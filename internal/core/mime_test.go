package core

import "testing"

func TestSniffMIME(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want string
	}{
		{"png", []byte("\x89PNG\r\n\x1a\nrestofdata"), "image/png"},
		{"jpeg", []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}, "image/jpeg"},
		{"gif87a", []byte("GIF87a...."), "image/gif"},
		{"gif89a", []byte("GIF89a...."), "image/gif"},
		{"webp", append([]byte("RIFF"), append([]byte{0x00, 0x00, 0x00, 0x00}, []byte("WEBPrest")...)...), "image/webp"},
		{"bmp", []byte("BM...."), "image/bmp"},
		{"svg", []byte("<svg xmlns=\"...\">"), "image/svg+xml"},
		{"xml-prolog svg", []byte("<?xml version=\"1.0\"?><svg/>"), "image/svg+xml"},
		{"pdf", []byte("%PDF-1.4 rest"), "application/pdf"},
		{"unknown", []byte("just some random text"), "application/octet-stream"},
		{"empty", []byte{}, "application/octet-stream"},
	}
	for _, c := range cases {
		got := sniffMIME(c.data)
		if got != c.want {
			t.Errorf("%s: sniffMIME(...) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestSniffMIME_TooShortForPrefixDoesNotPanic(t *testing.T) {
	// Every branch in sniffMIME must len-check before indexing —
	// feeding truncated/tiny inputs (shorter than any real magic-byte
	// prefix) must degrade to the default case, never panic on an
	// out-of-range index.
	for _, data := range [][]byte{{}, {0x89}, {0xFF, 0xD8}, []byte("RIF"), []byte("B")} {
		got := sniffMIME(data)
		if got != "application/octet-stream" {
			t.Errorf("sniffMIME(%v) = %q, want application/octet-stream for undersized input", data, got)
		}
	}
}

func TestSniffMIME_JPEGRequiresAllThreeMagicBytes(t *testing.T) {
	// JPEG's check is a 3-byte match (FF D8 FF), not just the 2-byte
	// SOI marker (FF D8) — the third byte matters because FF D8
	// followed by something other than FF is not actually a valid
	// JPEG start-of-frame.
	notJPEG := sniffMIME([]byte{0xFF, 0xD8, 0x00})
	if notJPEG == "image/jpeg" {
		t.Error("FF D8 00 should not be sniffed as JPEG — the third byte must also be FF")
	}
}

func TestSniffMIME_WebPRequiresRIFFAndWEBPMarker(t *testing.T) {
	// A plain RIFF container that isn't WEBP (e.g. a WAV file, which
	// also starts with RIFF....WAVE) must not be misidentified.
	wav := append([]byte("RIFF"), append([]byte{0, 0, 0, 0}, []byte("WAVEfmt ")...)...)
	got := sniffMIME(wav)
	if got == "image/webp" {
		t.Error("a RIFF/WAVE file should not be sniffed as image/webp")
	}
}

func TestHasPrefix(t *testing.T) {
	if !hasPrefix([]byte("hello world"), "hello") {
		t.Error("hasPrefix should match a genuine prefix")
	}
	if hasPrefix([]byte("hello"), "hello world") {
		t.Error("hasPrefix should not match when data is shorter than prefix")
	}
	if hasPrefix([]byte("world hello"), "hello") {
		t.Error("hasPrefix should not match a substring that isn't at position 0")
	}
	if !hasPrefix([]byte("x"), "") {
		t.Error("an empty prefix should always match")
	}
}
