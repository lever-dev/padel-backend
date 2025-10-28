package reservation

import (
	"time"

	"github.com/lever-dev/padel-backend/internal/entities"
)

type dto struct {
	ID      string
	CourtID string

	Status       string
	ReservedFrom time.Time
	ReservedTo   time.Time
	ReservedBy   string
	CancelledBy  string

	CreatedAt time.Time
}

func newDTO(r *entities.Reservation) dto {
	return dto{
		ID:           r.ID,
		CourtID:      r.CourtID,
		Status:       string(r.Status),
		ReservedFrom: r.ReservedFrom,
		ReservedTo:   r.ReservedTo,
		ReservedBy:   r.ReservedBy,
		CancelledBy:  r.CancelledBy,
		CreatedAt:    r.CreatedAt,
	}
}

func (d dto) toEntity() entities.Reservation {
	return entities.Reservation{
		ID:           d.ID,
		CourtID:      d.CourtID,
		Status:       entities.ReservationStatus(d.Status),
		ReservedFrom: d.ReservedFrom,
		ReservedTo:   d.ReservedTo,
		ReservedBy:   d.ReservedBy,
		CancelledBy:  d.CancelledBy,
		CreatedAt:    d.CreatedAt,
	}
}
