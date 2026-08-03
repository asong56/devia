package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBase64EncodeTextKnownVector(t *testing.T) {
	// "hello" -> "aGVsbG8=" is the textbook base64 example (RFC 4648).
	got := Base64EncodeText("hello", false)
	want := "aGVsbG8="
	if got != want {
		t.Errorf("Base64EncodeText(hello) = %s, want %s", got, want)
	}
}

func TestBase64RoundTrip(t *testing.T) {
	original := "Round trip: special chars !@#$%^&*()_+ and unicode 日本語"
	encoded := Base64EncodeText(original, false)
	decoded, err := Base64DecodeText(encoded, false)
	if err != nil {
		t.Fatal(err)
	}
	if decoded != original {
		t.Errorf("round trip failed: got %q, want %q", decoded, original)
	}
}

func TestBase64URLSafeRoundTrip(t *testing.T) {
	// Pick input whose standard base64 form is known to contain '+' or
	// '/', so the URL-safe alphabet actually gets exercised.
	original := string([]byte{0xfb, 0xff, 0xbf})
	stdEncoded := Base64EncodeText(original, false)
	if !strings.ContainsAny(stdEncoded, "+/") {
		t.Fatalf("test input %q does not exercise +/ in standard base64 (%s); fixture needs adjusting", original, stdEncoded)
	}

	urlEncoded := Base64EncodeText(original, true)
	if strings.ContainsAny(urlEncoded, "+/") {
		t.Errorf("url-safe encoding should not contain +/ , got %s", urlEncoded)
	}
	decoded, err := Base64DecodeText(urlEncoded, true)
	if err != nil {
		t.Fatal(err)
	}
	if decoded != original {
		t.Errorf("url-safe round trip failed: got %q, want %q", decoded, original)
	}
}

func TestBase64DecodeInvalidInput(t *testing.T) {
	_, err := Base64DecodeText("not valid base64!!!", false)
	if err == nil {
		t.Fatal("expected an error for invalid base64 input")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("invalid base64 should be a CodeInput error, got code %d", CodeOf(err))
	}
}

func TestBase64DecodeTrimsWhitespace(t *testing.T) {
	decoded, err := Base64DecodeText("  aGVsbG8=  \n", false)
	if err != nil {
		t.Fatal(err)
	}
	if decoded != "hello" {
		t.Errorf("decode should trim surrounding whitespace, got %q", decoded)
	}
}

func TestBase64EncodeFileWithDataURI(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "image.png")
	pngHeader := []byte("\x89PNG\r\n\x1a\nrest-of-file-does-not-matter-for-sniffing")
	if err := os.WriteFile(path, pngHeader, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := Base64EncodeFile(path, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "data:image/png;base64,") {
		t.Errorf("expected a PNG data URI prefix, got %s", out[:min(40, len(out))]) // min: builtin, Go 1.21+
	}
}

func TestBase64EncodeFileMissing(t *testing.T) {
	_, err := Base64EncodeFile(filepath.Join(t.TempDir(), "nope.bin"), false)
	if err == nil {
		t.Fatal("expected an error for a missing file")
	}
	if CodeOf(err) != CodeNotFound {
		t.Errorf("missing file should be CodeNotFound, got code %d", CodeOf(err))
	}
}

func TestBase64DecodeToFileStripsDataURI(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.bin")

	dataURI := "data:text/plain;base64," + Base64EncodeText("payload", false)
	if err := Base64DecodeToFile(dataURI, outPath); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "payload" {
		t.Errorf("decoded file content = %q, want %q", got, "payload")
	}
}
