package core

import (
	"strings"
	"testing"
)

func TestJSONFormat_Basic(t *testing.T) {
	got, err := JSONFormat(`{"b":2,"a":1}`, "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// json.MarshalIndent on a map produces keys in sorted order, so "a"
	// (inserted second in the source but alphabetically first) comes
	// before "b" in the output.
	want := "{\n  \"a\": 1,\n  \"b\": 2\n}"
	if got != want {
		t.Errorf("JSONFormat = %q, want %q", got, want)
	}
}

func TestJSONFormat_DefaultIndentWhenEmpty(t *testing.T) {
	withEmpty, err := JSONFormat(`{"a":1}`, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	withTwoSpaces, err := JSONFormat(`{"a":1}`, "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if withEmpty != withTwoSpaces {
		t.Errorf("empty indent should default to two spaces: %q vs %q", withEmpty, withTwoSpaces)
	}
}

func TestJSONFormat_CustomIndent(t *testing.T) {
	got, err := JSONFormat(`{"a":1}`, "\t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "\t\"a\"") {
		t.Errorf("expected a tab-indented line, got %q", got)
	}
}

func TestJSONFormat_InvalidJSON(t *testing.T) {
	_, err := JSONFormat(`{"a":}`, "  ")
	if err == nil {
		t.Fatal("expected an error for malformed JSON, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d)", CodeOf(err), CodeInput)
	}
}

func TestJSONFormat_TrailingGarbageIsRejected(t *testing.T) {
	// json.Decoder.Decode alone only consumes the first JSON value and
	// silently ignores whatever comes after — without an explicit
	// trailing-data check, "{"a":1} garbage" would format just the
	// first object and drop "garbage" with no error, while `devia json
	// validate` (which uses json.Unmarshal, not Decoder) would
	// correctly reject the very same input. That inconsistency is what
	// this test guards against.
	_, err := JSONFormat(`{"a":1} garbage`, "  ")
	if err == nil {
		t.Error("expected an error for trailing garbage after valid JSON")
	}
}

func TestJSONFormat_TrailingWhitespaceIsFine(t *testing.T) {
	// The trailing-data check above must not get overzealous and start
	// rejecting the ordinary trailing newline a file or `echo` adds.
	_, err := JSONFormat("{\"a\":1}\n", "  ")
	if err != nil {
		t.Errorf("trailing whitespace-only content should not be an error: %v", err)
	}
}

func TestJSONFormat_ConcatenatedObjectsAreRejected(t *testing.T) {
	// Two back-to-back valid JSON objects is exactly the case
	// json.Decoder is designed to accept as a *stream* of values —
	// which is precisely why JSONFormat needs its own explicit check:
	// devia's `format` is a single-document formatter, not a JSON
	// Lines processor, so this must be treated as invalid input rather
	// than silently formatting only the first object.
	_, err := JSONFormat(`{"a":1}{"b":2}`, "  ")
	if err == nil {
		t.Error("expected an error for two concatenated top-level JSON values")
	}
}

func TestJSONFormat_PreservesLargeIntegers(t *testing.T) {
	// json.Number decoding (dec.UseNumber()) exists specifically so a
	// large integer that would lose precision as a float64 survives
	// the round trip unchanged. This is the test that would catch a
	// regression back to plain interface{} decoding.
	big := `{"id":9007199254740993}` // 2^53 + 1 — not exactly representable as float64
	got, err := JSONFormat(big, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "9007199254740993") {
		t.Errorf("large integer was not preserved exactly, got %q", got)
	}
}

func TestJSONMinify_RemovesWhitespace(t *testing.T) {
	got, err := JSONMinify("{\n  \"a\" : 1,\n  \"b\" : [1, 2, 3]\n}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{"a":1,"b":[1,2,3]}`
	if got != want {
		t.Errorf("JSONMinify = %q, want %q", got, want)
	}
}

func TestJSONMinify_InvalidJSON(t *testing.T) {
	_, err := JSONMinify(`not json at all`)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d)", CodeOf(err), CodeInput)
	}
}

func TestJSONMinify_TrailingGarbageIsRejected(t *testing.T) {
	// Same reasoning as TestJSONFormat_TrailingGarbageIsRejected —
	// JSONMinify shares decodeSingleJSONValue with JSONFormat, so it
	// must reject the same malformed input the same way.
	_, err := JSONMinify(`{"a":1} garbage`)
	if err == nil {
		t.Error("expected an error for trailing garbage after valid JSON")
	}
}

func TestJSONValidate_Valid(t *testing.T) {
	cases := []string{`{}`, `[]`, `null`, `true`, `1`, `"a string"`, `{"nested":{"a":[1,2,3]}}`}
	for _, c := range cases {
		if err := JSONValidate(c); err != nil {
			t.Errorf("JSONValidate(%q) returned an error for valid JSON: %v", c, err)
		}
	}
}

func TestJSONValidate_Invalid(t *testing.T) {
	cases := []string{``, `{`, `{"a":}`, `{'a':1}`, `undefined`, `{"a":1,}`}
	for _, c := range cases {
		err := JSONValidate(c)
		if err == nil {
			t.Errorf("JSONValidate(%q) = nil, want an error", c)
			continue
		}
		if CodeOf(err) != CodeInput {
			t.Errorf("JSONValidate(%q): CodeOf(err) = %d, want CodeInput (%d)", c, CodeOf(err), CodeInput)
		}
	}
}
