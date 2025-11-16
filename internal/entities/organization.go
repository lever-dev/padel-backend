package entities

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID        string
	Name      string
	City      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewOrganization(name, city string) *Organization {
	return &Organization{
		ID:        uuid.NewString(),
		Name:      name,
		City:      city,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}
