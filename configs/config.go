package configs

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	AI        AIConfig        `mapstructure:"ai"`
	Security  SecurityConfig  `mapstructure:"security"`
	LLM       LLMConfig       `mapstructure:"llm"`
	Queue     QueueConfig     `mapstructure:"queue"`
	Embedding EmbeddingConfig `mapstructure:"embedding"`
	Qdrant    QdrantConfig    `mapstructure:"qdrant"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type DatabaseConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	User         string        `mapstructure:"user"`
	Password     string        `mapstructure:"password"`
	DBName       string        `mapstructure:"dbname"`
	SSLMode      string        `mapstructure:"sslmode"`
	MaxOpenConns int           `mapstructure:"max_open_conns"`
	MaxIdleConns int           `mapstructure:"max_idle_conns"`
	MaxLifetime  time.Duration `mapstructure:"max_lifetime"`
}

type RedisConfig struct {
	Host     string        `mapstructure:"host"`
	Port     int           `mapstructure:"port"`
	Password string        `mapstructure:"password"`
	DB       int           `mapstructure:"db"`
	PoolSize int           `mapstructure:"pool_size"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

type AIConfig struct {
	EmbeddingModel    string  `mapstructure:"embedding_model"`
	EmbeddingDim      int     `mapstructure:"embedding_dim"`
	DefaultThreshold  float64 `mapstructure:"default_threshold"`
	MaxContextLength  int     `mapstructure:"max_context_length"`
	DefaultImportance int     `mapstructure:"default_importance"`
	AutoSummary       bool    `mapstructure:"auto_summary"`
}

type SecurityConfig struct {
	JWTSecret      string        `mapstructure:"jwt_secret"`
	JWTExpiry      time.Duration `mapstructure:"jwt_expiry"`
	BCryptCost     int           `mapstructure:"bcrypt_cost"`
	RateLimit      int           `mapstructure:"rate_limit"`
	AllowedOrigins []string      `mapstructure:"allowed_origins"`
}

type LLMConfig struct {
	Provider        string `mapstructure:"provider"`
	APIKey          string `mapstructure:"api_key"`
	BaseURL         string `mapstructure:"base_url"`
	EmbeddingModel  string `mapstructure:"embedding_model"`
	CompletionModel string `mapstructure:"completion_model"`
	TimeoutSeconds  int    `mapstructure:"timeout_seconds"`
	MaxRetries      int    `mapstructure:"max_retries"`
	RateLimit       int    `mapstructure:"rate_limit"`
}

type QueueConfig struct {
	QueueName           string        `mapstructure:"queue_name"`
	MaxRetries          int           `mapstructure:"max_retries"`
	RetryDelay          time.Duration `mapstructure:"retry_delay"`
	JobTimeout          time.Duration `mapstructure:"job_timeout"`
	ResultTTL           time.Duration `mapstructure:"result_ttl"`
	DefaultConcurrency  int           `mapstructure:"default_concurrency"`
	PollInterval        time.Duration `mapstructure:"poll_interval"`
	CleanupInterval     time.Duration `mapstructure:"cleanup_interval"`
	StatsEnabled        bool          `mapstructure:"stats_enabled"`
	StatsUpdateInterval time.Duration `mapstructure:"stats_update_interval"`
}

type EmbeddingConfig struct {
	MaxTextLength          int  `mapstructure:"max_text_length"`
	CacheEnabled           bool `mapstructure:"cache_enabled"`
	CacheTTLMinutes        int  `mapstructure:"cache_ttl_minutes"`
	BatchSize              int  `mapstructure:"batch_size"`
	NormalizeWhitespace    bool `mapstructure:"normalize_whitespace"`
	ToLowercase            bool `mapstructure:"to_lowercase"`
	RemoveExtraPunctuation bool `mapstructure:"remove_extra_punctuation"`
	ChunkSize              int  `mapstructure:"chunk_size"`
	ChunkOverlap           int  `mapstructure:"chunk_overlap"`
}

type QdrantConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	CollectionName string `mapstructure:"collection_name"`
	VectorSize     int    `mapstructure:"vector_size"`
	UseHTTPS       bool   `mapstructure:"use_https"`
	APIKey         string `mapstructure:"api_key"`
	Enabled        bool   `mapstructure:"enabled"`
}

// LoadConfig loads configuration with proper priority: env vars > config file > defaults
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	// Step 1: Set all default values
	setViperDefaults()

	// Step 2: Setup viper and bind environment variables
	setupViper(configPath)

	// Step 3: Try to read config file (overrides defaults)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; use environment variables and defaults
			fmt.Printf("No config file found, using environment variables and defaults\n")
		} else if strings.Contains(err.Error(), "no such file or directory") {
			// Explicitly specified config file not found
			fmt.Printf("Specified config file not found, using environment variables and defaults\n")
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Step 4: Unmarshal config (env vars automatically override)
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Step 5: Post-process special configurations
	postProcessConfig(config)

	// Step 6: Validate final configuration
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// setViperDefaults sets all default configuration values
func setViperDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", "10s")
	viper.SetDefault("server.write_timeout", "10s")
	viper.SetDefault("server.idle_timeout", "60s")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.max_lifetime", "5m")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.timeout", "5s")

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")
	viper.SetDefault("logging.max_size", 100)
	viper.SetDefault("logging.max_backups", 3)
	viper.SetDefault("logging.max_age", 28)
	viper.SetDefault("logging.compress", true)

	// AI defaults
	viper.SetDefault("ai.embedding_model", "text-embedding-ada-002")
	viper.SetDefault("ai.embedding_dim", 1536)
	viper.SetDefault("ai.default_threshold", 0.8)
	viper.SetDefault("ai.max_context_length", 4000)
	viper.SetDefault("ai.default_importance", 5)
	viper.SetDefault("ai.auto_summary", false)

	// Security defaults
	viper.SetDefault("security.jwt_expiry", "24h")
	viper.SetDefault("security.bcrypt_cost", 12)
	viper.SetDefault("security.rate_limit", 100)
	viper.SetDefault("security.allowed_origins", []string{"*"})

	// LLM defaults
	viper.SetDefault("llm.provider", "openai")
	viper.SetDefault("llm.base_url", "https://api.openai.com/v1")
	viper.SetDefault("llm.embedding_model", "text-embedding-ada-002")
	viper.SetDefault("llm.completion_model", "gpt-3.5-turbo")
	viper.SetDefault("llm.timeout_seconds", 30)
	viper.SetDefault("llm.max_retries", 3)
	viper.SetDefault("llm.rate_limit", 60)

	// Queue defaults
	viper.SetDefault("queue.queue_name", "mem_bank_queue")
	viper.SetDefault("queue.max_retries", 3)
	viper.SetDefault("queue.retry_delay", "5s")
	viper.SetDefault("queue.job_timeout", "300s")
	viper.SetDefault("queue.result_ttl", "3600s")
	viper.SetDefault("queue.default_concurrency", 10)
	viper.SetDefault("queue.poll_interval", "1s")
	viper.SetDefault("queue.cleanup_interval", "3600s")
	viper.SetDefault("queue.stats_enabled", true)
	viper.SetDefault("queue.stats_update_interval", "10s")

	// Embedding defaults
	viper.SetDefault("embedding.max_text_length", 8192)
	viper.SetDefault("embedding.cache_enabled", true)
	viper.SetDefault("embedding.cache_ttl_minutes", 60)
	viper.SetDefault("embedding.batch_size", 100)
	viper.SetDefault("embedding.normalize_whitespace", true)
	viper.SetDefault("embedding.to_lowercase", true)
	viper.SetDefault("embedding.remove_extra_punctuation", true)
	viper.SetDefault("embedding.chunk_size", 1000)
	viper.SetDefault("embedding.chunk_overlap", 100)

	// Qdrant defaults
	viper.SetDefault("qdrant.host", "localhost")
	viper.SetDefault("qdrant.port", 6333)
	viper.SetDefault("qdrant.collection_name", "mem_bank_vectors")
	viper.SetDefault("qdrant.vector_size", 1536)
	viper.SetDefault("qdrant.use_https", false)
	viper.SetDefault("qdrant.enabled", true)
}

// setupViper configures viper for reading configuration
func setupViper(configPath string) {
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath("/etc/mem_bank/")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Enable reading from environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("MEM_BANK")

	// Replace dots and dashes with underscores for env vars
	replacer := strings.NewReplacer(".", "_", "-", "_")
	viper.SetEnvKeyReplacer(replacer)

	// Bind specific environment variables
	bindEnvironmentVariables()
}

// bindEnvironmentVariables binds specific environment variables for backward compatibility
// and to support common naming conventions (e.g., DB_HOST vs MEM_BANK_DATABASE_HOST)
func bindEnvironmentVariables() {
	// Server config - support both prefixed and unprefixed versions
	viper.BindEnv("server.host", "MEM_BANK_SERVER_HOST", "SERVER_HOST")
	viper.BindEnv("server.port", "MEM_BANK_SERVER_PORT", "SERVER_PORT")
	viper.BindEnv("server.mode", "MEM_BANK_SERVER_MODE", "SERVER_MODE", "GIN_MODE")
	viper.BindEnv("server.read_timeout", "MEM_BANK_SERVER_READ_TIMEOUT", "SERVER_READ_TIMEOUT")
	viper.BindEnv("server.write_timeout", "MEM_BANK_SERVER_WRITE_TIMEOUT", "SERVER_WRITE_TIMEOUT")
	viper.BindEnv("server.idle_timeout", "MEM_BANK_SERVER_IDLE_TIMEOUT", "SERVER_IDLE_TIMEOUT")

	// Database config - support common database env vars
	viper.BindEnv("database.host", "MEM_BANK_DATABASE_HOST", "DB_HOST", "DATABASE_HOST")
	viper.BindEnv("database.port", "MEM_BANK_DATABASE_PORT", "DB_PORT", "DATABASE_PORT")
	viper.BindEnv("database.user", "MEM_BANK_DATABASE_USER", "DB_USER", "DATABASE_USER")
	viper.BindEnv("database.password", "MEM_BANK_DATABASE_PASSWORD", "DB_PASSWORD", "DATABASE_PASSWORD")
	viper.BindEnv("database.dbname", "MEM_BANK_DATABASE_DBNAME", "DB_NAME", "DATABASE_NAME")
	viper.BindEnv("database.sslmode", "MEM_BANK_DATABASE_SSLMODE", "DB_SSLMODE", "DATABASE_SSLMODE")
	viper.BindEnv("database.max_open_conns", "MEM_BANK_DATABASE_MAX_OPEN_CONNS", "DB_MAX_OPEN_CONNS")
	viper.BindEnv("database.max_idle_conns", "MEM_BANK_DATABASE_MAX_IDLE_CONNS", "DB_MAX_IDLE_CONNS")
	viper.BindEnv("database.max_lifetime", "MEM_BANK_DATABASE_MAX_LIFETIME", "DB_MAX_LIFETIME")

	// Redis config
	viper.BindEnv("redis.host", "MEM_BANK_REDIS_HOST", "REDIS_HOST")
	viper.BindEnv("redis.port", "MEM_BANK_REDIS_PORT", "REDIS_PORT")
	viper.BindEnv("redis.password", "MEM_BANK_REDIS_PASSWORD", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "MEM_BANK_REDIS_DB", "REDIS_DB")
	viper.BindEnv("redis.pool_size", "MEM_BANK_REDIS_POOL_SIZE", "REDIS_POOL_SIZE")
	viper.BindEnv("redis.timeout", "MEM_BANK_REDIS_TIMEOUT", "REDIS_TIMEOUT")

	// Logging config
	viper.BindEnv("logging.level", "MEM_BANK_LOGGING_LEVEL", "LOG_LEVEL")
	viper.BindEnv("logging.format", "MEM_BANK_LOGGING_FORMAT", "LOG_FORMAT")
	viper.BindEnv("logging.output", "MEM_BANK_LOGGING_OUTPUT", "LOG_OUTPUT")

	// AI config
	viper.BindEnv("ai.embedding_model", "MEM_BANK_AI_EMBEDDING_MODEL")
	viper.BindEnv("ai.embedding_dim", "MEM_BANK_AI_EMBEDDING_DIM")
	viper.BindEnv("ai.default_threshold", "MEM_BANK_AI_DEFAULT_THRESHOLD")

	// Security config
	viper.BindEnv("security.jwt_secret", "MEM_BANK_SECURITY_JWT_SECRET", "JWT_SECRET")
	viper.BindEnv("security.jwt_expiry", "MEM_BANK_SECURITY_JWT_EXPIRY", "JWT_EXPIRY")
	viper.BindEnv("security.bcrypt_cost", "MEM_BANK_SECURITY_BCRYPT_COST")
	viper.BindEnv("security.rate_limit", "MEM_BANK_SECURITY_RATE_LIMIT", "RATE_LIMIT")
	viper.BindEnv("security.allowed_origins", "MEM_BANK_SECURITY_ALLOWED_ORIGINS", "ALLOWED_ORIGINS")

	// LLM config - support OpenAI-compatible env vars
	viper.BindEnv("llm.provider", "MEM_BANK_LLM_PROVIDER", "LLM_PROVIDER")
	viper.BindEnv("llm.api_key", "MEM_BANK_LLM_API_KEY", "LLM_API_KEY", "OPENAI_API_KEY")
	viper.BindEnv("llm.base_url", "MEM_BANK_LLM_BASE_URL", "LLM_BASE_URL", "OPENAI_BASE_URL")
	viper.BindEnv("llm.embedding_model", "MEM_BANK_LLM_EMBEDDING_MODEL", "LLM_EMBEDDING_MODEL")
	viper.BindEnv("llm.completion_model", "MEM_BANK_LLM_COMPLETION_MODEL", "LLM_COMPLETION_MODEL")
	viper.BindEnv("llm.timeout_seconds", "MEM_BANK_LLM_TIMEOUT_SECONDS", "LLM_TIMEOUT_SECONDS")
	viper.BindEnv("llm.max_retries", "MEM_BANK_LLM_MAX_RETRIES", "LLM_MAX_RETRIES")
	viper.BindEnv("llm.rate_limit", "MEM_BANK_LLM_RATE_LIMIT", "LLM_RATE_LIMIT")

	// Queue config
	viper.BindEnv("queue.queue_name", "MEM_BANK_QUEUE_QUEUE_NAME", "QUEUE_NAME")
	viper.BindEnv("queue.max_retries", "MEM_BANK_QUEUE_MAX_RETRIES")
	viper.BindEnv("queue.default_concurrency", "MEM_BANK_QUEUE_DEFAULT_CONCURRENCY")

	// Embedding config
	viper.BindEnv("embedding.max_text_length", "MEM_BANK_EMBEDDING_MAX_TEXT_LENGTH")
	viper.BindEnv("embedding.cache_enabled", "MEM_BANK_EMBEDDING_CACHE_ENABLED")
	viper.BindEnv("embedding.batch_size", "MEM_BANK_EMBEDDING_BATCH_SIZE")

	// Qdrant config
	viper.BindEnv("qdrant.host", "MEM_BANK_QDRANT_HOST", "QDRANT_HOST")
	viper.BindEnv("qdrant.port", "MEM_BANK_QDRANT_PORT", "QDRANT_PORT")
	viper.BindEnv("qdrant.api_key", "MEM_BANK_QDRANT_API_KEY", "QDRANT_API_KEY")
	viper.BindEnv("qdrant.enabled", "MEM_BANK_QDRANT_ENABLED", "QDRANT_ENABLED")
	viper.BindEnv("qdrant.collection_name", "MEM_BANK_QDRANT_COLLECTION_NAME", "QDRANT_COLLECTION_NAME")
	viper.BindEnv("qdrant.vector_size", "MEM_BANK_QDRANT_VECTOR_SIZE", "QDRANT_VECTOR_SIZE")
	viper.BindEnv("qdrant.use_https", "MEM_BANK_QDRANT_USE_HTTPS", "QDRANT_USE_HTTPS")
}

// postProcessConfig handles special configuration processing that requires custom logic
func postProcessConfig(config *Config) {
	// Parse comma-separated allowed origins from environment or config
	if originsStr := viper.GetString("security.allowed_origins"); originsStr != "" {
		// Handle case where allowed_origins is provided as a comma-separated string
		if strings.Contains(originsStr, ",") {
			origins := strings.Split(originsStr, ",")
			for i, origin := range origins {
				origins[i] = strings.TrimSpace(origin)
			}
			config.Security.AllowedOrigins = origins
		}
	}
}

// validateConfig validates the final configuration
func validateConfig(config *Config) error {
	// Database validation
	if config.Database.User == "" {
		return fmt.Errorf("database user is required (set MEM_BANK_DATABASE_USER environment variable or config)")
	}
	if config.Database.DBName == "" {
		return fmt.Errorf("database name is required (set MEM_BANK_DATABASE_DBNAME environment variable or config)")
	}

	// Security validation
	if config.Security.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required (set MEM_BANK_SECURITY_JWT_SECRET environment variable or config)")
	}

	// Validate JWT secret strength
	if len(config.Security.JWTSecret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long for security")
	}

	// LLM validation (if provider is set)
	if config.LLM.Provider != "" {
		if config.LLM.APIKey == "" {
			return fmt.Errorf("LLM API key is required when provider is specified (set MEM_BANK_LLM_API_KEY environment variable)")
		}
	}

	// Rate limit validation
	if config.Security.RateLimit <= 0 {
		return fmt.Errorf("invalid rate limit: %d (must be positive)", config.Security.RateLimit)
	}

	// Validate allowed origins format
	for _, origin := range config.Security.AllowedOrigins {
		if origin != "*" && !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
			return fmt.Errorf("invalid origin format: %s (must start with http:// or https://)", origin)
		}
	}

	// Validate timeout values
	if config.Server.ReadTimeout <= 0 {
		return fmt.Errorf("invalid server read timeout: %v (must be positive)", config.Server.ReadTimeout)
	}
	if config.Server.WriteTimeout <= 0 {
		return fmt.Errorf("invalid server write timeout: %v (must be positive)", config.Server.WriteTimeout)
	}

	// Validate database connection pool settings
	if config.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("invalid max open connections: %d (must be positive)", config.Database.MaxOpenConns)
	}
	if config.Database.MaxIdleConns < 0 {
		return fmt.Errorf("invalid max idle connections: %d (must be non-negative)", config.Database.MaxIdleConns)
	}
	if config.Database.MaxIdleConns > config.Database.MaxOpenConns {
		return fmt.Errorf("max idle connections (%d) cannot exceed max open connections (%d)",
			config.Database.MaxIdleConns, config.Database.MaxOpenConns)
	}

	return nil
}
