package court

import (
	"context"
	"errors"
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
		return nil, fmt.Errorf("input organization ids dont match for court: %s: %w", courtID, entities.ErrNotFound)
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

func (s *Service) UpdateName(ctx context.Context, organizationID string, courtID string, name string) (*entities.Court, error) {
	court, err := s.courtsRepo.UpdateName(ctx, organizationID, courtID, name)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			return nil, entities.ErrNotFound
		}

		return nil, fmt.Errorf("update court name: %w", err)
	}

	return court, nil
}
