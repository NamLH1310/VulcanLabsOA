package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/namlh/vulcanLabsOA/consts/ctxkey"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (rw *responseRecorder) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logging(logger *slog.Logger, next http.Handler) http.Handler {
	var requestCounter atomic.Int64

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := &responseRecorder{ResponseWriter: w}
		start := time.Now()

		requestID := requestCounter.Add(1)

		ww.Header().Set("Request-ID", strconv.FormatInt(requestID, 10))

		ctx := context.WithValue(r.Context(), ctxkey.RequestID{}, requestID)
		r = r.WithContext(ctx)

		defer func() {
			elapsed := time.Since(start)
			logger.LogAttrs(
				r.Context(),
				slog.LevelInfo,
				"incoming request",
				slog.Duration("elapsed", elapsed),
				slog.Int("status", ww.status),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)
		}()

		next.ServeHTTP(ww, r)
	})
}
