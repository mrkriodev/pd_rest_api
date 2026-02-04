package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pdrest/internal/config"
	"pdrest/internal/data"
	"pdrest/internal/database"
	"pdrest/internal/interfaces/http"
	"pdrest/internal/interfaces/services"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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
	var ratingService *services.RatingService
	var eventService *services.EventService
	var rouletteService *services.RouletteService
	var betService *services.BetService
	var betScheduler *services.BetScheduler
	var achievementService *services.AchievementService
	authService := services.NewAuthService(cfg.JWT.SecretKey, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)

	// Create Google auth service
	var googleAuthService *services.GoogleAuthService
	if cfg.Google.ClientID != "" {
		var err error
		googleAuthService, err = services.NewGoogleAuthService(cfg.Google.ClientID)
		if err != nil {
			log.Printf("Warning: Failed to create Google auth service: %v", err)
			log.Println("Google token verification will be unavailable")
			googleAuthService = nil
		} else {
			log.Println("Google authentication service initialized successfully")
		}
	} else {
		log.Println("Warning: GOOGLE_CLIENT_ID not configured, Google token verification will be unavailable")
		googleAuthService = nil
	}

	var googleOAuthConfig *oauth2.Config
	if cfg.Google.ClientID != "" && cfg.Google.ClientSecret != "" {
		googleOAuthConfig = &oauth2.Config{
			ClientID:     cfg.Google.ClientID,
			ClientSecret: cfg.Google.ClientSecret,
			RedirectURL:  cfg.Google.RedirectURL,
			Endpoint:     google.Endpoint,
			Scopes:       []string{"openid", "email", "profile"},
		}
	} else {
		log.Println("Warning: GOOGLE_CLIENT_ID/GOOGLE_CLIENT_SECRET not configured, Google OAuth2 code exchange will be unavailable")
	}

	// Create Telegram auth service
	telegramAuthService := services.NewTelegramAuthService(cfg.Telegram.BotToken)
	if cfg.Telegram.BotToken == "" {
		log.Println("Warning: Telegram bot token not configured, hash verification will be disabled")
	}

	// Try to connect to PostgreSQL database
	db, err := database.New(cfg.GetDatabaseURL())
	if err != nil {
		log.Printf("Warning: Failed to connect to PostgreSQL: %v", err)
		log.Println("Falling back to in-memory repository")
		repo = data.NewInMemoryUserRepository()
		userService = services.NewUserService(repo)
		ratingService = nil
		// Event, roulette, bet, and achievement services require database - will return error if accessed
		eventService = nil
		rouletteService = nil
		betService = nil
		achievementService = nil
	} else {
		defer db.Close()
		log.Println("Successfully connected to PostgreSQL database")

		// Create PostgreSQL repositories
		postgresRepo := data.NewPostgresUserRepository(db.Pool)
		eventRepo := data.NewPostgresEventRepository(db.Pool)
		rouletteRepo := data.NewPostgresRouletteRepository(db.Pool)
		betRepo := data.NewPostgresBetRepository(db.Pool)
		ratingRepo := data.NewPostgresRatingRepository(db.Pool)
		achievementRepo := data.NewPostgresAchievementRepository(db.Pool)
		prizeRepo := data.NewPostgresPrizeRepository(db.Pool)
		prizeValueRepo := data.NewPostgresPrizeValueRepository(db.Pool)

		repo = postgresRepo

		// Create services
		userService = services.NewUserService(repo)
		ratingService = services.NewRatingService(ratingRepo, prizeRepo, betRepo)
		eventService = services.NewEventService(eventRepo)
		rouletteService = services.NewRouletteService(rouletteRepo, repo, prizeRepo, prizeValueRepo, eventRepo)
		priceProvider := services.NewPriceProvider("") // Uses Binance API by default
		betScheduler = services.NewBetScheduler(betRepo, priceProvider)
		betService = services.NewBetService(betRepo, priceProvider, betScheduler)
		achievementService = services.NewAchievementService(achievementRepo)
	}

	// Register HTTP handlers (eventService, rouletteService, betService, and achievementService may be nil if database unavailable)
	http.NewHTTPHandler(e, userService, ratingService, eventService, rouletteService, betService, achievementService, authService, googleAuthService, googleOAuthConfig, telegramAuthService, cfg.JWT.SecretKey, cfg.JWT.StrictMode)

	// Start server in a goroutine
	addr := cfg.GetAddress()
	fmt.Printf("Server starting on %s\n", addr)

	go func() {
		if err := e.Start(addr); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Gracefully shutdown bet scheduler first
	if betScheduler != nil {
		betScheduler.Shutdown()
	}

	// Gracefully shutdown Echo server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server exited")
}
