package entities

import "time"

type Organization struct {
	ID        string
	Name      string
	City      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TO DO: NewOrganization constructor and validation
