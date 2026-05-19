package router

import (
	"github.com/gin-gonic/gin"
	"GoLangBackend/internal/handler"
	"GoLangBackend/internal/middleware"
	"GoLangBackend/internal/service"
)

func SetupRouter() *gin.Engine {
	r := gin.New()

	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", handler.Health)

	repoService := service.NewGitHubService()
	repoHandler := handler.NewRepositoryHandler(repoService)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/repositories/check", repoHandler.CheckRepository)
	}

	return r
}