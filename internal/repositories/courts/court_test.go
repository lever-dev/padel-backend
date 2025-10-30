package court_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/lever-dev/padel-backend/internal/repositories/court"
)

type repositorySuite struct {
	suite.Suite
	repo *court.Repository
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(repositorySuite))
}

func (s *repositorySuite) SetupTest() {
	connString := os.Getenv("POSTGRES_CONNECTION_URL")

	repo := court.NewRepository(connString)

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

func (s *repositorySuite) TestCreateAndGetCourt() {
	ctx := context.Background()
	c := &entities.Court{
		ID:             "court-create-1",
		OrganizationID: "org-1",
		Name:           "Central Court",
		SurfaceType:    "Hard",
		IsIndoor:       true,
		CreatedAt:      time.Date(2024, 7, 1, 8, 0, 0, 0, time.UTC),
	}

	err := s.repo.Create(ctx, c)
	s.Require().NoError(err)

	cDb, err := s.repo.GetByID(ctx, c.ID)
	s.Require().NoError(err)
	s.Require().Equal(c, cDb)
}

func (s *repositorySuite) TestListCourtsByOrganizationID() {
	ctx := context.Background()
	s.seedCourts(ctx, []*entities.Court{
		{
			ID:             "court-list-1",
			OrganizationID: "org-2",
			Name:           "Court A",
			SurfaceType:    "Clay",
			IsIndoor:       false,
		},
		{
			ID:             "court-list-2",
			OrganizationID: "org-2",
			Name:           "Court B",
			SurfaceType:    "Grass",
			IsIndoor:       true,
		},
		{
			ID:             "court-list-3",
			OrganizationID: "org-3",
			Name:           "Court C",
			SurfaceType:    "Hard",
			IsIndoor:       false,
		},
	})

	courts, err := s.repo.ListCourtsByOrganizationID(ctx, "org-2")
	s.Require().NoError(err)
	s.Len(courts, 2)
	s.Equal("Court A", courts[0].Name)
	s.Equal("Court B", courts[1].Name)
}

func (s *repositorySuite) TestUpdateCourt() {
	ctx := context.Background()

	c := &entities.Court{
		ID:             "court-update-1",
		OrganizationID: "org-5",
		Name:           "Court Old",
		SurfaceType:    "Clay",
		IsIndoor:       false,
	}

	s.seedCourts(ctx, []*entities.Court{c})

	c.Name = "Court Updated"
	c.IsIndoor = true
	err := s.repo.Update(ctx, c)
	s.Require().NoError(err)

	cDb, err := s.repo.GetByID(ctx, c.ID)
	s.Require().NoError(err)
	s.Equal("Court Updated", cDb.Name)
	s.True(cDb.IsIndoor)
}

func (s *repositorySuite) TestDeleteCourt() {
	ctx := context.Background()

	c := &entities.Court{
		ID:             "court-delete-1",
		OrganizationID: "org-6",
		Name:           "To Be Deleted",
		SurfaceType:    "Hard",
		IsIndoor:       false,
	}

	s.seedCourts(ctx, []*entities.Court{c})

	err := s.repo.Delete(ctx, c.ID)
	s.Require().NoError(err)

	_, err = s.repo.GetByID(ctx, c.ID)
	s.Require().Error(err)
	s.ErrorIs(err, entities.ErrNotFound)
}

func (s *repositorySuite) seedCourts(ctx context.Context, courts []*entities.Court) {
	s.T().Helper()
	for _, c := range courts {
		s.Require().NoError(s.repo.Create(ctx, c), "seeding court %s", c.ID)
	}
}
