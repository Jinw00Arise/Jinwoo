package main

import (
	"log"

	"github.com/Jinw00Arise/Jinwoo/config"
	"github.com/Jinw00Arise/Jinwoo/internal/channel"
	"github.com/Jinw00Arise/Jinwoo/internal/database"
	"github.com/Jinw00Arise/Jinwoo/internal/database/repository"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
)

func main() {
	log.Println("Starting Jinwoo Channel Server...")

	cfg := config.LoadChannel()
	network.SetDebugPackets(cfg.DebugPackets)

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

