package service

import (
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