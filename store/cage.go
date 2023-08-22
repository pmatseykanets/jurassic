package store

import (
	"context"
	"database/sql"

	"github.com/pmatseykanets/jurassic/app"
)

// CageStore is a DB implementation of api.CageStore.
type CageStore struct {
	DB *sql.DB
}

// Add a new cage.
func (s *CageStore) Add(ctx context.Context, cage *app.Cage) (*app.Cage, error) {
	var c app.Cage
	query := `
	INSERT INTO cages (capacity, status) VALUES ($1, $2) 
	RETURNING id, capacity, status, created_at, updated_at`
	err := s.DB.QueryRowContext(ctx, query, cage.Capacity, cage.Status).Scan(
		&c.ID,
		&c.Capacity,
		&c.Status,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// Get a cage by id.
func (s *CageStore) Get(ctx context.Context, id string) (*app.Cage, error) {
	return getCage(ctx, s.DB, id)
}

// List cages.
func (s *CageStore) List(ctx context.Context, status app.CageStatus) ([]app.Cage, error) {
	var cages []app.Cage
	query := `
	SELECT c.id, c.capacity, c.status, c.created_at, c.updated_at, COUNT(d.id)
	  FROM cages c 
	  LEFT JOIN dinosaurs d ON d.cage_id = c.id`

	var args []any
	if !status.IsUnspecified() {
		query += " WHERE status = $1"
		args = append(args, status)
	}

	query += " GROUP BY c.id, c.capacity, c.status, c.created_at, c.updated_at"

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cage app.Cage
		if err := rows.Scan(
			&cage.ID,
			&cage.Capacity,
			&cage.Status,
			&cage.CreatedAt,
			&cage.UpdatedAt,
			&cage.Occupancy,
		); err != nil {
			return nil, err
		}

		cages = append(cages, cage)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cages, nil
}

// Change status of a cage.
func (s *CageStore) ChangeStatus(ctx context.Context, id string, status app.CageStatus) (*app.Cage, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // nolint:errcheck

	cage, err := getCage(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	if status == cage.Status {
		return cage, nil // Nothing to do.
	}

	if status == app.CageStatusDown && cage.Occupancy > 0 {
		return nil, app.ErrConflict
	}

	query := `
	UPDATE cages
	   SET status = $1, updated_at = NOW()
	 WHERE id = $2
	RETURNING status, updated_at`

	err = tx.QueryRowContext(ctx, query, status, id).Scan(
		&cage.Status,
		&cage.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return cage, nil
}

// Delete a cage.
func (s *CageStore) Delete(ctx context.Context, id string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // nolint:errcheck

	cage, err := getCage(ctx, tx, id)
	if err != nil {
		return err
	}

	if cage.Occupancy > 0 {
		return app.ErrConflict
	}

	query := `
	DELETE FROM cages
	 WHERE id = $1`
	res, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return app.ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// getCage returns a cage by id including its occupancy.
func getCage(ctx context.Context, q queryable, id string) (*app.Cage, error) {
	var cage app.Cage
	query := `
	SELECT c.id, c.capacity, c.status, c.created_at, c.updated_at, COUNT(d.id)
	  FROM cages c 
	  LEFT JOIN dinosaurs d ON d.cage_id = c.id
	 WHERE c.id = $1
	 GROUP BY c.id, c.capacity, c.status, c.created_at, c.updated_at`

	err := q.QueryRowContext(ctx, query, id).Scan(
		&cage.ID,
		&cage.Capacity,
		&cage.Status,
		&cage.CreatedAt,
		&cage.UpdatedAt,
		&cage.Occupancy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, app.ErrNotFound
		}

		return nil, err
	}

	return &cage, nil
}

// checkCageCompatibility checks if a dinosaur can be added or moved to a cage.
func checkCageCompatibility(
	ctx context.Context,
	q queryable,
	id string,
	species app.DinosaurSpecies,
) error {
	// To satisfy the species compatibility requirements we just need to know
	// the species of any of the occupying dinosaurs- thus the use of MIN(d.species).
	query := `
	SELECT c.capacity, c.status, COUNT(d.id), COALESCE(MIN(d.species), '')
	  FROM cages c
	  LEFT JOIN dinosaurs d ON d.cage_id = c.id
	 WHERE c.id = $1
	 GROUP BY c.capacity, c.status`

	var (
		capacity    int
		status      app.CageStatus
		occupancy   int
		cageSpecies app.DinosaurSpecies
	)
	err := q.QueryRowContext(ctx, query, id).Scan(
		&capacity,
		&status,
		&occupancy,
		&cageSpecies,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return app.ErrNotFound
		}

		return err
	}
	if status == app.CageStatusDown {
		return app.ErrCagePoweredDown
	}

	if occupancy >= capacity {
		return app.ErrCapacityExceeded
	}

	// If a cage is occupied we need to make sure species are compatible.
	if occupancy > 0 {
		speciesType := species.Type()
		// All dinosaurs in the cage must be of the same species type.
		if speciesType != cageSpecies.Type() {
			return app.ErrSpeciesMismatch
		}

		// Carnivores can only be in a cage with the same species.
		if speciesType == app.DinosaurTypeCarnivore && species != cageSpecies {
			return app.ErrSpeciesMismatch
		}
	}

	return nil
}
