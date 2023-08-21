//go:build unit
// +build unit

package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/pmatseykanets/jurassic/app"
)

type fakeCageStore struct {
	cage   app.Cage
	id     string
	status app.CageStatus
	err    error
}

func (s *fakeCageStore) Add(_ context.Context, cage *app.Cage) (*app.Cage, error) {
	if s.err != nil {
		return nil, s.err
	}

	now := time.Now()
	s.cage = *cage
	s.cage.ID = uuid.NewString()
	s.cage.CreatedAt = now
	s.cage.UpdatedAt = now
	c := s.cage

	return &c, nil
}

func (s *fakeCageStore) Get(_ context.Context, id string) (*app.Cage, error) {
	if s.err != nil {
		return nil, s.err
	}

	c := s.cage
	s.id = id

	return &c, nil
}

func (s *fakeCageStore) List(_ context.Context, status app.CageStatus) ([]app.Cage, error) {
	if s.err != nil {
		return nil, s.err
	}

	s.status = status

	if s.cage.ID == "" {
		return nil, nil
	}

	return []app.Cage{s.cage}, nil
}

func (s *fakeCageStore) ChangeStatus(_ context.Context, id string, status app.CageStatus) (*app.Cage, error) {
	if s.err != nil {
		return nil, s.err
	}

	s.cage.Status = status
	s.id = id
	s.status = status
	c := s.cage

	return &c, nil
}

func (s *fakeCageStore) Delete(_ context.Context, id string) error {
	if s.err != nil {
		return s.err
	}

	s.id = id

	return nil
}

func TestListCages(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	now := time.Now()
	store := &fakeCageStore{
		cage: app.Cage{
			ID:        uuid.NewString(),
			Capacity:  1,
			Status:    app.CageStatusActive,
			Occupancy: 1,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	svc := &Server{
		Logger:    logger,
		CageStore: store,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/cages", nil)

	svc.ListCages().ServeHTTP(w, r)

	if want, got := http.StatusOK, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}

	if w.Body.Len() == 0 {
		t.Fatal("Expected body")
	}

	response := struct {
		Data []app.Cage `json:"data"`
	}{}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if want, got := 1, len(response.Data); want != got {
		t.Fatalf("Expected cages %d got %d", want, got)
	}

	if want, got := store.cage.ID, response.Data[0].ID; want != got {
		t.Fatalf("Expected ID %s got %s", want, got)
	}
	if want, got := store.cage.Status, response.Data[0].Status; want != got {
		t.Fatalf("Expected Status %s got %s", want, got)
	}
	if want, got := store.cage.Capacity, response.Data[0].Capacity; want != got {
		t.Fatalf("Expected Capacity %d got %d", want, got)
	}
	if want, got := store.cage.Occupancy, response.Data[0].Occupancy; want != got {
		t.Fatalf("Expected ID %d got %d", want, got)
	}
	if response.Data[0].CreatedAt.IsZero() {
		t.Fatal("Expected CreatedAt got empty")
	}
	if response.Data[0].UpdatedAt.IsZero() {
		t.Fatal("Expected UpdatedAt got empty")
	}
}

func TestListCagesInvalidFilter(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	store := &fakeCageStore{}
	svc := &Server{
		Logger:    logger,
		CageStore: store,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/cages?status=foo", nil)

	svc.ListCages().ServeHTTP(w, r)

	if want, got := http.StatusBadRequest, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}
}

func TestListCagesEmptyListIsRenderedCorrectly(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	svc := &Server{
		Logger:    logger,
		CageStore: &fakeCageStore{},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/cages", nil)

	svc.ListCages().ServeHTTP(w, r)

	if want, got := http.StatusOK, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}

	if want, got := `{"data":[]}`, strings.TrimSpace(w.Body.String()); want != got {
		t.Fatalf("Expected body %s got %s", want, got)
	}
}

func TestListCagesInternalError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	store := &fakeCageStore{
		err: errors.New("something went wrong"),
	}

	svc := &Server{
		Logger:    logger,
		CageStore: store,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/cages", nil)

	svc.ListCages().ServeHTTP(w, r)

	if want, got := http.StatusInternalServerError, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}
}

func TestAddCage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	store := &fakeCageStore{}

	svc := &Server{
		Logger:    logger,
		CageStore: store,
	}

	body := `{"capacity": 1, "status": "active"}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/cages", strings.NewReader(body))

	svc.AddCage().ServeHTTP(w, r)

	if want, got := http.StatusCreated, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}

	if w.Body.Len() == 0 {
		t.Fatal("Expected body")
	}

	response := struct {
		Data app.Cage `json:"data"`
	}{}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if response.Data.ID == "" {
		t.Fatalf("Expected ID got empty")
	}
	if want, got := app.CageStatusActive, response.Data.Status; want != got {
		t.Fatalf("Expected Status %s got %s", want, got)
	}
	if want, got := 1, response.Data.Capacity; want != got {
		t.Fatalf("Expected Capacity %d got %d", want, got)
	}
	if want, got := 0, response.Data.Occupancy; want != got {
		t.Fatalf("Expected ID %d got %d", want, got)
	}
	if response.Data.CreatedAt.IsZero() {
		t.Fatal("Expected CreatedAt got empty")
	}
	if response.Data.UpdatedAt.IsZero() {
		t.Fatal("Expected UpdatedAt got empty")
	}
}

func TestAddCageBadRequest(t *testing.T) {
	tests := []struct {
		desc string
		body string
	}{
		{
			desc: "no body",
			body: "",
		},
		{
			desc: "no capacity",
			body: `{"status": "active"}`,
		},
		{
			desc: "no status",
			body: `{"capacity": 1}`,
		},
		{
			desc: "zero capacity",
			body: `{"capacity": 0, "status": "active"}`,
		},
		{
			desc: "negative capacity",
			body: `{"capacity": -1, "status": "active"}`,
		},
		{
			desc: "invalid status",
			body: `{"capacity": 1, "status": "foo"}`,
		},
		{
			desc: "invalid request body",
			body: `{"capacity": 1, "status": "foo"`,
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	svc := &Server{
		Logger:    logger,
		CageStore: &fakeCageStore{},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var bodyReader io.Reader
			if tt.body != "" {
				bodyReader = strings.NewReader(tt.body)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/cages", bodyReader)

			svc.AddCage().ServeHTTP(w, r)

			if want, got := http.StatusBadRequest, w.Code; want != got {
				t.Fatalf("Expected %d got %d", want, got)
			}
		})
	}
}

func TestAddCageInternalError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	store := &fakeCageStore{
		err: errors.New("something went wrong"),
	}

	svc := &Server{
		Logger:    logger,
		CageStore: store,
	}

	body := `{"capacity": 1, "status": "active"}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/cages", strings.NewReader(body))

	svc.AddCage().ServeHTTP(w, r)

	if want, got := http.StatusInternalServerError, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}
}

func TestGetCage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	id := uuid.NewString()
	now := time.Now()
	store := &fakeCageStore{
		cage: app.Cage{
			ID:        uuid.NewString(),
			Capacity:  1,
			Status:    app.CageStatusActive,
			Occupancy: 1,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	svc := &Server{
		Logger:    logger,
		CageStore: store,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/cages/"+id, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	svc.GetCage().ServeHTTP(w, r)

	if want, got := http.StatusOK, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}

	if w.Body.Len() == 0 {
		t.Fatal("Expected body got empty")
	}

	response := struct {
		Data app.Cage `json:"data"`
	}{}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if want, got := store.cage.ID, response.Data.ID; want != got {
		t.Fatalf("Expected ID %s got %s", want, got)
	}
	if want, got := store.cage.Status, response.Data.Status; want != got {
		t.Fatalf("Expected Status %s got %s", want, got)
	}
	if want, got := store.cage.Capacity, response.Data.Capacity; want != got {
		t.Fatalf("Expected Capacity %d got %d", want, got)
	}
	if want, got := store.cage.Occupancy, response.Data.Occupancy; want != got {
		t.Fatalf("Expected ID %d got %d", want, got)
	}
	if response.Data.CreatedAt.IsZero() {
		t.Fatal("Expected CreatedAt got empty")
	}
	if response.Data.UpdatedAt.IsZero() {
		t.Fatal("Expected UpdatedAt got empty")
	}
}

func TestGetCageNotFoundError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	id := uuid.NewString()
	store := &fakeCageStore{
		err: app.ErrNotFound,
	}

	svc := &Server{
		Logger:    logger,
		CageStore: store,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/cages/"+id, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	svc.GetCage().ServeHTTP(w, r)

	if want, got := http.StatusNotFound, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}
}

func TestChangeCageStatus(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	id := uuid.NewString()
	now := time.Now()
	store := &fakeCageStore{
		cage: app.Cage{
			ID:        id,
			Capacity:  1,
			Status:    app.CageStatusActive,
			Occupancy: 0,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	svc := &Server{
		Logger:    logger,
		CageStore: store,
	}

	body := `{"status": "down"}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/cages/"+id, strings.NewReader(body))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	svc.ChangeCageStatus().ServeHTTP(w, r)

	if want, got := http.StatusOK, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}

	if want, got := id, store.id; want != got {
		t.Fatalf("Expected ID %s got %s", want, got)
	}
	if want, got := app.CageStatusDown, store.status; want != got {
		t.Fatalf("Expected ID %s got %s", want, got)
	}

	if w.Body.Len() == 0 {
		t.Fatal("Expected body got empty")
	}

	response := struct {
		Data app.Cage `json:"data"`
	}{}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if want, got := app.CageStatusDown, response.Data.Status; want != got {
		t.Fatalf("Expected Status %s got %s", want, got)
	}
}

func TestChangeCageStatusNotFoundError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	id := uuid.NewString()
	store := &fakeCageStore{
		err: app.ErrNotFound,
	}

	svc := &Server{
		Logger:    logger,
		CageStore: store,
	}

	body := `{"status": "down"}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/cages/"+id, strings.NewReader(body))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	svc.ChangeCageStatus().ServeHTTP(w, r)

	if want, got := http.StatusNotFound, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}
}

func TestChangeCageStatusInternalError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	id := uuid.NewString()
	store := &fakeCageStore{
		err: errors.New("something went wrong"),
	}

	svc := &Server{
		Logger:    logger,
		CageStore: store,
	}

	body := `{"status": "down"}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/cages/"+id, strings.NewReader(body))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	svc.ChangeCageStatus().ServeHTTP(w, r)

	if want, got := http.StatusInternalServerError, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}
}

func TestChangeCageStatusConflict(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	id := uuid.NewString()
	store := &fakeCageStore{
		err: app.ErrConflict,
	}

	svc := &Server{
		Logger:    logger,
		CageStore: store,
	}

	body := `{"status": "down"}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/cages/"+id, strings.NewReader(body))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	svc.ChangeCageStatus().ServeHTTP(w, r)

	if want, got := http.StatusConflict, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}
}

func TestChangeCageStatusBadRequest(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	id := uuid.NewString()
	store := &fakeCageStore{}

	svc := &Server{
		Logger:    logger,
		CageStore: store,
	}

	body := `{"status": "foo"}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/cages/"+uuid.NewString(), strings.NewReader(body))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	svc.ChangeCageStatus().ServeHTTP(w, r)

	if want, got := http.StatusBadRequest, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}
}

func TestDeleteCage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	id := uuid.NewString()
	store := &fakeCageStore{}

	svc := &Server{
		Logger:    logger,
		CageStore: store,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/cages/"+id, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	svc.DeleteCage().ServeHTTP(w, r)

	if want, got := http.StatusOK, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}
	if want, got := id, store.id; want != got {
		t.Fatalf("Expected %s got %s", want, got)
	}
}
