package core

import (
	"strings"
	"testing"
)

func TestEscapeUnescapeJSON(t *testing.T) {
	original := `hello "world"` + "\n\ttabbed"
	escaped, err := EscapeJSON(original)
	if err != nil {
		t.Fatal(err)
	}
	if escaped == original {
		t.Error("escaping should have changed the string")
	}
	unescaped, err := UnescapeJSON(escaped)
	if err != nil {
		t.Fatal(err)
	}
	if unescaped != original {
		t.Errorf("round trip failed: got %q, want %q", unescaped, original)
	}
}

func TestEscapeJSONHasNoSurroundingQuotes(t *testing.T) {
	got, err := EscapeJSON("x")
	if err != nil {
		t.Fatal(err)
	}
	if got != "x" {
		t.Errorf("EscapeJSON(%q) = %q, want %q (no added quotes for plain text)", "x", got, "x")
	}
}

func TestUnescapeJSONInvalid(t *testing.T) {
	_, err := UnescapeJSON(`unterminated \`)
	if err == nil {
		t.Fatal("expected an error for an invalid JSON-escaped string")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("expected CodeInput, got %d", CodeOf(err))
	}
}

func TestEscapeURLQuery(t *testing.T) {
	// url.QueryEscape encodes space as '+'.
	got := EscapeURL("a b", false)
	if got != "a+b" {
		t.Errorf("EscapeURL query = %q, want %q", got, "a+b")
	}
	back, err := UnescapeURL(got, false)
	if err != nil {
		t.Fatal(err)
	}
	if back != "a b" {
		t.Errorf("UnescapeURL query = %q, want %q", back, "a b")
	}
}

func TestEscapeURLPath(t *testing.T) {
	// url.PathEscape encodes space as '%20', not '+'.
	got := EscapeURL("a b", true)
	if got != "a%20b" {
		t.Errorf("EscapeURL path = %q, want %q", got, "a%20b")
	}
	back, err := UnescapeURL(got, true)
	if err != nil {
		t.Fatal(err)
	}
	if back != "a b" {
		t.Errorf("UnescapeURL path = %q, want %q", back, "a b")
	}
}

func TestUnescapeURLInvalid(t *testing.T) {
	_, err := UnescapeURL("%zz", false)
	if err == nil {
		t.Fatal("expected an error for an invalid percent-encoding")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("expected CodeInput, got %d", CodeOf(err))
	}
}

func TestEscapeUnescapeHTML(t *testing.T) {
	original := `<a href="x">it's & it's</a>`
	escaped := EscapeHTML(original)
	if escaped == original {
		t.Error("escaping should have changed the string")
	}
	// The angle brackets and ampersand must not survive raw.
	for _, ch := range []string{"<a", "</a>", "it's &"} {
		if strings.Contains(escaped, ch) {
			t.Errorf("escaped HTML still contains raw fragment %q: %s", ch, escaped)
		}
	}
	back := UnescapeHTML(escaped)
	if back != original {
		t.Errorf("round trip failed: got %q, want %q", back, original)
	}
}

func TestEscapeUnicodeASCIIPassesThrough(t *testing.T) {
	got := EscapeUnicode("plain ascii 123")
	if got != "plain ascii 123" {
		t.Errorf("ASCII text should pass through unchanged, got %q", got)
	}
}

func TestEscapeUnescapeUnicodeBMP(t *testing.T) {
	original := "café"
	escaped := EscapeUnicode(original)
	if escaped == original {
		t.Error("escaping should have changed a string with non-ASCII characters")
	}
	back, err := UnescapeUnicode(escaped)
	if err != nil {
		t.Fatal(err)
	}
	if back != original {
		t.Errorf("round trip failed: got %q, want %q", back, original)
	}
}

func TestEscapeUnescapeUnicodeSurrogatePair(t *testing.T) {
	// U+1F600 (grinning face emoji) is outside the Basic Multilingual
	// Plane and must round-trip through a UTF-16 surrogate pair.
	original := "\U0001F600"
	escaped := EscapeUnicode(original)
	want := `\ud83d\ude00`
	if escaped != want {
		t.Errorf("EscapeUnicode(emoji) = %s, want %s", escaped, want)
	}
	back, err := UnescapeUnicode(escaped)
	if err != nil {
		t.Fatal(err)
	}
	if back != original {
		t.Errorf("surrogate pair round trip failed: got %q, want %q", back, original)
	}
}

func TestUnescapeUnicodeTruncated(t *testing.T) {
	_, err := UnescapeUnicode(`\u12`)
	if err == nil {
		t.Fatal("expected an error for a truncated \\u escape")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("expected CodeInput, got %d", CodeOf(err))
	}
}
