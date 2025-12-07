package main

import (
	"fmt"
	"log"
	"time"

	"pdrest/internal/config"
	"pdrest/internal/data"
	"pdrest/internal/database"
	"pdrest/internal/interfaces/http"
	"pdrest/internal/interfaces/services"

	"github.com/labstack/echo/v4"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create Echo instance
	e := echo.New()

	// Setup WAF middleware
	wafConfig := &http.WAFConfig{
		Active:              cfg.WAF.Active,
		RequireSessionID:    cfg.WAF.RequireSessionID,
		SessionIDHeader:     cfg.WAF.SessionIDHeader,
		SessionIDCookie:     cfg.WAF.SessionIDCookie,
		BanOnMissingSession: cfg.WAF.BanOnMissingSession,
		BanTTL:              time.Duration(cfg.WAF.BanTTLHours) * time.Hour,
		WhitelistedPaths:    cfg.WAF.GetWhitelistedPaths(),
		IPBanService:        http.NewIPBanService(time.Duration(cfg.WAF.BanTTLHours) * time.Hour),
	}

	// Apply WAF middleware globally (will be bypassed if Active is false)
	e.Use(http.WAFMiddleware(wafConfig))

	// Log WAF status
	if cfg.WAF.Active {
		log.Println("WAF is ACTIVE - IP banning and session validation enabled")
	} else {
		log.Println("WAF is DISABLED - running in development mode")
	}

	var repo data.UserRepository
	var userService *services.UserService
	var eventService *services.EventService
	var rouletteService *services.RouletteService
	var betService *services.BetService
	authService := services.NewAuthService(cfg.JWT.SecretKey, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)

	// Create Google auth service
	googleAuthService, err := services.NewGoogleAuthService()
	if err != nil {
		log.Printf("Warning: Failed to create Google auth service: %v", err)
		log.Println("Google token verification will be unavailable")
		googleAuthService = nil
	}

	// Create Telegram auth service
	telegramAuthService := services.NewTelegramAuthService(cfg.Telegram.BotToken)
	if cfg.Telegram.BotToken == "" {
		log.Println("Warning: Telegram bot token not configured, hash verification will be disabled")
	}

	// Try to connect to PostgreSQL database
	db, err := database.New(cfg.GetDatabaseURL(), cfg.Database.MaxConns)
	if err != nil {
		log.Printf("Warning: Failed to connect to PostgreSQL: %v", err)
		log.Println("Falling back to in-memory repository")
		repo = data.NewInMemoryUserRepository()
		userService = services.NewUserService(repo)
		// Event, roulette, and bet services require database - will return error if accessed
		eventService = nil
		rouletteService = nil
		betService = nil
	} else {
		defer db.Close()
		log.Println("Successfully connected to PostgreSQL database")

		// Create PostgreSQL repositories
		postgresRepo := data.NewPostgresUserRepository(db.Pool)
		eventRepo := data.NewPostgresEventRepository(db.Pool)
		rouletteRepo := data.NewPostgresRouletteRepository(db.Pool)
		betRepo := data.NewPostgresBetRepository(db.Pool)

		repo = postgresRepo

		// Create services
		userService = services.NewUserService(repo)
		eventService = services.NewEventService(eventRepo)
		rouletteService = services.NewRouletteService(rouletteRepo, repo)
		priceProvider := services.NewPriceProvider("") // Uses Binance API by default
		betService = services.NewBetService(betRepo, priceProvider)
	}

	// Register HTTP handlers (eventService, rouletteService, and betService may be nil if database unavailable)
	http.NewHTTPHandler(e, userService, eventService, rouletteService, betService, authService, googleAuthService, telegramAuthService, cfg.JWT.SecretKey, cfg.JWT.StrictMode)

	// Start server
	addr := cfg.GetAddress()
	fmt.Printf("Server starting on %s\n", addr)
	if err := e.Start(addr); err != nil {
		log.Fatal(err)
	}
}
