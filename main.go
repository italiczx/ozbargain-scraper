package main

import (
	"log"
	"net/http"
	"os"

	"ozbarginscraper.com/handlers"

	"github.com/joho/godotenv"
)

func main() {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	
	// Register routes
	mux.HandleFunc("/", handlers.HealthHandler)
	mux.HandleFunc("/healthz", handlers.HealthHandler)
	
	// Start scheduled scrapers in background
	handlers.StartScheduledScrapers()
	
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Starting server on port %s", port)
	
	// Start the HTTP server
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}