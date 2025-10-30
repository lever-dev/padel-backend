package reservation

import (
	"context"
	"time"

	"github.com/lever-dev/padel-backend/internal/entities"
)

type Service struct {
	reservationsRepo ReservationsRepository
}

func (s Service) ReserveCourt(ctx context.Context, courtID string, reservation entities.Reservation) error {
	/* TODO:
	1. obtain lock by courtID
	2. create reservation
	*/

	return nil
}

func (s Service) ListReservations(ctx context.Context, courtID string, from, to time.Time) ([]entities.Reservation, error) {
	return nil, nil
}

func (s Service) CancelReservation(ctx context.Context, courtID string, reservation entities.Reservation, cancelledBy string) ([]entities.Reservation, error) {
	return nil, nil
}
