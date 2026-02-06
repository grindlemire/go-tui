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

func TestHandleClicks_Hit(t *testing.T) {
	ref := NewRef()
	called := false

	// Create element and attach to ref
	el := New(WithWidth(10), WithHeight(5))
	el.layout.Rect = NewRect(0, 0, 10, 5)
	ref.Set(el)

	// Simulate click inside element
	handled := HandleClicks(
		MouseEvent{Button: MouseLeft, Action: MousePress, X: 5, Y: 2},
		Click(ref, func() { called = true }),
	)

	if !handled {
		t.Fatal("expected click to be handled")
	}
	if !called {
		t.Fatal("handler not called")
	}
}

func TestHandleClicks_Miss(t *testing.T) {
	ref := NewRef()
	called := false

	// Create element and attach to ref
	el := New(WithWidth(10), WithHeight(5))
	el.layout.Rect = NewRect(0, 0, 10, 5)
	ref.Set(el)

	// Simulate click outside element
	handled := HandleClicks(
		MouseEvent{Button: MouseLeft, Action: MousePress, X: 15, Y: 10},
		Click(ref, func() { called = true }),
	)

	if handled {
		t.Fatal("expected click to not be handled")
	}
	if called {
		t.Fatal("handler should not be called")
	}
}

func TestHandleClicks_NilRef(t *testing.T) {
	ref := NewRef()
	// Don't attach element - ref.El() will be nil

	handled := HandleClicks(
		MouseEvent{Button: MouseLeft, Action: MousePress, X: 5, Y: 2},
		Click(ref, func() { t.Fatal("should not be called") }),
	)

	if handled {
		t.Fatal("expected nil ref to not handle click")
	}
}

func TestHandleClicks_MultipleBindings(t *testing.T) {
	ref1 := NewRef()
	ref2 := NewRef()
	var clicked string

	// Create elements
	el1 := New(WithWidth(10), WithHeight(5))
	el1.layout.Rect = NewRect(0, 0, 10, 5)
	ref1.Set(el1)

	el2 := New(WithWidth(10), WithHeight(5))
	el2.layout.Rect = NewRect(20, 0, 10, 5)
	ref2.Set(el2)

	// Click on second element
	handled := HandleClicks(
		MouseEvent{Button: MouseLeft, Action: MousePress, X: 25, Y: 2},
		Click(ref1, func() { clicked = "first" }),
		Click(ref2, func() { clicked = "second" }),
	)

	if !handled {
		t.Fatal("expected click to be handled")
	}
	if clicked != "second" {
		t.Fatalf("expected 'second', got '%s'", clicked)
	}
}

func TestHandleClicks_IgnoresNonLeftClick(t *testing.T) {
	ref := NewRef()
	el := New(WithWidth(10), WithHeight(5))
	el.layout.Rect = NewRect(0, 0, 10, 5)
	ref.Set(el)

	// Right click
	handled := HandleClicks(
		MouseEvent{Button: MouseRight, Action: MousePress, X: 5, Y: 2},
		Click(ref, func() { t.Fatal("should not be called") }),
	)

	if handled {
		t.Fatal("expected right click to not be handled")
	}
}

func TestHandleClicks_IgnoresRelease(t *testing.T) {
	ref := NewRef()
	el := New(WithWidth(10), WithHeight(5))
	el.layout.Rect = NewRect(0, 0, 10, 5)
	ref.Set(el)

	// Mouse release (not press)
	handled := HandleClicks(
		MouseEvent{Button: MouseLeft, Action: MouseRelease, X: 5, Y: 2},
		Click(ref, func() { t.Fatal("should not be called") }),
	)

	if handled {
		t.Fatal("expected release to not be handled")
	}
}
