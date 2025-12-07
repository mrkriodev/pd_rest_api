package http

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWTMiddleware creates a middleware that validates JWT tokens
func JWTMiddleware(secretKey string, strictMode bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get token from Authorization header
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

			// If strict mode is disabled, only check that token is non-empty
			if !strictMode {
				if tokenString == "" {
					return c.JSON(http.StatusUnauthorized, map[string]string{"error": "token cannot be empty"})
				}
				// Try to extract UUID from token if possible (without validation)
				// This is optional - if parsing fails, we still allow the request
				if token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					return []byte(secretKey), nil
				}); err == nil {
					if claims, ok := token.Claims.(jwt.MapClaims); ok {
						if uuid, ok := claims["uuid"].(string); ok {
							c.Set("user_uuid", uuid)
						}
					}
				}
				return next(c)
			}

			// Strict mode: Parse and validate token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Validate signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secretKey), nil
			})

			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			}

			if !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			}

			// Extract claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
			}

			// Store user UUID in context for use in handlers
			if uuid, ok := claims["uuid"].(string); ok {
				c.Set("user_uuid", uuid)
			}

			return next(c)
		}
	}
}
