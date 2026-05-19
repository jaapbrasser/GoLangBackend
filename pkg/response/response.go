package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Success: true, Data: data})
}

func Fail(c *gin.Context, status int, message string) {
	c.JSON(status, Response{Success: false, Error: message})
}