package pay_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/cristosal/pay"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPriceSync(t *testing.T) {
	db, err := sql.Open("pgx", os.Getenv("TEST_CONNECTION_STRING"))

	if err != nil {
		t.Fatal(err)
	}

	s := pay.NewStripeProvider(&pay.StripeConfig{
		EntityRepo:      pay.NewEntityRepo(db),
		StripeWebhookEventRepo: pay.NewStripeEventRepo(db),
		Key:             os.Getenv("TEST_STRIPE_KEY"),
		WebhookSecret:   os.Getenv("TEST_STRIPE_WEBHOOK_SECRET"),
	})

	ctx := context.Background()

	if err := s.Init(ctx); err != nil {
		t.Fatal(err)
	}

	if err := s.SyncPrices(ctx); err != nil {
		t.Fatal(err)
	}

}
