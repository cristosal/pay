package pay

import (
	"time"

	"github.com/cristosal/pgxx"
)

type Plan struct {
	ID         pgxx.ID `db:"id"`
	ProviderID string  `db:"provider_id"`
	Provider   string  `db:"provider"`
	Name       string  `db:"name"`
	Active     bool    `db:"active"` // or not but should be active
	Price      int64   `db:"price"`  // monthly price
	TrialDays  int64   `db:"trial_days"`
}

// HasTrial returns true if plan has a set amount of trial days
func (p *Plan) HasTrial() bool {
	return p.TrialDays > 0
}

// TrialEnd returns the time at which the trial would end if it started now
func (p *Plan) TrialEnd() time.Time {
	return time.Now().Add(time.Hour * 24 * time.Duration(p.TrialDays))
}

func (p *Plan) DisplayPrice() float64 {
	return float64(p.Price) / 100
}

// IsFree returns wether the plan requires any payment to use
func (p *Plan) IsFree() bool {
	return p.Price == 0
}
