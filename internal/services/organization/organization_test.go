package organization_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/lever-dev/padel-backend/internal/services/organization"
	"github.com/lever-dev/padel-backend/internal/services/organization/mocks"
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

func (s *ServiceSuite) TestCreateOrganization() {
	tests := []struct {
		name         string
		organization *entities.Organization
		setupMocks   func(mockRepo *mocks.MockOrganizationsRepository, org *entities.Organization)
		wantErr      bool
	}{
		{
			name: "success",
			organization: &entities.Organization{
				ID:        "org-1",
				Name:      "Padel Astana",
				City:      "Astana",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			setupMocks: func(mockRepo *mocks.MockOrganizationsRepository, org *entities.Organization) {
				mockRepo.EXPECT().Create(gomock.Any(), org).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "db error",
			organization: &entities.Organization{
				ID:        "org-1",
				Name:      "Padel Astana",
				City:      "Astana",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			setupMocks: func(mockRepo *mocks.MockOrganizationsRepository, org *entities.Organization) {
				mockRepo.EXPECT().Create(gomock.Any(), org).Return(fmt.Errorf("db error"))
			},
			wantErr: true,
		},
		{
			name: "data conflict",
			organization: &entities.Organization{
				ID:        "org-1",
				Name:      "Padel Astana",
				City:      "Astana",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			setupMocks: func(mockRepo *mocks.MockOrganizationsRepository, org *entities.Organization) {
				mockRepo.EXPECT().Create(gomock.Any(), org).Return(fmt.Errorf("create organization: duplicate key"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			mockRepo := mocks.NewMockOrganizationsRepository(s.ctrl)
			service := organization.NewService(mockRepo)

			tt.setupMocks(mockRepo, tt.organization)

			err := service.CreateOrganization(ctx, tt.organization)
			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *ServiceSuite) TestGetOrganization() {
	tests := []struct {
		name       string
		orgID      string
		setupMocks func(mockRepo *mocks.MockOrganizationsRepository, orgID string)
		wantErr    bool
		wantOrg    *entities.Organization
	}{
		{
			name:  "success",
			orgID: "org-1",
			setupMocks: func(mockRepo *mocks.MockOrganizationsRepository, orgID string) {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), orgID).
					Return(&entities.Organization{
						ID:        "org-1",
						Name:      "Padel Astana",
						City:      "Astana",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)
			},
			wantErr: false,
			wantOrg: &entities.Organization{
				ID:        "org-1",
				Name:      "Padel Astana",
				City:      "Astana",
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:  "not found",
			orgID: "non-existent",
			setupMocks: func(mockRepo *mocks.MockOrganizationsRepository, orgID string) {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), orgID).
					Return(nil, entities.ErrNotFound)
			},
			wantErr: true,
			wantOrg: nil,
		},
		{
			name:  "db error",
			orgID: "org-1",
			setupMocks: func(mockRepo *mocks.MockOrganizationsRepository, orgID string) {
				mockRepo.EXPECT().
					GetByID(gomock.Any(), orgID).
					Return(nil, fmt.Errorf("db error"))
			},
			wantErr: true,
			wantOrg: nil,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			mockRepo := mocks.NewMockOrganizationsRepository(s.ctrl)
			service := organization.NewService(mockRepo)

			tt.setupMocks(mockRepo, tt.orgID)

			org, err := service.GetOrganization(ctx, tt.orgID)

			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tt.wantOrg, org)
			}
		})
	}
}

func (s *ServiceSuite) TestGetOrganizationsByCity() {
	tests := []struct {
		name       string
		city       string
		setupMocks func(mockRepo *mocks.MockOrganizationsRepository, city string)
		wantErr    bool
		wantCount  int
	}{
		{
			name: "success",
			city: "Astana",
			setupMocks: func(mockRepo *mocks.MockOrganizationsRepository, city string) {
				mockRepo.EXPECT().
					GetOrganizationsByCity(gomock.Any(), city).
					Return([]entities.Organization{
						{
							ID:   "org-1",
							Name: "Padel Club 1",
							City: "Astana",
						},
						{
							ID:   "org-2",
							Name: "Padel Club 2",
							City: "Astana",
						},
						{
							ID:   "org-3",
							Name: "Padel Club 3",
							City: "Astana",
						},
					}, nil)
			},
			wantErr:   false,
			wantCount: 3,
		},
		{
			name: "success with empty list",
			city: "Shymkent",
			setupMocks: func(mockRepo *mocks.MockOrganizationsRepository, city string) {
				mockRepo.EXPECT().
					GetOrganizationsByCity(gomock.Any(), city).
					Return([]entities.Organization{}, nil)
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "db error",
			city: "Almaty",
			setupMocks: func(mockRepo *mocks.MockOrganizationsRepository, city string) {
				mockRepo.EXPECT().
					GetOrganizationsByCity(gomock.Any(), city).
					Return(nil, fmt.Errorf("db error"))
			},
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			mockRepo := mocks.NewMockOrganizationsRepository(s.ctrl)
			service := organization.NewService(mockRepo)

			tt.setupMocks(mockRepo, tt.city)

			orgs, err := service.GetOrganizationsByCity(ctx, tt.city)

			if tt.wantErr {
				s.Error(err)
				s.Nil(orgs)
			} else {
				s.NoError(err)
				s.Len(orgs, tt.wantCount)
			}
		})
	}
}

func (s *ServiceSuite) TestUpdateOrganization() {
	tests := []struct {
		name       string
		org        *entities.Organization
		setupMocks func(mockRepo *mocks.MockOrganizationsRepository, org *entities.Organization)
		wantErr    bool
	}{
		{
			name: "success",
			org: &entities.Organization{
				ID:        "org-1",
				Name:      "Updated Name",
				City:      "Updated City",
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			setupMocks: func(mockRepo *mocks.MockOrganizationsRepository, org *entities.Organization) {
				mockRepo.EXPECT().
					Update(gomock.Any(), org).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			org: &entities.Organization{
				ID:   "non-existent",
				Name: "Test",
				City: "Test",
			},
			setupMocks: func(mockRepo *mocks.MockOrganizationsRepository, org *entities.Organization) {
				mockRepo.EXPECT().
					Update(gomock.Any(), org).
					Return(entities.ErrNotFound)
			},
			wantErr: true,
		},
		{
			name: "db error",
			org: &entities.Organization{
				ID:   "org-error",
				Name: "Test",
				City: "Test",
			},
			setupMocks: func(mockRepo *mocks.MockOrganizationsRepository, org *entities.Organization) {
				mockRepo.EXPECT().
					Update(gomock.Any(), org).
					Return(fmt.Errorf("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			mockRepo := mocks.NewMockOrganizationsRepository(s.ctrl)
			service := organization.NewService(mockRepo)

			tt.setupMocks(mockRepo, tt.org)

			err := service.UpdateOrganization(ctx, tt.org)

			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}
