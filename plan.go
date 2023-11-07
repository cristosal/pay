package pay

import "github.com/cristosal/pgxx"

type Plan struct {
	ID         pgxx.ID `db:"id"`
	ProviderID string  `db:"provider_id"`
	Provider   string  `db:"provider"`
	Name       string  `db:"name"`
	Active     bool    `db:"active"` // or not but should be active
	Price      int64   `db:"price"`  // monthly price
}

func (p *Plan) DisplayPrice() float64 {
	return float64(p.Price) / 100
}

// IsFree returns wether the plan requires any payment to use
func (p *Plan) IsFree() bool {
	return p.Price == 0
}
