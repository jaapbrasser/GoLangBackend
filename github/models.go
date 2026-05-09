package models

// ─── Requests ─────────────────────────────────────────────────────────────────

// CreateIssueRequest is the JSON body accepted by POST /api/v1/repos/:owner/:repo/issues.
type CreateIssueRequest struct {
	Title     string   `json:"title"     validate:"required,min=1,max=256"`
	Body      string   `json:"body"      validate:"max=65536"`
	Assignees []string `json:"assignees" validate:"omitempty,dive,alphanum"`
	Labels    []string `json:"labels"    validate:"omitempty,dive,min=1"`
	Milestone *int     `json:"milestone" validate:"omitempty,min=1"`
}

// ─── Responses ────────────────────────────────────────────────────────────────

// IssueResponse is returned after successfully creating a GitHub issue.
type IssueResponse struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	State   string `json:"state"`
	URL     string `json:"url"`
	Created string `json:"created_at"`
}

// RepoValidationResponse is returned by GET /api/v1/repos/:owner/:repo/validate.
type RepoValidationResponse struct {
	Exists      bool   `json:"exists"`
	FullName    string `json:"full_name,omitempty"`
	Description string `json:"description,omitempty"`
	Private     bool   `json:"private,omitempty"`
	URL         string `json:"url,omitempty"`
	Stars       int    `json:"stars,omitempty"`
	Language    string `json:"language,omitempty"`
	DefaultBranch string `json:"default_branch,omitempty"`
}

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
