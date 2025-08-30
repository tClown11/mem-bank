package response

import (
	"time"

	"github.com/gin-gonic/gin"
)

// StandardResponse represents the standard API response format
type StandardResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// APIError represents an API error
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	Type    string      `json:"type,omitempty"`
}

// Meta represents metadata for paginated responses
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
	TotalCount int64 `json:"total_count,omitempty"`
}

// ValidationError represents field validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ErrorType constants
const (
	ErrorTypeValidation     = "validation_error"
	ErrorTypeAuthentication = "authentication_error"
	ErrorTypeAuthorization  = "authorization_error"
	ErrorTypeNotFound       = "not_found_error"
	ErrorTypeInternal       = "internal_error"
	ErrorTypeBadRequest     = "bad_request_error"
	ErrorTypeConflict       = "conflict_error"
	ErrorTypeRateLimit      = "rate_limit_error"
)

// Success sends a successful response
func Success(c *gin.Context, statusCode int, data interface{}) {
	response := StandardResponse{
		Success:   true,
		Data:      data,
		RequestID: getRequestID(c),
		Timestamp: time.Now().UTC(),
	}
	c.JSON(statusCode, response)
}

// SuccessWithMeta sends a successful response with metadata
func SuccessWithMeta(c *gin.Context, statusCode int, data interface{}, meta *Meta) {
	response := StandardResponse{
		Success:   true,
		Data:      data,
		Meta:      meta,
		RequestID: getRequestID(c),
		Timestamp: time.Now().UTC(),
	}
	c.JSON(statusCode, response)
}

// Error sends an error response
func Error(c *gin.Context, statusCode int, code, message string) {
	response := StandardResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Type:    getErrorTypeFromCode(code),
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now().UTC(),
	}
	c.JSON(statusCode, response)
}

// ErrorWithDetails sends an error response with additional details
func ErrorWithDetails(c *gin.Context, statusCode int, code, message string, details interface{}) {
	response := StandardResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
			Type:    getErrorTypeFromCode(code),
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now().UTC(),
	}
	c.JSON(statusCode, response)
}

// ValidationError sends a validation error response
func ValidationErrors(c *gin.Context, statusCode int, errors []ValidationError) {
	response := StandardResponse{
		Success: false,
		Error: &APIError{
			Code:    "validation_failed",
			Message: "Request validation failed",
			Details: errors,
			Type:    ErrorTypeValidation,
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now().UTC(),
	}
	c.JSON(statusCode, response)
}

// InternalError sends an internal server error response
func InternalError(c *gin.Context, message string) {
	response := StandardResponse{
		Success: false,
		Error: &APIError{
			Code:    "internal_server_error",
			Message: message,
			Type:    ErrorTypeInternal,
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now().UTC(),
	}
	c.JSON(500, response)
}

// NotFound sends a not found error response
func NotFound(c *gin.Context, resource string) {
	response := StandardResponse{
		Success: false,
		Error: &APIError{
			Code:    "not_found",
			Message: resource + " not found",
			Type:    ErrorTypeNotFound,
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now().UTC(),
	}
	c.JSON(404, response)
}

// Unauthorized sends an unauthorized error response
func Unauthorized(c *gin.Context, message string) {
	response := StandardResponse{
		Success: false,
		Error: &APIError{
			Code:    "unauthorized",
			Message: message,
			Type:    ErrorTypeAuthentication,
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now().UTC(),
	}
	c.JSON(401, response)
}

// Forbidden sends a forbidden error response
func Forbidden(c *gin.Context, message string) {
	response := StandardResponse{
		Success: false,
		Error: &APIError{
			Code:    "forbidden",
			Message: message,
			Type:    ErrorTypeAuthorization,
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now().UTC(),
	}
	c.JSON(403, response)
}

// BadRequest sends a bad request error response
func BadRequest(c *gin.Context, code, message string) {
	response := StandardResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Type:    ErrorTypeBadRequest,
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now().UTC(),
	}
	c.JSON(400, response)
}

// TooManyRequests sends a rate limit error response
func TooManyRequests(c *gin.Context, message string) {
	response := StandardResponse{
		Success: false,
		Error: &APIError{
			Code:    "rate_limit_exceeded",
			Message: message,
			Type:    ErrorTypeRateLimit,
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now().UTC(),
	}
	c.JSON(429, response)
}

// getRequestID extracts request ID from context
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

// getErrorTypeFromCode maps error codes to error types
func getErrorTypeFromCode(code string) string {
	switch code {
	case "validation_failed", "invalid_request", "missing_field", "invalid_field":
		return ErrorTypeValidation
	case "unauthorized", "invalid_credentials", "missing_auth_header", "invalid_token", "token_expired":
		return ErrorTypeAuthentication
	case "forbidden", "insufficient_permissions":
		return ErrorTypeAuthorization
	case "not_found":
		return ErrorTypeNotFound
	case "conflict", "already_exists":
		return ErrorTypeConflict
	case "rate_limit_exceeded":
		return ErrorTypeRateLimit
	default:
		return ErrorTypeInternal
	}
}
