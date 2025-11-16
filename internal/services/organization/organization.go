package organization

import (
	"context"
	"fmt"

	"github.com/lever-dev/padel-backend/internal/entities"
)

type Service struct {
	organizationsRepo OrganizationsRepository
}

func NewService(repo OrganizationsRepository) *Service {
	return &Service{
		organizationsRepo: repo,
	}
}

func (s *Service) CreateOrganization(ctx context.Context, organization *entities.Organization) error {
	if err := s.organizationsRepo.Create(ctx, organization); err != nil {
		return fmt.Errorf("create organization: %w", err)
	}
	return nil
}

func (s *Service) GetOrganization(ctx context.Context, id string) (*entities.Organization, error) {
	org, err := s.organizationsRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get organization: %w", err)
	}
	return org, nil
}

func (s *Service) GetOrganizationsByCity(ctx context.Context, city string) ([]entities.Organization, error) {
	orgs, err := s.organizationsRepo.GetOrganizationsByCity(ctx, city)
	if err != nil {
		return nil, fmt.Errorf("get organizations by city: %w", err)
	}
	return orgs, nil
}

func (s *Service) UpdateOrganization(ctx context.Context, org *entities.Organization) error {
	if err := s.organizationsRepo.Update(ctx, org); err != nil {
		return fmt.Errorf("update organization: %w", err)
	}
	return nil
}

func (s *Service) DeleteOrganization(ctx context.Context, id string) error {
	if err := s.organizationsRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete organization: %w", err)
	}
	return nil
}
