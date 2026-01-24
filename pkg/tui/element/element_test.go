package element

import (
	"testing"

	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
)

func TestNew_DefaultValues(t *testing.T) {
	e := New()

	// Should have Auto dimensions by default
	if !e.style.Width.IsAuto() {
		t.Error("New() should have Auto width")
	}
	if !e.style.Height.IsAuto() {
		t.Error("New() should have Auto height")
	}

	// Should be dirty
	if !e.IsDirty() {
		t.Error("New() should be dirty")
	}

	// Should have no children
	if len(e.Children()) != 0 {
		t.Errorf("New() should have no children, got %d", len(e.Children()))
	}

	// Should have no parent
	if e.Parent() != nil {
		t.Error("New() should have no parent")
	}
}

func TestNew_WithOptions(t *testing.T) {
	type tc struct {
		name    string
		opts    []Option
		check   func(*Element) bool
		message string
	}

	tests := map[string]tc{
		"WithWidth": {
			opts:    []Option{WithWidth(100)},
			check:   func(e *Element) bool { return e.style.Width == layout.Fixed(100) },
			message: "WithWidth should set fixed width",
		},
		"WithHeight": {
			opts:    []Option{WithHeight(50)},
			check:   func(e *Element) bool { return e.style.Height == layout.Fixed(50) },
			message: "WithHeight should set fixed height",
		},
		"WithSize": {
			opts: []Option{WithSize(80, 40)},
			check: func(e *Element) bool {
				return e.style.Width == layout.Fixed(80) && e.style.Height == layout.Fixed(40)
			},
			message: "WithSize should set both dimensions",
		},
		"WithDirection": {
			opts:    []Option{WithDirection(layout.Column)},
			check:   func(e *Element) bool { return e.style.Direction == layout.Column },
			message: "WithDirection should set direction",
		},
		"WithBorder": {
			opts:    []Option{WithBorder(tui.BorderRounded)},
			check:   func(e *Element) bool { return e.border == tui.BorderRounded },
			message: "WithBorder should set border style",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New(tt.opts...)
			if !tt.check(e) {
				t.Error(tt.message)
			}
		})
	}
}

func TestElement_AddChild(t *testing.T) {
	parent := New()
	child1 := New()
	child2 := New()

	// Clear dirty flag to test that AddChild marks dirty
	parent.dirty = false

	parent.AddChild(child1, child2)

	if len(parent.Children()) != 2 {
		t.Errorf("AddChild: len(Children) = %d, want 2", len(parent.Children()))
	}
	if parent.Children()[0] != child1 {
		t.Error("AddChild: first child mismatch")
	}
	if parent.Children()[1] != child2 {
		t.Error("AddChild: second child mismatch")
	}
	if child1.Parent() != parent {
		t.Error("AddChild: child1.Parent() not set")
	}
	if child2.Parent() != parent {
		t.Error("AddChild: child2.Parent() not set")
	}
	if !parent.IsDirty() {
		t.Error("AddChild should mark parent dirty")
	}
}

func TestElement_RemoveChild(t *testing.T) {
	type tc struct {
		setup       func() (*Element, *Element, *Element)
		removeChild func(*Element, *Element, *Element) *Element
		expectFound bool
		expectLen   int
	}

	tests := map[string]tc{
		"remove existing child": {
			setup: func() (*Element, *Element, *Element) {
				parent := New()
				child1 := New()
				child2 := New()
				parent.AddChild(child1, child2)
				parent.dirty = false
				return parent, child1, child2
			},
			removeChild: func(parent, child1, child2 *Element) *Element { return child1 },
			expectFound: true,
			expectLen:   1,
		},
		"remove non-existent child": {
			setup: func() (*Element, *Element, *Element) {
				parent := New()
				child1 := New()
				parent.AddChild(child1)
				parent.dirty = false
				otherChild := New()
				return parent, child1, otherChild
			},
			removeChild: func(parent, child1, otherChild *Element) *Element { return otherChild },
			expectFound: false,
			expectLen:   1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			parent, child1, child2 := tt.setup()
			toRemove := tt.removeChild(parent, child1, child2)
			wasDirty := parent.IsDirty()

			found := parent.RemoveChild(toRemove)

			if found != tt.expectFound {
				t.Errorf("RemoveChild returned %v, want %v", found, tt.expectFound)
			}
			if len(parent.Children()) != tt.expectLen {
				t.Errorf("len(Children) = %d, want %d", len(parent.Children()), tt.expectLen)
			}
			if tt.expectFound {
				if toRemove.Parent() != nil {
					t.Error("removed child's parent should be nil")
				}
				if !parent.IsDirty() {
					t.Error("RemoveChild should mark parent dirty")
				}
			} else {
				if parent.IsDirty() != wasDirty {
					t.Error("RemoveChild should not change dirty state when child not found")
				}
			}
		})
	}
}

func TestElement_MarkDirty_PropagatesUp(t *testing.T) {
	// Build tree: root -> middle -> leaf
	root := New()
	middle := New()
	leaf := New()
	root.AddChild(middle)
	middle.AddChild(leaf)

	// Clear all dirty flags
	root.dirty = false
	middle.dirty = false
	leaf.dirty = false

	// Mark leaf dirty
	leaf.MarkDirty()

	if !leaf.IsDirty() {
		t.Error("leaf should be dirty")
	}
	if !middle.IsDirty() {
		t.Error("middle should be dirty (propagated from leaf)")
	}
	if !root.IsDirty() {
		t.Error("root should be dirty (propagated from leaf)")
	}
}

func TestElement_MarkDirty_StopsAtAlreadyDirty(t *testing.T) {
	// Build tree: root -> middle -> leaf
	root := New()
	middle := New()
	leaf := New()
	root.AddChild(middle)
	middle.AddChild(leaf)

	// Clear all dirty flags
	root.dirty = false
	middle.dirty = false
	leaf.dirty = false

	// Mark middle dirty first
	middle.dirty = true

	// Mark leaf dirty - should stop at middle since it's already dirty
	leaf.MarkDirty()

	if !leaf.IsDirty() {
		t.Error("leaf should be dirty")
	}
	if !middle.IsDirty() {
		t.Error("middle should still be dirty")
	}
	// Root should still be clean because propagation stopped at middle
	if root.IsDirty() {
		t.Error("root should still be clean (propagation stopped at middle)")
	}
}

func TestElement_Calculate(t *testing.T) {
	parent := New(WithSize(100, 80), WithDirection(layout.Row))
	child1 := New(WithWidth(30))
	child2 := New(WithFlexGrow(1))

	parent.AddChild(child1, child2)
	parent.Calculate(200, 200)

	// Parent should have correct dimensions
	if parent.Rect().Width != 100 {
		t.Errorf("parent.Rect().Width = %d, want 100", parent.Rect().Width)
	}
	if parent.Rect().Height != 80 {
		t.Errorf("parent.Rect().Height = %d, want 80", parent.Rect().Height)
	}

	// Child1 should have fixed width
	if child1.Rect().Width != 30 {
		t.Errorf("child1.Rect().Width = %d, want 30", child1.Rect().Width)
	}

	// Child2 should grow to fill remaining space
	if child2.Rect().Width != 70 {
		t.Errorf("child2.Rect().Width = %d, want 70", child2.Rect().Width)
	}

	// All should be clean after Calculate
	if parent.IsDirty() || child1.IsDirty() || child2.IsDirty() {
		t.Error("all elements should be clean after Calculate")
	}
}

func TestElement_SetStyle(t *testing.T) {
	e := New()
	e.dirty = false

	newStyle := layout.DefaultStyle()
	newStyle.Width = layout.Fixed(200)
	e.SetStyle(newStyle)

	if e.Style().Width != layout.Fixed(200) {
		t.Errorf("SetStyle did not update style, Width = %+v", e.Style().Width)
	}
	if !e.IsDirty() {
		t.Error("SetStyle should mark element dirty")
	}
}

func TestElement_ImplementsLayoutable(t *testing.T) {
	// Compile-time check
	var _ layout.Layoutable = (*Element)(nil)

	// Runtime check that methods work
	e := New(WithSize(50, 50), WithPadding(5))
	e.Calculate(100, 100)

	if e.LayoutStyle().Width != layout.Fixed(50) {
		t.Error("LayoutStyle() should return the style")
	}

	l := e.GetLayout()
	if l.Rect.Width != 50 || l.Rect.Height != 50 {
		t.Errorf("GetLayout().Rect = %dx%d, want 50x50", l.Rect.Width, l.Rect.Height)
	}

	// ContentRect should be inset by padding
	if l.ContentRect.Width != 40 || l.ContentRect.Height != 40 {
		t.Errorf("GetLayout().ContentRect = %dx%d, want 40x40",
			l.ContentRect.Width, l.ContentRect.Height)
	}
}

func TestElement_VisualProperties(t *testing.T) {
	e := New(
		WithBorder(tui.BorderRounded),
		WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan)),
		WithBackground(tui.NewStyle().Background(tui.Blue)),
	)

	if e.Border() != tui.BorderRounded {
		t.Error("Border() should return the border style")
	}

	if e.BorderStyle().Fg != tui.Cyan {
		t.Error("BorderStyle() should return the border color style")
	}

	if e.Background() == nil || e.Background().Bg != tui.Blue {
		t.Error("Background() should return the background style")
	}

	// Test setters
	e.SetBorder(tui.BorderDouble)
	if e.Border() != tui.BorderDouble {
		t.Error("SetBorder() should update border style")
	}

	e.SetBorderStyle(tui.NewStyle().Foreground(tui.Red))
	if e.BorderStyle().Fg != tui.Red {
		t.Error("SetBorderStyle() should update border color style")
	}

	e.SetBackground(nil)
	if e.Background() != nil {
		t.Error("SetBackground(nil) should clear background")
	}
}
