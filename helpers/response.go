package helpers

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ValidationErrors struct {
	Message string            `json:"message"`
	Errors  map[string][]string `json:"errors"`
}

func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func SuccessDataResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, Response{
		Status: "success",
		Data:   data,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Status:  "error",
		Message: message,
	})
}

func ValidationErrorResponse(c *gin.Context, statusCode int, message string, errors map[string][]string) {
	c.JSON(statusCode, ValidationErrors{
		Message: message,
		Errors:  errors,
	})
}

func HealthResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"success": true,
		"message": message,
	})
}