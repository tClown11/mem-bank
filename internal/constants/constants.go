package constants

import "time"

// Application constants
const (
	// Application metadata
	AppName    = "mem_bank"
	AppVersion = "1.0.0"

	// Default configuration values
	DefaultServerPort   = 8080
	DefaultDBPort       = 5432
	DefaultRedisPort    = 6379
	DefaultQdrantPort   = 6333
	DefaultMaxOpenConns = 25
	DefaultMaxIdleConns = 5
	DefaultRateLimit    = 100
	DefaultBCryptCost   = 12
	DefaultEmbeddingDim = 1536
	DefaultImportance   = 5

	// Timeouts and durations
	DefaultServerReadTimeout  = 10 * time.Second
	DefaultServerWriteTimeout = 10 * time.Second
	DefaultServerIdleTimeout  = 60 * time.Second
	DefaultDBMaxLifetime      = 5 * time.Minute
	DefaultRedisTimeout       = 5 * time.Second
	DefaultJWTExpiry          = 24 * time.Hour
	DefaultLLMTimeout         = 30 * time.Second
	DefaultSlowQueryThreshold = 2 * time.Second

	// Request limits
	DefaultRequestSizeLimit = 10 << 20 // 10MB
	DefaultTextLength       = 8000
	DefaultBatchSize        = 100
	MaxEmailLength          = 254
	MinPasswordLength       = 8
	MinJWTSecretLength      = 32

	// Queue configuration
	DefaultQueueName       = "mem_bank_jobs"
	DefaultMaxRetries      = 3
	DefaultRetryDelay      = 30 * time.Second
	DefaultJobTimeout      = 5 * time.Minute
	DefaultResultTTL       = 24 * time.Hour
	DefaultConcurrency     = 5
	DefaultPollInterval    = 1 * time.Second
	DefaultCleanupInterval = 10 * time.Minute

	// Cache TTL values
	DefaultCacheTTL   = 5 * time.Minute
	UserCacheTTL      = 15 * time.Minute
	MemoryCacheTTL    = 10 * time.Minute
	EmbeddingCacheTTL = 1 * time.Hour
	SearchCacheTTL    = 5 * time.Minute

	// Pagination defaults
	DefaultPageSize = 20
	MaxPageSize     = 100

	// Vector search defaults
	DefaultSimilarityThreshold = 0.8
	DefaultSearchLimit         = 10
	MaxSearchLimit             = 50
)

// HTTP Status Messages
const (
	MessageSuccess          = "Operation completed successfully"
	MessageCreated          = "Resource created successfully"
	MessageUpdated          = "Resource updated successfully"
	MessageDeleted          = "Resource deleted successfully"
	MessageNotFound         = "Resource not found"
	MessageUnauthorized     = "Authentication required"
	MessageForbidden        = "Access forbidden"
	MessageValidationFailed = "Validation failed"
	MessageInternalError    = "Internal server error"
	MessageBadRequest       = "Bad request"
	MessageConflict         = "Resource conflict"
	MessageTooManyRequests  = "Rate limit exceeded"
)

// Error codes
const (
	// Authentication errors
	ErrCodeInvalidCredentials = "invalid_credentials"
	ErrCodeTokenExpired       = "token_expired"
	ErrCodeInvalidToken       = "invalid_token"
	ErrCodeMissingAuthHeader  = "missing_auth_header"
	ErrCodeInvalidAuthFormat  = "invalid_auth_format"

	// Authorization errors
	ErrCodeForbidden         = "forbidden"
	ErrCodeInsufficientPerms = "insufficient_permissions"

	// Validation errors
	ErrCodeValidationFailed = "validation_failed"
	ErrCodeInvalidRequest   = "invalid_request"
	ErrCodeInvalidUUID      = "invalid_uuid"
	ErrCodeMissingUUID      = "missing_uuid"
	ErrCodeInvalidEmail     = "invalid_email"
	ErrCodeWeakPassword     = "weak_password"

	// Resource errors
	ErrCodeNotFound      = "not_found"
	ErrCodeAlreadyExists = "already_exists"
	ErrCodeConflict      = "conflict"

	// System errors
	ErrCodeInternalError = "internal_server_error"
	ErrCodeDatabaseError = "database_error"
	ErrCodeCacheError    = "cache_error"
	ErrCodeQueueError    = "queue_error"
	ErrCodeLLMError      = "llm_error"

	// Request errors
	ErrCodeRequestTooLarge       = "request_too_large"
	ErrCodeInvalidContentType    = "invalid_content_type"
	ErrCodeRateLimitExceeded     = "rate_limit_exceeded"
	ErrCodePotentialSQLInjection = "potential_sql_injection"
	ErrCodePotentialXSS          = "potential_xss"
)

// Job types
const (
	JobTypeGenerateEmbedding = "generate_embedding"
	JobTypeBatchEmbedding    = "batch_embedding"
	JobTypeMemoryUpdate      = "memory_update"
	JobTypeMemoryDelete      = "memory_delete"
	JobTypeUserCleanup       = "user_cleanup"
)

// Job statuses
const (
	JobStatusPending    = "pending"
	JobStatusProcessing = "processing"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
	JobStatusRetrying   = "retrying"
)

// User roles
const (
	RoleUser   = "user"
	RoleAdmin  = "admin"
	RoleSystem = "system"
)

// Content types
const (
	ContentTypeJSON = "application/json"
	ContentTypeText = "text/plain"
	ContentTypeHTML = "text/html"
)

// Header names
const (
	HeaderContentType     = "Content-Type"
	HeaderAuthorization   = "Authorization"
	HeaderAPIKey          = "X-API-Key"
	HeaderRequestID       = "X-Request-ID"
	HeaderResponseTime    = "X-Response-Time"
	HeaderResponseTimeMs  = "X-Response-Time-Ms"
	HeaderRateLimitLimit  = "X-RateLimit-Limit"
	HeaderRateLimitRemain = "X-RateLimit-Remaining"
)

// Security headers
const (
	HeaderXContentTypeOptions = "X-Content-Type-Options"
	HeaderXFrameOptions       = "X-Frame-Options"
	HeaderXXSSProtection      = "X-XSS-Protection"
	HeaderReferrerPolicy      = "Referrer-Policy"
	HeaderCSP                 = "Content-Security-Policy"
	HeaderHSTS                = "Strict-Transport-Security"
)

// Environment names
const (
	EnvDevelopment = "development"
	EnvTesting     = "testing"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

// Logging levels
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)

// Database constraints
const (
	MaxUsernameLength      = 50
	MaxMemoryContentLength = 10000
	MaxMemorySummaryLength = 500
	MaxTagsPerMemory       = 20
	MaxTagLength           = 50
)

// Regular expressions (as constants for reuse)
const (
	EmailRegexPattern    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	UUIDRegexPattern     = `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	UsernameRegexPattern = `^[a-zA-Z0-9_-]{3,30}$`
)

// API versioning
const (
	APIVersion1 = "v1"
	APIPrefix   = "/api"
)

// Context keys (use unexported type to avoid collisions)
type contextKey string

const (
	ContextKeyUserID    contextKey = "user_id"
	ContextKeyUsername  contextKey = "username"
	ContextKeyEmail     contextKey = "email"
	ContextKeyRole      contextKey = "role"
	ContextKeyClaims    contextKey = "claims"
	ContextKeyRequestID contextKey = "request_id"
)

// Cache key prefixes
const (
	CacheKeyUser         = "user"
	CacheKeyMemory       = "memory"
	CacheKeyEmbedding    = "embedding"
	CacheKeyUserMemories = "user_memories"
	CacheKeySimilar      = "similar"
	CacheKeySearch       = "search"
)

// Memory importance levels
const (
	ImportanceVeryLow  = 1
	ImportanceLow      = 3
	ImportanceNormal   = 5
	ImportanceHigh     = 7
	ImportanceVeryHigh = 9
	ImportanceCritical = 10
)

// Vector similarity thresholds
const (
	SimilarityThresholdVeryHigh = 0.95
	SimilarityThresholdHigh     = 0.85
	SimilarityThresholdMedium   = 0.75
	SimilarityThresholdLow      = 0.65
	SimilarityThresholdVeryLow  = 0.55
)

// File size limits
const (
	MaxAvatarSize       = 2 << 20   // 2MB
	MaxDocumentSize     = 10 << 20  // 10MB
	MaxImportFileSize   = 50 << 20  // 50MB
	MaxBatchRequestSize = 100 << 20 // 100MB
)
