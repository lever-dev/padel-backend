package reservation_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/lever-dev/padel-backend/internal/services/reservation"
	"github.com/lever-dev/padel-backend/internal/services/reservation/mocks"
	"github.com/stretchr/testify/suite"
)

type ServiceSuite struct {
	suite.Suite
	ctrl *gomock.Controller
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
}

func (s *ServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *ServiceSuite) TestReserveCourt() {
	tests := []struct {
		name        string
		courtID     string
		reservation *entities.Reservation
		setupMocks  func(mockRepo *mocks.MockReservationsRepository, res *entities.Reservation)
		wantErr     bool
	}{
		{
			name:    "success",
			courtID: "court-1",
			reservation: &entities.Reservation{
				ID:           "reservation-1",
				CourtID:      "court-1",
				ReservedBy:   "user-1",
				Status:       entities.PendingReservationStatus,
				ReservedFrom: time.Now().Add(1 * time.Hour),
				ReservedTo:   time.Now().Add(2 * time.Hour),
				CreatedAt:    time.Now(),
			},
			setupMocks: func(mockRepo *mocks.MockReservationsRepository, res *entities.Reservation) {
				mockRepo.EXPECT().
					ListByCourtAndTimeRange(gomock.Any(), "court-1", res.ReservedFrom, res.ReservedTo).
					Return([]entities.Reservation{}, nil)
				mockRepo.EXPECT().
					Create(gomock.Any(), res).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "conflict - court reserved",
			courtID: "court-1",
			reservation: &entities.Reservation{
				ID:           "reservation-2",
				CourtID:      "court-1",
				ReservedBy:   "user-2",
				Status:       entities.PendingReservationStatus,
				ReservedFrom: time.Now().Add(1 * time.Hour),
				ReservedTo:   time.Now().Add(2 * time.Hour),
				CreatedAt:    time.Now(),
			},
			setupMocks: func(mockRepo *mocks.MockReservationsRepository, res *entities.Reservation) {
				mockRepo.EXPECT().
					ListByCourtAndTimeRange(gomock.Any(), "court-1", res.ReservedFrom, res.ReservedTo).
					Return([]entities.Reservation{
						{
							ID:           "existing-reservation",
							CourtID:      "court-1",
							ReservedBy:   "other-user",
							Status:       entities.ReservedReservationStatus,
							ReservedFrom: time.Now().Add(1 * time.Hour),
							ReservedTo:   time.Now().Add(2 * time.Hour),
						},
					}, nil)
			},
			wantErr: true,
		},
		{
			name:    "create error",
			courtID: "court-1",
			reservation: &entities.Reservation{
				ID:           "reservation-3",
				CourtID:      "court-1",
				ReservedBy:   "user-3",
				Status:       entities.PendingReservationStatus,
				ReservedFrom: time.Now().Add(1 * time.Hour),
				ReservedTo:   time.Now().Add(2 * time.Hour),
				CreatedAt:    time.Now(),
			},
			setupMocks: func(mockRepo *mocks.MockReservationsRepository, res *entities.Reservation) {
				mockRepo.EXPECT().
					ListByCourtAndTimeRange(gomock.Any(), "court-1", res.ReservedFrom, res.ReservedTo).
					Return([]entities.Reservation{}, nil)
				mockRepo.EXPECT().
					Create(gomock.Any(), res).
					Return(fmt.Errorf("database insert error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			mockRepo := mocks.NewMockReservationsRepository(s.ctrl)
			locker := reservation.NewLocalLocker()
			service := reservation.NewService(mockRepo, locker)

			tt.setupMocks(mockRepo, tt.reservation)

			err := service.ReserveCourt(ctx, tt.courtID, tt.reservation)

			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *ServiceSuite) TestReserveCourt_ConcurrentReservations() {
	ctx := context.Background()
	courtID := "court-1"
	reservedFrom := time.Now().Add(1 * time.Hour)
	reservedTo := time.Now().Add(2 * time.Hour)

	mockRepo := mocks.NewMockReservationsRepository(s.ctrl)
	locker := reservation.NewLocalLocker()
	service := reservation.NewService(mockRepo, locker)

	firstCall := mockRepo.EXPECT().
		ListByCourtAndTimeRange(ctx, courtID, gomock.Any(), gomock.Any()).
		Return([]entities.Reservation{}, nil)

	mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil).Times(1).After(firstCall)

	mockRepo.EXPECT().
		ListByCourtAndTimeRange(ctx, courtID, gomock.Any(), gomock.Any()).
		Return([]entities.Reservation{
			{
				ID:           "existing",
				CourtID:      courtID,
				ReservedBy:   "user-1",
				Status:       entities.ReservedReservationStatus,
				ReservedFrom: reservedFrom,
				ReservedTo:   reservedTo,
			},
		}, nil).
		AnyTimes().
		After(firstCall)

	var (
		wg      sync.WaitGroup
		results = make(chan error)

		wantSuccess = 1
		wantErrors  = 100
		total       = wantSuccess + wantErrors
	)

	for range total {
		wg.Go(func() {
			reservation := &entities.Reservation{
				ID:           "reservation-1",
				CourtID:      courtID,
				ReservedBy:   "user-1",
				Status:       entities.PendingReservationStatus,
				ReservedFrom: reservedFrom,
				ReservedTo:   reservedTo,
				CreatedAt:    time.Now(),
			}

			results <- service.ReserveCourt(ctx, courtID, reservation)
		})
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	successCount := 0
	errorCount := 0

	for err := range results {
		if err != nil {
			errorCount++
			continue
		}

		successCount++
	}

	s.Equal(wantSuccess, successCount)
	s.Equal(wantErrors, errorCount)
}

func (s *ServiceSuite) TestReserveCourt_DifferentCourts() {
	ctx := context.Background()

	mockRepo := mocks.NewMockReservationsRepository(s.ctrl)
	locker := reservation.NewLocalLocker()
	service := reservation.NewService(mockRepo, locker)

	mockRepo.EXPECT().
		ListByCourtAndTimeRange(ctx, "court-1", gomock.Any(), gomock.Any()).
		Return([]entities.Reservation{}, nil)

	mockRepo.EXPECT().
		ListByCourtAndTimeRange(ctx, "court-2", gomock.Any(), gomock.Any()).
		Return([]entities.Reservation{}, nil)

	mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil).Times(2)

	var wg sync.WaitGroup
	results := make(chan error)

	for i, courtID := range []string{"court-1", "court-2"} {
		wg.Add(1)
		go func(id int, cID string) {
			defer wg.Done()

			reservation := &entities.Reservation{
				ID:           "reservation-1",
				CourtID:      cID,
				ReservedBy:   "user-1",
				Status:       entities.PendingReservationStatus,
				ReservedFrom: time.Now().Add(1 * time.Hour),
				ReservedTo:   time.Now().Add(2 * time.Hour),
				CreatedAt:    time.Now(),
			}

			results <- service.ReserveCourt(ctx, cID, reservation)
		}(i, courtID)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for err := range results {
		s.NoError(err)
	}
}

func (s *ServiceSuite) TestReserveCourt_LockIsReleased() {
	ctx := context.Background()
	courtID := "court-1"

	mockRepo := mocks.NewMockReservationsRepository(s.ctrl)
	locker := reservation.NewLocalLocker()
	service := reservation.NewService(mockRepo, locker)

	reservation := &entities.Reservation{
		ID:           "reservation-1",
		CourtID:      courtID,
		ReservedBy:   "user-1",
		Status:       entities.PendingReservationStatus,
		ReservedFrom: time.Now().Add(1 * time.Hour),
		ReservedTo:   time.Now().Add(2 * time.Hour),
		CreatedAt:    time.Now(),
	}

	mockRepo.EXPECT().
		ListByCourtAndTimeRange(ctx, courtID, reservation.ReservedFrom, reservation.ReservedTo).
		Return([]entities.Reservation{}, nil)
	mockRepo.EXPECT().Create(ctx, reservation).Return(fmt.Errorf("fail"))

	err := service.ReserveCourt(ctx, courtID, reservation)
	s.Require().Error(err)

	mockRepo.EXPECT().
		ListByCourtAndTimeRange(ctx, courtID, reservation.ReservedFrom, reservation.ReservedTo).
		Return([]entities.Reservation{}, nil)
	mockRepo.EXPECT().Create(ctx, reservation).Return(nil)

	err = service.ReserveCourt(ctx, courtID, reservation)
	s.NoError(err)
}

func (s *ServiceSuite) TestCancelReservation() {
	ctx := context.Background()
	reservationID := "reservation-1"
	cancelledBy := "user-123"

	tests := []struct {
		name       string
		setupMocks func(mockRepo *mocks.MockReservationsRepository)
		wantErr    error
	}{
		{
			name: "success",
			setupMocks: func(mockRepo *mocks.MockReservationsRepository) {
				mockRepo.EXPECT().
					CancelReservation(ctx, reservationID, cancelledBy).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "reservation not found",
			setupMocks: func(mockRepo *mocks.MockReservationsRepository) {
				mockRepo.EXPECT().
					CancelReservation(ctx, reservationID, cancelledBy).
					Return(entities.ErrNotFound)
			},
			wantErr: entities.ErrNotFound,
		},
		{
			name: "internal error",
			setupMocks: func(mockRepo *mocks.MockReservationsRepository) {
				mockRepo.EXPECT().
					CancelReservation(ctx, reservationID, cancelledBy).
					Return(fmt.Errorf("db error"))
			},
			wantErr: fmt.Errorf("db error"),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockRepo := mocks.NewMockReservationsRepository(s.ctrl)
			locker := reservation.NewLocalLocker()
			service := reservation.NewService(mockRepo, locker)

			tt.setupMocks(mockRepo)

			err := service.CancelReservation(ctx, reservationID, cancelledBy)

			if tt.wantErr != nil {
				s.Require().Error(err)
				if errors.Is(tt.wantErr, entities.ErrNotFound) {
					s.ErrorIs(err, entities.ErrNotFound)
				} else {
					s.Contains(err.Error(), "cancel reservation")
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *ServiceSuite) TestGetListReservations() {
	ctx := context.Background()
	courtID := "court-1"
	from := time.Now().Add(1 * time.Hour)
	to := time.Now().Add(24 * time.Hour)

	rev := entities.Reservation{
		ID:           "reservation-1",
		CourtID:      courtID,
		ReservedBy:   "user-1",
		Status:       entities.PendingReservationStatus,
		ReservedFrom: time.Now().Add(1 * time.Hour),
		ReservedTo:   time.Now().Add(2 * time.Hour),
		CreatedAt:    time.Now(),
	}

	tests := []struct {
		name       string
		setupMocks func(*mocks.MockReservationsRepository)
		wantErr    bool
		wantRevs   []entities.Reservation
	}{
		{
			name: "success",
			setupMocks: func(mockRepo *mocks.MockReservationsRepository) {
				mockRepo.EXPECT().ListByCourtAndTimeRange(
					ctx,
					courtID,
					from,
					to,
				).Return([]entities.Reservation{rev}, nil)
			},
			wantErr:  false,
			wantRevs: []entities.Reservation{rev},
		},
		{
			name: "success - empty list",
			setupMocks: func(mockRepo *mocks.MockReservationsRepository) {
				mockRepo.EXPECT().ListByCourtAndTimeRange(
					ctx,
					courtID,
					from,
					to,
				).Return([]entities.Reservation{}, nil)
			},
			wantErr:  false,
			wantRevs: []entities.Reservation{},
		},
		{
			name: "internal error",
			setupMocks: func(mockRepo *mocks.MockReservationsRepository) {
				mockRepo.EXPECT().ListByCourtAndTimeRange(
					ctx,
					courtID,
					from,
					to,
				).Return(nil, fmt.Errorf("error"))
			},
			wantErr:  true,
			wantRevs: nil,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockRepo := mocks.NewMockReservationsRepository(s.ctrl)
			locker := reservation.NewLocalLocker()
			service := reservation.NewService(mockRepo, locker)

			tt.setupMocks(mockRepo)

			listRevs, err := service.ListReservations(ctx, courtID, from, to)

			if tt.wantErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}

			s.Equal(tt.wantRevs, listRevs)
		})
	}
}
