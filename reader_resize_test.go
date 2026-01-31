package tui

import (
	"os"
	"testing"
	"time"
)

// testStdinReader creates a stdinReader with a pipe for testing.
// Returns the reader and a function to send simulated resize events.
func testStdinReader(t *testing.T) (*stdinReader, func(width, height int), func()) {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}

	reader := &stdinReader{
		fd:    int(r.Fd()),
		buf:   make([]byte, 256),
		sigCh: make(chan os.Signal, 10), // Buffered for testing
	}

	// simulateResize sends a resize event with the given dimensions.
	// It directly manipulates the reader's internal state to simulate SIGWINCH.
	simulateResize := func(width, height int) {
		// Update pending resize directly (simulates what drainResizeSignals does)
		reader.pendingResize = &ResizeEvent{Width: width, Height: height}
		reader.lastResizeTime = time.Now()
	}

	cleanup := func() {
		r.Close()
		w.Close()
	}

	return reader, simulateResize, cleanup
}

func TestStdinReader_SingleResizeEmitsAfterDebounce(t *testing.T) {
	type tc struct {
		width  int
		height int
	}

	tests := map[string]tc{
		"standard terminal size": {
			width:  80,
			height: 24,
		},
		"large terminal size": {
			width:  200,
			height: 50,
		},
		"small terminal size": {
			width:  40,
			height: 10,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			reader, simulateResize, cleanup := testStdinReader(t)
			defer cleanup()

			// Simulate resize
			simulateResize(tt.width, tt.height)

			// Wait for debounce window to pass
			time.Sleep(resizeDebounceWindow + 5*time.Millisecond)

			// Poll should return the resize event
			event, ok := reader.PollEvent(10 * time.Millisecond)
			if !ok {
				t.Error("PollEvent() returned false, expected resize event")
				return
			}

			re, isResize := event.(ResizeEvent)
			if !isResize {
				t.Errorf("PollEvent() returned %T, expected ResizeEvent", event)
				return
			}

			if re.Width != tt.width || re.Height != tt.height {
				t.Errorf("ResizeEvent = {%d, %d}, want {%d, %d}",
					re.Width, re.Height, tt.width, tt.height)
			}
		})
	}
}

func TestStdinReader_RapidResizesAreCoalesced(t *testing.T) {
	type tc struct {
		sizes      []struct{ w, h int }
		finalWidth int
		finalHeight int
	}

	tests := map[string]tc{
		"two rapid resizes": {
			sizes: []struct{ w, h int }{
				{80, 24},
				{100, 30},
			},
			finalWidth:  100,
			finalHeight: 30,
		},
		"three rapid resizes": {
			sizes: []struct{ w, h int }{
				{80, 24},
				{90, 28},
				{120, 40},
			},
			finalWidth:  120,
			finalHeight: 40,
		},
		"shrink then grow": {
			sizes: []struct{ w, h int }{
				{100, 50},
				{60, 20},
				{80, 30},
			},
			finalWidth:  80,
			finalHeight: 30,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			reader, simulateResize, cleanup := testStdinReader(t)
			defer cleanup()

			// Simulate rapid resizes (all within debounce window)
			for _, size := range tt.sizes {
				simulateResize(size.w, size.h)
			}

			// Wait for debounce window to pass
			time.Sleep(resizeDebounceWindow + 5*time.Millisecond)

			// Poll should return only ONE resize event with final dimensions
			event, ok := reader.PollEvent(10 * time.Millisecond)
			if !ok {
				t.Error("PollEvent() returned false, expected resize event")
				return
			}

			re, isResize := event.(ResizeEvent)
			if !isResize {
				t.Errorf("PollEvent() returned %T, expected ResizeEvent", event)
				return
			}

			if re.Width != tt.finalWidth || re.Height != tt.finalHeight {
				t.Errorf("ResizeEvent = {%d, %d}, want {%d, %d}",
					re.Width, re.Height, tt.finalWidth, tt.finalHeight)
			}

			// No more resize events should be pending
			event2, ok2 := reader.PollEvent(5 * time.Millisecond)
			if ok2 {
				t.Errorf("Unexpected second event: %v", event2)
			}
		})
	}
}

func TestStdinReader_ResizeNotEmittedBeforeDebounceWindow(t *testing.T) {
	type tc struct {
		description string
	}

	tests := map[string]tc{
		"resize is held during debounce": {
			description: "resize event should not be emitted immediately",
		},
	}

	for name := range tests {
		t.Run(name, func(t *testing.T) {
			reader, simulateResize, cleanup := testStdinReader(t)
			defer cleanup()

			// Simulate resize
			simulateResize(100, 30)

			// Poll immediately (before debounce window passes)
			// Use a very short timeout to avoid waiting
			event, ok := reader.PollEvent(1 * time.Millisecond)

			// Should not get the event yet (debounce window not passed)
			if ok && event != nil {
				if _, isResize := event.(ResizeEvent); isResize {
					t.Error("ResizeEvent should not be emitted before debounce window")
				}
			}
		})
	}
}

func TestStdinReader_NormalEventsNotAffectedByDebouncing(t *testing.T) {
	type tc struct {
		input    []byte
		expected KeyEvent
	}

	tests := map[string]tc{
		"single key": {
			input:    []byte{'a'},
			expected: KeyEvent{Key: KeyRune, Rune: 'a'},
		},
		"escape key": {
			input:    []byte{0x1b},
			expected: KeyEvent{Key: KeyEscape},
		},
		"enter key": {
			input:    []byte{'\r'},
			expected: KeyEvent{Key: KeyEnter},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create pipe for stdin simulation
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("os.Pipe() error = %v", err)
			}
			defer r.Close()
			defer w.Close()

			reader := &stdinReader{
				fd:    int(r.Fd()),
				buf:   make([]byte, 256),
				sigCh: make(chan os.Signal, 1),
			}

			// Write test input
			_, err = w.Write(tt.input)
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			// Poll should return the key event (not affected by resize debouncing)
			event, ok := reader.PollEvent(50 * time.Millisecond)
			if !ok {
				t.Error("PollEvent() returned false, expected key event")
				return
			}

			ke, isKey := event.(KeyEvent)
			if !isKey {
				t.Errorf("PollEvent() returned %T, expected KeyEvent", event)
				return
			}

			if ke.Key != tt.expected.Key || ke.Rune != tt.expected.Rune {
				t.Errorf("KeyEvent = %+v, want %+v", ke, tt.expected)
			}
		})
	}
}

func TestStdinReader_FinalResizeIsNotSwallowed(t *testing.T) {
	type tc struct {
		description string
	}

	tests := map[string]tc{
		"final resize always emitted": {
			description: "even after waiting, the resize should eventually be emitted",
		},
	}

	for name := range tests {
		t.Run(name, func(t *testing.T) {
			reader, simulateResize, cleanup := testStdinReader(t)
			defer cleanup()

			// Simulate resize
			simulateResize(150, 45)

			// Poll multiple times with short timeouts, eventually we should get the event
			var gotResize bool
			for i := 0; i < 10; i++ {
				event, ok := reader.PollEvent(5 * time.Millisecond)
				if ok {
					if re, isResize := event.(ResizeEvent); isResize {
						gotResize = true
						if re.Width != 150 || re.Height != 45 {
							t.Errorf("ResizeEvent = {%d, %d}, want {150, 45}", re.Width, re.Height)
						}
						break
					}
				}
			}

			if !gotResize {
				t.Error("Resize event was never emitted (swallowed)")
			}
		})
	}
}

func TestStdinReader_DebounceWindowConstant(t *testing.T) {
	// Verify the debounce window is set to 16ms as specified
	expected := 16 * time.Millisecond
	if resizeDebounceWindow != expected {
		t.Errorf("resizeDebounceWindow = %v, want %v", resizeDebounceWindow, expected)
	}
}
