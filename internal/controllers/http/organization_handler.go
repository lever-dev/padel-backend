package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/lever-dev/padel-backend/pkg/httputil"
	"github.com/rs/zerolog/log"
)

type OrganizationService interface {
	CreateOrganization(ctx context.Context, organization *entities.Organization) error
	GetOrganizationsByCity(ctx context.Context, city string) ([]entities.Organization, error)
	GetOrganization(ctx context.Context, organizationID string) (*entities.Organization, error)
	UpdateOrganization(ctx context.Context, organization *entities.Organization) error
	DeleteOrganization(ctx context.Context, organizationID string) error
}

type OrganizationHandler struct {
	orgService OrganizationService
}

func NewOrganizationHandler(service OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{
		orgService: service,
	}
}

type CreateOrganizationRequest struct {
	// Name is a organization name
	// example: Padel Club #1
	Name string `json:"name" example:"Padel club"`

	// City is the city in which the organization itself is located
	// example: Almaty
	City string `json:"city" example:"Astana"`
}

type CreateOrganizationResponse struct{}

// CreateOrganization godoc
// @Summary Create a new organization
// @Description Create a new organization in the city
// @Tags organizations
// @Security BearerAuth
// @Accept json
// @Param organization body CreateOrganizationRequest true "Organization payload"
// @Success 200
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500
// @Router /v1/organizations [post]
func (o *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("failed to decode create organization request")

		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{Message: "invalid json"})
		return
	}

	if req.Name == "" || req.City == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "name or city are required",
		})
		return
	}

	org := entities.NewOrganization(req.Name, req.City)

	if err := o.orgService.CreateOrganization(r.Context(), org); err != nil {
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			httputil.JSON(w, http.StatusConflict, ErrorResponse{
				Message: "organization with this name already exists in this city",
			})

			log.Error().Err(err).Str("name", org.Name).Str("city", org.City).Msg("organization already exists")
			return
		}

		log.Error().Err(err).Msg("failed to create organization")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Info().
		Str("organization id", org.ID).
		Str("name", org.Name).
		Str("city", org.City).
		Msg("organization was created")
}

type OrganizationResponse struct {
	// ID is the unique identifier of the created organization
	// example: org-123
	ID string `json:"id" example:"org-1"`

	// Name is the organization name
	// example: Padel Club Almaty
	Name string `json:"name" example:"Padel Club Almaty"`

	// City is the city where the organization is located
	// example: Almaty
	City string `json:"city" example:"Almaty"`

	// CreatedAt is the timestamp when the organization was created
	// example: 2025-11-01T10:00:00Z
	CreatedAt time.Time `json:"createdAt" example:"2025-11-01T10:00:00Z"`

	// UpdatedAt is the timestamp when the organization was last updated (RFC3339 format)
	// example: 2025-11-01T10:00:00Z
	UpdatedAt time.Time `json:"updatedAt" example:"2025-11-01T10:00:00Z"`
}

// GetOrganization godoc
// @Summary Get an organization
// @Description Retrieves an organization by ID
// @Tags organizations
// @Security BearerAuth
// @Produce json
// @Param orgID path string true "Organization ID"
// @Success 200 {object} OrganizationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404
// @Failure 500
// @Router /v1/organizations/{orgID} [get]
func (h *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")
	if orgID == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{Message: "orgID is required"})
		return
	}

	org, err := h.orgService.GetOrganization(r.Context(), orgID)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)

			log.Info().Str("orgID", orgID).Msg("organization not found") // Log ERR?
			return
		}
		log.Error().Err(err).Str("orgID", orgID).Msg("failed to get organization")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := OrganizationResponse{
		ID:        org.ID,
		Name:      org.Name,
		City:      org.City,
		CreatedAt: org.CreatedAt,
		UpdatedAt: org.UpdatedAt,
	}

	httputil.JSON(w, http.StatusOK, resp)
}

type ListOrganizationsResponse struct {
	// Organizations is the list of organizations
	Organizations []OrganizationResponse `json:"organizations"`
}

// GetOrganizationsByCity godoc
// @Summary List organizations by city
// @Description Returns all organizations in a specific city
// @Tags organizations
// @Security BearerAuth
// @Produce json
// @Param city query string true "City name"
// @Success 200 {object} ListOrganizationsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500
// @Router /v1/organizations [get]
func (h *OrganizationHandler) GetOrganizationsByCity(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{Message: "city query parameter is required"})
		return
	}

	orgs, err := h.orgService.GetOrganizationsByCity(r.Context(), city)
	if err != nil {
		log.Error().Err(err).Str("city", city).Msg("failed to get organizations by city")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := ListOrganizationsResponse{
		Organizations: make([]OrganizationResponse, 0, len(orgs)),
	}

	for _, org := range orgs {
		resp.Organizations = append(resp.Organizations, OrganizationResponse{
			ID:        org.ID,
			Name:      org.Name,
			City:      org.City,
			CreatedAt: org.CreatedAt,
			UpdatedAt: org.UpdatedAt,
		})
	}

	httputil.JSON(w, http.StatusOK, resp)
	log.Info().Str("city", city).Int("count", len(orgs)).Msg("listed organizations by city")
}

type UpdateOrganizationRequest struct {
	// Name is the updated organization name
	// example: Updated Padel Club
	Name string `json:"name" example:"Updated Padel Club"`

	// City is the updated city where the organization is located
	// example: Astana
	City string `json:"city" example:"Astana"`
}

// UpdateOrganization godoc
// @Summary Update an organization
// @Description Updates an existing organization
// @Tags organizations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param orgID path string true "Organization ID"
// @Param organization body UpdateOrganizationRequest true "Organization payload"
// @Success 200
// @Failure 400 {object} ErrorResponse
// @Failure 404
// @Failure 500
// @Router /v1/organizations/{orgID} [put]
func (h *OrganizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")
	if orgID == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{Message: "orgID is required"})
		return
	}

	var req UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{Message: "invalid json"})
		return
	}

	if req.Name == "" || req.City == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{Message: "name and city are required"})
		return
	}

	org := &entities.Organization{
		ID:        orgID,
		Name:      req.Name,
		City:      req.City,
		UpdatedAt: time.Now().UTC(), // Maybe delete this row
	}

	if err := h.orgService.UpdateOrganization(r.Context(), org); err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			log.Info().Str("orgID", orgID).Msg("organization not found for update") // or ERR log
			return
		}
		log.Error().Err(err).Str("orgID", orgID).Msg("failed to update organization")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Info().Str("organization_id", org.ID).Msg("organization updated successfully")
}

// DeleteOrganization godoc
// @Summary Delete an organization
// @Description Deletes an organization by ID
// @Tags organizations
// @Security BearerAuth
// @Param orgID path string true "Organization ID"
// @Success 200
// @Failure 400 {object} ErrorResponse
// @Failure 404
// @Failure 500
// @Router /v1/organizations/{orgID} [delete]
func (h *OrganizationHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")
	if orgID == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{Message: "orgID is required"})
		return
	}

	if err := h.orgService.DeleteOrganization(r.Context(), orgID); err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			log.Info().Str("orgID", orgID).Msg("organization not found for deletion") // or err Log
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Str("orgID", orgID).Msg("failed to delete organization")
		return
	}

	log.Info().Str("organization_id", orgID).Msg("organization deleted successfully")
}
