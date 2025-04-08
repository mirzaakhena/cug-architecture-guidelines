package main

import (
	"fmt"
	"log"
	"net/http"
	"project/utils"
	"project/wiring"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Initialize database
	db, err := initDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Auto migrate database schema
	err = db.AutoMigrate(
	// Put model to auto migrate here ...
	)
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}

	// Create a new HTTP server mux
	mux := http.NewServeMux()

	// Create API printer for documentation
	apiPrinter := utils.NewApiPrinter()

	wiring.SetupWiring(mux, db, apiPrinter)

	apiPrinter.PublishAPI(mux, "http://localhost:8080", "/api/docs")

	// Start the server
	port := 8080
	serverAddr := fmt.Sprintf(":%d", port)
	log.Printf("Server started on http://localhost%s", serverAddr)
	log.Printf("API documentation available at http://localhost%s/api/docs", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, mux))
}

func initDatabase() (*gorm.DB, error) {
	// Open a connection to the SQLite database
	db, err := gorm.Open(sqlite.Open("todo.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate the schema
	err = db.AutoMigrate(
	// Add your models here
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database schema: %w", err)
	}

	return db, nil
}
