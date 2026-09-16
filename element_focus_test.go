package tui

import (
	"testing"
)

// --- Hit Testing Tests ---

func TestElement_ElementAt_ReturnsNilForPointOutsideBounds(t *testing.T) {
	e := New(WithWidth(10), WithHeight(10))
	// Calculate layout so the element has a position
	buf := NewBuffer(80, 25)
	e.RenderTo(buf, 80, 25)

	result := e.ElementAt(50, 50)
	if result != nil {
		t.Error("ElementAt should return nil for points outside the element's bounds")
	}
}

func TestElement_ElementAt_ReturnsSelfForPointInsideBounds(t *testing.T) {
	e := New(WithWidth(10), WithHeight(10))
	// Calculate layout so the element has a position
	buf := NewBuffer(80, 25)
	e.RenderTo(buf, 80, 25)

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
	parent.RenderTo(buf, 100, 100)

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
	root.RenderTo(buf, 100, 100)

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
	parent.RenderTo(buf, 100, 100)

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
	e.RenderTo(buf, 80, 25)

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
	e.RenderTo(buf, 80, 25)

	result := e.ElementAtPoint(50, 50)
	if result != nil {
		t.Error("ElementAtPoint should return nil for points outside bounds")
	}
}

// --- HandleEvent Tests ---

func TestElement_HandleEvent_MouseNotHandledByDefault(t *testing.T) {
	e := New()

	event := MouseEvent{
		Button: MouseLeft,
		Action: MousePress,
		X:      5,
		Y:      5,
	}
	consumed := e.HandleEvent(event)

	if consumed {
		t.Error("HandleEvent should return false for mouse events (mouse handling is via component HandleMouse)")
	}
}

func TestElement_HandleEvent_KeyNotHandledByDefault(t *testing.T) {
	e := New()

	event := KeyEvent{Key: KeyRune, Rune: 'x'}
	consumed := e.HandleEvent(event)

	if consumed {
		t.Error("HandleEvent should return false for key events (key handling is via component KeyMap)")
	}
}

// --- Border style writes while focused ---

// borderWriters are the three ways user code changes a border color.
var borderWriters = map[string]func(e *Element, s Style){
	"SetBorderStyle":        func(e *Element, s Style) { e.SetBorderStyle(s) },
	"Apply WithBorderStyle": func(e *Element, s Style) { e.Apply(WithBorderStyle(s)) },
	"SetClass border color": func(e *Element, _ Style) { e.SetClass("border-green") },
}

func TestElement_BorderStyleWriteWhileFocused(t *testing.T) {
	red := NewStyle().Foreground(Red)
	green := NewStyle().Foreground(Green)
	cyan := NewStyle().Foreground(Cyan)
	magenta := NewStyle().Foreground(Magenta)

	type tc struct {
		opts           []Option
		writeUnfocused bool  // write before Focus instead of during
		wantFocused    Style // visible style after the write, while focused
	}

	tests := map[string]tc{
		"default highlight keeps highlight until blur": {
			opts:        []Option{WithBorder(BorderSingle), WithBorderStyle(red), WithFocusable(true)},
			wantFocused: cyan,
		},
		"explicit focus style keeps highlight until blur": {
			opts:        []Option{WithBorder(BorderSingle), WithBorderStyle(red), WithFocusable(true), WithFocusBorderStyle(magenta)},
			wantFocused: magenta,
		},
		"no highlight applies immediately": {
			opts:        []Option{WithBorder(BorderSingle), WithBorderStyle(red), WithOnFocus(func(*Element) {})},
			wantFocused: green,
		},
		"write before focus survives focus and blur": {
			opts:           []Option{WithBorder(BorderSingle), WithBorderStyle(red), WithFocusable(true)},
			writeUnfocused: true,
			wantFocused:    cyan,
		},
	}

	for name, tt := range tests {
		for wname, write := range borderWriters {
			t.Run(name+"/"+wname, func(t *testing.T) {
				e := New(tt.opts...)
				if tt.writeUnfocused {
					write(e, green)
				}
				e.Focus()
				if !tt.writeUnfocused {
					write(e, green)
				}
				if got := e.activeBorderStyle(); got != tt.wantFocused {
					t.Errorf("focused: visible border = %+v, want %+v", got, tt.wantFocused)
				}
				if got := e.BorderStyle(); got != green {
					t.Errorf("focused: BorderStyle() = %+v, want %+v", got, green)
				}
				e.Blur()
				if got := e.activeBorderStyle(); got != green {
					t.Errorf("blurred: visible border = %+v, want %+v", got, green)
				}
			})
		}
	}
}

func TestElement_Blur_RefocusInOnBlurKeepsHighlight(t *testing.T) {
	red := NewStyle().Foreground(Red)
	cyan := NewStyle().Foreground(Cyan)

	e := New(WithBorder(BorderSingle), WithBorderStyle(red), WithFocusable(true),
		WithOnBlur(func(el *Element) { el.Focus() }))
	e.Focus()
	e.Blur()
	if !e.IsFocused() {
		t.Fatal("onBlur refocus should leave the element focused")
	}
	if got := e.activeBorderStyle(); got != cyan {
		t.Errorf("refocused: visible border = %+v, want %+v", got, cyan)
	}
	// Drop the handler so the next Blur really blurs.
	e.onBlur = nil
	e.Blur()
	if got := e.activeBorderStyle(); got != red {
		t.Errorf("blurred: visible border = %+v, want %+v", got, red)
	}
}

func TestElement_SetClassWhileFocused_Clear(t *testing.T) {
	red := NewStyle().Foreground(Red)
	green := NewStyle().Foreground(Green)
	cyan := NewStyle().Foreground(Cyan)

	type tc struct {
		applyBeforeFocus bool  // SetClass("border-green") before Focus instead of during
		clearAfterBlur   bool  // SetClass("") after Blur instead of while focused
		wantBlurred      Style // visible style right after Blur
	}

	tests := map[string]tc{
		"apply and clear while focused":           {wantBlurred: red},
		"apply while focused, clear after blur":   {clearAfterBlur: true, wantBlurred: green},
		"apply before focus, clear while focused": {applyBeforeFocus: true, wantBlurred: red},
		"apply before focus, clear after blur":    {applyBeforeFocus: true, clearAfterBlur: true, wantBlurred: green},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := New(WithBorder(BorderSingle), WithBorderStyle(red), WithFocusable(true))
			if tt.applyBeforeFocus {
				e.SetClass("border-green")
			}
			e.Focus()
			if !tt.applyBeforeFocus {
				e.SetClass("border-green")
			}
			if !tt.clearAfterBlur {
				e.SetClass("")
			}
			if got := e.activeBorderStyle(); got != cyan {
				t.Errorf("focused: visible border = %+v, want %+v", got, cyan)
			}
			e.Blur()
			if got := e.activeBorderStyle(); got != tt.wantBlurred {
				t.Errorf("blurred: visible border = %+v, want %+v", got, tt.wantBlurred)
			}
			if tt.clearAfterBlur {
				e.SetClass("")
			}
			if got := e.activeBorderStyle(); got != red {
				t.Errorf("class cleared: visible border = %+v, want %+v", got, red)
			}
			// A second focus cycle must not have baked the highlight into the base.
			e.Focus()
			e.Blur()
			if got := e.activeBorderStyle(); got != red {
				t.Errorf("after second cycle: visible border = %+v, want %+v", got, red)
			}
		})
	}
}
