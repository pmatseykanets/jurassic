package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/pmatseykanets/jurassic/app"
)

// ListCages lists all cages.
// GET /cages[&status=active|down]
func (s *Server) ListCages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		logger := s.Logger.With("requestId", requestID)

		status := app.CageStatus(r.URL.Query().Get("status"))
		if !status.IsUnspecified() {
			if err := status.Validate(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		cages, err := s.CageStore.List(r.Context(), status)
		if err != nil {
			logger.Error("Error getting cages", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if cages == nil {
			cages = []app.Cage{}
		}

		w.Header().Set("Content-Type", "application/json")

		response := struct {
			Data []app.Cage `json:"data"`
		}{
			Data: cages,
		}
		if err = json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("Error marshalling response", "error", err)
			return
		}
	}
}

// GetCage gets a cage by id.
// GET /cages/:id
func (s *Server) GetCage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		logger := s.Logger.With("requestId", requestID)
		id := chi.URLParam(r, "id")

		cage, err := s.CageStore.Get(r.Context(), id)
		if err != nil {
			if err == app.ErrNotFound {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			logger.Error("Error getting cage", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		response := struct {
			Data *app.Cage `json:"data"`
		}{
			Data: cage,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("Error marshalling response", "error", err)
			return
		}
	}
}

type AddCageRequest struct {
	Capacity int            `json:"capacity"`
	Status   app.CageStatus `json:"status"`
}

func (r AddCageRequest) Validate() error {
	if r.Capacity <= 0 {
		return errors.New("invalid capacity")
	}

	return r.Status.Validate()
}

// AddCage adds a new cage.
// POST /cages
func (s *Server) AddCage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		logger := s.Logger.With("requestId", requestID)

		var req AddCageRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil && !errors.Is(err, io.EOF) {
			logger.Error("Error decoding request body", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if err := req.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cage, err := s.CageStore.Add(r.Context(), &app.Cage{
			Capacity: req.Capacity,
			Status:   req.Status,
		})
		if err != nil {
			logger.Error("Error adding cage", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		response := struct {
			Data *app.Cage `json:"data"`
		}{
			Data: cage,
		}

		if err = json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("Error marshalling response", "error", err)
			return
		}
	}
}

type UpdateCageRequest struct {
	Status app.CageStatus `json:"status"`
}

func (r UpdateCageRequest) Validate() error {
	return r.Status.Validate()
}

// ChangeCageStatus changes the status of a cage.
// PUT /cages/:id
func (s *Server) ChangeCageStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		logger := s.Logger.With("requestId", requestID)
		id := chi.URLParam(r, "id")

		var req UpdateCageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("Error decoding request body", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if err := req.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cage, err := s.CageStore.ChangeStatus(r.Context(), id, req.Status)
		if err != nil {
			switch err {
			case app.ErrNotFound:
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			case app.ErrConflict:
				http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			default:
				logger.Error("Error updating cage", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}

			return
		}

		w.Header().Set("Content-Type", "application/json")

		response := struct {
			Data *app.Cage `json:"data"`
		}{
			Data: cage,
		}

		if err = json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("Error marshalling response", "error", err)
			return
		}
	}
}

// DeleteCage deletes a cage.
// DELETE /cages/:id
func (s *Server) DeleteCage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		logger := s.Logger.With("requestId", requestID)
		id := chi.URLParam(r, "id")

		err := s.CageStore.Delete(r.Context(), id)
		if err != nil {
			switch err {
			case app.ErrNotFound:
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			case app.ErrConflict:
				http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			default:
				logger.Error("Error deleting cage", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
