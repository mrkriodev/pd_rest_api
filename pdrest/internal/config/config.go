package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	WAF      WAFConfig
	JWT      JWTConfig
	Telegram TelegramConfig
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
}

type WAFConfig struct {
	Active              bool // Enable/disable WAF (default: false)
	RequireSessionID    bool
	SessionIDHeader     string
	SessionIDCookie     string
	BanOnMissingSession bool
	BanTTLHours         int
	WhitelistedPaths    string // Comma-separated list of paths
}

type JWTConfig struct {
	SecretKey       string
	AccessTokenTTL  int  // in hours
	RefreshTokenTTL int  // in hours
	StrictMode      bool // if false, only check token is non-empty
}

type TelegramConfig struct {
	BotToken string
}

func Load() *Config {
	// Configuration loading order:
	// 1. First, try to load .env file (loads values that don't exist in environment)
	// 2. Then, environment variables override any values from .env file
	// This ensures environment variables take precedence over .env file values

	// Try to load .env file from multiple locations
	envPaths := []string{
		".env",                      // Current directory
		"pdrest/.env",               // pdrest subdirectory
		"../.env",                   // Parent directory (project root)
		filepath.Join("..", ".env"), // Parent directory (alternative)
	}

	loaded := false
	for _, envPath := range envPaths {
		// godotenv.Load() does NOT override existing environment variables
		// So: .env file sets values for vars that don't exist in environment
		//     Environment variables that exist will remain (overriding file values)
		if err := godotenv.Load(envPath); err == nil {
			log.Printf("Loaded .env file from: %s (environment variables will override .env values)", envPath)
			loaded = true
			break
		}
	}

	if !loaded {
		log.Println("Warning: .env file not found, using environment variables and defaults")
	}

	// Parse whitelisted paths
	whitelistedPathsStr := getEnv("WAF_WHITELISTED_PATHS", "/api/status")

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
		},
		WAF: WAFConfig{
			Active:              getEnvAsBool("WAF_ACTIVE", false),
			RequireSessionID:    getEnvAsBool("WAF_REQUIRE_SESSION_ID", true),
			SessionIDHeader:     getEnv("WAF_SESSION_ID_HEADER", "X-SESSION-ID"),
			SessionIDCookie:     getEnv("WAF_SESSION_ID_COOKIE", "X-SESSION-ID"),
			BanOnMissingSession: getEnvAsBool("WAF_BAN_ON_MISSING_SESSION", true),
			BanTTLHours:         getEnvAsInt("WAF_BAN_TTL_HOURS", 24),
			WhitelistedPaths:    whitelistedPathsStr,
		},
		JWT: JWTConfig{
			SecretKey:       getEnv("JWT_SECRET_KEY", "your-secret-key-change-in-production"),
			AccessTokenTTL:  getEnvAsInt("JWT_ACCESS_TOKEN_TTL_HOURS", 1),
			RefreshTokenTTL: getEnvAsInt("JWT_REFRESH_TOKEN_TTL_HOURS", 24),
			StrictMode:      getEnvAsBool("JWT_STRICT_MODE", true),
		},
		Telegram: TelegramConfig{
			BotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
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

// GetWhitelistedPaths returns a slice of whitelisted paths
func (c *WAFConfig) GetWhitelistedPaths() []string {
	if c.WhitelistedPaths == "" {
		return []string{"/api/status"}
	}

	paths := []string{}
	// Simple split by comma, trim whitespace
	for _, path := range splitAndTrim(c.WhitelistedPaths, ",") {
		if path != "" {
			paths = append(paths, path)
		}
	}
	return paths
}

func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// getEnv retrieves environment variable value
// Priority: environment variable > .env file > default value
// (Environment variables override .env file values because godotenv.Load() doesn't override existing env vars)
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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
