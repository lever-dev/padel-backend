package court

import (
	"github.com/lever-dev/padel-backend/internal/entities"
	"time"
)

type dto struct {
	ID             string
	OrganizationID string
	Name           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func newDTO(c *entities.Court) dto {
	return dto{
		ID:             c.ID,
		OrganizationID: c.OrganizationID,
		Name:           c.Name,
		UpdatedAt:      c.UpdatedAt,
	}
}

func (d dto) toEntity() entities.Court {
	return entities.Court{
		ID:             d.ID,
		OrganizationID: d.OrganizationID,
		Name:           d.Name,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}
