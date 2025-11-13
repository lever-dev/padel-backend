//go:generate mockgen -source=dependency.go -destination=./mocks/mocks.go -package=mocks

package auth

import (
	"context"

	"github.com/lever-dev/padel-backend/internal/entities"
)

type UsersRepository interface {
	GetByNickname(ctx context.Context, nickname string) (entities.User, error)
	Create(ctx context.Context, user *entities.User) error
}
