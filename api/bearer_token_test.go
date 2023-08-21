//go:build unit
// +build unit

package api

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestBearerToken(t *testing.T) {
	token := base64.RawURLEncoding.EncodeToString([]byte(uuid.NewString()))
	invalidToken := base64.RawURLEncoding.EncodeToString([]byte(uuid.NewString()))

	tests := []struct {
		desc   string
		header string
		code   int
	}{
		{
			desc:   "empty header",
			header: "",
			code:   401,
		},
		{
			desc:   "invalid scheme",
			header: "Basic " + token,
			code:   401,
		},
		{
			desc:   "invalid token length",
			header: "Bearer " + token[:len(token)-1],
			code:   401,
		},
		{
			desc:   "invalid token",
			header: "Bearer " + invalidToken,
			code:   401,
		},
		{
			desc:   "valid token",
			header: "Bearer " + token,
			code:   200,
		},
		{
			desc:   "valid token with extra spaces",
			header: "Bearer   " + token,
			code:   200,
		},
		{
			desc:   "valid token with lowercase scheme",
			header: "bearer " + token,
			code:   200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/foo", nil)
			if tt.header != "" {
				r.Header.Set("Authorization", tt.header)
			}
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			BearerToken(token)(h).ServeHTTP(w, r)

			if want, got := tt.code, w.Code; want != got {
				t.Errorf("Expected status code %d got %d", want, got)
			}
		})
	}
}
