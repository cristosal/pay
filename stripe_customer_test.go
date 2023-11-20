package pay_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/cristosal/pay"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewStripeService(t *testing.T) *pay.StripeService {
	db, err := sql.Open("pgx", os.Getenv("CONNECTION_STRING"))
	if err != nil {
		t.Fatal(err)
	}

	s := pay.NewStripeProvider(&pay.StripeConfig{
		Key:        os.Getenv("STRIPE_API_KEY"),
		EntityRepo: pay.NewRepo(db),
	})

	return s
}

func TestListingCustomerSubscriptions(t *testing.T) {
	ss := NewStripeService(t)
	cust, err := ss.Repo().CustomerByEmail("admin@cibera.com.mx")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", cust)
}
