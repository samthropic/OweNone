package config

import (
	"fmt"
	"os"
	"path/filepath"
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
		Address:          listenAddress(),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		FrontendOrigin:   frontendOrigin(),
		AvatarDir:        avatarDir(),
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

// listenAddress prefers HTTP_ADDRESS, then Vercel's PORT, then :8080.
func listenAddress() string {
	if value := os.Getenv("HTTP_ADDRESS"); value != "" {
		return value
	}
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return ":8080"
}

// frontendOrigin prefers FRONTEND_ORIGIN, then https://VERCEL_URL, then localhost.
func frontendOrigin() string {
	if value := os.Getenv("FRONTEND_ORIGIN"); value != "" {
		return value
	}
	if host := os.Getenv("VERCEL_URL"); host != "" {
		return "https://" + host
	}
	return "http://localhost:3000"
}

func avatarDir() string {
	if value := os.Getenv("AVATAR_DIR"); value != "" {
		return value
	}
	// Vercel’s app filesystem is read-only; keep uploads under /tmp.
	if os.Getenv("VERCEL") != "" {
		return filepath.Join(os.TempDir(), "owenone-avatars")
	}
	return "data/avatars"
}
