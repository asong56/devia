package core

import (
	"regexp"
	"testing"
)

var uuidV4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestNewUUIDv4Format(t *testing.T) {
	for i := 0; i < 20; i++ {
		id, err := NewUUIDv4()
		if err != nil {
			t.Fatal(err)
		}
		if !uuidV4Pattern.MatchString(id) {
			t.Fatalf("UUID %q does not match the expected v4 format (8-4-4-4-12, version nibble 4, variant nibble 8-b)", id)
		}
	}
}

func TestNewUUIDv4Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := NewUUIDv4()
		if err != nil {
			t.Fatal(err)
		}
		if seen[id] {
			t.Fatalf("duplicate UUID generated: %s", id)
		}
		seen[id] = true
	}
}

func TestNewUUIDsCount(t *testing.T) {
	ids, err := NewUUIDs(5, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 5 {
		t.Fatalf("expected 5 UUIDs, got %d", len(ids))
	}
	for _, id := range ids {
		if !uuidV4Pattern.MatchString(id) {
			t.Errorf("UUID %q does not match expected format", id)
		}
	}
}

func TestNewUUIDsZeroOrNegativeDefaultsToOne(t *testing.T) {
	ids, err := NewUUIDs(0, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 {
		t.Errorf("count 0 should default to 1 UUID, got %d", len(ids))
	}

	ids, err = NewUUIDs(-3, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 {
		t.Errorf("negative count should default to 1 UUID, got %d", len(ids))
	}
}

func TestNewUUIDsUppercase(t *testing.T) {
	ids, err := NewUUIDs(3, true)
	if err != nil {
		t.Fatal(err)
	}
	upperPattern := regexp.MustCompile(`^[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}$`)
	for _, id := range ids {
		if !upperPattern.MatchString(id) {
			t.Errorf("expected uppercase UUID, got %s", id)
		}
	}
}
