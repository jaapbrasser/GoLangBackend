package router

import (
	"encoding/json"
	"os"

	"GoLangBackend/internal/handler"
	"GoLangBackend/internal/middleware"
	"GoLangBackend/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.New()

	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", handler.Health)

	var repoService service.GitHubService
	if tokensJSON := os.Getenv("GITHUB_TOKENS"); tokensJSON != "" {
		var tokens map[string]string
		if err := json.Unmarshal([]byte(tokensJSON), &tokens); err == nil {
			repoService = service.NewGitHubServiceWithTokens(tokens)
		} else {
			repoService = service.NewGitHubService()
		}
	} else {
		repoService = service.NewGitHubService()
	}
	repoHandler := handler.NewRepositoryHandler(repoService)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/repositories/check", repoHandler.CheckRepository)
		v1.POST("/repositories/issues", repoHandler.CreateIssue)
		v1.GET("/repositories/issues/:number", repoHandler.GetIssue)
	}

	return r
}
