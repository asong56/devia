package core

import (
	"errors"
	"testing"
)

func TestCodeOf_CodedError(t *testing.T) {
	err := NewInputError("bad input")
	if got := CodeOf(err); got != CodeInput {
		t.Errorf("CodeOf(NewInputError(...)) = %d, want %d", got, CodeInput)
	}

	err = NewNotFoundError("missing")
	if got := CodeOf(err); got != CodeNotFound {
		t.Errorf("CodeOf(NewNotFoundError(...)) = %d, want %d", got, CodeNotFound)
	}
}

func TestCodeOf_PlainError(t *testing.T) {
	// A raw error (e.g. bubbled up from os.ReadFile for a permission
	// error, not a missing-file error) must default to CodeError, not
	// panic or silently misreport as CodeOK.
	err := errors.New("something went wrong")
	if got := CodeOf(err); got != CodeError {
		t.Errorf("CodeOf(plain error) = %d, want %d", got, CodeError)
	}
}

func TestCodeOf_WrappedCodedError(t *testing.T) {
	// CodeOf uses errors.As, so a CodedError wrapped by another error
	// (e.g. via fmt.Errorf("%w", ...) in calling code) must still
	// resolve to its original code, not fall back to CodeError.
	inner := NewInputError("bad input")
	wrapped := &wrapErr{msg: "context", err: inner}
	if got := CodeOf(wrapped); got != CodeInput {
		t.Errorf("CodeOf(wrapped CodedError) = %d, want %d", got, CodeInput)
	}
}

func TestCodedError_ErrorAndUnwrap(t *testing.T) {
	inner := errors.New("disk full")
	ce := &CodedError{Code: CodeError, Err: inner}
	if ce.Error() != "disk full" {
		t.Errorf("Error() = %q, want %q", ce.Error(), "disk full")
	}
	if !errors.Is(ce, inner) {
		t.Error("errors.Is(ce, inner) = false, want true (Unwrap must expose the inner error)")
	}
}

type wrapErr struct {
	msg string
	err error
}

func (w *wrapErr) Error() string { return w.msg }
func (w *wrapErr) Unwrap() error { return w.err }
