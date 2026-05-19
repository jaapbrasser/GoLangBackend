package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"GoLangBackend/internal/model"
)

type GitHubService interface {
	CheckRepositoryExists(owner, repo string) (*model.Repository, error)
}

type githubService struct {
	client *http.Client
	baseURL string
}

type githubRepoResponse struct {
	HTMLURL string `json:"html_url"`
}

func NewGitHubService() GitHubService {
	return &githubService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://api.github.com",
	}
}

func (s *githubService) CheckRepositoryExists(owner, repo string) (*model.Repository, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", s.baseURL, owner, repo)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &model.Repository{
			Owner:  owner,
			Name:   repo,
			Exists: false,
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github API returned status %d: %s", resp.StatusCode, string(body))
	}

	var repoResp githubRepoResponse
	if err := json.NewDecoder(resp.Body).Decode(&repoResp); err != nil {
		return nil, err
	}

	return &model.Repository{
		Owner:  owner,
		Name:   repo,
		Exists: true,
		URL:    repoResp.HTMLURL,
	}, nil
}