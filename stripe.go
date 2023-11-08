package pay

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cristosal/pgxx"
	"github.com/jackc/pgx/v5"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"github.com/stripe/stripe-go/v74/product"
)

const StripeProvider = "stripe"

var ErrCheckoutFailed = errors.New("checkout failed")

type (
	// StripeConfig configures StripeService with necessary credentials and callbacks
	StripeConfig struct {
		CustomerRepo
		PlanRepo
		SubscriptionRepo
		StripeEventRepo
		Key           string
		WebhookSecret string
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
func (s *StripeService) Init() error {
	if err := s.Customers().Init(); err != nil {
		return fmt.Errorf("error initializing customers: %w", err)
	}

	if err := s.Plans().Init(); err != nil {
		return fmt.Errorf("error initializing plans: %w", err)
	}

	if err := s.Subscriptions().Init(); err != nil {
		return fmt.Errorf("error initializing subscriptions: %w", err)
	}

	if err := s.EventRepo().Init(); err != nil {
		return fmt.Errorf("error initializing stripe event store: %w", err)
	}

	if err := s.Sync(); err != nil {
		return fmt.Errorf("error syncing data: %w", err)
	}

	return nil
}

// Sync data with stripe
func (s *StripeService) Sync() error {
	if err := s.syncCustomers(); err != nil {
		return fmt.Errorf("error syncing customers: %w", err)
	}

	if err := s.syncPlans(); err != nil {
		return fmt.Errorf("error syncing plans: %w", err)
	}

	if err := s.syncSubscriptions(); err != nil {
		return fmt.Errorf("error syncing subscriptions: %w", err)
	}

	return nil
}

func (s *StripeService) EventRepo() StripeEventRepo {
	return s.cfg.StripeEventRepo
}

// Customers returns the underlying customer repo
func (s *StripeService) Customers() CustomerRepo {
	return s.cfg.CustomerRepo
}

// Plans returns the underlying plan repo
func (s *StripeService) Plans() PlanRepo {
	return s.cfg.PlanRepo
}

// Plans returns the underlying plan repo
func (s *StripeService) Subscriptions() SubscriptionRepo {
	return s.cfg.SubscriptionRepo
}

func (s *StripeService) OnSubscriptionAdded(fn func(*Subscription)) {
	s.subAddCallback = fn
}

func (s *StripeService) OnSubscriptionUpdated(fn func(*Subscription)) {
	s.subUpdatedCallback = fn
}

func (s *StripeService) ListPlans() ([]Plan, error) {
	return s.Plans().List()
}

// CurrentUserPlan returns the current active plan for a given user
func (s *StripeService) PlanByUser(uid pgxx.ID) (*Plan, error) {
	cust, err := s.Customers().ByUserID(uid)
	if err != nil {
		return nil, ErrNoPlan
	}

	subs, err := s.Subscriptions().ByCustomerID(cust.ID)
	if err != nil {
		return nil, ErrNoPlan
	}

	var sub *Subscription
	for i := range subs {
		if subs[i].Active {
			sub = &subs[i]
			break
		}
	}

	if sub == nil || !sub.Active {
		return nil, ErrNoPlan
	}

	return s.Plans().ByID(sub.PlanID)
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
	cust, err := s.AddCustomer(req.UserID, req.Name, req.Email)

	if errors.Is(err, ErrNotFound) {
		err = nil
	}

	if err != nil {
		return
	}

	pl, err := s.Plans().ByName(req.Plan)
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
