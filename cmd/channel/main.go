package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/crypto"
	"github.com/Jinw00Arise/Jinwoo/internal/data/db"
	"github.com/Jinw00Arise/Jinwoo/internal/data/providers"
	"github.com/Jinw00Arise/Jinwoo/internal/data/repositories"
	"github.com/Jinw00Arise/Jinwoo/internal/game/channel"
	"github.com/Jinw00Arise/Jinwoo/internal/game/field"
)

const shutdownTimeout = 30 * time.Second

func main() {
	// Parse command-line flags
	channelID := flag.Int("channel", -1, "Channel ID (overrides env)")
	port := flag.String("port", "", "Channel port (overrides env)")
	flag.Parse()

	log.Println("Starting Channel Server")

	if err := crypto.Init(); err != nil {
		log.Fatalf("crypto.Init() failed: %v", err)
	}

	cfg := channel.LoadChannel()

	// Override with command-line flags if provided
	if *channelID >= 0 {
		cfg.ChannelID = byte(*channelID)
	}
	if *port != "" {
		cfg.Port = *port
	}

	// Database Connection
	dbConn, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	log.Println("[Channel] Database connected")

	charRepo := repositories.NewCharacterRepo(dbConn)
	invRepo := repositories.NewInventoryRepo(dbConn)

	// Initialize WZ data providers
	log.Printf("[Channel] Loading WZ data from: %s", cfg.WZPath)
	wzProvider := providers.NewWzProvider(cfg.WZPath)
	mapProvider := providers.NewMapProvider(wzProvider)

	// Create field manager with map provider
	fieldMgr := field.NewManager(mapProvider)
	log.Println("[Channel] Field manager initialized")

	srv := channel.NewServer(cfg, charRepo, invRepo, fieldMgr)

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
