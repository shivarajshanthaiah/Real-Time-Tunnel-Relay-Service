package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/shivarajshanthaiah/tunnel-relay-service/internal/client"
	"github.com/shivarajshanthaiah/tunnel-relay-service/internal/hub"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func ClientWSHandler(h *hub.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Query("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
			return
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			h.Logger().Error("client ugrade failed", zap.Error(err))
			return
		}
		// construct client and serve
		cl := client.NewClient(id, conn, h)
		cl.Serve()
	}
}
