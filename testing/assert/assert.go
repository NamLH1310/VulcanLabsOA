package assert

import "testing"

func NoError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func Equal[T comparable](t *testing.T, expect, got T) {
	if expect != got {
		t.Errorf("Mismatch expect: %v, got %v\n", expect, got)
	}
}
