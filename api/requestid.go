package api

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// RequestID is a replacement for chi's RequestID middleware
// that injects a request ID into the context of each request.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(middleware.RequestIDHeader)
		if requestID == "" {
			guid := uuid.New()
			requestID = base64.RawURLEncoding.EncodeToString(guid[:])
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), middleware.RequestIDKey, requestID)))
	})
}
