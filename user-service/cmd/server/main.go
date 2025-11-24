package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"

	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/config"
	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/logger"
	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/rabbitmq"
	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/service"
	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/transports/http/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log := logger.NewZerologLogger("info", os.Stderr)

	cfg, err := config.NewConfig(log)
	if err != nil {
		log.Fatal(err.Error())
	}

	rmq := rabbitmq.NewRabbitMQ(ctx, &rabbitmq.Opts{
		Logger: log,
		Config: cfg.AMQP,
	})

	healthService := service.NewHealthService(&service.HealthServiceOpts{
		Checks: map[string]service.DependencyHealthCheck{
			"rabbitmq": func(ctx context.Context) error {
				return rmq.Health()
			},
		},
	})
	userService := service.NewUserService(&service.UserServiceOpts{
		Rabbitmq: rmq,
		Logger:   log,
	})

	httpServer := server.NewServer(&server.Opts{
		Config:        cfg.HTTPServer,
		Logger:        log,
		HealthService: healthService,
		UserService:   userService,
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
