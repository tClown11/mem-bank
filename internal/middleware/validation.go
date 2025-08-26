package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ValidateUUID middleware validates UUID parameters
func ValidateUUID(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuidStr := c.Param(paramName)
		if uuidStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "UUID parameter is required",
				"code":    "missing_uuid",
				"param":   paramName,
			})
			c.Abort()
			return
		}
		
		if _, err := uuid.Parse(uuidStr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid UUID format",
				"code":    "invalid_uuid",
				"param":   paramName,
				"value":   uuidStr,
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// ValidateJSON middleware ensures request has valid JSON body
func ValidateJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "application/json" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Content-Type must be application/json",
					"code":    "invalid_content_type",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// RequestSizeLimit middleware limits request body size
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":     "Request body too large",
				"code":      "request_too_large",
				"max_size":  maxSize,
				"your_size": c.Request.ContentLength,
			})
			c.Abort()
			return
		}
		
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}