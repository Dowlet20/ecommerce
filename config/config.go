package config

import (
    "fmt"
    "os"
)

// Config holds application configuration
type Config struct {
    DBUser     string
    DBPassword string
    DBName     string
    Redis      string
    JWTSecret  string
    ServerAddr string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
    cfg := &Config{
        DBUser:     os.Getenv("DB_USER"),
        DBPassword: os.Getenv("DB_PASSWORD"),
        DBName:     os.Getenv("DB_NAME"),
        Redis:      os.Getenv("REDIS_ADDR"),
        JWTSecret:  os.Getenv("JWT_SECRET"),
        ServerAddr: os.Getenv("SERVER_ADDR"),
    }

    // Validate required fields
    if cfg.DBUser == "" || cfg.DBName == "" {
        return nil, fmt.Errorf("DB_USER and DB_NAME are required")
    }
    if cfg.JWTSecret == "" {
        return nil, fmt.Errorf("JWT_SECRET is required")
    }
    if len(cfg.JWTSecret) < 32 {
        return nil, fmt.Errorf("JWT_SECRET must be at least 32 bytes")
    }
    if cfg.ServerAddr == "" {
        cfg.ServerAddr = ":8080" // Default port
    }

    return cfg, nil
}