package pay

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
)

// addCustomer creates a customer in Stripe and inserts it into the repo.
func (s *StripeService) addCustomer(name, email string) (*Customer, error) {
	c, err := s.Entities().GetCustomerByEmail(email)

	if errors.Is(err, pgx.ErrNoRows) {
		return s.createCustomer(name, email)
	}

	// otherwise we got another err
	if err != nil {
		return nil, err
	}

	if err := s.Entities().UpdateCustomerByID(c); err != nil {
		return nil, err
	}

	return c, nil
}

// SyncCustomers pulls all customers from stripe and upserts them in the repository
func (s *StripeService) SyncCustomers() error {
	it := customer.List(nil)
	for it.Next() {
		cust := it.Customer()

		c := &Customer{
			Name:       cust.Name,
			Email:      cust.Email,
			Provider:   ProviderStripe,
			ProviderID: cust.ID,
		}

		//
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

// creates customer in stripe as part of checkout session logic
func (s *StripeService) createCustomer(name, email string) (*Customer, error) {
	cust, err := customer.New(&stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	})

	if err != nil {
		return nil, err
	}

	c := &Customer{
		Provider:   ProviderStripe,
		ProviderID: cust.ID,
		Name:       name,
		Email:      email,
	}

	if err := s.Entities().AddCustomer(c); err != nil {
		return nil, err
	}

	return c, nil
}

// handler for when a customer is deleted from stripe
func (s *StripeService) handleCustomerDeleted(data *stripe.EventData) error {
	var c stripe.Customer
	if err := json.Unmarshal(data.Raw, &c); err != nil {
		return err
	}

	return s.Entities().DeleteCustomerByProvider(ProviderStripe, c.ID)
}
