package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/pmatseykanets/jurassic/app"
)

// AddDinosaurRequest is a request to add a dinosaur.
type AddDinosaurRequest struct {
	Name    string              `json:"name"`
	Species app.DinosaurSpecies `json:"species"`
}

// Validate validates the request.
func (r AddDinosaurRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}

	if r.Species.IsUnspecified() {
		return errors.New("species is required")
	}

	return r.Species.Validate()
}

// AddDinosaur adds a dinosaur to a cage.
// POST /cages/:id/dinosaurs
func (s *Server) AddDinosaur() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		logger := s.Logger.With("requestId", requestID)
		id := chi.URLParam(r, "id")

		if err := app.ValidateID(id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var req AddDinosaurRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("Error decoding request body", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if err := req.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dinosaur, err := s.DinosaurStore.Add(r.Context(), &app.Dinosaur{
			Name:    req.Name,
			Species: req.Species,
			CageID:  id,
		})
		if err != nil {
			switch err {
			case app.ErrNotFound:
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			case app.ErrCagePoweredDown, app.ErrCapacityExceeded, app.ErrSpeciesMismatch:
				http.Error(w, err.Error(), http.StatusConflict)
			default:
				logger.Error("Error adding dinosaur", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		response := struct {
			Data *app.Dinosaur `json:"data"`
		}{
			Data: dinosaur,
		}

		if err = json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("Error marshalling response", "error", err)
			return
		}
	}
}

// ListCageDinosaurs lists dinosaurs in a cage.
// GET /cages/:id/dinosaurs[?species=...]
func (s *Server) ListCageDinosaurs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		logger := s.Logger.With("requestId", requestID)
		id := chi.URLParam(r, "id")

		if err := app.ValidateID(id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		species := app.DinosaurSpecies(r.URL.Query().Get("species"))
		if !species.IsUnspecified() {
			if err := species.Validate(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		dinosaurs, err := s.DinosaurStore.List(r.Context(), id, species)
		if err != nil {
			if err == app.ErrNotFound {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			logger.Error("Error getting dinosaurs", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if dinosaurs == nil {
			dinosaurs = []app.Dinosaur{}
		}

		w.Header().Set("Content-Type", "application/json")

		response := struct {
			Data []app.Dinosaur `json:"data"`
		}{
			Data: dinosaurs,
		}
		if err = json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("Error marshalling response", "error", err)
			return
		}
	}
}

// ListAllDinosaurs lists all dinosaurs.
// GET dinosaurs[&species=...]
func (s *Server) ListAllDinosaurs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		logger := s.Logger.With("requestId", requestID)

		species := app.DinosaurSpecies(r.URL.Query().Get("species"))
		if !species.IsUnspecified() {
			if err := species.Validate(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		dinosaurs, err := s.DinosaurStore.List(r.Context(), app.IDUnspecified, species)
		if err != nil {
			if err == app.ErrNotFound {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			logger.Error("Error getting dinosaurs", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if dinosaurs == nil {
			dinosaurs = []app.Dinosaur{}
		}

		w.Header().Set("Content-Type", "application/json")

		response := struct {
			Data []app.Dinosaur `json:"data"`
		}{
			Data: dinosaurs,
		}
		if err = json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("Error marshalling response", "error", err)
			return
		}
	}
}

// GetDinosaur gets a dinosaur by id.
// GET /dinosaurs/:id
func (s *Server) GetDinosaur() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		logger := s.Logger.With("requestId", requestID)
		id := chi.URLParam(r, "id")

		if err := app.ValidateID(id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dinosaur, err := s.DinosaurStore.Get(r.Context(), id)
		if err != nil {
			if err == app.ErrNotFound {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			logger.Error("Error getting dinosaur", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		response := struct {
			Data *app.Dinosaur `json:"data"`
		}{
			Data: dinosaur,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("Error marshalling response", "error", err)
			return
		}
	}
}

// MoveDinosaurRequest is a request to move a dinosaur to a different cage.
type MoveDinosaurRequest struct {
	CageID string `json:"cageId"`
}

// Validate validates the request.
func (r *MoveDinosaurRequest) Validate() error {
	if r.CageID == "" {
		return errors.New("cageId is required")
	}

	return nil
}

// MoveDinosaur moves a dinosaur to a different cage.
// PUT /dinosaurs/:id
func (s *Server) MoveDinosaur() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		logger := s.Logger.With("requestId", requestID)
		id := chi.URLParam(r, "id")

		if err := app.ValidateID(id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var req MoveDinosaurRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("Error decoding request body", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if err := req.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dinosaur, err := s.DinosaurStore.Move(r.Context(), id, req.CageID)
		if err != nil {
			switch err {
			case app.ErrNotFound:
				// NOTE: This can be improved by differentiating between
				// a dinosaur or a cage being not found.
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			case app.ErrCagePoweredDown, app.ErrCapacityExceeded, app.ErrSpeciesMismatch:
				http.Error(w, err.Error(), http.StatusConflict)
			default:
				logger.Error("Error moving dinosaur", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}

			return
		}

		w.Header().Set("Content-Type", "application/json")

		response := struct {
			Data *app.Dinosaur `json:"data"`
		}{
			Data: dinosaur,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("Error marshalling response", "error", err)
			return
		}
	}
}

// DeleteDinosaur deletes a dinosaur.
// DELETE /dinosaur/:id
func (s *Server) DeleteDinosaur() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		logger := s.Logger.With("requestId", requestID)
		id := chi.URLParam(r, "id")

		if err := app.ValidateID(id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := s.DinosaurStore.Delete(r.Context(), id)
		if err != nil {
			switch err {
			case app.ErrNotFound:
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			case app.ErrConflict:
				http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			default:
				logger.Error("Error deleting dinosaur", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
