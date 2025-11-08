package organization

import (
	"context"
)

type Service struct {
	organizationsRepo OrganizationsRepository
}

func NewService(repo OrganizationsRepository) *Service {
	return &Service{
		organizationsRepo: repo,
	}
}

func (s *Service) CreateOrganization(ctx context.Context /* TODO: add params */) error {
	return nil
}
