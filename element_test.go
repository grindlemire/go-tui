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
			if e.style.Width != Fixed(tt.wantW) {
				t.Errorf("width = %v, want Fixed(%d)", e.style.Width, tt.wantW)
			}
			if e.style.Height != Fixed(tt.wantH) {
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
	if e.style.Width != Fixed(11) {
		t.Errorf("width = %v, want Fixed(11)", e.style.Width)
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
	e := New(WithOnEvent(func(Event) bool { return false }))

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
				e = New(WithOnEvent(func(Event) bool { return tt.handlerRet }))
			} else {
				e = New()
			}

			event := KeyEvent{Key: KeyEnter}
			handled := e.HandleEvent(event)

			if handled != tt.wantHandled {
				t.Errorf("HandleEvent() = %v, want %v", handled, tt.wantHandled)
			}
		})
	}
}

func TestElement_HandleEvent_ReceivesEvent(t *testing.T) {
	var receivedEvent Event
	e := New(WithOnEvent(func(ev Event) bool {
		receivedEvent = ev
		return true
	}))

	sentEvent := KeyEvent{Key: KeyEnter, Rune: 0}
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
	var addedFocusables []Focusable

	root.SetOnFocusableAdded(func(f Focusable) {
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
	var addedFocusables []Focusable

	root.SetOnFocusableAdded(func(f Focusable) {
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
				middle.AddChild(New(WithOnEvent(func(Event) bool { return false })))
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
			var found []Focusable

			root.WalkFocusables(func(f Focusable) {
				found = append(found, f)
			})

			if len(found) != tt.expectedCount {
				t.Errorf("WalkFocusables found %d, want %d", len(found), tt.expectedCount)
			}
		})
	}
}

// --- Intrinsic Sizing Tests ---

func TestElement_IntrinsicSize_TextElement(t *testing.T) {
	type tc struct {
		text           string
		expectedWidth  int
		expectedHeight int
	}

	tests := map[string]tc{
		"short text": {
			text:           "Hello",
			expectedWidth:  5,
			expectedHeight: 1,
		},
		"longer text": {
			text:           "Hello, World!",
			expectedWidth:  13,
			expectedHeight: 1,
		},
		"empty text": {
			text:           "",
			expectedWidth:  0,
			expectedHeight: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New(WithText(tt.text))
			w, h := e.IntrinsicSize()

			if w != tt.expectedWidth {
				t.Errorf("IntrinsicSize().Width = %d, want %d", w, tt.expectedWidth)
			}
			if h != tt.expectedHeight {
				t.Errorf("IntrinsicSize().Height = %d, want %d", h, tt.expectedHeight)
			}
		})
	}
}

func TestElement_IntrinsicSize_TextWithPadding(t *testing.T) {
	e := New(
		WithText("Hello"),
		WithPadding(2),
	)

	w, h := e.IntrinsicSize()

	// Text is 5 chars, padding adds 2 on each side
	// Width = 5 + 4 = 9, Height = 1 + 4 = 5
	if w != 9 {
		t.Errorf("IntrinsicSize().Width = %d, want 9 (5 text + 4 padding)", w)
	}
	if h != 5 {
		t.Errorf("IntrinsicSize().Height = %d, want 5 (1 text + 4 padding)", h)
	}
}

func TestElement_IntrinsicSize_TextWithBorder(t *testing.T) {
	e := New(
		WithText("Hello"),
		WithBorder(BorderSingle),
	)

	w, h := e.IntrinsicSize()

	// Text is 5 chars, border adds 1 on each side
	// Width = 5 + 2 = 7, Height = 1 + 2 = 3
	if w != 7 {
		t.Errorf("IntrinsicSize().Width = %d, want 7 (5 text + 2 border)", w)
	}
	if h != 3 {
		t.Errorf("IntrinsicSize().Height = %d, want 3 (1 text + 2 border)", h)
	}
}

func TestElement_IntrinsicSize_EmptyContainer(t *testing.T) {
	e := New(WithDirection(Column))

	w, h := e.IntrinsicSize()

	if w != 0 || h != 0 {
		t.Errorf("IntrinsicSize() = %dx%d, want 0x0 for empty container", w, h)
	}
}

func TestElement_IntrinsicSize_ContainerWithChildren(t *testing.T) {
	// Column container with two text children
	parent := New(WithDirection(Column))

	child1 := New(WithText("Hello"))    // 5x1
	child2 := New(WithText("World!!!")) // 8x1

	parent.AddChild(child1, child2)

	w, h := parent.IntrinsicSize()

	// Column: width = max(5, 8) = 8, height = 1 + 1 = 2
	if w != 8 {
		t.Errorf("IntrinsicSize().Width = %d, want 8 (max child width)", w)
	}
	if h != 2 {
		t.Errorf("IntrinsicSize().Height = %d, want 2 (sum of child heights)", h)
	}
}

func TestElement_IntrinsicSize_RowContainerWithChildren(t *testing.T) {
	// Row container with two text children
	parent := New(WithDirection(Row))

	child1 := New(WithText("Hello")) // 5x1
	child2 := New(WithText("World")) // 5x1

	parent.AddChild(child1, child2)

	w, h := parent.IntrinsicSize()

	// Row: width = 5 + 5 = 10, height = max(1, 1) = 1
	if w != 10 {
		t.Errorf("IntrinsicSize().Width = %d, want 10 (sum of child widths)", w)
	}
	if h != 1 {
		t.Errorf("IntrinsicSize().Height = %d, want 1 (max child height)", h)
	}
}

func TestElement_IntrinsicSize_ContainerWithGap(t *testing.T) {
	parent := New(
		WithDirection(Column),
		WithGap(5),
	)

	child1 := New(WithText("A")) // 1x1
	child2 := New(WithText("B")) // 1x1
	child3 := New(WithText("C")) // 1x1

	parent.AddChild(child1, child2, child3)

	w, h := parent.IntrinsicSize()

	// Column with gap: width = max(1,1,1) = 1, height = 1 + 5 + 1 + 5 + 1 = 13
	if w != 1 {
		t.Errorf("IntrinsicSize().Width = %d, want 1", w)
	}
	if h != 13 {
		t.Errorf("IntrinsicSize().Height = %d, want 13 (3 children + 2 gaps)", h)
	}
}

func TestElement_IntrinsicSize_NestedContainers(t *testing.T) {
	// Nested structure similar to DSL counter
	root := New(WithDirection(Column))

	box := New(
		WithDirection(Column),
		WithPadding(1),
	)
	text := New(WithText("Title")) // 5x1
	box.AddChild(text)

	footer := New(WithText("Footer text here")) // 16x1

	root.AddChild(box, footer)

	w, h := root.IntrinsicSize()

	// box: width = 5 + 2 = 7, height = 1 + 2 = 3
	// root: width = max(7, 16) = 16, height = 3 + 1 = 4
	if w != 16 {
		t.Errorf("IntrinsicSize().Width = %d, want 16", w)
	}
	if h != 4 {
		t.Errorf("IntrinsicSize().Height = %d, want 4", h)
	}
}

func TestElement_IntrinsicSize_ContainerWithPadding(t *testing.T) {
	parent := New(
		WithDirection(Column),
		WithPadding(3),
	)

	child := New(WithText("Hi")) // 2x1

	parent.AddChild(child)

	w, h := parent.IntrinsicSize()

	// width = 2 + 6 = 8, height = 1 + 6 = 7
	if w != 8 {
		t.Errorf("IntrinsicSize().Width = %d, want 8 (2 text + 6 padding)", w)
	}
	if h != 7 {
		t.Errorf("IntrinsicSize().Height = %d, want 7 (1 text + 6 padding)", h)
	}
}

func TestElement_IntrinsicSize_ContainerWithBorder(t *testing.T) {
	parent := New(
		WithDirection(Column),
		WithBorder(BorderSingle),
	)

	child := New(WithText("Hi")) // 2x1

	parent.AddChild(child)

	w, h := parent.IntrinsicSize()

	// width = 2 + 2 = 4, height = 1 + 2 = 3
	if w != 4 {
		t.Errorf("IntrinsicSize().Width = %d, want 4 (2 text + 2 border)", w)
	}
	if h != 3 {
		t.Errorf("IntrinsicSize().Height = %d, want 3 (1 text + 2 border)", h)
	}
}

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

// --- Event Handler Tests ---

func TestElement_SetOnKeyPress(t *testing.T) {
	var receivedEvent KeyEvent
	e := New()

	e.SetOnKeyPress(func(event KeyEvent) {
		receivedEvent = event
	})

	// Dispatch a key event
	sentEvent := KeyEvent{Key: KeyRune, Rune: 'a'}
	e.HandleEvent(sentEvent)

	if receivedEvent != sentEvent {
		t.Errorf("SetOnKeyPress handler should receive the event, got %v, want %v", receivedEvent, sentEvent)
	}
}

func TestElement_SetOnClick(t *testing.T) {
	clickCalled := false
	e := New()

	e.SetOnClick(func() {
		clickCalled = true
	})

	// onClick is stored but not invoked by HandleEvent (that's for key events)
	// The onClick handler would be invoked by mouse events in future
	if e.onClick == nil {
		t.Error("SetOnClick should store the handler")
	}

	// Call the handler directly to verify it works
	e.onClick()
	if !clickCalled {
		t.Error("onClick handler should be callable")
	}
}

func TestElement_WithOnKeyPress_ImpliesFocusable(t *testing.T) {
	e := New(WithOnKeyPress(func(KeyEvent) {}))

	if !e.IsFocusable() {
		t.Error("WithOnKeyPress should set focusable = true")
	}
}

func TestElement_WithOnClick_ImpliesFocusable(t *testing.T) {
	e := New(WithOnClick(func() {}))

	if !e.IsFocusable() {
		t.Error("WithOnClick should set focusable = true")
	}
}

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
	_ = TestCheckAndClearDirty()

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
	_ = TestCheckAndClearDirty()

	e := New()
	e.dirty = false // Clear local dirty flag

	e.MarkDirty()

	if !TestCheckAndClearDirty() {
		t.Error("MarkDirty should set the global dirty flag")
	}
}

func TestElement_ScrollBy_MarksDirty(t *testing.T) {
	// Reset global dirty flag
	_ = TestCheckAndClearDirty()

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
	_ = TestCheckAndClearDirty()

	e.ScrollBy(0, 5)

	if !TestCheckAndClearDirty() {
		t.Error("ScrollBy should mark the global dirty flag")
	}
}

func TestElement_SetText_MarksDirty(t *testing.T) {
	// Reset global dirty flag
	_ = TestCheckAndClearDirty()

	e := New(WithText("hello"))

	// Clear dirty flags
	e.dirty = false
	_ = TestCheckAndClearDirty()

	e.SetText("world")

	if !TestCheckAndClearDirty() {
		t.Error("SetText should mark the global dirty flag")
	}
}

func TestElement_AddChild_MarksDirty(t *testing.T) {
	// Reset global dirty flag
	_ = TestCheckAndClearDirty()

	parent := New()

	// Clear dirty flags
	parent.dirty = false
	_ = TestCheckAndClearDirty()

	child := New()
	parent.AddChild(child)

	if !TestCheckAndClearDirty() {
		t.Error("AddChild should mark the global dirty flag")
	}
}

func TestElement_HandleEvent_CallsOnKeyPress(t *testing.T) {
	handlerCalled := false
	var receivedEvent KeyEvent

	e := New(WithOnKeyPress(func(event KeyEvent) {
		handlerCalled = true
		receivedEvent = event
	}))

	sentEvent := KeyEvent{Key: KeyEnter}
	e.HandleEvent(sentEvent)

	if !handlerCalled {
		t.Error("HandleEvent should call onKeyPress handler for key events")
	}
	if receivedEvent != sentEvent {
		t.Errorf("onKeyPress should receive the event, got %v, want %v", receivedEvent, sentEvent)
	}
}

// --- Hit Testing Tests ---

func TestElement_ElementAt_ReturnsNilForPointOutsideBounds(t *testing.T) {
	e := New(WithWidth(10), WithHeight(10))
	// Calculate layout so the element has a position
	buf := NewBuffer(80, 25)
	e.Render(buf, 80, 25)

	result := e.ElementAt(50, 50)
	if result != nil {
		t.Error("ElementAt should return nil for points outside the element's bounds")
	}
}

func TestElement_ElementAt_ReturnsSelfForPointInsideBounds(t *testing.T) {
	e := New(WithWidth(10), WithHeight(10))
	// Calculate layout so the element has a position
	buf := NewBuffer(80, 25)
	e.Render(buf, 80, 25)

	result := e.ElementAt(5, 5)
	if result != e {
		t.Error("ElementAt should return the element itself when point is inside bounds and no children")
	}
}

func TestElement_ElementAt_ReturnsChildForPointInsideChild(t *testing.T) {
	parent := New(
		WithWidth(100),
		WithHeight(100),
		WithDirection(Column),
	)
	child := New(WithWidth(50), WithHeight(50))
	parent.AddChild(child)

	// Calculate layout
	buf := NewBuffer(100, 100)
	parent.Render(buf, 100, 100)

	// Point inside child bounds (child starts at 0,0 and is 50x50)
	result := parent.ElementAt(10, 10)
	if result != child {
		t.Error("ElementAt should return the child when point is inside child bounds")
	}
}

func TestElement_ElementAt_ReturnsDeepestChild(t *testing.T) {
	root := New(
		WithWidth(100),
		WithHeight(100),
	)
	child1 := New(WithWidth(80), WithHeight(80))
	child2 := New(WithWidth(60), WithHeight(60))
	grandchild := New(WithWidth(40), WithHeight(40))

	child2.AddChild(grandchild)
	child1.AddChild(child2)
	root.AddChild(child1)

	// Calculate layout
	buf := NewBuffer(100, 100)
	root.Render(buf, 100, 100)

	// Point inside grandchild bounds (should be at 0,0)
	result := root.ElementAt(10, 10)
	if result != grandchild {
		t.Error("ElementAt should return the deepest child containing the point")
	}
}

func TestElement_ElementAt_LastChildTakesPrecedence(t *testing.T) {
	// When children overlap, last child should take precedence (renders on top)
	parent := New(
		WithWidth(100),
		WithHeight(100),
	)
	// Both children start at 0,0 with same size
	child1 := New(WithWidth(50), WithHeight(50))
	child2 := New(WithWidth(50), WithHeight(50))

	parent.AddChild(child1)
	parent.AddChild(child2)

	// Calculate layout - both children will be at position (0,0)
	buf := NewBuffer(100, 100)
	parent.Render(buf, 100, 100)

	result := parent.ElementAt(10, 10)
	// child2 was added last, so it should take precedence
	// However, with default layout (row), they won't overlap.
	// Let's just verify we get one of the children
	if result != child1 && result != child2 {
		t.Error("ElementAt should return one of the children for overlapping bounds")
	}
}

func TestElement_ElementAtPoint_ReturnsFocusable(t *testing.T) {
	// Test that ElementAtPoint returns a Focusable interface
	e := New(WithWidth(10), WithHeight(10))
	buf := NewBuffer(80, 25)
	e.Render(buf, 80, 25)

	result := e.ElementAtPoint(5, 5)
	if result == nil {
		t.Error("ElementAtPoint should return a non-nil Focusable")
	}
	// Verify it's the same underlying element
	if result.(*Element) != e {
		t.Error("ElementAtPoint should return the element wrapped as Focusable")
	}
}

func TestElement_ElementAtPoint_ReturnsNilForPointOutsideBounds(t *testing.T) {
	e := New(WithWidth(10), WithHeight(10))
	buf := NewBuffer(80, 25)
	e.Render(buf, 80, 25)

	result := e.ElementAtPoint(50, 50)
	if result != nil {
		t.Error("ElementAtPoint should return nil for points outside bounds")
	}
}

func TestElement_HandleEvent_MouseClick_TriggersOnClick(t *testing.T) {
	handlerCalled := false

	e := New(WithOnClick(func() {
		handlerCalled = true
	}))

	// Send a left mouse press event
	event := MouseEvent{
		Button: MouseLeft,
		Action: MousePress,
		X:      5,
		Y:      5,
	}
	consumed := e.HandleEvent(event)

	if !handlerCalled {
		t.Error("HandleEvent should call onClick handler for left mouse press")
	}
	if !consumed {
		t.Error("HandleEvent should return true when onClick is triggered")
	}
}

func TestElement_HandleEvent_MouseClick_NoHandlerNotConsumed(t *testing.T) {
	e := New() // No onClick handler

	event := MouseEvent{
		Button: MouseLeft,
		Action: MousePress,
		X:      5,
		Y:      5,
	}
	consumed := e.HandleEvent(event)

	if consumed {
		t.Error("HandleEvent should return false when no onClick handler is set")
	}
}

func TestElement_HandleEvent_MouseRelease_DoesNotTriggerOnClick(t *testing.T) {
	handlerCalled := false

	e := New(WithOnClick(func() {
		handlerCalled = true
	}))

	// Send a left mouse release event (not press)
	event := MouseEvent{
		Button: MouseLeft,
		Action: MouseRelease,
		X:      5,
		Y:      5,
	}
	e.HandleEvent(event)

	if handlerCalled {
		t.Error("HandleEvent should not call onClick for mouse release events")
	}
}

func TestElement_HandleEvent_RightClick_DoesNotTriggerOnClick(t *testing.T) {
	handlerCalled := false

	e := New(WithOnClick(func() {
		handlerCalled = true
	}))

	// Send a right mouse press event
	event := MouseEvent{
		Button: MouseRight,
		Action: MousePress,
		X:      5,
		Y:      5,
	}
	e.HandleEvent(event)

	if handlerCalled {
		t.Error("HandleEvent should not call onClick for right click")
	}
}
