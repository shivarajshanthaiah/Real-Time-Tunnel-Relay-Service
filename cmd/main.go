package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shivarajshanthaiah/tunnel-relay-service/configs"
	"github.com/shivarajshanthaiah/tunnel-relay-service/internal/api"
	"github.com/shivarajshanthaiah/tunnel-relay-service/internal/boot"
	"github.com/shivarajshanthaiah/tunnel-relay-service/internal/handlers"
	"github.com/shivarajshanthaiah/tunnel-relay-service/internal/hub"
	"go.uber.org/zap"
)

func main() {
	// initialise gin with logger and recovery mode
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	logger := boot.NewLogger()
	defer logger.Sync()

	// create hub
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := hub.NewHub(ctx, logger)
	go h.Run()

	// register REST endpoints
	api.RegisterREST(r, h)

	// register WS handlers
	r.GET("/ws/client", handlers.ClientWSHandler(h))
	r.GET("/ws/admin", handlers.AdminWSHandler(h))

	// Load config
	cnfg := configs.LoadConfig()
	srv := boot.NewHTTPServer(":"+cnfg.SERVERPORT, r)

	logger.Info("server starting", zap.String("addr", srv.Addr))

	// run server
	go func() {
		log.Printf("server: listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server listen failed", zap.Error(err))
		}
	}()

	// graceful shutdown on SIGINT/SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	logger.Info("shutdown signal received")

	// stop accepting new requests
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	//stop hub and wait for all goroutines
	logger.Info("initiating hub shutdown")

	h.Shutdown()
	h.Wg().Wait()

	logger.Info("shutdown complete")
}
