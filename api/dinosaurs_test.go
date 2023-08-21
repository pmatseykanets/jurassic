//go:build unit
// +build unit

package api

import (
	"context"
	"encoding/json"
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

type fakeDinosaurStore struct {
	dinosaur app.Dinosaur
	id       string
	species  app.DinosaurSpecies
	cageID   string
	err      error
}

func (s *fakeDinosaurStore) Add(_ context.Context, dinosaur *app.Dinosaur) (*app.Dinosaur, error) {
	if s.err != nil {
		return nil, s.err
	}

	now := time.Now()
	s.dinosaur = *dinosaur
	s.dinosaur.ID = uuid.NewString()
	s.dinosaur.CreatedAt = now
	s.dinosaur.UpdatedAt = now
	c := s.dinosaur

	return &c, nil
}

func (s *fakeDinosaurStore) Get(_ context.Context, id string) (*app.Dinosaur, error) {
	if s.err != nil {
		return nil, s.err
	}

	d := s.dinosaur
	s.id = id

	return &d, nil
}

func (s *fakeDinosaurStore) List(_ context.Context, cageID string, species app.DinosaurSpecies) ([]app.Dinosaur, error) {
	if s.err != nil {
		return nil, s.err
	}

	s.cageID = cageID
	s.species = species

	if s.dinosaur.ID == "" {
		return nil, nil
	}

	return []app.Dinosaur{s.dinosaur}, nil
}

func (s *fakeDinosaurStore) Move(_ context.Context, id string, cageID string) (*app.Dinosaur, error) {
	if s.err != nil {
		return nil, s.err
	}

	s.dinosaur.CageID = cageID
	s.id = id
	s.cageID = cageID
	d := s.dinosaur

	return &d, nil
}

func (s *fakeDinosaurStore) Delete(_ context.Context, id string) error {
	if s.err != nil {
		return s.err
	}

	s.id = id

	return nil
}

func TestAddDinosaur(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	id := uuid.NewString()
	store := &fakeDinosaurStore{}

	svc := &Server{
		Logger:        logger,
		DinosaurStore: store,
	}

	body := `{"name": "Tyrannosaurus Rex", "species": "tyrannosaurus"}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/cages/"+id+"/dinosaurs", strings.NewReader(body))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	svc.AddDinosaur().ServeHTTP(w, r)

	if want, got := http.StatusCreated, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}

	if w.Body.Len() == 0 {
		t.Fatal("Expected body")
	}

	response := struct {
		Data app.Dinosaur `json:"data"`
	}{}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if response.Data.ID == "" {
		t.Fatalf("Expected ID got empty")
	}
	if want, got := "Tyrannosaurus Rex", response.Data.Name; want != got {
		t.Fatalf("Expected Name %s got %s", want, got)
	}
	if want, got := app.DinosaurSpeciesTyrannosaurus, response.Data.Species; want != got {
		t.Fatalf("Expected Species %s got %s", want, got)
	}
	if want, got := id, response.Data.CageID; want != got {
		t.Fatalf("Expected CageID %s got %s", want, got)
	}
	if response.Data.CreatedAt.IsZero() {
		t.Fatal("Expected CreatedAt got empty")
	}
	if response.Data.UpdatedAt.IsZero() {
		t.Fatal("Expected UpdatedAt got empty")
	}
}

func TestGetDinosaur(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	id := uuid.NewString()
	cageID := uuid.NewString()
	now := time.Now()
	store := &fakeDinosaurStore{
		dinosaur: app.Dinosaur{
			ID:        id,
			Name:      "Tyrannosaurus Rex",
			Species:   app.DinosaurSpeciesTyrannosaurus,
			CageID:    cageID,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	svc := &Server{
		Logger:        logger,
		DinosaurStore: store,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/dinosaurs/"+id, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	svc.GetDinosaur().ServeHTTP(w, r)

	if want, got := http.StatusOK, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}

	if w.Body.Len() == 0 {
		t.Fatal("Expected body got empty")
	}

	response := struct {
		Data app.Dinosaur `json:"data"`
	}{}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if want, got := store.dinosaur.ID, response.Data.ID; want != got {
		t.Fatalf("Expected ID %s got %s", want, got)
	}
	if want, got := "Tyrannosaurus Rex", response.Data.Name; want != got {
		t.Fatalf("Expected Name %s got %s", want, got)
	}
	if want, got := app.DinosaurSpeciesTyrannosaurus, response.Data.Species; want != got {
		t.Fatalf("Expected Species %s got %s", want, got)
	}
	if want, got := cageID, response.Data.CageID; want != got {
		t.Fatalf("Expected CageID %s got %s", want, got)
	}
	if response.Data.CreatedAt.IsZero() {
		t.Fatal("Expected CreatedAt got empty")
	}
	if response.Data.UpdatedAt.IsZero() {
		t.Fatal("Expected UpdatedAt got empty")
	}
}

func TestListCageDinosaurs(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	id := uuid.NewString()
	cageID := uuid.NewString()

	now := time.Now()
	store := &fakeDinosaurStore{
		dinosaur: app.Dinosaur{
			ID:        id,
			Name:      "Tyrannosaurus Rex",
			Species:   app.DinosaurSpeciesTyrannosaurus,
			CageID:    cageID,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	svc := &Server{
		Logger:        logger,
		DinosaurStore: store,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/cages/"+id+"/dinosaurs", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	svc.ListCageDinosaurs().ServeHTTP(w, r)

	if want, got := http.StatusOK, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}

	if w.Body.Len() == 0 {
		t.Fatal("Expected body got empty")
	}

	response := struct {
		Data []app.Dinosaur `json:"data"`
	}{}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if want, got := 1, len(response.Data); want != got {
		t.Fatalf("Expected dinosaurs %d got %d", want, got)
	}
	if want, got := id, response.Data[0].ID; want != got {
		t.Fatalf("Expected ID %s got %s", want, got)
	}
}

func TestListAllDinosaurs(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	id := uuid.NewString()
	cageID := uuid.NewString()
	now := time.Now()
	store := &fakeDinosaurStore{
		dinosaur: app.Dinosaur{
			ID:        id,
			Name:      "Tyrannosaurus Rex",
			Species:   app.DinosaurSpeciesTyrannosaurus,
			CageID:    cageID,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	svc := &Server{
		Logger:        logger,
		DinosaurStore: store,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/dinosaurs", nil)

	svc.ListAllDinosaurs().ServeHTTP(w, r)

	if want, got := http.StatusOK, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}

	if w.Body.Len() == 0 {
		t.Fatal("Expected body got empty")
	}

	response := struct {
		Data []app.Dinosaur `json:"data"`
	}{}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if want, got := 1, len(response.Data); want != got {
		t.Fatalf("Expected dinosaurs %d got %d", want, got)
	}
	if want, got := id, response.Data[0].ID; want != got {
		t.Fatalf("Expected ID %s got %s", want, got)
	}
}

func TestMoveDinosaur(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	id := uuid.NewString()
	cage1ID := uuid.NewString()
	cage2ID := uuid.NewString()
	now := time.Now()
	store := &fakeDinosaurStore{
		dinosaur: app.Dinosaur{
			ID:        id,
			Name:      "Tyrannosaurus Rex",
			Species:   app.DinosaurSpeciesTyrannosaurus,
			CageID:    cage1ID,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	svc := &Server{
		Logger:        logger,
		DinosaurStore: store,
	}

	body := `{"cageId": "` + cage2ID + `"}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/dinosaurs/"+id, strings.NewReader(body))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	svc.MoveDinosaur().ServeHTTP(w, r)

	if want, got := http.StatusOK, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}
	if want, got := id, store.id; want != got {
		t.Errorf("Expected id %s got %s", want, got)
	}
	if want, got := cage2ID, store.cageID; want != got {
		t.Errorf("Expected cageID %s got %s", want, got)
	}
	if want, got := cage2ID, store.dinosaur.CageID; want != got {
		t.Errorf("Expected CageID %s got %s", want, got)
	}
}

func TestDeleteDinosaur(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	id := uuid.NewString()
	cageID := uuid.NewString()
	now := time.Now()
	store := &fakeDinosaurStore{
		dinosaur: app.Dinosaur{
			ID:        id,
			Name:      "Tyrannosaurus Rex",
			Species:   app.DinosaurSpeciesTyrannosaurus,
			CageID:    cageID,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	svc := &Server{
		Logger:        logger,
		DinosaurStore: store,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/dinosaurs/"+id, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	svc.DeleteDinosaur().ServeHTTP(w, r)

	if want, got := http.StatusOK, w.Code; want != got {
		t.Fatalf("Expected %d got %d", want, got)
	}
	if want, got := id, store.id; want != got {
		t.Errorf("Expected id %s got %s", want, got)
	}
}
