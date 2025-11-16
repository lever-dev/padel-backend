package court

import (
	"context"
	"fmt"

	"github.com/lever-dev/padel-backend/internal/entities"
)

type Service struct {
	courtsRepo CourtsRepository
}

func NewService(repo CourtsRepository) *Service {
	return &Service{
		courtsRepo: repo,
	}
}

func (s *Service) Create(ctx context.Context, court *entities.Court) error {
	if err := s.courtsRepo.Create(ctx, court); err != nil {
		return fmt.Errorf("create court: %w", err)
	}
	return nil
}

func (s *Service) ListByOrganizationID(ctx context.Context, organizationID string) ([]entities.Court, error) {
	courts, err := s.courtsRepo.ListByOrganizationID(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list courts by organization id: %w", err)
	}
	return courts, nil
}

func (s *Service) GetByID(ctx context.Context, organizationID, courtID string) (*entities.Court, error) {
	court, err := s.courtsRepo.GetByID(ctx, courtID)
	if err != nil {
		return nil, fmt.Errorf("get court by id: %w", err)
	}

	if court.OrganizationID != organizationID {
		return nil, entities.ErrNotFound
	}

	return court, nil
}

func (s *Service) Update(ctx context.Context, court *entities.Court) error {
	err := s.courtsRepo.Update(ctx, court)
	if err != nil {
		return fmt.Errorf("update court: %w", err)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, courtID string) error {
	err := s.courtsRepo.Delete(ctx, courtID)
	if err != nil {
		return fmt.Errorf("delete court: %w", err)
	}
	return nil
}
