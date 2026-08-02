package core

import (
	"strings"
	"unicode"
)

// splitWords tokenizes an identifier-like string — camelCase,
// PascalCase, snake_case, kebab-case, space separated, or an acronym
// run like "HTTPServer" — into lowercase words. This is the shared
// foundation every case-style conversion below is built on: convert
// once to a canonical word list, then reassemble in the target style.
func splitWords(s string) []string {
	var words []string
	var cur []rune
	runes := []rune(s)

	flush := func() {
		if len(cur) > 0 {
			words = append(words, strings.ToLower(string(cur)))
			cur = nil
		}
	}

	for i, r := range runes {
		switch {
		case r == '_' || r == '-' || r == ' ' || r == '.':
			flush()
		case unicode.IsUpper(r):
			prevLower := i > 0 && (unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1]))
			nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
			// Break before an uppercase letter that follows a lowercase
			// one ("aB"), and before the last letter of an acronym run
			// that's followed by a new lowercase word ("HTTPServer" ->
			// break before the "S").
			if prevLower || (len(cur) > 0 && nextLower && allUpper(cur)) {
				flush()
			}
			cur = append(cur, r)
		default:
			cur = append(cur, r)
		}
	}
	flush()
	return words
}

func allUpper(rs []rune) bool {
	for _, r := range rs {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// TextTransform applies one of the supported case transformations.
// Supported modes: lower, upper, sentence, title, camel, pascal,
// snake, constant, kebab, cobol, train, alternating, inverse.
func TextTransform(mode, text string) (string, error) {
	switch mode {
	case "lower":
		return strings.ToLower(text), nil
	case "upper":
		return strings.ToUpper(text), nil
	case "sentence":
		return capitalize(strings.ToLower(text)), nil
	case "title":
		words := strings.Fields(text)
		for i, w := range words {
			words[i] = capitalize(strings.ToLower(w))
		}
		return strings.Join(words, " "), nil
	case "camel":
		words := splitWords(text)
		for i, w := range words {
			if i == 0 {
				words[i] = strings.ToLower(w)
			} else {
				words[i] = capitalize(w)
			}
		}
		return strings.Join(words, ""), nil
	case "pascal":
		words := splitWords(text)
		for i, w := range words {
			words[i] = capitalize(w)
		}
		return strings.Join(words, ""), nil
	case "snake":
		return strings.Join(splitWords(text), "_"), nil
	case "constant":
		words := splitWords(text)
		for i, w := range words {
			words[i] = strings.ToUpper(w)
		}
		return strings.Join(words, "_"), nil
	case "kebab":
		return strings.Join(splitWords(text), "-"), nil
	case "cobol":
		words := splitWords(text)
		for i, w := range words {
			words[i] = strings.ToUpper(w)
		}
		return strings.Join(words, "-"), nil
	case "train":
		words := splitWords(text)
		for i, w := range words {
			words[i] = capitalize(w)
		}
		return strings.Join(words, "-"), nil
	case "alternating":
		var b strings.Builder
		upper := false
		for _, r := range text {
			if unicode.IsLetter(r) {
				if upper {
					b.WriteRune(unicode.ToUpper(r))
				} else {
					b.WriteRune(unicode.ToLower(r))
				}
				upper = !upper
			} else {
				b.WriteRune(r)
			}
		}
		return b.String(), nil
	case "inverse":
		var b strings.Builder
		for _, r := range text {
			switch {
			case unicode.IsUpper(r):
				b.WriteRune(unicode.ToLower(r))
			case unicode.IsLower(r):
				b.WriteRune(unicode.ToUpper(r))
			default:
				b.WriteRune(r)
			}
		}
		return b.String(), nil
	default:
		return "", NewInputError("unsupported text mode: " + mode)
	}
}
