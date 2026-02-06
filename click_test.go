package tui

import "testing"

func TestClick_Type(t *testing.T) {
	ref := NewRef()
	called := false
	c := Click(ref, func() { called = true })

	if c.Ref != ref {
		t.Fatal("ref not set")
	}

	c.Fn()
	if !called {
		t.Fatal("fn not called")
	}
}
