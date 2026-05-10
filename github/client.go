// Package githubclient provides a thin, secure client around the GitHub REST API.
// The personal access token is read once from the environment at construction
// time and is never exposed outside this package.
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const baseURL = "https://api.github.com"

// Client is a GitHub REST API client. Always construct via NewClient().
type Client struct {
	httpClient *http.Client
	token      string // never exported
}

// NewClient reads GITHUB_TOKEN from the environment and returns a ready-to-use
// Client. Returns an error if the token is absent.
func NewClient() (*Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is not set")
	}
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		token:      token,
	}, nil
}

// ─── Request helpers ──────────────────────────────────────────────────────────

func (c *Client) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		buf = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, buf)
	if err != nil {
		return nil, err
	}

	// Authorization header — token never leaks into query strings or logs
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (c *Client) do(req *http.Request, out any) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MB cap
	if err != nil {
		return resp, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var ghErr GitHubError
		if jsonErr := json.Unmarshal(respBody, &ghErr); jsonErr == nil && ghErr.Message != "" {
			ghErr.StatusCode = resp.StatusCode // ← was never set; broke all status-code checks
			return resp, &ghErr
		}
		return resp, fmt.Errorf("github API error %d", resp.StatusCode)
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return resp, fmt.Errorf("decoding response: %w", err)
		}
	}
	return resp, nil
}

// ─── Public API methods ───────────────────────────────────────────────────────

// RepositoryExists returns (true, nil) when the repo is accessible,
// (false, nil) when it does not exist / is inaccessible, or an error for
// unexpected failures.
func (c *Client) RepositoryExists(ctx context.Context, owner, repo string) (bool, *RepositoryInfo, error) {
	req, err := c.newRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s", owner, repo), nil)
	if err != nil {
		return false, nil, err
	}

	var info RepositoryInfo
	if _, err := c.do(req, &info); err != nil {
		if ghErr, ok := err.(*GitHubError); ok && (ghErr.StatusCode == 404 || ghErr.StatusCode == 403) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, &info, nil
}

// CreateIssue opens a new issue in the given repository.
func (c *Client) CreateIssue(ctx context.Context, owner, repo string, payload CreateIssuePayload) (*Issue, error) {
	req, err := c.newRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/issues", owner, repo), payload)
	if err != nil {
		return nil, err
	}

	var issue Issue
	if _, err := c.do(req, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}
