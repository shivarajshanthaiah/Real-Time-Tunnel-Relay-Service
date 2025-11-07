package hub_test

import (
	"context"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shivarajshanthaiah/tunnel-relay-service/internal/client"
	"github.com/shivarajshanthaiah/tunnel-relay-service/internal/hub"
	"github.com/shivarajshanthaiah/tunnel-relay-service/pkg"
	"go.uber.org/zap"
)

// fakeConn mocks a websocket.Conn for testing
type fakeConn struct{}

func (f *fakeConn) WriteMessage(mt int, data []byte) error            { return nil }
func (f *fakeConn) NextWriter(mt int) (writer interface{}, err error) { return nil, nil }
func (f *fakeConn) Close() error                                      { return nil }
func (f *fakeConn) SetWriteDeadline(t time.Time) error                { return nil }
func (f *fakeConn) SetReadDeadline(t time.Time) error                 { return nil }
func (f *fakeConn) SetReadLimit(n int64)                              {}
func (f *fakeConn) SetPongHandler(h func(string) error)               {}

func TestHub_RegisterUnregister(t *testing.T) {
	logger := zap.NewExample()
	ctx := context.Background()
	h := hub.NewHub(ctx, logger)
	go h.Run()

	c := client.NewClient("test1", &websocket.Conn{}, h)
	h.Register(c)
	time.Sleep(100 * time.Millisecond)

	clients := h.ListClients()
	if len(clients) != 1 || clients[0] != "test1" {
		t.Fatalf("expected test1 in clients list, got %v", clients)
	}

	h.Unregister(c)
	time.Sleep(100 * time.Millisecond)
	clients = h.ListClients()
	if len(clients) != 0 {
		t.Fatalf("expected no clients, got %v", clients)
	}

	h.Shutdown()
}

func TestHub_SendAdminMessage(t *testing.T) {
	logger := zap.NewExample()
	ctx := context.Background()
	h := hub.NewHub(ctx, logger)
	go h.Run()

	// client with mock send buffer
	c := client.NewClient("c1", &websocket.Conn{}, h)
	h.Register(c)
	time.Sleep(100 * time.Millisecond)

	msg := pkg.AdminMessage{Target: "c1", Message: "hello"}
	h.SendAdminMessage(msg)

	// give hub a moment
	time.Sleep(100 * time.Millisecond)

	h.Shutdown()
}