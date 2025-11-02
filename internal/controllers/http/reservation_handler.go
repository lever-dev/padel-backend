package http

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/rs/zerolog/log"
)

type ReservationService interface {
	ReserveCourt(ctx context.Context, courtID string, reservation entities.Reservation) error
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
}

type ReserveCourtResponse struct {
}

// TODO: implement
func (h *ReservationHandler) ReserveCourt(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgID")
	courtID := chi.URLParam(r, "courtID")

	// TODO: unmarshal request body to ReserveCourtRequest

	if err := h.rsvService.ReserveCourt(r.Context(), courtID, entities.Reservation{}); err != nil {
		log.Error().Err(err).
			Str("organization id", orgID).
			Str("court id", courtID).
			Msg("failed to reserve court")

		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
