package utils

import "github.com/google/uuid"

// ValidateUUID checks if the provided ID is a valid UUID.
func ValidateUUID(id string) error {
	_, err := uuid.Parse(id)
	return err
}
