package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lever-dev/padel-backend/internal/entities"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	usersRepo UsersRepository
	jwtSecret []byte
}

func NewService(repo UsersRepository) *Service {
	return &Service{
		usersRepo: repo,
		jwtSecret: []byte("some-jwt-key"), // TODO: use from env
	}
}

func (s *Service) LoginViaPassword(ctx context.Context, nickname, password string) (string, error) {
	user, err := s.usersRepo.GetByNickname(ctx, nickname)
	if err != nil {
		return "", fmt.Errorf("get by nickname: %w", err)
	}

	if err := s.comparePasswords(user, password); err != nil {
		return "", fmt.Errorf("%w: %w", entities.ErrInvalidCredentials, err)
	}

	tok, err := s.issueToken(user)
	if err != nil {
		return "", fmt.Errorf("issue token: %w", err)
	}

	return tok, nil
}

func (s *Service) RegisterUser(ctx context.Context, user *entities.User, password string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user.HashedPassword = string(hashed)

	if err := s.usersRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (s *Service) VerifyToken(tokenStr string) error {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, entities.ErrInvalidToken
		}

		return s.jwtSecret, nil
	})
	if err != nil {
		return fmt.Errorf("jwt parse: %w", err)
	}

	if !token.Valid {
		return fmt.Errorf("invalid token ")
	}

	return nil
}

func (s *Service) comparePasswords(user entities.User, providedPass string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(providedPass))
}

func (s *Service) issueToken(user entities.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"nick": user.Nickname,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
