package pay

import (
	"context"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/price"
)

func (s *StripeService) SyncPrices(ctx context.Context) error {
	it := price.List(nil)

	for it.Next() {
		p := it.Price()
		plan, _ := s.Repository().GetPlanByProviderID(p.Product.ID)

		if plan == nil {
			continue
		}

		if err := s.Repository().AddPrice(&Price{
			PlanID:     plan.ID,
			Amount:     p.UnitAmount,
			Provider:   ProviderStripe,
			ProviderID: p.ID,
			Currency:   string(p.Currency),
			Schedule:   s.getPricing(p),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (StripeService) getPricing(p *stripe.Price) string {
	switch p.Type {
	case stripe.PriceTypeOneTime:
		return PricingOnce
	case stripe.PriceTypeRecurring:
		switch p.Recurring.Interval {
		case stripe.PriceRecurringIntervalMonth:
			return PricingMonthly
		case stripe.PriceRecurringIntervalYear:
			return PricingAnnual
		default:
			return ""
		}
	default:
		return ""
	}
}
