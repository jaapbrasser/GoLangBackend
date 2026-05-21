package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"GoLangBackend/internal/dto"
	"GoLangBackend/internal/model"
	"GoLangBackend/internal/service"
	"GoLangBackend/pkg/response"
)

type MockGitHubService struct {
	mock.Mock
}

func (m *MockGitHubService) CheckRepositoryExists(owner, repo string) (*model.Repository, error) {
	args := m.Called(owner, repo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Repository), args.Error(1)
}

func (m *MockGitHubService) CreateIssue(owner, repo, title, body string) (*model.Issue, error) {
	args := m.Called(owner, repo, title, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Issue), args.Error(1)
}

func (m *MockGitHubService) GetIssue(owner, repo string, issueNumber int) (*model.Issue, error) {
	args := m.Called(owner, repo, issueNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Issue), args.Error(1)
}

func TestCheckRepository(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid repository", func(t *testing.T) {
		mockService := new(MockGitHubService)
		mockService.On("CheckRepositoryExists", "octocat", "Hello-World").Return(&model.Repository{
			Owner:  "octocat",
			Name:   "Hello-World",
			Exists: true,
			URL:    "https://github.com/octocat/Hello-World",
		}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/repositories/check",
			bytes.NewBufferString(`{"owner":"octocat","repo":"Hello-World"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler := NewRepositoryHandler(mockService)
		handler.CheckRepository(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp response.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		var data dto.CheckRepositoryResponse
		dataBytes, _ := json.Marshal(resp.Data)
		json.Unmarshal(dataBytes, &data)
		assert.True(t, data.Exists)
		assert.Equal(t, "https://github.com/octocat/Hello-World", data.HTMLURL)
	})

	t.Run("non-existing repository", func(t *testing.T) {
		mockService := new(MockGitHubService)
		mockService.On("CheckRepositoryExists", "nonexistent", "repo").Return(&model.Repository{
			Owner:  "nonexistent",
			Name:   "repo",
			Exists: false,
		}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/repositories/check",
			bytes.NewBufferString(`{"owner":"nonexistent","repo":"repo"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler := NewRepositoryHandler(mockService)
		handler.CheckRepository(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp response.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		var data dto.CheckRepositoryResponse
		dataBytes, _ := json.Marshal(resp.Data)
		json.Unmarshal(dataBytes, &data)
		assert.False(t, data.Exists)
	})

	t.Run("invalid request", func(t *testing.T) {
		mockService := new(MockGitHubService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/repositories/check", 
			bytes.NewBufferString(`{invalid json`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler := NewRepositoryHandler(mockService)
		handler.CheckRepository(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCreateIssue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("issue created", func(t *testing.T) {
		mockService := new(MockGitHubService)
		mockService.On("CreateIssue", "octocat", "Hello-World", "Test Issue", "Test Body").Return(&model.Issue{
			Number:  42,
			HTMLURL: "https://github.com/octocat/Hello-World/issues/42",
			Title:   "Test Issue",
			Body:    "Test Body",
			State:   "open",
		}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/repositories/issues",
			bytes.NewBufferString(`{"owner":"octocat","repo":"Hello-World","title":"Test Issue","body":"Test Body"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler := NewRepositoryHandler(mockService)
		handler.CreateIssue(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp response.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		var data dto.CreateIssueResponse
		dataBytes, _ := json.Marshal(resp.Data)
		json.Unmarshal(dataBytes, &data)
		assert.Equal(t, 42, data.Number)
		assert.Equal(t, "https://github.com/octocat/Hello-World/issues/42", data.HTMLURL)
		assert.Equal(t, "Test Issue", data.Title)
		assert.Equal(t, "open", data.State)
	})

	t.Run("token not configured", func(t *testing.T) {
		mockService := new(MockGitHubService)
		mockService.On("CreateIssue", "octocat", "Hello-World", "Test Issue", "Test Body").Return(nil, service.ErrTokenNotConfigured)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/repositories/issues",
			bytes.NewBufferString(`{"owner":"octocat","repo":"Hello-World","title":"Test Issue","body":"Test Body"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler := NewRepositoryHandler(mockService)
		handler.CreateIssue(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		mockService := new(MockGitHubService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/v1/repositories/issues",
			bytes.NewBufferString(`{invalid json`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler := NewRepositoryHandler(mockService)
		handler.CreateIssue(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetIssue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("issue found", func(t *testing.T) {
		mockService := new(MockGitHubService)
		mockService.On("GetIssue", "octocat", "Hello-World", 42).Return(&model.Issue{
			Number:  42,
			HTMLURL: "https://github.com/octocat/Hello-World/issues/42",
			Title:   "Test Issue",
			Body:    "Test Body",
			State:   "open",
		}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/repositories/issues/42?owner=octocat&repo=Hello-World", nil)
		c.Params = gin.Params{{Key: "number", Value: "42"}}

		handler := NewRepositoryHandler(mockService)
		handler.GetIssue(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp response.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		var data dto.GetIssueResponse
		dataBytes, _ := json.Marshal(resp.Data)
		json.Unmarshal(dataBytes, &data)
		assert.Equal(t, 42, data.Number)
		assert.Equal(t, "https://github.com/octocat/Hello-World/issues/42", data.HTMLURL)
		assert.Equal(t, "Test Issue", data.Title)
		assert.Equal(t, "Test Body", data.Body)
		assert.Equal(t, "open", data.State)
	})

	t.Run("token not configured", func(t *testing.T) {
		mockService := new(MockGitHubService)
		mockService.On("GetIssue", "octocat", "Hello-World", 42).Return(nil, service.ErrTokenNotConfigured)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/repositories/issues/42?owner=octocat&repo=Hello-World", nil)
		c.Params = gin.Params{{Key: "number", Value: "42"}}

		handler := NewRepositoryHandler(mockService)
		handler.GetIssue(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		mockService := new(MockGitHubService)
		mockService.On("GetIssue", "octocat", "Hello-World", 999).Return(nil, service.ErrIssueNotFound)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/repositories/issues/999?owner=octocat&repo=Hello-World", nil)
		c.Params = gin.Params{{Key: "number", Value: "999"}}

		handler := NewRepositoryHandler(mockService)
		handler.GetIssue(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid number", func(t *testing.T) {
		mockService := new(MockGitHubService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/repositories/issues/abc?owner=octocat&repo=Hello-World", nil)
		c.Params = gin.Params{{Key: "number", Value: "abc"}}

		handler := NewRepositoryHandler(mockService)
		handler.GetIssue(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}