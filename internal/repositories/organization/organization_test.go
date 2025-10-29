package organization_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/lever-dev/padel-backend/internal/repositories/organization"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type repositorySuite struct {
	suite.Suite

	repo *organization.Repository
}

func TestRepositorySuit(t *testing.T) {
	suite.Run(t, new(repositorySuite))
}

func (s *repositorySuite) SetupTest() {
	connString := os.Getenv("POSTGRES_CONNECTION_URL")

	repo := organization.NewRepository(connString)

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

func (s *repositorySuite) TestCreateOrganization() {
	ctx := context.Background()

	org := &entities.Organization{
		ID:        "org-create-1",
		Name:      "org-1",
		City:      "Almaty",
		CreatedAt: time.Date(2024, 7, 1, 8, 0, 0, 0, time.UTC),
	}

	err := s.repo.Create(ctx, org)
	s.Require().NoError(err)

	orgDB, err := s.repo.GetByID(ctx, org.ID)
	s.Require().NoError(err)

	s.Require().Equal(org, orgDB)
}

func (s *repositorySuite) TestGetOrganizationByCity() {
	ctx := context.Background()
	defCity := "Astana"

	s.seedOrganizations(ctx, []*entities.Organization{
		{
			ID:   "org-list-1",
			Name: "org-1",
			City: "Almaty",
		},
		{
			ID:   "org-list-2",
			Name: "org-2",
			City: "Astana",
		},
		{
			ID:   "org-list-3",
			Name: "org-3",
			City: "Astana",
		},
		{
			ID:   "org-list-4",
			Name: "org-4",
			City: "Astana",
		},
	})

	reservations, err := s.repo.GetOrganizationsByCity(ctx, defCity)
	s.Require().NoError(err)

	s.Len(reservations, 3)
}

func (s *repositorySuite) seedOrganizations(ctx context.Context, organizations []*entities.Organization) {
	s.T().Helper()
	for _, org := range organizations {
		s.Require().NoError(s.repo.Create(ctx, org), "seeding organization %s", org.ID)
	}
}
