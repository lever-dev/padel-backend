package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/lever-dev/padel-backend/internal/entities"
	"github.com/lever-dev/padel-backend/pkg/httputil"
	"github.com/rs/zerolog/log"
)

type TokenVerifier interface {
	VerifyToken(token string) error
}

func NewAuthMiddleware(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if authHeader == "" {
				httputil.JSON(w, http.StatusUnauthorized, ErrorResponse{Message: "missing authorization header"})
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
				httputil.JSON(w, http.StatusUnauthorized, ErrorResponse{Message: "invalid authorization header"})
				return
			}

			token := strings.TrimSpace(parts[1])

			if err := verifier.VerifyToken(token); err != nil {
				if errors.Is(err, entities.ErrInvalidToken) || errors.Is(err, entities.ErrExpiredToken) {
					httputil.JSON(w, http.StatusUnauthorized, ErrorResponse{Message: "invalid token"})
					return
				}

				log.Error().Err(err).Msg("verify token failed")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
