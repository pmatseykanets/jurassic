//go:build integration
// +build integration

package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pmatseykanets/jurassic/app"
)

func TestDinosaurStore(t *testing.T) {
	setUpTestDB(t)
	t.Cleanup(func() {
		testDB.Exec("TRUNCATE TABLE cages CASCADE")
	})

	ctx := context.Background()
	cageStore := CageStore{DB: testDB}
	dinosaurStore := DinosaurStore{DB: testDB}

	// Add an active (powered) cage.
	cage1, err := cageStore.Add(ctx, &app.Cage{
		Capacity: 2,
		Status:   app.CageStatusActive,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Make sure listing cage dinosaurs comes back empty.
	list, err := dinosaurStore.List(ctx, app.IDUnspecified, app.DinosaurSpeciesUnspecified)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 0, len(list); want != got {
		t.Fatalf("Expected dinosaurs %d got %d", want, got)
	}

	// Add a dinosaur.
	dinosaur1, err := dinosaurStore.Add(ctx, &app.Dinosaur{
		Name:    "Tyrannosaurus Rex",
		Species: app.DinosaurSpeciesTyrannosaurus,
		CageID:  cage1.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	if dinosaur1.ID == "" {
		t.Error("Expected ID got empty")
	}
	if dinosaur1.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt got empty")
	}
	if dinosaur1.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt got empty")
	}
	if want, got := "Tyrannosaurus Rex", dinosaur1.Name; want != got {
		t.Errorf("Expected Name %s got %s", want, got)
	}
	if want, got := app.DinosaurSpeciesTyrannosaurus, dinosaur1.Species; want != got {
		t.Errorf("Expected Species %s got %s", want, got)
	}
	if want, got := cage1.ID, dinosaur1.CageID; want != got {
		t.Errorf("Expected CageID %s got %s", want, got)
	}

	// Make sure listing cage dinosaurs comes back with one dinosaur.
	list, err = dinosaurStore.List(ctx, cage1.ID, app.DinosaurSpeciesUnspecified)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 1, len(list); want != got {
		t.Fatalf("Expected dinosaurs %d got %d", want, got)
	}

	// Make sure we can't add a carnivore of different species to the cage.
	_, err = dinosaurStore.Add(ctx, &app.Dinosaur{
		Name:    "Velociraptor",
		Species: app.DinosaurSpeciesVelociraptor,
		CageID:  cage1.ID,
	})
	if err == nil {
		t.Fatal("Expected error got nil")
	}

	// Make sure we can't add a herbivore to the cage with carnivores.
	_, err = dinosaurStore.Add(ctx, &app.Dinosaur{
		Name:    "Brachiosaurus",
		Species: app.DinosaurSpeciesBrachiosaurus,
		CageID:  cage1.ID,
	})
	if err == nil {
		t.Fatal("Expected error got nil")
	}

	// Adding a carnivore of the same species should work though.
	dinosaur2, err := dinosaurStore.Add(ctx, &app.Dinosaur{
		Name:    "Tyrannosaurus Pex",
		Species: app.DinosaurSpeciesTyrannosaurus,
		CageID:  cage1.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Add another cage.
	cage2, err := cageStore.Add(ctx, &app.Cage{
		Capacity: 2,
		Status:   app.CageStatusActive,
	})
	if err != nil {
		t.Fatal(err)
	}

	// And move the dinosaur2 to the new cage.
	dinosaur2, err = dinosaurStore.Move(ctx, dinosaur2.ID, cage2.ID)
	if err != nil {
		t.Fatal(err)
	}

	// List all dinosaurs.
	list, err = dinosaurStore.List(ctx, app.IDUnspecified, app.DinosaurSpeciesUnspecified)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 2, len(list); want != got {
		t.Fatalf("Expected dinosaurs %d got %d", want, got)
	}

	// Adding a dinosaur to a non-existent cage should fail.
	_, err = dinosaurStore.Add(ctx, &app.Dinosaur{
		Name:    "Tyrannosaurus",
		Species: app.DinosaurSpeciesTyrannosaurus,
		CageID:  uuid.NewString(),
	})
	if err == nil {
		t.Fatal("Expected error got nil")
	}

	// Adding a dinosaur to a powered down cage should fail.
	cage3, err := cageStore.Add(ctx, &app.Cage{
		Capacity: 2,
		Status:   app.CageStatusDown,
	})

	_, err = dinosaurStore.Add(ctx, &app.Dinosaur{
		Name:    "Tyrannosaurus",
		Species: app.DinosaurSpeciesTyrannosaurus,
		CageID:  cage3.ID,
	})
	if err == nil {
		t.Fatal("Expected error got nil")
	}

	// Delete a dinosaur.
	err = dinosaurStore.Delete(ctx, dinosaur2.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure we no longer see the deleted dinosaur.
	list, err = dinosaurStore.List(ctx, app.IDUnspecified, app.DinosaurSpeciesUnspecified)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 1, len(list); want != got {
		t.Fatalf("Expected dinosaurs %d got %d", want, got)
	}

	_, err = dinosaurStore.Get(ctx, dinosaur2.ID)
	if want, got := app.ErrNotFound, err; want != got {
		t.Fatalf("Expected error %s got %s", want, got)
	}
}
