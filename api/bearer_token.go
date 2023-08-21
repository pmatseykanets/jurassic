package api

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// BearerToken is a an authentication middleware.
func BearerToken(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			fields := strings.Fields(header)
			if len(fields) != 2 ||
				!strings.EqualFold(fields[0], "Bearer") ||
				len(fields[1]) != len(key) {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			if subtle.ConstantTimeCompare([]byte(key), []byte(fields[1])) != 1 {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
