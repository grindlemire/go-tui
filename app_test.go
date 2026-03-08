package tui

import (
	"testing"
)

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
			app := &App{
				focus:  newFocusManager(),
				buffer: NewBuffer(80, 24),
			}

			if tt.createRoot {
				root := New()
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

func TestApp_FocusedNoElements(t *testing.T) {
	app := &App{
		focus: newFocusManager(),
	}

	if app.Focused() != nil {
		t.Error("Focused() should return nil when no elements are registered")
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
				focus:  newFocusManager(),
				buffer: buffer,
			}

			var root *Element
			if tt.hasRoot {
				root = New()
				app.SetRoot(root)
				// Clear dirty state so we can detect the resize marking it dirty
				root.SetDirty(false)
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
			if tt.hasRoot && !root.IsDirty() {
				t.Error("Root should be dirty after resize")
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
			focus := newFocusManager()

			if tt.hasFocused {
				mock := newMockFocusable("a", true)
				mock.handled = tt.handled
				focus.Register(mock)
				focus.SetFocus(mock)
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

func TestApp_RenderWithRoot(t *testing.T) {
	buffer := NewBuffer(80, 24)

	app := &App{
		buffer: buffer,
		focus:  newFocusManager(),
	}

	root := New(WithText("hello"))
	app.SetRoot(root)

	// Render directly to buffer
	root.Render(buffer, 80, 24)

	// After rendering, element should no longer be dirty
	if root.IsDirty() {
		t.Error("Root should not be dirty after Render()")
	}
}
