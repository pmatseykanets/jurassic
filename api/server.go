package api

import (
	"context"
	"log/slog"

	"github.com/pmatseykanets/jurassic/app"
)

type CageStore interface {
	Add(ctx context.Context, cage *app.Cage) (*app.Cage, error)
	Get(ctx context.Context, id string) (*app.Cage, error)
	List(ctx context.Context, status app.CageStatus) ([]app.Cage, error)
	ChangeStatus(ctx context.Context, id string, status app.CageStatus) (*app.Cage, error)
	Delete(ctx context.Context, id string) error
}

type DinosaurStore interface {
	Add(ctx context.Context, dinosaur *app.Dinosaur) (*app.Dinosaur, error)
	List(ctx context.Context, cageID string, species app.DinosaurSpecies) ([]app.Dinosaur, error)
	Get(ctx context.Context, id string) (*app.Dinosaur, error)
	Move(ctx context.Context, id string, cageID string) (*app.Dinosaur, error)
	Delete(ctx context.Context, id string) error
}

type Server struct {
	Addr          string
	Logger        *slog.Logger
	CageStore     CageStore
	DinosaurStore DinosaurStore
}
