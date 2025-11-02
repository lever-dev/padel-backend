package reservation

import (
	"context"
	"fmt"
	"time"

	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/rs/zerolog"
)

type Service struct {
	reservationsRepo ReservationsRepository
	locker           Locker
	logger           zerolog.Logger
}

func NewService(repo ReservationsRepository, locker Locker, logger zerolog.Logger) *Service {
	return &Service{
		reservationsRepo: repo,
		locker:           locker,
		logger:           logger,
	}
}

func (s *Service) ReserveCourt(ctx context.Context, courtID string, reservation entities.Reservation) error {
	if err := s.locker.Lock(ctx, courtID); err != nil {
		return fmt.Errorf("failed to lock court: %w", err)
	}

	defer func() {
		if err := s.locker.Unlock(ctx, courtID); err != nil {
			s.logger.Error().Err(err).Str("court_id", courtID).Msg("failed to unlock court")
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
