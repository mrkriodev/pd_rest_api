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

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
)

type HTTPHandler struct {
	userService         *services.UserService
	ratingService       *services.RatingService
	eventService        *services.EventService
	rouletteService     *services.RouletteService
	betService          *services.BetService
	achievementService  *services.AchievementService
	authService         *services.AuthService
	googleAuthService   *services.GoogleAuthService
	telegramAuthService *services.TelegramAuthService
	jwtSecretKey        string
	jwtStrictMode       bool
}

func NewHTTPHandler(e *echo.Echo, userService *services.UserService, ratingService *services.RatingService, eventService *services.EventService, rouletteService *services.RouletteService, betService *services.BetService, achievementService *services.AchievementService, authService *services.AuthService, googleAuthService *services.GoogleAuthService, telegramAuthService *services.TelegramAuthService, jwtSecretKey string, jwtStrictMode bool) {
	h := &HTTPHandler{
		userService:         userService,
		ratingService:       ratingService,
		eventService:        eventService,
		rouletteService:     rouletteService,
		betService:          betService,
		achievementService:  achievementService,
		authService:         authService,
		googleAuthService:   googleAuthService,
		telegramAuthService: telegramAuthService,
		jwtSecretKey:        jwtSecretKey,
		jwtStrictMode:       jwtStrictMode,
	}

	api := e.Group("/api")
	api.GET("/status", h.Status)
	api.GET("/available_events", h.AvailableEvents)
	api.GET("/available_achievements", h.AvailableAchievements)
	api.GET("/globalrating", h.GlobalRating)
	api.GET("/getidbypreauth", h.GetUserIDByPreauth)
	api.GET("/getidbysession", h.GetUserIDBySession)

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
	googleAuth.POST("/register", h.RegisterGoogleUser)
	telegramAuth := auth.Group("/telegram")
	telegramAuth.GET("/verify", h.VerifyTelegramToken)

	// User endpoints (protected by JWT)
	user := api.Group("/user")
	user.Use(JWTMiddleware(jwtSecretKey, jwtStrictMode))
	user.GET("/last_login/:uuid", h.UserLastLogin)
	user.GET("/profile/:uuid", h.UserProfile)
	user.GET("/assets", h.UserAssets)
	user.POST("/assets", h.UserAssets)
	user.GET("/referral_link", h.UserReferralLink)
	user.GET("/friends_ratings", h.UserFriendsRatings)
	user.GET("/achievements", h.UserAchievements)
	user.POST("/openbet", h.OpenBet)
	user.GET("/betstatus", h.BetStatus)
	user.GET("/unfinished_bets/:uuid", h.UnfinishedBets)

	// Roulette endpoints
	roulette := api.Group("/roulette")
	roulette.GET("/status", h.GetRouletteStatus)
	roulette.GET("/get", h.GetRouletteConfig)
	roulette.POST("/spin", h.Spin)
	roulette.POST("/take-prize", h.TakePrize)
	roulette.GET("/get_preauth_token", h.GetPreauthToken)
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
- GET /api/user/unfinished_bets/:uuid - Get unfinished bets for user
- GET /api/getidbysession - Get user ID by session_id + IP

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

func (h *HTTPHandler) UserAssets(c echo.Context) error {
	if h.ratingService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for user assets"})
	}

	var req struct {
		UserID string `json:"userId" query:"userId"`
	}

	// Try to bind userId from body or query params
	_ = c.Bind(&req)

	// Also check query param if not in body
	if req.UserID == "" {
		req.UserID = c.QueryParam("userId")
	}

	var userUUID string

	// In strict mode, validate JWT token from Authorization header and match with userId
	if h.jwtStrictMode {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid authorization header format"})
		}

		tokenString := parts[1]
		if tokenString == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "token cannot be empty"})
		}

		// Validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(h.jwtSecretKey), nil
		})

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		// Extract user UUID from token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
		}

		tokenUserUUID, ok := claims["uuid"].(string)
		if !ok || tokenUserUUID == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token: missing user UUID"})
		}

		// If userId is provided in body, verify it matches the token
		if req.UserID != "" {
			if req.UserID != tokenUserUUID {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "userId does not match token"})
			}
			userUUID = req.UserID
		} else {
			userUUID = tokenUserUUID
		}
	} else {
		// Non-strict mode: get userId from body
		if req.UserID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "userId is required in request body"})
		}
		userUUID = req.UserID
	}

	assets, err := h.ratingService.GetUserAssets(c.Request().Context(), userUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Return response with userId (not userID)
	response := map[string]interface{}{
		"userId":       assets.UserID,
		"points":       assets.Points,
		"total_points": assets.TotalPoints,
	}

	return c.JSON(http.StatusOK, response)
}

// GetUserIDByPreauth returns user UUID (as userId) for a valid preauth token
// Requires X-SESSION-ID header that must correspond to the preauth token
func (h *HTTPHandler) GetUserIDByPreauth(c echo.Context) error {
	if h.rouletteService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required"})
	}

	// Get preauth_token from header or query parameter
	preauthToken := c.Request().Header.Get("X-Preauth-Token")
	if preauthToken == "" {
		preauthToken = c.QueryParam("preauth_token")
	}
	if preauthToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "preauth_token is required (header X-Preauth-Token or query parameter)"})
	}

	// Get X-SESSION-ID header (mandatory)
	sessionID := c.Request().Header.Get("X-SESSION-ID")
	if sessionID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "X-SESSION-ID header is required"})
	}

	// Get client IP
	ipAddress := c.Request().Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = c.Request().Header.Get("X-Real-IP")
	}
	if ipAddress == "" {
		ipAddress = c.RealIP()
	}
	if ipAddress == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "could not determine IP address"})
	}

	ctx := context.Background()
	userID, err := h.rouletteService.GetUserIDByPreauthToken(ctx, preauthToken, sessionID, ipAddress)
	if err != nil {
		if strings.Contains(err.Error(), "does not match") || strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not linked") {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"userId": userID})
}

// GetUserIDBySession returns user UUID (as userId) for a valid session_id + IP
// Requires X-SESSION-ID header and derives preauth token from session_id + IP
func (h *HTTPHandler) GetUserIDBySession(c echo.Context) error {
	if h.rouletteService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required"})
	}

	// Get X-SESSION-ID header (mandatory)
	sessionID := c.Request().Header.Get("X-SESSION-ID")
	if sessionID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "X-SESSION-ID header is required"})
	}

	// Get client IP
	ipAddress := c.Request().Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = c.Request().Header.Get("X-Real-IP")
	}
	if ipAddress == "" {
		ipAddress = c.RealIP()
	}
	if ipAddress == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "could not determine IP address"})
	}

	ctx := context.Background()
	userID, err := h.rouletteService.GetUserIDBySessionIP(ctx, sessionID, ipAddress)
	if err != nil {
		if strings.Contains(err.Error(), "does not match") || strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not linked") {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"userId": userID})
}

func (h *HTTPHandler) UserReferralLink(c echo.Context) error {
	if h.userService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for referral link"})
	}

	userUUID, ok := c.Get("user_uuid").(string)
	if !ok || userUUID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	scheme := "http"
	if c.IsTLS() {
		scheme = "https"
	}
	host := c.Request().Host
	referralLink := fmt.Sprintf("%s://%s/ref/%s", scheme, host, userUUID)

	return c.JSON(http.StatusOK, map[string]string{
		"referral_link": referralLink,
		"code":          userUUID,
	})
}

func (h *HTTPHandler) UserFriendsRatings(c echo.Context) error {
	if h.ratingService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for friends ratings"})
	}

	// Get user UUID from context (set by JWT middleware)
	userUUID, ok := c.Get("user_uuid").(string)
	if !ok || userUUID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Parse pagination parameters
	limit := 50 // Default limit
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // Default offset
	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	result, err := h.ratingService.GetFriendsRatings(c.Request().Context(), userUUID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"friends": result,
	})
}

func (h *HTTPHandler) GlobalRating(c echo.Context) error {
	if h.ratingService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for global rating"})
	}

	// Parse pagination parameters
	limit := 50 // Default limit
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // Default offset
	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	entries, err := h.ratingService.GetGlobalRating(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, entries)
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

func (h *HTTPHandler) UnfinishedBets(c echo.Context) error {
	if h.betService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for bets"})
	}

	userUUID := c.Param("uuid")
	if userUUID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "uuid is required"})
	}

	ctx := context.Background()
	bets, err := h.betService.GetUnfinishedBetsByUser(ctx, userUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"bets": bets,
	})
}

func (h *HTTPHandler) RefreshToken(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
		UserID       string `json:"userID"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// If JWT_STRICT_MODE=true, check Authorization header and extract userID from token
	if h.jwtStrictMode {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid authorization header format"})
		}

		tokenString := parts[1]
		if tokenString == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "token cannot be empty"})
		}

		// Validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(h.jwtSecretKey), nil
		})

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		// Extract user UUID from token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
		}

		userUUID, ok := claims["uuid"].(string)
		if !ok || userUUID == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token: missing user UUID"})
		}

		// Generate new token pair for the user
		tokenPair, err := h.authService.GenerateTokenPair(userUUID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		}

		return c.JSON(http.StatusOK, tokenPair)
	}

	// Non-strict mode: if userID is provided, generate new tokens
	if req.UserID != "" {
		tokenPair, err := h.authService.GenerateTokenPair(req.UserID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		}
		return c.JSON(http.StatusOK, tokenPair)
	}

	// Fallback: use refresh_token if provided
	if req.RefreshToken != "" {
		tokenPair, err := h.authService.RefreshToken(req.RefreshToken)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, tokenPair)
	}

	return c.JSON(http.StatusBadRequest, map[string]string{"error": "userID or refresh_token is required"})
}

func (h *HTTPHandler) AuthStatus(c echo.Context) error {
	// Get user UUID from context (set by JWT middleware)
	userUUID, ok := c.Get("user_uuid").(string)
	if !ok || userUUID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	return c.JSON(http.StatusOK, map[string]string{"userID": userUUID})
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
	tokenPair, err := h.authService.GenerateTokenPair(user.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
	}

	return c.JSON(http.StatusOK, tokenPair)
}

// RegisterGoogleUser registers a new user with Google OAuth information
// This endpoint is called when a user first registers with Google OAuth
func (h *HTTPHandler) RegisterGoogleUser(c echo.Context) error {
	if h.googleAuthService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Google authentication service unavailable"})
	}

	if h.userService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "User service unavailable"})
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

	// Parse request body
	var req struct {
		UserID       string `json:"userID"`
		PreauthToken string `json:"preauth_token"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.UserID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "userID is required"})
	}

	if req.PreauthToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "preauth_token is required"})
	}

	// Check if Google ID is already registered to a different user
	existingUser, err := h.userService.GetUserByGoogleID(googleUserInfo.ID)
	if err == nil && existingUser != nil {
		// User exists with this Google ID
		if existingUser.UserID != req.UserID {
			// Google ID is registered to a different user - conflict
			return c.JSON(http.StatusConflict, map[string]string{"error": "Google account is already registered to another user"})
		}
		// Same user - allow update (will be handled by CreateOrUpdateUserWithGoogleInfo)
	} else if err != nil && !strings.Contains(err.Error(), "not found") {
		// Database error (not "user not found") - return error
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to check existing user: " + err.Error()})
	}
	// If user not found (err != nil with "not found" message) or err == nil with existingUser == nil, proceed with registration

	// Register user with Google info
	ctx := context.Background()
	if err := h.userService.RegisterUserWithGoogle(ctx, req.UserID, googleUserInfo.ID, googleUserInfo.Email, googleUserInfo.Name); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Link preauth token to user
	if h.rouletteService != nil {
		if err := h.rouletteService.LinkPreauthTokenToUser(ctx, req.PreauthToken, req.UserID); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to link preauth token: " + err.Error()})
		}
	}

	// Generate JWT token pair for the user
	tokenPair, err := h.authService.GenerateTokenPair(req.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
	}

	return c.JSON(http.StatusOK, tokenPair)
}

func (h *HTTPHandler) VerifyTelegramToken(c echo.Context) error {
	if h.telegramAuthService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Telegram authentication service unavailable"})
	}

	// Get optional preauth_token from header or query parameter
	preauthToken := c.Request().Header.Get("X-Preauth-Token")
	if preauthToken == "" {
		preauthToken = c.QueryParam("preauth_token")
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

	// Link preauth token to user if provided
	if preauthToken != "" && h.rouletteService != nil {
		ctx := context.Background()
		if err := h.rouletteService.LinkPreauthTokenToUser(ctx, preauthToken, user.UserID); err != nil {
			// Log error but don't fail the auth request
			_ = err
		}
	}

	// Generate JWT token pair for the user
	tokenPair, err := h.authService.GenerateTokenPair(user.UserID)
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

func (h *HTTPHandler) AvailableAchievements(c echo.Context) error {
	if h.achievementService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for achievements"})
	}

	ctx := context.Background()
	result, err := h.achievementService.GetAvailableAchievements(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) UserAchievements(c echo.Context) error {
	if h.achievementService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for user achievements"})
	}

	// Get user UUID from context (set by JWT middleware)
	userUUID, ok := c.Get("user_uuid").(string)
	if !ok || userUUID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	ctx := context.Background()
	result, err := h.achievementService.GetUserAchievements(ctx, userUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

// GetRouletteStatus gets the current status of roulette
// If roulette_id = 1: checks X-SESSION-ID header
// If roulette_id != 1: checks JWT Header
func (h *HTTPHandler) GetRouletteStatus(c echo.Context) error {
	if h.rouletteService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for roulette"})
	}

	// Get roulette_id from query parameter
	rouletteIDStr := c.QueryParam("roulette_id")
	if rouletteIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "roulette_id query parameter is required"})
	}

	rouletteID, err := strconv.Atoi(rouletteIDStr)
	if err != nil || rouletteID <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid roulette_id"})
	}

	var preauthToken string

	// Authentication based on roulette_id
	if rouletteID == 1 {
		// For roulette_id = 1, check X-SESSION-ID header
		sessionID := c.Request().Header.Get("X-SESSION-ID")
		if sessionID == "" {
			cookie, err := c.Cookie("X-SESSION-ID")
			if err == nil && cookie != nil && cookie.Value != "" {
				sessionID = cookie.Value
			}
		}
		if sessionID == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "X-SESSION-ID header is required for roulette_id=1"})
		}

		// Get client IP
		ipAddress := c.Request().Header.Get("X-Forwarded-For")
		if ipAddress == "" {
			ipAddress = c.Request().Header.Get("X-Real-IP")
		}
		if ipAddress == "" {
			ipAddress = c.RealIP()
		}
		if ipAddress == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "could not determine IP address"})
		}

		// Generate preauth token from session_id + IP (same logic as GetPreauthToken)
		ctx := context.Background()
		preauthToken, err = h.rouletteService.GetPreauthToken(ctx, sessionID, ipAddress)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	} else {
		// For roulette_id != 1, check JWT Header only
		userID, err := h.validateJWTToken(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authorization required for this roulette: " + err.Error()})
		}
		// Store userID in context for potential future use
		c.Set("userID", userID)

		// Get roulette status by user UUID and roulette config ID
		ctx := context.Background()
		status, err := h.rouletteService.GetRouletteStatusByUser(ctx, userID, rouletteID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, status)
	}

	// For roulette_id = 1, use preauth token
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

// validateJWTToken validates JWT token from Authorization header and returns userID if valid
func (h *HTTPHandler) validateJWTToken(c echo.Context) (string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	// Extract token from "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	tokenString := parts[1]

	// Parse and validate token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.jwtSecretKey), nil
	})

	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	// Extract user UUID
	userID, ok := claims["uuid"].(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("user ID not found in token")
	}

	return userID, nil
}

// GetRouletteConfig returns roulette config by id
func (h *HTTPHandler) GetRouletteConfig(c echo.Context) error {
	if h.rouletteService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for roulette"})
	}

	idStr := c.QueryParam("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "id query parameter is required"})
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	// JWT is required for roulette id != 1
	if id != 1 {
		userID, err := h.validateJWTToken(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authorization required for this roulette: " + err.Error()})
		}
		// Store userID in context for potential future use
		c.Set("userID", userID)
	}

	ctx := context.Background()
	config, err := h.rouletteService.GetRouletteConfigByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if config == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "roulette config not found"})
	}

	return c.JSON(http.StatusOK, config)
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

	if req.RouletteID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "roulette_id is required"})
	}

	// JWT is required for roulette id != 1
	if req.RouletteID != 1 {
		userID, err := h.validateJWTToken(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authorization required for this roulette: " + err.Error()})
		}
		// Store userID in context for potential future use
		c.Set("userID", userID)
	}

	// Get preauth_token from header, query, or body (optional)
	// Priority: header > query > body
	preauthToken := c.Request().Header.Get("X-Preauth-Token")
	if preauthToken == "" {
		preauthToken = c.QueryParam("preauth_token")
	}
	if preauthToken == "" {
		preauthToken = req.PreauthToken
	}

	// Extract session_id and IP address (required if preauth_token is not provided)
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

	ctx := context.Background()
	// Pass Authorization header for event-based roulette auth enforcement
	authHeader := c.Request().Header.Get("Authorization")
	ctx = context.WithValue(ctx, services.ContextKeyAuthHeader, authHeader)

	// Pass session_id and IP for preauth token generation if needed
	if sessionID != "" {
		ctx = context.WithValue(ctx, services.ContextKeySessionID, sessionID)
	}
	if ipAddress != "" {
		ctx = context.WithValue(ctx, services.ContextKeyIPAddress, ipAddress)
	}

	// Pass preauthToken only as parameter (Spin method doesn't use req.PreauthToken)
	response, err := h.rouletteService.Spin(ctx, preauthToken, &req)
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

	if req.RouletteID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "roulette_id is required"})
	}

	// JWT is required for roulette id != 1
	if req.RouletteID != 1 {
		userID, err := h.validateJWTToken(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authorization required for this roulette: " + err.Error()})
		}
		// Store userID in context for potential future use
		c.Set("userID", userID)
	}

	// Get preauth_token from header, query, or body (optional)
	preauthToken := c.Request().Header.Get("X-Preauth-Token")
	if preauthToken == "" {
		preauthToken = c.QueryParam("preauth_token")
	}
	if preauthToken == "" {
		preauthToken = req.PreauthToken
	}

	// Extract session_id and IP address (required if preauth_token is not provided)
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

	// If preauth_token is not provided, session_id and IP are required
	if preauthToken == "" {
		if sessionID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "preauth_token is required, or X-SESSION-ID header/cookie must be provided"})
		}
		if ipAddress == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "preauth_token is required, or IP address must be available"})
		}
	}

	// Create context with session_id and IP for internal registration and token generation
	ctx := context.Background()
	// Pass Authorization header for event-based roulette auth enforcement
	authHeader := c.Request().Header.Get("Authorization")
	ctx = context.WithValue(ctx, services.ContextKeyAuthHeader, authHeader)

	if sessionID != "" {
		ctx = context.WithValue(ctx, services.ContextKeySessionID, sessionID)
	}
	if ipAddress != "" {
		ctx = context.WithValue(ctx, services.ContextKeyIPAddress, ipAddress)
	}

	response, err := h.rouletteService.TakePrize(ctx, preauthToken, &req)
	if err != nil {
		// Check if it's a business logic error (should return 400) or server error (500)
		if strings.Contains(err.Error(), "invalid") ||
			strings.Contains(err.Error(), "already") ||
			strings.Contains(err.Error(), "must complete") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "expired") ||
			strings.Contains(err.Error(), "required") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, response)
}

// GetPreauthToken gets or creates a preauth token for on_start roulette
// Only for unauthenticated users, based on X-SESSION-ID and IP address
func (h *HTTPHandler) GetPreauthToken(c echo.Context) error {
	if h.rouletteService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "database connection required for roulette"})
	}

	// Check if user is authenticated - this endpoint is only for unauthenticated users
	_, ok := c.Get("user_uuid").(string)
	if ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "this endpoint is only for unauthenticated users"})
	}

	// Get session ID from header or cookie
	sessionID := c.Request().Header.Get("X-SESSION-ID")
	if sessionID == "" {
		cookie, err := c.Cookie("X-SESSION-ID")
		if err == nil && cookie != nil && cookie.Value != "" {
			sessionID = cookie.Value
		}
	}
	if sessionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "X-SESSION-ID header or cookie is required"})
	}

	// Get client IP
	ipAddress := c.Request().Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = c.Request().Header.Get("X-Real-IP")
	}
	if ipAddress == "" {
		ipAddress = c.RealIP()
	}
	if ipAddress == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "could not determine IP address"})
	}

	ctx := context.Background()
	token, err := h.rouletteService.GetPreauthToken(ctx, sessionID, ipAddress)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "inactive") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"preauth_token": token,
	})
}
