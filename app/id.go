package app

import (
	"errors"

	"github.com/google/uuid"
)

// IDUnspecified ia a centinel value to denote any ID.
var IDUnspecified = ""

const idLength = 36 // We only accept IDs in the form xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.

// ValidateID validates the ID.
func ValidateID(id string) error {
	invalidErr := errors.New("invalid ID")

	if len(id) != idLength {
		return invalidErr
	}

	_, err := uuid.Parse(id)
	if err != nil {
		return invalidErr
	}

	return nil
}
