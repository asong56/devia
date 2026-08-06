// Package core contains all business logic for devia, with zero
// dependencies outside the Go standard library. It is deliberately
// decoupled from both the CLI (main package) and the HTTP API
// (server.go) so the same tested logic backs both entry points.
package core

import "errors"

// Standard exit / status codes shared by the CLI and the HTTP API.
// The CLI uses these directly as process exit codes; the HTTP API
// maps them to HTTP status codes (see server.go).
const (
	CodeOK       = 0 // success
	CodeError    = 1 // internal / unexpected error
	CodeUsage    = 2 // bad command-line usage (missing/invalid flags)
	CodeInput    = 3 // well-formed invocation, but invalid input data
	CodeNotFound = 4 // referenced file / resource does not exist
)

// CodedError carries one of the standard codes above alongside a
// human-readable message. Every user-facing error returned by this
// package is a *CodedError so callers can make consistent decisions
// (which exit code to use, which HTTP status to return) without
// string-matching error messages.
type CodedError struct {
	Code int
	Err  error
}

func (e *CodedError) Error() string { return e.Err.Error() }
func (e *CodedError) Unwrap() error { return e.Err }

// NewInputError wraps msg as a CodeInput error (bad user-supplied data:
// invalid JSON, invalid base64, invalid regex, etc).
func NewInputError(msg string) error {
	return &CodedError{Code: CodeInput, Err: errors.New(msg)}
}

// NewNotFoundError wraps msg as a CodeNotFound error (missing file).
func NewNotFoundError(msg string) error {
	return &CodedError{Code: CodeNotFound, Err: errors.New(msg)}
}

// CodeOf extracts the standard code from err, defaulting to CodeError
// for plain errors that didn't originate in this package (e.g. raw OS
// errors bubbled up from os.ReadFile).
func CodeOf(err error) int {
	var ce *CodedError
	if errors.As(err, &ce) {
		return ce.Code
	}
	return CodeError
}
