package handler

import "github.com/gin-gonic/gin"

func (h *Handler) Home(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello, World!",
	})
}
