package court_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/lever-dev/padel-backend/internal/repositories/courts"
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
	require.NotEmpty(s.T(), connString, "POSTGRES_CONNECTION_URL must be set")

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

func (s *repositorySuite) seedCourts(ctx context.Context, courts []*entities.Court) {
	s.T().Helper()
	for _, c := range courts {
		err := s.repo.Create(ctx, c)
		s.Require().NoError(err, "failed to seed court %s", c.ID)
	}
}

func (s *repositorySuite) TestCreateAndGetCourt() {
	ctx := context.Background()

	c := &entities.Court{
		ID:             "court-create-1",
		OrganizationID: "org-1",
		Name:           "Central Court",
		CreatedAt:      time.Now().UTC(),
	}

	err := s.repo.Create(ctx, c)
	s.Require().NoError(err)

	cDb, err := s.repo.GetByID(ctx, c.ID)
	s.Require().NoError(err)
	s.Equal(c.ID, cDb.ID)
	s.Equal(c.OrganizationID, cDb.OrganizationID)
	s.Equal(c.Name, cDb.Name)
}

func (s *repositorySuite) TestListByOrganizationID() {
	ctx := context.Background()

	courts := []*entities.Court{
		{
			ID:             "court-list-1",
			OrganizationID: "org-2",
			Name:           "Court A",
		},
		{
			ID:             "court-list-2",
			OrganizationID: "org-2",
			Name:           "Court B",
		},
		{
			ID:             "court-list-3",
			OrganizationID: "org-3",
			Name:           "Court C",
		},
	}

	s.seedCourts(ctx, courts)

	list, err := s.repo.ListByOrganizationID(ctx, "org-2")
	s.Require().NoError(err)
	s.Len(list, 2)
	s.Equal("court-list-1", list[0].ID)
	s.Equal("court-list-2", list[1].ID)
	s.Equal("Court A", list[0].Name)
	s.Equal("Court B", list[1].Name)

}

func (s *repositorySuite) TestUpdateCourt() {
	ctx := context.Background()

	c := &entities.Court{
		ID:             "court-update-1",
		OrganizationID: "org-4",
		Name:           "Old Name",
	}

	s.seedCourts(ctx, []*entities.Court{c})

	c.Name = "Updated Name"
	err := s.repo.Update(ctx, c)
	s.Require().NoError(err)

	cDb, err := s.repo.GetByID(ctx, c.ID)
	s.Require().NoError(err)
	s.False(cDb.UpdatedAt.IsZero(), "UpdatedAt should be set after update")
	s.Equal("Updated Name", cDb.Name)

}

func (s *repositorySuite) TestDeleteCourt() {
	ctx := context.Background()

	c := &entities.Court{
		ID:             "court-delete-1",
		OrganizationID: "org-5",
		Name:           "To Delete",
	}

	s.seedCourts(ctx, []*entities.Court{c})

	err := s.repo.Delete(ctx, c.ID)
	s.Require().NoError(err)

	_, err = s.repo.GetByID(ctx, c.ID)
	s.Require().Error(err)
	s.ErrorIs(err, entities.ErrNotFound)
}
