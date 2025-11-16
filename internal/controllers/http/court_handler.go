package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/lever-dev/padel-backend/pkg/httputil"
	"github.com/rs/zerolog/log"
)

type CourtService interface {
	Create(ctx context.Context, court *entities.Court) error
	GetByID(ctx context.Context, organizationID, courtID string) (*entities.Court, error)
	ListByOrganizationID(ctx context.Context, organizationID string) ([]entities.Court, error)
	UpdateName(ctx context.Context, organizationID, courtID, name string) (*entities.Court, error)
}

type CourtHandler struct {
	courtService CourtService
}

func NewCourtHandler(service CourtService) *CourtHandler {
	return &CourtHandler{
		courtService: service,
	}
}

// swagger:model CreateCourtRequest
type CreateCourtRequest struct {
	Name string `json:"name" example:"Court 1"`
}

// swagger:model CreateCourtResponse
type CreateCourtResponse struct {
	ID             string `json:"id"             example:"court-123"`
	OrganizationID string `json:"organizationId" example:"org-456"`
	Name           string `json:"name"           example:"Court 1"`
}

// swagger:model CourtResponse
type CourtResponse struct {
	ID             string     `json:"id"             example:"court-123"`
	OrganizationID string     `json:"organizationId" example:"org-456"`
	Name           string     `json:"name"           example:"Court 1"`
	CreatedAt      time.Time  `json:"createdAt"      example:"2025-11-01T10:00:00Z" format:"date-time"`
	UpdatedAt      *time.Time `json:"updatedAt,omitempty" example:"2025-11-01T10:00:00Z" format:"date-time"`
}

// swagger:model ListCourtsResponse
type ListCourtsResponse struct {
	Courts []CourtResponse `json:"courts"`
}

// swagger:model UpdateCourtRequest
type UpdateCourtRequest struct {
	Name string `json:"name" example:"Updated Court Name"`
}

// CreateCourt godoc
// @Summary Create a new court
// @Description Creates a new court for the specified organization
// @Tags courts
// @Security BearerAuth
// @Param orgID path string true "Organization ID"
// @Accept json
// @Produce json
// @Param court body CreateCourtRequest true "Court creation payload"
// @Success 201 {object} CreateCourtResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500
// @Router /v1/organizations/{orgID}/courts [post]
func (h *CourtHandler) CreateCourt(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")

	if orgID == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "orgID is required",
		})
		return
	}

	var req CreateCourtRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Str("orgID", orgID).Msg("failed to decode create court request")
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{Message: "invalid json"})
		return
	}

	if req.Name == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "name is required",
		})
		return
	}

	court := entities.NewCourt(orgID, req.Name)

	if err := h.courtService.Create(r.Context(), court); err != nil {
		log.Error().Err(err).Str("orgID", orgID).Msg("failed to create court")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := CreateCourtResponse{
		ID:             court.ID,
		OrganizationID: court.OrganizationID,
		Name:           court.Name,
	}

	httputil.JSON(w, http.StatusCreated, resp)

	log.Info().
		Str("orgID", orgID).
		Str("courtID", court.ID).
		Msg("court created successfully")
}

// GetCourt godoc
// @Summary Get a court by ID
// @Description Returns a single court by its ID
// @Tags courts
// @Security BearerAuth
// @Param orgID path string true "Organization ID"
// @Param courtID path string true "Court ID"
// @Produce json
// @Success 200 {object} CourtResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500
// @Router /v1/organizations/{orgID}/courts/{courtID} [get]
func (h *CourtHandler) GetCourt(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")
	courtID := chi.URLParam(r, "courtID")

	if orgID == "" || courtID == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "orgID and courtID are required",
		})
		return
	}

	court, err := h.courtService.GetByID(r.Context(), orgID, courtID)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			httputil.JSON(w, http.StatusNotFound, ErrorResponse{
				Message: "court not found",
			})
			return
		}

		log.Error().
			Err(err).
			Str("orgID", orgID).
			Str("courtID", courtID).
			Msg("failed to get court")

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := CourtResponse{
		ID:             court.ID,
		OrganizationID: court.OrganizationID,
		Name:           court.Name,
		CreatedAt:      court.CreatedAt,
	}

	if !court.UpdatedAt.IsZero() {
		resp.UpdatedAt = &court.UpdatedAt
	}

	httputil.JSON(w, http.StatusOK, resp)

	log.Info().
		Str("orgID", orgID).
		Str("courtID", courtID).
		Msg("successfully retrieved court")
}

// ListCourts godoc
// @Summary List all courts for an organization
// @Description Returns all courts belonging to the specified organization
// @Tags courts
// @Security BearerAuth
// @Param orgID path string true "Organization ID"
// @Produce json
// @Success 200 {object} ListCourtsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500
// @Router /v1/organizations/{orgID}/courts [get]
func (h *CourtHandler) ListCourts(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")

	if orgID == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "orgID is required",
		})
		return
	}

	courts, err := h.courtService.ListByOrganizationID(r.Context(), orgID)
	if err != nil {
		log.Error().
			Err(err).
			Str("orgID", orgID).
			Msg("failed to list courts")

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dtos := make([]CourtResponse, 0, len(courts))
	for _, c := range courts {
		resp := CourtResponse{
			ID:             c.ID,
			OrganizationID: c.OrganizationID,
			Name:           c.Name,
			CreatedAt:      c.CreatedAt,
		}

		if !c.UpdatedAt.IsZero() {
			resp.UpdatedAt = &c.UpdatedAt
		}

		dtos = append(dtos, resp)
	}

	httputil.JSON(w, http.StatusOK, ListCourtsResponse{Courts: dtos})

	log.Info().
		Str("orgID", orgID).
		Int("count", len(courts)).
		Msg("successfully listed courts")
}

// UpdateCourt godoc
// @Summary Update a court
// @Description Updates an existing court's information
// @Tags courts
// @Security BearerAuth
// @Param orgID path string true "Organization ID"
// @Param courtID path string true "Court ID"
// @Accept json
// @Produce json
// @Param court body UpdateCourtRequest true "Court update payload"
// @Success 200 {object} CourtResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500
// @Router /v1/organizations/{orgID}/courts/{courtID} [put]
func (h *CourtHandler) UpdateCourt(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")
	courtID := chi.URLParam(r, "courtID")

	if orgID == "" || courtID == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "orgID and courtID are required",
		})
		return
	}

	var req UpdateCourtRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Str("courtID", courtID).Msg("failed to decode update court request")
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{Message: "invalid json"})
		return
	}

	if req.Name == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "name is required",
		})
		return
	}

	updatedCourt, err := h.courtService.UpdateName(r.Context(), orgID, courtID, req.Name)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			httputil.JSON(w, http.StatusNotFound, ErrorResponse{
				Message: "court not found",
			})
			return
		}

		log.Error().Err(err).Str("courtID", courtID).Msg("failed to update court")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := CourtResponse{
		ID:             updatedCourt.ID,
		OrganizationID: updatedCourt.OrganizationID,
		Name:           updatedCourt.Name,
		CreatedAt:      updatedCourt.CreatedAt,
	}

	if !updatedCourt.UpdatedAt.IsZero() {
		resp.UpdatedAt = &updatedCourt.UpdatedAt
	}

	httputil.JSON(w, http.StatusOK, resp)

	log.Info().
		Str("orgID", orgID).
		Str("courtID", courtID).
		Msg("court updated successfully")
}
