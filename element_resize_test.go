package tui

import "testing"

// TestElement_SetHeight_TriggersRelayout verifies that SetHeight on a retained
// element changes its laid-out height on the next frame.
func TestElement_SetHeight_TriggersRelayout(t *testing.T) {
	app := newTestApp(40, 24)

	// A retained child with an explicit starting height inside a column root.
	child := New(WithHeight(1), WithWidth(10))
	root := New(WithDirection(Column))
	root.AddChild(child)
	app.SetRoot(root)

	app.Render()
	if got := child.Rect().Height; got != 1 {
		t.Fatalf("initial child height = %d, want 1", got)
	}

	// Grow the retained child; the next frame must re-lay it out.
	child.SetHeight(Fixed(8))
	app.Render()
	if got := child.Rect().Height; got != 8 {
		t.Fatalf("child height after SetHeight = %d, want 8", got)
	}
}

// TestElement_SetWidth_TriggersRelayout verifies that SetWidth on a retained
// element changes its laid-out width on the next frame.
func TestElement_SetWidth_TriggersRelayout(t *testing.T) {
	app := newTestApp(40, 24)

	child := New(WithHeight(1), WithWidth(5))
	root := New(WithDirection(Column))
	root.AddChild(child)
	app.SetRoot(root)

	app.Render()
	if got := child.Rect().Width; got != 5 {
		t.Fatalf("initial child width = %d, want 5", got)
	}

	child.SetWidth(Fixed(20))
	app.Render()
	if got := child.Rect().Width; got != 20 {
		t.Fatalf("child width after SetWidth = %d, want 20", got)
	}
}

// TestElement_SetHeight_MarksDirty verifies the element flag is set so the
// app re-renders.
func TestElement_SetHeight_MarksDirty(t *testing.T) {
	app := newTestApp(40, 24)
	child := New(WithHeight(1))
	root := New()
	root.AddChild(child)
	app.SetRoot(root)
	app.Render() // clears dirty

	if app.checkAndClearDirty() {
		t.Fatal("app should be clean after render")
	}
	child.SetHeight(Fixed(4))
	if !child.IsDirty() {
		t.Fatal("child should be dirty after SetHeight")
	}
	if !app.checkAndClearDirty() {
		t.Fatal("app should be dirty after SetHeight")
	}
}
