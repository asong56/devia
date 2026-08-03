package core

import (
	"math/rand"
	"strings"
)

var loremWords = []string{
	"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing",
	"elit", "sed", "do", "eiusmod", "tempor", "incididunt", "ut", "labore",
	"et", "dolore", "magna", "aliqua", "enim", "ad", "minim", "veniam",
	"quis", "nostrud", "exercitation", "ullamco", "laboris", "nisi",
	"aliquip", "ex", "ea", "commodo", "consequat", "duis", "aute", "irure",
	"in", "reprehenderit", "voluptate", "velit", "esse", "cillum", "eu",
	"fugiat", "nulla", "pariatur", "excepteur", "sint", "occaecat",
	"cupidatat", "non", "proident", "sunt", "culpa", "qui", "officia",
	"deserunt", "mollit", "anim", "id", "est", "laborum",
}

func loremWordsN(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = loremWords[rand.Intn(len(loremWords))]
	}
	return out
}

func loremSentence(minWords, maxWords int) string {
	n := minWords + rand.Intn(maxWords-minWords+1)
	words := loremWordsN(n)
	words[0] = capitalize(words[0])
	return strings.Join(words, " ") + "."
}

// LoremGenerate produces lorem-ipsum text of the given kind ("word",
// "sentence", or "paragraph") and count, optionally forcing the
// classic "Lorem ipsum dolor sit amet..." opener.
func LoremGenerate(kind string, count int, classicStart bool) (string, error) {
	if count < 1 {
		count = 1
	}
	switch kind {
	case "word":
		words := loremWordsN(count)
		if classicStart && len(words) > 0 {
			words[0] = "lorem"
		}
		return strings.Join(words, " "), nil
	case "sentence":
		sentences := make([]string, count)
		for i := range sentences {
			sentences[i] = loremSentence(5, 12)
		}
		if classicStart {
			sentences[0] = "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
		}
		return strings.Join(sentences, " "), nil
	case "paragraph", "":
		paragraphs := make([]string, count)
		for i := range paragraphs {
			sCount := 3 + rand.Intn(4)
			sentences := make([]string, sCount)
			for j := range sentences {
				sentences[j] = loremSentence(5, 12)
			}
			paragraphs[i] = strings.Join(sentences, " ")
		}
		if classicStart {
			paragraphs[0] = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " + paragraphs[0]
		}
		return strings.Join(paragraphs, "\n\n"), nil
	default:
		return "", NewInputError("unsupported lorem type: " + kind + " (want word|sentence|paragraph)")
	}
}
