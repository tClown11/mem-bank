package testutil

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"mem_bank/configs"
	"mem_bank/pkg/database"
)

type TestDB struct {
	Pool   *pgxpool.Pool
	config *configs.DatabaseConfig
}

func SetupTestDB() (*TestDB, error) {
	config := &configs.DatabaseConfig{
		Host:         getEnvWithDefault("TEST_DB_HOST", "localhost"),
		Port:         5432,
		User:         getEnvWithDefault("TEST_DB_USER", "test_user"),
		Password:     getEnvWithDefault("TEST_DB_PASSWORD", "test_password"),
		DBName:       getEnvWithDefault("TEST_DB_NAME", "test_mem_bank"),
		SSLMode:      "disable",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
		MaxLifetime:  5 * time.Minute,
	}

	db, err := database.NewPostgresConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	testDB := &TestDB{
		Pool:   db.Pool,
		config: config,
	}

	if err := testDB.setupSchema(); err != nil {
		return nil, fmt.Errorf("failed to setup test schema: %w", err)
	}

	return testDB, nil
}

func (db *TestDB) setupSchema() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		`CREATE EXTENSION IF NOT EXISTS vector`,
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			username VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			profile JSONB DEFAULT '{}',
			settings JSONB DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			last_login TIMESTAMP WITH TIME ZONE,
			is_active BOOLEAN DEFAULT true
		)`,
		`CREATE TABLE IF NOT EXISTS memories (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			summary TEXT,
			embedding VECTOR(1536),
			importance INTEGER DEFAULT 5 CHECK (importance >= 1 AND importance <= 10),
			memory_type VARCHAR(50) DEFAULT 'general',
			tags TEXT[] DEFAULT '{}',
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			last_accessed TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			access_count INTEGER DEFAULT 0
		)`,
	}

	for _, query := range queries {
		if _, err := db.Pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}

	return nil
}

func (db *TestDB) CleanupTables() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	queries := []string{
		"DELETE FROM memories",
		"DELETE FROM users",
	}

	for _, query := range queries {
		if _, err := db.Pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to cleanup table: %w", err)
		}
	}

	return nil
}

func (db *TestDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}