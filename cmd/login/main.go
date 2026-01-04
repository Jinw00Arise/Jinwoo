package main

import (
	"log"

	"github.com/Jinw00Arise/Jinwoo/config"
	"github.com/Jinw00Arise/Jinwoo/internal/database"
	"github.com/Jinw00Arise/Jinwoo/internal/database/repository"
	"github.com/Jinw00Arise/Jinwoo/internal/login"
	"github.com/Jinw00Arise/Jinwoo/internal/network"
)

func main() {
	log.Println("Starting Jinwoo Login Server...")

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
	server := login.NewServer(cfg, accounts, characters)

	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
