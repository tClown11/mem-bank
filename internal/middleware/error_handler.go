package middleware

import (
	"github.com/gin-gonic/gin"
	"mem_bank/pkg/logger"
	"mem_bank/pkg/response"
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

			// Return appropriate error response using standardized format
			switch err.Type {
			case gin.ErrorTypeBind:
				response.BadRequest(c, "invalid_request", "Invalid request format: "+err.Error())
			case gin.ErrorTypePublic:
				response.BadRequest(c, "validation_error", "Request validation failed: "+err.Error())
			default:
				response.InternalError(c, "An unexpected error occurred")
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

		response.InternalError(c, "An unexpected error occurred during request processing")
	})
}
