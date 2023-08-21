//go:build unit
// +build unit

package app

import "testing"

func TestDinosaurSpeciesValidate(t *testing.T) {
	tests := []struct {
		species DinosaurSpecies
		valid   bool
	}{
		{DinosaurSpeciesTyrannosaurus, true},
		{DinosaurSpeciesVelociraptor, true},
		{DinosaurSpeciesSpinosaurus, true},
		{DinosaurSpeciesMegalosaurus, true},
		{DinosaurSpeciesBrachiosaurus, true},
		{DinosaurSpeciesStegosaurus, true},
		{DinosaurSpeciesAnkylosaurus, true},
		{DinosaurSpeciesTriceratops, true},
		{DinosaurSpeciesUnspecified, false},
		{DinosaurSpecies("foo"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.species), func(t *testing.T) {
			err := tt.species.Validate()
			if want, got := tt.valid, err == nil; want != got {
				t.Errorf("Expected %t got %t", want, got)
			}
		})
	}
}

func TestDinosaurSpeciesType(t *testing.T) {
	tests := []struct {
		species     DinosaurSpecies
		speciesType DinosaurType
	}{
		{DinosaurSpeciesTyrannosaurus, DinosaurTypeCarnivore},
		{DinosaurSpeciesVelociraptor, DinosaurTypeCarnivore},
		{DinosaurSpeciesSpinosaurus, DinosaurTypeCarnivore},
		{DinosaurSpeciesMegalosaurus, DinosaurTypeCarnivore},
		{DinosaurSpeciesBrachiosaurus, DinosaurTypeHerbivore},
		{DinosaurSpeciesStegosaurus, DinosaurTypeHerbivore},
		{DinosaurSpeciesAnkylosaurus, DinosaurTypeHerbivore},
		{DinosaurSpeciesTriceratops, DinosaurTypeHerbivore},
	}

	for _, tt := range tests {
		t.Run(string(tt.species), func(t *testing.T) {
			if want, got := tt.speciesType, tt.species.Type(); want != got {
				t.Errorf("Expected %s got %s", want, got)
			}
		})
	}
}

func TestDinosaurSpeciesTypePanicsOnInvalidSpecies(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()

	_ = DinosaurSpecies("foo").Type()
}
