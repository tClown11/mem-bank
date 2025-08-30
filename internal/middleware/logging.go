package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"mem_bank/pkg/logger"
)

// RequestLogger middleware for structured request logging
func RequestLogger(appLogger logger.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		// Log to structured logger instead of stdout
		fields := map[string]interface{}{
			"timestamp":   params.TimeStamp.Format(time.RFC3339),
			"method":      params.Method,
			"path":        params.Path,
			"status_code": params.StatusCode,
			"latency":     params.Latency.String(),
			"client_ip":   params.ClientIP,
			"user_agent":  params.Request.UserAgent(),
			"body_size":   params.BodySize,
		}

		if params.ErrorMessage != "" {
			fields["error"] = params.ErrorMessage
		}

		// Log with appropriate level based on status code
		switch {
		case params.StatusCode >= 500:
			appLogger.WithFields(fields).Error("HTTP request completed with server error")
		case params.StatusCode >= 400:
			appLogger.WithFields(fields).Warn("HTTP request completed with client error")
		default:
			appLogger.WithFields(fields).Info("HTTP request completed")
		}

		return "" // Return empty string since we're logging elsewhere
	})
}

// RequestID middleware adds a unique ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate a simple request ID
			requestID = generateRequestID()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Simple request ID generator (you might want to use a more sophisticated one)
func generateRequestID() string {
	// For now, use timestamp + random suffix
	return time.Now().Format("20060102150405") + "-" + generateRandomString(6)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
