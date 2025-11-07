package client

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	writeWait   = 10 * time.Second
	pongWait    = 60 * time.Second
	pingPeriod  = 30 * time.Second
	maxMsgSize  = 512
	sendBufSize = 256
)

// Hub defines the minimal interface client needs to register/unregister.
type Hub interface {
	Register(c *Client)
	Unregister(c *Client)
	Wg() *sync.WaitGroup
	Logger() *zap.Logger
}

// Client represents a connected websocket client.
type Client struct {
	id     string
	conn   *websocket.Conn
	hub    Hub
	send   chan []byte
	sendMu sync.Mutex // protects closing send channel
	mu     sync.Mutex // protects closed
	closed bool
}

func NewClient(id string, conn *websocket.Conn, h Hub) *Client {
	conn.SetReadLimit(maxMsgSize)
	return &Client{
		id:   id,
		conn: conn,
		hub:  h,
		send: make(chan []byte, sendBufSize),
	}
}

// ID returns client id.
func (c *Client) ID() string {
	return c.id
}

// Send attempts to queue a message for this client.
// It returns true if queued, false if the client send buffer is full or client closed.
func (c *Client) Send(b []byte) bool {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return false
	}
	c.mu.Unlock()

	select {
	case c.send <- b:
		return true
	default:
		// buffer full -> drop
		return false
	}
}

// Close closes the websocket and marks client closed
func (c *Client) Close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	c.mu.Unlock()

	_ = c.conn.Close()

	// close send channel
	c.sendMu.Lock()
	close(c.send)
	c.sendMu.Unlock()
}

func (c *Client) readPump() {
	defer func() {
		// ask hub to remove this client
		c.hub.Unregister(c)
		c.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.hub.Logger().Debug("received pong", zap.String("client", c.id))
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.Logger().Warn("read error", zap.String("client", c.id), zap.Error(err))
			}
			break
		}
		// No client->server message handling needed for this assignment
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// channel closed
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				_ = w.Close()
				return
			}
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.hub.Logger().Debug("sending ping", zap.String("client", c.id))
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.hub.Logger().Debug("send ping error", zap.String("client", c.id), zap.Error(err))
				return
			}
		}
	}
}

// maintaining the counter for go routines
func (c *Client) runPump(fn func()) {
	c.hub.Wg().Add(1)
	go func() {
		defer c.hub.Wg().Done()
		fn()
	}()
}


// Serve registers the client with hub and starts read/write pumps.
func (c *Client) Serve() {
	c.hub.Register(c)
	c.runPump(c.readPump)
	c.runPump(c.writePump)
}
