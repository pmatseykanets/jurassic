//go:build unit
// +build unit

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	requestID := uuid.NewString()
	body := `{"bar":"baz"}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/foo", strings.NewReader(body))
	r = r.WithContext(context.WithValue(
		r.Context(),
		middleware.RequestIDKey,
		requestID,
	))

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.Copy(w, r.Body); err != nil {
			t.Fatal(err)
		}
	})

	Logger(logger)(h).ServeHTTP(w, r)

	var entry = struct {
		Time     time.Time     `json:"time"`
		Level    string        `json:"level"`
		Msg      string        `json:"msg"`
		ID       string        `json:"id"`
		Method   string        `json:"method"`
		Path     string        `json:"path"`
		Status   int           `json:"status"`
		Bytes    int           `json:"bytes"`
		IP       string        `json:"ip"`
		Duration time.Duration `json:"duration"`
	}{}

	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}

	if entry.Time.IsZero() {
		t.Error("Expected time got zero")
	}
	if want, got := "INFO", entry.Level; want != got {
		t.Errorf("Expected level %s got %s", want, got)
	}
	if want, got := "Request", entry.Msg; want != got {
		t.Errorf("Expected msg %s got %s", want, got)
	}
	if want, got := requestID, entry.ID; want != got {
		t.Errorf("Expected id %s got %s", want, got)
	}
	if want, got := http.MethodPost, entry.Method; want != got {
		t.Errorf("Expected method %s got %s", want, got)
	}
	if want, got := "/foo", entry.Path; want != got {
		t.Errorf("Expected path %s got %s", want, got)
	}
	if want, got := len(body), entry.Bytes; want != got {
		t.Errorf("Expected path %d got %d", want, got)
	}
	if entry.IP == "" {
		t.Error("Expected ip got zero")
	}
	if entry.Duration == 0 {
		t.Error("Expected duration got zero")
	}
}
