package reservation_test

import (
	"context"
	"fmt"
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
	service  *reservation.Service
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = mocks.NewMockReservationsRepository(s.ctrl)
	s.service = reservation.NewService(s.mockRepo)
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
		ReservedTo:   time.Now().Add(1 * time.Hour),
		CreatedAt:    time.Now(),
	}

	s.mockRepo.EXPECT().Create(ctx, &reservation).Return(nil)

	err := s.service.ReserveCourt(ctx, courtID, reservation)
	s.NoError(err)
}

func (s *ServiceSuite) TestReserveCourt_RepoErr() {
	ctx := context.Background()
	courtID := "court-1"

	reservation := entities.Reservation{
		ID:           "reservation-1",
		CourtID:      courtID,
		ReservedBy:   "user-1",
		Status:       entities.PendingReservationStatus,
		ReservedFrom: time.Now().Add(1 * time.Hour),
		ReservedTo:   time.Now().Add(1 * time.Hour),
		CreatedAt:    time.Now(),
	}

	fail := fmt.Errorf("fail in repo")
	s.mockRepo.EXPECT().Create(ctx, &reservation).Return(fail)

	err := s.service.ReserveCourt(ctx, courtID, reservation)
	s.Error(err)
	s.ErrorIs(err, fail)
}

// func (s *ServiceSuite) TestReserveCourt_CheckMutex() {
// 	ctx := context.Background()
// 	courtID := "court-1"

// 	reservation := entities.Reservation{
// 		ID:           "reservation-1",
// 		CourtID:      courtID,
// 		ReservedBy:   "user-1",
// 		Status:       entities.PendingReservationStatus,
// 		ReservedFrom: time.Now().Add(1 * time.Hour),
// 		ReservedTo:   time.Now().Add(1 * time.Hour),
// 		CreatedAt:    time.Now(),
// 	}

// }
