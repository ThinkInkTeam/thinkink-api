package api

import (
	"log"

	"github.com/ThinkInkTeam/thinkink-core-backend/docs"
	"github.com/ThinkInkTeam/thinkink-core-backend/handlers"
	"github.com/ThinkInkTeam/thinkink-core-backend/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// @title ThinkInk API
// @version 1.0
// @description ThinkInk backend API with Stripe Checkout integration
// @host localhost:8080
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.

// SetupRouter configures the API routes and returns the router
func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(CORSMiddleware())

	// Set up Swagger
	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Stripe-Signature"}
	r.Use(cors.New(config))

	// Public routes
	r.POST("/signin", handlers.SignIn)
	r.POST("/signup", handlers.SignUp)
	
	// Stripe webhook handler - needs to be public to receive Stripe events
	r.POST("/stripe/webhook", handlers.StripeWebhookHandler)

	// Protected routes - require authentication
	authenticated := r.Group("/")
	authenticated.Use(middleware.AuthMiddleware())
	{
		// User routes
		authenticated.GET("/user/:id", handlers.GetUser)
		authenticated.PUT("/user/:id/update", handlers.UpdateUser)

		// File upload route
		authenticated.POST("/upload", handlers.UploadSignalFile)

		// Reports routes
		authenticated.GET("/reports", handlers.GetUserReports)
		authenticated.GET("/reports/sorted", handlers.GetUserReportsSortedByScale)
		
		// Payment routes
		payment := authenticated.Group("/payment")
		{
			// Checkout sessions
			payment.POST("/checkout/subscription", handlers.CreateCheckoutSessionHandler)
			payment.POST("/checkout/one-time", handlers.CreateOneTimeCheckoutHandler)
			
			// Subscription management
			payment.GET("/subscription", handlers.GetSubscriptionHandler)
			payment.POST("/subscription/cancel", handlers.CancelSubscriptionHandler)
		}
	}

	return r
}

// RunServer starts the API server on the specified port
func RunServer(port string) {
	r := SetupRouter()
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}