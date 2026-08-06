package core

import (
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"strconv"
	"strings"
)

// EscapeJSON returns the JSON-escaped form of text, without the
// surrounding quotes json.Marshal normally adds (so the output is a
// drop-in fragment, not a full JSON string literal).
func EscapeJSON(text string) (string, error) {
	b, err := json.Marshal(text)
	if err != nil {
		return "", err
	}
	s := string(b)
	return s[1 : len(s)-1], nil
}

// UnescapeJSON reverses EscapeJSON.
func UnescapeJSON(text string) (string, error) {
	var s string
	if err := json.Unmarshal([]byte(`"`+text+`"`), &s); err != nil {
		return "", NewInputError("invalid JSON-escaped string: " + err.Error())
	}
	return s, nil
}

func EscapeURL(text string, path bool) string {
	if path {
		return url.PathEscape(text)
	}
	return url.QueryEscape(text)
}

func UnescapeURL(text string, path bool) (string, error) {
	var s string
	var err error
	if path {
		s, err = url.PathUnescape(text)
	} else {
		s, err = url.QueryUnescape(text)
	}
	if err != nil {
		return "", NewInputError("invalid URL-escaped string: " + err.Error())
	}
	return s, nil
}

func EscapeHTML(text string) string   { return html.EscapeString(text) }
func UnescapeHTML(text string) string { return html.UnescapeString(text) }

// EscapeUnicode converts non-ASCII runes to \uXXXX escapes, encoding
// characters outside the Basic Multilingual Plane as UTF-16 surrogate
// pairs (the same convention JSON string literals use).
func EscapeUnicode(text string) string {
	var b strings.Builder
	for _, r := range text {
		if r <= 127 {
			b.WriteRune(r)
			continue
		}
		if r > 0xFFFF {
			r1, r2 := utf16Pair(r)
			fmt.Fprintf(&b, "\\u%04x\\u%04x", r1, r2)
		} else {
			fmt.Fprintf(&b, "\\u%04x", r)
		}
	}
	return b.String()
}

func utf16Pair(r rune) (rune, rune) {
	r -= 0x10000
	return 0xD800 + (r >> 10), 0xDC00 + (r & 0x3FF)
}

// UnescapeUnicode reverses \uXXXX escapes, recombining surrogate pairs
// back into a single rune.
func UnescapeUnicode(text string) (string, error) {
	var b strings.Builder
	i := 0
	for i < len(text) {
		if text[i] == '\\' && i+1 < len(text) && text[i+1] == 'u' {
			if i+6 > len(text) {
				return "", NewInputError("truncated \\u escape near position " + strconv.Itoa(i))
			}
			code, err := strconv.ParseInt(text[i+2:i+6], 16, 32)
			if err != nil {
				return "", NewInputError("invalid \\u escape: " + text[i:i+6])
			}
			r := rune(code)
			i += 6
			if r >= 0xD800 && r <= 0xDBFF && i+6 <= len(text) && text[i] == '\\' && text[i+1] == 'u' {
				code2, err2 := strconv.ParseInt(text[i+2:i+6], 16, 32)
				if err2 == nil && code2 >= 0xDC00 && code2 <= 0xDFFF {
					r = ((r - 0xD800) << 10) + (rune(code2) - 0xDC00) + 0x10000
					i += 6
				}
			}
			b.WriteRune(r)
		} else {
			b.WriteByte(text[i])
			i++
		}
	}
	return b.String(), nil
}
