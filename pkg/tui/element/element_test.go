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
			wantH:   1,
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
			if e.style.Width != layout.Fixed(tt.wantW) {
				t.Errorf("width = %v, want Fixed(%d)", e.style.Width, tt.wantW)
			}
			if e.style.Height != layout.Fixed(tt.wantH) {
				t.Errorf("height = %v, want Fixed(%d)", e.style.Height, tt.wantH)
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
	if e.style.Width != layout.Fixed(11) {
		t.Errorf("width = %v, want Fixed(11)", e.style.Width)
	}
	if !e.IsDirty() {
		t.Error("SetText should mark element dirty")
	}
}

func TestElement_TextStyle(t *testing.T) {
	style := tui.NewStyle().Foreground(tui.Red).Bold()
	e := New(
		WithText("Hello"),
		WithTextStyle(style),
	)

	if e.TextStyle() != style {
		t.Errorf("TextStyle() = %v, want %v", e.TextStyle(), style)
	}

	newStyle := tui.NewStyle().Foreground(tui.Blue)
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

// --- Focus Tests ---

func TestElement_WithOnFocus_ImpliesFocusable(t *testing.T) {
	e := New(WithOnFocus(func() {}))

	if !e.IsFocusable() {
		t.Error("WithOnFocus should set focusable = true")
	}
}

func TestElement_WithOnBlur_ImpliesFocusable(t *testing.T) {
	e := New(WithOnBlur(func() {}))

	if !e.IsFocusable() {
		t.Error("WithOnBlur should set focusable = true")
	}
}

func TestElement_WithOnEvent_ImpliesFocusable(t *testing.T) {
	e := New(WithOnEvent(func(tui.Event) bool { return false }))

	if !e.IsFocusable() {
		t.Error("WithOnEvent should set focusable = true")
	}
}

func TestElement_Focus_SetsAndCallsCallback(t *testing.T) {
	focusCalled := false
	e := New(WithOnFocus(func() { focusCalled = true }))

	if e.IsFocused() {
		t.Error("element should not be focused initially")
	}

	e.Focus()

	if !e.IsFocused() {
		t.Error("Focus() should set focused = true")
	}
	if !focusCalled {
		t.Error("Focus() should call onFocus callback")
	}
}

func TestElement_Blur_ClearsAndCallsCallback(t *testing.T) {
	blurCalled := false
	e := New(WithOnBlur(func() { blurCalled = true }))

	// First focus, then blur
	e.Focus()
	e.Blur()

	if e.IsFocused() {
		t.Error("Blur() should set focused = false")
	}
	if !blurCalled {
		t.Error("Blur() should call onBlur callback")
	}
}

func TestElement_Focus_CascadesToChildren(t *testing.T) {
	parent := New(WithOnFocus(func() {}))
	child := New()
	grandchild := New()

	parent.AddChild(child)
	child.AddChild(grandchild)

	parent.Focus()

	if !parent.IsFocused() {
		t.Error("parent should be focused")
	}
	if !child.IsFocused() {
		t.Error("child should be focused when parent is focused")
	}
	if !grandchild.IsFocused() {
		t.Error("grandchild should be focused when parent is focused")
	}
}

func TestElement_Blur_CascadesToChildren(t *testing.T) {
	parent := New(WithOnBlur(func() {}))
	child := New()
	grandchild := New()

	parent.AddChild(child)
	child.AddChild(grandchild)

	// Focus first, then blur
	parent.Focus()
	parent.Blur()

	if parent.IsFocused() {
		t.Error("parent should not be focused after blur")
	}
	if child.IsFocused() {
		t.Error("child should not be focused when parent is blurred")
	}
	if grandchild.IsFocused() {
		t.Error("grandchild should not be focused when parent is blurred")
	}
}

func TestElement_Focus_ChildCallbacksCalled(t *testing.T) {
	parentFocusCalled := false
	childFocusCalled := false

	parent := New(WithOnFocus(func() { parentFocusCalled = true }))
	child := New(WithOnFocus(func() { childFocusCalled = true }))

	parent.AddChild(child)
	parent.Focus()

	if !parentFocusCalled {
		t.Error("parent onFocus should be called")
	}
	if !childFocusCalled {
		t.Error("child onFocus should be called when parent is focused")
	}
}

func TestElement_Blur_ChildCallbacksCalled(t *testing.T) {
	parentBlurCalled := false
	childBlurCalled := false

	parent := New(WithOnBlur(func() { parentBlurCalled = true }))
	child := New(WithOnBlur(func() { childBlurCalled = true }))

	parent.AddChild(child)
	parent.Focus()
	parent.Blur()

	if !parentBlurCalled {
		t.Error("parent onBlur should be called")
	}
	if !childBlurCalled {
		t.Error("child onBlur should be called when parent is blurred")
	}
}

func TestElement_HandleEvent_DelegatesToOnEvent(t *testing.T) {
	type tc struct {
		hasHandler  bool
		handlerRet  bool
		wantHandled bool
	}

	tests := map[string]tc{
		"no handler returns false": {
			hasHandler:  false,
			wantHandled: false,
		},
		"handler returns true": {
			hasHandler:  true,
			handlerRet:  true,
			wantHandled: true,
		},
		"handler returns false": {
			hasHandler:  true,
			handlerRet:  false,
			wantHandled: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var e *Element
			if tt.hasHandler {
				e = New(WithOnEvent(func(tui.Event) bool { return tt.handlerRet }))
			} else {
				e = New()
			}

			event := tui.KeyEvent{Key: tui.KeyEnter}
			handled := e.HandleEvent(event)

			if handled != tt.wantHandled {
				t.Errorf("HandleEvent() = %v, want %v", handled, tt.wantHandled)
			}
		})
	}
}

func TestElement_HandleEvent_ReceivesEvent(t *testing.T) {
	var receivedEvent tui.Event
	e := New(WithOnEvent(func(ev tui.Event) bool {
		receivedEvent = ev
		return true
	}))

	sentEvent := tui.KeyEvent{Key: tui.KeyEnter, Rune: 0}
	e.HandleEvent(sentEvent)

	if receivedEvent != sentEvent {
		t.Error("handler should receive the exact event passed to HandleEvent")
	}
}

func TestElement_NotFocusableByDefault(t *testing.T) {
	e := New()

	if e.IsFocusable() {
		t.Error("element should not be focusable by default")
	}
	if e.IsFocused() {
		t.Error("element should not be focused by default")
	}
}

// --- Child Notification Tests ---

func TestElement_SetOnChildAdded_Callback(t *testing.T) {
	root := New()
	var addedChildren []*Element

	root.SetOnChildAdded(func(child *Element) {
		addedChildren = append(addedChildren, child)
	})

	child := New()
	root.AddChild(child)

	if len(addedChildren) != 1 || addedChildren[0] != child {
		t.Error("onChildAdded should be called when child is added")
	}
}

func TestElement_AddChild_TriggersRootCallback(t *testing.T) {
	root := New()
	middle := New()
	root.AddChild(middle)

	var addedChildren []*Element
	root.SetOnChildAdded(func(child *Element) {
		addedChildren = append(addedChildren, child)
	})

	leaf := New()
	middle.AddChild(leaf)

	if len(addedChildren) != 1 || addedChildren[0] != leaf {
		t.Error("onChildAdded should be called on root when leaf is added to middle")
	}
}

func TestElement_SetOnFocusableAdded_Callback(t *testing.T) {
	root := New()
	var addedFocusables []tui.Focusable

	root.SetOnFocusableAdded(func(f tui.Focusable) {
		addedFocusables = append(addedFocusables, f)
	})

	focusable := New(WithOnFocus(func() {}))
	root.AddChild(focusable)

	if len(addedFocusables) != 1 {
		t.Errorf("onFocusableAdded should be called, got %d calls", len(addedFocusables))
	}
}

func TestElement_SetOnFocusableAdded_NotCalledForNonFocusable(t *testing.T) {
	root := New()
	var addedFocusables []tui.Focusable

	root.SetOnFocusableAdded(func(f tui.Focusable) {
		addedFocusables = append(addedFocusables, f)
	})

	nonFocusable := New()
	root.AddChild(nonFocusable)

	if len(addedFocusables) != 0 {
		t.Error("onFocusableAdded should not be called for non-focusable elements")
	}
}

func TestElement_WalkFocusables(t *testing.T) {
	type tc struct {
		setupTree    func() *Element
		expectedCount int
	}

	tests := map[string]tc{
		"empty tree": {
			setupTree: func() *Element {
				return New()
			},
			expectedCount: 0,
		},
		"root is focusable": {
			setupTree: func() *Element {
				return New(WithOnFocus(func() {}))
			},
			expectedCount: 1,
		},
		"child is focusable": {
			setupTree: func() *Element {
				root := New()
				root.AddChild(New(WithOnFocus(func() {})))
				return root
			},
			expectedCount: 1,
		},
		"multiple focusables in tree": {
			setupTree: func() *Element {
				root := New()
				root.AddChild(New(WithOnFocus(func() {})))
				root.AddChild(New(WithOnBlur(func() {})))
				middle := New()
				middle.AddChild(New(WithOnEvent(func(tui.Event) bool { return false })))
				root.AddChild(middle)
				return root
			},
			expectedCount: 3,
		},
		"mixed focusable and non-focusable": {
			setupTree: func() *Element {
				root := New()
				root.AddChild(New(WithOnFocus(func() {})))
				root.AddChild(New()) // non-focusable
				root.AddChild(New(WithOnBlur(func() {})))
				return root
			},
			expectedCount: 2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			root := tt.setupTree()
			var found []tui.Focusable

			root.WalkFocusables(func(f tui.Focusable) {
				found = append(found, f)
			})

			if len(found) != tt.expectedCount {
				t.Errorf("WalkFocusables found %d, want %d", len(found), tt.expectedCount)
			}
		})
	}
}
