package main

import (
	"log"
	"os"

	"github.com/bookaroo/bookaroo-platform-be/config"
	"github.com/bookaroo/bookaroo-platform-be/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	// Initialize database
	db := config.InitDB()

	// Create a new Gin router
	r := gin.Default()

	// Initialize routes
	routes.SetupRoutes(r, db)

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	r.Run(":" + port)
}
