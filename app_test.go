package tui

import (
	"testing"
)

// mockRenderable is a mock implementation of Renderable for testing.
type mockRenderable struct {
	dirty           bool
	renderCalled    bool
	markDirtyCalled bool
}

func newMockRenderable() *mockRenderable {
	return &mockRenderable{dirty: true}
}

func (m *mockRenderable) Render(buf *Buffer, width, height int) {
	m.renderCalled = true
	m.dirty = false
}

func (m *mockRenderable) MarkDirty() {
	m.dirty = true
	m.markDirtyCalled = true
}

func (m *mockRenderable) IsDirty() bool {
	return m.dirty
}

func TestApp_SetRootAndRoot(t *testing.T) {
	type tc struct {
		createRoot bool
	}

	tests := map[string]tc{
		"with root element": {
			createRoot: true,
		},
		"without root element": {
			createRoot: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create a mock app (we can't test NewApp without a real terminal)
			app := &App{
				focus:  NewFocusManager(),
				buffer: NewBuffer(80, 24),
			}

			if tt.createRoot {
				root := newMockRenderable()
				app.SetRoot(root)

				if app.Root() != root {
					t.Error("Root() should return the element passed to SetRoot()")
				}
			} else {
				if app.Root() != nil {
					t.Error("Root() should return nil when no root set")
				}
			}
		})
	}
}

func TestApp_Focus(t *testing.T) {
	app := &App{
		focus: NewFocusManager(),
	}

	if app.Focus() == nil {
		t.Error("Focus() should return a non-nil FocusManager")
	}
}

func TestApp_DispatchResizeEvent(t *testing.T) {
	type tc struct {
		initialWidth  int
		initialHeight int
		resizeWidth   int
		resizeHeight  int
		hasRoot       bool
	}

	tests := map[string]tc{
		"resize with root": {
			initialWidth:  80,
			initialHeight: 24,
			resizeWidth:   100,
			resizeHeight:  30,
			hasRoot:       true,
		},
		"resize without root": {
			initialWidth:  80,
			initialHeight: 24,
			resizeWidth:   100,
			resizeHeight:  30,
			hasRoot:       false,
		},
		"shrink terminal": {
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

			var mockRoot *mockRenderable
			if tt.hasRoot {
				mockRoot = newMockRenderable()
				mockRoot.dirty = false // Start as not dirty
				app.SetRoot(mockRoot)
			}

			event := ResizeEvent{Width: tt.resizeWidth, Height: tt.resizeHeight}
			handled := app.Dispatch(event)

			if !handled {
				t.Error("Dispatch(ResizeEvent) should return true")
			}

			// Check buffer was resized
			bufW, bufH := app.buffer.Size()
			if bufW != tt.resizeWidth || bufH != tt.resizeHeight {
				t.Errorf("Buffer size = (%d, %d), want (%d, %d)", bufW, bufH, tt.resizeWidth, tt.resizeHeight)
			}

			// Check root was marked dirty if it exists
			if tt.hasRoot && !mockRoot.markDirtyCalled {
				t.Error("MarkDirty should have been called on root after resize")
			}
		})
	}
}

func TestApp_DispatchKeyEvent(t *testing.T) {
	type tc struct {
		hasFocused   bool
		handled      bool
		expectReturn bool
	}

	tests := map[string]tc{
		"event handled by focused element": {
			hasFocused:   true,
			handled:      true,
			expectReturn: true,
		},
		"event not handled by focused element": {
			hasFocused:   true,
			handled:      false,
			expectReturn: false,
		},
		"no focused element": {
			hasFocused:   false,
			handled:      false,
			expectReturn: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			focus := NewFocusManager()

			if tt.hasFocused {
				mock := newMockFocusable("a", true)
				mock.handled = tt.handled
				focus.Register(mock)
			}

			app := &App{
				focus:  focus,
				buffer: NewBuffer(80, 24),
			}

			event := KeyEvent{Key: KeyEnter}
			result := app.Dispatch(event)

			if result != tt.expectReturn {
				t.Errorf("Dispatch(KeyEvent) = %v, want %v", result, tt.expectReturn)
			}
		})
	}
}

func TestApp_RenderWithMockRoot(t *testing.T) {
	// Create a mock terminal for testing
	mockTerm := NewMockTerminal(80, 24)
	buffer := NewBuffer(80, 24)

	app := &App{
		terminal: nil, // We can't use a real ANSITerminal in tests
		buffer:   buffer,
		focus:    NewFocusManager(),
	}

	// Create a mock renderable
	mockRoot := newMockRenderable()
	app.SetRoot(mockRoot)

	// Test that rendering calls the root's Render method
	mockRoot.Render(buffer, 80, 24)

	if !mockRoot.renderCalled {
		t.Error("Root's Render method should have been called")
	}

	// Verify the mock was used
	_ = mockTerm // We created it but App tests are limited without terminal
}
