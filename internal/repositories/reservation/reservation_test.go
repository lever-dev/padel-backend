package reservation_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/lever-dev/padel-backend/internal/repositories/reservation"
)

type repositorySuite struct {
	suite.Suite

	repo *reservation.Repository
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(repositorySuite))
}

func (s *repositorySuite) SetupTest() {
	// TODO: use config
	connString := "postgres://test:test@localhost:5432/db0?sslmode=disable"

	repo := reservation.NewRepository(connString)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := repo.Connect(ctx)
	require.NoError(s.T(), err)

	s.repo = repo
}

func (s *repositorySuite) TearDownTest() {
	if s.repo != nil {
		s.repo.Close()
	}
}

func (s *repositorySuite) TestCreateReservation() {
	ctx := context.Background()
	res := &entities.Reservation{
		ID:           "res-create-1",
		CourtID:      "court-1",
		Status:       entities.ReservedReservationStatus,
		ReservedFrom: time.Date(2024, 7, 20, 9, 0, 0, 0, time.UTC),
		ReservedTo:   time.Date(2024, 7, 20, 10, 0, 0, 0, time.UTC),
		ReservedBy:   "user-1",
		CreatedAt:    time.Date(2024, 7, 1, 8, 0, 0, 0, time.UTC),
	}

	err := s.repo.Create(ctx, res)
	s.Require().NoError(err)

	resDb, err := s.repo.GetByID(ctx, res.ID)
	s.Require().NoError(err)

	s.Require().Equal(res, resDb)
}

func (s *repositorySuite) TestListReservations() {
	ctx := context.Background()
	base := time.Date(2024, 7, 21, 9, 0, 0, 0, time.UTC)

	s.seedReservations(ctx, []*entities.Reservation{
		{
			ID:           "res-list-1",
			CourtID:      "court-1",
			Status:       entities.ReservedReservationStatus,
			ReservedFrom: base,
			ReservedTo:   base.Add(1 * time.Hour),
			ReservedBy:   "alice",
			CreatedAt:    base.Add(-24 * time.Hour),
		},
		{
			ID:           "res-list-2",
			CourtID:      "court-1",
			Status:       entities.PendingReservationStatus,
			ReservedFrom: base.Add(2 * time.Hour),
			ReservedTo:   base.Add(3 * time.Hour),
			ReservedBy:   "bob",
			CreatedAt:    base.Add(-23 * time.Hour),
		},
		{
			ID:           "res-list-3",
			CourtID:      "court-1",
			Status:       entities.CancelledReservationStatus,
			ReservedFrom: base.Add(4 * time.Hour),
			ReservedTo:   base.Add(5 * time.Hour),
			ReservedBy:   "carol",
			CancelledBy:  "carol",
			CreatedAt:    base.Add(-22 * time.Hour),
		},
		{
			ID:           "res-list-4",
			CourtID:      "court-2",
			Status:       entities.ReservedReservationStatus,
			ReservedFrom: base,
			ReservedTo:   base.Add(1 * time.Hour),
			ReservedBy:   "dave",
			CreatedAt:    base.Add(-21 * time.Hour),
		},
	})

	reservations, err := s.repo.ListByTimeRange(
		ctx,
		"court-1",
		base.Add(-30*time.Minute),
		base.Add(3*time.Hour+30*time.Minute),
	)
	s.Require().NoError(err)

	s.Len(reservations, 2)
	s.Equal("res-list-1", reservations[0].ID)
	s.Equal("res-list-2", reservations[1].ID)
	s.Equal(string(entities.PendingReservationStatus), string(reservations[1].Status))
	s.Empty(reservations[0].CancelledBy)
}

func (s *repositorySuite) TestCancelReservation() {
	ctx := context.Background()

	res := &entities.Reservation{
		ID:           "res-cancel-1",
		CourtID:      "court-1",
		Status:       entities.ReservedReservationStatus,
		ReservedFrom: time.Date(2024, 7, 22, 9, 0, 0, 0, time.UTC),
		ReservedTo:   time.Date(2024, 7, 22, 10, 0, 0, 0, time.UTC),
		ReservedBy:   "user-1",
		CreatedAt:    time.Date(2024, 7, 1, 9, 0, 0, 0, time.UTC),
	}

	s.seedReservations(ctx, []*entities.Reservation{res})

	cancelledBy := "admin-user"
	err := s.repo.CancelReservation(ctx, res.ID, cancelledBy)
	s.Require().NoError(err)

	resDb, err := s.repo.GetByID(ctx, res.ID)
	s.Require().NoError(err)

	s.Equal(entities.CancelledReservationStatus, resDb.Status)
	s.Equal(cancelledBy, resDb.CancelledBy)
}

func (s *repositorySuite) TestCancelReservation_NotFound() {
	ctx := context.Background()
	err := s.repo.CancelReservation(ctx, "does-not-exist-id", "someone")
	s.Require().Error(err)
	s.ErrorIs(err, entities.ErrNotFound)
}

func (s *repositorySuite) seedReservations(ctx context.Context, reservations []*entities.Reservation) {
	s.T().Helper()
	for _, res := range reservations {
		s.Require().NoError(s.repo.Create(ctx, res), "seeding reservation %s", res.ID)
	}
}
