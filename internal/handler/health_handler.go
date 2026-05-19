package handler

import (
	"github.com/gin-gonic/gin"
	"GoLangBackend/pkg/response"
)

func Health(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok"})
}