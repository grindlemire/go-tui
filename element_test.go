package tui

import (
	"testing"
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
			check:   func(e *Element) bool { return e.style.Width == Fixed(100) },
			message: "WithWidth should set fixed width",
		},
		"WithHeight": {
			opts:    []Option{WithHeight(50)},
			check:   func(e *Element) bool { return e.style.Height == Fixed(50) },
			message: "WithHeight should set fixed height",
		},
		"WithSize": {
			opts: []Option{WithSize(80, 40)},
			check: func(e *Element) bool {
				return e.style.Width == Fixed(80) && e.style.Height == Fixed(40)
			},
			message: "WithSize should set both dimensions",
		},
		"WithDirection": {
			opts:    []Option{WithDirection(Column)},
			check:   func(e *Element) bool { return e.style.Direction == Column },
			message: "WithDirection should set direction",
		},
		"WithBorder": {
			opts:    []Option{WithBorder(BorderRounded)},
			check:   func(e *Element) bool { return e.border == BorderRounded },
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
	parent := New(WithSize(100, 80), WithDirection(Row))
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

	newStyle := DefaultLayoutStyle()
	newStyle.Width = Fixed(200)
	e.SetStyle(newStyle)

	if e.Style().Width != Fixed(200) {
		t.Errorf("SetStyle did not update style, Width = %+v", e.Style().Width)
	}
	if !e.IsDirty() {
		t.Error("SetStyle should mark element dirty")
	}
}

func TestElement_ImplementsLayoutable(t *testing.T) {
	// Compile-time check
	var _ Layoutable = (*Element)(nil)

	// Runtime check that methods work
	e := New(WithSize(50, 50), WithPadding(5))
	e.Calculate(100, 100)

	if e.LayoutStyle().Width != Fixed(50) {
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

func TestElement_WithText_SetsIntrinsicSize(t *testing.T) {
	type tc struct {
		content string
		wantW   int
		wantH   int
	}

	tests := map[string]tc{
		"simple text": {
			content: "Hello",
			wantW:   5,
			wantH:   1,
		},
		"empty text": {
			content: "",
			wantW:   0,
			wantH:   0,
		},
		"text with spaces": {
			content: "Hello World",
			wantW:   11,
			wantH:   1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New(WithText(tt.content))
			// WithText leaves Width/Height as Auto; IntrinsicSize computes
			// the correct dimensions including padding and border.
			w, h := e.IntrinsicSize()
			if w != tt.wantW {
				t.Errorf("IntrinsicSize width = %d, want %d", w, tt.wantW)
			}
			if h != tt.wantH {
				t.Errorf("IntrinsicSize height = %d, want %d", h, tt.wantH)
			}
			if e.Text() != tt.content {
				t.Errorf("Text() = %q, want %q", e.Text(), tt.content)
			}
		})
	}
}

func TestElement_SetText_UpdatesWidthAndMarksDirty(t *testing.T) {
	e := New(WithText("Hi"))
	e.dirty = false // Clear dirty flag

	e.SetText("Hello World")

	if e.Text() != "Hello World" {
		t.Errorf("Text() = %q, want %q", e.Text(), "Hello World")
	}
	// IntrinsicSize should reflect the new text dimensions
	w, _ := e.IntrinsicSize()
	if w != 11 {
		t.Errorf("IntrinsicSize width = %d, want 11", w)
	}
	if !e.IsDirty() {
		t.Error("SetText should mark element dirty")
	}
}

func TestElement_TextStyle(t *testing.T) {
	style := NewStyle().Foreground(Red).Bold()
	e := New(
		WithText("Hello"),
		WithTextStyle(style),
	)

	if e.TextStyle() != style {
		t.Errorf("TextStyle() = %v, want %v", e.TextStyle(), style)
	}

	newStyle := NewStyle().Foreground(Blue)
	e.SetTextStyle(newStyle)

	if e.TextStyle() != newStyle {
		t.Errorf("SetTextStyle() did not update, got %v, want %v", e.TextStyle(), newStyle)
	}
}

func TestElement_TextAlign(t *testing.T) {
	type tc struct {
		align TextAlign
	}

	tests := map[string]tc{
		"left":   {align: TextAlignLeft},
		"center": {align: TextAlignCenter},
		"right":  {align: TextAlignRight},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New(
				WithText("Test"),
				WithTextAlign(tt.align),
			)

			if e.TextAlign() != tt.align {
				t.Errorf("TextAlign() = %v, want %v", e.TextAlign(), tt.align)
			}
		})
	}

	// Test SetTextAlign
	e := New(WithText("Test"))
	e.SetTextAlign(TextAlignRight)
	if e.TextAlign() != TextAlignRight {
		t.Errorf("SetTextAlign() did not update, got %v, want %v", e.TextAlign(), TextAlignRight)
	}
}

func TestElement_VisualProperties(t *testing.T) {
	e := New(
		WithBorder(BorderRounded),
		WithBorderStyle(NewStyle().Foreground(Cyan)),
		WithBackground(NewStyle().Background(Blue)),
	)

	if e.Border() != BorderRounded {
		t.Error("Border() should return the border style")
	}

	if e.BorderStyle().Fg != Cyan {
		t.Error("BorderStyle() should return the border color style")
	}

	if e.Background() == nil || e.Background().Bg != Blue {
		t.Error("Background() should return the background style")
	}

	// Test setters
	e.SetBorder(BorderDouble)
	if e.Border() != BorderDouble {
		t.Error("SetBorder() should update border style")
	}

	e.SetBorderStyle(NewStyle().Foreground(Red))
	if e.BorderStyle().Fg != Red {
		t.Error("SetBorderStyle() should update border color style")
	}

	e.SetBackground(nil)
	if e.Background() != nil {
		t.Error("SetBackground(nil) should clear background")
	}
}
