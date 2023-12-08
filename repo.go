package pay

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cristosal/orm"
)

// DefaultSchema where tables will be stored can be overriden using
const DefaultSchema = "pay"

// Repo contains methods for storing entities within an sql database
type Repo struct {
	events
	db             *sql.DB
	migrationTable string
	schema         string
}

// NewEntityRepo is a constructor for *Repo
func NewEntityRepo(db *sql.DB) *Repo {
	return &Repo{
		db:     db,
		schema: DefaultSchema,
	}
}

// SetMigrationsTable for setting up migrations during init
func (r *Repo) SetMigrationsTable(table string) {
	r.migrationTable = table
}

// SetSchema used for storing entity tables
func (r *Repo) SetSchema(schema string) {
	r.schema = schema
}

// Init creates the required tables and migrations for entities.
// The call to init is idempotent and can therefore be called many times acheiving the same result.
func (r *Repo) Init(ctx context.Context) error {
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
func (r *Repo) GetPriceByID(priceID int64) (*Price, error) {
	var p Price
	if err := orm.Get(r.db, &p, "WHERE id = $1", priceID); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPriceByID returns the price by a given id
func (r *Repo) GetPriceByProvider(provider, providerID string) (*Price, error) {
	var p Price
	if err := orm.Get(r.db, &p, "WHERE provider = $1 AND provider_id = $2", provider, providerID); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPricesByPlanID returns the planID
func (r *Repo) GetPricesByPlanID(planID int64) ([]Price, error) {
	var p []Price
	if err := orm.List(r.db, &p, "WHERE plan_id = $1", planID); err != nil {
		return nil, err
	}
	return p, nil
}

// Destroy removes all tables and relationships
func (r *Repo) Destroy(ctx context.Context) error {
	orm.SetSchema(r.schema)
	orm.SetMigrationTable(r.migrationTable)
	_, err := orm.RemoveAllMigrations(r.db)
	return err
}

// ListAllCustomers returns a list of prices
func (r *Repo) ListAllCustomers() ([]Customer, error) {
	var customers []Customer
	if err := orm.ListAll(r.db, &customers); err != nil {
		return nil, err
	}

	return customers, nil
}

// ListAllWebhookEvents returns a list of all webhook events
func (r *Repo) ListAllWebhookEvents() ([]WebhookEvent, error) {
	var webhookEvents []WebhookEvent
	if err := orm.ListAll(r.db, &webhookEvents); err != nil {
		return nil, err
	}

	return webhookEvents, nil
}

// ListAllPrices returns a list of prices
func (r *Repo) ListAllPrices() ([]Price, error) {
	var prices []Price
	if err := orm.ListAll(r.db, &prices); err != nil {
		return nil, err
	}

	return prices, nil
}

// ListPrices returns a list of prices
func (r *Repo) ListPricesByPlanID(planID int64) ([]Price, error) {
	var prices []Price
	if err := orm.List(r.db, &prices, "WHERE plan_id = $1", planID); err != nil {
		return nil, err
	}

	return prices, nil
}

func (r *Repo) ListAllSubscriptions() ([]Subscription, error) {
	var subs []Subscription
	if err := orm.ListAll(r.db, &subs); err != nil {
		return nil, err
	}
	return subs, nil
}

// addPrice to plan
func (r *Repo) addPrice(p *Price) error {
	if err := orm.Add(r.db, p); err != nil {
		return err
	}
	r.priceAdded(p)
	return nil
}

// UpdatePriceByProvider
func (r *Repo) updatePriceByProvider(p *Price) error {
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
func (r *Repo) removePriceByProvider(p *Price) error {
	err := orm.Remove(r.db, "WHERE provider = $1 AND provider_id = $2", p.Provider, p.ProviderID)
	if err != nil {
		return err
	}

	r.priceRemoved(p)
	return nil
}

// GetCustomerByID returns the customer by its id field
func (r *Repo) GetCustomerByID(id int64) (*Customer, error) {
	var c Customer
	if err := orm.Get(r.db, &c, "WHERE id = $1", id); err != nil {
		return nil, err
	}
	return &c, nil
}

// GetCustomerByEmail returns the customer with a given email
func (r *Repo) GetCustomerByEmail(email string) (*Customer, error) {
	var c Customer
	if err := orm.Get(r.db, &c, "WHERE email = $1", email); err != nil {
		return nil, err
	}
	return &c, nil
}

// GetCustomerByProvider returns the customer with provider id.
// Provider id refers to the id given to the customer by an external provider such as stripe or paypal.
func (r *Repo) GetCustomerByProvider(provider, providerID string) (*Customer, error) {
	var c Customer
	if err := orm.Get(r.db, &c, "WHERE provider_id = $1 AND provider = $2", providerID, provider); err != nil {
		return nil, err
	}

	return &c, nil
}

// UpdateCustomerByProvider updates a given customer by id field
func (r *Repo) updateCustomerByProvider(c *Customer) error {
	var prev Customer
	if err := orm.Get(r.db, &prev, "WHERE provider = $1 AND provider_id = $2", c.Provider, c.ProviderID); err != nil {
		return err
	}

	if err := orm.Update(r.db, c, "WHERE provider = $1 AND provider_id = $2", c.Provider, c.ProviderID); err != nil {
		return err
	}

	r.customerUpdated(&prev, c)
	return nil
}

// AddCustomer inserts a customer into the repository
func (r *Repo) addCustomer(c *Customer) error {
	if err := orm.Add(r.db, c); err != nil {
		return err
	}
	r.customerAdded(c)
	return nil
}

// RemoveCustomerByProviderID removes customer by given provider
func (r *Repo) removeCustomerByProvider(provider, providerID string) error {
	var c Customer
	if err := orm.Get(r.db, &c, "WHERE provider = $1 AND provider_id = $2", provider, providerID); err != nil {
		return err
	}

	if err := orm.Remove(r.db, &c, "WHERE provider = $1 AND provider_id = $2", provider, providerID); err != nil {
		return err
	}

	r.customerRemoved(&c)
	return nil
}

// ListActivePlans returns a list of all active plans in alphabetic order
func (r *Repo) ListActivePlans() ([]Plan, error) {
	var plans []Plan
	if err := orm.List(r.db, &plans, "WHERE active = TRUE ORDER BY name ASC"); err != nil {
		return nil, err
	}

	return plans, nil
}

// AddPlan adds a plan to the repository
func (r *Repo) addPlan(p *Plan) error {
	if err := orm.Add(r.db, p); err != nil {
		return err
	}

	r.planAdded(p)
	return nil
}

// RemovePlanByProviderID deletes a plan by provider id from the repository
func (r *Repo) removePlanByProvider(provider, providerID string) error {
	var p Plan
	if err := orm.Get(r.db, &p, "WHERE provider = $1 AND provider_id = $2", provider, providerID); err != nil {
		return err
	}
	if err := orm.Remove(r.db, &p, "WHERE provider = $1 AND provider_id = $2", provider, providerID); err != nil {
		return err
	}
	r.planRemoved(&p)
	return nil
}

// UpdatePlanByProvider updates the plan matching the provider and provider id
func (r *Repo) updatePlanByProvider(p *Plan) error {
	var prev Plan
	if err := orm.Get(r.db, &prev, "WHERE provider = $1 AND provider_id = $2", p.Provider, p.ProviderID); err != nil {
		return err
	}

	if err := orm.Update(r.db, p, "WHERE provider = $1 AND provider_id = $2", p.Provider, p.ProviderID); err != nil {
		return err
	}

	r.planUpdated(&prev, p)
	return nil
}

// GetPlanByID returns the plan matching the internal id
func (r *Repo) GetPlanByID(id int64) (*Plan, error) {
	var p Plan
	if err := orm.Get(r.db, &p, "WHERE id = $1", id); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPlanByProvider returns the plan which matches provider and provider id
func (r *Repo) GetPlanByProvider(provider, providerID string) (*Plan, error) {
	var p Plan

	if err := orm.Get(r.db, &p, "WHERE provider = $1 AND provider_id = $2", provider, providerID); err != nil {
		return nil, err
	}

	return &p, nil
}

// GetPlanByName returns the plan with given name
func (r *Repo) GetPlanByName(name string) (*Plan, error) {
	var p Plan
	if err := orm.Get(r.db, &p, "WHERE name = $1", name); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPlanByCustomerEmail returns the plan associated with the email of a given customer
func (r *Repo) GetPlanByCustomerEmail(email string) (*Plan, error) {
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

func (r *Repo) addSubscription(s *Subscription) error {
	if err := orm.Add(r.db, s); err != nil {
		return err
	}
	r.subAdded(s)
	return nil
}

func (r *Repo) updateSubscriptionByProvider(s *Subscription) error {
	var prev Subscription
	if err := orm.Get(r.db, &prev, "WHERE provider = $1 AND provider_id = $2", s.Provider, s.ProviderID); err != nil {
		return err
	}

	if err := orm.Update(r.db, s, "WHERE provider = $1 AND provider_id = $2", s.Provider, s.ProviderID); err != nil {
		return err
	}

	r.subUpdated(&prev, s)
	return nil
}

func (r *Repo) removeSubscriptionByProvider(s *Subscription) error {
	if err := orm.Remove(r.db, s, "WHERE provider = $1 AND provider_id = $2", s.Provider, s.ProviderID); err != nil {
		return err
	}

	r.subRemoved(s)
	return nil
}

func (r *Repo) GetSubscriptionByCustomerID(customerID int64) ([]Subscription, error) {
	var s []Subscription
	if err := orm.List(r.db, &s, "WHERE customer_id = $1", customerID); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repo) GetSubscriptionByPlanID(planID int64) (*Subscription, error) {
	var s Subscription
	if err := orm.Get(r.db, &s, "WHERE plan_id = $1", planID); err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *Repo) GetSubscriptionByProvider(provider, providerID string) (*Subscription, error) {
	var s Subscription
	if err := orm.Get(r.db, &s, "WHERE provider = $1 AND provider_id = $2", provider, providerID); err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *Repo) hasWebhookEvent(provider, providerID string) bool {
	var e WebhookEvent
	if err := orm.Get(r.db, &e, "WHERE provider = $1 AND provider_id = $2", provider, providerID); err != nil {
		return false
	}

	return e.Provider == provider && e.ProviderID == providerID
}

func (r *Repo) addWebhookEvent(e *WebhookEvent) error {
	return orm.Add(r.db, e)
}

func (r *Repo) GetPlanBySubscriptionID(subID int64) (*Plan, error) {
	var p Plan

	sql := fmt.Sprintf("SELECT %s FROM %s p INNER JOIN %s pr ON pr.plan_id = p.id INNER JOIN %s s ON s.price_id = pr.id WHERE s.id = $1",
		orm.Columns(&p).PrefixedList("p"),
		orm.TableName(&p),
		orm.TableName(&Price{}),
		orm.TableName(&Subscription{}),
	)

	if err := orm.QueryRow(r.db, &p, sql, subID); err != nil {
		return nil, err
	}

	return &p, nil
}
