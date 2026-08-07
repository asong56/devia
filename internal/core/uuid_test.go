package core

import (
	"regexp"
	"strings"
	"testing"
)

var uuidv4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestNewUUIDv4_Format(t *testing.T) {
	id, err := NewUUIDv4()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !uuidv4Pattern.MatchString(id) {
		t.Errorf("NewUUIDv4() = %q, does not match the RFC 4122 v4 pattern (8-4-4-4-12, version nibble 4, variant nibble 8/9/a/b)", id)
	}
}

func TestNewUUIDv4_Uniqueness(t *testing.T) {
	// Not a proof of randomness, but generating a decent batch and
	// checking for zero collisions is a real smoke test that
	// crypto/rand is actually wired up (a broken/zeroed source would
	// produce identical or trivially patterned IDs).
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id, err := NewUUIDv4()
		if err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}
		if seen[id] {
			t.Fatalf("duplicate UUID generated: %s", id)
		}
		seen[id] = true
	}
}

func TestNewUUIDs_Count(t *testing.T) {
	ids, err := NewUUIDs(5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 5 {
		t.Errorf("len(ids) = %d, want 5", len(ids))
	}
}

func TestNewUUIDs_ZeroOrNegativeCountDefaultsToOne(t *testing.T) {
	for _, n := range []int{0, -1, -100} {
		ids, err := NewUUIDs(n, false)
		if err != nil {
			t.Fatalf("NewUUIDs(%d): unexpected error: %v", n, err)
		}
		if len(ids) != 1 {
			t.Errorf("NewUUIDs(%d): len = %d, want 1 (should clamp up, not return zero or panic on a negative make())", n, len(ids))
		}
	}
}

func TestNewUUIDs_Upper(t *testing.T) {
	ids, err := NewUUIDs(3, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, id := range ids {
		if id != strings.ToUpper(id) {
			t.Errorf("expected uppercase UUID, got %q", id)
		}
	}
}

func TestNewUUIDs_AllUnique(t *testing.T) {
	ids, err := NewUUIDs(50, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			t.Fatalf("NewUUIDs produced a duplicate: %s", id)
		}
		seen[id] = true
	}
}
