//go:generate mockgen -source=dependency.go -destination=./mocks/mocks.go -package=mocks

package reservation

import (
	"context"
	"time"

	"github.com/lever-dev/padel-backend/internal/entities"
)

type ReservationsRepository interface {
	Create(ctx context.Context, reservation *entities.Reservation) error
	ListByCourtAndTimeRange(ctx context.Context, courtID string, from, to time.Time) ([]entities.Reservation, error)
	GetByID(ctx context.Context, reservationID string) (*entities.Reservation, error)
	CancelReservation(ctx context.Context, reservationID string, cancelledBy string) error
}
