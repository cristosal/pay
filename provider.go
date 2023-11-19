package pay

import (
	"errors"
	"net/http"

	"github.com/cristosal/dbx"
	"github.com/cristosal/pgxx"
)

var (
	ErrNotFound = errors.New("not found")
	ErrNoPlan   = errors.New("no plan")
)

type Provider interface {
	Init() error
	Sync() error
	AddCustomer(uid pgxx.ID, name, email string) (*Customer, error)
	PlanByUser(uid pgxx.ID) (*Plan, error)
	ListPlans() ([]Plan, error)
	VerifyCheckout(string) error
	Checkout(*CheckoutRequest) (url string, err error)
	Webhook() http.HandlerFunc
	OnSubscriptionAdded(func(s *Subscription))
	OnSubscriptionUpdated(func(s *Subscription))
}

// CheckoutRequest
type CheckoutRequest struct {
	UserID      dbx.ID `json:"user_id"`
	Plan        string `json:"plan_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	RedirectURL string `json:"redirect_url"`
}
