package reservation

import (
	"context"
	"fmt"
	"time"

	"github.com/lever-dev/padel-backend/internal/entities"
)

type Service struct {
	reservationsRepo ReservationsRepository
	locker           Locker
}

func NewService(repo ReservationsRepository, l Locker) *Service {
	return &Service{
		reservationsRepo: repo,
		locker:           l,
	}
}

func (s *Service) ReserveCourt(ctx context.Context, courtID string, reservation entities.Reservation) error {
	if err := s.locker.Lock(ctx, courtID); err != nil {
		return fmt.Errorf("failed to lock court: %w", err)
	}

	defer func() {
		if err := s.locker.Unlock(ctx, courtID); err != nil {
			// TO DO: Maybe log
		}
	}()

	hasConflict, err := s.reservationsRepo.HasOverlapping(
		ctx,
		courtID,
		reservation.ReservedFrom,
		reservation.ReservedTo,
	)
	if err != nil {
		return fmt.Errorf("failed to check overlapping reservations: %w", err)
	}

	if hasConflict {
		return entities.ErrCourtAlreadyReserved
	}

	return s.reservationsRepo.Create(ctx, &reservation)
}

func (s *Service) ListReservations(
	ctx context.Context,
	courtID string,
	from, to time.Time,
) ([]entities.Reservation, error) {
	return nil, nil
}

func (s *Service) CancelReservation(
	ctx context.Context,
	courtID string,
	reservation entities.Reservation,
	cancelledBy string,
) ([]entities.Reservation, error) {
	return nil, nil
}
