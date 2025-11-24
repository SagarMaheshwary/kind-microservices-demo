package config

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/gofor-little/env"
	"github.com/sagarmaheshwary/kind-microservices-demo/user-service/internal/logger"
)

type LoaderOptions struct {
	EnvPath   string
	EnvLoader func(string) error
	Logger    logger.Logger
}

type Config struct {
	HTTPServer *HTTPServer
	AMQP       *AMQP
}

type HTTPServer struct {
	URL             string
	ShutdownTimeout time.Duration
}

type AMQP struct {
	Host                    string
	Port                    int
	Username                string
	Password                string
	PublishTimeout          time.Duration
	ConnectionRetryInterval time.Duration
	ConnectionRetryAttempts int
}

func NewConfig(log logger.Logger) (*Config, error) {
	return NewConfigWithOptions(LoaderOptions{
		EnvPath: path.Join(rootDir(), "..", ".env"),
		Logger:  log,
	})
}

func NewConfigWithOptions(opts LoaderOptions) (*Config, error) {
	log := opts.Logger
	if log == nil {
		log = logger.NewZerologLogger("info", os.Stderr)
	}

	envLoader := opts.EnvLoader
	if envLoader == nil {
		envLoader = func(path string) error {
			_, err := os.Stat(path)
			if err != nil {
				return err
			}

			return env.Load(path)
		}
	}

	if err := envLoader(opts.EnvPath); err == nil {
		log.Info("Loaded environment variables from" + opts.EnvPath)
	} else {
		log.Info("failed to load .env file, using system environment variables")
	}

	cfg := &Config{
		HTTPServer: &HTTPServer{
			URL:             getEnv("HTTP_SERVER_URL", ":4000"),
			ShutdownTimeout: getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 5*time.Second),
		},
		AMQP: &AMQP{
			Host:                    getEnv("AMQP_HOST", "rabbitmq"),
			Port:                    getEnvInt("AMQP_PORT", 5672),
			Username:                getEnv("AMQP_USERNAME", "default"),
			Password:                getEnv("AMQP_PASSWORD", "default"),
			PublishTimeout:          getEnvDuration("AMQP_PUBLISH_TIMEOUT_SECONDS", time.Second*5),
			ConnectionRetryInterval: getEnvDuration("AMQP_CONNECTION_RETRY_INTERVAL_SECONDS", time.Second*5),
			ConnectionRetryAttempts: getEnvInt("AMQP_CONNECTION_RETRY_ATTEMPTS", 10),
		},
	}

	return cfg, nil
}

func rootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))

	return filepath.Dir(d)
}

func getEnv(key string, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return val
	}

	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val, err := time.ParseDuration(os.Getenv(key)); err == nil {
		return val
	}

	return defaultVal
}
