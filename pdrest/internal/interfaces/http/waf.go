package http

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// IPBanService manages banned IP addresses with time-based expiration
type IPBanService struct {
	bannedIPs map[string]time.Time
	mu        sync.RWMutex
	banTTL    time.Duration // Time to live for bans
}

// NewIPBanService creates a new IP ban service
func NewIPBanService(banTTL time.Duration) *IPBanService {
	service := &IPBanService{
		bannedIPs: make(map[string]time.Time),
		banTTL:    banTTL,
	}

	// Start cleanup goroutine to remove expired bans
	go service.cleanupExpiredBans()

	return service
}

// IsBanned checks if an IP is currently banned
func (s *IPBanService) IsBanned(ip string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	banTime, exists := s.bannedIPs[ip]
	if !exists {
		return false
	}

	// Check if ban has expired
	if time.Now().After(banTime) {
		// Ban expired, remove it
		s.mu.RUnlock()
		s.mu.Lock()
		delete(s.bannedIPs, ip)
		s.mu.Unlock()
		s.mu.RLock()
		return false
	}

	return true
}

// BanIP bans an IP address for the configured TTL
func (s *IPBanService) BanIP(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bannedIPs[ip] = time.Now().Add(s.banTTL)
}

// UnbanIP removes a ban from an IP address
func (s *IPBanService) UnbanIP(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.bannedIPs, ip)
}

// cleanupExpiredBans periodically removes expired bans
func (s *IPBanService) cleanupExpiredBans() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for ip, banTime := range s.bannedIPs {
			if now.After(banTime) {
				delete(s.bannedIPs, ip)
			}
		}
		s.mu.Unlock()
	}
}

// WAFConfig holds configuration for the WAF middleware
type WAFConfig struct {
	// Active enables/disables WAF (if false, middleware is bypassed)
	Active bool
	// RequireSessionID requires X-SESSION-ID in cookies or headers
	RequireSessionID bool
	// SessionIDHeader is the header name to check (default: "X-SESSION-ID")
	SessionIDHeader string
	// SessionIDCookie is the cookie name to check (default: "X-SESSION-ID")
	SessionIDCookie string
	// BanOnMissingSession bans IPs that don't provide session ID
	BanOnMissingSession bool
	// BanTTL is the duration for which IPs are banned
	BanTTL time.Duration
	// WhitelistedPaths are paths that don't require session ID
	WhitelistedPaths []string
	// IPBanService manages banned IPs
	IPBanService *IPBanService
}

// DefaultWAFConfig returns a default WAF configuration
func DefaultWAFConfig() *WAFConfig {
	return &WAFConfig{
		Active:              false, // Disabled by default
		RequireSessionID:    true,
		SessionIDHeader:     "X-SESSION-ID",
		SessionIDCookie:     "X-SESSION-ID",
		BanOnMissingSession: true,
		BanTTL:              24 * time.Hour, // Ban for 24 hours by default
		WhitelistedPaths:    []string{"/api/status"},
		IPBanService:        NewIPBanService(24 * time.Hour),
	}
}

// WAFMiddleware creates middleware for Web Application Firewall functionality
func WAFMiddleware(config *WAFConfig) echo.MiddlewareFunc {
	if config == nil {
		config = DefaultWAFConfig()
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// If WAF is not active, bypass all checks
			if !config.Active {
				return next(c)
			}

			// Get client IP
			ip := getClientIP(c)

			// Check if IP is banned
			if config.IPBanService.IsBanned(ip) {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "IP address is banned",
				})
			}

			// Check if path is whitelisted
			path := c.Request().URL.Path
			isWhitelisted := false
			for _, whitelistedPath := range config.WhitelistedPaths {
				if path == whitelistedPath {
					isWhitelisted = true
					break
				}
			}

			// If session ID is required and path is not whitelisted
			if config.RequireSessionID && !isWhitelisted {
				sessionID := getSessionID(c, config.SessionIDHeader, config.SessionIDCookie)

				if sessionID == "" {
					// Ban IP if configured to do so
					if config.BanOnMissingSession {
						config.IPBanService.BanIP(ip)
					}

					return c.JSON(http.StatusForbidden, map[string]string{
						"error": "X-SESSION-ID is required in cookies or headers",
					})
				}
			}

			// Continue to next handler
			return next(c)
		}
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(c echo.Context) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	forwardedFor := c.Request().Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		return forwardedFor
	}

	// Check X-Real-IP header
	realIP := c.Request().Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return c.RealIP()
}

// getSessionID retrieves the session ID from headers or cookies
func getSessionID(c echo.Context, headerName, cookieName string) string {
	// First, check header
	sessionID := c.Request().Header.Get(headerName)
	if sessionID != "" {
		return sessionID
	}

	// Then, check cookie
	cookie, err := c.Cookie(cookieName)
	if err == nil && cookie != nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}
