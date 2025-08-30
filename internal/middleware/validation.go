package middleware

import (
	"html"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"mem_bank/pkg/response"
)

// ValidateUUID middleware validates UUID parameters
func ValidateUUID(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuidStr := c.Param(paramName)
		if uuidStr == "" {
			response.BadRequest(c, "missing_uuid", "UUID parameter '"+paramName+"' is required")
			c.Abort()
			return
		}

		if _, err := uuid.Parse(uuidStr); err != nil {
			response.BadRequest(c, "invalid_uuid", "Invalid UUID format for parameter '"+paramName+"'")
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
			// Check for JSON content type (allow charset specification)
			if !strings.HasPrefix(contentType, "application/json") {
				response.BadRequest(c, "invalid_content_type", "Content-Type must be application/json")
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
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusRequestEntityTooLarge, response.StandardResponse{
				Success: false,
				Error: &response.APIError{
					Code:    "request_too_large",
					Message: "Request body too large",
					Type:    "bad_request_error",
				},
				RequestID: getRequestID(c),
				Timestamp: time.Now().UTC(),
			})
			c.Abort()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// getRequestID extracts request ID from context (helper function)
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

// XSSProtection middleware provides XSS protection by sanitizing inputs
func XSSProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		for key, values := range c.Request.URL.Query() {
			for i, value := range values {
				c.Request.URL.Query()[key][i] = html.EscapeString(value)
			}
		}

		// Add XSS protection headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// SQLInjectionProtection middleware provides basic SQL injection protection
func SQLInjectionProtection() gin.HandlerFunc {
	// Common SQL injection patterns
	sqlInjectionPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union|select|insert|delete|update|drop|create|alter|exec|execute)`),
		regexp.MustCompile(`(?i)(or|and)\s+\d+\s*=\s*\d+`),
		regexp.MustCompile(`(?i)'.*?('|--|#)`),
		regexp.MustCompile(`(?i)(script|javascript|vbscript|onload|onerror)`),
	}

	return func(c *gin.Context) {
		// Check query parameters
		for _, values := range c.Request.URL.Query() {
			for _, value := range values {
				if containsSQLInjection(value, sqlInjectionPatterns) {
					response.BadRequest(c, "potential_sql_injection", "Potentially malicious input detected")
					c.Abort()
					return
				}
			}
		}

		// Check path parameters
		for _, param := range c.Params {
			if containsSQLInjection(param.Value, sqlInjectionPatterns) {
				response.BadRequest(c, "potential_sql_injection", "Potentially malicious input detected")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// containsSQLInjection checks if input contains potential SQL injection
func containsSQLInjection(input string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// RateLimit middleware provides basic rate limiting
func RateLimit(maxRequests int, windowSeconds int) gin.HandlerFunc {
	// Simple in-memory rate limiting (for production, use Redis)
	requests := make(map[string][]time.Time)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()
		window := time.Duration(windowSeconds) * time.Second

		// Clean old requests outside the window
		if clientRequests, exists := requests[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range clientRequests {
				if now.Sub(reqTime) < window {
					validRequests = append(validRequests, reqTime)
				}
			}
			requests[clientIP] = validRequests
		}

		// Check if rate limit exceeded
		if len(requests[clientIP]) >= maxRequests {
			c.Header("X-RateLimit-Limit", strconv.Itoa(maxRequests))
			c.Header("X-RateLimit-Remaining", "0")
			response.TooManyRequests(c, "Rate limit exceeded")
			c.Abort()
			return
		}

		// Add current request
		requests[clientIP] = append(requests[clientIP], now)

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(maxRequests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(maxRequests-len(requests[clientIP])))

		c.Next()
	}
}

// SecurityHeaders middleware adds security headers to all responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// XSS Protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (basic)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' https:; connect-src 'self' https:")

		// HTTP Strict Transport Security (for HTTPS)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		c.Next()
	}
}
