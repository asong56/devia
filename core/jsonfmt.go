package core

import (
	"encoding/json"
	"strings"
)

// JSONFormat pretty-prints text with the given indent string (default
// two spaces). json.Number is used during decode so large integers and
// exact decimal literals survive the round-trip unchanged.
func JSONFormat(text, indent string) (string, error) {
	var v interface{}
	dec := json.NewDecoder(strings.NewReader(text))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return "", NewInputError("invalid JSON: " + err.Error())
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
	var v interface{}
	dec := json.NewDecoder(strings.NewReader(text))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return "", NewInputError("invalid JSON: " + err.Error())
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
