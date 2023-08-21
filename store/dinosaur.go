package store

import (
	"context"
	"database/sql"

	"github.com/pmatseykanets/jurassic/app"
)

// DinosaurStore is a DB implementation of api.DinosaurStore store.
type DinosaurStore struct {
	DB *sql.DB
}

// Add a dinosaur to a cage.
func (s *DinosaurStore) Add(ctx context.Context, dinosaur *app.Dinosaur) (*app.Dinosaur, error) {
	return nil, nil
}

// List dinosaurs.
func (s *DinosaurStore) List(ctx context.Context, cageID string, species app.DinosaurSpecies) ([]app.Dinosaur, error) {
	var dinosaurs []app.Dinosaur
	return dinosaurs, nil
}

// Get a dinosaur by id.
func (s *DinosaurStore) Get(ctx context.Context, id string) (*app.Dinosaur, error) {
	return nil, nil
}

// Move a dinosaur to a different cage.
func (s *DinosaurStore) Move(ctx context.Context, id string, cageID string) (*app.Dinosaur, error) {
	return nil, nil
}

// Delete a dinosaur.
func (s *DinosaurStore) Delete(ctx context.Context, id string) error {
	return nil
}
