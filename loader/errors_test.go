package loader

import (
	"errors"
	"testing"
)

func TestErrorsUnwrap(t *testing.T) {
	first := errors.New("first")
	second := errors.New("second")
	err := &Errors{Errs: []error{first, second}}

	if got, want := err.Error(), "could not compile: first"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
	if !errors.Is(err, second) {
		t.Fatal("errors.Is did not find an error in Errors.Errs")
	}
}
