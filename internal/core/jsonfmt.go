package core

import (
	"encoding/json"
	"strings"
)

// decodeSingleJSONValue decodes exactly one top-level JSON value from
// text and errors on anything but whitespace left over afterward.
//
// json.Decoder.Decode alone does NOT do this — it's built for reading
// a stream of concatenated values, so `{"a":1} garbage` or `{"a":1}
// {"b":2}` decodes successfully as just the first value, silently
// discarding the rest. json.Unmarshal (what JSONValidate below uses)
// already rejects trailing data, so without this check `devia json
// validate` would correctly reject trailing garbage while `devia json
// format`/`minify` silently accepted and reformatted only the first
// fragment — same input, contradictory verdicts from two commands in
// the same tool. dec.More() is the standard way to ask "is there
// another value queued in the stream", which at the top level after
// one full Decode is exactly "is there trailing non-whitespace data".
func decodeSingleJSONValue(text string) (interface{}, error) {
	var v interface{}
	dec := json.NewDecoder(strings.NewReader(text))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return nil, NewInputError("invalid JSON: " + err.Error())
	}
	if dec.More() {
		return nil, NewInputError("invalid JSON: unexpected data after the top-level value")
	}
	return v, nil
}

// JSONFormat pretty-prints text with the given indent string (default
// two spaces). json.Number is used during decode so large integers and
// exact decimal literals survive the round-trip unchanged.
func JSONFormat(text, indent string) (string, error) {
	v, err := decodeSingleJSONValue(text)
	if err != nil {
		return "", err
	}
	if indent == "" {
		indent = "  "
	}
	b, err := json.MarshalIndent(v, "", indent)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// JSONMinify removes all insignificant whitespace.
func JSONMinify(text string) (string, error) {
	v, err := decodeSingleJSONValue(text)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// JSONValidate reports only whether text is syntactically valid JSON —
// useful in scripts as `devia json validate < f.json && echo ok`.
func JSONValidate(text string) error {
	var v interface{}
	if err := json.Unmarshal([]byte(text), &v); err != nil {
		return NewInputError("invalid JSON: " + err.Error())
	}
	return nil
}
