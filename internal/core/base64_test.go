package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBase64EncodeText_RoundTrip(t *testing.T) {
	text := "hello world"
	encoded := Base64EncodeText(text, false)
	if encoded != "aGVsbG8gd29ybGQ=" {
		t.Errorf("Base64EncodeText(%q) = %s, want aGVsbG8gd29ybGQ=", text, encoded)
	}
	decoded, err := Base64DecodeText(encoded, false)
	if err != nil {
		t.Fatalf("unexpected error decoding: %v", err)
	}
	if decoded != text {
		t.Errorf("round trip = %q, want %q", decoded, text)
	}
}

func TestBase64_StandardVsURLSafeAlphabet(t *testing.T) {
	// "Kr~" is chosen specifically because its standard base64 form
	// contains a '+' (which becomes '-' in the URL-safe alphabet) — a
	// string like "hello" wouldn't actually distinguish the two codecs.
	text := "Kr~"
	std := Base64EncodeText(text, false)
	url := Base64EncodeText(text, true)
	if std != "S3J+" {
		t.Errorf("standard encode = %s, want S3J+", std)
	}
	if url != "S3J-" {
		t.Errorf("url-safe encode = %s, want S3J-", url)
	}
	if std == url {
		t.Error("standard and url-safe encodings should differ for input containing a '+' byte")
	}

	// Cross-decoding must fail: standard-encoded text is not valid
	// url-safe base64 once it contains '+', and vice versa.
	if _, err := Base64DecodeText(std, true); err == nil {
		t.Error("expected an error decoding standard-alphabet output with the url-safe codec")
	}
	if _, err := Base64DecodeText(url, false); err == nil {
		t.Error("expected an error decoding url-safe-alphabet output with the standard codec")
	}
}

func TestBase64DecodeText_InvalidInput(t *testing.T) {
	_, err := Base64DecodeText("not valid base64!!!", false)
	if err == nil {
		t.Fatal("expected an error for invalid base64 input, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d)", CodeOf(err), CodeInput)
	}
}

func TestBase64DecodeText_TrimsWhitespace(t *testing.T) {
	// Piping through `echo` (as documented in the README) adds a
	// trailing newline — decode must not choke on it.
	got, err := Base64DecodeText("aGVsbG8gd29ybGQ=\n", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestBase64EncodeFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.bin")
	content := []byte("arbitrary file content \x00\x01\x02")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	encoded, err := Base64EncodeFile(path, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	decoded, err := Base64DecodeBytes(encoded)
	if err != nil {
		t.Fatalf("unexpected error decoding: %v", err)
	}
	if string(decoded) != string(content) {
		t.Errorf("round trip through file did not preserve binary content")
	}
}

func TestBase64EncodeFile_NotFound(t *testing.T) {
	_, err := Base64EncodeFile(filepath.Join(t.TempDir(), "nope.bin"), false)
	if err == nil {
		t.Fatal("expected an error for a missing file, got nil")
	}
	if CodeOf(err) != CodeNotFound {
		t.Errorf("CodeOf(err) = %d, want CodeNotFound (%d)", CodeOf(err), CodeNotFound)
	}
}

func TestBase64EncodeFile_DataURI_SniffsPNG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "img.png")
	// Real PNG magic bytes followed by arbitrary content — sniffMIME
	// only looks at the header, so the rest doesn't need to be a valid
	// PNG for this test.
	content := append([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, []byte("restofdata")...)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	uri, err := Base64EncodeFile(path, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "data:image/png;base64,iVBORw0KGgpyZXN0b2ZkYXRh"
	if uri != want {
		t.Errorf("data URI = %s, want %s", uri, want)
	}
}

func TestBase64EncodeFile_DataURI_UnknownTypeFallsBackToOctetStream(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.bin")
	if err := os.WriteFile(path, []byte("not a known magic-byte format"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	uri, err := Base64EncodeFile(path, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(uri) < len("data:application/octet-stream;base64,") ||
		uri[:len("data:application/octet-stream;base64,")] != "data:application/octet-stream;base64," {
		t.Errorf("expected an application/octet-stream data URI for unrecognized content, got %s", uri)
	}
}

func TestBase64DecodeBytes_StripsDataURIPrefix(t *testing.T) {
	// Base64DecodeBytes should accept the data: URI it just produced —
	// this is the property --dry-run and --out both rely on: they must
	// agree on what "valid input" means, including data URIs.
	got, err := Base64DecodeBytes("data:image/png;base64,aGVsbG8=")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("got %q, want %q", string(got), "hello")
	}
}

func TestBase64DecodeToFile_WritesExactBytes(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.bin")
	if err := Base64DecodeToFile("aGVsbG8gd29ybGQ=", out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}
	if string(got) != "hello world" {
		t.Errorf("written content = %q, want %q", string(got), "hello world")
	}
}

func TestBase64DecodeToFile_InvalidInputDoesNotCreateFile(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "should-not-exist.bin")
	err := Base64DecodeToFile("!!!not base64!!!", out)
	if err == nil {
		t.Fatal("expected an error for invalid base64, got nil")
	}
	if _, statErr := os.Stat(out); statErr == nil {
		t.Error("Base64DecodeToFile must not create the output file when decoding fails")
	}
}
