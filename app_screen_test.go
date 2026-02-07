package tui

import (
	"testing"
)

func TestApp_IsInAlternateScreen_InitiallyFalse(t *testing.T) {
	type tc struct {
		inlineHeight int
	}

	tests := map[string]tc{
		"full screen mode": {
			inlineHeight: 0,
		},
		"inline mode": {
			inlineHeight: 5,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app := &App{
				buffer:       NewBuffer(80, 24),
				focus:        NewFocusManager(),
				inlineHeight: tt.inlineHeight,
			}

			if app.IsInAlternateScreen() {
				t.Error("IsInAlternateScreen() should be false initially")
			}
		})
	}
}

func TestApp_EnterAlternateScreen_Basic(t *testing.T) {
	type tc struct {
		inlineHeight       int
		inlineStartRow     int
		expectedInAltState bool
	}

	tests := map[string]tc{
		"from inline mode": {
			inlineHeight:       5,
			inlineStartRow:     19,
			expectedInAltState: true,
		},
		"from full screen mode": {
			inlineHeight:       0,
			inlineStartRow:     0,
			expectedInAltState: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			term := NewMockTerminal(80, 24)
			app := &App{
				terminal:       term,
				buffer:         NewBuffer(80, tt.inlineHeight),
				focus:          NewFocusManager(),
				inlineHeight:   tt.inlineHeight,
				inlineStartRow: tt.inlineStartRow,
			}

			err := app.EnterAlternateScreen()
			if err != nil {
				t.Fatalf("EnterAlternateScreen() returned error: %v", err)
			}

			if !app.IsInAlternateScreen() {
				t.Error("IsInAlternateScreen() should be true after EnterAlternateScreen()")
			}

			if !term.IsInAltScreen() {
				t.Error("terminal should be in alternate screen")
			}

			if term.AltScreenEnterCount() != 1 {
				t.Errorf("EnterAltScreen() call count = %d, want 1", term.AltScreenEnterCount())
			}
		})
	}
}

func TestApp_EnterAlternateScreen_SavesInlineState(t *testing.T) {
	type tc struct {
		inlineHeight   int
		inlineStartRow int
	}

	tests := map[string]tc{
		"saves inline state": {
			inlineHeight:   5,
			inlineStartRow: 19,
		},
		"saves zero inline state": {
			inlineHeight:   0,
			inlineStartRow: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			term := NewMockTerminal(80, 24)
			app := &App{
				terminal:       term,
				buffer:         NewBuffer(80, 24),
				focus:          NewFocusManager(),
				inlineHeight:   tt.inlineHeight,
				inlineStartRow: tt.inlineStartRow,
			}

			_ = app.EnterAlternateScreen()

			if app.savedInlineHeight != tt.inlineHeight {
				t.Errorf("savedInlineHeight = %d, want %d", app.savedInlineHeight, tt.inlineHeight)
			}

			if app.savedInlineStartRow != tt.inlineStartRow {
				t.Errorf("savedInlineStartRow = %d, want %d", app.savedInlineStartRow, tt.inlineStartRow)
			}
		})
	}
}

func TestApp_EnterAlternateScreen_Idempotent(t *testing.T) {
	type tc struct {
		callCount int
	}

	tests := map[string]tc{
		"double enter": {
			callCount: 2,
		},
		"triple enter": {
			callCount: 3,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			term := NewMockTerminal(80, 24)
			app := &App{
				terminal:     term,
				buffer:       NewBuffer(80, 24),
				focus:        NewFocusManager(),
				inlineHeight: 5,
			}

			for i := 0; i < tt.callCount; i++ {
				_ = app.EnterAlternateScreen()
			}

			// Should only call terminal EnterAltScreen once
			if term.AltScreenEnterCount() != 1 {
				t.Errorf("EnterAltScreen() call count = %d, want 1", term.AltScreenEnterCount())
			}
		})
	}
}

func TestApp_ExitAlternateScreen_Basic(t *testing.T) {
	type tc struct {
		savedInlineHeight   int
		savedInlineStartRow int
	}

	tests := map[string]tc{
		"restore to inline mode": {
			savedInlineHeight:   5,
			savedInlineStartRow: 19,
		},
		"restore to full screen mode": {
			savedInlineHeight:   0,
			savedInlineStartRow: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			term := NewMockTerminal(80, 24)
			app := &App{
				terminal:            term,
				buffer:              NewBuffer(80, 24),
				focus:               NewFocusManager(),
				inAlternateScreen:   true,
				savedInlineHeight:   tt.savedInlineHeight,
				savedInlineStartRow: tt.savedInlineStartRow,
			}

			err := app.ExitAlternateScreen()
			if err != nil {
				t.Fatalf("ExitAlternateScreen() returned error: %v", err)
			}

			if app.IsInAlternateScreen() {
				t.Error("IsInAlternateScreen() should be false after ExitAlternateScreen()")
			}

			if app.inlineHeight != tt.savedInlineHeight {
				t.Errorf("inlineHeight = %d, want %d", app.inlineHeight, tt.savedInlineHeight)
			}
		})
	}
}

func TestApp_ExitAlternateScreen_Idempotent(t *testing.T) {
	type tc struct {
		callCount int
	}

	tests := map[string]tc{
		"double exit": {
			callCount: 2,
		},
		"triple exit": {
			callCount: 3,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			term := NewMockTerminal(80, 24)
			app := &App{
				terminal:          term,
				buffer:            NewBuffer(80, 24),
				focus:             NewFocusManager(),
				inAlternateScreen: true,
				savedInlineHeight: 5,
			}

			for i := 0; i < tt.callCount; i++ {
				_ = app.ExitAlternateScreen()
			}

			// Should only call terminal ExitAltScreen once
			if term.AltScreenExitCount() != 1 {
				t.Errorf("ExitAltScreen() call count = %d, want 1", term.AltScreenExitCount())
			}
		})
	}
}

func TestApp_ExitAlternateScreen_WhenNotInAlternate(t *testing.T) {
	type tc struct {
		description string
	}

	tests := map[string]tc{
		"no-op when not in alternate": {
			description: "should not call terminal ExitAltScreen",
		},
	}

	for name := range tests {
		t.Run(name, func(t *testing.T) {
			term := NewMockTerminal(80, 24)
			app := &App{
				terminal:          term,
				buffer:            NewBuffer(80, 24),
				focus:             NewFocusManager(),
				inAlternateScreen: false,
			}

			_ = app.ExitAlternateScreen()

			if term.AltScreenExitCount() != 0 {
				t.Errorf("ExitAltScreen() call count = %d, want 0", term.AltScreenExitCount())
			}
		})
	}
}

func TestApp_EnterAlternateScreen_ResizesBuffer(t *testing.T) {
	type tc struct {
		initialInlineHeight int
		termWidth           int
		termHeight          int
	}

	tests := map[string]tc{
		"buffer resized to full terminal": {
			initialInlineHeight: 5,
			termWidth:           80,
			termHeight:          24,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			term := NewMockTerminal(tt.termWidth, tt.termHeight)
			buffer := NewBuffer(tt.termWidth, tt.initialInlineHeight)
			app := &App{
				terminal:     term,
				buffer:       buffer,
				focus:        NewFocusManager(),
				inlineHeight: tt.initialInlineHeight,
			}

			_ = app.EnterAlternateScreen()

			if app.buffer.Width() != tt.termWidth {
				t.Errorf("buffer width = %d, want %d", app.buffer.Width(), tt.termWidth)
			}

			if app.buffer.Height() != tt.termHeight {
				t.Errorf("buffer height = %d, want %d", app.buffer.Height(), tt.termHeight)
			}
		})
	}
}

func TestApp_EnterAlternateScreen_SetsNeedsFullRedraw(t *testing.T) {
	type tc struct {
		description string
	}

	tests := map[string]tc{
		"sets needsFullRedraw flag": {
			description: "should set needsFullRedraw to true",
		},
	}

	for name := range tests {
		t.Run(name, func(t *testing.T) {
			term := NewMockTerminal(80, 24)
			app := &App{
				terminal:     term,
				buffer:       NewBuffer(80, 24),
				focus:        NewFocusManager(),
				inlineHeight: 5,
			}

			_ = app.EnterAlternateScreen()

			if !app.needsFullRedraw {
				t.Error("needsFullRedraw should be true after EnterAlternateScreen()")
			}
		})
	}
}

func TestApp_ExitAlternateScreen_SetsNeedsFullRedraw(t *testing.T) {
	type tc struct {
		description string
	}

	tests := map[string]tc{
		"sets needsFullRedraw flag": {
			description: "should set needsFullRedraw to true",
		},
	}

	for name := range tests {
		t.Run(name, func(t *testing.T) {
			term := NewMockTerminal(80, 24)
			app := &App{
				terminal:          term,
				buffer:            NewBuffer(80, 24),
				focus:             NewFocusManager(),
				inAlternateScreen: true,
				savedInlineHeight: 5,
			}

			_ = app.ExitAlternateScreen()

			if !app.needsFullRedraw {
				t.Error("needsFullRedraw should be true after ExitAlternateScreen()")
			}
		})
	}
}

func TestApp_RoundTrip_InlineToAlternateToInline(t *testing.T) {
	type tc struct {
		inlineHeight   int
		inlineStartRow int
	}

	tests := map[string]tc{
		"inline mode round trip": {
			inlineHeight:   5,
			inlineStartRow: 19,
		},
		"different inline height": {
			inlineHeight:   10,
			inlineStartRow: 14,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			term := NewMockTerminal(80, 24)
			app := &App{
				terminal:       term,
				buffer:         NewBuffer(80, tt.inlineHeight),
				focus:          NewFocusManager(),
				inlineHeight:   tt.inlineHeight,
				inlineStartRow: tt.inlineStartRow,
			}

			// Enter alternate screen
			_ = app.EnterAlternateScreen()

			if !app.IsInAlternateScreen() {
				t.Error("should be in alternate screen after enter")
			}

			// Exit alternate screen
			_ = app.ExitAlternateScreen()

			if app.IsInAlternateScreen() {
				t.Error("should not be in alternate screen after exit")
			}

			// Inline height should be restored
			if app.inlineHeight != tt.inlineHeight {
				t.Errorf("inlineHeight = %d after round trip, want %d", app.inlineHeight, tt.inlineHeight)
			}

			// Terminal should have entered and exited once each
			if term.AltScreenEnterCount() != 1 {
				t.Errorf("EnterAltScreen count = %d, want 1", term.AltScreenEnterCount())
			}

			if term.AltScreenExitCount() != 1 {
				t.Errorf("ExitAltScreen count = %d, want 1", term.AltScreenExitCount())
			}
		})
	}
}

func TestApp_DispatchResize_InAlternateMode(t *testing.T) {
	type tc struct {
		newWidth  int
		newHeight int
	}

	tests := map[string]tc{
		"resize in alternate mode": {
			newWidth:  100,
			newHeight: 30,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			term := NewMockTerminal(80, 24)
			app := &App{
				terminal:          term,
				buffer:            NewBuffer(80, 24),
				focus:             NewFocusManager(),
				inAlternateScreen: true,
				savedInlineHeight: 5, // Was in inline mode before
			}

			event := ResizeEvent{Width: tt.newWidth, Height: tt.newHeight}
			handled := app.Dispatch(event)

			if !handled {
				t.Error("Dispatch(ResizeEvent) should return true")
			}

			// Buffer should be resized to full terminal (not inline height)
			if app.buffer.Width() != tt.newWidth {
				t.Errorf("buffer width = %d, want %d", app.buffer.Width(), tt.newWidth)
			}

			if app.buffer.Height() != tt.newHeight {
				t.Errorf("buffer height = %d, want %d", app.buffer.Height(), tt.newHeight)
			}
		})
	}
}
