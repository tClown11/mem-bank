package configs

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestLoadConfigDefaults(t *testing.T) {
	// Clear any existing environment variables
	clearTestEnv()
	defer clearTestEnv()

	// Test with non-existent config file to get true defaults
	viper.Reset()
	config, err := LoadConfig("non_existent_config.yaml")
	if err == nil {
		t.Error("Expected error for missing required config (database user, JWT secret)")
	}

	// Set minimum required environment variables
	os.Setenv("MEM_BANK_DATABASE_USER", "testuser")
	os.Setenv("MEM_BANK_DATABASE_DBNAME", "testdb")
	os.Setenv("MEM_BANK_SECURITY_JWT_SECRET", "this_is_a_very_long_jwt_secret_key_for_testing_purposes_32_chars_plus")
	os.Setenv("MEM_BANK_LLM_API_KEY", "test-api-key-for-testing")

	// Clear viper and reload to ensure clean state
	viper.Reset()

	config, err = LoadConfig("non_existent_config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config with required vars: %v", err)
	}

	// Test default values (not overridden by config file)
	if config.Server.Host != "localhost" {
		t.Errorf("Expected server host 'localhost', got '%s'", config.Server.Host)
	}
	if config.Server.Port != 8080 {
		t.Errorf("Expected server port 8080, got %d", config.Server.Port)
	}
	if config.Database.Port != 5432 {
		t.Errorf("Expected database port 5432, got %d", config.Database.Port)
	}
	if config.AI.EmbeddingDim != 1536 {
		t.Errorf("Expected embedding dim 1536, got %d", config.AI.EmbeddingDim)
	}
}

func TestEnvironmentVariableOverride(t *testing.T) {
	clearTestEnv()
	defer clearTestEnv()
	viper.Reset()

	// Set required and test environment variables
	os.Setenv("MEM_BANK_DATABASE_USER", "testuser")
	os.Setenv("MEM_BANK_DATABASE_DBNAME", "testdb")
	os.Setenv("MEM_BANK_SECURITY_JWT_SECRET", "this_is_a_very_long_jwt_secret_key_for_testing_purposes_32_chars_plus")
	os.Setenv("MEM_BANK_LLM_API_KEY", "test-api-key-for-testing")

	// Override some defaults
	os.Setenv("MEM_BANK_SERVER_HOST", "0.0.0.0")
	os.Setenv("MEM_BANK_SERVER_PORT", "9090")
	os.Setenv("MEM_BANK_DATABASE_HOST", "db.example.com")
	os.Setenv("MEM_BANK_AI_EMBEDDING_DIM", "512")

	config, err := LoadConfig("non_existent_config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test environment variable overrides
	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Expected server host '0.0.0.0', got '%s'", config.Server.Host)
	}
	if config.Server.Port != 9090 {
		t.Errorf("Expected server port 9090, got %d", config.Server.Port)
	}
	if config.Database.Host != "db.example.com" {
		t.Errorf("Expected database host 'db.example.com', got '%s'", config.Database.Host)
	}
	if config.AI.EmbeddingDim != 512 {
		t.Errorf("Expected embedding dim 512, got %d", config.AI.EmbeddingDim)
	}
}

func TestBackwardCompatibilityEnvVars(t *testing.T) {
	clearTestEnv()
	defer clearTestEnv()
	viper.Reset()

	// Set required vars
	os.Setenv("DB_USER", "testuser")                                                                 // Test backward compatibility
	os.Setenv("DB_NAME", "testdb")                                                                   // Test backward compatibility
	os.Setenv("JWT_SECRET", "this_is_a_very_long_jwt_secret_key_for_testing_purposes_32_chars_plus") // Test backward compatibility
	os.Setenv("OPENAI_API_KEY", "test-api-key-for-testing")                                          // Test backward compatibility

	// Use old-style env vars
	os.Setenv("DB_HOST", "old.db.com")
	os.Setenv("REDIS_HOST", "old.redis.com")
	os.Setenv("SERVER_MODE", "release")

	config, err := LoadConfig("non_existent_config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config with old-style env vars: %v", err)
	}

	// Test backward compatibility
	if config.Database.Host != "old.db.com" {
		t.Errorf("Expected database host 'old.db.com', got '%s'", config.Database.Host)
	}
	if config.Redis.Host != "old.redis.com" {
		t.Errorf("Expected redis host 'old.redis.com', got '%s'", config.Redis.Host)
	}
	if config.Server.Mode != "release" {
		t.Errorf("Expected server mode 'release', got '%s'", config.Server.Mode)
	}
}

func TestAllowedOriginsPostProcessing(t *testing.T) {
	clearTestEnv()
	defer clearTestEnv()
	viper.Reset()

	// Set required vars
	os.Setenv("MEM_BANK_DATABASE_USER", "testuser")
	os.Setenv("MEM_BANK_DATABASE_DBNAME", "testdb")
	os.Setenv("MEM_BANK_SECURITY_JWT_SECRET", "this_is_a_very_long_jwt_secret_key_for_testing_purposes_32_chars_plus")
	os.Setenv("MEM_BANK_LLM_API_KEY", "test-api-key-for-testing")

	// Test comma-separated origins
	os.Setenv("MEM_BANK_SECURITY_ALLOWED_ORIGINS", "http://localhost:3000, https://app.example.com, https://admin.example.com")

	config, err := LoadConfig("non_existent_config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	expectedOrigins := []string{"http://localhost:3000", "https://app.example.com", "https://admin.example.com"}
	if len(config.Security.AllowedOrigins) != len(expectedOrigins) {
		t.Errorf("Expected %d origins, got %d", len(expectedOrigins), len(config.Security.AllowedOrigins))
	}

	for i, expected := range expectedOrigins {
		if i >= len(config.Security.AllowedOrigins) || config.Security.AllowedOrigins[i] != expected {
			t.Errorf("Expected origin[%d] '%s', got '%s'", i, expected, config.Security.AllowedOrigins[i])
		}
	}
}

func TestConfigValidation(t *testing.T) {
	clearTestEnv()
	defer clearTestEnv()

	// Test missing required database user
	viper.Reset()
	_, err := LoadConfig("non_existent_config.yaml")
	if err == nil {
		t.Error("Expected error for missing database user")
	}

	// Test missing JWT secret
	clearTestEnv()
	viper.Reset()
	os.Setenv("MEM_BANK_DATABASE_USER", "testuser")
	os.Setenv("MEM_BANK_DATABASE_DBNAME", "testdb")
	os.Setenv("MEM_BANK_LLM_API_KEY", "test-api-key-for-testing") // Provide LLM key so we test JWT validation
	_, err = LoadConfig("non_existent_config.yaml")
	if err == nil {
		t.Error("Expected error for missing JWT secret")
	}

	// Test short JWT secret
	clearTestEnv()
	viper.Reset()
	os.Setenv("MEM_BANK_DATABASE_USER", "testuser")
	os.Setenv("MEM_BANK_DATABASE_DBNAME", "testdb")
	os.Setenv("MEM_BANK_SECURITY_JWT_SECRET", "short")
	os.Setenv("MEM_BANK_LLM_API_KEY", "test-api-key-for-testing")
	_, err = LoadConfig("non_existent_config.yaml")
	if err == nil {
		t.Error("Expected error for short JWT secret")
	}

	// Test invalid origins
	clearTestEnv()
	viper.Reset()
	os.Setenv("MEM_BANK_DATABASE_USER", "testuser")
	os.Setenv("MEM_BANK_DATABASE_DBNAME", "testdb")
	os.Setenv("MEM_BANK_SECURITY_JWT_SECRET", "this_is_a_very_long_jwt_secret_key_for_testing_purposes_32_chars_plus")
	os.Setenv("MEM_BANK_LLM_API_KEY", "test-api-key-for-testing")
	os.Setenv("MEM_BANK_SECURITY_ALLOWED_ORIGINS", "invalid-origin")
	_, err = LoadConfig("non_existent_config.yaml")
	if err == nil {
		t.Error("Expected error for invalid origin format")
	}
}

func TestTimeoutParsing(t *testing.T) {
	clearTestEnv()
	defer clearTestEnv()
	viper.Reset()

	// Set required vars
	os.Setenv("MEM_BANK_DATABASE_USER", "testuser")
	os.Setenv("MEM_BANK_DATABASE_DBNAME", "testdb")
	os.Setenv("MEM_BANK_SECURITY_JWT_SECRET", "this_is_a_very_long_jwt_secret_key_for_testing_purposes_32_chars_plus")
	os.Setenv("MEM_BANK_LLM_API_KEY", "test-api-key-for-testing")

	// Test timeout parsing
	os.Setenv("MEM_BANK_SERVER_READ_TIMEOUT", "30s")
	os.Setenv("MEM_BANK_DATABASE_MAX_LIFETIME", "10m")

	config, err := LoadConfig("non_existent_config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.Server.ReadTimeout != 30*time.Second {
		t.Errorf("Expected read timeout 30s, got %v", config.Server.ReadTimeout)
	}
	if config.Database.MaxLifetime != 10*time.Minute {
		t.Errorf("Expected max lifetime 10m, got %v", config.Database.MaxLifetime)
	}
}

func clearTestEnv() {
	// Clear all possible test environment variables
	envVars := []string{
		// Required vars
		"MEM_BANK_DATABASE_USER", "DB_USER", "DATABASE_USER",
		"MEM_BANK_DATABASE_DBNAME", "DB_NAME", "DATABASE_NAME",
		"MEM_BANK_SECURITY_JWT_SECRET", "JWT_SECRET",
		"MEM_BANK_LLM_API_KEY", "LLM_API_KEY", "OPENAI_API_KEY",

		// Test vars
		"MEM_BANK_SERVER_HOST", "SERVER_HOST",
		"MEM_BANK_SERVER_PORT", "SERVER_PORT",
		"MEM_BANK_DATABASE_HOST", "DB_HOST", "DATABASE_HOST",
		"MEM_BANK_AI_EMBEDDING_DIM",
		"MEM_BANK_REDIS_HOST", "REDIS_HOST",
		"MEM_BANK_SERVER_MODE", "SERVER_MODE", "GIN_MODE",
		"MEM_BANK_SECURITY_ALLOWED_ORIGINS", "ALLOWED_ORIGINS",
		"MEM_BANK_SERVER_READ_TIMEOUT",
		"MEM_BANK_DATABASE_MAX_LIFETIME",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}

	// Reset viper
	viper.Reset()
}
