package pay

import (
	"context"

	"github.com/cristosal/pgxx"
	"github.com/stripe/stripe-go/v74"
)

type StripeEventRepo interface {
	Init() error
	Add(*stripe.Event) error
	Has(*stripe.Event) bool
}

type StripeEventPgxRepo struct{ pgxx.DB }

func NewStripeEventPgxRepo(db pgxx.DB) *StripeEventPgxRepo {
	return &StripeEventPgxRepo{db}
}

// Init initializes events
func (s *StripeEventPgxRepo) Init() error {
	sql := `create table if not exists stripe_events (
		event_id varchar(255) not null primary key,
		event_type varchar(255) not null,
		payload jsonb not null,
		processed bool not null default false
	)`

	return pgxx.Exec(s, sql)
}

// Has true if we have already stored the event
func (r *StripeEventPgxRepo) Has(ev *stripe.Event) bool {
	var id string
	row := r.QueryRow(context.Background(), "select event_id from stripe_events where event_id = $1")
	_ = row.Scan(&id)
	return id != ""
}

// Add inserts the event in the store
func (r *StripeEventPgxRepo) Add(ev *stripe.Event) error {
	_, err := r.Exec(context.Background(), "insert into stripe_events (event_id, event_type, payload) values ($1, $2, $3)", ev.ID, ev.Type, ev.Data)
	return err
}
