//go:build unit
// +build unit

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

func TestRequestIDPropagate(t *testing.T) {
	requestID := uuid.NewString()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/foo", nil)
	r.Header.Set(middleware.RequestIDHeader, requestID)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if want, got := requestID, middleware.GetReqID(r.Context()); want != got {
			t.Errorf("Expected request id %s got %s", want, got)
		}
	})

	RequestID(h).ServeHTTP(w, r)
}

func TestRequestIDNotGenerate(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/foo", nil)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		if requestID == "" {
			t.Error("Expected request id got empty")
		}
	})

	RequestID(h).ServeHTTP(w, r)
}
