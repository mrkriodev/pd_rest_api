package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"pdrest/internal/domain"
	"pdrest/internal/interfaces/services"

	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
)

type HTTPHandler struct {
	userService         *services.UserService
	eventService        *services.EventService
	rouletteService     *services.RouletteService
	betService          *services.BetService
	authService         *services.AuthService
	googleAuthService   *services.GoogleAuthService
	telegramAuthService *services.TelegramAuthService
	jwtSecretKey        string
	jwtStrictMode       bool
}

func NewHTTPHandler(e *echo.Echo, userService *services.UserService, eventService *services.EventService, rouletteService *services.RouletteService, betService *services.BetService, authService *services.AuthService, googleAuthService *services.GoogleAuthService, telegramAuthService *services.TelegramAuthService, jwtSecretKey string, jwtStrictMode bool) {
	h := &HTTPHandler{
		userService:         userService,
		eventService:        eventService,
		rouletteService:     rouletteService,
		betService:          betService,
		authService:         authService,
		googleAuthService:   googleAuthService,
		telegramAuthService: telegramAuthService,
		jwtSecretKey:        jwtSecretKey,
		jwtStrictMode:       jwtStrictMode,
	}

	api := e.Group("/api")
	api.GET("/status", h.Status)
	api.GET("/available_events", h.AvailableEvents)

	// Documentation endpoints
	api.GET("/docs", h.GetAPIDocumentation)
	api.GET("/docs/openapi.yaml", h.GetOpenAPISpec)
	api.GET("/docs/openapi.json", h.GetOpenAPISpecJSON)
	api.GET("/swagger/*", h.SwaggerUI)

	// Auth endpoints
	auth := api.Group("/auth")
	auth.POST("/refresh", h.RefreshToken)
	auth.GET("/status", h.AuthStatus, JWTMiddleware(jwtSecretKey, jwtStrictMode))
	googleAuth := auth.Group("/google")
	googleAuth.GET("/verify", h.VerifyGoogleToken)
	telegramAuth := auth.Group("/telegram")
	telegramAuth.GET("/verify", h.VerifyTelegramToken)

	// User endpoints (protected by JWT)
	user := api.Group("/user")
	user.Use(JWTMiddleware(jwtSecretKey, jwtStrictMode))
	user.GET("/last_login/:uuid", h.UserLastLogin)
	user.GET("/profile/:uuid", h.UserProfile)
	user.POST("/openbet", h.OpenBet)
	user.GET("/betstatus", h.BetStatus)

	// Roulette endpoints
	roulette := api.Group("/roulette")
	roulette.GET("/status", h.GetRouletteStatus)
	roulette.POST("/spin", h.Spin)
	roulette.POST("/take-prize", h.TakePrize)
	roulette.POST("/preauth-token", h.CreatePreauthToken)
}

func (h *HTTPHandler) Status(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HTTPHandler) GetAPIDocumentation(c echo.Context) error {
	// Try to read the API documentation markdown file
	docPath := filepath.Join("docs", "API_DOCUMENTATION.md")
	content, err := os.ReadFile(docPath)
	if err != nil {
		// If file not found, return a simple text response with endpoint info
		doc := `PD REST API Documentation

Available endpoints:
- GET /api/status - Health check
- GET /api/available_events - Get available events
- POST /api/auth/refresh - Refresh JWT token
- GET /api/auth/status - Check JWT authorization status
- GET /api/auth/google/verify - Verify Google OAuth token
- GET /api/auth/telegram/verify - Verify Telegram Web Login
- GET /api/user/last_login/:uuid - Get user last login time
- GET /api/user/profile/:uuid - Get user profile
- POST /api/user/openbet - Create a new bet
- GET /api/user/betstatus?id=<bet_id> - Get bet status

For full documentation, see: /api/docs/openapi.yaml or /api/docs/openapi.json
`
		c.Response().Header().Set(echo.HeaderContentType, "text/plain; charset=utf-8")
		return c.String(http.StatusOK, doc)
	}

	c.Response().Header().Set(echo.HeaderContentType, "text/plain; charset=utf-8")
	return c.String(http.StatusOK, string(content))
}

func (h *HTTPHandler) GetOpenAPISpec(c echo.Context) error {
	// Try to read the OpenAPI YAML file
	specPath := filepath.Join("docs", "openapi.yaml")
	content, err := os.ReadFile(specPath)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "OpenAPI specification not found"})
	}

	c.Response().Header().Set(echo.HeaderContentType, "application/x-yaml")
	return c.String(http.StatusOK, string(content))
}

func (h *HTTPHandler) GetOpenAPISpecJSON(c echo.Context) error {
	// Read the OpenAPI YAML file and convert to JSON
	specPath := filepath.Join("docs", "openapi.yaml")
	content, err := os.ReadFile(specPath)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "OpenAPI specification not found"})
	}

	// Parse YAML and convert to JSON
	var spec map[string]interface{}
	if err := yaml.Unmarshal(content, &spec); err != nil {
		// If YAML parsing fails, return error
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to parse OpenAPI specification"})
	}

	return c.JSON(http.StatusOK, spec)
}

func (h *HTTPHandler) SwaggerUI(c echo.Context) error {
	// Serve Swagger UI HTML
	swaggerHTML := `<!DOCTYPE html>
<html>
<head>
	<title>PD REST API - Swagger UI</title>
	<link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
	<style>
		html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
		*, *:before, *:after { box-sizing: inherit; }
		body { margin:0; background: #fafafa; }
	</style>
</head>
<body>
	<div id="swagger-ui"></div>
	<script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
	<script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
	<script>
		window.onload = function() {
			const ui = SwaggerUIBundle({
				url: "/api/docs/openapi.yaml",
				dom_id: '#swagger-ui',
				deepLinking: true,
				presets: [
					SwaggerUIBundle.presets.apis,
					SwaggerUIStandalonePreset
				],
				plugins: [
					SwaggerUIBundle.plugins.DownloadUrl
				],
				layout: "StandaloneLayout"
			});
		};
	</script>
</body>
</html>`

	c.Response().Header().Set(echo.HeaderContentType, "text/html; charset=utf-8")
	return c.HTML(http.StatusOK, swaggerHTML)
}

func (h *HTTPHandler) UserLastLogin(c echo.Context) error {
	uuid := c.Param("uuid")
	if uuid == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "uuid is required"})
	}

	result, err := h.userService.GetLastLogin(uuid)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) UserProfile(c echo.Context) error {
	uuid := c.Param("uuid")
	if uuid == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "uuid is required"})
	}

	result, err := h.userService.GetProfile(uuid)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) OpenBet(c echo.Context) error {
	if h.betService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for bets"})
	}

	// Get user UUID from context (set by JWT middleware)
	userUUID, ok := c.Get("user_uuid").(string)
	if !ok || userUUID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req domain.OpenBetRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	ctx := context.Background()
	response, err := h.betService.OpenBet(ctx, userUUID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "must be") || strings.Contains(err.Error(), "required") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, response)
}

func (h *HTTPHandler) BetStatus(c echo.Context) error {
	if h.betService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for bets"})
	}

	// Get user UUID from context (set by JWT middleware)
	userUUID, ok := c.Get("user_uuid").(string)
	if !ok || userUUID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Get bet ID from query parameter
	betIDStr := c.QueryParam("id")
	if betIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id query parameter is required"})
	}

	var betID int
	if _, err := fmt.Sscanf(betIDStr, "%d", &betID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid bet id"})
	}

	ctx := context.Background()
	response, err := h.betService.GetBetStatus(ctx, betID, userUUID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, response)
}

func (h *HTTPHandler) RefreshToken(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.RefreshToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "refresh_token is required"})
	}

	tokenPair, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, tokenPair)
}

func (h *HTTPHandler) AuthStatus(c echo.Context) error {
	// Get user UUID from context (set by JWT middleware)
	userUUID, ok := c.Get("user_uuid").(string)
	if !ok || userUUID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	return c.JSON(http.StatusOK, map[string]string{"uuid": userUUID})
}

func (h *HTTPHandler) VerifyGoogleToken(c echo.Context) error {
	if h.googleAuthService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Google authentication service unavailable"})
	}

	// Get Google token from Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
	}

	// Extract token from "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid authorization header format"})
	}

	googleToken := parts[1]

	// Validate Google token
	googleUserInfo, err := h.googleAuthService.ValidateWithGoogle(googleToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	// Find user by Google ID
	user, err := h.userService.GetUserByGoogleID(googleUserInfo.ID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user not found"})
	}

	// Generate JWT token pair for the user
	tokenPair, err := h.authService.GenerateTokenPair(user.UUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
	}

	return c.JSON(http.StatusOK, tokenPair)
}

func (h *HTTPHandler) VerifyTelegramToken(c echo.Context) error {
	if h.telegramAuthService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Telegram authentication service unavailable"})
	}

	// Telegram Web Login sends data as query parameters
	var authData services.TelegramAuthData

	// Parse ID
	if idStr := c.QueryParam("id"); idStr != "" {
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			authData.ID = id
		}
	}

	// Parse auth_date
	if authDateStr := c.QueryParam("auth_date"); authDateStr != "" {
		if authDate, err := strconv.ParseInt(authDateStr, 10, 64); err == nil {
			authData.AuthDate = authDate
		}
	}

	authData.FirstName = c.QueryParam("first_name")
	authData.LastName = c.QueryParam("last_name")
	authData.Username = c.QueryParam("username")
	authData.PhotoURL = c.QueryParam("photo_url")
	authData.Hash = c.QueryParam("hash")

	// If no query params, try JSON body
	if authData.ID == 0 && authData.Hash == "" {
		if err := c.Bind(&authData); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body or missing query parameters"})
		}
	}

	// Validate Telegram token
	telegramUserInfo, err := h.telegramAuthService.ValidateWithTelegram(authData)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	// Find user by Telegram ID
	user, err := h.userService.GetUserByTelegramID(telegramUserInfo.ID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user not found"})
	}

	// Generate JWT token pair for the user
	tokenPair, err := h.authService.GenerateTokenPair(user.UUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
	}

	return c.JSON(http.StatusOK, tokenPair)
}

func (h *HTTPHandler) AvailableEvents(c echo.Context) error {
	if h.eventService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for events"})
	}

	ctx := context.Background()
	events, err := h.eventService.GetAvailableEvents(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"events": events,
	})
}

// getUserUUID extracts user UUID from header or query parameter (optional)
func (h *HTTPHandler) getUserUUID(c echo.Context) *string {
	// Try header first
	userUUID := c.Request().Header.Get("X-User-UUID")
	if userUUID != "" {
		return &userUUID
	}

	// Try query parameter
	userUUID = c.QueryParam("user_uuid")
	if userUUID != "" {
		return &userUUID
	}

	return nil
}

// GetRouletteStatus gets the current status of roulette by preauth token
func (h *HTTPHandler) GetRouletteStatus(c echo.Context) error {
	if h.rouletteService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for roulette"})
	}

	// Get preauth token from query parameter or header
	preauthToken := c.QueryParam("preauth_token")
	if preauthToken == "" {
		preauthToken = c.Request().Header.Get("X-Preauth-Token")
	}
	if preauthToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "preauth_token is required (query parameter or X-Preauth-Token header)"})
	}

	ctx := context.Background()
	status, err := h.rouletteService.GetRouletteStatus(ctx, preauthToken)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, status)
}

// Spin performs a spin using preauth token
func (h *HTTPHandler) Spin(c echo.Context) error {
	if h.rouletteService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for roulette"})
	}

	var req domain.SpinRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.PreauthToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "preauth_token is required"})
	}

	ctx := context.Background()
	response, err := h.rouletteService.Spin(ctx, &req)
	if err != nil {
		// Check if it's a business logic error (should return 400) or server error (500)
		if strings.Contains(err.Error(), "invalid") ||
			strings.Contains(err.Error(), "already") ||
			strings.Contains(err.Error(), "maximum") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "expired") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, response)
}

// TakePrize allows user to take the prize after completing all spins
func (h *HTTPHandler) TakePrize(c echo.Context) error {
	if h.rouletteService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for roulette"})
	}

	var req domain.TakePrizeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.PreauthToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "preauth_token is required"})
	}

	// Extract session_id and IP address for internal user registration
	sessionID := c.Request().Header.Get("X-SESSION-ID")
	if sessionID == "" {
		cookie, err := c.Cookie("X-SESSION-ID")
		if err == nil && cookie != nil && cookie.Value != "" {
			sessionID = cookie.Value
		}
	}

	// Get client IP
	ipAddress := c.Request().Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = c.Request().Header.Get("X-Real-IP")
	}
	if ipAddress == "" {
		ipAddress = c.RealIP()
	}

	// Create context with session_id and IP for internal registration
	ctx := context.Background()
	if sessionID != "" {
		ctx = context.WithValue(ctx, "session_id", sessionID)
	}
	if ipAddress != "" {
		ctx = context.WithValue(ctx, "ip_address", ipAddress)
	}

	response, err := h.rouletteService.TakePrize(ctx, &req)
	if err != nil {
		// Check if it's a business logic error (should return 400) or server error (500)
		if strings.Contains(err.Error(), "invalid") ||
			strings.Contains(err.Error(), "already") ||
			strings.Contains(err.Error(), "must complete") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "expired") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, response)
}

// CreatePreauthToken creates a preauth token (typically called from browser)
func (h *HTTPHandler) CreatePreauthToken(c echo.Context) error {
	if h.rouletteService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for roulette"})
	}

	// Get optional user UUID (for authenticated users)
	userUUID := h.getUserUUID(c)

	var req struct {
		Token     string  `json:"token"`
		Type      string  `json:"type"`
		EventID   *string `json:"event_id,omitempty"`
		ExpiresAt int64   `json:"expires_at"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.Token == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "token is required"})
	}

	if req.Type == "" {
		req.Type = "on_start" // Default
	}

	rouletteType := domain.RouletteType(req.Type)
	if rouletteType != domain.RouletteTypeOnStart && rouletteType != domain.RouletteTypeDuringEvent {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid type, must be 'on_start' or 'during_event'"})
	}

	if rouletteType == domain.RouletteTypeDuringEvent && req.EventID == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "event_id is required for during_event type"})
	}

	if req.ExpiresAt == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "expires_at is required"})
	}

	ctx := context.Background()
	err := h.rouletteService.CreatePreauthToken(ctx, rouletteType, req.EventID, req.Token, req.ExpiresAt, userUUID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "inactive") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Preauth token created successfully",
	})
}
