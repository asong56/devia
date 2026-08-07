package core

import (
	"strings"
	"testing"
)

func TestEscapeJSON_Basic(t *testing.T) {
	got, err := EscapeJSON(`hello "world"` + "\n\ttab")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `hello \"world\"\n\ttab`
	if got != want {
		t.Errorf("EscapeJSON = %q, want %q", got, want)
	}
}

func TestEscapeJSON_NoSurroundingQuotes(t *testing.T) {
	// EscapeJSON is documented to strip the surrounding quotes
	// json.Marshal normally adds, so the result is a drop-in fragment.
	got, _ := EscapeJSON("x")
	if len(got) > 0 && (got[0] == '"' || got[len(got)-1] == '"') {
		t.Errorf("EscapeJSON should not include surrounding quotes, got %q", got)
	}
}

func TestEscapeUnescapeJSON_RoundTrip(t *testing.T) {
	original := "quotes \" and backslash \\ and newline \n and tab \t and unicode 你好"
	escaped, err := EscapeJSON(original)
	if err != nil {
		t.Fatalf("escape error: %v", err)
	}
	back, err := UnescapeJSON(escaped)
	if err != nil {
		t.Fatalf("unescape error: %v", err)
	}
	if back != original {
		t.Errorf("round trip = %q, want %q", back, original)
	}
}

func TestUnescapeJSON_InvalidEscape(t *testing.T) {
	_, err := UnescapeJSON(`bad \x escape`)
	if err == nil {
		t.Fatal("expected an error for an invalid JSON escape sequence, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d)", CodeOf(err), CodeInput)
	}
}

func TestUnescapeJSON_UnterminatedString(t *testing.T) {
	// A bare unescaped quote makes the wrapped `"...."` invalid JSON —
	// this is the "malformed input from a different tool" case: text
	// containing a literal " that wasn't meant as a JSON escape target.
	_, err := UnescapeJSON(`unterminated " string`)
	if err == nil {
		t.Fatal("expected an error for an unescaped embedded quote, got nil")
	}
}

func TestEscapeURL_QueryVsPath(t *testing.T) {
	text := "a b/c"
	query := EscapeURL(text, false)
	path := EscapeURL(text, true)
	// QueryEscape encodes space as '+', PathEscape encodes it as '%20'
	// — this is the exact distinction the --url-path flag exists for.
	if query != "a+b%2Fc" {
		t.Errorf("query escape = %q, want %q", query, "a+b%2Fc")
	}
	if path != "a%20b%2Fc" {
		t.Errorf("path escape = %q, want %q", path, "a%20b%2Fc")
	}
}

func TestEscapeUnescapeURL_RoundTrip(t *testing.T) {
	original := "hello world & more=stuff?"
	for _, isPath := range []bool{true, false} {
		escaped := EscapeURL(original, isPath)
		back, err := UnescapeURL(escaped, isPath)
		if err != nil {
			t.Fatalf("unescape error (path=%v): %v", isPath, err)
		}
		if back != original {
			t.Errorf("round trip (path=%v) = %q, want %q", isPath, back, original)
		}
	}
}

func TestUnescapeURL_InvalidPercentEncoding(t *testing.T) {
	_, err := UnescapeURL("100%", false)
	if err == nil {
		t.Fatal("expected an error for a truncated percent-escape, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d)", CodeOf(err), CodeInput)
	}
}

func TestEscapeHTML_Basic(t *testing.T) {
	got := EscapeHTML(`<script>alert("hi")</script>`)
	want := `&lt;script&gt;alert(&#34;hi&#34;)&lt;/script&gt;`
	if got != want {
		t.Errorf("EscapeHTML = %q, want %q", got, want)
	}
}

func TestEscapeUnescapeHTML_RoundTrip(t *testing.T) {
	original := `<b>bold</b> & "quoted" & 'single'`
	escaped := EscapeHTML(original)
	back := UnescapeHTML(escaped)
	if back != original {
		t.Errorf("round trip = %q, want %q", back, original)
	}
}

func TestEscapeUnicode_ASCIIPassesThrough(t *testing.T) {
	got := EscapeUnicode("plain ascii text 123")
	if got != "plain ascii text 123" {
		t.Errorf("ASCII text should pass through unchanged, got %q", got)
	}
}

func TestEscapeUnicode_BMPCharacter(t *testing.T) {
	got := EscapeUnicode("你")
	want := `\u4f60`
	if got != want {
		t.Errorf("EscapeUnicode(你) = %q, want %q", got, want)
	}
}

func TestEscapeUnicode_SurrogatePairForAstralCharacter(t *testing.T) {
	// U+1F600 (😀) is outside the Basic Multilingual Plane and must be
	// encoded as a UTF-16 surrogate pair, matching how JSON string
	// literals represent characters beyond U+FFFF.
	got := EscapeUnicode("😀")
	want := `\ud83d\ude00`
	if got != want {
		t.Errorf("EscapeUnicode(😀) = %q, want %q", got, want)
	}
}

func TestEscapeUnescapeUnicode_RoundTrip(t *testing.T) {
	original := "mixed ascii, 你好, and emoji 😀🎉"
	escaped := EscapeUnicode(original)
	back, err := UnescapeUnicode(escaped)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if back != original {
		t.Errorf("round trip = %q, want %q", back, original)
	}
}

func TestUnescapeUnicode_TruncatedEscape(t *testing.T) {
	_, err := UnescapeUnicode(`abc\u12`)
	if err == nil {
		t.Fatal("expected an error for a truncated \\u escape, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d)", CodeOf(err), CodeInput)
	}
}

func TestUnescapeUnicode_InvalidHexDigits(t *testing.T) {
	_, err := UnescapeUnicode(`\uZZZZ`)
	if err == nil {
		t.Fatal("expected an error for non-hex digits in a \\u escape, got nil")
	}
}

func TestUnescapeUnicode_LoneHighSurrogateDoesNotErrorOrHang(t *testing.T) {
	// A high surrogate with no valid low surrogate following it is
	// malformed UTF-16. UnescapeUnicode's recombination check simply
	// doesn't fire in that case, and the lone half gets passed to
	// strings.Builder.WriteRune — which substitutes the U+FFFD
	// replacement character, since a bare surrogate half is not a
	// valid Unicode scalar value on its own and can't be UTF-8
	// encoded directly. The behavior worth pinning down here is that
	// this path doesn't error out or hang; it degrades to a
	// replacement character instead.
	got, err := UnescapeUnicode(`\ud800end`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Error("expected some output for a lone high surrogate, got empty string")
	}
	if !strings.HasSuffix(got, "end") {
		t.Errorf("expected the literal trailing text to survive, got %q", got)
	}
}
