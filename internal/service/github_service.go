package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"GoLangBackend/internal/model"
)

type githubService struct {
	client  *http.Client
	baseURL string
	tokens  map[string]string
}

type githubRepoResponse struct {
	HTMLURL string `json:"html_url"`
}

type githubIssueRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type githubIssueResponse struct {
	Number  int    `json:"number"`
	HTMLURL string `json:"html_url"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	State   string `json:"state"`
}

var (
	ErrTokenNotConfigured = errors.New("no github token configured for the requested repository")
	ErrUnauthorized       = errors.New("unauthorized: invalid github token")
	ErrIssueNotFound      = errors.New("issue not found")
)

type GitHubService interface {
	CheckRepositoryExists(owner, repo string) (*model.Repository, error)
	CreateIssue(owner, repo, title, body string) (*model.Issue, error)
	GetIssue(owner, repo string, issueNumber int) (*model.Issue, error)
	CloseIssue(owner, repo string, issueNumber int) (*model.Issue, error)
}

func NewGitHubService() GitHubService {
	return &githubService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://api.github.com",
	}
}

func NewGitHubServiceWithTokens(tokens map[string]string) GitHubService {
	return &githubService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://api.github.com",
		tokens:  tokens,
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

func (s *githubService) getToken(owner, repo string) (string, error) {
	if s.tokens == nil {
		return "", ErrTokenNotConfigured
	}
	key := owner + "/" + repo
	token, ok := s.tokens[key]
	if !ok {
		return "", ErrTokenNotConfigured
	}
	if token == "" {
		return "", ErrTokenNotConfigured
	}
	return token, nil
}

func (s *githubService) CreateIssue(owner, repo, title, body string) (*model.Issue, error) {
	token, err := s.getToken(owner, repo)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/repos/%s/%s/issues", s.baseURL, owner, repo)

	reqBody, err := json.Marshal(githubIssueRequest{Title: title, Body: body})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github API returned status %d: %s", resp.StatusCode, string(body))
	}

	var issueResp githubIssueResponse
	if err := json.NewDecoder(resp.Body).Decode(&issueResp); err != nil {
		return nil, err
	}

	return &model.Issue{
		Number:  issueResp.Number,
		HTMLURL: issueResp.HTMLURL,
		Title:   issueResp.Title,
		Body:    issueResp.Body,
		State:   issueResp.State,
	}, nil
}

func (s *githubService) GetIssue(owner, repo string, issueNumber int) (*model.Issue, error) {
	token, err := s.getToken(owner, repo)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", s.baseURL, owner, repo, issueNumber)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrIssueNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github API returned status %d: %s", resp.StatusCode, string(body))
	}

	var issueResp githubIssueResponse
	if err := json.NewDecoder(resp.Body).Decode(&issueResp); err != nil {
		return nil, err
	}

	return &model.Issue{
		Number:  issueResp.Number,
		HTMLURL: issueResp.HTMLURL,
		Title:   issueResp.Title,
		Body:    issueResp.Body,
		State:   issueResp.State,
	}, nil
}

func (s *githubService) CloseIssue(owner, repo string, issueNumber int) (*model.Issue, error) {
	token, err := s.getToken(owner, repo)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", s.baseURL, owner, repo, issueNumber)

	// To close an issue, we need to PATCH it with state: "closed"
	reqBody, err := json.Marshal(map[string]string{"state": "closed"})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github API returned status %d: %s", resp.StatusCode, string(body))
	}

	var issueResp githubIssueResponse
	if err := json.NewDecoder(resp.Body).Decode(&issueResp); err != nil {
		return nil, err
	}

	return &model.Issue{
		Number:  issueResp.Number,
		HTMLURL: issueResp.HTMLURL,
		Title:   issueResp.Title,
		Body:    issueResp.Body,
		State:   issueResp.State,
	}, nil
}
