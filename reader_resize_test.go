//go:build !windows

package tui

import (
	"os"
	"testing"
	"time"
)

func TestStdinReader_NormalEventsNotAffectedByReaderChanges(t *testing.T) {
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
				fd:  int(r.Fd()),
				buf: make([]byte, 256),
			}

			// Write test input
			_, err = w.Write(tt.input)
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			// Poll should return the key event
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
