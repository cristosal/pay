package pay

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/subscription"
)

// SyncSubscriptions pulls in all subscriptions from stripe
func (s *StripeService) SyncSubscriptions() error {
	it := subscription.List(nil)
	for it.Next() {
		sub := it.Subscription()

		if err := s.saveSubscription(sub); err != nil {
			log.Printf("error while syncing subscriptions: %v", err)
		}
	}

	return it.Err()
}

// handleSubscriptionEvent updates or adds an internal representation of stripe event
func (s *StripeService) handleSubscriptionEvent(e *stripe.Event) error {
	var stripeSub stripe.Subscription
	if err := json.Unmarshal(e.Data.Raw, &stripeSub); err != nil {
		return err
	}

	return s.saveSubscription(&stripeSub)
}

// saveSubscription is an upsert
func (s *StripeService) saveSubscription(stripeSub *stripe.Subscription) error {
	sub, err := s.getSubscription(stripeSub)
	if err != nil {
		return err
	}

	if sub.ID == 0 {
		// subscription was added we can fire an event
		if err := s.Repository().AddSubscription(sub); err != nil {
			return err
		}

		// fire callback
		s.subAddCallback(sub)
		return nil
	}

	// check if the subscription
	if err := s.Repository().UpdateSubscriptionByID(sub); err != nil {
		return err
	}

	s.subUpdatedCallback(sub)
	return nil
}

// returns a subscription with local customer and plan data from database
// performs 3 database queries are performed
func (s *StripeService) getSubscription(stripeSub *stripe.Subscription) (*Subscription, error) {
	if len(stripeSub.Items.Data) == 0 {
		return nil, errors.New("subscription item is not present in data")
	}

	sub := Subscription{
		ProviderID: stripeSub.ID,
		Provider:   ProviderStripe,
		Active:     stripeSub.Status == "active",
	}

	// lookup to find local id in database +1
	found, _ := s.Repository().GetSubscriptionByProviderID(stripeSub.ID)
	if found != nil {
		sub.ID = found.ID
	}

	// lookup the customer
	cust, err := s.Repository().GetCustomerByProvider(ProviderStripe, stripeSub.Customer.ID)
	if err != nil {
		return nil, err
	}

	sub.CustomerID = cust.ID
	sub.Customer = cust

	productID := stripeSub.Items.Data[0].Plan.Product.ID
	// look up the plan + 3
	plan, err := s.Repository().GetPlanByProviderID(productID)
	if err != nil {
		return nil, err
	}

	sub.PlanID = plan.ID
	sub.Plan = plan

	return &sub, nil
}
