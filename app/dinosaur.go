package app

import (
	"errors"
	"time"
)

// DinosaurSpecies represents a dinosaur species.
type DinosaurSpecies string

// List of all supported dinosaur species.
const (
	DinosaurSpeciesUnspecified   DinosaurSpecies = ""
	DinosaurSpeciesTyrannosaurus DinosaurSpecies = "tyrannosaurus"
	DinosaurSpeciesVelociraptor  DinosaurSpecies = "velociraptor"
	DinosaurSpeciesSpinosaurus   DinosaurSpecies = "spinosaurus"
	DinosaurSpeciesMegalosaurus  DinosaurSpecies = "megalosaurus"
	DinosaurSpeciesBrachiosaurus DinosaurSpecies = "brachiosaurus"
	DinosaurSpeciesStegosaurus   DinosaurSpecies = "stegosaurus"
	DinosaurSpeciesAnkylosaurus  DinosaurSpecies = "ankylosaurus"
	DinosaurSpeciesTriceratops   DinosaurSpecies = "triceratops"
)

// DinosaurType represents a dinosaur type.
type DinosaurType string

// List of all supported dinosaur types.
const (
	DinosaurTypeCarnivore DinosaurType = "carnivore"
	DinosaurTypeHerbivore DinosaurType = "herbivore"
)

// Validate the dinosaur species value.
func (s DinosaurSpecies) Validate() error {
	switch s {
	case DinosaurSpeciesTyrannosaurus,
		DinosaurSpeciesVelociraptor,
		DinosaurSpeciesSpinosaurus,
		DinosaurSpeciesMegalosaurus,
		DinosaurSpeciesBrachiosaurus,
		DinosaurSpeciesStegosaurus,
		DinosaurSpeciesAnkylosaurus,
		DinosaurSpeciesTriceratops:
		return nil
	default:
		return errors.New("invalid species")
	}
}

// IsUnspecified returns true if the dinosaur species value is empty.
func (s DinosaurSpecies) IsUnspecified() bool {
	return s == DinosaurSpeciesUnspecified
}

// Type returns the dinosaur type.
func (s DinosaurSpecies) Type() DinosaurType {
	switch s {
	case DinosaurSpeciesTyrannosaurus,
		DinosaurSpeciesVelociraptor,
		DinosaurSpeciesSpinosaurus,
		DinosaurSpeciesMegalosaurus:
		return DinosaurTypeCarnivore
	case DinosaurSpeciesBrachiosaurus,
		DinosaurSpeciesStegosaurus,
		DinosaurSpeciesAnkylosaurus,
		DinosaurSpeciesTriceratops:
		return DinosaurTypeHerbivore
	default:
		panic("invalid species")
	}
}

// Dinosaur represents a dinosaur.
type Dinosaur struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Species   DinosaurSpecies `json:"species"`
	CageID    string          `json:"cageId"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}
