package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/app"
	"github.com/Jinw00Arise/Jinwoo/internal/crypto"
	"github.com/Jinw00Arise/Jinwoo/internal/data/db"
	"github.com/Jinw00Arise/Jinwoo/internal/data/repositories"
)

const shutdownTimeout = 30 * time.Second

func main() {
	log.Println("Starting Channel Server")

	if err := crypto.Init(); err != nil {
		log.Fatalf("crypto.Init() failed: %v", err)
	}

	cfg := app.LoadChannel()

	// Database Connection
	dbConn, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	log.Println("[Channel] Database connected")

	charRepo := repositories.NewCharacterRepo(dbConn)
	invRepo := repositories.NewInventoryRepo(dbConn)

	srv := app.NewServer(cfg, charRepo, invRepo)

	// Start server
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Start()
	}()

	// Wait for SIGINT/SIGTERM or server error
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-sig:
		log.Printf("Received signal %s, shutting down...", s)
	case err := <-serverErr:
		if err != nil {
			log.Printf("Server stopped with error: %v", err)
		} else {
			log.Printf("Server stopped")
		}
	}

	// Graceful stop with timeout
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	} else {
		log.Println("Channel server shutdown complete")
	}
}
