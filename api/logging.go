package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// Logger is a replacement for chi's Logger middleware
// that logs requests using log/slog.
func Logger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			defer func() {
				logger.Info(
					"Request",
					"id", middleware.GetReqID(r.Context()),
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.Status(),
					"bytes", ww.BytesWritten(),
					"ip", r.RemoteAddr,
					"duration", time.Since(start),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
