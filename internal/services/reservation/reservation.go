package reservation

import (
	"context"
	"errors"
	"fmt"
	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/rs/zerolog/log"
	"time"
)

type Service struct {
	reservationsRepo ReservationsRepository
	locker           Locker
}

func NewService(repo ReservationsRepository, locker Locker) *Service {
	return &Service{
		reservationsRepo: repo,
		locker:           locker,
	}
}

func (s *Service) ReserveCourt(ctx context.Context, courtID string, reservation *entities.Reservation) error {
	if err := s.locker.Lock(ctx, courtID); err != nil {
		return fmt.Errorf("failed to lock court: %w", err)
	}

	defer func() {
		if err := s.locker.Unlock(ctx, courtID); err != nil {
			log.Error().Err(err).Str("court_id", courtID).Msg("failed to unlock court")
		}
	}()

	overlapping, err := s.reservationsRepo.ListByCourtAndTimeRange(
		ctx,
		courtID,
		reservation.ReservedFrom,
		reservation.ReservedTo,
	)
	if err != nil {
		return fmt.Errorf("failed to check overlapping reservations: %w", err)
	}

	if len(overlapping) > 0 {
		return entities.ErrCourtAlreadyReserved
	}

	return s.reservationsRepo.Create(ctx, reservation)
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
	reservationID string,
	cancelledBy string,
) error {
	reservation, err := s.reservationsRepo.GetByID(ctx, reservationID)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			return entities.ErrNotFound
		}
		return fmt.Errorf("failed to get reservation: %w", err)
	}

	if reservation.Status == entities.CancelledReservationStatus {
		return fmt.Errorf("reservation already cancelled")
	}

	if err := s.reservationsRepo.CancelReservation(ctx, reservationID, cancelledBy); err != nil {
		return fmt.Errorf("failed to cancel reservation: %w", err)
	}

	return nil
}
