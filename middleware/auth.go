// Package middleware provides Echo middleware for this service.
package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github-api/models"

	"github.com/labstack/echo/v4"
)

// APIKeyAuth returns an Echo middleware that validates the X-API-Key header
// against the API_KEY environment variable. Requests to /health bypass auth.
//
// Security notes:
//   - The expected key is read once from the environment at middleware
//     construction, never from an in-memory store or database.
//   - Comparison uses crypto/subtle.ConstantTimeCompare to prevent
//     timing-based side-channel attacks.
func APIKeyAuth() echo.MiddlewareFunc {
	expectedKey := os.Getenv("API_KEY")
	if expectedKey == "" {
		panic("API_KEY environment variable is required")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Bypass auth for health check
			if c.Path() == "/health" {
				return next(c)
			}

			provided := strings.TrimSpace(c.Request().Header.Get("X-API-Key"))
			if provided == "" {
				return c.JSON(http.StatusUnauthorized, models.ErrorResponse{
					Error: "X-API-Key header is required",
				})
			}

			if subtle.ConstantTimeCompare([]byte(provided), []byte(expectedKey)) != 1 {
				return c.JSON(http.StatusForbidden, models.ErrorResponse{
					Error: "invalid API key",
				})
			}

			return next(c)
		}
	}
}
