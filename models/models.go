package github

import "fmt"

// ─── Request payloads ─────────────────────────────────────────────────────────

// CreateIssuePayload is what we send to GitHub when opening an issue.
type CreateIssuePayload struct {
	Title     string   `json:"title"`
	Body      string   `json:"body,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	Milestone *int     `json:"milestone,omitempty"`
}

// ─── Response models ──────────────────────────────────────────────────────────

// Issue is a trimmed representation of a GitHub issue.
type Issue struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	HTMLURL   string `json:"html_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	User      *User  `json:"user,omitempty"`
}

// RepositoryInfo holds the subset of repository metadata we surface to callers.
type RepositoryInfo struct {
	ID            int64  `json:"id"`
	FullName      string `json:"full_name"`
	Description   string `json:"description"`
	Private       bool   `json:"private"`
	HTMLURL       string `json:"html_url"`
	Fork          bool   `json:"fork"`
	StarCount     int    `json:"stargazers_count"`
	WatchCount    int    `json:"watchers_count"`
	Language      string `json:"language"`
	DefaultBranch string `json:"default_branch"`
}

// User is a minimal GitHub user object.
type User struct {
	Login   string `json:"login"`
	HTMLURL string `json:"html_url"`
}

// ─── Error type ───────────────────────────────────────────────────────────────

// GitHubError represents an error response from the GitHub API.
type GitHubError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"-"` // populated by the client
}

func (e *GitHubError) Error() string {
	return fmt.Sprintf("github: %s (HTTP %d)", e.Message, e.StatusCode)
}
