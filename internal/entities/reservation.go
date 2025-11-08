package entities

import (
	"time"

	"github.com/google/uuid"
)

type ReservationStatus string

const (
	PendingReservationStatus   ReservationStatus = "pending"
	ReservedReservationStatus  ReservationStatus = "reserved"
	CancelledReservationStatus ReservationStatus = "cancelled"
)

type Reservation struct {
	ID      string
	CourtID string

	Status       ReservationStatus
	ReservedFrom time.Time
	ReservedTo   time.Time
	ReservedBy   string
	CancelledBy  string

	CreatedAt time.Time
}

func NewReservation(courtID string, from, to time.Time, reservedBy string) *Reservation {
	return &Reservation{
		ID:           uuid.New().String(),
		CourtID:      courtID,
		ReservedFrom: from,
		ReservedTo:   to,
		ReservedBy:   reservedBy,
		CreatedAt:    time.Now(),
	}
}
