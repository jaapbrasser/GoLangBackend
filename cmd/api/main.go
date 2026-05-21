package main

import (
	"GoLangBackend/internal/config"
	"GoLangBackend/internal/router"
	"GoLangBackend/pkg/logger"
)

const version = "0.1.1"

func main() {
	cfg := config.Load()

	logger.Init(cfg.Environment)
	defer logger.Sync()

	r := router.SetupRouter()

	logger.L().Info("Starting server", "port", cfg.Port, "version", version)
	if err := r.Run(":" + cfg.Port); err != nil {
		logger.L().Error("Failed to start server", "error", err)
	}
}
