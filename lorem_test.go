package core

import (
	"strings"
	"testing"
)

func TestLoremGenerateWordCount(t *testing.T) {
	got, err := LoremGenerate("word", 5, false)
	if err != nil {
		t.Fatal(err)
	}
	words := strings.Fields(got)
	if len(words) != 5 {
		t.Errorf("expected 5 words, got %d (%q)", len(words), got)
	}
}

func TestLoremGenerateSentenceCount(t *testing.T) {
	got, err := LoremGenerate("sentence", 3, false)
	if err != nil {
		t.Fatal(err)
	}
	// Sentences are joined with a single space and each ends in ".",
	// so the number of sentences equals the number of "." characters.
	if n := strings.Count(got, "."); n != 3 {
		t.Errorf("expected 3 sentences (3 periods), got %d in %q", n, got)
	}
}

func TestLoremGenerateParagraphCount(t *testing.T) {
	got, err := LoremGenerate("paragraph", 2, false)
	if err != nil {
		t.Fatal(err)
	}
	paragraphs := strings.Split(got, "\n\n")
	if len(paragraphs) != 2 {
		t.Errorf("expected 2 paragraphs, got %d", len(paragraphs))
	}
}

func TestLoremGenerateClassicOpener(t *testing.T) {
	word, err := LoremGenerate("word", 3, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(word, "lorem") {
		t.Errorf("classic word output should start with 'lorem', got %q", word)
	}

	sentence, err := LoremGenerate("sentence", 2, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(sentence, "Lorem ipsum dolor sit amet, consectetur adipiscing elit.") {
		t.Errorf("classic sentence output should start with the standard opener, got %q", sentence)
	}

	paragraph, err := LoremGenerate("paragraph", 1, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(paragraph, "Lorem ipsum dolor sit amet, consectetur adipiscing elit.") {
		t.Errorf("classic paragraph output should start with the standard opener, got %q", paragraph)
	}
}

func TestLoremGenerateInvalidKind(t *testing.T) {
	_, err := LoremGenerate("haiku", 1, false)
	if err == nil {
		t.Fatal("expected an error for an unsupported lorem type")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("expected CodeInput, got code %d", CodeOf(err))
	}
}

func TestLoremGenerateCountBelowOneDefaultsToOne(t *testing.T) {
	got, err := LoremGenerate("word", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(strings.Fields(got)) != 1 {
		t.Errorf("count 0 should default to 1 word, got %q", got)
	}
}
