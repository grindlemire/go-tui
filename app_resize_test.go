package tui

import (
	"testing"
)

// trackingTerminal wraps MockTerminal to track render operations.
// It counts Clear() calls to distinguish full vs diff renders.
type trackingTerminal struct {
	*MockTerminal
	clearCount int
	flushCount int
}

func newTrackingTerminal(width, height int) *trackingTerminal {
	return &trackingTerminal{
		MockTerminal: NewMockTerminal(width, height),
	}
}

func (t *trackingTerminal) Clear() {
	t.clearCount++
	t.MockTerminal.Clear()
}

func (t *trackingTerminal) Flush(changes []CellChange) {
	t.flushCount++
	t.MockTerminal.Flush(changes)
}

// testableApp is a helper to create an App with a tracking terminal for testing.
func testableApp(width, height int) (*App, *trackingTerminal) {
	term := newTrackingTerminal(width, height)
	buffer := NewBuffer(width, height)
	focus := NewFocusManager()

	app := &App{
		terminal: nil, // We'll use renderWithTerminal helper
		buffer:   buffer,
		focus:    focus,
	}

	return app, term
}

// renderWithTerminal renders the app using the given terminal (for testing).
func renderWithTerminal(app *App, term Terminal) {
	width, height := term.Size()

	// Clear buffer
	app.buffer.Clear()

	// If root exists, render the element tree
	if app.root != nil {
		app.root.Render(app.buffer, width, height)
	}

	// Use full redraw after resize to clear artifacts, otherwise use diff-based render
	if app.needsFullRedraw {
		RenderFull(term, app.buffer)
		app.needsFullRedraw = false
	} else {
		Render(term, app.buffer)
	}
}

func TestApp_DispatchResizeEvent_SetsNeedsFullRedraw(t *testing.T) {
	type tc struct {
		initialWidth  int
		initialHeight int
		resizeWidth   int
		resizeHeight  int
		hasRoot       bool
	}

	tests := map[string]tc{
		"resize sets flag": {
			initialWidth:  80,
			initialHeight: 24,
			resizeWidth:   100,
			resizeHeight:  30,
			hasRoot:       false,
		},
		"resize with root sets flag": {
			initialWidth:  80,
			initialHeight: 24,
			resizeWidth:   100,
			resizeHeight:  30,
			hasRoot:       true,
		},
		"shrink sets flag": {
			initialWidth:  100,
			initialHeight: 50,
			resizeWidth:   60,
			resizeHeight:  20,
			hasRoot:       true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buffer := NewBuffer(tt.initialWidth, tt.initialHeight)
			app := &App{
				focus:  NewFocusManager(),
				buffer: buffer,
			}

			if tt.hasRoot {
				mockRoot := newMockRenderable()
				app.SetRoot(mockRoot)
			}

			// Flag should initially be false
			if app.needsFullRedraw {
				t.Error("needsFullRedraw should initially be false")
			}

			// Dispatch resize event
			event := ResizeEvent{Width: tt.resizeWidth, Height: tt.resizeHeight}
			handled := app.Dispatch(event)

			if !handled {
				t.Error("Dispatch(ResizeEvent) should return true")
			}

			// Flag should now be true
			if !app.needsFullRedraw {
				t.Error("needsFullRedraw should be true after resize event")
			}
		})
	}
}

func TestApp_Render_ClearsNeedsFullRedrawFlag(t *testing.T) {
	type tc struct {
		description string
	}

	tests := map[string]tc{
		"flag is cleared after render": {
			description: "needsFullRedraw should be false after Render()",
		},
	}

	for name := range tests {
		t.Run(name, func(t *testing.T) {
			app, term := testableApp(80, 24)

			// Set flag manually
			app.needsFullRedraw = true

			// Render should clear the flag
			renderWithTerminal(app, term)

			if app.needsFullRedraw {
				t.Error("needsFullRedraw should be false after Render()")
			}
		})
	}
}

func TestApp_Render_UsesFullRedrawWhenFlagSet(t *testing.T) {
	type tc struct {
		setFlag          bool
		expectClearCount int
		description      string
	}

	tests := map[string]tc{
		"full redraw when flag set": {
			setFlag:          true,
			expectClearCount: 1, // RenderFull calls Clear()
			description:      "should call Clear() for full redraw",
		},
		"diff render when flag not set": {
			setFlag:          false,
			expectClearCount: 0, // Render() does not call Clear()
			description:      "should not call Clear() for diff render",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app, term := testableApp(80, 24)
			app.needsFullRedraw = tt.setFlag

			renderWithTerminal(app, term)

			if term.clearCount != tt.expectClearCount {
				t.Errorf("Clear() count = %d, want %d (%s)", term.clearCount, tt.expectClearCount, tt.description)
			}
		})
	}
}

func TestApp_MultipleRenders_OnlyOneFullRedraw(t *testing.T) {
	type tc struct {
		renderCount      int
		expectClearCount int
	}

	tests := map[string]tc{
		"three renders after resize": {
			renderCount:      3,
			expectClearCount: 1, // Only first render should be full
		},
		"five renders after resize": {
			renderCount:      5,
			expectClearCount: 1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app, term := testableApp(80, 24)

			// Dispatch resize to set flag
			app.Dispatch(ResizeEvent{Width: 100, Height: 30})

			// Render multiple times
			for i := 0; i < tt.renderCount; i++ {
				renderWithTerminal(app, term)
			}

			if term.clearCount != tt.expectClearCount {
				t.Errorf("Clear() count = %d after %d renders, want %d",
					term.clearCount, tt.renderCount, tt.expectClearCount)
			}
		})
	}
}

func TestApp_DispatchNonResizeEvent_DoesNotSetFlag(t *testing.T) {
	type tc struct {
		event Event
	}

	tests := map[string]tc{
		"key enter event": {
			event: KeyEvent{Key: KeyEnter},
		},
		"key tab event": {
			event: KeyEvent{Key: KeyTab},
		},
		"key escape event": {
			event: KeyEvent{Key: KeyEscape},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app := &App{
				focus:  NewFocusManager(),
				buffer: NewBuffer(80, 24),
			}

			// Flag should be false before
			if app.needsFullRedraw {
				t.Error("needsFullRedraw should initially be false")
			}

			// Dispatch non-resize event
			app.Dispatch(tt.event)

			// Flag should still be false
			if app.needsFullRedraw {
				t.Error("needsFullRedraw should still be false after non-resize event")
			}
		})
	}
}

func TestApp_DispatchResizeEvent_InlineWidthChange_InvalidatesLayout(t *testing.T) {
	app := &App{
		focus:          NewFocusManager(),
		buffer:         NewBuffer(80, 3),
		inlineHeight:   3,
		inlineStartRow: 21,
		inlineLayout:   newInlineLayoutState(21),
	}
	app.inlineLayout.visibleRows = 2
	app.inlineLayout.contentStartRow = 19

	handled := app.Dispatch(ResizeEvent{Width: 100, Height: 24})
	if !handled {
		t.Fatal("Dispatch(ResizeEvent) should return true")
	}

	if app.inlineLayout.valid {
		t.Fatalf("inline layout should be invalidated after width change: %+v", app.inlineLayout)
	}
}

func TestApp_DispatchResizeEvent_InlineHeightChange_KeepsLayoutValid(t *testing.T) {
	app := &App{
		focus:          NewFocusManager(),
		buffer:         NewBuffer(80, 3),
		inlineHeight:   3,
		inlineStartRow: 21,
		inlineLayout:   newInlineLayoutState(21),
	}
	app.inlineLayout.visibleRows = 2
	app.inlineLayout.contentStartRow = 19

	handled := app.Dispatch(ResizeEvent{Width: 80, Height: 30})
	if !handled {
		t.Fatal("Dispatch(ResizeEvent) should return true")
	}

	if !app.inlineLayout.valid {
		t.Fatalf("inline layout should remain valid when only height changes: %+v", app.inlineLayout)
	}
}
