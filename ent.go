package pay

import (
	"time"
)

const (
	PricingAnnual  = "annual"
	PricingMonthly = "monthly"
	PricingOnce    = "once"
)

type Price struct {
	ID         int64 // internal identifier
	PlanID     int64 // id of associated plan
	Provider   string
	ProviderID string
	Amount     int64  // in lowest common denominator
	Currency   string // three letter currency code
	Schedule   string // one of PricingAnnual | PricingMonthly | PricingOnce
}

// Plan
type Plan struct {
	ID         int64
	Name       string
	Provider   string
	ProviderID string
	TrialDays  int64
	Active     bool
	Features   []string
}

func (p *Plan) Table() string {
	return "pay.plan"
}

// HasTrial returns true if plan has a set amount of trial days
func (p *Plan) HasTrial() bool {
	return p.TrialDays > 0
}

// TrialEnd returns the time at which the trial would end if it started now
func (p *Plan) TrialEnd() time.Time {
	return time.Now().Add(time.Hour * 24 * time.Duration(p.TrialDays))
}

// Customer represents a paying customer attached to a third party service like stripe or paypal
type Customer struct {
	ID         int64  // internal (to this service)
	ProviderID string // external providers id
	Provider   string // the provider for this customer
	Name       string // customers name
	Email      string // customers email
}

func (c *Customer) Table() string {
	return "pay.customer"
}

// Subscription represents a customers subscription to a Plan
type Subscription struct {
	ID         int64
	CustomerID int64
	PlanID     int64
	Provider   string
	ProviderID string
	Active     bool
	// Plan attached to this subscription
	Plan *Plan `db:"-"`
	// Customer attached to this subscription
	Customer *Customer `db:"-"`
}
