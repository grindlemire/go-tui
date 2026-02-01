package tui

import (
	"fmt"
	"time"
)

// AppOption is a functional option for configuring an App.
type AppOption func(*App) error

// WithInputLatency sets the polling timeout for the event reader.
// Default is 50ms. Use InputLatencyBlocking (-1) for blocking mode.
// A value of 0 is not allowed and will return an error.
func WithInputLatency(d time.Duration) AppOption {
	return func(a *App) error {
		if d == 0 {
			return fmt.Errorf("input latency of 0 (busy polling) is not allowed; use a positive duration or InputLatencyBlocking")
		}
		a.inputLatency = d
		return nil
	}
}

// WithFrameRate sets the target frame rate for the render loop.
// Default is 60 fps (16ms frame duration). Valid range is 1-240 fps.
func WithFrameRate(fps int) AppOption {
	return func(a *App) error {
		if fps < 1 {
			return fmt.Errorf("frame rate must be at least 1 fps")
		}
		if fps > 240 {
			return fmt.Errorf("frame rate cannot exceed 240 fps")
		}
		a.frameDuration = time.Second / time.Duration(fps)
		return nil
	}
}

// WithEventQueueSize sets the capacity of the event queue buffer.
// Default is 256. Must be at least 1.
func WithEventQueueSize(size int) AppOption {
	return func(a *App) error {
		if size < 1 {
			return fmt.Errorf("event queue size must be at least 1")
		}
		a.eventQueueSize = size
		return nil
	}
}

// WithGlobalKeyHandler sets a handler that runs before dispatching to focused element.
// If the handler returns true, the event is consumed and not dispatched further.
// Use this for app-level key bindings like quit.
func WithGlobalKeyHandler(fn func(KeyEvent) bool) AppOption {
	return func(a *App) error {
		a.globalKeyHandler = fn
		return nil
	}
}

// WithRoot sets the root view for rendering. Accepts:
//   - A view struct implementing Viewable (extracts Root, starts watchers)
//   - A raw Renderable (element.Element)
//
// The root is set after the app is fully initialized.
func WithRoot(v Viewable) AppOption {
	return func(a *App) error {
		a.pendingRoot = v
		return nil
	}
}

// WithoutMouse disables mouse event reporting.
// By default, mouse events are enabled.
func WithoutMouse() AppOption {
	return func(a *App) error {
		a.mouseEnabled = false
		return nil
	}
}

// WithCursor keeps the cursor visible during app execution.
// By default, the cursor is hidden.
func WithCursor() AppOption {
	return func(a *App) error {
		a.cursorVisible = true
		return nil
	}
}

// WithInlineHeight enables inline widget mode.
// The TUI manages only the bottom N rows of the terminal, allowing normal
// terminal output above. Use PrintAbove() or PrintAboveln() to print
// scrolling content above the widget.
// In inline mode, alternate screen is not used, so terminal history is preserved.
func WithInlineHeight(rows int) AppOption {
	return func(a *App) error {
		if rows < 1 {
			return fmt.Errorf("inline height must be at least 1 row")
		}
		a.inlineHeight = rows
		return nil
	}
}
