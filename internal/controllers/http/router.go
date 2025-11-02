package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

func NewRouter(reservationHandler *ReservationHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Route("/v1", func(r chi.Router) {
		// POST /v1/organizations/{orgID}/courts/{courtID}/reservations
		r.Post("/reservations/{orgID}/courts/{courtID}", reservationHandler.ReserveCourt)
	})

	return r
}
