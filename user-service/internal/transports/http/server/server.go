package server

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/config"
	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/logger"
	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/service"
	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/transports/http/server/handler"
	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/transports/http/server/middleware"
)

type Opts struct {
	Config        *config.HTTPServer
	Logger        logger.Logger
	HealthService service.HealthService
	UserService   service.UserService
}

type HTTPServer struct {
	Config *config.HTTPServer
	Server *http.Server
	Logger logger.Logger
}

func NewServer(opts *Opts) *HTTPServer {
	r := gin.New()
	gin.SetMode(gin.ReleaseMode)

	r.Use(
		gin.Recovery(),
		middleware.ZerologMiddleware(),
	)

	healthHandler := handler.NewHealthHandler(&handler.HealthHandlerOpts{
		HealthService: opts.HealthService,
		Logger:        opts.Logger,
	})
	userHandler := handler.NewUserHandler(&handler.UserHandlerOpts{
		UserService: opts.UserService,
	})

	r.GET("/livez", healthHandler.Livez)
	r.GET("/readyz", healthHandler.Readyz)
	r.POST("/users", userHandler.CreateUser)

	return &HTTPServer{
		Config: opts.Config,
		Server: &http.Server{
			Addr:    opts.Config.URL,
			Handler: r,
		},
		Logger: opts.Logger,
	}
}

func (h *HTTPServer) ServeListener(listener net.Listener) error {
	h.Logger.Info("HTTP server started", logger.Field{Key: "address", Value: listener.Addr().String()})
	if err := h.Server.Serve(listener); err != nil && err != http.ErrServerClosed {
		h.Logger.Error("HTTP server failed", logger.Field{Key: "error", Value: err.Error()})
		return err
	}
	return nil
}

func (h *HTTPServer) Serve() error {
	listener, err := net.Listen("tcp", h.Config.URL)
	if err != nil {
		h.Logger.Error("Failed to create HTTP listener",
			logger.Field{Key: "address", Value: h.Config.URL},
			logger.Field{Key: "error", Value: err.Error()},
		)
		return err
	}

	return h.ServeListener(listener)
}
