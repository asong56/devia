package core

import (
	"strings"
	"testing"
)

func TestLoremGenerate_WordCount(t *testing.T) {
	got, err := LoremGenerate("word", 5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	words := strings.Fields(got)
	if len(words) != 5 {
		t.Errorf("got %d words, want 5 (output: %q)", len(words), got)
	}
}

func TestLoremGenerate_WordClassicStart(t *testing.T) {
	got, err := LoremGenerate("word", 3, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	words := strings.Fields(got)
	if len(words) != 3 {
		t.Fatalf("got %d words, want 3", len(words))
	}
	if words[0] != "lorem" {
		t.Errorf("first word with classicStart=true = %q, want %q", words[0], "lorem")
	}
}

func TestLoremGenerate_SentenceCount(t *testing.T) {
	got, err := LoremGenerate("sentence", 3, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Sentences are joined by a single space and each ends in ".", so
	// counting ". " boundaries (plus the final period) gives the
	// sentence count.
	count := strings.Count(got, ". ") + 1
	if count != 3 {
		t.Errorf("got %d sentences (by '. ' boundary count), want 3: %q", count, got)
	}
	if !strings.HasSuffix(got, ".") {
		t.Errorf("output should end with a period: %q", got)
	}
}

func TestLoremGenerate_SentenceStartsCapitalized(t *testing.T) {
	got, err := LoremGenerate("sentence", 1, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Fatal("expected non-empty output")
	}
	first := got[0]
	if first < 'A' || first > 'Z' {
		t.Errorf("sentence should start with a capital letter, got %q", got)
	}
}

func TestLoremGenerate_SentenceClassicStart(t *testing.T) {
	got, err := LoremGenerate("sentence", 2, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	classic := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
	if !strings.HasPrefix(got, classic) {
		t.Errorf("expected output to start with the classic opener %q, got %q", classic, got)
	}
}

func TestLoremGenerate_ParagraphCount(t *testing.T) {
	got, err := LoremGenerate("paragraph", 3, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	paragraphs := strings.Split(got, "\n\n")
	if len(paragraphs) != 3 {
		t.Errorf("got %d paragraphs, want 3", len(paragraphs))
	}
	for i, p := range paragraphs {
		if strings.TrimSpace(p) == "" {
			t.Errorf("paragraph %d is empty", i)
		}
	}
}

func TestLoremGenerate_ParagraphClassicStart(t *testing.T) {
	got, err := LoremGenerate("paragraph", 1, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	classic := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
	if !strings.HasPrefix(got, classic) {
		t.Errorf("expected paragraph to start with the classic opener, got %q", got)
	}
}

func TestLoremGenerate_EmptyKindDefaultsToParagraph(t *testing.T) {
	withEmpty, err := LoremGenerate("", 1, true)
	if err != nil {
		t.Fatalf("unexpected error for empty kind: %v", err)
	}
	classic := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
	if !strings.HasPrefix(withEmpty, classic) {
		t.Errorf("empty kind should behave like 'paragraph', got %q", withEmpty)
	}
}

func TestLoremGenerate_ZeroOrNegativeCountDefaultsToOne(t *testing.T) {
	for _, n := range []int{0, -1, -50} {
		got, err := LoremGenerate("word", n, false)
		if err != nil {
			t.Fatalf("LoremGenerate(word, %d): unexpected error: %v", n, err)
		}
		words := strings.Fields(got)
		if len(words) != 1 {
			t.Errorf("LoremGenerate(word, %d): got %d words, want 1 (count should clamp up)", n, len(words))
		}
	}
}

func TestLoremGenerate_UnsupportedKind(t *testing.T) {
	_, err := LoremGenerate("sentences", 1, false) // typo'd plural, not a real mode
	if err == nil {
		t.Fatal("expected an error for an unsupported lorem kind, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d)", CodeOf(err), CodeInput)
	}
}

func TestLoremGenerate_WordsAreFromTheKnownList(t *testing.T) {
	got, err := LoremGenerate("word", 30, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	known := make(map[string]bool, len(loremWords))
	for _, w := range loremWords {
		known[w] = true
	}
	for _, w := range strings.Fields(got) {
		if !known[w] {
			t.Errorf("word %q is not in the known lorem word list", w)
		}
	}
}
