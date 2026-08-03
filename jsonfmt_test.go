package core

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONFormatDefaultIndent(t *testing.T) {
	// encoding/json marshals map keys in sorted order regardless of the
	// order they appeared in the source text, so "b" before "a" in the
	// input still comes out "a" before "b" here.
	got, err := JSONFormat(`{"b":1,"a":2}`, "")
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"a\": 2,\n  \"b\": 1\n}"
	if got != want {
		t.Errorf("JSONFormat = %q, want %q", got, want)
	}
}

func TestJSONFormatCustomIndent(t *testing.T) {
	got, err := JSONFormat(`{"a":1}`, "\t")
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n\t\"a\": 1\n}"
	if got != want {
		t.Errorf("JSONFormat with tab indent = %q, want %q", got, want)
	}
}

func TestJSONFormatPreservesLargeIntegers(t *testing.T) {
	// A number larger than float64 can represent exactly must round-trip
	// unchanged (this is what json.Number / UseNumber() is for).
	got, err := JSONFormat(`{"id":9223372036854775807}`, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "9223372036854775807") {
		t.Errorf("large integer was not preserved exactly, got %q", got)
	}
}

func TestJSONFormatInvalidInput(t *testing.T) {
	_, err := JSONFormat(`{not valid json`, "")
	if err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("invalid JSON should be a CodeInput error, got code %d", CodeOf(err))
	}
}

func TestJSONMinify(t *testing.T) {
	got, err := JSONMinify("{\n  \"a\" : 1,\n  \"b\" : [1, 2, 3]\n}")
	if err != nil {
		t.Fatal(err)
	}
	want := `{"a":1,"b":[1,2,3]}`
	if got != want {
		t.Errorf("JSONMinify = %q, want %q", got, want)
	}
}

func TestJSONMinifyInvalidInput(t *testing.T) {
	_, err := JSONMinify(`[1, 2,`)
	if err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("invalid JSON should be a CodeInput error, got code %d", CodeOf(err))
	}
}

func TestJSONValidate(t *testing.T) {
	if err := JSONValidate(`{"a": [1, 2, true, null, "x"]}`); err != nil {
		t.Errorf("expected valid JSON to pass, got %v", err)
	}
	if err := JSONValidate(`{"a": }`); err == nil {
		t.Error("expected invalid JSON to fail")
	} else if CodeOf(err) != CodeInput {
		t.Errorf("invalid JSON should be a CodeInput error, got code %d", CodeOf(err))
	}
}

// Sanity check that our expectations above actually match what
// encoding/json itself considers minified/compact, rather than a
// hand-typed guess.
func TestJSONMinifyMatchesStdlibCompact(t *testing.T) {
	// Keys are already alphabetical here on purpose: JSONMinify decodes
	// and re-marshals (which sorts map keys), while json.Compact only
	// strips whitespace byte-for-byte and preserves source order. The
	// two only agree when the source order already happens to be sorted.
	input := `{"a": 1, "z": [1,2,3]}`
	got, err := JSONMinify(input)
	if err != nil {
		t.Fatal(err)
	}
	var buf strings.Builder
	if err := json.Compact(&buf, []byte(input)); err != nil {
		t.Fatal(err)
	}
	if got != buf.String() {
		t.Errorf("JSONMinify = %q, want %q (json.Compact)", got, buf.String())
	}
}
