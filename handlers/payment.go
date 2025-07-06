package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ThinkInkTeam/thinkink-core-backend/database"
	"github.com/ThinkInkTeam/thinkink-core-backend/models"
	"github.com/ThinkInkTeam/thinkink-core-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/sub"
	"github.com/stripe/stripe-go/v72/webhook"
)

// CreateCheckoutSessionRequest represents the request body for creating a checkout session
type CreateCheckoutSessionRequest struct {
	PlanID     string `json:"plan_id" binding:"required" example:"price_1Oxy3JExamplePriceID"`
	SuccessURL string `json:"success_url" binding:"required" example:"https://yourapp.com/success?session_id={CHECKOUT_SESSION_ID}"`
	CancelURL  string `json:"cancel_url" binding:"required" example:"https://yourapp.com/cancel"`
}

// CreateOneTimeCheckoutRequest represents the request body for one-time checkout
type CreateOneTimeCheckoutRequest struct {
	Amount      int64  `json:"amount" binding:"required" example:"2000"` // Amount in cents, e.g., 2000 = $20.00
	Currency    string `json:"currency" binding:"required" example:"usd"`
	ProductName string `json:"product_name" binding:"required" example:"Premium Report"`
	SuccessURL  string `json:"success_url" binding:"required" example:"https://yourapp.com/success?session_id={CHECKOUT_SESSION_ID}"`
	CancelURL   string `json:"cancel_url" binding:"required" example:"https://yourapp.com/cancel"`
}

// CheckoutResponse is the response returned for checkout session creation
type CheckoutResponse struct {
	SessionID string `json:"sessionId" example:"cs_test_a1b2c3d4e5f6g7h8i9j0"`
	URL       string `json:"url" example:"https://checkout.stripe.com/pay/cs_test_a1b2c3d4e5f6g7h8i9j0"`
}

// SubscriptionResponse represents a subscription response
type SubscriptionResponse struct {
	HasSubscription   bool       `json:"has_subscription" example:"true"`
	SubscriptionID    string     `json:"subscription_id,omitempty" example:"sub_12345"`
	PlanID            string     `json:"plan_id,omitempty" example:"price_1Oxy3JExamplePriceID"`
	Status            string     `json:"status,omitempty" example:"active"`
	CancelAtPeriodEnd bool       `json:"cancel_at_period_end,omitempty" example:"false"`
	CurrentPeriodEnd  *time.Time `json:"current_period_end,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Error message"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

// CancelSubscriptionResponse represents the response when canceling a subscription
type CancelSubscriptionResponse struct {
	Message      string              `json:"message" example:"Subscription will be canceled at the end of the current billing period"`
	Subscription SubscriptionDetails `json:"subscription"`
}

// SubscriptionDetails represents details about a subscription
type SubscriptionDetails struct {
	ID                string    `json:"id" example:"sub_12345"`
	Status            string    `json:"status" example:"active"`
	CancelAtPeriodEnd bool      `json:"cancel_at_period_end" example:"true"`
	CurrentPeriodEnd  time.Time `json:"current_period_end"`
}

// WebhookResponse represents a response for webhook processing
type WebhookResponse struct {
	Received bool `json:"received" example:"true"`
}

// CreateCheckoutSessionHandler creates a Stripe Checkout session for subscription
// @Summary Create a subscription checkout session
// @Description Creates a Stripe checkout session for subscription payments
// @Tags payment
// @Accept json
// @Produce json
// @Param request body CreateCheckoutSessionRequest true "Checkout session details"
// @Success 200 {object} CheckoutResponse "Checkout session created"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /payment/checkout/subscription [post]
func CreateCheckoutSessionHandler(c *gin.Context) {
	// Parse request
	var req CreateCheckoutSessionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Get authenticated user from context
	userID := c.GetUint("userID")

	// Get user from database
	db := database.DB
	user, err := models.FindUserByID(db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	// Create or retrieve customer
	var customerID string
	if user.StripeCustomerID != nil {
		customerID = *user.StripeCustomerID
	} else {
		// Create new customer in Stripe
		customerParams := user.ToStripeCustomerParams()
		newCustomer, err := customer.New(customerParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Error creating Stripe customer: %v", err)})
			return
		}

		// Update user with Stripe customer ID
		if err := user.UpdateStripeData(db, newCustomer.ID, ""); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Error updating user data: %v", err)})
			return
		}

		customerID = newCustomer.ID
	}

	// Create checkout session params
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(req.PlanID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(req.SuccessURL),
		CancelURL:  stripe.String(req.CancelURL),
	}

	// Add metadata to identify user in webhook
	params.AddMetadata("user_id", fmt.Sprintf("%d", user.ID))
	params.AddMetadata("plan_id", req.PlanID)

	sess, err := session.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Error creating checkout session: %v", err)})
		return
	}

	c.JSON(http.StatusOK, CheckoutResponse{
		SessionID: sess.ID,
		URL:       sess.URL,
	})
}

// CreateOneTimeCheckoutHandler creates a Stripe Checkout session for one-time payment
// @Summary Create a one-time payment checkout session
// @Description Creates a Stripe checkout session for one-time payments
// @Tags payment
// @Accept json
// @Produce json
// @Param request body CreateOneTimeCheckoutRequest true "One-time checkout details"
// @Success 200 {object} CheckoutResponse "Checkout session created"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /payment/checkout/one-time [post]
func CreateOneTimeCheckoutHandler(c *gin.Context) {
	// Parse request
	var req CreateOneTimeCheckoutRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Get authenticated user from context
	userID := c.GetUint("userID")

	// Get user from database
	db := database.DB
	user, err := models.FindUserByID(db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	// Create or retrieve customer
	var customerID string
	if user.StripeCustomerID != nil {
		customerID = *user.StripeCustomerID
	} else {
		// Create new customer in Stripe
		customerParams := user.ToStripeCustomerParams()
		newCustomer, err := customer.New(customerParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Error creating Stripe customer: %v", err)})
			return
		}

		// Update user with Stripe customer ID
		if err := user.UpdateStripeData(db, newCustomer.ID, ""); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Error updating user data: %v", err)})
			return
		}

		customerID = newCustomer.ID
	}

	// Create checkout session params
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(req.Currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(req.ProductName),
					},
					UnitAmount: stripe.Int64(req.Amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(req.SuccessURL),
		CancelURL:  stripe.String(req.CancelURL),
	}

	// Add metadata to identify user in webhook
	params.AddMetadata("user_id", fmt.Sprintf("%d", user.ID))

	sess, err := session.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Error creating checkout session: %v", err)})
		return
	}

	c.JSON(http.StatusOK, CheckoutResponse{
		SessionID: sess.ID,
		URL:       sess.URL,
	})
}

// CancelSubscriptionHandler cancels a subscription at the end of the current period
// @Summary Cancel a subscription
// @Description Cancels the user's subscription at the end of the current billing period
// @Tags payment
// @Accept json
// @Produce json
// @Success 200 {object} CancelSubscriptionResponse "Subscription canceled"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /payment/subscription/cancel [post]
func CancelSubscriptionHandler(c *gin.Context) {
	// Get authenticated user from context
	userID := c.GetUint("userID")

	// Get user from database
	db := database.DB
	user, err := models.FindUserByID(db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	// Check if user has an active subscription
	if user.SubscriptionID == nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "No active subscription found"})
		return
	}

	// Cancel the subscription at period end
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}

	// Make the API call to cancel
	subscription, err := sub.Update(*user.SubscriptionID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Error canceling subscription: %v", err)})
		return
	}

	// Update subscription status in database
	periodEnd := time.Unix(subscription.CurrentPeriodEnd, 0)
	if err := user.UpdateSubscriptionData(db, subscription.ID, *user.CurrentPlanID, string(subscription.Status), &periodEnd); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Error updating subscription data: %v", err)})
		return
	}

	c.JSON(http.StatusOK, CancelSubscriptionResponse{
		Message: "Subscription will be canceled at the end of the current billing period",
		Subscription: SubscriptionDetails{
			ID:                subscription.ID,
			Status:            string(subscription.Status),
			CancelAtPeriodEnd: subscription.CancelAtPeriodEnd,
			CurrentPeriodEnd:  time.Unix(subscription.CurrentPeriodEnd, 0),
		},
	})
}

// GetSubscriptionHandler gets the current subscription status
// @Summary Get subscription details
// @Description Returns details about the user's current subscription
// @Tags payment
// @Accept json
// @Produce json
// @Success 200 {object} SubscriptionResponse "Subscription details"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /payment/subscription [get]
func GetSubscriptionHandler(c *gin.Context) {
	// Get authenticated user from context
	userID := c.GetUint("userID")

	// Get user from database
	db := database.DB
	user, err := models.FindUserByID(db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	// Check if user has a subscription
	if user.SubscriptionID == nil {
		c.JSON(http.StatusOK, SubscriptionResponse{
			HasSubscription: false,
		})
		return
	}

	// Get subscription details from Stripe
	subscription, err := sub.Get(*user.SubscriptionID, nil)

	if err != nil {
		// If can't retrieve from Stripe, return the local data
		endsAt := user.SubscriptionEndsAt

		c.JSON(http.StatusOK, SubscriptionResponse{
			HasSubscription:  user.IsSubscribed(),
			PlanID:           *user.CurrentPlanID,
			Status:           *user.SubscriptionStatus,
			CurrentPeriodEnd: endsAt,
		})
		return
	}

	// Return subscription details
	periodEnd := time.Unix(subscription.CurrentPeriodEnd, 0)

	c.JSON(http.StatusOK, SubscriptionResponse{
		HasSubscription:   subscription.Status == stripe.SubscriptionStatusActive || subscription.Status == stripe.SubscriptionStatusTrialing,
		SubscriptionID:    subscription.ID,
		PlanID:            *user.CurrentPlanID,
		Status:            string(subscription.Status),
		CancelAtPeriodEnd: subscription.CancelAtPeriodEnd,
		CurrentPeriodEnd:  &periodEnd,
	})
}

// StripeWebhookHandler processes incoming webhook events from Stripe
// @Summary Process Stripe webhook events
// @Description Handles Stripe webhook events for subscription updates, payments, etc.
// @Tags webhook
// @Accept json
// @Produce json
// @Success 200 {object} WebhookResponse "Webhook processed"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Router /stripe/webhook [post]
func StripeWebhookHandler(c *gin.Context) {
	// Read request body
	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Error reading request"})
		return
	}

	// Get webhook secret from env
	webhookSecret := utils.GetEnvWithDefault("STRIPE_WEBHOOK_SECRET", "whsec_your_webhook_secret")

	// Verify signature
	event, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), webhookSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Webhook signature verification failed: %v", err)})
		return
	}

	db := database.DB

	// Handle the event based on its type
	switch event.Type {
	case "checkout.session.completed":
		var sess stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &sess)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Error parsing webhook payload"})
			return
		}

		// Process the session completion
		userIDStr, ok := sess.Metadata["user_id"]
		if !ok {
			fmt.Println("No user_id in session metadata")
			break
		}

		var userID uint
		fmt.Sscanf(userIDStr, "%d", &userID)

		user, err := models.FindUserByID(db, userID)
		if err != nil {
			fmt.Printf("User not found: %v\n", err)
			break
		}

		// Update the payment method if available
		if sess.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid && sess.Customer != nil {
			customerID := sess.Customer.ID

			// Update customer ID if needed
			if user.StripeCustomerID == nil {
				user.UpdateStripeData(db, customerID, "")
			}

			// If this was a subscription purchase
			if sess.Mode == stripe.CheckoutSessionModeSubscription && sess.Subscription != nil {
				// Get subscription details
				subscription, err := sub.Get(sess.Subscription.ID, nil)

				if err != nil {
					fmt.Printf("Error retrieving subscription: %v\n", err)
					break
				}

				// Get plan ID
				var planID string
				if len(subscription.Items.Data) > 0 && subscription.Items.Data[0].Price != nil {
					planID = subscription.Items.Data[0].Price.ID
				} else {
					planID = sess.Metadata["plan_id"]
				}

				// Store subscription details
				periodEnd := time.Unix(subscription.CurrentPeriodEnd, 0)
				if err := user.UpdateSubscriptionData(db, subscription.ID, planID, string(subscription.Status), &periodEnd); err != nil {
					fmt.Printf("Error updating subscription data: %v\n", err)
				}
			}

			// Get customer's payment methods and set the default if needed
			if user.StripeDefaultPM == nil {
				// Get customer to find default payment method
				cus, err := customer.Get(customerID, nil)
				if err == nil && cus.InvoiceSettings.DefaultPaymentMethod != nil {
					user.UpdateStripeData(db, customerID, cus.InvoiceSettings.DefaultPaymentMethod.ID)
				}
			}
		}

	case "customer.subscription.updated", "customer.subscription.created":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Error parsing webhook payload"})
			return
		}

		// Find customer in our database
		if subscription.Customer == nil {
			fmt.Println("No customer attached to subscription")
			break
		}

		// Find user by Stripe customer ID
		var user models.User
		if err := db.Where("stripe_customer_id = ?", subscription.Customer.ID).First(&user).Error; err != nil {
			fmt.Printf("User with Stripe customer ID not found: %v\n", err)
			break
		}

		// Get plan ID
		var planID string
		if len(subscription.Items.Data) > 0 && subscription.Items.Data[0].Price != nil {
			planID = subscription.Items.Data[0].Price.ID
		}

		// Update subscription details
		periodEnd := time.Unix(subscription.CurrentPeriodEnd, 0)
		if err := user.UpdateSubscriptionData(db, subscription.ID, planID, string(subscription.Status), &periodEnd); err != nil {
			fmt.Printf("Error updating subscription data: %v\n", err)
		}

	case "customer.subscription.deleted":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Error parsing webhook payload"})
			return
		}

		// Find customer in our database
		if subscription.Customer == nil {
			fmt.Println("No customer attached to subscription")
			break
		}

		// Find user by Stripe customer ID
		var user models.User
		if err := db.Where("stripe_customer_id = ?", subscription.Customer.ID).First(&user).Error; err != nil {
			fmt.Printf("User with Stripe customer ID not found: %v\n", err)
			break
		}

		// Clear subscription details
		if err := user.UpdateSubscriptionData(db, "", "", "canceled", nil); err != nil {
			fmt.Printf("Error updating subscription data: %v\n", err)
		}

	case "payment_method.attached":
		var pm stripe.PaymentMethod
		err := json.Unmarshal(event.Data.Raw, &pm)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Error parsing webhook payload"})
			return
		}

		// Find customer in our database
		if pm.Customer == nil {
			fmt.Println("No customer attached to payment method")
			break
		}

		// Find user by Stripe customer ID
		var user models.User
		if err := db.Where("stripe_customer_id = ?", pm.Customer.ID).First(&user).Error; err != nil {
			fmt.Printf("User with Stripe customer ID not found: %v\n", err)
			break
		}

		// If this is the first payment method, set it as default
		if user.StripeDefaultPM == nil {
			user.UpdateStripeData(db, pm.Customer.ID, pm.ID)
		}
	}

	c.JSON(http.StatusOK, WebhookResponse{Received: true})
}
