package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"

	"github.com/sagarmaheshwary/kind-microservices-demo/api-gateway/internal/config"
	"github.com/sagarmaheshwary/kind-microservices-demo/api-gateway/internal/logger"
	"github.com/sagarmaheshwary/kind-microservices-demo/api-gateway/internal/service"
	"github.com/sagarmaheshwary/kind-microservices-demo/api-gateway/internal/transports/http/client"
	"github.com/sagarmaheshwary/kind-microservices-demo/api-gateway/internal/transports/http/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log := logger.NewZerologLogger("info", os.Stderr)

	cfg, err := config.NewConfig(log)
	if err != nil {
		log.Fatal(err.Error())
	}

	httpClient := client.NewUserClient(client.NewHTTPClient(cfg.UserClient.URL, cfg.UserClient.Timeout))

	healthService := service.NewHealthService(&service.HealthServiceOpts{
		Checks: map[string]service.DependencyHealthCheck{},
	})

	httpServer := server.NewServer(&server.Opts{
		Config:        cfg.HTTPServer,
		Logger:        log,
		HealthService: healthService,
		UserService:   service.NewUserService(httpClient),
	})
	go func() {
		err = httpServer.Serve()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			stop()
		}
	}()

	<-ctx.Done()

	log.Warn("Shutdown signal received, closing services!")

	healthService.SetReady(false)

	httpCtx, httpCancel := context.WithTimeout(context.Background(), cfg.HTTPServer.ShutdownTimeout)
	if err := httpServer.Server.Shutdown(httpCtx); err != nil {
		log.Error("failed to close http server", logger.Field{Key: "error", Value: err.Error()})
	}
	httpCancel()

	log.Info("Shutdown complete!")
}
