package controller

import (
	"log/slog"
	"net/http"
)

func HealthCheck(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		if _, err := w.Write([]byte("OK\n")); err != nil {
			logger.ErrorContext(r.Context(), "write response failed", "error", err)
			internalErrResponse(w)
		}
	}
}
