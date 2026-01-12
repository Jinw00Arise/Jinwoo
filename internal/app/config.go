package app

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Host string
	Port string
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

func LoadChannel() *ChannelConfig {
	_ = godotenv.Load() // Ignore error - .env is optional

	return &ChannelConfig{
		ServerConfig: ServerConfig{
			Host: getEnv("CHANNEL_HOST", "127.0.0.1"),
			Port: getEnv("CHANNEL_PORT", "8585"),
		},
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://localhost:5432/jinwoo?sslmode=disable"),
		DebugPackets: getEnv("DEBUG_PACKETS", "") != "",
		GameVersion:  95,
		PatchVersion: "1",
		Locale:       8,
		WorldID:      0,
		ChannelID:    getEnvByte("CHANNEL_ID", 0),
		WZPath:       getEnv("WZ_PATH", "data/wz"),
		ScriptsPath:  getEnv("SCRIPTS_PATH", "scripts"),
		ExpRate:      getEnvFloat("EXP_RATE", 1.0),
		QuestExpRate: getEnvFloat("QUEST_EXP_RATE", 1.0),
		MesoRate:     getEnvFloat("MESO_RATE", 1.0),
		DropRate:     getEnvFloat("DROP_RATE", 1.0),
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

func getEnvByte(key string, fallback byte) byte {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.ParseInt(val, 10, 8); err == nil {
			return byte(i)
		}
	}
	return fallback
}
