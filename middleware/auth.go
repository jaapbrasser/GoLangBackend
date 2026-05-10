package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	apimodels "github-api/models"

	"github.com/labstack/echo/v4"
)

func APIKeyAuth() echo.MiddlewareFunc {
	expectedKey := os.Getenv("API_KEY")
	if expectedKey == "" {
		panic("API_KEY environment variable is required")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Path() == "/health" {
				return next(c)
			}

			provided := strings.TrimSpace(c.Request().Header.Get("X-API-Key"))
			if provided == "" {
				return c.JSON(http.StatusUnauthorized, apimodels.ErrorResponse{
					Error: "X-API-Key header is required",
				})
			}

			if subtle.ConstantTimeCompare([]byte(provided), []byte(expectedKey)) != 1 {
				return c.JSON(http.StatusForbidden, apimodels.ErrorResponse{
					Error: "invalid API key",
				})
			}

			return next(c)
		}
	}
}
