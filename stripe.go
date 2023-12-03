package pay

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cristosal/dbx"
	"github.com/jackc/pgx/v5"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"github.com/stripe/stripe-go/v74/product"
)

const ProviderStripe = "stripe"

var ErrCheckoutFailed = errors.New("checkout failed")

type (
	// StripeConfig configures StripeService with necessary credentials and callbacks
	StripeConfig struct {
		EntityRepo             *EntityRepo
		StripeWebhookEventRepo *StripeEventRepo
		Key                    string
		WebhookSecret          string
	}

	// StripeService interfaces with stripe for customer, plan and subscription data
	StripeService struct {
		config             *StripeConfig
		subAddCallback     func(*Subscription)
		subUpdatedCallback func(*Subscription)
	}
)

// NewStripeProvider creates a provider service for interacting with stripe
func NewStripeProvider(config *StripeConfig) *StripeService {

	if config == nil {
		config = new(StripeConfig)
	}

	stripe.Key = config.Key

	return &StripeService{
		config:             config,
		subAddCallback:     func(*Subscription) {},
		subUpdatedCallback: func(*Subscription) {},
	}
}

// Init creates necessary tables by executing migrations
func (s *StripeService) Init(ctx context.Context) error {
	if err := s.Entities().Init(ctx); err != nil {
		return fmt.Errorf("error initializing customers: %w", err)
	}

	if err := s.WebhookEvents().Init(ctx); err != nil {
		return fmt.Errorf("error initializing stripe event store: %w", err)
	}

	return nil
}

// Sync repository data with stripe
func (s *StripeService) Sync() error {
	if err := s.SyncCustomers(); err != nil {
		return fmt.Errorf("error syncing customers: %w", err)
	}

	if err := s.SyncPlans(); err != nil {
		return fmt.Errorf("error syncing plans: %w", err)
	}

	if err := s.SyncPrices(context.Background()); err != nil {
		return fmt.Errorf("error syncing prices: %w", err)
	}

	if err := s.SyncSubscriptions(); err != nil {
		return fmt.Errorf("error syncing subscriptions: %w", err)
	}

	return nil
}

// WebhookEvents returns the stripe event repository
func (s *StripeService) WebhookEvents() *StripeEventRepo {
	return s.config.StripeWebhookEventRepo
}

// Entities returns the entity repository
func (s *StripeService) Entities() *EntityRepo {
	return s.config.EntityRepo
}

// OnSubscriptionAdded registers a callback that is invoked whenever a subscription is added to the repository
func (s *StripeService) OnSubscriptionAdded(fn func(*Subscription)) {
	s.subAddCallback = fn
}

// OnSubscriptionUpdated registers a callback that is invoked whenever a subscription is updated in the repository
func (s *StripeService) OnSubscriptionUpdated(fn func(*Subscription)) {
	s.subUpdatedCallback = fn
}

// Verify that the checkout was completed
func (s *StripeService) VerifyCheckout(sessionID string) error {
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
	UserID      dbx.ID `json:"user_id"`
	Plan        string `json:"plan"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	RedirectURL string `json:"redirect_url"`
}

// Checkout returns the url that a user has to visit in order to complete payment
// it registers the customer if it was unavailable
func (s *StripeService) Checkout(req *CheckoutRequest) (url string, err error) {
	cust, err := s.addCustomer(req.Name, req.Email)
	if errors.Is(err, ErrNotFound) {
		err = nil
	}

	if err != nil {
		return
	}

	pl, err := s.Entities().GetPlanByName(req.Plan)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}

	if err != nil {
		return
	}

	prod, err := product.Get(pl.ProviderID, nil)
	if err != nil {
		return "", ErrNotFound
	}

	var trialEnd *int64 = nil

	if pl.TrialDays > 0 {
		// we add one day of grace so that stripe displays the correct amount.
		// since trial end is calculated from current time, being one second off will result in days -1 being displayed in stripe checkout
		trialEnd = stripe.Int64(pl.TrialEnd().Add(time.Hour * 24).Unix())
	}

	params := &stripe.CheckoutSessionParams{
		Customer:                stripe.String(cust.ProviderID),
		SuccessURL:              stripe.String(req.RedirectURL),
		ClientReferenceID:       stripe.String(strconv.FormatInt(int64(req.UserID), 10)),
		Mode:                    stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		PaymentMethodCollection: stripe.String("if_required"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(prod.DefaultPrice.ID),
				Quantity: stripe.Int64(1),
			},
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			TrialEnd: trialEnd,
			Metadata: map[string]string{
				"user_id": req.UserID.String(),
			},
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return
	}

	url = sess.URL
	return
}
