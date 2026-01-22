package server

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// WorldConfig represents configuration for a single world
type WorldConfig struct {
	WorldID   byte
	WorldName string
	Channels  []ChannelConfig
}

// ChannelConfig represents configuration for a single channel
type ChannelConfig struct {
	ChannelID byte
	Port      int
}

// Config holds the unified server configuration
type Config struct {
	// Server binding
	Host string

	// Login server
	LoginPort int

	// Database
	DatabaseURL string

	// Debug
	DebugPackets bool
	AutoRegister bool

	// Game version
	GameVersion  uint16
	PatchVersion string
	Locale       byte

	// Worlds configuration
	Worlds []WorldConfig

	// Paths
	WZPath      string
	ScriptsPath string

	// Rate multipliers
	ExpRate      float64
	QuestExpRate float64
	MesoRate     float64
	DropRate     float64
}

// Load loads the server configuration from environment variables
func Load() *Config {
	_ = godotenv.Load() // Ignore error - .env is optional

	cfg := &Config{
		Host:         getEnv("HOST", "127.0.0.1"),
		LoginPort:    getEnvInt("LOGIN_PORT", 8484),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://localhost:5432/jinwoo?sslmode=disable"),
		DebugPackets: getEnv("DEBUG_PACKETS", "") != "",
		AutoRegister: getEnv("AUTO_REGISTER", "true") != "",
		GameVersion:  95,
		PatchVersion: "1",
		Locale:       8,
		WZPath:       getEnv("WZ_PATH", "./wz"),
		ScriptsPath:  getEnv("SCRIPTS_PATH", "scripts"),
		ExpRate:      getEnvFloat("EXP_RATE", 1.0),
		QuestExpRate: getEnvFloat("QUEST_EXP_RATE", 1.0),
		MesoRate:     getEnvFloat("MESO_RATE", 1.0),
		DropRate:     getEnvFloat("DROP_RATE", 1.0),
	}

	// Build worlds configuration
	channelCount := getEnvInt("CHANNEL_COUNT", 1)
	channelBasePort := getEnvInt("CHANNEL_BASE_PORT", 8585)

	channels := make([]ChannelConfig, channelCount)
	for i := range channelCount {
		channels[i] = ChannelConfig{
			ChannelID: byte(i),
			Port:      channelBasePort + i,
		}
	}

	cfg.Worlds = []WorldConfig{
		{
			WorldID:   0,
			WorldName: getEnv("WORLD_NAME", "Scania"),
			Channels:  channels,
		},
	}

	return cfg
}

// GetChannelPort returns the port for a specific world and channel
func (c *Config) GetChannelPort(worldID, channelID byte) int {
	for _, world := range c.Worlds {
		if world.WorldID == worldID {
			for _, ch := range world.Channels {
				if ch.ChannelID == channelID {
					return ch.Port
				}
			}
		}
	}
	return 0
}

// GetWorld returns the world config for a given world ID
func (c *Config) GetWorld(worldID byte) (*WorldConfig, bool) {
	for i := range c.Worlds {
		if c.Worlds[i].WorldID == worldID {
			return &c.Worlds[i], true
		}
	}
	return nil, false
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

func getEnvFloat(key string, fallback float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return fallback
}
