package pay

import "github.com/cristosal/pgxx"

type (
	PlanRepo interface {
		Init() error
		List() ([]Plan, error)
		Add(*Plan) error
		ByID(pgxx.ID) (*Plan, error)
		ByName(string) (*Plan, error)
		ByProviderID(string) (*Plan, error)
		RemoveByProviderID(string) error
		Update(*Plan) error
	}

	PlanPgxRepo struct{ pgxx.DB }
)

func NewPlanPgxRepo(db pgxx.DB) *PlanPgxRepo {
	return &PlanPgxRepo{db}
}

func (r *PlanPgxRepo) Init() error {
	return pgxx.Exec(r, `create table if not exists plan (
		id serial primary key,
		name varchar(255) not null,
		provider varchar(255) not null,
		provider_id varchar(255) not null,
		active bool not null,
		trial_days int not null default 0,
		price int not null
	)`)
}

func (r *PlanPgxRepo) List() ([]Plan, error) {
	var plans []Plan

	if err := pgxx.Many(r, &plans, "where active = true order by price asc"); err != nil {
		return nil, err
	}

	return plans, nil
}

func (r *PlanPgxRepo) Add(p *Plan) error {
	return pgxx.Insert(r, p)
}

func (r *PlanPgxRepo) RemoveByProviderID(providerID string) error {
	return pgxx.Exec(r, "delete from plan where provider_id = $1", providerID)
}

func (r *PlanPgxRepo) Update(p *Plan) error {
	return pgxx.Update(r, p)
}

func (r *PlanPgxRepo) ByID(id pgxx.ID) (*Plan, error) {
	var p Plan
	if err := pgxx.One(r, &p, "where id = $1", id); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PlanPgxRepo) ByProviderID(providerID string) (*Plan, error) {
	var p Plan
	if err := pgxx.One(r, &p, "where provider_id = $1", providerID); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PlanPgxRepo) ByName(name string) (*Plan, error) {
	var p Plan
	if err := pgxx.One(r, &p, "where name = $1", name); err != nil {
		return nil, err
	}
	return &p, nil
}
