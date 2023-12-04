package pay

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/cristosal/dbx"
	"github.com/cristosal/migra"
)

// DefaultSchema where tables will be stored can be overriden using
const DefaultSchema = "pay"

//go:embed migrations
var migrations embed.FS // directory containing migration files

// EntityRepo contains methods for storing entities within an sql database
type EntityRepo struct {
	db             *sql.DB
	migrationTable string
	schema         string
}

// NewEntityRepo is a constructor for *Repo
func NewEntityRepo(db *sql.DB) *EntityRepo {
	return &EntityRepo{
		db:             db,
		migrationTable: migra.DefaultMigrationTable,
		schema:         DefaultSchema,
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
	m := migra.New(r.db).
		SetSchema(r.schema).
		SetMigrationTable(r.migrationTable)

	if err := m.CreateMigrationTable(ctx); err != nil {
		return err
	}

	if err := m.PushDirFS(ctx, migrations, "migrations"); err != nil {
		return err
	}

	_, err := r.db.Exec(fmt.Sprintf("SET search_path = %s;", r.schema))
	return err
}

// GetPriceByID returns the price by a given id
func (r *EntityRepo) GetPriceByID(priceID int64) (*Price, error) {
	var p Price
	if err := dbx.One(r.db, &p, "WHERE id = $1", priceID); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPriceByID returns the price by a given id
func (r *EntityRepo) GetPriceByProvider(provider, providerID string) (*Price, error) {
	var p Price
	if err := dbx.One(r.db, &p, "WHERE provider = $1 AND provider_id = $2", provider, providerID); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPricesByPlanID returns the planID
func (r *EntityRepo) GetPricesByPlanID(planID int64) ([]Price, error) {
	var p []Price
	if err := dbx.Many(r.db, &p, "WHERE plan_id = $1", planID); err != nil {
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
	return dbx.Insert(r.db, p)
}

// UpdatePriceByProvider
func (r *EntityRepo) UpdatePriceByProvider(p *Price) error {
	return dbx.Update(r.db, p, "WHERE provider = $1 AND provider_id = $2",
		p.Provider, p.ProviderID)
}

// RemovePrice deletes price from repository
func (r *EntityRepo) RemovePrice(p *Price) error {
	return dbx.Exec(r.db, "DELETE FROM price WHERE id = $1", p.ID)
}

// RemovePrice deletes price from repository
func (r *EntityRepo) RemovePriceByProvider(p *Price) error {
	return dbx.Exec(r.db, "DELETE FROM price WHERE provider = $1 AND provider_id = $2", p.Provider, p.ProviderID)
}

// ClearCustomers removes all customers from the database
func (r *EntityRepo) ClearCustomers() error {
	return dbx.Exec(r.db, "DELETE FROM customer")
}

// GetCustomerByID returns the customer by its id field
func (r *EntityRepo) GetCustomerByID(id int64) (*Customer, error) {
	var c Customer
	if err := dbx.One(r.db, &c, "WHERE id = $1", id); err != nil {
		return nil, err
	}
	return &c, nil
}

// GetCustomerByEmail returns the customer with a given email
func (r *EntityRepo) GetCustomerByEmail(email string) (*Customer, error) {
	var c Customer
	if err := dbx.One(r.db, &c, "WHERE email = $1", email); err != nil {
		return nil, err
	}
	return &c, nil
}

// GetCustomerByProvider returns the customer with provider id.
// Provider id refers to the id given to the customer by an external provider such as stripe or paypal.
func (r *EntityRepo) GetCustomerByProvider(provider, providerID string) (*Customer, error) {
	var c Customer
	if err := dbx.One(r.db, &c, "WHERE provider_id = $1 AND provider = $2", providerID, provider); err != nil {
		return nil, err
	}

	return &c, nil
}

// UpdateCustomerByID updates a given customer by id field
func (r *EntityRepo) UpdateCustomerByID(c *Customer) error {
	return dbx.UpdateByID(r.db, c)
}

// UpdateCustomerByProvider updates a given customer by id field
func (r *EntityRepo) UpdateCustomerByProvider(c *Customer) error {
	return dbx.Update(r.db, c, "WHERE provider = $1 AND provider_id = $2", c.Provider, c.ProviderID)
}

// AddCustomer inserts a customer into the repository
func (r *EntityRepo) AddCustomer(c *Customer) error {
	return dbx.Insert(r.db, c)
}

// RemoveCustomerByProviderID removes customer by given provider
func (r *EntityRepo) DeleteCustomerByProvider(provider, providerID string) error {
	return dbx.Exec(r.db, "DELETE FROM customer WHERE provider = $1 AND provider_id = $2", provider, providerID)
}

// ListPlans returns a list of all active plans
func (r *EntityRepo) ListPlans() ([]Plan, error) {
	var plans []Plan
	if err := dbx.Many(r.db, &plans, "WHERE active = true ORDER BY price ASC"); err != nil {
		return nil, err
	}

	return plans, nil
}

// AddPlan adds a plan to the repository
func (r *EntityRepo) AddPlan(p *Plan) error {
	return dbx.Insert(r.db, p)
}

// RemovePlanByProviderID deletes a plan by provider id from the repository
func (r *EntityRepo) RemovePlanByProvider(provider, providerID string) error {
	return dbx.Exec(r.db, "DELETE FROM plan WHERE provider = $1 AND provider_id = $2", provider, providerID)
}

// UpdatePlanByID updates the plan matching the id field
func (r *EntityRepo) UpdatePlanByID(p *Plan) error {
	return dbx.UpdateByID(r.db, p)
}

// UpdatePlanByProvider updates the plan matching the provider and provider id
func (r *EntityRepo) UpdatePlanByProvider(p *Plan) error {
	return dbx.Update(r.db, p, "WHERE provider = $1 AND provider_id = $2", p.Provider, p.ProviderID)
}

// GetPlanByID returns the plan matching the internal id
func (r *EntityRepo) GetPlanByID(id int64) (*Plan, error) {
	var p Plan
	if err := dbx.One(r.db, &p, "WHERE id = $1", id); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPlanByProvider returns the plan which matches provider and provider id
func (r *EntityRepo) GetPlanByProvider(provider, providerID string) (*Plan, error) {
	var p Plan

	if err := dbx.One(r.db, &p, "WHERE provider = $1 AND provider_id = $2", provider, providerID); err != nil {
		return nil, err
	}

	return &p, nil
}

// GetPlanByName returns the plan with given name
func (r *EntityRepo) GetPlanByName(name string) (*Plan, error) {
	var p Plan
	if err := dbx.One(r.db, &p, "WHERE name = $1", name); err != nil {
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
	if err := dbx.QueryRow(r.db, &p, sql, email); err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *EntityRepo) RemovePriceByID(id int) error {
	return dbx.Exec(r.db, "DELETE FROM price WHERE id = $1")

}

func (r *EntityRepo) AddSubscription(s *Subscription) error {
	return dbx.Insert(r.db, s)
}

func (r *EntityRepo) UpdateSubscriptionByID(s *Subscription) error {
	return dbx.UpdateByID(r.db, s)
}

func (r *EntityRepo) RemoveSubscriptionByProviderID(providerID string) error {
	return dbx.Exec(r.db, "delete from subscription where provider_id = $1", providerID)
}

func (r *EntityRepo) GetSubscriptionByCustomerID(customerID int64) ([]Subscription, error) {
	var s []Subscription
	if err := dbx.Many(r.db, &s, "where customer_id = $1", customerID); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *EntityRepo) GetSubscriptionByPlanID(planID int64) (*Subscription, error) {
	var s Subscription
	if err := dbx.One(r.db, &s, "where plan_id = $1", planID); err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *EntityRepo) GetSubscriptionByProviderID(providerID string) (*Subscription, error) {
	var s Subscription
	if err := dbx.One(r.db, &s, "where provider_id = $1", providerID); err != nil {
		return nil, err
	}

	return &s, nil
}
