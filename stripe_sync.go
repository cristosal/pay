package pay

import (
	"context"
	"fmt"
	"log"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
	"github.com/stripe/stripe-go/v74/subscription"
)

// Sync repository data with stripe
func (s *StripeService) Sync() error {
	if err := s.syncCustomers(); err != nil {
		return fmt.Errorf("error syncing customers: %w", err)
	}

	if err := s.syncPlans(); err != nil {
		return fmt.Errorf("error syncing plans: %w", err)
	}

	if err := s.syncPrices(context.Background()); err != nil {
		return fmt.Errorf("error syncing prices: %w", err)
	}

	if err := s.syncSubscriptions(); err != nil {
		return fmt.Errorf("error syncing subscriptions: %w", err)
	}

	return nil
}

func (s *StripeService) syncPrices(ctx context.Context) error {
	it := price.List(nil)

	for it.Next() {
		p := it.Price()
		plan, _ := s.Entities().GetPlanByProvider(ProviderStripe, p.Product.ID)

		if plan == nil {
			continue
		}

		if err := s.Entities().AddPrice(&Price{
			PlanID:     plan.ID,
			Amount:     p.UnitAmount,
			Provider:   ProviderStripe,
			ProviderID: p.ID,
			Currency:   string(p.Currency),
			Schedule:   s.convertPricingSchedule(p),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *StripeService) syncCustomers() error {
	it := customer.List(nil)
	for it.Next() {
		cust := it.Customer()
		c := &Customer{
			Provider:   ProviderStripe,
			ProviderID: cust.ID,
			Name:       cust.Name,
			Email:      cust.Email,
		}

		found, _ := s.Entities().GetCustomerByProvider(ProviderStripe, cust.ID)
		if found == nil {
			if err := s.Entities().AddCustomer(c); err != nil {
				log.Printf("error while adding stripe customer with id %s: %v", c.ProviderID, err)
			}
			continue
		}

		c.ID = found.ID
		// avoid db queries and check if we really have to update the customer
		// we can only change email or name from stripe portal all other fields are internal
		if c.Email != found.Email || c.Name != found.Name {
			if err := s.Entities().UpdateCustomerByID(c); err != nil {
				log.Printf("error while updating stripe customer with id %s: %v", c.ProviderID, err)
			}
		}
	}

	if it.Err() != nil {
		return it.Err()
	}

	return nil
}

func (s *StripeService) syncPlans() error {
	plans, err := s.fetchPlans()
	if err != nil {
		return err
	}

	for _, p := range plans {
		// check if we have it
		found, _ := s.Entities().GetPlanByProvider(p.Provider, p.ProviderID)

		if found == nil {
			if err := s.Entities().AddPlan(&p); err != nil {
				log.Printf("error adding plan during sync: %v", err)
			}

			// get next
			continue
		}

		p.ID = found.ID
		if err := s.Entities().UpdatePlanByID(&p); err != nil {
			log.Printf("error updating plan during sync: %v", err)
		}
	}

	return nil
}

// syncSubscriptions pulls in all subscriptions from stripe
func (s *StripeService) syncSubscriptions() error {
	it := subscription.List(nil)
	for it.Next() {
		_ = it.Subscription()
		// TODO: upsert subscriptions
	}

	return it.Err()
}

// fetchPlans from stripe
func (s *StripeService) fetchPlans() ([]Plan, error) {
	params := &stripe.ProductListParams{}
	params.AddExpand("data.default_price")

	it := product.List(params)

	var plans []Plan
	for it.Next() {
		p := it.Product()
		if p.DefaultPrice == nil {
			continue
		}

		plans = append(plans, Plan{
			ProviderID: p.ID,
			Provider:   ProviderStripe,
			Name:       p.Name,
			Active:     p.Active,
		})
	}

	if it.Err() != nil {
		return nil, it.Err()
	}

	return plans, nil
}
