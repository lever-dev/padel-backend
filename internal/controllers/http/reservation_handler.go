package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/rs/zerolog/log"
)

type ReservationService interface {
	ReserveCourt(ctx context.Context, courtID string, reservation *entities.Reservation) error
	CancelReservation(ctx context.Context, reservationID string, cancelledBy string) error
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
	// example: 2025-11-04T18:30Z
	// format: date-time
	StartTime time.Time `json:"startTime" example:"2025-11-04T18:30Z" format:"date-time"`

	// EndTime is the reservation end timestamp in RFC3339 format
	// example: 2025-11-04T19:45Z
	// format: date-time
	EndTime time.Time `json:"endTime" example:"2025-11-04T19:45Z" format:"date-time"`
}

type ReserveCourtResponse struct {
}

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
// @Param orgID path string true "Organization ID"
// @Param courtID path string true "Court ID"
// @Accept json
// @Param reservation body ReserveCourtRequest true "Reservation payload"
// @Success 200
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/reservations/{orgID}/courts/{courtID} [post]
func (h *ReservationHandler) ReserveCourt(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")
	courtID := chi.URLParam(r, "courtID")

	var req ReserveCourtRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().
			Err(err).
			Str("organization id", orgID).
			Str("court id", courtID).
			Msg("failed to decode reserve court request")

		h.sendJSONResponse(w, http.StatusBadRequest, ErrorResponse{Message: "invalid json"})
		return
	}

	// TODO: add user_id after auth service implementation
	reservation := entities.NewReservation(courtID, req.StartTime, req.EndTime, "")

	if err := h.rsvService.ReserveCourt(r.Context(), courtID, reservation); err != nil {
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
// @Param reservationID path string true "Reservation ID"
// @Accept json
// @Param cancel body CancelReservationRequest true "Cancellation payload"
// @Success 200
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/reservations/{reservationID} [delete]
func (h *ReservationHandler) CancelReservation(w http.ResponseWriter, r *http.Request) {
	reservationID := chi.URLParam(r, "reservationID")
	if reservationID == "" {
		http.Error(w, "reservation ID is required", http.StatusBadRequest)
		return
	}

	var req CancelReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("failed to decode request body")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.CancelledBy == "" {
		http.Error(w, "cancelled_by is required", http.StatusBadRequest)
		return
	}

	if err := h.rsvService.CancelReservation(r.Context(), reservationID, req.CancelledBy); err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			http.Error(w, "reservation not found", http.StatusNotFound)
			return
		}

		log.Error().Err(err).
			Str("reservation_id", reservationID).
			Msg("failed to cancel reservation")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

// TODO: move to pkg library
func (h *ReservationHandler) sendJSONResponse(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Error().
			Err(err).
			Msg("failed to encode body")

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
}
