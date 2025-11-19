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

func NewRouter(
	reservationHandler *ReservationHandler,
	authHandler *AuthHandler,
	courtHandler *CourtHandler,
	authMiddleware func(http.Handler) http.Handler,
) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Get("/_docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/_docs/index.html", http.StatusMovedPermanently)
	})

	r.Get("/_docs/*", swagger.Handler())

	r.Route("/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			if authMiddleware != nil {
				r.Use(authMiddleware)
			}

			r.Post("/organizations/{orgID}/courts/{courtID}/reservations", reservationHandler.ReserveCourt)
			r.Delete(
				"/organizations/{orgID}/courts/{courtID}/reservations/{reservationID}",
				reservationHandler.CancelReservation,
			)
			r.Get("/organizations/{orgID}/courts/{courtID}/reservations", reservationHandler.ListReservations)
			r.Get(
				"/organizations/{orgID}/courts/{courtID}/reservations/{reservationID}",
				reservationHandler.GetReservation,
			)

			r.Post("/organizations/{orgID}/courts", courtHandler.CreateCourt)
			r.Get("/organizations/{orgID}/courts", courtHandler.ListCourts)
			r.Get("/organizations/{orgID}/courts/{courtID}", courtHandler.GetCourt)
			r.Put("/organizations/{orgID}/courts/{courtID}", courtHandler.UpdateCourt)
		})

		r.Post("/auth/register", authHandler.RegisterUser)
		r.Post("/auth/login", authHandler.Login)
	})

	return r
}
