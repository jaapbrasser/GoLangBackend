package service

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetIssue_Integration(t *testing.T) {
	// Skip if no token provided
	token := os.Getenv("GITHUB_ISSUETOKEN")
	if token == "" {
		t.Skip("Skipping integration test because GITHUB_ISSUETOKEN is not set")
	}

	// Use the real GitHub API
	svc := NewGitHubServiceWithTokens(map[string]string{
		"jaapbrasser/GoLangBackend": token,
	})

	// Get issue #1
	issue, err := svc.GetIssue("jaapbrasser", "GoLangBackend", 1)
	assert.NoError(t, err, "Failed to get issue #1")
	assert.NotNil(t, issue)

	// Verify the title
	assert.Equal(t, "Active Test Issue for Integration Test", issue.Title, "Issue title does not match expected value")

	// Verify the description/body (normalizing whitespace and newlines)
	expectedBody := "Do not remove, thank you.  9119  🦄"
	// Normalize both strings by replacing newlines with spaces and trimming spaces
	normalizedExpected := strings.Join(strings.Fields(expectedBody), " ")
	normalizedActual := strings.Join(strings.Fields(issue.Body), " ")
	assert.Equal(t, normalizedExpected, normalizedActual, "Issue body does not match expected value")
}

func TestCreateIssue_Integration(t *testing.T) {
	// Skip if no token provided
	token := os.Getenv("GITHUB_ISSUETOKEN")
	if token == "" {
		t.Skip("Skipping integration test because GITHUB_ISSUETOKEN is not set")
	}

	// Use the real GitHub API
	svc := NewGitHubServiceWithTokens(map[string]string{
		"jaapbrasser/GoLangBackend": token,
	})

	// Create a unique title and body to avoid duplicate issues
	timestamp := time.Now().Unix()
	title := "Integration Test Issue " + strconv.FormatInt(timestamp, 10)
	body := "This is a test issue created by integration test at " + strconv.FormatInt(timestamp, 10)

	// Create the issue
	issue, err := svc.CreateIssue("jaapbrasser", "GoLangBackend", title, body)
	assert.NoError(t, err, "Failed to create issue")
	assert.NotNil(t, issue)
	assert.NotEqual(t, 0, issue.Number, "Issue number should be non-zero")
	assert.Equal(t, title, issue.Title, "Created issue title does not match")
	assert.Equal(t, body, issue.Body, "Created issue body does not match")

	// Clean up: close the created issue
	closedIssue, err := svc.CloseIssue("jaapbrasser", "GoLangBackend", issue.Number)
	assert.NoError(t, err, "Failed to close issue")
	assert.NotNil(t, closedIssue)
	assert.Equal(t, "closed", closedIssue.State, "Issue should be closed after cleanup")

	// Verify the issue is actually closed by retrieving it
	retrievedIssue, err := svc.GetIssue("jaapbrasser", "GoLangBackend", issue.Number)
	assert.NoError(t, err, "Failed to retrieve issue after closing")
	assert.NotNil(t, retrievedIssue)
	assert.Equal(t, "closed", retrievedIssue.State, "Issue should remain closed when retrieved")
}
