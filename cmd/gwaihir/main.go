// Gwaihir is a minimal HTTP API that sends Wake-on-LAN packets on command to other servers in the network.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/josimar-silva/gwaihir/internal/config"
	httpdelivery "github.com/josimar-silva/gwaihir/internal/delivery/http"
	"github.com/josimar-silva/gwaihir/internal/infrastructure"
	"github.com/josimar-silva/gwaihir/internal/repository"
	"github.com/josimar-silva/gwaihir/internal/usecase"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := loadConfiguration()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	logger := initializeLogger(cfg)

	metrics, err := initializeMetrics(logger)
	if err != nil {
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	repo, err := initializeRepository(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	logMachineConfiguration(logger, metrics, repo)
	useCase := initializeUseCase(repo, logger, metrics)
	handler := initializeHandler(useCase, logger, metrics)
	router := initializeRouter(handler, cfg, logger)

	if err := startServer(cfg, router, logger); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func loadConfiguration() (*config.Config, error) {
	configPath := os.Getenv("GWAIHIR_CONFIG")
	if configPath == "" {
		configPath = "/etc/gwaihir/gwaihir.yaml"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
	}

	return cfg, nil
}

func initializeLogger(cfg *config.Config) *infrastructure.Logger {
	logger := infrastructure.NewLogger(cfg.Server.Log.Format, cfg.Server.Log.Level)
	logger.Info("Configuration loaded",
		infrastructure.String("summary", cfg.String()),
	)
	return logger
}

func initializeMetrics(logger *infrastructure.Logger) (*infrastructure.Metrics, error) {
	metrics, err := infrastructure.NewMetrics()
	if err != nil {
		logger.Error("Failed to initialize metrics", infrastructure.Any("error", err))
		return nil, fmt.Errorf("metrics initialization failed: %w", err)
	}
	return metrics, nil
}

func initializeRepository(cfg *config.Config, logger *infrastructure.Logger) (*repository.InMemoryMachineRepository, error) {
	repo, err := repository.NewInMemoryMachineRepository(cfg)
	if err != nil {
		logger.Error("Failed to initialize machine repository", infrastructure.Any("error", err))
		return nil, fmt.Errorf("repository initialization failed: %w", err)
	}
	return repo, nil
}

func logMachineConfiguration(logger *infrastructure.Logger, metrics *infrastructure.Metrics, repo *repository.InMemoryMachineRepository) {
	machines, _ := repo.GetAll()
	logger.Info("Machine configuration loaded", infrastructure.Int("count", len(machines)))
	for _, m := range machines {
		logger.Debug("Machine registered", infrastructure.String("name", m.Name), infrastructure.String("id", m.ID))
	}
	metrics.ConfiguredMachines.Set(float64(len(machines)))
}

func initializeUseCase(repo *repository.InMemoryMachineRepository, logger *infrastructure.Logger, metrics *infrastructure.Metrics) *usecase.WoLUseCase {
	packetSender := repository.NewWoLPacketSender()
	return usecase.NewWoLUseCase(repo, packetSender, logger, metrics)
}

func initializeHandler(useCase *usecase.WoLUseCase, logger *infrastructure.Logger, metrics *infrastructure.Metrics) *httpdelivery.Handler {
	return httpdelivery.NewHandler(useCase, logger, metrics, Version, BuildTime, GitCommit)
}

func initializeRouter(handler *httpdelivery.Handler, cfg *config.Config, logger *infrastructure.Logger) *gin.Engine {
	if cfg.Server.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	if cfg.Authentication.APIKey == "" {
		logger.Warn("No API key configured - protected endpoints will not require authentication")
	}

	return httpdelivery.NewRouterWithConfig(handler, cfg)
}

func startServer(cfg *config.Config, router *gin.Engine, logger *infrastructure.Logger) error {
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	logger.Info("Server starting",
		infrastructure.String("address", addr),
		infrastructure.String("version", Version),
		infrastructure.String("buildTime", BuildTime),
		infrastructure.String("gitCommit", GitCommit),
	)

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		logger.Info("Shutdown signal received")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Server shutdown error", infrastructure.Any("error", err))
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Server listen error", infrastructure.Any("error", err))
		return fmt.Errorf("server listen failed: %w", err)
	}

	logger.Info("Server stopped gracefully")
	return nil
}
