package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// BasicAuth provides basic authentication middleware
func BasicAuth(username, password string) gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		username: password,
	})
}

// ApiKeyAuth provides API key authentication middleware
func ApiKeyAuth(validApiKeys map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}
		
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key required",
				"code":  "missing_api_key",
			})
			c.Abort()
			return
		}
		
		// Check if API key is valid
		userID, exists := validApiKeys[apiKey]
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
				"code":  "invalid_api_key",
			})
			c.Abort()
			return
		}
		
		// Set user ID in context
		c.Set("user_id", userID)
		c.Next()
	}
}

// OptionalAuth allows both authenticated and unauthenticated access
func OptionalAuth(validApiKeys map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}
		
		if apiKey != "" {
			userID, exists := validApiKeys[apiKey]
			if exists {
				c.Set("user_id", userID)
			}
		}
		
		c.Next()
	}
}

// BearerTokenAuth provides JWT/Bearer token authentication
func BearerTokenAuth(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "missing_auth_header",
			})
			c.Abort()
			return
		}
		
		// Check Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
				"code":  "invalid_auth_format",
			})
			c.Abort()
			return
		}
		
		token := tokenParts[1]
		// TODO: Implement JWT validation here
		// For now, just check if token is not empty
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Empty bearer token",
				"code":  "empty_token",
			})
			c.Abort()
			return
		}
		
		// TODO: Extract user ID from JWT and set in context
		c.Set("token", token)
		c.Next()
	}
}