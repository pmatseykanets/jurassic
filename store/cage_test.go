//go:build integration
// +build integration

package store

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/pmatseykanets/jurassic/app"
)

func TestCageStore(t *testing.T) {
	ctx := context.Background()
	store := CageStore{DB: testDB}

	t.Cleanup(func() {
		testDB.Exec("TRUNCATE TABLE cages CASCADE")
	})

	list, err := store.List(ctx, app.CageStatusUnspecified)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 0, len(list); want != got {
		t.Fatalf("Expected cages %d got %d", want, got)
	}

	// Add an active (powered) cage.
	cage1, err := store.Add(ctx, &app.Cage{
		Capacity: 2,
		Status:   app.CageStatusActive,
	})
	if err != nil {
		t.Fatal(err)
	}

	if cage1.ID == "" {
		t.Fatal("Expected ID got empty")
	}
	if cage1.CreatedAt.IsZero() {
		t.Fatal("Expected CreatedAt got empty")
	}
	if cage1.UpdatedAt.IsZero() {
		t.Fatal("Expected UpdatedAt got empty")
	}
	if want, got := app.CageStatusActive, cage1.Status; want != got {
		t.Fatalf("Expected Status %s got %s", want, got)
	}
	if want, got := 2, cage1.Capacity; want != got {
		t.Fatalf("Expected Capacity %d got %d", want, got)
	}
	if want, got := 0, cage1.Occupancy; want != got {
		t.Fatalf("Expected Occupancy %d got %d", want, got)
	}

	// Add a powered down cage.
	cage2, err := store.Add(ctx, &app.Cage{
		Capacity: 2,
		Status:   app.CageStatusDown,
	})
	if err != nil {
		t.Fatal(err)
	}

	if want, got := app.CageStatusDown, cage2.Status; want != got {
		t.Fatalf("Expected Status %s got %s", want, got)
	}

	list, err = store.List(ctx, app.CageStatusUnspecified)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 2, len(list); want != got {
		t.Fatalf("Expected cages %d got %d", want, got)
	}

	// List only active cages.
	list, err = store.List(ctx, app.CageStatusActive)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 1, len(list); want != got {
		t.Fatalf("Expected cages %d got %d", want, got)
	}
	if want, got := cage1.ID, list[0].ID; want != got {
		t.Fatalf("Expected ID %s got %s", want, got)
	}

	// List only powered down cages.
	list, err = store.List(ctx, app.CageStatusDown)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 1, len(list); want != got {
		t.Fatalf("Expected cages %d got %d", want, got)
	}
	if want, got := cage2.ID, list[0].ID; want != got {
		t.Fatalf("Expected ID %s got %s", want, got)
	}

	// Power down the active cage.
	cage1, err = store.ChangeStatus(ctx, cage1.ID, app.CageStatusDown)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := app.CageStatusDown, cage1.Status; want != got {
		t.Fatalf("Expected Status %s got %s", want, got)
	}

	// Power up the powered down cage.
	cage2, err = store.ChangeStatus(ctx, cage2.ID, app.CageStatusActive)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := app.CageStatusActive, cage2.Status; want != got {
		t.Fatalf("Expected Status %s got %s", want, got)
	}

	// Add a dinosaur.
	query := "INSERT INTO dinosaurs (name, species, cage_id) VALUES ($1, $2, $3)"
	_, err = testDB.Exec(query, "foo", app.DinosaurSpeciesTyrannosaurus, cage2.ID)
	if err != nil {
		t.Fatal(err)
	}

	// And make sure we can't power down an occupied cage.
	_, err = store.ChangeStatus(ctx, cage2.ID, app.CageStatusDown)
	if want, got := app.ErrConflict, err; want != got {
		t.Fatalf("Expected error %v got %v", want, got)
	}

	// Get the cage and see that occupancy is correctly reflected.
	cage2, err = store.Get(ctx, cage2.ID)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 1, cage2.Occupancy; want != got {
		t.Fatalf("Expected Occupancy %d got %d", want, got)
	}

	// Getting a non-existent cage should fail.
	_, err = store.Get(ctx, uuid.NewString())
	if want, got := app.ErrNotFound, err; want != got {
		t.Fatalf("Expected error %v got %v", want, got)
	}

	// Changing the status of a non-existent cage should fail.
	_, err = store.ChangeStatus(ctx, uuid.NewString(), app.CageStatusActive)
	if want, got := app.ErrNotFound, err; want != got {
		t.Fatalf("Expected error %v got %v", want, got)
	}

	// Delete a cage.
	err = store.Delete(ctx, cage1.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure we no longer see the deleted cage.
	list, err = store.List(ctx, app.CageStatusUnspecified)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 1, len(list); want != got {
		t.Fatalf("Expected cages %d got %d", want, got)
	}

	// Make sure an occupied cage can't be deleted.
	err = store.Delete(ctx, cage2.ID)
	if want, got := app.ErrConflict, err; want != got {
		t.Fatalf("Expected error %v got %v", want, got)
	}
}
