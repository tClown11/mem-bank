package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"mem_bank/pkg/logger"
)

// PerformanceMonitoring middleware tracks request performance
func PerformanceMonitoring(logger logger.Logger) gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			// Log performance metrics
			logger.WithFields(map[string]interface{}{
				"timestamp":     param.TimeStamp.Format(time.RFC3339),
				"method":        param.Method,
				"path":          param.Path,
				"status_code":   param.StatusCode,
				"latency":       param.Latency.String(),
				"client_ip":     param.ClientIP,
				"user_agent":    param.Request.UserAgent(),
				"response_size": param.BodySize,
				"request_id":    param.Keys["request_id"],
			}).Info("HTTP Request")

			return ""
		},
		Output: gin.DefaultWriter,
	})
}

// ResponseTime middleware adds response time headers
func ResponseTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		// Add response time headers
		c.Header("X-Response-Time", duration.String())
		c.Header("X-Response-Time-Ms", strconv.FormatInt(duration.Nanoseconds()/1000000, 10))
	}
}

// SlowQueryDetector middleware detects slow requests
func SlowQueryDetector(threshold time.Duration, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		if duration > threshold {
			logger.WithFields(map[string]interface{}{
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"duration":   duration.String(),
				"client_ip":  c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
				"request_id": c.GetString("request_id"),
			}).Warn("Slow request detected")
		}
	}
}
