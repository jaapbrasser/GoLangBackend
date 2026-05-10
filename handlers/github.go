package handlers

import (
	"fmt"
	"net/http"
	"strings"

	githubclient "github-api/github"
	apimodels "github-api/models"

	"github.com/labstack/echo/v4"
)

type GitHubHandler struct {
	client *githubclient.Client
}

func NewGitHubHandler() *GitHubHandler {
	c, err := githubclient.NewClient()
	if err != nil {
		panic("failed to initialise GitHub client: " + err.Error())
	}
	return &GitHubHandler{client: c}
}

func (h *GitHubHandler) CreateIssue(c echo.Context) error {
	owner, repo, err := ownerRepo(c)
	if err != nil {
		return badRequest(c, err.Error(), "")
	}

	var req apimodels.CreateIssueRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid JSON body", err.Error())
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return badRequest(c, "title is required", "")
	}
	if len(req.Title) > 256 {
		return badRequest(c, "title must be 256 characters or fewer", "")
	}
	if len(req.Body) > 65536 {
		return badRequest(c, "body must be 65 536 characters or fewer", "")
	}

	payload := githubclient.CreateIssuePayload{
		Title:     req.Title,
		Body:      req.Body,
		Assignees: req.Assignees,
		Labels:    req.Labels,
		Milestone: req.Milestone,
	}

	issue, err := h.client.CreateIssue(c.Request().Context(), owner, repo, payload)
	if err != nil {
		return githubError(c, err)
	}

	return c.JSON(http.StatusCreated, apimodels.IssueResponse{
		Number:  issue.Number,
		Title:   issue.Title,
		Body:    issue.Body,
		State:   issue.State,
		URL:     issue.HTMLURL,
		Created: issue.CreatedAt,
	})
}

func (h *GitHubHandler) ValidateRepository(c echo.Context) error {
	owner, repo, err := ownerRepo(c)
	if err != nil {
		return badRequest(c, err.Error(), "")
	}

	exists, info, err := h.client.RepositoryExists(c.Request().Context(), owner, repo)
	if err != nil {
		return githubError(c, err)
	}

	if !exists {
		return c.JSON(http.StatusOK, apimodels.RepoValidationResponse{Exists: false})
	}

	return c.JSON(http.StatusOK, apimodels.RepoValidationResponse{
		Exists:        true,
		FullName:      info.FullName,
		Description:   info.Description,
		Private:       info.Private,
		URL:           info.HTMLURL,
		Stars:         info.StarCount,
		Language:      info.Language,
		DefaultBranch: info.DefaultBranch,
	})
}

func ownerRepo(c echo.Context) (string, string, error) {
	owner := strings.TrimSpace(c.Param("owner"))
	repo := strings.TrimSpace(c.Param("repo"))
	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("owner and repo path parameters are required")
	}
	if !validGitHubName(owner) {
		return "", "", fmt.Errorf("invalid owner name")
	}
	if !validGitHubName(repo) {
		return "", "", fmt.Errorf("invalid repo name")
	}
	return owner, repo, nil
}

func validGitHubName(s string) bool {
	if len(s) == 0 || len(s) > 100 {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.') {
			return false
		}
	}
	return true
}

func badRequest(c echo.Context, msg, detail string) error {
	return c.JSON(http.StatusBadRequest, apimodels.ErrorResponse{Error: msg, Details: detail})
}

func githubError(c echo.Context, err error) error {
	if ghErr, ok := err.(*githubclient.GitHubError); ok {
		status := http.StatusBadGateway
		switch ghErr.StatusCode {
		case http.StatusNotFound:
			status = http.StatusNotFound
		case http.StatusUnprocessableEntity:
			status = http.StatusUnprocessableEntity
		case http.StatusForbidden:
			status = http.StatusForbidden
		case http.StatusUnauthorized:
			status = http.StatusUnauthorized
		}
		return c.JSON(status, apimodels.ErrorResponse{Error: ghErr.Message})
	}
	c.Logger().Errorf("upstream error: %v", err)
	return c.JSON(http.StatusInternalServerError, apimodels.ErrorResponse{Error: "internal server error"})
}
