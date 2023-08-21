package store

import (
	"context"
	"database/sql"

	"github.com/pmatseykanets/jurassic/app"
)

// CageStore is a DB implementation of api.CageStore store.
type CageStore struct {
	DB *sql.DB
}

// Add a new cage.
func (s *CageStore) Add(ctx context.Context, cage *app.Cage) (*app.Cage, error) {
	return nil, nil
}

// Get a cage by id.
func (s *CageStore) Get(ctx context.Context, id string) (*app.Cage, error) {
	return nil, nil
}

// List cages.
func (s *CageStore) List(ctx context.Context, status app.CageStatus) ([]app.Cage, error) {
	var cages []app.Cage
	return cages, nil
}

// Change status of a cage.
func (s *CageStore) ChangeStatus(ctx context.Context, id string, status app.CageStatus) (*app.Cage, error) {
	return nil, nil
}

// Delete a cage.
func (s *CageStore) Delete(ctx context.Context, id string) error {
	return nil
}
