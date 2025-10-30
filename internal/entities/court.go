package entities

import "time"

type Court struct {
	ID             string
	OrganizationID string
	Name           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
