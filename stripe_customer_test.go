package pay_test

import (
	"context"
	"os"
	"testing"

	"github.com/cristosal/pay"
	"github.com/jackc/pgx/v5"
)

func NewStripeService(t *testing.T) *pay.StripeService {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("CONNECTION_STRING"))
	if err != nil {
		t.Fatal(err)
	}

	s := pay.NewStripeProvider(&pay.StripeConfig{
		Key:              os.Getenv("STRIPE_API_KEY"),
		CustomerRepo:     pay.NewCustomerPgxRepo(conn),
		PlanRepo:         pay.NewPlanPgxRepo(conn),
		SubscriptionRepo: pay.NewSubscriptionPgxRepo(conn),
	})

	return s
}

func TestListingCustomerSubscriptions(t *testing.T) {
	ss := NewStripeService(t)
	cust, err := ss.Customers().ByEmail("admin@cibera.com.mx")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", cust)
}
