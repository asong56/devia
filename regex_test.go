package core

import "testing"

func TestRegexTestFindsMatchesWithGroups(t *testing.T) {
	matches, err := RegexTest(`(\d+)-(\d+)`, "", "range 10-20 and 30-40")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].Match != "10-20" || matches[0].Groups[0] != "10" || matches[0].Groups[1] != "20" {
		t.Errorf("first match = %+v, want Match=10-20 Groups=[10 20]", matches[0])
	}
	if matches[1].Match != "30-40" || matches[1].Groups[0] != "30" || matches[1].Groups[1] != "40" {
		t.Errorf("second match = %+v, want Match=30-40 Groups=[30 40]", matches[1])
	}
}

func TestRegexTestCaseInsensitiveFlag(t *testing.T) {
	matches, err := RegexTest(`hello`, "i", "HELLO there")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 case-insensitive match, got %d", len(matches))
	}
	if matches[0].Match != "HELLO" {
		t.Errorf("match = %q, want %q", matches[0].Match, "HELLO")
	}

	// Without the flag, it should not match.
	noMatches, err := RegexTest(`hello`, "", "HELLO there")
	if err != nil {
		t.Fatal(err)
	}
	if len(noMatches) != 0 {
		t.Errorf("expected no case-sensitive match, got %d", len(noMatches))
	}
}

func TestRegexTestNoMatches(t *testing.T) {
	matches, err := RegexTest(`xyz`, "", "abc def")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestRegexTestInvalidPattern(t *testing.T) {
	_, err := RegexTest(`(unclosed`, "", "text")
	if err == nil {
		t.Fatal("expected an error for an invalid regex pattern")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("invalid pattern should be a CodeInput error, got code %d", CodeOf(err))
	}
}

func TestRegexReplaceWithGroupReferences(t *testing.T) {
	got, err := RegexReplace(`(\w+)@(\w+)\.com`, "", "contact us at info@example.com today", "$1 at $2")
	if err != nil {
		t.Fatal(err)
	}
	want := "contact us at info at example today"
	if got != want {
		t.Errorf("RegexReplace = %q, want %q", got, want)
	}
}

func TestRegexReplaceInvalidPattern(t *testing.T) {
	_, err := RegexReplace(`[unclosed`, "", "text", "x")
	if err == nil {
		t.Fatal("expected an error for an invalid regex pattern")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("invalid pattern should be a CodeInput error, got code %d", CodeOf(err))
	}
}
