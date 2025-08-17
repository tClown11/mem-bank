package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	// Initialize GORM connection
	dsn := "host=192.168.64.23 user=mem_bank_user password=mem_bank_password dbname=mem_bank port=30432 sslmode=disable TimeZone=UTC"
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

	// Apply custom methods to User model
	user := g.GenerateModel("users", gen.FieldType("last_login", "*time.Time"))
	memory := g.GenerateModel("memories", 
		gen.FieldType("last_login", "*time.Time"),
		gen.FieldIgnore("embedding")) // Skip embedding field for now

	g.ApplyBasic(user, memory)

	// Execute the generator
	g.Execute()

	fmt.Println("GORM Gen code generation completed!")
}