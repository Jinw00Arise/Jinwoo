package login

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Host string
	Port string
}

type LoginConfig struct {
	ServerConfig
	DatabaseURL  string
	DebugPackets bool
	AutoRegister bool
	GameVersion  uint16
	PatchVersion string
	Locale       byte
	ChannelHost  string
	ChannelPort  string
}

type ChannelConfig struct {
	ServerConfig
	DatabaseURL  string
	DebugPackets bool
	GameVersion  uint16
	PatchVersion string
	Locale       byte
	WorldID      byte
	ChannelID    byte
	WZPath       string
	ScriptsPath  string
	// Rate multipliers
	ExpRate      float64 // General EXP rate multiplier (default 1.0)
	QuestExpRate float64 // Quest EXP rate multiplier (default 1.0)
	MesoRate     float64 // Meso drop rate multiplier (default 1.0)
	DropRate     float64 // Item drop rate multiplier (default 1.0)
}

func Load() *LoginConfig {
	_ = godotenv.Load() // Ignore error - .env is optional

	return &LoginConfig{
		ServerConfig: ServerConfig{
			Host: getEnv("LOGIN_HOST", "127.0.0.1"),
			Port: getEnv("LOGIN_PORT", "8484"),
		},
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://localhost:5432/jinwoo?sslmode=disable"),
		DebugPackets: getEnv("DEBUG_PACKETS", "") != "",
		AutoRegister: getEnv("AUTO_REGISTER", "") != "",
		GameVersion:  95,
		PatchVersion: "1",
		Locale:       8, // GMS
		ChannelHost:  getEnv("CHANNEL_HOST", "127.0.0.1"),
		ChannelPort:  getEnv("CHANNEL_PORT", "8585"),
	}
}

func LoadLogin() *LoginConfig {
	_ = godotenv.Load() // Ignore error - .env is optional

	return &LoginConfig{
		ServerConfig: ServerConfig{
			Host: getEnv("LOGIN_HOST", "127.0.0.1"),
			Port: getEnv("LOGIN_PORT", "8484"),
		},
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://localhost:5432/jinwoo?sslmode=disable"),
		DebugPackets: getEnv("DEBUG_PACKETS", "") != "",
		GameVersion:  95,
		PatchVersion: "1",
		Locale:       8,
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return fallback
}
