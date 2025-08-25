package configs

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	AI       AIConfig       `mapstructure:"ai"`
	Security SecurityConfig `mapstructure:"security"`
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

func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	err = viper.Unmarshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	setDefaults(config)
	validateConfig(config)

	return config, nil
}

func setDefaults(config *Config) {
	if config.Server.Host == "" {
		config.Server.Host = "localhost"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.Mode == "" {
		config.Server.Mode = "debug"
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 10 * time.Second
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 10 * time.Second
	}
	if config.Server.IdleTimeout == 0 {
		config.Server.IdleTimeout = 60 * time.Second
	}

	if config.Database.Host == "" {
		config.Database.Host = "192.168.64.23"
	}
	if config.Database.Port == 0 {
		config.Database.Port = 30432
	}
	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "disable"
	}
	if config.Database.MaxOpenConns == 0 {
		config.Database.MaxOpenConns = 25
	}
	if config.Database.MaxIdleConns == 0 {
		config.Database.MaxIdleConns = 5
	}
	if config.Database.MaxLifetime == 0 {
		config.Database.MaxLifetime = 5 * time.Minute
	}

	if config.Redis.Host == "" {
		config.Redis.Host = "192.168.64.23"
	}
	if config.Redis.Port == 0 {
		config.Redis.Port = 30379
	}
	if config.Redis.PoolSize == 0 {
		config.Redis.PoolSize = 10
	}
	if config.Redis.Timeout == 0 {
		config.Redis.Timeout = 5 * time.Second
	}

	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}
	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}

	if config.AI.EmbeddingDim == 0 {
		config.AI.EmbeddingDim = 1536
	}
	if config.AI.DefaultThreshold == 0 {
		config.AI.DefaultThreshold = 0.8
	}
	if config.AI.MaxContextLength == 0 {
		config.AI.MaxContextLength = 4000
	}
	if config.AI.DefaultImportance == 0 {
		config.AI.DefaultImportance = 5
	}

	if config.Security.BCryptCost == 0 {
		config.Security.BCryptCost = 12
	}
	if config.Security.RateLimit == 0 {
		config.Security.RateLimit = 100
	}
	if config.Security.JWTExpiry == 0 {
		config.Security.JWTExpiry = 24 * time.Hour
	}
}

func validateConfig(config *Config) error {
	if config.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if config.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}
	if config.Security.JWTSecret == "" {
		config.Security.JWTSecret = os.Getenv("JWT_SECRET")
		if config.Security.JWTSecret == "" {
			return fmt.Errorf("JWT secret is required")
		}
	}
	return nil
}

func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}
