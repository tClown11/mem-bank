package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"mem_bank/pkg/auth"
)

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

// JWTAuth provides JWT authentication middleware
func JWTAuth(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header required",
				"code":    "missing_auth_header",
			})
			c.Abort()
			return
		}

		// Extract token from header
		token := auth.ExtractTokenFromHeader(authHeader)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid authorization header format",
				"code":    "invalid_auth_format",
			})
			c.Abort()
			return
		}

		// Validate JWT token
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			var code string
			switch {
			case errors.Is(err, auth.ErrExpiredToken):
				code = "token_expired"
			case errors.Is(err, auth.ErrInvalidToken):
				code = "invalid_token"
			default:
				code = "auth_error"
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   err.Error(),
				"code":    code,
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID.String())
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("claims", claims)
		c.Next()
	}
}

// OptionalJWTAuth provides optional JWT authentication
func OptionalJWTAuth(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, continue as unauthenticated user
			c.Next()
			return
		}

		token := auth.ExtractTokenFromHeader(authHeader)
		if token == "" {
			// Invalid format, continue as unauthenticated user
			c.Next()
			return
		}

		// Try to validate token
		claims, err := jwtService.ValidateToken(token)
		if err == nil {
			// Valid token, set user info
			c.Set("user_id", claims.UserID.String())
			c.Set("username", claims.Username)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)
			c.Set("claims", claims)
			c.Set("authenticated", true)
		} else {
			// Invalid or expired token, continue as unauthenticated user
			c.Set("authenticated", false)
		}

		c.Next()
	}
}

// RequireRole requires specific role for access
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Authentication required",
				"code":    "auth_required",
			})
			c.Abort()
			return
		}

		role := userRole.(string)
		for _, requiredRole := range roles {
			if role == requiredRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Insufficient permissions",
			"code":    "insufficient_permissions",
		})
		c.Abort()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, errors.New("user ID not found in context")
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

// GetClaims extracts JWT claims from context
func GetClaims(c *gin.Context) (*auth.Claims, error) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, errors.New("JWT claims not found in context")
	}

	jwtClaims, ok := claims.(*auth.Claims)
	if !ok {
		return nil, errors.New("invalid JWT claims in context")
	}

	return jwtClaims, nil
}

// IsAuthenticated checks if user is authenticated
func IsAuthenticated(c *gin.Context) bool {
	authenticated, exists := c.Get("authenticated")
	if !exists {
		// If not set, check if user_id exists (for backward compatibility)
		_, exists := c.Get("user_id")
		return exists
	}

	return authenticated.(bool)
}
