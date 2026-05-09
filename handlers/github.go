package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github-api/github"
	"github-api/models"

	"github.com/labstack/echo/v4"
)

// GitHubHandler holds a shared GitHub client for all route handlers.
type GitHubHandler struct {
	client *github.Client
}

// NewGitHubHandler constructs a GitHubHandler. Panics on startup if the
// GitHub token is missing — it is better to fail fast than serve broken routes.
func NewGitHubHandler() *GitHubHandler {
	c, err := github.NewClient()
	if err != nil {
		panic("failed to initialise GitHub client: " + err.Error())
	}
	return &GitHubHandler{client: c}
}

// ─── POST /api/v1/repos/:owner/:repo/issues ───────────────────────────────────

// CreateIssue validates the request body and opens a new issue on GitHub.
//
//	POST /api/v1/repos/{owner}/{repo}/issues
//	Body: models.CreateIssueRequest (JSON)
func (h *GitHubHandler) CreateIssue(c echo.Context) error {
	owner, repo, err := ownerRepo(c)
	if err != nil {
		return badRequest(c, err.Error(), "")
	}

	var req models.CreateIssueRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid JSON body", err.Error())
	}

	// Manual validation (swap in go-playground/validator if you prefer)
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

	payload := github.CreateIssuePayload{
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

	return c.JSON(http.StatusCreated, models.IssueResponse{
		Number:  issue.Number,
		Title:   issue.Title,
		Body:    issue.Body,
		State:   issue.State,
		URL:     issue.HTMLURL,
		Created: issue.CreatedAt,
	})
}

// ─── GET /api/v1/repos/:owner/:repo/validate ──────────────────────────────────

// ValidateRepository checks whether a GitHub repository exists and is
// accessible with the configured token.
//
//	GET /api/v1/repos/{owner}/{repo}/validate
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
		return c.JSON(http.StatusOK, models.RepoValidationResponse{Exists: false})
	}

	return c.JSON(http.StatusOK, models.RepoValidationResponse{
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

// ─── Helpers ──────────────────────────────────────────────────────────────────

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

// validGitHubName allows only characters GitHub itself permits in owner/repo
// names: alphanumerics, hyphens, underscores, and dots.
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
	return c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: msg, Details: detail})
}

func githubError(c echo.Context, err error) error {
	if ghErr, ok := err.(*github.GitHubError); ok {
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
		return c.JSON(status, models.ErrorResponse{Error: ghErr.Message})
	}
	// Log the internal detail server-side; never expose it to the caller.
	c.Logger().Errorf("upstream error: %v", err)
	return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error"})
}
