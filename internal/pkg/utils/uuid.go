package utils

import (
	"github.com/google/uuid"
)

// GenerateID generates a new UUID string
func GenerateID() string {
	return uuid.New().String()
}

// GenerateShortID generates a shorter UUID (first 8 characters)
func GenerateShortID() string {
	return uuid.New().String()[:8]
}

// ValidateID validates if a string is a valid UUID
func ValidateID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// ParseID parses a UUID string and returns the UUID object
func ParseID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}
