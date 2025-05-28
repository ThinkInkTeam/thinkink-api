# thinkink-core-backend

ThinkInk Backend Server - A Go-based backend service for ThinkInk.

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

### Database Configuration

The application uses PostgreSQL defaults and can be configured with the following environment variables:

```bash
# Database configuration
export DB_HOST="localhost"
export DB_USER="postgres"
export DB_PASSWORD="your_db_password"
export DB_NAME="postgres"
export DB_PORT="5432"
export DB_SSL_MODE="disable"  # Use "require" in production
```

### API Server

```bash
# Server port (defaults to 8080)
export PORT="8080"

# JWT secret for authentication tokens
export JWT_SECRET="your_jwt_secret"
```

## Stripe Checkout Integration

ThinkInk includes payment processing capabilities powered by Stripe Checkout, making it simple to accept payments with a secure, hosted payment page.

### Stripe Configuration

Set the following environment variables:

```bash
# Your Stripe API key
export STRIPE_SECRET_KEY="sk_test_your_key_here"

# Your webhook signing secret for verifying Stripe events
export STRIPE_WEBHOOK_SECRET="whsec_your_webhook_secret_here"
```

For development, default test keys are used if these environment variables are not set.

### Features

- **Checkout Sessions**: Create hosted checkout pages for both subscriptions and one-time payments
- **Subscription Management**: View and cancel subscription plans
- **Automatic Updates**: Process Stripe webhook events to keep subscription data updated

### API Endpoints

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

```go
type User struct {
    // Regular user fields
    ID          uint   `json:"id"`
    Name        string `json:"name"`
    Email       string `json:"email"`
    // ... more user fields

    // Stripe-related fields
    StripeCustomerID    string     `json:"stripe_customer_id,omitempty"`
    StripeDefaultPM     string     `json:"stripe_default_payment_method,omitempty"`
    CurrentPlanID       string     `json:"current_plan_id,omitempty"`
    SubscriptionID      string     `json:"subscription_id,omitempty"`
    SubscriptionStatus  string     `json:"subscription_status,omitempty"`
    SubscriptionEndsAt  *time.Time `json:"subscription_ends_at,omitempty"`
}
```

### Testing

For testing, use Stripe's test keys and test cards:

- Test card success: `4242 4242 4242 4242`
- Test card decline: `4000 0000 0000 0002`

For more test cards, see [Stripe's documentation](https://stripe.com/docs/testing).