package pay

import (
	"errors"
)

var (
	ErrNotFound = errors.New("not found")
	ErrNoPlan   = errors.New("no plan")
)
