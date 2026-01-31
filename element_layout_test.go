package tui

import (
	"testing"
)

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
