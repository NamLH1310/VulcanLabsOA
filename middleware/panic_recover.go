package middleware

import (
	"log/slog"
	"net/http"
)

func PanicRecover(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				logger.ErrorContext(
					r.Context(),
					"panic recover",
					"recover_msg", rec,
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
