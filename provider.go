package pay

import (
	"errors"
	"net/http"

	"github.com/cristosal/dbx"
)

var (
	ErrNotFound = errors.New("not found")
	ErrNoPlan   = errors.New("no plan")
)

type Provider interface {
	Init() error
	Sync() error
	AddCustomer(uid dbx.ID, name, email string) (*Customer, error)
	PlanByUser(uid dbx.ID) (*Plan, error)
	VerifyCheckout(string) error
	Checkout(*CheckoutRequest) (url string, err error)
	Webhook() http.HandlerFunc
	OnSubscriptionAdded(func(s *Subscription))
	OnSubscriptionUpdated(func(s *Subscription))
}
