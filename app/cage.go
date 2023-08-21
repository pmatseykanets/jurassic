package app

import (
	"errors"
	"time"
)

// CageStatus represents a cage status.
type CageStatus string

const (
	CageStatusUnspecified CageStatus = ""
	CageStatusActive      CageStatus = "active"
	CageStatusDown        CageStatus = "down"
)

// Validate the cage staus value.
func (s CageStatus) Validate() error {
	switch s {
	case CageStatusActive, CageStatusDown:
		return nil
	default:
		return errors.New("invalid status")
	}
}

// IsUnspecified returns true if the cage status is empty.
func (s CageStatus) IsUnspecified() bool {
	return s == CageStatusUnspecified
}

type Cage struct {
	ID        string     `json:"id"`
	Status    CageStatus `json:"status"`
	Capacity  int        `json:"capacity"`
	Occupancy int        `json:"occupancy"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

var CageIDUnspecified = "" // A centinel value to denote any cage ID.
