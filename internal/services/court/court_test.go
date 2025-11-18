package court_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/lever-dev/padel-backend/internal/services/court"
	"github.com/lever-dev/padel-backend/internal/services/court/mocks"
	"github.com/stretchr/testify/suite"
)

type ServiceSuite struct {
	suite.Suite
	ctrl *gomock.Controller
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
}

func (s *ServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *ServiceSuite) TestCreate() {
	ctx := context.Background()

	tests := []struct {
		name       string
		court      *entities.Court
		setupMocks func(mockRepo *mocks.MockCourtsRepository)
		wantErr    bool
	}{
		{
			name: "success",
			court: &entities.Court{
				ID:             "court-1",
				OrganizationID: "org-1",
				Name:           "Court A",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			setupMocks: func(mockRepo *mocks.MockCourtsRepository) {
				mockRepo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repository error",
			court: &entities.Court{
				ID:             "court-2",
				OrganizationID: "org-1",
				Name:           "Court B",
			},
			setupMocks: func(mockRepo *mocks.MockCourtsRepository) {
				mockRepo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(fmt.Errorf("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockRepo := mocks.NewMockCourtsRepository(s.ctrl)
			service := court.NewService(mockRepo)

			tt.setupMocks(mockRepo)

			err := service.Create(ctx, tt.court)

			if tt.wantErr {
				s.Error(err)
				s.Contains(err.Error(), "create court")
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *ServiceSuite) TestGetByID() {
	ctx := context.Background()
	organizationID := "org-1"
	courtID := "court-1"

	tests := []struct {
		name       string
		orgID      string
		courtID    string
		setupMocks func(mockRepo *mocks.MockCourtsRepository)
		wantCourt  *entities.Court
		wantErr    error
	}{
		{
			name:    "success",
			orgID:   organizationID,
			courtID: courtID,
			setupMocks: func(mockRepo *mocks.MockCourtsRepository) {
				mockRepo.EXPECT().
					GetByID(ctx, courtID).
					Return(&entities.Court{
						ID:             courtID,
						OrganizationID: organizationID,
						Name:           "Court A",
						CreatedAt:      time.Now(),
						UpdatedAt:      time.Now(),
					}, nil)
			},
			wantCourt: &entities.Court{
				ID:             courtID,
				OrganizationID: organizationID,
				Name:           "Court A",
			},
			wantErr: nil,
		},
		{
			name:    "court not found",
			orgID:   organizationID,
			courtID: "non-existent",
			setupMocks: func(mockRepo *mocks.MockCourtsRepository) {
				mockRepo.EXPECT().
					GetByID(ctx, "non-existent").
					Return(nil, entities.ErrNotFound)
			},
			wantCourt: nil,
			wantErr:   entities.ErrNotFound,
		},
		{
			name:    "court belongs to different organization",
			orgID:   "org-1",
			courtID: courtID,
			setupMocks: func(mockRepo *mocks.MockCourtsRepository) {
				mockRepo.EXPECT().
					GetByID(ctx, courtID).
					Return(&entities.Court{
						ID:             courtID,
						OrganizationID: "org-999", // другая организация
						Name:           "Court A",
					}, nil)
			},
			wantCourt: nil,
			wantErr:   entities.ErrNotFound,
		},
		{
			name:    "repository error",
			orgID:   organizationID,
			courtID: courtID,
			setupMocks: func(mockRepo *mocks.MockCourtsRepository) {
				mockRepo.EXPECT().
					GetByID(ctx, courtID).
					Return(nil, fmt.Errorf("db error"))
			},
			wantCourt: nil,
			wantErr:   fmt.Errorf("db error"),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockRepo := mocks.NewMockCourtsRepository(s.ctrl)
			service := court.NewService(mockRepo)

			tt.setupMocks(mockRepo)

			result, err := service.GetByID(ctx, tt.orgID, tt.courtID)

			if tt.wantErr != nil {
				s.Error(err)
				if errors.Is(tt.wantErr, entities.ErrNotFound) {
					s.ErrorIs(err, entities.ErrNotFound)
				}
			} else {
				s.NoError(err)
				s.NotNil(result)
				s.Equal(tt.wantCourt.ID, result.ID)
				s.Equal(tt.wantCourt.OrganizationID, result.OrganizationID)
				s.Equal(tt.wantCourt.Name, result.Name)
			}
		})
	}
}

func (s *ServiceSuite) TestListByOrganizationID() {
	ctx := context.Background()
	organizationID := "org-1"

	tests := []struct {
		name       string
		orgID      string
		setupMocks func(mockRepo *mocks.MockCourtsRepository)
		wantCourts []entities.Court
		wantErr    bool
	}{
		{
			name:  "success with multiple courts",
			orgID: organizationID,
			setupMocks: func(mockRepo *mocks.MockCourtsRepository) {
				mockRepo.EXPECT().
					ListByOrganizationID(ctx, organizationID).
					Return([]entities.Court{
						{
							ID:             "court-1",
							OrganizationID: organizationID,
							Name:           "Court A",
						},
						{
							ID:             "court-2",
							OrganizationID: organizationID,
							Name:           "Court B",
						},
					}, nil)
			},
			wantCourts: []entities.Court{
				{
					ID:             "court-1",
					OrganizationID: organizationID,
					Name:           "Court A",
				},
				{
					ID:             "court-2",
					OrganizationID: organizationID,
					Name:           "Court B",
				},
			},
			wantErr: false,
		},
		{
			name:  "success with empty list",
			orgID: organizationID,
			setupMocks: func(mockRepo *mocks.MockCourtsRepository) {
				mockRepo.EXPECT().
					ListByOrganizationID(ctx, organizationID).
					Return([]entities.Court{}, nil)
			},
			wantCourts: []entities.Court{},
			wantErr:    false,
		},
		{
			name:  "repository error",
			orgID: organizationID,
			setupMocks: func(mockRepo *mocks.MockCourtsRepository) {
				mockRepo.EXPECT().
					ListByOrganizationID(ctx, organizationID).
					Return(nil, fmt.Errorf("db error"))
			},
			wantCourts: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockRepo := mocks.NewMockCourtsRepository(s.ctrl)
			service := court.NewService(mockRepo)

			tt.setupMocks(mockRepo)

			result, err := service.ListByOrganizationID(ctx, tt.orgID)

			if tt.wantErr {
				s.Error(err)
				s.Contains(err.Error(), "list courts by organization id")
			} else {
				s.NoError(err)
				s.Equal(len(tt.wantCourts), len(result))

				for i, expectedCourt := range tt.wantCourts {
					s.Equal(expectedCourt.ID, result[i].ID)
					s.Equal(expectedCourt.OrganizationID, result[i].OrganizationID)
					s.Equal(expectedCourt.Name, result[i].Name)
				}
			}
		})
	}
}

func (s *ServiceSuite) TestUpdate() {
	ctx := context.Background()

	tests := []struct {
		name       string
		court      *entities.Court
		setupMocks func(mockRepo *mocks.MockCourtsRepository)
		wantErr    error
	}{
		{
			name: "success",
			court: &entities.Court{
				ID:             "court-1",
				OrganizationID: "org-1",
				Name:           "Updated Court Name",
				UpdatedAt:      time.Now(),
			},
			setupMocks: func(mockRepo *mocks.MockCourtsRepository) {
				mockRepo.EXPECT().
					Update(ctx, gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "court not found",
			court: &entities.Court{
				ID:             "non-existent",
				OrganizationID: "org-1",
				Name:           "Court Name",
			},
			setupMocks: func(mockRepo *mocks.MockCourtsRepository) {
				mockRepo.EXPECT().
					Update(ctx, gomock.Any()).
					Return(entities.ErrNotFound)
			},
			wantErr: entities.ErrNotFound,
		},
		{
			name: "repository error",
			court: &entities.Court{
				ID:             "court-1",
				OrganizationID: "org-1",
				Name:           "Court Name",
			},
			setupMocks: func(mockRepo *mocks.MockCourtsRepository) {
				mockRepo.EXPECT().
					Update(ctx, gomock.Any()).
					Return(fmt.Errorf("db error"))
			},
			wantErr: fmt.Errorf("db error"),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockRepo := mocks.NewMockCourtsRepository(s.ctrl)
			service := court.NewService(mockRepo)

			tt.setupMocks(mockRepo)

			err := service.Update(ctx, tt.court)

			if tt.wantErr != nil {
				s.Error(err)
				if errors.Is(tt.wantErr, entities.ErrNotFound) {
					s.ErrorIs(err, entities.ErrNotFound)
				} else {
					s.Contains(err.Error(), "update court")
				}
			} else {
				s.NoError(err)
			}
		})
	}
}
func (s *ServiceSuite) TestUpdateName() {
	ctx := context.Background()
	orgID := "org-1"
	courtID := "court-1"
	newName := "New Court Name"

	tests := []struct {
		name       string
		orgID      string
		courtID    string
		newName    string
		setupMocks func(mockRepo *mocks.MockCourtsRepository)
		wantCourt  *entities.Court
		wantErr    error
	}{
		{
			name:    "success",
			orgID:   orgID,
			courtID: courtID,
			newName: newName,
			setupMocks: func(mockRepo *mocks.MockCourtsRepository) {
				existing := &entities.Court{
					ID:             courtID,
					OrganizationID: orgID,
					Name:           "Old Name",
					CreatedAt:      time.Now(),
				}

				gomock.InOrder(
					mockRepo.EXPECT().
						GetByID(ctx, courtID).
						Return(existing, nil),

					mockRepo.EXPECT().
						UpdateName(ctx, gomock.AssignableToTypeOf(&entities.Court{})).
						Return(nil),
				)
			},
			wantCourt: &entities.Court{
				ID:             courtID,
				OrganizationID: orgID,
				Name:           newName,
			},
			wantErr: nil,
		},
		{
			name:    "court not found (GetByID)",
			orgID:   orgID,
			courtID: "non-existent",
			newName: newName,
			setupMocks: func(mockRepo *mocks.MockCourtsRepository) {
				mockRepo.EXPECT().
					GetByID(ctx, "non-existent").
					Return(nil, entities.ErrNotFound)
			},
			wantCourt: nil,
			wantErr:   entities.ErrNotFound,
		},
		{
			name:    "repository error on UpdateName",
			orgID:   orgID,
			courtID: courtID,
			newName: newName,
			setupMocks: func(mockRepo *mocks.MockCourtsRepository) {
				existing := &entities.Court{
					ID:             courtID,
					OrganizationID: orgID,
					Name:           "Old Name",
					CreatedAt:      time.Now(),
				}

				gomock.InOrder(
					mockRepo.EXPECT().
						GetByID(ctx, courtID).
						Return(existing, nil),

					mockRepo.EXPECT().
						UpdateName(ctx, gomock.AssignableToTypeOf(&entities.Court{})).
						Return(fmt.Errorf("db error")),
				)
			},
			wantCourt: nil,
			wantErr:   fmt.Errorf("db error"),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockRepo := mocks.NewMockCourtsRepository(s.ctrl)
			service := court.NewService(mockRepo)

			tt.setupMocks(mockRepo)

			result, err := service.UpdateName(ctx, tt.orgID, tt.courtID, tt.newName)

			if tt.wantErr != nil {
				s.Error(err)

				if errors.Is(tt.wantErr, entities.ErrNotFound) {
					s.ErrorIs(err, entities.ErrNotFound)
				} else {
					if tt.name == "repository error on UpdateName" {
						s.Contains(err.Error(), "update court name")
					} else {
						s.Contains(err.Error(), "get court by id")
					}
				}

				s.Nil(result)
			} else {
				s.NoError(err)
				s.NotNil(result)
				s.Equal(tt.wantCourt.ID, result.ID)
				s.Equal(tt.wantCourt.OrganizationID, result.OrganizationID)
				s.Equal(tt.wantCourt.Name, result.Name)
			}
		})
	}
}
