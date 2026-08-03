package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareChecksumMatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}

	actual, _, err := CompareChecksum("sha256", path, "placeholder")
	if err != nil {
		t.Fatal(err)
	}

	// Compare against the file's own real checksum, uppercased and with
	// surrounding whitespace, to exercise the case-insensitive / trimmed
	// comparison behaviour that CompareChecksum documents.
	_, match, err := CompareChecksum("sha256", path, "  "+strings.ToUpper(actual)+"\n")
	if err != nil {
		t.Fatal(err)
	}
	if !match {
		t.Error("expected checksum to match itself regardless of case/whitespace")
	}
}

func TestCompareChecksumMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, match, err := CompareChecksum("sha256", path, "0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	if match {
		t.Error("expected mismatch against an obviously wrong checksum")
	}
}

func TestCompareChecksumMissingFile(t *testing.T) {
	_, _, err := CompareChecksum("sha256", filepath.Join(t.TempDir(), "nope"), "x")
	if err == nil {
		t.Fatal("expected an error for a missing file")
	}
	if CodeOf(err) != CodeNotFound {
		t.Errorf("missing file should be CodeNotFound, got code %d", CodeOf(err))
	}
}
