package app

import (
	"errors"

	"github.com/google/uuid"
)

var IDUnspecified = "" // A centinel value to denote any ID.

const idLength = 36 // We only accept IDs in the form xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.

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
