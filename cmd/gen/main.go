package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	// Initialize GORM connection from environment variables
	host := getEnvWithDefault("GEN_DB_HOST", "localhost")
	port := getEnvWithDefault("GEN_DB_PORT", "5432")
	user := getEnvWithDefault("GEN_DB_USER", "mem_bank_user")
	password := getEnvWithDefault("GEN_DB_PASSWORD", "mem_bank_password")
	dbname := getEnvWithDefault("GEN_DB_NAME", "mem_bank")
	sslmode := getEnvWithDefault("GEN_DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		host, port, user, password, dbname, sslmode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create generator
	g := gen.NewGenerator(gen.Config{
		OutPath:       "./internal/query",
		Mode:          gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable: true,
	})

	// Use the connection
	g.UseDB(db)

	// Apply basic types to the generator
	g.ApplyBasic(
		// Generate structs for all tables
		g.GenerateAllTable()...,
	)

	// Apply custom methods to specific models
	g.ApplyInterface(func() {}, g.GenerateModel("users"), g.GenerateModel("memories"))

	// Execute the generator
	g.Execute()

	fmt.Println("GORM Gen code generation completed!")
}

// getEnvWithDefault gets environment variable with default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
