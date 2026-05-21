package handler

import (
	"net/http"
	"strconv"

	"GoLangBackend/internal/dto"
	"GoLangBackend/internal/service"
	"GoLangBackend/pkg/response"

	"github.com/gin-gonic/gin"
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

func (h *RepositoryHandler) CreateIssue(c *gin.Context) {
	var req dto.CreateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	issue, err := h.githubService.CreateIssue(req.Owner, req.Repo, req.Title, req.Body)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	resp := dto.CreateIssueResponse{
		Number:  issue.Number,
		HTMLURL: issue.HTMLURL,
		Title:   issue.Title,
		State:   issue.State,
	}
	response.OK(c, resp)
}

func (h *RepositoryHandler) GetIssue(c *gin.Context) {
	issueNumberStr := c.Param("number")
	issueNumber, err := strconv.Atoi(issueNumberStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid issue number")
		return
	}

	owner := c.Query("owner")
	repo := c.Query("repo")
	if owner == "" || repo == "" {
		response.Fail(c, http.StatusBadRequest, "owner and repo query parameters required")
		return
	}

	issue, err := h.githubService.GetIssue(owner, repo, issueNumber)
	if err != nil {
		response.Fail(c, http.StatusNotFound, err.Error())
		return
	}

	resp := dto.GetIssueResponse{
		Number:  issue.Number,
		HTMLURL: issue.HTMLURL,
		Title:   issue.Title,
		Body:    issue.Body,
		State:   issue.State,
	}
	response.OK(c, resp)
}
