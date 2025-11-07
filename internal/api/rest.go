package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shivarajshanthaiah/tunnel-relay-service/internal/hub"
)

func RegisterREST(r *gin.Engine, h *hub.Hub) {
	r.GET("/clients", func(c *gin.Context) {
		clients := h.ListClients()
		c.JSON(http.StatusOK, gin.H{"connected_clients": clients})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
