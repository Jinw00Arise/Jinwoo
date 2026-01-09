package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Jinw00Arise/Jinwoo/config"
	"github.com/Jinw00Arise/Jinwoo/internal/channel"
	"github.com/Jinw00Arise/Jinwoo/internal/crypto"
	"github.com/Jinw00Arise/Jinwoo/internal/database"
	"github.com/Jinw00Arise/Jinwoo/internal/database/repository"
	"github.com/Jinw00Arise/Jinwoo/internal/game/drops"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
	"github.com/Jinw00Arise/Jinwoo/internal/script"
	"github.com/Jinw00Arise/Jinwoo/internal/wz"
)

const shutdownTimeout = 30 * time.Second

func main() {
	log.Println("Starting Jinwoo Channel Server...")

	// Initialize crypto subsystem first
	if err := crypto.Init(); err != nil {
		log.Fatalf("Crypto initialization failed: %v", err)
	}

	cfg := config.LoadChannel()
	network.SetDebugPackets(cfg.DebugPackets)

	// Initialize WZ data
	if err := wz.Init(cfg.WZPath); err != nil {
		log.Printf("Warning: WZ data initialization failed: %v", err)
	}
	log.Println("WZ data initialized")

	// Initialize script engine
	if err := script.Init(cfg.ScriptsPath); err != nil {
		log.Printf("Warning: Script initialization failed: %v", err)
	}
	log.Println("Script engine initialized")

	// Initialize drop tables
	dropTablePath := filepath.Join("data", "drop_tables.json")
	if _, err := os.Stat(dropTablePath); err == nil {
		if err := drops.GetInstance().LoadFromFile(dropTablePath); err != nil {
			log.Printf("Warning: Failed to load drop tables: %v", err)
		}
	}
	log.Println("Drop tables initialized")

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	log.Println("Database connected")

	characters := repository.NewCharacterRepository(db)
	inventories := repository.NewInventoryRepository(db)
	server := channel.NewServer(cfg, characters, inventories)

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

	// Graceful shutdown - save all player states
	log.Println("Saving all player states...")
	server.SaveAllPlayers()

	// Stop accepting new connections and tick loop
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
		log.Println("Channel server shutdown complete")
	}
}

