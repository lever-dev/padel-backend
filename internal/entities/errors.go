package entities

import "errors"

var (
	ErrNotFound             = errors.New("not found")
	ErrCourtAlreadyReserved = errors.New("court is already reserved for this time slot")
)
