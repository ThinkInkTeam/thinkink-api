package main

import (
	"log"
	"net"
	"os"
	"sync"

	"github.com/ThinkInkTeam/thinkink-core-backend/api"
	"github.com/ThinkInkTeam/thinkink-core-backend/database"
	"github.com/ThinkInkTeam/thinkink-core-backend/services/validation"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v72"
	"google.golang.org/grpc"
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
	restPort := os.Getenv("PORT")
	if restPort == "" {
		restPort = "8080" // Default port
	}

	grpcPort := getEnvWithDefault("GRPC_PORT", "50051")

	// Create a WaitGroup to run both servers concurrently
	var wg sync.WaitGroup
	wg.Add(2)

	// Start the gRPC server in a goroutine
	go func() {
		defer wg.Done()
		startGRPCServer(grpcPort)
	}()

	// Start the REST API server in a goroutine
	go func() {
		defer wg.Done()
		api.RunServer(restPort)
	}()

	log.Printf("Starting servers - REST API on port %s, gRPC on port %s", restPort, grpcPort)

	// Wait for both servers to finish
	wg.Wait()
}

// getEnvWithDefault returns the environment variable value or a default if not set
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// startGRPCServer starts the gRPC validation server
func startGRPCServer(port string) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	validationServer := validation.NewServer()
	validation.RegisterTokenValidationServiceServer(grpcServer, validationServer)

	log.Printf("gRPC server listening on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
