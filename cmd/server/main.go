package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jinw00Arise/Jinwoo/internal/crypto"
	"github.com/Jinw00Arise/Jinwoo/internal/data/db"
	"github.com/Jinw00Arise/Jinwoo/internal/data/providers"
	"github.com/Jinw00Arise/Jinwoo/internal/data/providers/wz"
	"github.com/Jinw00Arise/Jinwoo/internal/data/repositories"
	"github.com/Jinw00Arise/Jinwoo/internal/game/server"
)

const shutdownTimeout = 30 * time.Second

func main() {
	log.Println("Starting Unified Game Server")

	// Initialize crypto
	if err := crypto.Init(); err != nil {
		log.Fatalf("crypto.Init() failed: %v", err)
	}

	// Load configuration
	cfg := server.Load()

	// Database Connection
	dbConn, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	log.Println("[Server] Database connected")

	// Create repositories
	repos := server.Repositories{
		Accounts:   repositories.NewAccountRepository(dbConn),
		Characters: repositories.NewCharacterRepo(dbConn),
		Items:      repositories.NewItemRepo(dbConn),
		Quests:     repositories.NewQuestRepo(dbConn),
	}

	// Initialize WZ data providers
	log.Printf("[Server] Loading WZ data from: %s", cfg.WZPath)
	wzProvider := wz.NewWzProvider(cfg.WZPath)

	itemProvider, err := providers.NewItemProvider(wzProvider)
	if err != nil {
		log.Fatalf("Failed to initialize item provider: %v", err)
	}

	mapProvider := providers.NewMapProvider(wzProvider)

	questProvider, err := providers.NewQuestProvider(wzProvider)
	if err != nil {
		log.Fatalf("Failed to initialize quest provider: %v", err)
	}

	npcProvider := providers.NewNPCProvider(wzProvider)

	provs := server.Providers{
		Items:  itemProvider,
		Maps:   mapProvider,
		Quests: questProvider,
		NPCs:   npcProvider,
	}

	// Create unified server
	srv := server.NewServer(cfg, repos, provs)

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
		log.Println("Server shutdown complete")
	}
}
