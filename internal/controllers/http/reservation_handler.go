package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/lever-dev/padel-backend/pkg/httputil"
	"github.com/rs/zerolog/log"
)

type ReservationService interface {
	ReserveCourt(ctx context.Context, courtID string, reservation *entities.Reservation) error
	ListReservations(ctx context.Context, courtID string, from, to time.Time) ([]entities.Reservation, error)
	CancelReservation(ctx context.Context, reservationID string, cancelledBy string) error
	GetReservation(ctx context.Context, courtID, reservstionID string) (*entities.Reservation, error)
}

type ReservationHandler struct {
	rsvService ReservationService
}

func NewReservationHandler(service ReservationService) *ReservationHandler {
	return &ReservationHandler{
		rsvService: service,
	}
}

type ReserveCourtRequest struct {
	// StartTime is the reservation start timestamp in RFC3339 format
	// example: 2025-11-04T18:30
	// format: date-time
	StartTime time.Time `json:"startTime" example:"2025-11-04T18:30" format:"date-time"`

	// EndTime is the reservation end timestamp in RFC3339 format
	// example: 2025-11-04T19:45
	// format: date-time
	EndTime time.Time `json:"endTime" example:"2025-11-04T19:45" format:"date-time"`
}

func (r *ReserveCourtRequest) UnmarshalJSON(data []byte) error {
	// Define a temporary structure with strings
	var raw struct {
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decode request: %w", err)
	}

	var err error
	r.StartTime, err = httputil.ParseTime(raw.StartTime)
	if err != nil {
		return fmt.Errorf("parse startTime: %w", err)
	}

	r.EndTime, err = httputil.ParseTime(raw.EndTime)
	if err != nil {
		return fmt.Errorf("parse endTime: %w", err)
	}

	return nil
}

type ReserveCourtResponse struct{}

// ErrorResponse represents a standard error body.
// swagger:model ErrorResponse
type ErrorResponse struct {
	// Message provides a human-readable description of the error.
	Message string `json:"message" example:"invalid JSON body"`
}

// ReserveCourt godoc
// @Summary Reserve a court
// @Description Creates a reservation for the specified organization and court.
// @Tags reservations
// @Security BearerAuth
// @Param orgID path string true "Organization ID"
// @Param courtID path string true "Court ID"
// @Accept json
// @Param reservation body ReserveCourtRequest true "Reservation payload"
// @Success 200
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500
// @Router /v1/organizations/{orgID}/courts/{courtID}/reserve [post]
func (h *ReservationHandler) ReserveCourt(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")
	courtID := chi.URLParam(r, "courtID")

	if orgID == "" || courtID == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "orgID and courtID are required",
		})
		return
	}

	var req ReserveCourtRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().
			Err(err).
			Str("organization id", orgID).
			Str("court id", courtID).
			Msg("failed to decode reserve court request")

		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{Message: "invalid json"})
		return
	}

	if req.StartTime.IsZero() || req.EndTime.IsZero() {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "startTime and endTime are required",
		})
		return
	}

	if !req.StartTime.Before(req.EndTime) {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "startTime must be before endTime",
		})
		return
	}

	// TODO: add user_id after auth service implementation
	reservation := entities.NewReservation(courtID, req.StartTime, req.EndTime, "")

	if err := h.rsvService.ReserveCourt(r.Context(), courtID, reservation); err != nil {
		if errors.Is(err, entities.ErrCourtAlreadyReserved) {
			httputil.JSON(w, http.StatusConflict, ErrorResponse{
				Message: "court is already reserved for this time slot",
			})
			return
		}
		log.Error().
			Err(err).
			Str("organization id", orgID).
			Str("court id", courtID).
			Msg("failed to reserve court")

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Info().
		Str("organization id", orgID).
		Str("court id", courtID).
		Time("start time", req.StartTime).
		Time("end time", req.EndTime).
		Msg("court was reserved")
}

// CancelReservationRequest represents the payload for cancelling a reservation
// swagger:model CancelReservationRequest
type CancelReservationRequest struct {
	// CancelledBy is the ID or name of the user performing the cancellation
	// example: "user_123"
	CancelledBy string `json:"cancelledBy"`
}

// CancelReservation godoc
// @Summary Cancel a reservation
// @Description Cancels the reservation with the specified ID.
// @Tags reservations
// @Security BearerAuth
// @Param reservationID path string true "Reservation ID"
// @Accept json
// @Param cancel body CancelReservationRequest true "Cancellation payload"
// @Success 200
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500
// @Router /v1/organizations/{orgID}/courts/{courtID}/reservations/{reservationID} [delete]
func (h *ReservationHandler) CancelReservation(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")
	courtID := chi.URLParam(r, "courtID")
	reservationID := chi.URLParam(r, "reservationID")

	if orgID == "" || courtID == "" || reservationID == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "orgID, courtID, reservationID are required",
		})
		return
	}

	var req CancelReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("failed to decode request body")
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{Message: "invalid json"})
		return
	}

	if req.CancelledBy == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{Message: "cancelledBy is required"})
		return
	}

	if err := h.rsvService.CancelReservation(r.Context(), reservationID, req.CancelledBy); err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			httputil.JSON(w, http.StatusNotFound, ErrorResponse{Message: "reservation not found"})
			return
		}

		log.Error().Err(err).
			Str("reservation_id", reservationID).
			Msg("failed to cancel reservation")

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Info().
		Str("reservation_id", reservationID).
		Str("cancelled_by", req.CancelledBy).
		Msg("reservation cancelled successfully")
}

type ReservationResponse struct {
	ID           string    `json:"id"                    example:"res-123"`
	CourtID      string    `json:"courtId"               example:"court-456"`
	Status       string    `json:"status"                example:"reserved"`
	ReservedFrom time.Time `json:"reservedFrom"          example:"2025-11-04T18:30Z"    format:"date-time"`
	ReservedTo   time.Time `json:"reservedTo"            example:"2025-11-04T19:45Z"    format:"date-time"`
	ReservedBy   string    `json:"reservedBy"            example:"user-789"`
	CancelledBy  string    `json:"cancelledBy,omitempty" example:""`
	CreatedAt    time.Time `json:"createdAt"             example:"2025-11-01T10:00:00Z" format:"date-time"`
}

type ListReservationsResponse struct {
	Reservations []ReservationResponse `json:"reservations"`
}

// ListReservations godoc
// @Summary List reservations
// @Description Returns all reservations for a court within a time range
// @Tags reservations
// @Security BearerAuth
// @Param orgID path string true "Organization ID"
// @Param courtID path string true "Court ID"
// @Param from query string true "Start time in RFC3339 format" format:"date-time"
// @Param to query string true "End time in RFC3339 format" format:"date-time"
// @Produce json
// @Success 200 {object} ListReservationsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500
// @Router /v1/organizations/{orgID}/courts/{courtID}/reservations [get]
func (h *ReservationHandler) ListReservations(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")
	courtID := chi.URLParam(r, "courtID")

	if orgID == "" || courtID == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "orgID and courtID are required",
		})
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "from and to query parameters are required (RFC3339 format)",
		})
		return
	}

	from, err := httputil.ParseTime(fromStr)
	if err != nil {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "invalid from time format, expected RFC3339",
		})
		return
	}

	to, err := httputil.ParseTime(toStr)
	if err != nil {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "invalid to time format, expected RFC3339",
		})
		return
	}

	if !from.Before(to) {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "from must be before to",
		})
		return
	}

	revs, err := h.rsvService.ListReservations(r.Context(), courtID, from, to)
	if err != nil {
		log.Error().
			Err(err).
			Str("organization id", orgID).
			Str("court id", courtID).
			Msg("failed to get reservations")

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dtos := make([]ReservationResponse, 0, len(revs))
	for _, res := range revs {
		dtos = append(dtos, ReservationResponse{
			ID:           res.ID,
			CourtID:      res.CourtID,
			ReservedBy:   res.ReservedBy,
			Status:       string(res.Status),
			ReservedFrom: res.ReservedFrom,
			ReservedTo:   res.ReservedTo,
			CreatedAt:    res.CreatedAt,
		})
	}

	httputil.JSON(w, http.StatusOK, ListReservationsResponse{Reservations: dtos})

	log.Info().
		Str("organization_id", orgID).
		Str("court_id", courtID).
		Msg("successfully listed reservations")
}

// GetReservation godoc
// @Summary Get a reservation
// @Description Retrieves the reservation with the specified ID.
// @Tags reservations
// @Security BearerAuth
// @Param orgID path string true "Organization ID"
// @Param courtID path string true "Court ID"
// @Param reservationID path string true "Reservation ID"
// @Produce json
// @Success 200 {object} ReservationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404
// @Failure 500
// @Router /v1/organizations/{orgID}/courts/{courtID}/reservations/{reservationID} [get]
func (h *ReservationHandler) GetReservation(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")
	courtID := chi.URLParam(r, "courtID")
	reservationID := chi.URLParam(r, "reservationID")

	if orgID == "" || courtID == "" || reservationID == "" {
		httputil.JSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "orgID, courtID and reservationID are required",
		})
		return
	}

	rev, err := h.rsvService.GetReservation(r.Context(), courtID, reservationID)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		log.Error().
			Err(err).
			Str("orgID", orgID).
			Str("courtID", courtID).
			Str("reservationID", reservationID).
			Msg("failed to get reservation")

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := ReservationResponse{
		ID:           rev.ID,
		CourtID:      rev.CourtID,
		ReservedBy:   rev.ReservedBy,
		Status:       string(rev.Status),
		ReservedFrom: rev.ReservedFrom,
		ReservedTo:   rev.ReservedTo,
		CancelledBy:  rev.CancelledBy,
		CreatedAt:    rev.CreatedAt,
	}

	httputil.JSON(w, http.StatusOK, resp)
}
