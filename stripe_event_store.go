package pay

import (
	"context"
	"database/sql"

	"github.com/cristosal/migra"
	"github.com/stripe/stripe-go/v74"
)

type StripeEventRepo struct {
	db *sql.DB
}

func NewStripeEventRepo(db *sql.DB) *StripeEventRepo {
	return &StripeEventRepo{db}
}

// Init table and migrations
func (s *StripeEventRepo) Init(ctx context.Context) error {
	m := migra.New(s.db).SetMigrationTable("pay_migrations")

	if err := m.CreateMigrationTable(ctx); err != nil {
		return err
	}

	return m.PushMany(ctx, []migra.Migration{
		{
			Name:        "stripe_events",
			Description: "Create stripe_events table",
			Up: `CREATE TABLE IF NOT EXISTS stripe_events (
				event_id VARCHAR(255) NOT NULL PRIMARY KEY,
				event_type VARCHAR(255) NOT NULL,
				payload JSONB NOT NULL,
				processed BOOL NOT NULL DEFAULT FALSE
			)`,
			Down: "DROP TABLE strip_events",
		},
	})
}

// Has true if we have already stored the event
func (r *StripeEventRepo) Has(ev *stripe.Event) bool {
	var id string
	row := r.db.QueryRow("SELECT event_id FROM stripe_events WHERE event_id = $1")
	_ = row.Scan(&id)
	return id != ""
}

// Add inserts the event in the store
func (r *StripeEventRepo) Add(ev *stripe.Event) error {
	_, err := r.db.Exec("insert into stripe_events (event_id, event_type, payload) values ($1, $2, $3)", ev.ID, ev.Type, ev.Data)
	return err
}
