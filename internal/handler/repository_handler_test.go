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