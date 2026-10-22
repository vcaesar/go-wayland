package cursor

import (
	"errors"
	"testing"
)

func TestMultiError(t *testing.T) {
	var e MultiError
	e.Add(nil)
	if e.Err() != nil {
		t.Fatalf("Err() = %v, want nil", e.Err())
	}

	e.Add(errors.New("a"))
	e.Add(errors.New("b"))
	if e.Err() == nil {
		t.Fatal("Err() = nil, want error")
	}
	if got, want := e.Error(), "a\nb\n"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}
