package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/shivarajshanthaiah/tunnel-relay-service/internal/hub"
	"github.com/shivarajshanthaiah/tunnel-relay-service/pkg"
	"go.uber.org/zap"
)

var adminUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func AdminWSHandler(h *hub.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := adminUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			h.Logger().Error("admin ugrade failed", zap.Error(err))
			return
		}
		defer conn.Close()
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					h.Logger().Error("admin ws read error", zap.Error(err))
				}
				return
			}
			var m pkg.AdminMessage
			if err := json.Unmarshal(data, &m); err != nil {
				h.Logger().Error("admin: invalid json", zap.Error(err))
				_ = conn.WriteMessage(websocket.TextMessage, []byte("invalid json"))
				continue
			}
			if m.Target == "" {
				_ = conn.WriteMessage(websocket.TextMessage, []byte("target required"))
				continue
			}
			// forward to hub
			h.SendAdminMessage(m)
			// acknowledge
			ack := map[string]string{"status": "ok", "target": m.Target}
			if ackb, err := json.Marshal(ack); err == nil {
				_ = conn.WriteMessage(websocket.TextMessage, ackb)
			}
		}
	}
}
