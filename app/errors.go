package app

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrConflict         = errors.New("conflict")
	ErrCapacityExceeded = errors.New("capacity exceeded")
	ErrCagePoweredDown  = errors.New("cage powered down")
	ErrSpeciesMismatch  = errors.New("species mismatch")
)
