package core

import "strings"

// CompareChecksum hashes the file at path and compares it against an
// expected value (case-insensitive, whitespace-trimmed on both sides,
// since users routinely paste hashes with trailing newlines or in the
// wrong case).
func CompareChecksum(algo, path, expected string) (actual string, match bool, err error) {
	actual, err = HashFile(algo, path, false)
	if err != nil {
		return "", false, err
	}
	match = strings.EqualFold(strings.TrimSpace(actual), strings.TrimSpace(expected))
	return actual, match, nil
}
