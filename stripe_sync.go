package pay

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/cristosal/orm"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
	"github.com/stripe/stripe-go/v74/subscription"
)

// Sync repository data with stripe
func (s *StripeProvider) Sync() error {
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

func (s *StripeProvider) syncPrices(ctx context.Context) error {
	it := price.List(nil)

	for it.Next() {
		p := it.Price()

		pr, err := s.convertPrice(p)
		if err != nil {
			log.Printf("error converting price %s: %v", p.ID, err)
			continue
		}

		_, err = s.Entities().GetPriceByProvider(ProviderStripe, p.ID)

		if errors.Is(err, orm.ErrNotFound) {
			if err := s.Entities().UpdatePriceByProvider(pr); err != nil {
				log.Printf("error updating price %s: %v", pr.ProviderID, err)
			}
			continue
		}

		if err != nil {
			log.Printf("error getting price %s: %v", p.ID, err)
			continue
		}

		if err := s.Entities().AddPrice(pr); err != nil {
			log.Printf("error adding price %s: %v", pr.ProviderID, err)
			continue
		}
	}

	return nil
}

func (s *StripeProvider) syncCustomers() error {
	it := customer.List(nil)
	for it.Next() {
		cust := it.Customer()
		c := s.convertCustomer(cust)

		found, _ := s.Entities().GetCustomerByProvider(ProviderStripe, cust.ID)
		if found == nil {
			if err := s.Entities().AddCustomer(c); err != nil {
				log.Printf("error while adding stripe customer with id %s: %v", c.ProviderID, err)
			}
			continue
		}

		c.ID = found.ID
		if err := s.Entities().UpdateCustomerByProvider(c); err != nil {
			log.Printf("error while updating stripe customer with id %s: %v", c.ProviderID, err)
		}
	}

	if it.Err() != nil {
		return it.Err()
	}

	return nil
}

func (s *StripeProvider) syncPlans() error {
	it := product.List(nil)

	for it.Next() {
		p := it.Product()
		pl := s.convertProduct(p)

		// we need to see if we already have it
		_, err := s.Entities().GetPlanByProvider(ProviderStripe, p.ID)

		if errors.Is(err, orm.ErrNotFound) {
			if err := s.Entities().AddPlan(pl); err != nil {
				log.Printf("error while adding plan %s: %v", p.ID, err)
			}
			continue
		}

		if err != nil {
			log.Printf("error while getting plan %s: %v", p.ID, err)
			continue
		}

		if err := s.Entities().UpdatePlanByProvider(pl); err != nil {
			log.Printf("error updating plan %s: %v", pl.ProviderID, err)
		}
	}

	return it.Err()
}

// syncSubscriptions pulls in all subscriptions from stripe
func (s *StripeProvider) syncSubscriptions() error {
	it := subscription.List(nil)
	for it.Next() {
		sub := it.Subscription()
		subscr, err := s.convertSubscription(sub)
		if err != nil {
			log.Printf("error converting subscription %s: %v", sub.ID, err)
		}

		_, err = s.Entities().GetSubscriptionByProvider(ProviderStripe, sub.ID)
		if errors.Is(err, orm.ErrNotFound) {
			// we add it
			if err := s.Entities().AddSubscription(subscr); err != nil {
				log.Printf("error adding subscription %s: %v", subscr.ProviderID, err)
			}
			continue
		}

		if err != nil {
			log.Printf("error getting subscription %s: %v", sub.ID, err)
			continue
		}

		if err := s.Entities().UpdateSubscriptionByProvider(subscr); err != nil {
			log.Printf("error updating subscription %s: %v", subscr.ProviderID, err)
			continue
		}
	}

	return it.Err()
}
