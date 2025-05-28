package main

import (
	"log"
	"os"

	"github.com/ThinkInkTeam/thinkink-core-backend/api"
	"github.com/ThinkInkTeam/thinkink-core-backend/database"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v72"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Initialize database connection using environment variables
	databaseManager := database.NewDatabaseManager()
	
	// Get database configuration from environment variables
	host := getEnvWithDefault("DB_HOST", "localhost")
	user := getEnvWithDefault("DB_USER", "postgres")
	password := getEnvWithDefault("DB_PASSWORD", "postgres")
	dbname := getEnvWithDefault("DB_NAME", "postgres")
	port := getEnvWithDefault("DB_PORT", "5432")
	sslMode := getEnvWithDefault("DB_SSL_MODE", "disable")
	
	if err := databaseManager.Connect(host, user, password, dbname, port, sslMode); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return
	}

	// Initialize Stripe with the API key
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		// Use a default test key for development environments
		stripeKey = "sk_test_example_key_replace_in_production"
		log.Println("Warning: Using default Stripe test key. Set STRIPE_SECRET_KEY environment variable for production.")
	}
	stripe.Key = stripeKey

	// Determine port from environment variable or use default
	port = os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	// Start the API server
	api.RunServer(port)
}

// getEnvWithDefault returns the environment variable value or a default if not set
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
