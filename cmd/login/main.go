package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jinw00Arise/Jinwoo/config"
	"github.com/Jinw00Arise/Jinwoo/internal/crypto"
	"github.com/Jinw00Arise/Jinwoo/internal/database"
	"github.com/Jinw00Arise/Jinwoo/internal/database/repository"
	"github.com/Jinw00Arise/Jinwoo/internal/login"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
)

const shutdownTimeout = 30 * time.Second

func main() {
	log.Println("Starting Jinwoo Login Server...")

	// Initialize crypto subsystem first
	if err := crypto.Init(); err != nil {
		log.Fatalf("Crypto initialization failed: %v", err)
	}

	cfg := config.Load()
	network.SetDebugPackets(cfg.DebugPackets)
	login.SetAutoRegister(cfg.AutoRegister)

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	log.Println("Database connected")

	accounts := repository.NewAccountRepository(db)
	characters := repository.NewCharacterRepository(db)
	inventories := repository.NewInventoryRepository(db)
	server := login.NewServer(cfg, accounts, characters, inventories)

	// Setup signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := server.Start(); err != nil {
			serverErr <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case <-ctx.Done():
		log.Println("Shutdown signal received, initiating graceful shutdown...")
	case err := <-serverErr:
		log.Fatalf("Server error: %v", err)
	}

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Stop accepting new connections
	if err := server.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}

	// Wait for shutdown to complete or timeout
	select {
	case <-shutdownCtx.Done():
		if shutdownCtx.Err() == context.DeadlineExceeded {
			log.Println("Shutdown timeout exceeded, forcing exit")
		}
	default:
		log.Println("Login server shutdown complete")
	}
}
