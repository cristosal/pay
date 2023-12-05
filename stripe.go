package pay

import (
	"context"
	"errors"
	"time"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
)

const ProviderStripe = "stripe"

var ErrCheckoutFailed = errors.New("checkout failed")

type (
	// StripeConfig configures StripeService with necessary credentials and callbacks
	StripeConfig struct {
		EntityRepo    *Repo
		Key           string
		WebhookSecret string
	}

	// StripeProvider interfaces with stripe for customer, plan and subscription data
	StripeProvider struct {
		config *StripeConfig
	}
)

// NewStripeProvider creates a provider service for interacting with stripe
func NewStripeProvider(config *StripeConfig) *StripeProvider {
	if config == nil {
		config = new(StripeConfig)
	}
	stripe.Key = config.Key
	return &StripeProvider{config}
}

// Init creates necessary tables by executing migrations
func (s *StripeProvider) Init(ctx context.Context) error {
	return s.Repo().Init(ctx)
}

// Repo returns the entity repository
func (s *StripeProvider) Repo() *Repo {
	return s.config.EntityRepo
}

// Verify that the checkout was completed
func (s *StripeProvider) VerifyCheckout(sessionID string) error {
	sess, err := session.Get(sessionID, nil)
	if err != nil {
		return err
	}

	if sess.PaymentStatus == stripe.CheckoutSessionPaymentStatusUnpaid {
		return ErrCheckoutFailed
	}

	return nil
}

// CheckoutRequest
type CheckoutRequest struct {
	CustomerID  int64
	PriceID     int64
	RedirectURL string `json:"redirect_url"`
}

// Checkout returns the url that a user has to visit in order to complete payment
// it registers the customer if it was unavailable
func (s *StripeProvider) Checkout(request *CheckoutRequest) (url string, err error) {
	customer, err := s.Repo().GetCustomerByID(request.CustomerID)
	if err != nil {
		return
	}

	price, err := s.Repo().GetPriceByID(request.PriceID)
	if err != nil {
		return
	}

	var trialEnd *int64 = nil

	if price.TrialDays > 0 {
		// we add one day of grace so that stripe displays the correct amount.
		// since trial end is calculated from current time, being one second off will result in days -1 being displayed in stripe checkout
		trialEnd = stripe.Int64(price.TrialEnd().Add(time.Hour * 24).Unix())
	}

	params := &stripe.CheckoutSessionParams{
		Customer:                stripe.String(customer.ProviderID),
		SuccessURL:              stripe.String(request.RedirectURL),
		Mode:                    stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		PaymentMethodCollection: stripe.String("if_required"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(price.ProviderID),
				Quantity: stripe.Int64(1),
			},
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			TrialEnd: trialEnd,
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return
	}

	url = sess.URL
	return
}

// this should go here
func (StripeProvider) convertPricingSchedule(p *stripe.Price) PricingSchedule {
	switch p.Type {
	case stripe.PriceTypeOneTime:
		return PricingOnce
	case stripe.PriceTypeRecurring:
		switch p.Recurring.Interval {
		case stripe.PriceRecurringIntervalMonth:
			return PricingMonthly
		case stripe.PriceRecurringIntervalYear:
			return PricingAnnual
		default:
			return ""
		}
	default:
		return ""
	}
}
