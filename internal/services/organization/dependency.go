//go:generate mockgen -source=dependency.go -destination=./mocks/mocks.go -package=mocks

package organization

import (
	"context"

	"github.com/lever-dev/padel-backend/internal/entities"
)

type OrganizationsRepository interface {
	Create(ctx context.Context, organization *entities.Organization) error
	GetByID(ctx context.Context, organizationID string) (*entities.Organization, error)
	GetOrganizationsByCity(ctx context.Context, city string) ([]entities.Organization, error)
	Update(ctx context.Context, org *entities.Organization) error
	Delete(ctx context.Context, id string) error
}
