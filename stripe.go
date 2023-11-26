package pay

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

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
		EntityRepo      *Repo
		StripeEventRepo *StripeEventRepo
		Key             string
		WebhookSecret   string
	}

	// StripeService interfaces with stripe for customer, plan and subscription data
	StripeService struct {
		cfg                *StripeConfig
		subAddCallback     func(*Subscription)
		subUpdatedCallback func(*Subscription)
	}
)

// NewStripeProvider creates a service for interacting with stripe
func NewStripeProvider(cfg *StripeConfig) *StripeService {
	if cfg == nil {
		cfg = new(StripeConfig)
	}

	stripe.Key = cfg.Key

	return &StripeService{
		cfg:                cfg,
		subAddCallback:     func(*Subscription) {},
		subUpdatedCallback: func(*Subscription) {},
	}
}

// Init creates tables and syncs data
func (s *StripeService) Init(ctx context.Context) error {
	if err := s.Repository().Init(ctx); err != nil {
		return fmt.Errorf("error initializing customers: %w", err)
	}

	if err := s.Events().Init(ctx); err != nil {
		return fmt.Errorf("error initializing stripe event store: %w", err)
	}

	if err := s.Sync(); err != nil {
		return fmt.Errorf("error syncing data: %w", err)
	}

	return nil
}

// Sync data with stripe
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

func (s *StripeService) Events() *StripeEventRepo {
	return s.cfg.StripeEventRepo
}

// Customers returns the underlying customer repo
func (s *StripeService) Repository() *Repo {
	return s.cfg.EntityRepo
}

func (s *StripeService) OnSubscriptionAdded(fn func(*Subscription)) {
	s.subAddCallback = fn
}

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

func (s *StripeService) Checkout(req *CheckoutRequest) (url string, err error) {
	cust, err := s.AddCustomer(req.Name, req.Email)

	if errors.Is(err, ErrNotFound) {
		err = nil
	}

	if err != nil {
		return
	}

	pl, err := s.Repository().GetPlanByName(req.Plan)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNoPlan
	}

	if err != nil {
		return
	}

	prod, err := product.Get(pl.ProviderID, nil)
	if err != nil {
		return "", ErrNoPlan
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
