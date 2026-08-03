package core

import "testing"

func TestTextTransformModes(t *testing.T) {
	cases := []struct {
		mode string
		in   string
		want string
	}{
		{"lower", "Hello World", "hello world"},
		{"upper", "Hello World", "HELLO WORLD"},
		{"sentence", "HELLO world", "Hello world"},
		{"title", "the great gatsby", "The Great Gatsby"},
		{"camel", "hello-world", "helloWorld"},
		{"camel", "HTTPServer", "httpServer"},
		{"pascal", "hello-world", "HelloWorld"},
		{"pascal", "hello_world", "HelloWorld"},
		{"snake", "helloWorld", "hello_world"},
		{"snake", "HTTPServer", "http_server"},
		{"constant", "hello-world", "HELLO_WORLD"},
		{"kebab", "helloWorld", "hello-world"},
		{"cobol", "hello world", "HELLO-WORLD"},
		{"train", "hello world", "Hello-World"},
		{"alternating", "hello world", "hElLo WoRlD"},
		{"inverse", "Hello World", "hELLO wORLD"},
	}
	for _, c := range cases {
		got, err := TextTransform(c.mode, c.in)
		if err != nil {
			t.Fatalf("TextTransform(%s, %q): %v", c.mode, c.in, err)
		}
		if got != c.want {
			t.Errorf("TextTransform(%s, %q) = %q, want %q", c.mode, c.in, got, c.want)
		}
	}
}

func TestTextTransformUnknownMode(t *testing.T) {
	_, err := TextTransform("shout-whisper", "x")
	if err == nil {
		t.Fatal("expected an error for an unknown mode")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("unknown mode should be a CodeInput error, got code %d", CodeOf(err))
	}
}

func TestTextTransformEmptyInput(t *testing.T) {
	got, err := TextTransform("upper", "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("expected empty string to stay empty, got %q", got)
	}
}
