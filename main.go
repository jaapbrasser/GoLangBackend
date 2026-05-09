package main

import (
	"log"
	"net/http"
	"os"

	"github-api/handlers"
	"github-api/middleware"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load .env file — ignore error in production (env vars may be set externally)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading environment variables directly")
	}

	// Validate required env vars at startup
	if os.Getenv("GITHUB_TOKEN") == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}

	e := echo.New()
	e.HideBanner = true

	// ── Global middleware ──────────────────────────────────────────────────────
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.RequestID())
	e.Use(echomiddleware.SecureWithConfig(echomiddleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		ContentSecurityPolicy: "default-src 'self'",
	}))
	e.Use(echomiddleware.RateLimiter(echomiddleware.NewRateLimiterMemoryStore(20)))

	// Internal auth middleware — validates X-API-Key header
	e.Use(middleware.APIKeyAuth())

	// ── Routes ────────────────────────────────────────────────────────────────
	gh := handlers.NewGitHubHandler()

	api := e.Group("/api/v1")
	{
		// POST /api/v1/repos/:owner/:repo/issues  — create a new issue
		api.POST("/repos/:owner/:repo/issues", gh.CreateIssue)

		// GET  /api/v1/repos/:owner/:repo/validate — check a repo exists
		api.GET("/repos/:owner/:repo/validate", gh.ValidateRepository)
	}

	// Health check (no auth required)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	e.Logger.Fatal(e.Start(":" + port))
}
