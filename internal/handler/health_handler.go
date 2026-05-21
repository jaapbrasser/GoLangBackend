package handler

import (
	"GoLangBackend/pkg/response"
	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok"})
}
