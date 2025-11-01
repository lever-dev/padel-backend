package reservation_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/lever-dev/padel-backend/internal/services/reservation"
	"github.com/lever-dev/padel-backend/internal/services/reservation/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type ServiceSuite struct {
	suite.Suite

	ctrl     *gomock.Controller
	mockRepo *mocks.MockReservationsRepository
	locker   reservation.Locker
	service  *reservation.Service
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = mocks.NewMockReservationsRepository(s.ctrl)
	s.locker = reservation.NewLocalLocker()
	s.service = reservation.NewService(s.mockRepo, s.locker)
}

func (s *ServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *ServiceSuite) TestReserveCourt_Success() {
	ctx := context.Background()
	courtID := "court-1"

	reservation := entities.Reservation{
		ID:           "reservation-1",
		CourtID:      courtID,
		ReservedBy:   "user-1",
		Status:       entities.PendingReservationStatus,
		ReservedFrom: time.Now().Add(1 * time.Hour),
		ReservedTo:   time.Now().Add(2 * time.Hour),
		CreatedAt:    time.Now(),
	}

	s.mockRepo.EXPECT().
		HasOverlapping(ctx, courtID, reservation.ReservedFrom, reservation.ReservedTo).
		Return(false, nil)

	s.mockRepo.EXPECT().Create(ctx, &reservation).Return(nil)

	err := s.service.ReserveCourt(ctx, courtID, reservation)
	s.NoError(err)
}

func (s *ServiceSuite) TestReserveCourt_WithConflict() {
	ctx := context.Background()
	courtID := "court-1"

	reservation := entities.Reservation{
		ID:           "reservation-1",
		CourtID:      courtID,
		ReservedBy:   "user-1",
		Status:       entities.PendingReservationStatus,
		ReservedFrom: time.Now().Add(1 * time.Hour),
		ReservedTo:   time.Now().Add(2 * time.Hour),
		CreatedAt:    time.Now(),
	}

	s.mockRepo.EXPECT().HasOverlapping(ctx, courtID, reservation.ReservedFrom, reservation.ReservedTo).Return(true, nil)

	s.mockRepo.EXPECT().Create(ctx, &reservation).Times(0)

	err := s.service.ReserveCourt(ctx, courtID, reservation)
	s.Error(err)
	s.ErrorIs(err, entities.ErrCourtAlreadyReserved)
}

func (s *ServiceSuite) TestReserveCourt_CreateError() {
	ctx := context.Background()
	courtID := "court-1"

	reservation := entities.Reservation{
		ID:           "reservation-1",
		CourtID:      courtID,
		ReservedBy:   "user-1",
		Status:       entities.PendingReservationStatus,
		ReservedFrom: time.Now().Add(1 * time.Hour),
		ReservedTo:   time.Now().Add(2 * time.Hour),
		CreatedAt:    time.Now(),
	}

	s.mockRepo.EXPECT().
		HasOverlapping(ctx, courtID, reservation.ReservedFrom, reservation.ReservedTo).
		Return(false, nil)

	myErr := fmt.Errorf("Error")
	s.mockRepo.EXPECT().Create(ctx, &reservation).Return(myErr)

	err := s.service.ReserveCourt(ctx, courtID, reservation)
	s.Error(err)
	s.ErrorIs(err, myErr)
}

func (s *ServiceSuite) TestReserveCourt_ConcurrentReservations() {
	ctx := context.Background()
	courtID := "court-1"

	firstCall := s.mockRepo.EXPECT().HasOverlapping(ctx, courtID, gomock.Any(), gomock.Any()).Return(false, nil)

	s.mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil).Times(1).After(firstCall)

	s.mockRepo.EXPECT().
		HasOverlapping(ctx, courtID, gomock.Any(), gomock.Any()).
		Return(true, nil).
		AnyTimes().
		After(firstCall)

	var wg sync.WaitGroup
	results := make(chan error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			reservation := entities.Reservation{
				ID:           "reservation-1",
				CourtID:      courtID,
				ReservedBy:   "user-1",
				Status:       entities.PendingReservationStatus,
				ReservedFrom: time.Now().Add(1 * time.Hour),
				ReservedTo:   time.Now().Add(2 * time.Hour),
				CreatedAt:    time.Now(),
			}

			err := s.service.ReserveCourt(ctx, courtID, reservation)
			results <- err
		}()
	}

	wg.Wait()
	close(results)

	successCount := 0
	errorCount := 0

	for err := range results {
		if err == nil {
			successCount++
		} else {
			errorCount++
		}
	}

	s.Equal(1, successCount)
	s.Equal(1, errorCount)
}

func (s *ServiceSuite) TestReserveCourt_DifferentCourts() {
	ctx := context.Background()

	s.mockRepo.EXPECT().HasOverlapping(ctx, "court-1", gomock.Any(), gomock.Any()).Return(false, nil)

	s.mockRepo.EXPECT().HasOverlapping(ctx, "court-2", gomock.Any(), gomock.Any()).Return(false, nil)

	s.mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil).Times(2)

	var wg sync.WaitGroup
	results := make(chan error, 2)

	for i, courtID := range []string{"court-1", "court-2"} {
		wg.Add(1)
		go func(id int, cID string) {
			defer wg.Done()

			reservation := entities.Reservation{
				ID:           "reservation-1",
				CourtID:      cID,
				ReservedBy:   "user-1",
				Status:       entities.PendingReservationStatus,
				ReservedFrom: time.Now().Add(1 * time.Hour),
				ReservedTo:   time.Now().Add(2 * time.Hour),
				CreatedAt:    time.Now(),
			}

			err := s.service.ReserveCourt(ctx, cID, reservation)
			results <- err
		}(i, courtID)
	}

	wg.Wait()
	close(results)

	for err := range results {
		s.NoError(err)
	}
}

func (s *ServiceSuite) TestReserveCourt_LockIsReleased() {
	ctx := context.Background()
	courtID := "court-1"

	reservation := entities.Reservation{
		ID:           "reservation-1",
		CourtID:      courtID,
		ReservedBy:   "user-1",
		Status:       entities.PendingReservationStatus,
		ReservedFrom: time.Now().Add(1 * time.Hour),
		ReservedTo:   time.Now().Add(2 * time.Hour),
		CreatedAt:    time.Now(),
	}

	s.mockRepo.EXPECT().
		HasOverlapping(ctx, courtID, reservation.ReservedFrom, reservation.ReservedTo).
		Return(false, nil)
	s.mockRepo.EXPECT().Create(ctx, &reservation).Return(fmt.Errorf("fail"))

	err := s.service.ReserveCourt(ctx, courtID, reservation)
	s.Require().Error(err)

	s.mockRepo.EXPECT().
		HasOverlapping(ctx, courtID, reservation.ReservedFrom, reservation.ReservedTo).
		Return(false, nil)
	s.mockRepo.EXPECT().Create(ctx, &reservation).Return(nil)

	err = s.service.ReserveCourt(ctx, courtID, reservation)
	s.NoError(err)
}
