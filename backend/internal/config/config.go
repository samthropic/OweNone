package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Address          string
	DatabaseURL      string
	FrontendOrigin   string
	AvatarDir        string
	MaxDBConnections int32
	AutoMigrate      bool
	ShutdownTimeout  time.Duration
}

func Load() (Config, error) {
	config := Config{
		Address:          envOrDefault("HTTP_ADDRESS", ":8080"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		FrontendOrigin:   envOrDefault("FRONTEND_ORIGIN", "http://localhost:3000"),
		AvatarDir:        envOrDefault("AVATAR_DIR", "data/avatars"),
		MaxDBConnections: 10,
		AutoMigrate:      true,
		ShutdownTimeout:  10 * time.Second,
	}
	if config.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if value := os.Getenv("DATABASE_MAX_CONNECTIONS"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed < 1 {
			return Config{}, fmt.Errorf("DATABASE_MAX_CONNECTIONS must be a positive integer")
		}
		config.MaxDBConnections = int32(parsed)
	}
	if value := os.Getenv("AUTO_MIGRATE"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return Config{}, fmt.Errorf("AUTO_MIGRATE must be true or false")
		}
		config.AutoMigrate = parsed
	}
	return config, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
