package pay

import (
	"context"
	"database/sql"

	"github.com/cristosal/dbx"
	"github.com/cristosal/migra"
)

// Repo contains methods for storing entities within an sql database
type Repo struct{ *sql.DB }

// NewRepo is a constructor for *Repo
func NewRepo(db *sql.DB) *Repo { return &Repo{db} }

// Init creates the required tables and migrations for entities
func (r *Repo) Init(ctx context.Context) error {
	m := migra.New(r.DB)

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

func (r *Repo) CustomerByID(id int64) (*Customer, error) {
	var c Customer
	if err := dbx.One(r, &c, "WHERE id = $1", id); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repo) CustomerByProviderID(id string) (*Customer, error) {
	var c Customer
	if err := dbx.One(r, &c, "WHERE provider_id = $1", id); err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *Repo) UpdateCustomerByID(c *Customer) error {
	return dbx.Update(r, c)
}

func (r *Repo) AddCustomer(c *Customer) error {
	return dbx.Insert(r, c)
}

func (r *Repo) RemoveCustomersByProviderID(providerID string) error {
	return dbx.Exec(r, "DELETE FROM customer WHERE provider_id = $1", providerID)
}

func (r *Repo) ListPlans() ([]Plan, error) {
	var plans []Plan
	if err := dbx.Many(r, &plans, "WHERE active = true ORDER BY price ASC"); err != nil {
		return nil, err
	}

	return plans, nil
}

func (r *Repo) AddPlan(p *Plan) error {
	return dbx.Insert(r, p)
}

func (r *Repo) RemovePlanByProviderID(providerID string) error {
	return dbx.Exec(r, "DELETE FROM plan WHERE provider_id = $1", providerID)
}

func (r *Repo) UpdatePlanByID(p *Plan) error {
	return dbx.Update(r, p)
}

func (r *Repo) PlanByID(id int64) (*Plan, error) {
	var p Plan
	if err := dbx.One(r, &p, "where id = $1", id); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) PlanByProviderID(providerID string) (*Plan, error) {
	var p Plan
	if err := dbx.One(r, &p, "where provider_id = $1", providerID); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) PlanByName(name string) (*Plan, error) {
	var p Plan
	if err := dbx.One(r, &p, "where name = $1", name); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) PlanByEmail(email string) (*Plan, error) {
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
	if err := dbx.QueryRow(r.DB, &p, sql, email); err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *Repo) AddSubscription(s *Subscription) error {
	return dbx.Insert(r, s)
}

func (r *Repo) UpdateSubscriptionByID(s *Subscription) error {
	return dbx.Update(r, s)
}

func (r *Repo) RemoveSubscriptionByProviderID(providerID string) error {
	return dbx.Exec(r, "delete from subscription where provider_id = $1", providerID)
}

func (r *Repo) SubscriptionByCustomerID(customerID int64) ([]Subscription, error) {
	var s []Subscription
	if err := dbx.Many(r, &s, "where customer_id = $1", customerID); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repo) SubscriptionByPlanID(planID int64) (*Subscription, error) {
	var s Subscription
	if err := dbx.One(r, &s, "where plan_id = $1", planID); err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *Repo) SubscriptionByProviderID(providerID string) (*Subscription, error) {
	var s Subscription
	if err := dbx.One(r, &s, "where provider_id = $1", providerID); err != nil {
		return nil, err
	}

	return &s, nil
}
