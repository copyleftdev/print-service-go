package utils

import (
	"testing"

	"github.com/google/uuid"
)

func TestGenerateID(t *testing.T) {
	id := GenerateID()

	// Test that it's a valid UUID
	_, err := uuid.Parse(id)
	if err != nil {
		t.Errorf("GenerateID() returned invalid UUID: %v", err)
	}

	// Test that it's not empty
	if id == "" {
		t.Error("GenerateID() returned empty string")
	}

	// Test that multiple calls return different UUIDs
	id2 := GenerateID()
	if id == id2 {
		t.Error("GenerateID() returned same UUID on consecutive calls")
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid UUID", "550e8400-e29b-41d4-a716-446655440000", true},
		{"Valid UUID lowercase", "550e8400-e29b-41d4-a716-446655440000", true},
		{"Valid UUID uppercase", "550E8400-E29B-41D4-A716-446655440000", true},
		{"Invalid UUID - too short", "550e8400-e29b-41d4-a716", false},
		{"Invalid UUID - no hyphens", "550e8400e29b41d4a716446655440000", false},
		{"Invalid UUID - invalid chars", "550e8400-e29b-41d4-a716-44665544000g", false},
		{"Empty string", "", false},
		{"Random string", "not-a-uuid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateID(tt.input)
			if result != tt.expected {
				t.Errorf("ValidateID(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseID(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	parsed, err := ParseID(validUUID)
	if err != nil {
		t.Errorf("ParseID(%q) returned error: %v", validUUID, err)
	}

	if parsed.String() != validUUID {
		t.Errorf("ParseID(%q) = %q, want %q", validUUID, parsed.String(), validUUID)
	}

	// Test invalid UUID
	_, err = ParseID("invalid-uuid")
	if err == nil {
		t.Error("ParseID() should return error for invalid UUID")
	}
}

func TestGenerateShortID(t *testing.T) {
	id := GenerateShortID()

	// Test length (should be 8 characters)
	if len(id) != 8 {
		t.Errorf("GenerateShortID() returned ID of length %d, want 8", len(id))
	}

	// Test uniqueness
	id2 := GenerateShortID()
	if id == id2 {
		t.Error("GenerateShortID() returned same ID on consecutive calls")
	}
}

func TestUUIDPerformance(t *testing.T) {
	// Test that we can generate many UUIDs quickly
	const numUUIDs = 1000
	uuids := make(map[string]bool)

	for i := 0; i < numUUIDs; i++ {
		id := GenerateID()
		if uuids[id] {
			t.Errorf("Duplicate UUID generated: %s", id)
		}
		uuids[id] = true
	}

	if len(uuids) != numUUIDs {
		t.Errorf("Expected %d unique UUIDs, got %d", numUUIDs, len(uuids))
	}
}
