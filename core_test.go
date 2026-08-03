package core

import (
	"errors"
	"testing"
)

func TestCodeOf(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"input error", NewInputError("bad input"), CodeInput},
		{"not found error", NewNotFoundError("missing"), CodeNotFound},
		{"plain error defaults to CodeError", errors.New("boom"), CodeError},
		{"nil-wrapped plain error still defaults", errors.New(""), CodeError},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := CodeOf(c.err); got != c.want {
				t.Errorf("CodeOf(%v) = %d, want %d", c.err, got, c.want)
			}
		})
	}
}

func TestCodedErrorUnwrap(t *testing.T) {
	inner := errors.New("underlying problem")
	err := &CodedError{Code: CodeInput, Err: inner}

	if err.Error() != "underlying problem" {
		t.Errorf("Error() = %q, want %q", err.Error(), "underlying problem")
	}
	if !errors.Is(err, inner) {
		t.Error("errors.Is should see through CodedError to the wrapped error")
	}
}

func TestNewInputErrorAndNotFoundError(t *testing.T) {
	if got := CodeOf(NewInputError("x")); got != CodeInput {
		t.Errorf("NewInputError should carry CodeInput, got %d", got)
	}
	if got := CodeOf(NewNotFoundError("x")); got != CodeNotFound {
		t.Errorf("NewNotFoundError should carry CodeNotFound, got %d", got)
	}
}
