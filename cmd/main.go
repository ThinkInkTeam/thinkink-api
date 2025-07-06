package main

import (
	"log"
	"net"
	"sync"

	"github.com/ThinkInkTeam/thinkink-core-backend/api"
	"github.com/ThinkInkTeam/thinkink-core-backend/database"
	pb "github.com/ThinkInkTeam/thinkink-core-backend/proto-gen/proto/validation"
	"github.com/ThinkInkTeam/thinkink-core-backend/services/validation"
	"github.com/ThinkInkTeam/thinkink-core-backend/utils"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v72"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	_ = godotenv.Load()
	// Initialize database connection using environment variables
	databaseManager := database.NewDatabaseManager()

	// Get database configuration from environment variables
	host := utils.GetEnvWithDefault("DB_HOST", "localhost")
	user := utils.GetEnvWithDefault("DB_USER", "postgres")
	password := utils.GetEnvWithDefault("DB_PASSWORD", "postgres")
	dbname := utils.GetEnvWithDefault("DB_NAME", "postgres")
	port := utils.GetEnvWithDefault("DB_PORT", "5432")
	sslMode := utils.GetEnvWithDefault("DB_SSL_MODE", "disable")

	if err := databaseManager.Connect(host, user, password, dbname, port, sslMode); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return
	}

	// Initialize Stripe with the API key
	stripeKey := utils.GetEnvWithDefault("STRIPE_SECRET_KEY", "sk_test_example_key_replace_in_production")
	if stripeKey == "sk_test_example_key_replace_in_production" {
		log.Println("Warning: Using default Stripe test key. Set STRIPE_SECRET_KEY environment variable for production.")
	}
	stripe.Key = stripeKey

	// Determine port from environment variable or use default
	restPort := utils.GetEnvWithDefault("PORT", "8080")

	grpcPort := utils.GetEnvWithDefault("GRPC_PORT", "50051")

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

// startGRPCServer starts the gRPC validation server
func startGRPCServer(port string) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	validationServer := validation.NewServer()
	pb.RegisterTokenValidationServiceServer(grpcServer, validationServer)

	if utils.GetEnvWithDefault("APP_ENV", "development") != "production" {
		reflection.Register(grpcServer)
	}

	log.Printf("gRPC server listening on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
