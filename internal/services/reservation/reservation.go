package reservation

import (
	"context"
	"sync"
	"time"

	"github.com/lever-dev/padel-backend/internal/entities"
)

type Service struct {
	reservationsRepo ReservationsRepository
	courtMutexes     map[string]*sync.Mutex
	mu               sync.Mutex
}

func NewService(repo ReservationsRepository) *Service {
	return &Service{
		reservationsRepo: repo,
		courtMutexes:     make(map[string]*sync.Mutex),
	}
}

func (s *Service) ReserveCourt(ctx context.Context, courtID string, reservation entities.Reservation) error {
	// TO DO: validation

	courtMutex := s.getCourtMutex(courtID)
	courtMutex.Lock()
	defer courtMutex.Unlock()

	// TO DO: check conflict

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

func (s *Service) getCourtMutex(courtID string) *sync.Mutex {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.courtMutexes[courtID]; !exists {
		s.courtMutexes[courtID] = &sync.Mutex{}
	}
	return s.courtMutexes[courtID]
}
