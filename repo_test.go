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

	r := pay.NewRepo(db)
	if err := r.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
}
