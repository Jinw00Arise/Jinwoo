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
	DatabaseURL     string
	DebugPackets    bool
	AutoRegister    bool
	GameVersion     uint16
	PatchVersion    string
	Locale          byte
	ChannelCount    int
	ChannelBasePort int
	ChannelHost     string
	ChannelPort     string
	WZPath          string
}

func LoadLogin() *LoginConfig {
	_ = godotenv.Load() // Ignore error - .env is optional

	return &LoginConfig{
		ServerConfig: ServerConfig{
			Host: getEnv("LOGIN_HOST", "127.0.0.1"),
			Port: getEnv("LOGIN_PORT", "8484"),
		},
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://localhost:5432/jinwoo?sslmode=disable"),
		DebugPackets:    getEnv("DEBUG_PACKETS", "") != "",
		GameVersion:     95,
		PatchVersion:    "1",
		Locale:          8,
		ChannelHost:     getEnv("CHANNEL_HOST", "127.0.0.1"),
		ChannelPort:     getEnv("CHANNEL_PORT", "8585"),
		ChannelCount:    getEnvInt("CHANNEL_COUNT", 1),
		ChannelBasePort: getEnvInt("CHANNEL_BASE_PORT", 8585),
		WZPath:          getEnv("WZ_PATH", "./wz"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
