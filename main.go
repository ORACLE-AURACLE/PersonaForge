package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/PersonaForge/backend/docs"
	"github.com/PersonaForge/backend/internal/config"
	"github.com/PersonaForge/backend/internal/server"
)

// @title PersonaForge API
// @version 1.0
// @description Security-first backend for AI persona simulation
// @termsOfService http://swagger.io/terms/
//
// @contact.name API Support
// @contact.email support@personaforge.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /api
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Println("PersonaForge Backend")
	fmt.Println("===================")
	fmt.Printf("Environment: %s\n", cfg.Environment)
	fmt.Printf("Port: %s\n", cfg.Port)
	fmt.Printf("Gemini Model: %s\n", cfg.GeminiModel)
	fmt.Println()

	ctx := context.Background()

	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()

	if err := srv.Setup(ctx); err != nil {
		log.Fatalf("Failed to setup server: %v", err)
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		_ = srv.Close()
		os.Exit(0)
	}()

	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}


