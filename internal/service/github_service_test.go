package service

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckRepositoryExists(t *testing.T) {
	tests := []struct {
		name       string
		owner      string
		repo       string
		mockStatus int
		mockBody   string
		wantExists bool
		wantURL    string
		wantErr    bool
	}{
		{
			name:       "existing repository",
			owner:      "octocat",
			repo:       "Hello-World",
			mockStatus: http.StatusOK,
			mockBody:   `{"html_url":"https://github.com/octocat/Hello-World"}`,
			wantExists: true,
			wantURL:    "https://github.com/octocat/Hello-World",
		},
		{
			name:       "non-existing repository",
			owner:      "nonexistent",
			repo:       "repo",
			mockStatus: http.StatusNotFound,
			mockBody:   "",
			wantExists: false,
			wantURL:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatus)
				w.Write([]byte(tt.mockBody))
			}))
			defer server.Close()

			svc := &githubService{
				client:  server.Client(),
				baseURL: server.URL,
			}

			repo, err := svc.CheckRepositoryExists(tt.owner, tt.repo)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantExists, repo.Exists)
				assert.Equal(t, tt.wantURL, repo.URL)
			}
		})
	}
}

func TestCreateIssue(t *testing.T) {
	tests := []struct {
		name       string
		owner      string
		repo       string
		title      string
		body       string
		tokens     map[string]string
		mockStatus int
		mockBody   string
		wantErr    bool
		wantErrIs  error
	}{
		{
			name:       "issue created",
			owner:      "octocat",
			repo:       "Hello-World",
			title:      "Test Issue",
			body:       "Test body",
			tokens:     map[string]string{"octocat/Hello-World": "valid-token"},
			mockStatus: http.StatusCreated,
			mockBody:   `{"number":42,"html_url":"https://github.com/octocat/Hello-World/issues/42","title":"Test Issue","body":"Test body","state":"open"}`,
			wantErr:    false,
		},
		{
			name:      "no tokens configured",
			owner:     "octocat",
			repo:      "Hello-World",
			title:     "Test Issue",
			body:      "Test body",
			tokens:    nil,
			wantErr:   true,
			wantErrIs: ErrTokenNotConfigured,
		},
		{
			name:      "token not for repo",
			owner:     "octocat",
			repo:      "Hello-World",
			title:     "Test Issue",
			body:      "Test body",
			tokens:    map[string]string{"other/repo": "token"},
			wantErr:   true,
			wantErrIs: ErrTokenNotConfigured,
		},
		{
			name:       "github returns 401",
			owner:      "octocat",
			repo:       "Hello-World",
			title:      "Test Issue",
			body:       "Test body",
			tokens:     map[string]string{"octocat/Hello-World": "invalid-token"},
			mockStatus: http.StatusUnauthorized,
			mockBody:   "",
			wantErr:    true,
			wantErrIs:  ErrUnauthorized,
		},
		{
			name:       "github returns 500",
			owner:      "octocat",
			repo:       "Hello-World",
			title:      "Test Issue",
			body:       "Test body",
			tokens:     map[string]string{"octocat/Hello-World": "token"},
			mockStatus: http.StatusInternalServerError,
			mockBody:   "",
			wantErr:    true,
		},
		{
			name:       "bad response body",
			owner:      "octocat",
			repo:       "Hello-World",
			title:      "Test Issue",
			body:       "Test body",
			tokens:     map[string]string{"octocat/Hello-World": "token"},
			mockStatus: http.StatusCreated,
			mockBody:   `invalid json`,
			wantErr:    true,
		},
		{
			name:      "empty token configured",
			owner:     "octocat",
			repo:      "Hello-World",
			title:     "Test Issue",
			body:      "Test body",
			tokens:    map[string]string{"octocat/Hello-World": ""},
			wantErr:   true,
			wantErrIs: ErrTokenNotConfigured,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatus)
				w.Write([]byte(tt.mockBody))
			}))
			defer server.Close()

			svc := &githubService{
				client:  server.Client(),
				baseURL: server.URL,
				tokens:  tt.tokens,
			}

			issue, err := svc.CreateIssue(tt.owner, tt.repo, tt.title, tt.body)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, issue)
				assert.Equal(t, 42, issue.Number)
				assert.Equal(t, "Test Issue", issue.Title)
			}
		})
	}
}

func TestGetIssue(t *testing.T) {
	tests := []struct {
		name        string
		owner       string
		repo        string
		issueNumber int
		tokens      map[string]string
		mockStatus  int
		mockBody    string
		wantErr     bool
		wantErrIs   error
	}{
		{
			name:        "issue found",
			owner:       "octocat",
			repo:        "Hello-World",
			issueNumber: 42,
			tokens:      map[string]string{"octocat/Hello-World": "valid-token"},
			mockStatus:  http.StatusOK,
			mockBody:    `{"number":42,"html_url":"https://github.com/octocat/Hello-World/issues/42","title":"Test Issue","body":"Test body","state":"open"}`,
			wantErr:     false,
		},
		{
			name:        "no tokens configured",
			owner:       "octocat",
			repo:        "Hello-World",
			issueNumber: 42,
			tokens:      nil,
			wantErr:     true,
			wantErrIs:   ErrTokenNotConfigured,
		},
		{
			name:        "github 404 not found",
			owner:       "octocat",
			repo:        "Hello-World",
			issueNumber: 999,
			tokens:      map[string]string{"octocat/Hello-World": "token"},
			mockStatus:  http.StatusNotFound,
			mockBody:    "",
			wantErr:     true,
			wantErrIs:   ErrIssueNotFound,
		},
		{
			name:        "github returns 401",
			owner:       "octocat",
			repo:        "Hello-World",
			issueNumber: 42,
			tokens:      map[string]string{"octocat/Hello-World": "invalid-token"},
			mockStatus:  http.StatusUnauthorized,
			mockBody:    "",
			wantErr:     true,
			wantErrIs:   ErrUnauthorized,
		},
		{
			name:        "empty token configured",
			owner:       "octocat",
			repo:        "Hello-World",
			issueNumber: 42,
			tokens:      map[string]string{"octocat/Hello-World": ""},
			wantErr:     true,
			wantErrIs:   ErrTokenNotConfigured,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatus)
				w.Write([]byte(tt.mockBody))
			}))
			defer server.Close()

			svc := &githubService{
				client:  server.Client(),
				baseURL: server.URL,
				tokens:  tt.tokens,
			}

			issue, err := svc.GetIssue(tt.owner, tt.repo, tt.issueNumber)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, issue)
				assert.Equal(t, 42, issue.Number)
				assert.Equal(t, "Test Issue", issue.Title)
			}
		})
	}
}
