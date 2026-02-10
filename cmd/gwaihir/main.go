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
	configPath := os.Getenv("GWAIHIR_CONFIG")
	if configPath == "" {
		configPath = "/etc/gwaihir/gwaihir.yaml"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logger := infrastructure.NewLogger(cfg.Server.Log.Format, cfg.Server.Log.Level)

	logger.Info("Configuration loaded", infrastructure.String("path", configPath))

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	metrics, err := infrastructure.NewMetrics()
	if err != nil {
		logger.Error("Failed to initialize metrics", infrastructure.Any("error", err))
		os.Exit(1)
	}

	machineRepo, err := repository.NewYAMLMachineRepository(cfg)
	if err != nil {
		logger.Error("Failed to initialize machine repository", infrastructure.Any("error", err))
		os.Exit(1)
	}

	machines, _ := machineRepo.GetAll()
	logger.Info("Machine configuration loaded", infrastructure.Int("count", len(machines)))
	for _, m := range machines {
		logger.Debug("Machine registered", infrastructure.String("name", m.Name), infrastructure.String("id", m.ID))
	}
	metrics.ConfiguredMachines.Set(float64(len(machines)))

	packetSender := repository.NewWoLPacketSender()
	wolUseCase := usecase.NewWoLUseCase(machineRepo, packetSender, logger, metrics)
	handler := httpdelivery.NewHandler(wolUseCase, logger, metrics, Version, BuildTime, GitCommit)

	apiKey := cfg.Authentication.APIKey
	if apiKey == "" {
		logger.Warn("No API key configured - protected endpoints will not require authentication")
	}
	router := httpdelivery.NewRouterWithAuth(handler, apiKey)

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
		os.Exit(1)
	}

	logger.Info("Server stopped gracefully")
}
