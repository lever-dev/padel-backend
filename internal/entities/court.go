package entities

import (
	"time"

	"github.com/google/uuid"
)

type Court struct {
	ID             string
	OrganizationID string
	Name           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewCourt(orgID, name string) *Court {
	now := time.Now()
	return &Court{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		Name:           name,
		CreatedAt:      now,
	}
}
