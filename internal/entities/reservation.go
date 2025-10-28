package entities

import "time"

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
