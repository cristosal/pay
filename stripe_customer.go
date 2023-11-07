package pay

import (
	"encoding/json"
	"errors"

	"github.com/cristosal/pgxx"
	"github.com/jackc/pgx/v5"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
)

// AddCustomer creates a customer in Stripe and inserts it into the repo.
// If a customer with given email already exists, the user id is assigned to the customer.
func (s *StripeService) AddCustomer(uid pgxx.ID, name, email string) (*Customer, error) {
	c, err := s.Customers().ByEmail(email)

	if errors.Is(err, pgx.ErrNoRows) {
		return s.createCustomer(uid, name, email)
	}

	// otherwise we got another err
	if err != nil {
		return nil, err
	}

	c.UserID = &uid
	if err := s.Customers().Update(c); err != nil {
		return nil, err
	}

	return c, nil
}

// syncCustomers pulls all customers from stripe and adds or updates them in the repo
func (s *StripeService) syncCustomers() error {
	it := customer.List(nil)
	for it.Next() {
		cust := it.Customer()
		c := &Customer{
			ProviderID: cust.ID,
			Provider:   StripeProvider,
			Name:       cust.Name,
			Email:      cust.Email,
		}

		found, _ := s.Customers().ByProviderID(cust.ID)

		if found == nil {
			s.Customers().Add(c)
			continue
		}

		c.ID = found.ID
		// avoid db queries and check if we really have to update the customer
		// we can only change email or name from stripe portal all other fields are internal
		if c.Email != found.Email || c.Name != found.Name {
			s.Customers().Update(c)
		}
	}

	if it.Err() != nil {
		return it.Err()
	}

	return nil
}

// creates customer in stripe as part of checkout session logic
func (s *StripeService) createCustomer(uid pgxx.ID, name, email string) (*Customer, error) {
	cust, err := customer.New(&stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	})

	if err != nil {
		return nil, err
	}

	c := &Customer{
		UserID:     &uid,
		Provider:   StripeProvider,
		ProviderID: cust.ID,
		Name:       name,
		Email:      email,
	}

	if err := s.Customers().Add(c); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *StripeService) handleCustomerDeleted(data *stripe.EventData) error {
	var c stripe.Customer
	if err := json.Unmarshal(data.Raw, &c); err != nil {
		return err
	}

	return s.Customers().RemoveByProviderID(c.ID)
}
