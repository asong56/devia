package core

import "testing"

func TestTextTransform_BasicCases(t *testing.T) {
	cases := []struct {
		mode string
		in   string
		want string
	}{
		{"lower", "Hello World", "hello world"},
		{"upper", "Hello World", "HELLO WORLD"},
		{"sentence", "HELLO world", "Hello world"},
		{"title", "hello world", "Hello World"},
		{"camel", "hello world", "helloWorld"},
		{"camel", "hello_world", "helloWorld"},
		{"camel", "hello-world", "helloWorld"},
		{"pascal", "hello world", "HelloWorld"},
		{"snake", "helloWorld", "hello_world"},
		{"constant", "helloWorld", "HELLO_WORLD"},
		{"kebab", "helloWorld", "hello-world"},
		{"cobol", "helloWorld", "HELLO-WORLD"},
		{"train", "hello world", "Hello-World"},
	}
	for _, c := range cases {
		got, err := TextTransform(c.mode, c.in)
		if err != nil {
			t.Errorf("TextTransform(%q, %q): unexpected error: %v", c.mode, c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("TextTransform(%q, %q) = %q, want %q", c.mode, c.in, got, c.want)
		}
	}
}

func TestTextTransform_AcronymBoundary(t *testing.T) {
	// "HTTPServer" is the exact case splitWords' acronym-boundary logic
	// exists for: the run of capitals "HTTP" is one word, and the
	// break happens right before the "S" that starts "Server" — not
	// after the first capital, and not treating each capital letter as
	// its own word.
	cases := []struct {
		mode string
		want string
	}{
		{"camel", "httpServer"},
		{"pascal", "HttpServer"},
		{"snake", "http_server"},
		{"kebab", "http-server"},
		{"constant", "HTTP_SERVER"},
	}
	for _, c := range cases {
		got, err := TextTransform(c.mode, "HTTPServer")
		if err != nil {
			t.Errorf("TextTransform(%q, HTTPServer): unexpected error: %v", c.mode, err)
			continue
		}
		if got != c.want {
			t.Errorf("TextTransform(%q, HTTPServer) = %q, want %q", c.mode, got, c.want)
		}
	}
}

func TestTextTransform_SingleTrailingCapitalIsNotSplitOff(t *testing.T) {
	// A single capital at the end ("aB") is the plain camelCase boundary
	// case (break before an uppercase letter that follows a lowercase
	// one) — distinct from the acronym-run case above, and worth
	// pinning down separately since the two conditions in splitWords'
	// uppercase branch are independent triggers for the same flush.
	got, err := TextTransform("snake", "myID")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "my_id" {
		t.Errorf("TextTransform(snake, myID) = %q, want %q", got, "my_id")
	}
}

func TestTextTransform_Title_OnlySplitsOnWhitespace(t *testing.T) {
	// "title" mode is built on strings.Fields (whitespace only), not
	// splitWords (which also splits on _/-/.) — so unlike camel/snake/
	// kebab/etc., a hyphenated or underscored input is treated as ONE
	// word for title-casing, not split into separate words. This is a
	// real, easy-to-miss inconsistency between "title" and every other
	// mode, worth a test specifically because it's surprising.
	got, err := TextTransform("title", "hello-world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Hello-world" {
		t.Errorf("TextTransform(title, hello-world) = %q, want %q (only the leading letter of the single whitespace-delimited token is capitalized)", got, "Hello-world")
	}
}

func TestTextTransform_Alternating(t *testing.T) {
	got, err := TextTransform("alternating", "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "hElLo WoRlD"
	if got != want {
		t.Errorf("TextTransform(alternating, ...) = %q, want %q", got, want)
	}
}

func TestTextTransform_AlternatingSkipsNonLettersButKeepsCaseState(t *testing.T) {
	// Non-letter characters (spaces, punctuation, digits) pass through
	// unchanged and must NOT consume a turn of the upper/lower
	// alternation — only letters advance the toggle.
	got, err := TextTransform("alternating", "a1b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "a1B"
	if got != want {
		t.Errorf("TextTransform(alternating, a1b) = %q, want %q", got, want)
	}
}

func TestTextTransform_Inverse(t *testing.T) {
	got, err := TextTransform("inverse", "Hello World 123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "hELLO wORLD 123"
	if got != want {
		t.Errorf("TextTransform(inverse, ...) = %q, want %q", got, want)
	}
}

func TestTextTransform_EmptyInput(t *testing.T) {
	// Every mode should handle an empty string gracefully — no panic,
	// no error, just an empty (or near-empty) result.
	modes := []string{"lower", "upper", "sentence", "title", "camel",
		"pascal", "snake", "constant", "kebab", "cobol", "train",
		"alternating", "inverse"}
	for _, mode := range modes {
		got, err := TextTransform(mode, "")
		if err != nil {
			t.Errorf("TextTransform(%q, \"\"): unexpected error: %v", mode, err)
			continue
		}
		if got != "" {
			t.Errorf("TextTransform(%q, \"\") = %q, want empty string", mode, got)
		}
	}
}

func TestTextTransform_UnsupportedMode(t *testing.T) {
	_, err := TextTransform("shouty-kebab-reverse", "x")
	if err == nil {
		t.Fatal("expected an error for an unsupported mode, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d)", CodeOf(err), CodeInput)
	}
}

func TestTextTransform_UnicodeInput(t *testing.T) {
	// unicode.IsUpper/IsLower/ToUpper/ToLower are used throughout, not
	// ASCII-only byte comparisons — this should hold for non-ASCII
	// letters too (Cyrillic Б has real case folding).
	got, err := TextTransform("upper", "hello Привет")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "HELLO ПРИВЕТ"
	if got != want {
		t.Errorf("TextTransform(upper, ...) = %q, want %q", got, want)
	}
}
