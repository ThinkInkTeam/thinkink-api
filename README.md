# thinkink-core-backend

ThinkInk Backend Server - A Go-based backend service for ThinkInk with EEG signal processing capabilities, user management, payment integration, and machine learning services.

## Features

- **User Management**: Registration, authentication, profile management
- **EEG Signal Processing**: File upload, processing, and report generation
- **Payment Integration**: Stripe Checkout for subscriptions and one-time payments
- **Machine Learning Integration**: gRPC services for EEG signal translation and token validation
- **RESTful API**: Comprehensive REST API with Swagger documentation
- **Database Support**: PostgreSQL with GORM ORM
- **Security**: JWT-based authentication with token blacklisting
- **Docker Support**: Containerized deployment

## Prerequisites

Before running the application, ensure you have:

- **Go 1.23.1 or later** - [Installation Guide](https://go.dev/doc/install)
- **PostgreSQL database** - [Installation Guide](https://www.postgresql.org/download/)
- **Protocol Buffers compiler (protoc)** - [Installation Guide](https://protobuf.dev/installation/)
- **Make** - Build automation tool (usually pre-installed on Linux/macOS, [Windows installation](https://gnuwin32.sourceforge.net/packages/make.htm))
- **Git** - Version control system - [Installation Guide](https://git-scm.com/downloads)
- **Docker** (optional, for containerized deployment) - [Installation Guide](https://docs.docker.com/get-docker/)

### Additional Development Tools

For development and code generation, you'll also need:

- **Swag** (for Swagger documentation generation) - Installed automatically via `go install` during setup
- **protoc-gen-go** and **protoc-gen-go-grpc** (for Protocol Buffer Go code generation) - Installed automatically during setup


## Quick Start

1. **Clone the repository**
```bash
git clone https://github.com/ThinkInkTeam/thinkink-core-backend.git
cd thinkink-core-backend
```

2. **Install dependencies**
```bash
go mod download
```

3. **Generate Protocol Buffer files**
```bash
make gen-proto
```

4. **Generate Swagger documentation**
```bash
make gen-docs
```

5. **Start PostgreSQL database**
```bash
make db-start
```

6. **Run the server**
```bash
make run-server
```

Or run everything together:
```bash
make run-all
```

## API Documentation

The API is documented using Swagger. When the server is running, you can access the Swagger UI at:

```
http://localhost:8080/swagger/index.html
```

This provides an interactive documentation where you can:
- Browse all available API endpoints
- See request/response models
- Try out API calls directly from the browser

## Configuration

### Environment Variables

Create a `.env` file in the root directory with the following variables:

#### Database Configuration
```bash
# Database configuration
DB_HOST="localhost"
DB_USER="postgres"
DB_PASSWORD="your_db_password"
DB_NAME="postgres"
DB_PORT="5432"
DB_SSL_MODE="disable"  # Use "require" in production
DB_ENABLE_LOGS="false"  # Set to "true" for SQL query logging
```

#### Server Configuration
```bash
# Server ports
PORT="8080"           # REST API server port
GRPC_PORT="50051"     # gRPC server port

# JWT secret for authentication tokens
JWT_SECRET="your_jwt_secret_key_here"

# Application environment
APP_ENV="development"  # Use "production" for production
```

#### Stripe Configuration
```bash
# Stripe API keys (use your actual keys)
STRIPE_SECRET_KEY="sk_test_your_key_here"
STRIPE_WEBHOOK_SECRET="whsec_your_webhook_secret_here"
```

**Note**: For development, default test keys are used if these environment variables are not set.

### Make Commands

The project includes a Makefile with useful commands:

```bash
# Generate protobuf files
make gen-proto

# Generate Swagger documentation  
make gen-docs

# Database management
make db-start    # Start PostgreSQL container
make db-stop     # Stop PostgreSQL container

# Run application
make run-server  # Run the server only
make run-all     # Stop DB, start fresh DB, then run server
```

## Architecture

ThinkInk Core Backend is a dual-server application that runs both REST API and gRPC services concurrently:

### REST API Server (Port 8080)
- User authentication and management
- File upload and processing
- Payment handling with Stripe
- Report management
- Swagger documentation at `/swagger/index.html`

### gRPC Server (Port 50051)
- **Translation Service**: Converts EEG signals to text using ML models
- **Token Validation Service**: Validates JWT tokens for ML service access

### Key Components

#### Models
- **User**: User accounts with Stripe integration
- **Report**: EEG analysis reports with matching scales
- **SingleFile**: Temporary file storage before processing
- **Token**: JWT token blacklist management

#### Services
- **Translation Client**: Communicates with ML services for EEG processing
- **Token Validator**: Validates user tokens and subscription status

#### Protocol Buffers
- `proto/translation/translation.proto`: EEG translation service definitions
- `proto/validation/validation.proto`: Token validation service definitions

## gRPC Services

### Translation Service
Processes EEG data and converts it to readable text.

**Endpoint**: `localhost:50051`

**Methods**:
- `Translate(TranslateRequest) returns (TranslateResponse)`

**Request Format**:
```protobuf
message TranslateRequest {
  string token = 1;                // JWT authentication token
  repeated EegRow eeg = 2;         // 2D array: list of float32 lists
  repeated float msk = 3;          // 1D array: float32 mask
}
```

### Token Validation Service
Validates JWT tokens for ML service authentication.

**Endpoint**: `localhost:50051`

**Methods**:
- `ValidateMLToken(ValidateTokenRequest) returns (ValidateTokenResponse)`

**Request Format**:
```protobuf
message ValidateTokenRequest {
  string token = 1;  // JWT token to validate
}
```

**Response Format**:
```protobuf
message ValidateTokenResponse {
  bool is_valid = 1;  // Whether the token is valid
}
```

## API Endpoints

### Authentication
- `POST /signup` - User registration
- `POST /signin` - User login
- `POST /logout` - User logout (requires auth)
- `POST /refresh-token` - Refresh JWT token (requires auth)
- `GET /check-auth` - Validate current token (requires auth)
- `POST /forgot-password` - Request password reset
- `POST /reset-password` - Reset password with token
- `POST /validate-ml-token` - Validate token for ML services

### User Management
- `GET /user/{id}` - Get user profile (requires auth)
- `PUT /user/{id}/update` - Update user profile (requires auth)

### File Processing
- `POST /upload` - Upload EEG signal files (requires auth)

### Reports
- `GET /reports` - Get all user reports (requires auth)
- `GET /reports/sorted` - Get reports sorted by matching scale (requires auth)
- `POST /match` - Update report matching scale (requires auth)

### Payment Integration

### Payment Integration (Stripe Checkout)

ThinkInk includes payment processing capabilities powered by Stripe Checkout, making it simple to accept payments with a secure, hosted payment page.

#### Stripe Configuration

Set the following environment variables in your `.env` file:

```bash
# Your Stripe API key
STRIPE_SECRET_KEY="sk_test_your_key_here"

# Your webhook signing secret for verifying Stripe events
STRIPE_WEBHOOK_SECRET="whsec_your_webhook_secret_here"
```

For development, default test keys are used if these environment variables are not set.

#### Payment Features

- **Checkout Sessions**: Create hosted checkout pages for both subscriptions and one-time payments
- **Subscription Management**: View and cancel subscription plans
- **Automatic Updates**: Process Stripe webhook events to keep subscription data updated

#### Payment API Endpoints

#### Checkout Sessions
- `POST /payment/checkout/subscription` - Create a Stripe Checkout session for subscription
- `POST /payment/checkout/one-time` - Create a Stripe Checkout session for one-time payment

#### Subscription Management
- `GET /payment/subscription` - Get the active subscription details
- `POST /payment/subscription/cancel` - Cancel a subscription

#### Webhooks
- `POST /stripe/webhook` - Stripe event webhook (public endpoint)

### Database Integration

Stripe data is directly integrated into the User model to simplify the implementation:


### Testing

For testing, use Stripe's test keys and test cards:

- Test card success: `4242 4242 4242 4242`
- Test card decline: `4000 0000 0000 0002`

For more test cards, see [Stripe's documentation](https://stripe.com/docs/testing).

## Docker Deployment

The application includes a multi-stage Dockerfile for containerized deployment.

### Building the Docker Image

```bash
docker build -t thinkink-backend .
```

### Running with Docker

```bash
# Run the container
docker run -p 8080:8080 -p 50051:50051 \
  -e DB_HOST=your_db_host \
  -e DB_USER=your_db_user \
  -e DB_PASSWORD=your_db_password \
  -e STRIPE_SECRET_KEY=your_stripe_key \
  thinkink-backend
```

### Docker Features

- **Multi-stage build**: Optimized for production deployment
- **Alpine-based runtime**: Small image size
- **Protocol buffer generation**: Automatic protobuf file generation during build
- **Swagger documentation**: Generated during build process

## Development

### Project Structure

```
├── api/                    # REST API router and server setup
├── cmd/                    # Application entry point
├── database/               # Database connection and management
├── docs/                   # Swagger documentation (auto-generated)
├── handlers/               # HTTP request handlers
├── middleware/             # HTTP middleware (auth, CORS, etc.)
├── models/                 # Database models and business logic
├── proto/                  # Protocol Buffer definitions
├── proto-gen/              # Generated Protocol Buffer code
├── services/               # Business services and gRPC clients
├── uploads/                # File upload directory
├── utils/                  # Utility functions
├── dockerfile              # Docker configuration
├── go.mod                  # Go module dependencies
├── makefile               # Build and development commands
└── README.md              # This file
```

### Development Workflow

1. **Make changes** to your code
2. **Regenerate protobuf files** if you modify `.proto` files:
   ```bash
   make gen-proto
   ```
3. **Update Swagger docs** if you modify API endpoints:
   ```bash
   make gen-docs
   ```
4. **Restart the server**:
   ```bash
   make run-server
   ```

### Adding New API Endpoints

1. Add handler function in appropriate file under `handlers/`
2. Add route in `api/server.go`
3. Add Swagger annotations for documentation
4. Run `make gen-docs` to update documentation

### Adding New gRPC Services

1. Define service in `.proto` file under `proto/`
2. Run `make gen-proto` to generate Go code
3. Implement service in `services/` directory
4. Register service in `cmd/main.go`

## Security

- **JWT Authentication**: Secure token-based authentication
- **Token Blacklisting**: Revoked tokens are tracked in database
- **CORS Support**: Configurable cross-origin resource sharing
- **Input Validation**: Request validation using Gin binding
- **Environment Variables**: Sensitive data stored in environment variables

## Dependencies

### Core Dependencies
- **Gin**: HTTP web framework
- **GORM**: ORM for database operations
- **JWT-Go**: JSON Web Token implementation
- **Stripe Go**: Stripe payment processing
- **gRPC**: High-performance RPC framework
- **Protocol Buffers**: Serialization framework

### Development Dependencies
- **Swaggo**: Swagger documentation generation
- **Air** (optional): Live reload for development
- **Docker**: Containerization platform

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Support

For support and questions:
- Create an issue in the GitHub repository
- Check the Swagger documentation at `http://localhost:8080/swagger/index.html`
- Review the protocol buffer definitions in the `proto/` directory