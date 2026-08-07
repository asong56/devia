package core

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return path
}

func TestCompareChecksum_Match(t *testing.T) {
	path := writeTempFile(t, "hello world")
	sum, _ := HashText("sha256", "hello world", "", false)

	actual, match, err := CompareChecksum("sha256", path, sum)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Error("expected match=true for an identical checksum")
	}
	if actual != sum {
		t.Errorf("actual = %s, want %s", actual, sum)
	}
}

func TestCompareChecksum_Mismatch(t *testing.T) {
	path := writeTempFile(t, "hello world")
	_, match, err := CompareChecksum("sha256", path, "0000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatalf("a checksum mismatch is not itself an error: %v", err)
	}
	if match {
		t.Error("expected match=false for a deliberately wrong checksum")
	}
}

func TestCompareChecksum_CaseInsensitive(t *testing.T) {
	path := writeTempFile(t, "hello world")
	sum, _ := HashText("sha256", "hello world", "", false)
	upper := ""
	for _, r := range sum {
		if r >= 'a' && r <= 'f' {
			upper += string(r - 32)
		} else {
			upper += string(r)
		}
	}
	_, match, err := CompareChecksum("sha256", path, upper)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Error("expected an uppercase hex checksum to still match (comparison should be case-insensitive)")
	}
}

func TestCompareChecksum_TrimsWhitespace(t *testing.T) {
	// Users routinely paste a checksum with a trailing newline copied
	// from another tool's output — that must not cause a false mismatch.
	path := writeTempFile(t, "hello world")
	sum, _ := HashText("sha256", "hello world", "", false)
	_, match, err := CompareChecksum("sha256", path, "  "+sum+"\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Error("expected surrounding whitespace in the expected checksum to be trimmed before comparing")
	}
}

func TestCompareChecksum_FileNotFound(t *testing.T) {
	_, _, err := CompareChecksum("sha256", filepath.Join(t.TempDir(), "nope.bin"), "anything")
	if err == nil {
		t.Fatal("expected an error for a missing file, got nil")
	}
	if CodeOf(err) != CodeNotFound {
		t.Errorf("CodeOf(err) = %d, want CodeNotFound (%d)", CodeOf(err), CodeNotFound)
	}
}
