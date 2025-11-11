package reservation

import (
	"context"
	"fmt"
	"time"

	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/rs/zerolog/log"
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

	if err := s.reservationsRepo.Create(ctx, reservation); err != nil {
		return fmt.Errorf("create reservation: %w", err)
	}

	return nil
}

func (s *Service) ListReservations(
	ctx context.Context,
	courtID string,
	from, to time.Time,
) ([]entities.Reservation, error) {
	revs, err := s.reservationsRepo.ListByCourtAndTimeRange(ctx, courtID, from, to)
	if err != nil {
		return nil, fmt.Errorf("list reservations by court and time range: %w", err)
	}
	return revs, nil
}

func (s *Service) CancelReservation(ctx context.Context, reservationID string, cancelledBy string) error {
	if err := s.reservationsRepo.CancelReservation(ctx, reservationID, cancelledBy); err != nil {
		return fmt.Errorf("cancel reservation: %w", err)
	}

	return nil
}

func (s *Service) GetReservation(ctx context.Context, courtID, reservationID string) (*entities.Reservation, error) {
	rev, err := s.reservationsRepo.GetByID(ctx, reservationID)
	if err != nil {
		return nil, fmt.Errorf("get reservation by id: %w", err)
	}

	if rev.CourtID != courtID {
		log.Error().Err(err).
			Str("reservationID", reservationID).
			Str("courtID", courtID).
			Msg("reservation's court ID does not match the requested court ID")
		return nil, fmt.Errorf(
			"reservation court ids mismatch, reservation id: %s, court id: %s",
			reservationID,
			courtID,
		)
	}
	return rev, nil
}
