package organization

import (
	"time"

	"github.com/lever-dev/padel-backend/internal/entities"
)

type dto struct {
	ID        string
	Name      string
	City      string
	CreatedAt time.Time
}

func newDTO(o *entities.Organization) dto {
	return dto{
		ID:        o.ID,
		Name:      o.Name,
		City:      o.City,
		CreatedAt: o.CreatedAt,
	}
}

func (d dto) toEntity() entities.Organization {
	return entities.Organization{
		ID:        d.ID,
		Name:      d.Name,
		City:      d.City,
		CreatedAt: d.CreatedAt,
	}
}
