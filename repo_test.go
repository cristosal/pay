package pay_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/cristosal/pay"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func getEntityRepo(t *testing.T) *pay.Repo {
	db, err := sql.Open("pgx", "")
	if err != nil {
		t.Fatal(err)
	}

	r := pay.NewEntityRepo(db)

	if err := r.Init(context.Background()); err != nil {
		t.Fatal(fmt.Errorf("failed to initialize repository: %w", err))
	}

	return r
}

func TestCustomerSubscriptions(t *testing.T) {
	r := getEntityRepo(t)
	t.Cleanup(func() { r.ClearCustomers() })

	cust := pay.Customer{
		Name:       "Test Customer",
		Email:      "test@cibera.com.mx",
		Provider:   "test",
		ProviderID: "1",
	}

	// probably we want to replicate pricing as well.
	r.AddCustomer(&cust)
}

func TestCustomerRepo(t *testing.T) {
	r := getEntityRepo(t)

	t.Cleanup(func() { r.ClearCustomers() })

	c := &pay.Customer{
		Name:       "Test Customer",
		Email:      "test@cibera.com.mx",
		ProviderID: "123",
		Provider:   "test",
	}

	if err := r.AddCustomer(c); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		r.DeleteCustomerByProvider("test", "123")
	})

	c2, err := r.GetCustomerByID(c.ID)
	if err != nil {
		t.Fatal(err)
	}

	if c2.Name != c.Name {
		t.Fatal("expected to match customers")
	}

}
