//go:generate mockgen -source=dependency.go -destination=./mocks/mocks.go -package=mocks

package reservation

import (
	"context"

	"github.com/lever-dev/padel-backend/internal/entities"
)

type ReservationsRepository interface {
	Create(ctx context.Context, reservation *entities.Reservation) error
}
