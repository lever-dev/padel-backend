package render

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

func JSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Error().
			Err(err).
			Msg("failed to encode body")
		return
	}
}
