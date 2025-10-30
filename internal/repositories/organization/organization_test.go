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
		UpdatedAt: time.Date(2024, 7, 1, 8, 0, 0, 0, time.UTC),
	}

	err := s.repo.Create(ctx, org)
	s.Require().NoError(err)

	orgDB, err := s.repo.GetByID(ctx, org.ID)
	s.Require().NoError(err)

	s.Require().Equal(org, orgDB)
}

func (s *repositorySuite) TestGetOrganizationByID() {
	ctx := context.Background()

	org := &entities.Organization{
		ID:        "org-get-1",
		Name:      "org-1",
		City:      "Aktobe",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	s.seedOrganizations(ctx, []*entities.Organization{org})

	result, err := s.repo.GetByID(ctx, org.ID)
	s.Require().NoError(err)
	s.Require().NotNil(result)
	s.Equal(org.ID, result.ID)
	s.Equal(org.Name, result.Name)
	s.Equal(org.City, result.City)
}

func (s *repositorySuite) TestGetOrganizationByID_NotFound() {
	ctx := context.Background()

	_, err := s.repo.GetByID(ctx, "non-existent-id")
	s.Require().Error(err)
	s.ErrorIs(err, entities.ErrNotFound)
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

	results, err := s.repo.GetOrganizationsByCity(ctx, defCity)
	s.Require().NoError(err)

	count := 0
	for _, org := range results {
		if org.City == "Astana" {
			count++
		}
		s.Equal("Astana", org.City, "All results should be from Almaty")
	}

	s.GreaterOrEqual(count, 3, "Should find at least 3 Astana organizations")
}

func (s *repositorySuite) TestGetOrganizationByCity_Empty() {
	ctx := context.Background()

	results, err := s.repo.GetOrganizationsByCity(ctx, "NonExistentCity")
	s.Require().NoError(err)
	s.Require().Empty(results)
}

func (s *repositorySuite) TestUpdateOrganization() {
	ctx := context.Background()

	org := &entities.Organization{
		ID:        "org-update-1",
		Name:      "Old Name",
		City:      "Almaty",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	s.seedOrganizations(ctx, []*entities.Organization{org})

	org.Name = "New Name"
	org.City = "Astana"

	err := s.repo.Update(ctx, org)
	s.Require().NoError(err)

	updated, err := s.repo.GetByID(ctx, org.ID)
	s.Require().NoError(err)
	s.Equal("New Name", updated.Name)
	s.Equal("Astana", updated.City)
	s.NotEqual(org.UpdatedAt, updated.UpdatedAt)
}

func (s *repositorySuite) TestUpdateOrganization_NotFound() {
	ctx := context.Background()

	org := &entities.Organization{
		ID:   "non-existent",
		Name: "Test",
		City: "Test",
	}

	err := s.repo.Update(ctx, org)
	s.Require().Error(err)
	s.ErrorIs(err, entities.ErrNotFound)
}

func (s *repositorySuite) TestDeleteOrganization() {
	ctx := context.Background()

	org := &entities.Organization{
		ID:        "org-delete-1",
		Name:      "To Be Deleted",
		City:      "Almaty",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	s.seedOrganizations(ctx, []*entities.Organization{org})

	err := s.repo.Delete(ctx, org.ID)
	s.Require().NoError(err)

	_, err = s.repo.GetByID(ctx, org.ID)
	s.Require().Error(err)
	s.ErrorIs(err, entities.ErrNotFound)
}

func (s *repositorySuite) seedOrganizations(ctx context.Context, organizations []*entities.Organization) {
	s.T().Helper()
	for _, org := range organizations {
		s.Require().NoError(s.repo.Create(ctx, org), "seeding organization %s", org.ID)
	}
}
