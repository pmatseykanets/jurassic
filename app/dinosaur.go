package app

import (
	"errors"
	"time"
)

type DinosaurSpecies string

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

type DinosaurType string

const (
	DinosaurTypeCarnivore DinosaurType = "carnivore"
	DinosaurTypeHerbivore DinosaurType = "herbivore"
)

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

func (s DinosaurSpecies) IsUnspecified() bool {
	return s == DinosaurSpeciesUnspecified
}

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

type Dinosaur struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Species   DinosaurSpecies `json:"species"`
	CageID    string          `json:"cageId"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}
