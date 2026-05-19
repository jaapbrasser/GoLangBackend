package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"GoLangBackend/internal/dto"
	"GoLangBackend/internal/service"
	"GoLangBackend/pkg/response"
)

type RepositoryHandler struct {
	githubService service.GitHubService
}

func NewRepositoryHandler(gs service.GitHubService) *RepositoryHandler {
	return &RepositoryHandler{githubService: gs}
}

func (h *RepositoryHandler) CheckRepository(c *gin.Context) {
	var req dto.CheckRepositoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	repo, err := h.githubService.CheckRepositoryExists(req.Owner, req.Repo)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	resp := dto.CheckRepositoryResponse{
		Exists:  repo.Exists,
		HTMLURL: repo.URL,
	}
	response.OK(c, resp)
}