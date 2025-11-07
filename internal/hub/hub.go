package hub

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/shivarajshanthaiah/tunnel-relay-service/internal/client"
	"github.com/shivarajshanthaiah/tunnel-relay-service/pkg"
	"go.uber.org/zap"
)

// Hub mediates between admin and client connections.
type Hub struct {
	clients map[string]*client.Client
	mu      sync.RWMutex

	register   chan *client.Client
	unregister chan *client.Client
	adminMsg   chan pkg.AdminMessage

	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
	logger *zap.Logger
}

// NewHub creates a new Hub.
func NewHub(parentCtx context.Context, logger *zap.Logger) *Hub {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Hub{
		clients:    make(map[string]*client.Client),
		register:   make(chan *client.Client),
		unregister: make(chan *client.Client),
		adminMsg:   make(chan pkg.AdminMessage),
		ctx:        ctx,
		cancel:     cancel,
		wg:         &sync.WaitGroup{},
		logger:     logger,
	}
}

// Run starts the hub loop call in a goroutine.
func (h *Hub) Run() {
	h.logger.Info("hub started")

	for {
		select {
		case <-h.ctx.Done():
			h.logger.Info("hub shutdown initiated")
			h.mu.Lock()
			for _, c := range h.clients {
				c.Close()
			}
			h.clients = make(map[string]*client.Client)
			h.mu.Unlock()
			h.wg.Wait() // Wait for all client pumps to finish
			h.logger.Info("hub shutdown complete")
			return

		case c := <-h.register:
			h.mu.Lock()
			if old, ok := h.clients[c.ID()]; ok {
				h.logger.Warn("replacing existing client", zap.String("id", c.ID()))
				old.Close()
			}
			h.clients[c.ID()] = c
			h.mu.Unlock()
			h.logger.Info("client registered", zap.String("id", c.ID()))

		case c := <-h.unregister:
			h.mu.Lock()
			if cur, ok := h.clients[c.ID()]; ok && cur == c {
				delete(h.clients, c.ID())
				h.logger.Info("client unregistered", zap.String("id", c.ID()))
			}
			h.mu.Unlock()

		case msg := <-h.adminMsg:
			if msg.Target == "*" {
				h.mu.RLock()
				for _, c := range h.clients {
					h.safeSend(c, msg.Message)
				}
				h.mu.RUnlock()
				continue
			}

			h.mu.RLock()
			c, ok := h.clients[msg.Target]
			h.mu.RUnlock()
			if !ok {
				h.logger.Warn("admin target not found", zap.String("target", msg.Target))
				continue
			}
			h.safeSend(c, msg.Message)
		}
	}
}

// safeSend marshals message and pushes to client's send method.
func (h *Hub) safeSend(c *client.Client, message string) {
	payload := map[string]string{"from": "admin", "message": message}
	b, err := json.Marshal(payload)
	if err != nil {
		h.logger.Error("failed to marshal payload", zap.Error(err))
		return
	}
	if ok := c.Send(b); !ok {
		h.logger.Warn("client send buffer full or closed", zap.String("id", c.ID()))
		c.Close()
	}
}

// Register forwards client to register channel
func (h *Hub) Register(c *client.Client) {
	select {
	case h.register <- c:
	case <-h.ctx.Done():
	}
}

// Unregister forwards client to unregister channel
func (h *Hub) Unregister(c *client.Client) {
	select {
	case h.unregister <- c:
	case <-h.ctx.Done():
	}
}

// SendAdminMessage forwards admin message to hub
func (h *Hub) SendAdminMessage(m pkg.AdminMessage) {
	select {
	case h.adminMsg <- m:
	case <-h.ctx.Done():
	}
}

// ListClients returns a snapshot of client IDs
func (h *Hub) ListClients() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]string, 0, len(h.clients))
	for id := range h.clients {
		out = append(out, id)
	}
	return out
}

// Shutdown cancels hub context.
func (h *Hub) Shutdown() {
	h.logger.Warn("hub shutdown initiated")
	h.cancel()
}

func (h *Hub) Wg() *sync.WaitGroup {
	return h.wg
}

func (h *Hub) Logger() *zap.Logger {
	return h.logger
}
