package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mem_bank/pkg/logger"
)

// ErrorHandler middleware for centralized error handling
func ErrorHandler(appLogger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			appLogger.WithError(err.Err).WithFields(map[string]interface{}{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
				"ip":     c.ClientIP(),
			}).Error("Request processing error")
			
			// Return appropriate error response
			switch err.Type {
			case gin.ErrorTypeBind:
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid request format",
					"code":    "invalid_request",
					"message": err.Error(),
				})
			case gin.ErrorTypePublic:
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Request validation failed",
					"code":    "validation_error",
					"message": err.Error(),
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal server error",
					"code":    "internal_error",
					"message": "An unexpected error occurred",
				})
			}
		}
	}
}

// Recovery middleware with custom logging
func Recovery(appLogger logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		appLogger.WithFields(map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"ip":     c.ClientIP(),
			"panic":  recovered,
		}).Error("Request panic recovered")
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"code":    "panic_recovered",
			"message": "An unexpected error occurred",
		})
	})
}