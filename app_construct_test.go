//go:build !windows

package tui

import (
	"os"
	"testing"
	"time"
)

// stdinIsTTY reports whether the test process's stdin is a real terminal.
// go test normally wires stdin to /dev/null, so this returns false. If a
// real TTY is attached, construction tests skip instead of letting NewApp
// take over the user's terminal.
func stdinIsTTY(t *testing.T) bool {
	t.Helper()
	state, err := enableRawMode(int(os.Stdin.Fd()))
	if err != nil {
		return false
	}
	if err := disableRawMode(state); err != nil {
		t.Fatalf("disableRawMode() error = %v", err)
	}
	return true
}

// TestNewApp_FailsWithoutTTY exercises the deterministic error path of
// NewApp: when stdin is not a terminal, entering raw mode fails and NewApp
// must return a nil app and a non-nil error. The success path requires a
// real TTY and is skipped in that environment instead.
func TestNewApp_FailsWithoutTTY(t *testing.T) {
	type tc struct {
		opts []AppOption
	}

	tests := map[string]tc{
		"no options": {
			opts: nil,
		},
		"with options": {
			opts: []AppOption{
				WithFrameRate(30),
				WithInputLatency(10 * time.Millisecond),
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if stdinIsTTY(t) {
				t.Skip("stdin is a real TTY; NewApp would take over the terminal")
			}

			app, err := NewApp(tt.opts...)
			if err == nil {
				if app != nil {
					app.Close()
				}
				t.Fatal("NewApp() error = nil, want raw mode error when stdin is not a TTY")
			}
			if app != nil {
				t.Errorf("NewApp() app = %v, want nil on error", app)
			}
		})
	}
}

// TestNewAppWithReader_FailsWithoutTTY covers the same raw mode error path
// for NewAppWithReader. Even with an injected reader, the constructor still
// builds an ANSITerminal on os.Stdin and enters raw mode, so without a TTY
// it must fail before the reader is ever used.
func TestNewAppWithReader_FailsWithoutTTY(t *testing.T) {
	type tc struct {
		opts []AppOption
	}

	tests := map[string]tc{
		"no options": {
			opts: nil,
		},
		"with options": {
			opts: []AppOption{
				WithEventQueueSize(8),
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if stdinIsTTY(t) {
				t.Skip("stdin is a real TTY; NewAppWithReader would take over the terminal")
			}

			reader := NewMockEventReader(KeyEvent{Key: KeyEnter})

			app, err := NewAppWithReader(reader, tt.opts...)
			if err == nil {
				if app != nil {
					app.Close()
				}
				t.Fatal("NewAppWithReader() error = nil, want raw mode error when stdin is not a TTY")
			}
			if app != nil {
				t.Errorf("NewAppWithReader() app = %v, want nil on error", app)
			}

			// The raw mode failure happens before the reader is polled, so the
			// injected reader's queue must be untouched.
			if reader.Remaining() != 1 {
				t.Errorf("reader.Remaining() = %d, want 1 (reader should not be consumed)", reader.Remaining())
			}
		})
	}
}

func TestApp_BlurFocused(t *testing.T) {
	type tc struct {
		focusFirst   bool
		wantBlurCall bool
	}

	tests := map[string]tc{
		"blurs the focused element": {
			focusFirst:   true,
			wantBlurCall: true,
		},
		"no-op when nothing is focused": {
			focusFirst:   false,
			wantBlurCall: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app := &App{
				focus:  newFocusManager(),
				buffer: NewBuffer(80, 24),
				stopCh: make(chan struct{}),
			}
			defer close(app.stopCh)

			blurCalled := false
			first := New(WithFocusable(true), WithOnBlur(func(*Element) {
				blurCalled = true
			}))
			second := New(WithFocusable(true))
			root := New()
			root.AddChild(first)
			root.AddChild(second)
			app.SetRoot(root)

			if tt.focusFirst {
				app.FocusNext()
				if app.Focused() == nil {
					t.Fatal("FocusNext() should focus a focusable element")
				}
				if !first.IsFocused() {
					t.Fatal("first focusable child should be focused after FocusNext()")
				}
			}

			app.BlurFocused()

			if app.Focused() != nil {
				t.Errorf("Focused() = %v after BlurFocused(), want nil", app.Focused())
			}
			if first.IsFocused() {
				t.Error("element should not report focus after BlurFocused()")
			}
			if blurCalled != tt.wantBlurCall {
				t.Errorf("onBlur called = %v, want %v", blurCalled, tt.wantBlurCall)
			}

			// Calling again with nothing focused must remain a safe no-op.
			blurCalled = false
			app.BlurFocused()
			if app.Focused() != nil {
				t.Error("Focused() should stay nil after repeated BlurFocused()")
			}
			if blurCalled {
				t.Error("onBlur should not fire when nothing is focused")
			}
		})
	}
}

func TestApp_EventQueue(t *testing.T) {
	app := &App{
		watcherQueue: make(chan func(), 4),
	}

	q := app.EventQueue()
	if q == nil {
		t.Fatal("EventQueue() returned nil channel")
	}

	called := false
	q <- func() { called = true }

	select {
	case fn := <-app.watcherQueue:
		fn()
	default:
		t.Fatal("function sent on EventQueue() did not arrive on the watcher queue")
	}

	if !called {
		t.Error("function received from the watcher queue was not the one sent")
	}
}
