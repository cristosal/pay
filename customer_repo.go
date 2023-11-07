package pay

import "github.com/cristosal/pgxx"

// Customer represents a paying customer attached to a third party service like stripe or paypal
type Customer struct {
	ID         pgxx.ID  // internal (to this service)
	UserID     *pgxx.ID // our internal identifier
	ProviderID string   // external providers id
	Provider   string   // the provider for this customer
	Name       string   // customers name
	Email      string   // customers email
}

func (c *Customer) HasUser() bool {
	return c.UserID != nil
}

func (c *Customer) TableName() string {
	return "customer"
}

type (
	CustomerRepo interface {
		Init() error
		Add(*Customer) error
		RemoveByUserID(pgxx.ID) error
		RemoveByProviderID(string) error
		Update(*Customer) error
		ByProviderID(string) (*Customer, error)
		ByID(pgxx.ID) (*Customer, error)
		ByUserID(pgxx.ID) (*Customer, error)
		ByEmail(string) (*Customer, error)
	}

	CustomerPgxRepo struct{ pgxx.DB }
)

// NewCustomerPgxRepo returns an implementation of CustomerRepo using postgres as the underlying data store
func NewCustomerPgxRepo(db pgxx.DB) *CustomerPgxRepo {
	return &CustomerPgxRepo{db}
}

// Init creates customer table
func (r *CustomerPgxRepo) Init() error {
	return pgxx.Exec(r, `create table if not exists customer (
		id serial primary key,
		user_id int,
		provider_id varchar(255) not null,
		provider varchar(32) not null,
		name varchar(255) not null,
		email varchar(255) not null,
		unique (provider_id, provider)
	)`)
}

// ByProviderID returns a customer by the external providers id
func (r *CustomerPgxRepo) ByProviderID(id string) (*Customer, error) {
	var c Customer
	if err := pgxx.One(r, &c, "where provider_id = $1", id); err != nil {
		return nil, err
	}

	return &c, nil
}

// ByEmail returns a customer by given email
func (r *CustomerPgxRepo) ByEmail(email string) (*Customer, error) {
	var c Customer
	if err := pgxx.One(r, &c, "where email = $1", email); err != nil {
		return nil, err
	}

	return &c, nil
}

// ByEmail returns a customer by user id
func (r *CustomerPgxRepo) ByID(id pgxx.ID) (*Customer, error) {
	var c Customer
	if err := pgxx.One(r, &c, "where id = $1", id); err != nil {
		return nil, err
	}
	return &c, nil
}

// ByEmail returns a customer by user id
func (r *CustomerPgxRepo) ByUserID(id pgxx.ID) (*Customer, error) {
	var c Customer
	if err := pgxx.One(r, &c, "where user_id = $1", id); err != nil {
		return nil, err
	}
	return &c, nil
}

// Update a customer
func (r *CustomerPgxRepo) Update(c *Customer) error {
	return pgxx.Update(r, c)
}

// Add customer to repository
func (r *CustomerPgxRepo) Add(c *Customer) error {
	return pgxx.Insert(r, c)
}

// Remove customer from repository
func (r *CustomerPgxRepo) RemoveByUserID(userID pgxx.ID) error {
	return pgxx.Exec(r, "delete from customer where user_id = $1", userID)
}

// Remove customer from repository
func (r *CustomerPgxRepo) RemoveByProviderID(providerID string) error {
	return pgxx.Exec(r, "delete from customer where provider_id = $1", providerID)
}
