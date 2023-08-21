package store

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/pmatseykanets/jurassic/app"
)

// DinosaurStore is a DB implementation of api.DinosaurStore store.
type DinosaurStore struct {
	DB *sql.DB
}

// Add a dinosaur to a cage.
func (s *DinosaurStore) Add(ctx context.Context, dinosaur *app.Dinosaur) (*app.Dinosaur, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // nolint:errcheck

	err = checkCageCompatibility(ctx, tx, dinosaur.CageID, dinosaur.Species)
	if err != nil {
		return nil, err
	}

	var added app.Dinosaur
	query := `
	INSERT INTO dinosaurs (name, species, cage_id)
	VALUES ($1, $2, $3) 
	RETURNING id, name, species, cage_id, created_at, updated_at`
	err = s.DB.QueryRowContext(ctx, query, dinosaur.Name, dinosaur.Species, dinosaur.CageID).Scan(
		&added.ID,
		&added.Name,
		&added.Species,
		&added.CageID,
		&added.CreatedAt,
		&added.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &added, nil
}

// List dinosaurs.
func (s *DinosaurStore) List(ctx context.Context, cageID string, species app.DinosaurSpecies) ([]app.Dinosaur, error) {
	var dinosaurs []app.Dinosaur
	query := `
	SELECT id, name, species, cage_id, created_at, updated_at
	  FROM dinosaurs`

	var (
		where []string
		args  []any
	)
	if cageID != "" {
		where = append(where, "cage_id = ?")
		args = append(args, cageID)
	}
	if species != "" {
		where = append(where, "species = ?")
		args = append(args, species)
	}

	for i, predicate := range where {
		if i == 0 {
			query += " WHERE "
		} else {
			query += " AND "
		}

		query += strings.Replace(predicate, "?", "$"+strconv.Itoa(i+1), 1)
	}

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var dinosaur app.Dinosaur
		if err := rows.Scan(
			&dinosaur.ID,
			&dinosaur.Name,
			&dinosaur.Species,
			&dinosaur.CageID,
			&dinosaur.CreatedAt,
			&dinosaur.UpdatedAt,
		); err != nil {
			return nil, err
		}

		dinosaurs = append(dinosaurs, dinosaur)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dinosaurs, nil
}

// Get a dinosaur by id.
func (s *DinosaurStore) Get(ctx context.Context, id string) (*app.Dinosaur, error) {
	return getDinosaur(ctx, s.DB, id)
}

// Move a dinosaur to a different cage.
func (s *DinosaurStore) Move(ctx context.Context, id string, cageID string) (*app.Dinosaur, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // nolint:errcheck

	dinosaur, err := getDinosaur(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	err = checkCageCompatibility(ctx, tx, cageID, dinosaur.Species)
	if err != nil {
		return nil, err
	}

	query := `
	UPDATE dinosaurs
	   SET cage_id = $1, updated_at = NOW()
	 WHERE id = $2
	RETURNING cage_id, updated_at`
	err = s.DB.QueryRowContext(ctx, query, cageID, id).Scan(
		&dinosaur.CageID,
		&dinosaur.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return dinosaur, nil
}

// Delete a dinosaur.
func (s *DinosaurStore) Delete(ctx context.Context, id string) error {
	query := `
	DELETE FROM dinosaurs
	 WHERE id = $1`
	res, err := s.DB.ExecContext(ctx, query, id)
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

	return nil
}

func getDinosaur(ctx context.Context, q queryable, id string) (*app.Dinosaur, error) {
	var dinosaur app.Dinosaur
	query := `
	SELECT id, name, species, cage_id, created_at, updated_at
	  FROM dinosaurs
	 WHERE id = $1`

	err := q.QueryRowContext(ctx, query, id).Scan(
		&dinosaur.ID,
		&dinosaur.Name,
		&dinosaur.Species,
		&dinosaur.CageID,
		&dinosaur.CreatedAt,
		&dinosaur.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, app.ErrNotFound
		}

		return nil, err
	}

	return &dinosaur, nil
}
