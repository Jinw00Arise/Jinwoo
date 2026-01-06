package main

import (
	"log"

	"github.com/Jinw00Arise/Jinwoo/config"
	"github.com/Jinw00Arise/Jinwoo/internal/channel"
	"github.com/Jinw00Arise/Jinwoo/internal/database"
	"github.com/Jinw00Arise/Jinwoo/internal/database/repository"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
	"github.com/Jinw00Arise/Jinwoo/internal/script"
	"github.com/Jinw00Arise/Jinwoo/internal/wz"
)

func main() {
	log.Println("Starting Jinwoo Channel Server...")

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

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	log.Println("Database connected")

	characters := repository.NewCharacterRepository(db)
	server := channel.NewServer(cfg, characters)

	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

