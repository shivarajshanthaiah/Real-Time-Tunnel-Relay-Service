package boot

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NewHTTPServer creates an http.Server with given gin engine and address.
func NewHTTPServer(addr string, engine *gin.Engine) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: engine,
		//timeouts
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// NewLogger creates a structured zap logger.
func NewLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		panic("failed to initialize zap logger: " + err.Error())
	}
	return logger
}
