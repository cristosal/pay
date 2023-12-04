package pay_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/cristosal/pay"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestInit(t *testing.T) {
	var (
		r   = getPgxEntityRepo(t)
		n   = 3
		ctx = context.Background()
	)

	t.Cleanup(func() {
		_ = r.Destroy(ctx)
	})

	for i := 0; i < n; i++ {
		if err := r.Init(ctx); err != nil {
			t.Fatal(err)
		}
	}

	if err := r.Destroy(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestPlansAndPricing(t *testing.T) {
	var (
		r   = getPgxEntityRepo(t)
		ctx = context.Background()
	)

	t.Cleanup(func() {
		_ = r.Destroy(ctx)
	})

	if err := r.AddPlan(&pay.Plan{
		Name:       "Test Plan",
		Provider:   "test",
		ProviderID: "1",
		Active:     true,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestCustomerRepo(t *testing.T) {
	var (
		r   = getPgxEntityRepo(t)
		ctx = context.Background()
	)

	t.Cleanup(func() {
		_ = r.Destroy(ctx)
	})

	if err := r.Init(ctx); err != nil {
		t.Fatal(err)
	}

	cust := pay.Customer{
		Name:       "Test Customer",
		Email:      "test@cibera.com.mx",
		Provider:   "test",
		ProviderID: "1",
	}

	if err := r.AddCustomer(&cust); err != nil {
		t.Fatal(err)
	}

	found, err := r.GetCustomerByEmail(cust.Email)
	if err != nil {
		t.Fatal(err)
	}

	if found.ID != cust.ID {
		t.Fatal("expected to find customer by email")
	}
}

func getPgxEntityRepo(t *testing.T) *pay.EntityRepo {
	db, err := sql.Open("pgx", os.Getenv("TEST_CONNECTION_STRING"))
	if err != nil {
		t.Fatal(err)
	}

	r := pay.NewEntityRepo(db)
	return r
}
