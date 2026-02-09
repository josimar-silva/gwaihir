// Package main implements Gwaihir, the WoL messenger service.
// Gwaihir is a minimal HTTP API that sends Wake-on-LAN packets on command from Smaug.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	httpdelivery "github.com/josimar-silva/gwaihir/internal/delivery/http"
	"github.com/josimar-silva/gwaihir/internal/repository"
	"github.com/josimar-silva/gwaihir/internal/usecase"
)

func main() {
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

	log.Printf("Loading machine configuration from: %s", configPath)
	machineRepo, err := repository.NewYAMLMachineRepository(configPath)
	if err != nil {
		log.Fatalf("Failed to initialize machine repository: %v", err)
	}

	machines, _ := machineRepo.GetAll()
	log.Printf("Loaded %d machine(s) into allowlist", len(machines))
	for _, m := range machines {
		log.Printf("  - %s (%s): %s", m.Name, m.ID, m.MAC)
	}

	packetSender := repository.NewWoLPacketSender()
	wolUseCase := usecase.NewWoLUseCase(machineRepo, packetSender)
	handler := httpdelivery.NewHandler(wolUseCase, Version, BuildTime, GitCommit)
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

	log.Printf("Gwaihir, the Windlord, is listening on %s...", addr)
	log.Printf("Version: %s, BuildTime: %s, GitCommit: %s", Version, BuildTime, GitCommit)

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	log.Println("Server stopped gracefully")
}
