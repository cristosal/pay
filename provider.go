package pay

import (
	"errors"
	"net/http"
)

var (
	ErrNotFound = errors.New("not found")
	ErrNoPlan   = errors.New("no plan")
)

type Provider interface {
	Init() error
	Sync() error
	VerifyCheckout(string) error
	Checkout(*CheckoutRequest) (url string, err error)
	Webhook() http.HandlerFunc
	OnSubscriptionAdded(func(s *Subscription))
	OnSubscriptionUpdated(func(s *Subscription))
}
