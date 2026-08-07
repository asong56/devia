package core

import "testing"

func TestRegexTest_BasicMatches(t *testing.T) {
	matches, err := RegexTest(`\d+`, "", "room 42, floor 3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("got %d matches, want 2: %+v", len(matches), matches)
	}
	if matches[0].Match != "42" || matches[0].Start != 5 || matches[0].End != 7 {
		t.Errorf("matches[0] = %+v, want Match=42 Start=5 End=7", matches[0])
	}
	if matches[1].Match != "3" {
		t.Errorf("matches[1].Match = %q, want %q", matches[1].Match, "3")
	}
}

func TestRegexTest_NoMatchesReturnsEmptyNotNilError(t *testing.T) {
	matches, err := RegexTest(`xyz`, "", "no such substring here")
	if err != nil {
		t.Fatalf("zero matches is not an error condition: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("got %d matches, want 0", len(matches))
	}
}

func TestRegexTest_CaptureGroups(t *testing.T) {
	matches, err := RegexTest(`(\w+)@(\w+)\.com`, "", "contact: alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	if len(matches[0].Groups) != 2 {
		t.Fatalf("got %d groups, want 2: %+v", len(matches[0].Groups), matches[0].Groups)
	}
	if matches[0].Groups[0] != "alice" || matches[0].Groups[1] != "example" {
		t.Errorf("Groups = %v, want [alice example]", matches[0].Groups)
	}
}

func TestRegexTest_OptionalGroupNotParticipating(t *testing.T) {
	// A group that's part of the pattern but didn't participate in a
	// particular match (e.g. an alternation that took the other
	// branch) must come back as "" rather than panicking on a
	// negative submatch index.
	matches, err := RegexTest(`(a)|(b)`, "", "b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}
	if len(matches[0].Groups) != 2 {
		t.Fatalf("got %d groups, want 2", len(matches[0].Groups))
	}
	if matches[0].Groups[0] != "" {
		t.Errorf("non-participating group[0] = %q, want empty string", matches[0].Groups[0])
	}
	if matches[0].Groups[1] != "b" {
		t.Errorf("group[1] = %q, want %q", matches[0].Groups[1], "b")
	}
}

func TestRegexTest_CaseInsensitiveFlag(t *testing.T) {
	matches, err := RegexTest(`hello`, "i", "HELLO world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 {
		t.Errorf("case-insensitive match failed: got %d matches, want 1", len(matches))
	}
}

func TestRegexTest_MultilineFlag(t *testing.T) {
	text := "first\nsecond\nthird"
	// Without 'm', ^ only matches at the very start of the string.
	without, err := RegexTest(`^\w+`, "", text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(without) != 1 {
		t.Errorf("without 'm' flag: got %d matches, want 1", len(without))
	}
	// With 'm', ^ matches at the start of every line.
	with, err := RegexTest(`^\w+`, "m", text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(with) != 3 {
		t.Errorf("with 'm' flag: got %d matches, want 3", len(with))
	}
}

func TestRegexTest_InvalidPattern(t *testing.T) {
	_, err := RegexTest(`[unclosed`, "", "text")
	if err == nil {
		t.Fatal("expected an error for an invalid/unclosed regex, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d)", CodeOf(err), CodeInput)
	}
}

func TestRegexTest_UnsupportedFlagsAreIgnoredNotErrors(t *testing.T) {
	// compileRegex only recognizes i,m,s,U — any other character is
	// silently dropped rather than causing a compile error. Pinning
	// this down so a future change to that filtering doesn't silently
	// start rejecting flag strings that used to work.
	_, err := RegexTest(`abc`, "izzz", "abc")
	if err != nil {
		t.Errorf("unexpected error for a flag string with unrecognized characters: %v", err)
	}
}

func TestRegexReplace_Basic(t *testing.T) {
	got, err := RegexReplace(`\d+`, "", "room 42, floor 3", "#")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "room #, floor #"
	if got != want {
		t.Errorf("RegexReplace = %q, want %q", got, want)
	}
}

func TestRegexReplace_GroupReferences(t *testing.T) {
	got, err := RegexReplace(`(\w+)@(\w+)\.com`, "", "alice@example.com", "$1 at $2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "alice at example"
	if got != want {
		t.Errorf("RegexReplace = %q, want %q", got, want)
	}
}

func TestRegexReplace_NoMatchReturnsInputUnchanged(t *testing.T) {
	got, err := RegexReplace(`xyz`, "", "unrelated text", "REPLACED")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "unrelated text" {
		t.Errorf("RegexReplace with no match = %q, want the original text unchanged", got)
	}
}

func TestRegexReplace_InvalidPattern(t *testing.T) {
	_, err := RegexReplace(`(unclosed`, "", "text", "x")
	if err == nil {
		t.Fatal("expected an error for an invalid regex, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d)", CodeOf(err), CodeInput)
	}
}

func TestRegexTest_ReDoSProneInputDoesNotHang(t *testing.T) {
	// This is the whole point of RE2/Go regexp instead of a
	// backtracking engine: a classically catastrophic-backtracking
	// pattern against adversarial input must still return quickly,
	// linear in input size, not hang. If this test ever times out, RE2
	// guarantees have been broken somehow.
	pattern := `(a+)+b`
	input := ""
	for i := 0; i < 40; i++ {
		input += "a"
	}
	// deliberately no trailing "b" — a backtracking engine would blow
	// up combinatorially trying every grouping before giving up.
	_, err := RegexTest(pattern, "", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// If we get here at all (rather than hanging), RE2's linear-time
	// guarantee held. No further assertion needed.
}
