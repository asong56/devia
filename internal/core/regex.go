package core

import "regexp"

type RegexMatch struct {
	Match  string   `json:"match"`
	Start  int      `json:"start"`
	End    int      `json:"end"`
	Groups []string `json:"groups,omitempty"`
}

// compileRegex turns supported single-letter flags (i,m,s,U) into a Go
// inline flag group, e.g. flags="im" -> "(?im)pattern".
func compileRegex(pattern, flags string) (*regexp.Regexp, error) {
	prefix := ""
	for _, f := range flags {
		switch f {
		case 'i', 'm', 's', 'U':
			prefix += string(f)
		}
	}
	full := pattern
	if prefix != "" {
		full = "(?" + prefix + ")" + pattern
	}
	re, err := regexp.Compile(full)
	if err != nil {
		return nil, NewInputError("invalid regex: " + err.Error())
	}
	return re, nil
}

// RegexTest returns every match of pattern in text, including capture
// groups. Note: Go's regexp is RE2-based, so it does not support
// backreferences or lookaround — trade-off for guaranteed linear-time
// matching (no ReDoS), which is the right default for a tool that
// might run untrusted patterns.
func RegexTest(pattern, flags, text string) ([]RegexMatch, error) {
	re, err := compileRegex(pattern, flags)
	if err != nil {
		return nil, err
	}
	idx := re.FindAllStringSubmatchIndex(text, -1)
	out := make([]RegexMatch, 0, len(idx))
	for _, m := range idx {
		match := RegexMatch{Match: text[m[0]:m[1]], Start: m[0], End: m[1]}
		for g := 2; g < len(m); g += 2 {
			if m[g] < 0 {
				match.Groups = append(match.Groups, "")
				continue
			}
			match.Groups = append(match.Groups, text[m[g]:m[g+1]])
		}
		out = append(out, match)
	}
	return out, nil
}

// RegexReplace replaces all matches with replacement ($1-style group
// references supported, per regexp.ReplaceAllString semantics).
func RegexReplace(pattern, flags, text, replacement string) (string, error) {
	re, err := compileRegex(pattern, flags)
	if err != nil {
		return "", err
	}
	return re.ReplaceAllString(text, replacement), nil
}
