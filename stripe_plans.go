package pay

import (
	"log"
	"strconv"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/product"
)

func (s *StripeService) syncPlans() error {
	plans, err := s.fetchPlans()
	if err != nil {
		return err
	}

	for _, p := range plans {
		// check if we have it
		found, _ := s.Plans().ByProviderID(p.ProviderID)

		if found == nil {
			if err := s.Plans().Add(&p); err != nil {
				log.Printf("error adding plan during sync: %v", err)
			}

			// get next
			continue
		}

		p.ID = found.ID
		if err := s.Plans().Update(&p); err != nil {
			log.Printf("error updating plan during sync: %v", err)
		}
	}

	return nil
}

// fetchPlans from stripe
func (s *StripeService) fetchPlans() ([]Plan, error) {
	params := &stripe.ProductListParams{}
	params.AddExpand("data.default_price")

	it := product.List(params)

	var plans []Plan
	for it.Next() {
		p := it.Product()
		if p.DefaultPrice == nil {
			continue
		}

		days, _ := strconv.ParseInt(p.Metadata["trial_days"], 10, 32)

		plans = append(plans, Plan{
			ProviderID: p.ID,
			Provider:   StripeProvider,
			Name:       p.Name,
			Active:     p.Active,
			Price:      p.DefaultPrice.UnitAmount,
			TrialDays:  days,
		})
	}

	if it.Err() != nil {
		return nil, it.Err()
	}

	return plans, nil
}
