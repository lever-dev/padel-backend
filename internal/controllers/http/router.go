package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	_ "github.com/lever-dev/padel-backend/docs"
	swagger "github.com/swaggo/http-swagger"
)

func NewRouter(reservationHandler *ReservationHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Get("/_docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/_docs/index.html", http.StatusMovedPermanently)
	})

	r.Get("/_docs/*", swagger.Handler())

	r.Route("/v1", func(r chi.Router) {
		// POST /v1/organizations/{orgID}/courts/{courtID}/reservations
		r.Post("/reservations/{orgID}/courts/{courtID}", reservationHandler.ReserveCourt)
		r.Delete("/reservations/{orgID}/courts/{courtID}/{reservationID}", reservationHandler.CancelReservation)
	})

	return r
}
