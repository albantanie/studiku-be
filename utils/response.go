package utils

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func SuccessJSON(c *gin.Context, status int, msg string, data interface{}) {
	c.JSON(status, Response{
		Success: true,
		Message: msg,
		Data:    data,
	})
}

func ErrorJSON(c *gin.Context, status int, err string) {
	c.JSON(status, Response{
		Success: false,
		Error:   err,
	})
}
