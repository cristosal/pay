package pay

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cristosal/migra"
	"github.com/cristosal/orm"
)

// DefaultSchema where tables will be stored can be overriden using
const DefaultSchema = "pay"

// EntityRepo contains methods for storing entities within an sql database
type EntityRepo struct {
	Events
	db             *sql.DB
	migrationTable string
	schema         string
}

// NewEntityRepo is a constructor for *Repo
func NewEntityRepo(db *sql.DB) *EntityRepo {
	return &EntityRepo{
		db:                       db,
		schema:                   DefaultSchema,
		subAddedCallbacks:        []func(*Subscription){},
		subUpdatedCallbacks:      []func(*Subscription, *Subscription){},
		subRemovedCallbacks:      []func(*Subscription){},
		customerAddedCallbacks:   []func(*Customer){},
		customerUpdatedCallbacks: []func(*Customer, *Customer){},
		customerRemovedCallbacks: []func(*Customer){},
	}
}

// SetMigrationsTable for setting up migrations during init
func (r *EntityRepo) SetMigrationsTable(table string) {
	r.migrationTable = table
}

// SetSchema used for storing entity tables
func (r *EntityRepo) SetSchema(schema string) {
	r.schema = schema
}

// Init creates the required tables and migrations for entities.
// The call to init is idempotent and can therefore be called many times acheiving the same result.
func (r *EntityRepo) Init(ctx context.Context) error {
	orm.SetSchema(r.schema)
	orm.SetMigrationTable(r.migrationTable)

	if err := orm.CreateMigrationTable(r.db); err != nil {
		return err
	}

	if err := orm.AddMigrations(r.db, migrations); err != nil {
		return err
	}

	return orm.Exec(r.db, fmt.Sprintf("SET search_path = %s;", r.schema))
}

// GetPriceByID returns the price by a given id
func (r *EntityRepo) GetPriceByID(priceID int64) (*Price, error) {
	var p Price
	if err := orm.Get(r.db, &p, "WHERE id = $1", priceID); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPriceByID returns the price by a given id
func (r *EntityRepo) GetPriceByProvider(provider, providerID string) (*Price, error) {
	var p Price
	if err := orm.Get(r.db, &p, "WHERE provider = $1 AND provider_id = $2", provider, providerID); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPricesByPlanID returns the planID
func (r *EntityRepo) GetPricesByPlanID(planID int64) ([]Price, error) {
	var p []Price
	if err := orm.List(r.db, &p, "WHERE plan_id = $1", planID); err != nil {
		return nil, err
	}
	return p, nil
}

// Destroy removes all tables and relationships
func (r *EntityRepo) Destroy(ctx context.Context) error {
	m := migra.New(r.db).
		SetSchema("pay").
		SetMigrationTable(r.migrationTable)

	_, err := m.PopAll(ctx)
	return err
}

// AddPrice to plan
func (r *EntityRepo) AddPrice(p *Price) error {
	if err := orm.Add(r.db, p); err != nil {
		return err
	}
	r.priceAdded(p)
	return nil
}

// UpdatePriceByProvider
func (r *EntityRepo) UpdatePriceByProvider(p *Price) error {
	var prev *Price
	_ = orm.Get(r.db, prev, "WHERE provider = $1 AND provider_id = $2", p.Provider, p.ProviderID)

	err := orm.Update(r.db, p, "WHERE provider = $1 AND provider_id = $2",
		p.Provider, p.ProviderID)

	if err != nil {
		return err
	}

	r.priceUpdated(prev, p)
	return nil
}

// RemovePrice deletes price from repository
func (r *EntityRepo) RemovePrice(p *Price) error {
	err := orm.Exec(r.db, "DELETE FROM pay.price WHERE id = $1", p.ID)
	if err != nil {
		return err
	}

	r.priceRemoved(p)
	return nil
}

// RemovePrice deletes price from repository
func (r *EntityRepo) RemovePriceByProvider(p *Price) error {
	return orm.Exec(r.db, "DELETE FROM pay.price WHERE provider = $1 AND provider_id = $2", p.Provider, p.ProviderID)
}

// RemoveCustomers removes all customers from the database
func (r *EntityRepo) RemoveCustomers() error {
	return orm.Exec(r.db, "DELETE FROM pay.customer")
}

// GetCustomerByID returns the customer by its id field
func (r *EntityRepo) GetCustomerByID(id int64) (*Customer, error) {
	var c Customer
	if err := orm.Get(r.db, &c, "WHERE id = $1", id); err != nil {
		return nil, err
	}
	return &c, nil
}

// GetCustomerByEmail returns the customer with a given email
func (r *EntityRepo) GetCustomerByEmail(email string) (*Customer, error) {
	var c Customer
	if err := orm.Get(r.db, &c, "WHERE email = $1", email); err != nil {
		return nil, err
	}
	return &c, nil
}

// GetCustomerByProvider returns the customer with provider id.
// Provider id refers to the id given to the customer by an external provider such as stripe or paypal.
func (r *EntityRepo) GetCustomerByProvider(provider, providerID string) (*Customer, error) {
	var c Customer
	if err := orm.Get(r.db, &c, "WHERE provider_id = $1 AND provider = $2", providerID, provider); err != nil {
		return nil, err
	}

	return &c, nil
}

// UpdateCustomerByID updates a given customer by id field
func (r *EntityRepo) UpdateCustomerByID(c *Customer) error {
	return orm.UpdateByID(r.db, c)
}

// UpdateCustomerByProvider updates a given customer by id field
func (r *EntityRepo) UpdateCustomerByProvider(c *Customer) error {
	return orm.Update(r.db, c, "WHERE provider = $1 AND provider_id = $2", c.Provider, c.ProviderID)
}

// AddCustomer inserts a customer into the repository
func (r *EntityRepo) AddCustomer(c *Customer) error {
	return orm.Add(r.db, c)
}

// RemoveCustomerByProviderID removes customer by given provider
func (r *EntityRepo) RemoveCustomerByProvider(provider, providerID string) error {
	return orm.Exec(r.db, "DELETE FROM pay.customer WHERE provider = $1 AND provider_id = $2", provider, providerID)
}

// ListPlans returns a list of all active plans
func (r *EntityRepo) ListPlans() ([]Plan, error) {
	var plans []Plan
	if err := orm.List(r.db, &plans, "WHERE active = true ORDER BY price ASC"); err != nil {
		return nil, err
	}

	return plans, nil
}

// AddPlan adds a plan to the repository
func (r *EntityRepo) AddPlan(p *Plan) error {
	return orm.Add(r.db, p)
}

// RemovePlanByProviderID deletes a plan by provider id from the repository
func (r *EntityRepo) RemovePlanByProvider(provider, providerID string) error {
	return orm.Exec(r.db, "DELETE FROM pay.plan WHERE provider = $1 AND provider_id = $2", provider, providerID)
}

// UpdatePlanByID updates the plan matching the id field
func (r *EntityRepo) UpdatePlanByID(p *Plan) error {
	return orm.UpdateByID(r.db, p)
}

// UpdatePlanByProvider updates the plan matching the provider and provider id
func (r *EntityRepo) UpdatePlanByProvider(p *Plan) error {
	return orm.Update(r.db, p, "WHERE provider = $1 AND provider_id = $2", p.Provider, p.ProviderID)
}

// GetPlanByID returns the plan matching the internal id
func (r *EntityRepo) GetPlanByID(id int64) (*Plan, error) {
	var p Plan
	if err := orm.Get(r.db, &p, "WHERE id = $1", id); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPlanByProvider returns the plan which matches provider and provider id
func (r *EntityRepo) GetPlanByProvider(provider, providerID string) (*Plan, error) {
	var p Plan

	if err := orm.Get(r.db, &p, "WHERE provider = $1 AND provider_id = $2", provider, providerID); err != nil {
		return nil, err
	}

	return &p, nil
}

// GetPlanByName returns the plan with given name
func (r *EntityRepo) GetPlanByName(name string) (*Plan, error) {
	var p Plan
	if err := orm.Get(r.db, &p, "WHERE name = $1", name); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPlanByCustomerEmail returns the plan associated with the email of a given customer
func (r *EntityRepo) GetPlanByCustomerEmail(email string) (*Plan, error) {
	sql := `SELECT p.*
	FROM
		plan p
	INNER JOIN
		subscription s
	ON
		s.plan_id = p.id
	INNER JOIN
		customer c
	ON
		s.customer_id = c.id AND c.email = $1`

	var p Plan
	if err := orm.QueryRow(r.db, &p, sql, email); err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *EntityRepo) RemovePriceByID(id int) error {
	return orm.Remove(r.db, &Price{}, "WHERE id = $1")
}

func (r *EntityRepo) AddSubscription(s *Subscription) error {
	if err := orm.Add(r.db, s); err != nil {
		return err
	}
	return nil
}

func (r *EntityRepo) UpdateSubscriptionByID(s *Subscription) error {
	return orm.UpdateByID(r.db, s)
}

func (r *EntityRepo) UpdateSubscriptionByProvider(s *Subscription) error {
	prev, _ := r.GetSubscriptionByProvider(s.Provider, s.ProviderID)
	err := orm.Update(r.db, s, "WHERE provider = $1 AND provider_id = $2", s.Provider, s.ProviderID)
	if err != nil {
		return err
	}
	r.subUpdated(prev, s)
	return nil
}

func (r *EntityRepo) RemoveSubscriptionByProvider(s *Subscription) error {
	err := orm.Remove(r.db, s, "WHERE provider = $1 AND provider_id = $2", s.Provider, s.ProviderID)
	if err != nil {
		return err
	}
	r.subRemoved(s)
	return nil
}

func (r *EntityRepo) GetSubscriptionByCustomerID(customerID int64) ([]Subscription, error) {
	var s []Subscription
	if err := orm.List(r.db, &s, "WHERE customer_id = $1", customerID); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *EntityRepo) GetSubscriptionByPlanID(planID int64) (*Subscription, error) {
	var s Subscription
	if err := orm.Get(r.db, &s, "WHERE plan_id = $1", planID); err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *EntityRepo) GetSubscriptionByProvider(provider, providerID string) (*Subscription, error) {
	var s Subscription
	if err := orm.Get(r.db, &s, "WHERE provider = $1 AND provider_id = $2", provider, providerID); err != nil {
		return nil, err
	}

	return &s, nil
}
