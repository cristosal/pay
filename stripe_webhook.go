package pay

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
)

// Webhook returns the http handler that is responsible for handling any event received from stripe
func (s *StripeService) Webhook() http.HandlerFunc {
	const MaxBodyBytes = int64(65536)
	whevents := make(chan stripe.Event)

	go func() {
		for event := range whevents {
			log.Printf("stripe webhook: recieved event: %s", event.Type)
			var err error

			switch event.Type {
			case "product.created":
				err = s.handleProductCreated(event.Data)
			case "product.updated":
				err = s.handleProductUpdated(event.Data)
			case "product.deleted":
				err = s.handleProductDeleted(event.Data)
			case "price.created":
				err = s.handlePriceCreated(event.Data)
			case "price.updated":
				err = s.handlePriceUpdated(event.Data)
			case "price.deleted":
				err = s.handlePriceDeleted(event.Data)
			case "customer.created":
				err = s.handleCustomerCreated(event.Data)
			case "customer.updated":
				err = s.handleCustomerUpdated(event.Data)
			case "customer.deleted":
				err = s.handleCustomerDeleted(event.Data)
			case "customer.subscription.created":
				err = s.handleSubscriptionCreated(event.Data)
			}

			if err != nil {
				log.Printf("error handling stripe event %s: %v", event.Type, err)
			}
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		payload, err := io.ReadAll(r.Body)

		if err != nil {
			log.Printf("Error reading request body: %v\n", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		// Pass the request body and Stripe-Signature header to ConstructEvent, along
		// with the webhook signing key.
		sig := r.Header.Get("Stripe-Signature")
		event, err := webhook.ConstructEvent(payload, sig, s.config.WebhookSecret)

		if err != nil {
			log.Printf("Error verifying webhook signature: %v\n", err)
			w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
			return
		}

		if s.WebhookEvents().Has(&event) {
			log.Printf("Already processed event with id %s", event.ID)
			w.WriteHeader(http.StatusOK)
			return
		}

		if err := s.WebhookEvents().Add(&event); err != nil {
			log.Printf("error while saving stripe event: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		whevents <- event
		w.WriteHeader(http.StatusOK)
	}
}

// convert subscription types
func (s StripeService) convertSubscription(sub *stripe.Subscription) (*Subscription, error) {
	// ensure that the first item is a subscription
	if sub.Items == nil ||
		len(sub.Items.Data) == 0 ||
		sub.Items.Data[0].Price == nil {
		return nil, errors.New("unable to get price id from subscription")
	}

	priceID := sub.Items.Data[0].Price.ID
	pr, err := s.Entities().GetPriceByProvider(ProviderStripe, priceID)
	if err != nil {
		return nil, fmt.Errorf("could not get price %s: %w", priceID, err)
	}

	cust, err := s.Entities().GetCustomerByProvider(ProviderStripe, sub.Customer.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get customer with provider_id = %s for subscription %s: %w",
			sub.Customer.ID, sub.ID, err)
	}

	subscr := Subscription{
		Provider:   ProviderStripe,
		ProviderID: sub.ID,
		CustomerID: cust.ID,
		PriceID:    pr.ID,
		Active:     sub.Status == stripe.SubscriptionStatusActive || sub.Status == stripe.SubscriptionStatusTrialing,
	}

	return &subscr, nil

	// subscription is done to a price so we have to note that
}

func (s *StripeService) handleSubscriptionCreated(data *stripe.EventData) error {
	var sub stripe.Subscription
	if err := sub.UnmarshalJSON(data.Raw); err != nil {
		return err
	}

	subscr, err := s.convertSubscription(&sub)
	if err != nil {
		return err
	}

	return s.Entities().AddSubscription(subscr)
}

func (s *StripeService) handleCustomerDeleted(data *stripe.EventData) error {
	var c stripe.Customer
	if err := json.Unmarshal(data.Raw, &c); err != nil {
		return err
	}
	return s.Entities().DeleteCustomerByProvider(ProviderStripe, c.ID)
}

func (StripeService) custconv(c *stripe.Customer) *Customer {
	return &Customer{
		ProviderID: c.ID,
		Provider:   ProviderStripe,
		Name:       c.Name,
		Email:      c.Email,
	}
}

func (s *StripeService) handleCustomerCreated(data *stripe.EventData) error {
	var c stripe.Customer
	if err := json.Unmarshal(data.Raw, &c); err != nil {
		return err
	}
	return s.Entities().AddCustomer(s.custconv(&c))
}

func (s *StripeService) handleCustomerUpdated(data *stripe.EventData) error {
	var c stripe.Customer
	if err := json.Unmarshal(data.Raw, &c); err != nil {
		return err
	}
	return s.Entities().UpdateCustomerByProvider(&Customer{
		ProviderID: c.ID,
		Provider:   ProviderStripe,
		Name:       c.Name,
		Email:      c.Email,
	})
}

func (s *StripeService) handlePriceDeleted(data *stripe.EventData) error {
	var p stripe.Price
	if err := json.Unmarshal(data.Raw, &p); err != nil {
		return err
	}
	return s.Entities().RemovePriceByProvider(&Price{
		Provider:   ProviderStripe,
		ProviderID: p.ID,
	})
}

func (s *StripeService) handlePriceUpdated(data *stripe.EventData) error {
	var p stripe.Price
	if err := json.Unmarshal(data.Raw, &p); err != nil {
		return err
	}

	pl, err := s.Entities().GetPlanByProvider(ProviderStripe, p.Product.ID)
	if err != nil {
		return err
	}
	return s.Entities().UpdatePriceByProvider(&Price{
		Provider:   ProviderStripe,
		ProviderID: p.ID,
		Amount:     p.UnitAmount,
		Currency:   string(p.Currency),
		Schedule:   s.getPricing(&p),
		TrialDays:  int(p.Recurring.TrialPeriodDays), // TODO: check if this is actually sent through in the webhook
		PlanID:     pl.ID,
	})
}

func (s *StripeService) handlePriceCreated(data *stripe.EventData) error {
	var p stripe.Price
	if err := json.Unmarshal(data.Raw, &p); err != nil {
		return err
	}
	pl, err := s.Entities().GetPlanByProvider(ProviderStripe, p.Product.ID)
	if err != nil {
		return err
	}
	return s.Entities().AddPrice(&Price{
		Provider:   ProviderStripe,
		ProviderID: p.ID,
		Amount:     p.UnitAmount,
		Currency:   string(p.Currency),
		Schedule:   s.getPricing(&p),
		TrialDays:  int(p.Recurring.TrialPeriodDays), // TODO: check if this is actually sent through in the webhook
		PlanID:     pl.ID,
	})
}

func (s *StripeService) handleProductCreated(data *stripe.EventData) error {
	var p stripe.Product
	if err := json.Unmarshal(data.Raw, &p); err != nil {
		return err
	}
	return s.Entities().AddPlan(&Plan{
		Name:       p.Name,
		Provider:   ProviderStripe,
		ProviderID: p.ID,
		Active:     p.Active,
	})
}

func (s *StripeService) handleProductDeleted(data *stripe.EventData) error {
	var p stripe.Product
	if err := json.Unmarshal(data.Raw, &p); err != nil {
		return err
	}
	return s.Entities().RemovePlanByProvider(ProviderStripe, p.ID)
}

func (s *StripeService) handleProductUpdated(data *stripe.EventData) error {
	var p stripe.Product
	if err := json.Unmarshal(data.Raw, &p); err != nil {
		return err
	}
	return s.Entities().UpdatePlanByProvider(&Plan{
		Name:       p.Name,
		Provider:   ProviderStripe,
		ProviderID: p.ID,
		Active:     p.Active,
	})
}
