package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	MaxConns int
}

func Load() *Config {
	// Try to load .env file from multiple locations
	envPaths := []string{
		".env",                      // Current directory
		"pdrest/.env",               // pdrest subdirectory
		"../.env",                   // Parent directory (project root)
		filepath.Join("..", ".env"), // Parent directory (alternative)
	}

	loaded := false
	for _, envPath := range envPaths {
		if err := godotenv.Load(envPath); err == nil {
			log.Printf("Loaded .env file from: %s", envPath)
			loaded = true
			break
		}
	}

	if !loaded {
		log.Println("Warning: .env file not found, using environment variables and defaults")
	}

	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", ""),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "pumpdump_db"),
			MaxConns: getEnvAsInt("DB_MAX_CONNS", 25),
		},
	}
}

func (c *Config) GetAddress() string {
	if c.Server.Host != "" {
		return c.Server.Host + ":" + c.Server.Port
	}
	return ":" + c.Server.Port
}

func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.DBName,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
