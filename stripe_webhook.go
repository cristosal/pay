package pay

import (
	"io"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
)

func (s *StripeService) Webhook() http.HandlerFunc {
	const MaxBodyBytes = int64(65536)
	whevents := make(chan stripe.Event)

	go func() {
		for event := range whevents {
			log.Printf("stripe webhook: recieved event: %s", event.Type)
			var err error

			switch event.Type {
			case "price.updated",
				"price.created",
				"product.created",
				"product.updated",
				"product.deleted":
				err = s.syncPlans() // resync plans
			case "customer.deleted":
				err = s.handleCustomerDeleted(event.Data)
			case "customer.subscription.created",
				"customer.subscription.updated",
				"customer.subscription.deleted":
				err = s.handleSubscriptionEvent(&event) // this is good
			}

			if err != nil {
				log.Printf("error handling stripe event: %v", err)
			}
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		payload, err := io.ReadAll(r.Body)

		if err != nil {
			log.Printf("Error reading request body: %v\n", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		// Pass the request body and Stripe-Signature header to ConstructEvent, along
		// with the webhook signing key.
		sig := r.Header.Get("Stripe-Signature")
		event, err := webhook.ConstructEvent(payload, sig, s.cfg.WebhookSecret)

		if err != nil {
			log.Printf("Error verifying webhook signature: %v\n", err)
			w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
			return
		}

		if s.EventRepo().Has(&event) {
			log.Printf("Already processed event with id %s", event.ID)
			w.WriteHeader(http.StatusOK)
			return
		}

		if err := s.EventRepo().Add(&event); err != nil {
			log.Printf("error while saving stripe event: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		whevents <- event
		w.WriteHeader(http.StatusOK)
	}
}
