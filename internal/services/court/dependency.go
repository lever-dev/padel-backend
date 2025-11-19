//go:generate mockgen -source=dependency.go -destination=./mocks/mocks.go -package=mocks

package court

import (
	"context"

	"github.com/lever-dev/padel-backend/internal/entities"
)

type CourtsRepository interface {
	Create(ctx context.Context, court *entities.Court) error
	ListByOrganizationID(ctx context.Context, organizationID string) ([]entities.Court, error)
	GetByID(ctx context.Context, courtID string) (*entities.Court, error)
	Update(ctx context.Context, court *entities.Court) error
	UpdateName(ctx context.Context, court *entities.Court) error
}
