package core

import (
	"crypto/rand"
	"fmt"
	"strings"
)

// NewUUIDv4 generates a cryptographically random RFC 4122 version 4
// UUID (crypto/rand, not math/rand — this identifies things, so it
// needs to actually be unpredictable).
func NewUUIDv4() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // RFC 4122 variant
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// NewUUIDs generates n UUIDv4 values, optionally uppercased.
func NewUUIDs(n int, upper bool) ([]string, error) {
	if n < 1 {
		n = 1
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		id, err := NewUUIDv4()
		if err != nil {
			return nil, err
		}
		if upper {
			id = strings.ToUpper(id)
		}
		out = append(out, id)
	}
	return out, nil
}
