// Package main implements Gwaihir, the WoL messenger service.
// Gwaihir is a minimal HTTP API that sends Wake-on-LAN packets on command from Smaug.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	httpdelivery "github.com/josimar-silva/gwaihir/internal/delivery/http"
	"github.com/josimar-silva/gwaihir/internal/infrastructure"
	"github.com/josimar-silva/gwaihir/internal/repository"
	"github.com/josimar-silva/gwaihir/internal/usecase"
)

func main() {
	inProduction, _ := strconv.ParseBool(os.Getenv("GWAIHIR_PRODUCTION"))
	logger := infrastructure.NewLogger(inProduction)

	configPath := os.Getenv("GWAIHIR_CONFIG")
	if configPath == "" {
		configPath = "/etc/gwaihir/machines.yaml"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	metrics, err := infrastructure.NewMetrics()
	if err != nil {
		logger.Error("Failed to initialize metrics", infrastructure.Any("error", err))
		os.Exit(1)
	}

	logger.Info("Loading machine configuration", infrastructure.String("path", configPath))
	machineRepo, err := repository.NewYAMLMachineRepository(configPath)
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
	router := httpdelivery.NewRouter(handler)

	addr := fmt.Sprintf(":%s", port)
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
