package pay

import (
	"context"

	"github.com/cristosal/pgxx"
	"github.com/jackc/pgx/v5"
)

type Subscription struct {
	ID         pgxx.ID
	CustomerID pgxx.ID
	PlanID     pgxx.ID
	Provider   string
	ProviderID string
	Active     bool

	// Plan attached to this subscription
	Plan *Plan `db:"-"`

	// Customer attached to this subscription
	Customer *Customer `db:"-"`
}

type (
	SubscriptionRepo interface {
		Init() error
		Add(*Subscription) error
		Remove(providerID string) error
		Update(*Subscription) error
		ByCustomerID(pgxx.ID) ([]Subscription, error)
		ByPlanID(pgxx.ID) (*Subscription, error)
		ByProviderID(string) (*Subscription, error)
		PlanByEmail(string) (*Plan, error)
	}

	SubscriptionPgxRepo struct{ pgxx.DB }
)

func NewSubscriptionPgxRepo(db pgxx.DB) *SubscriptionPgxRepo {
	return &SubscriptionPgxRepo{db}
}

func (r *SubscriptionPgxRepo) PlanByEmail(email string) (*Plan, error) {
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

	rows, err := r.Query(context.Background(), sql, email)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	return pgx.CollectOneRow[*Plan](rows, pgx.RowToAddrOfStructByName)
}

func (r *SubscriptionPgxRepo) Init() error {
	return pgxx.Exec(r, `create table if not exists subscription (
		id serial primary key,
		customer_id int not null,
		plan_id int not null,
		provider varchar(255) not null,
		provider_id varchar(255) not null,
		active bool not null default false,
		foreign key (customer_id) references customer(id),
		foreign key (plan_id) references plan(id)
	)`)
}

func (r *SubscriptionPgxRepo) Add(s *Subscription) error {
	return pgxx.Insert(r, s)
}

func (r *SubscriptionPgxRepo) Update(s *Subscription) error {
	return pgxx.Update(r, s)
}

func (r *SubscriptionPgxRepo) Remove(providerID string) error {
	return pgxx.Exec(r, "delete from subscription where provider_id = $1", providerID)
}

func (r *SubscriptionPgxRepo) ByCustomerID(customerID pgxx.ID) ([]Subscription, error) {
	var s []Subscription
	if err := pgxx.Many(r, &s, "where customer_id = $1", customerID); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *SubscriptionPgxRepo) ByPlanID(planID pgxx.ID) (*Subscription, error) {
	var s Subscription
	if err := pgxx.One(r, &s, "where plan_id = $1", planID); err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *SubscriptionPgxRepo) ByProviderID(providerID string) (*Subscription, error) {
	var s Subscription
	if err := pgxx.One(r, &s, "where provider_id = $1", providerID); err != nil {
		return nil, err
	}

	return &s, nil
}
