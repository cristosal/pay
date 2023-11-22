package pay

import (
	"context"
	"database/sql"

	"github.com/cristosal/dbx"
	"github.com/cristosal/migra"
)

// Repo contains methods for storing entities within an sql database
type Repo struct{ db *sql.DB }

// NewEntityRepo is a constructor for *Repo
func NewEntityRepo(db *sql.DB) *Repo { return &Repo{db} }

// Init creates the required tables and migrations for entities
func (r *Repo) Init(ctx context.Context) error {
	m := migra.New(r.db)

	m.SetMigrationsTable("pay_migrations")

	if err := m.Init(ctx); err != nil {
		return err
	}

	return m.PushMany(ctx, []migra.Migration{
		{
			Name:        "customer_table",
			Description: "Creates customer table",
			Up: `CREATE TABLE IF NOT EXISTS customer (
				id SERIAL PRIMARY KEY,
				user_id INT,
				provider_id VARCHAR(255) NOT NULL,
				provider VARCHAR(32) NOT NULL,
				name VARCHAR(255) NOT NULL,
				email VARCHAR(255) NOT NULL,
				UNIQUE (provider_id, provider)
			)`,
			Down: "DROP TABLE customer",
		},
		{
			Name:        "plan_table",
			Description: "Create plan table",
			Up: `CREATE TABLE IF NOT EXISTS plan (
				id SERIAL PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				provider VARCHAR(255) NOT NULL,
				provider_id VARCHAR(255) NOT NULL,
				active BOOL NOT NULL,
				trial_days INT NOT NULL DEFAULT 0,
				price INT NOT NULL
			)`,
			Down: "DROP TABLE plan",
		},
		{
			Name:        "subscription_table",
			Description: "Creates a subscription table",
			Up: `CREATE TABLE IF NOT EXISTS subscription (
				id SERIAL PRIMARY KEY,
				customer_id INT NOT NULL,
				plan_id INT NOT NULL,
				provider VARCHAR(255) NOT NULL,
				provider_id VARCHAR(255) NOT NULL,
				active BOOL NOT NULL DEFAULT FALSE,
				FOREIGN KEY (customer_id) REFERENCES customer(id),
				FOREIGN KEY (plan_id) REFERENCES plan(id)
			)`,
			Down: "DROP TABLE subscription",
		},
		// add more migrations here
	})
}

// GetCustomerByID returns the customer by its id field
func (r *Repo) GetCustomerByID(id int64) (*Customer, error) {
	var c Customer
	if err := dbx.One(r.db, &c, "WHERE id = $1", id); err != nil {
		return nil, err
	}
	return &c, nil
}

// GetCustomerByEmail returns the customer with a given email
func (r *Repo) GetCustomerByEmail(email string) (*Customer, error) {
	var c Customer
	if err := dbx.One(r.db, &c, "WHERE email = $1", email); err != nil {
		return nil, err
	}
	return &c, nil
}

// GetCustomerByProvider returns the customer with provider id.
// Provider id refers to the id given to the customer by an external provider such as stripe or paypal.
func (r *Repo) GetCustomerByProvider(provider, providerID string) (*Customer, error) {
	var c Customer
	if err := dbx.One(r.db, &c, "WHERE provider_id = $1 AND provider = $2", providerID, provider); err != nil {
		return nil, err
	}

	return &c, nil
}

// UpdateCustomerByID updates a given customer by id field
func (r *Repo) UpdateCustomerByID(c *Customer) error {
	return dbx.Update(r.db, c)
}

// AddCustomer inserts a customer into the repository
func (r *Repo) AddCustomer(c *Customer) error {
	return dbx.Insert(r.db, c)
}

// RemoveCustomerByProviderID removes customer by given provider
func (r *Repo) RemoveCustomerByProviderID(providerID string) error {
	return dbx.Exec(r.db, "DELETE FROM customer WHERE provider_id = $1", providerID)
}

func (r *Repo) ListPlans() ([]Plan, error) {
	var plans []Plan
	if err := dbx.Many(r.db, &plans, "WHERE active = true ORDER BY price ASC"); err != nil {
		return nil, err
	}

	return plans, nil
}

func (r *Repo) AddPlan(p *Plan) error {
	return dbx.Insert(r.db, p)
}

func (r *Repo) RemovePlanByProviderID(providerID string) error {
	return dbx.Exec(r.db, "DELETE FROM plan WHERE provider_id = $1", providerID)
}

func (r *Repo) UpdatePlanByID(p *Plan) error {
	return dbx.Update(r.db, p)
}

func (r *Repo) GetPlanByID(id int64) (*Plan, error) {
	var p Plan
	if err := dbx.One(r.db, &p, "where id = $1", id); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) GetPlanByProviderID(providerID string) (*Plan, error) {
	var p Plan
	if err := dbx.One(r.db, &p, "where provider_id = $1", providerID); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) GetPlanByName(name string) (*Plan, error) {
	var p Plan
	if err := dbx.One(r.db, &p, "where name = $1", name); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) GetPlanByEmail(email string) (*Plan, error) {
	sql := `select p.*
	from
		plan p
	inner join
		subscription s
	on
		s.plan_id = p.id
	inner join
		customer c
	on
		s.customer_id = c.id and c.email = $1
	`

	var p Plan
	if err := dbx.QueryRow(r.db, &p, sql, email); err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *Repo) AddSubscription(s *Subscription) error {
	return dbx.Insert(r.db, s)
}

func (r *Repo) UpdateSubscriptionByID(s *Subscription) error {
	return dbx.Update(r.db, s)
}

func (r *Repo) RemoveSubscriptionByProviderID(providerID string) error {
	return dbx.Exec(r.db, "delete from subscription where provider_id = $1", providerID)
}

func (r *Repo) GetSubscriptionByCustomerID(customerID int64) ([]Subscription, error) {
	var s []Subscription
	if err := dbx.Many(r.db, &s, "where customer_id = $1", customerID); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repo) GetSubscriptionByPlanID(planID int64) (*Subscription, error) {
	var s Subscription
	if err := dbx.One(r.db, &s, "where plan_id = $1", planID); err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *Repo) GetSubscriptionByProviderID(providerID string) (*Subscription, error) {
	var s Subscription
	if err := dbx.One(r.db, &s, "where provider_id = $1", providerID); err != nil {
		return nil, err
	}

	return &s, nil
}
