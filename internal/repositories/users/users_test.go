package users_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/lever-dev/padel-backend/internal/repositories/users"
)

type repositorySuite struct {
	suite.Suite

	repo *users.Repository
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(repositorySuite))
}

func (s *repositorySuite) SetupTest() {
	connString := os.Getenv("POSTGRES_CONNECTION_URL")
	repo := users.NewRepository(connString)

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

func (s *repositorySuite) TestCreateUser() {
	ctx := context.Background()
	createdAt := time.Date(2024, 7, 1, 10, 0, 0, 0, time.UTC)

	user := &entities.User{
		ID:             "user-create-1",
		Nickname:       "alice",
		HashedPassword: "hashed",
		PhoneNumber:    "+77010000001",
		FirstName:      "Alice",
		LastName:       "Baker",
		CreatedAt:      createdAt,
	}

	err := s.repo.Create(ctx, user)
	s.Require().NoError(err)

	userDB, err := s.repo.GetByID(ctx, user.ID)
	s.Require().NoError(err)

	s.Equal(*user, *userDB)
}

func (s *repositorySuite) TestGetByID_NotFound() {
	ctx := context.Background()

	_, err := s.repo.GetByID(ctx, "missing-user")
	s.Require().Error(err)
	s.ErrorIs(err, entities.ErrNotFound)
}

func (s *repositorySuite) TestGetByPhoneNumber() {
	ctx := context.Background()
	lastLogin := time.Date(2024, 7, 2, 11, 0, 0, 0, time.UTC)

	user := &entities.User{
		ID:             "user-phone-1",
		Nickname:       "bob",
		HashedPassword: "hashed",
		PhoneNumber:    "+77010000002",
		FirstName:      "Bob",
		LastName:       "Carter",
		LastLoginAt:    &lastLogin,
	}

	s.seedUsers(ctx, []*entities.User{user})

	result, err := s.repo.GetByPhoneNumber(ctx, user.PhoneNumber)
	s.Require().NoError(err)
	s.Equal(user.ID, result.ID)
	s.Require().NotNil(result.LastLoginAt)
	s.Equal(lastLogin.UTC(), result.LastLoginAt.UTC())
}

func (s *repositorySuite) TestGetByNickname() {
	ctx := context.Background()

	user := &entities.User{
		ID:             "user-nick-1",
		Nickname:       "carol",
		HashedPassword: "hashed",
		PhoneNumber:    "+77010000005",
		FirstName:      "Carol",
		LastName:       "Danvers",
	}

	s.seedUsers(ctx, []*entities.User{user})

	found, err := s.repo.GetByNickname(ctx, user.Nickname)
	s.Require().NoError(err)
	s.Equal(user.ID, found.ID)
	s.Equal(user.Nickname, found.Nickname)
	s.Equal(user.HashedPassword, found.HashedPassword)
}

func (s *repositorySuite) TestUpdateLastLogin() {
	ctx := context.Background()
	user := &entities.User{
		ID:             "user-login-1",
		Nickname:       "dave",
		HashedPassword: "hashed",
		PhoneNumber:    "+77010000003",
		FirstName:      "Carol",
		LastName:       "Davis",
	}

	s.seedUsers(ctx, []*entities.User{user})

	newLogin := time.Date(2024, 7, 3, 12, 0, 0, 0, time.UTC)

	err := s.repo.UpdateLastLogin(ctx, user.ID, newLogin)
	s.Require().NoError(err)

	updated, err := s.repo.GetByID(ctx, user.ID)
	s.Require().NoError(err)
	s.Require().NotNil(updated.LastLoginAt)
	s.Equal(newLogin.UTC(), updated.LastLoginAt.UTC())
}

func (s *repositorySuite) TestUpdateLastLogin_NotFound() {
	ctx := context.Background()

	err := s.repo.UpdateLastLogin(ctx, "unknown", time.Now())
	s.Require().Error(err)
	s.ErrorIs(err, entities.ErrNotFound)
}

func (s *repositorySuite) seedUsers(ctx context.Context, usersToSeed []*entities.User) {
	s.T().Helper()
	for _, u := range usersToSeed {
		if u.CreatedAt.IsZero() {
			u.CreatedAt = time.Now().UTC()
		}
		if u.Nickname == "" {
			u.Nickname = fmt.Sprintf("nick-%s", u.ID)
		}
		if u.HashedPassword == "" {
			u.HashedPassword = "hashed-password"
		}
		s.Require().NoError(s.repo.Create(ctx, u), "seeding user %s", u.ID)
	}
}
