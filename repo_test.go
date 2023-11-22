package pay_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/cristosal/pay"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestRepo(t *testing.T) {
	db, err := sql.Open("pgx", "")
	if err != nil {
		t.Fatal(err)
	}

	r := pay.NewEntityRepo(db)

	if err := r.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

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
		r.RemoveCustomerByProviderID("123")
	})

	c2, err := r.GetCustomerByID(c.ID)

	if err != nil {
		t.Fatal(err)
	}

	if c2.Name == c.Name {
		t.Fatal("expected to match customers")
	}

}
