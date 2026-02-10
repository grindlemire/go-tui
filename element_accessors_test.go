package tui

import (
	"testing"
)

// --- OnUpdate Hook Tests ---

func TestElement_SetOnUpdate_CalledDuringRender(t *testing.T) {
	updateCalled := false
	e := New(WithSize(10, 10))
	e.SetOnUpdate(func() {
		updateCalled = true
	})

	buf := NewBuffer(20, 20)
	e.Render(buf, 20, 20)

	if !updateCalled {
		t.Error("onUpdate hook should be called during Render()")
	}
}

func TestElement_Render_NilOnUpdateDoesNotPanic(t *testing.T) {
	// Create an element without an onUpdate hook
	e := New(WithSize(10, 10))

	buf := NewBuffer(20, 20)

	// This should not panic
	e.Render(buf, 20, 20)
}

func TestElement_WithOnUpdate_SetsHook(t *testing.T) {
	updateCalled := false
	e := New(
		WithSize(10, 10),
		WithOnUpdate(func() {
			updateCalled = true
		}),
	)

	buf := NewBuffer(20, 20)
	e.Render(buf, 20, 20)

	if !updateCalled {
		t.Error("WithOnUpdate should set the onUpdate hook")
	}
}

func TestElement_OnUpdate_CalledOnEachRender(t *testing.T) {
	callCount := 0
	e := New(WithSize(10, 10))
	e.SetOnUpdate(func() {
		callCount++
	})

	buf := NewBuffer(20, 20)

	// Render multiple times
	e.Render(buf, 20, 20)
	e.Render(buf, 20, 20)
	e.Render(buf, 20, 20)

	if callCount != 3 {
		t.Errorf("onUpdate should be called on each render, got %d calls, want 3", callCount)
	}
}

// --- Focusable Tests ---

func TestElement_WithFocusable(t *testing.T) {
	type tc struct {
		focusable bool
	}

	tests := map[string]tc{
		"focusable true":  {focusable: true},
		"focusable false": {focusable: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New(WithFocusable(tt.focusable))

			if e.IsFocusable() != tt.focusable {
				t.Errorf("WithFocusable(%v) = %v, want %v", tt.focusable, e.IsFocusable(), tt.focusable)
			}
		})
	}
}

func TestElement_SetFocusable(t *testing.T) {
	e := New()

	if e.IsFocusable() {
		t.Error("element should not be focusable by default")
	}

	e.SetFocusable(true)
	if !e.IsFocusable() {
		t.Error("SetFocusable(true) should make element focusable")
	}

	e.SetFocusable(false)
	if e.IsFocusable() {
		t.Error("SetFocusable(false) should make element not focusable")
	}
}

func TestElement_RemoveAllChildren(t *testing.T) {
	parent := New()
	child1 := New()
	child2 := New()
	child3 := New()

	parent.AddChild(child1, child2, child3)

	if len(parent.Children()) != 3 {
		t.Fatalf("setup failed: expected 3 children, got %d", len(parent.Children()))
	}

	// Clear dirty flag to test that RemoveAllChildren marks dirty
	parent.dirty = false
	MarkDirty() // Reset global dirty
	_ = checkAndClearDirty()

	parent.RemoveAllChildren()

	if len(parent.Children()) != 0 {
		t.Errorf("RemoveAllChildren should remove all children, got %d", len(parent.Children()))
	}

	if child1.Parent() != nil {
		t.Error("removed child1's parent should be nil")
	}
	if child2.Parent() != nil {
		t.Error("removed child2's parent should be nil")
	}
	if child3.Parent() != nil {
		t.Error("removed child3's parent should be nil")
	}

	if !parent.IsDirty() {
		t.Error("RemoveAllChildren should mark parent dirty")
	}
}

func TestElement_RemoveAllChildren_Empty(t *testing.T) {
	parent := New()

	// Should not panic on empty element
	parent.RemoveAllChildren()

	if len(parent.Children()) != 0 {
		t.Error("RemoveAllChildren on empty element should result in no children")
	}
}

// --- Global Dirty Flag Tests ---

func TestElement_MarkDirty_SetsGlobalDirtyFlag(t *testing.T) {
	// Reset global dirty flag
	_ = checkAndClearDirty()

	e := New()
	e.dirty = false // Clear local dirty flag

	e.MarkDirty()

	if !checkAndClearDirty() {
		t.Error("MarkDirty should set the global dirty flag")
	}
}

func TestElement_ScrollBy_MarksDirty(t *testing.T) {
	// Reset global dirty flag
	_ = checkAndClearDirty()

	e := New(
		WithHeight(10),
		WithScrollable(ScrollVertical),
		WithDirection(Column),
	)
	// Set up content that exceeds viewport
	for i := 0; i < 20; i++ {
		e.AddChild(New(WithHeight(1)))
	}

	// Render to compute content bounds (scrollable content needs this)
	buf := NewBuffer(80, 25)
	e.Render(buf, 80, 10)

	// Clear dirty flags
	e.dirty = false
	_ = checkAndClearDirty()

	e.ScrollBy(0, 5)

	if !checkAndClearDirty() {
		t.Error("ScrollBy should mark the global dirty flag")
	}
}

func TestElement_SetText_MarksDirty(t *testing.T) {
	// Reset global dirty flag
	_ = checkAndClearDirty()

	e := New(WithText("hello"))

	// Clear dirty flags
	e.dirty = false
	_ = checkAndClearDirty()

	e.SetText("world")

	if !checkAndClearDirty() {
		t.Error("SetText should mark the global dirty flag")
	}
}

func TestElement_AddChild_MarksDirty(t *testing.T) {
	// Reset global dirty flag
	_ = checkAndClearDirty()

	parent := New()

	// Clear dirty flags
	parent.dirty = false
	_ = checkAndClearDirty()

	child := New()
	parent.AddChild(child)

	if !checkAndClearDirty() {
		t.Error("AddChild should mark the global dirty flag")
	}
}

