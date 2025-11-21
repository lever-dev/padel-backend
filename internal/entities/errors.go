package entities

import "errors"

var (
	ErrNotFound                 = errors.New("not found")
	ErrCourtAlreadyReserved     = errors.New("court is already reserved for this time slot")
	ErrOrganizationAlreadyExist = errors.New("organization already exist")

	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("token has expired")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
