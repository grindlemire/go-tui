package tui

import (
	"os"
	"testing"
	"time"
)

func TestMockEventReader_ReturnsEventsInOrder(t *testing.T) {
	type tc struct {
		events   []Event
		expected []Event
	}

	tests := map[string]tc{
		"single key event": {
			events: []Event{
				KeyEvent{Key: KeyEnter},
			},
			expected: []Event{
				KeyEvent{Key: KeyEnter},
			},
		},
		"multiple key events": {
			events: []Event{
				KeyEvent{Key: KeyUp},
				KeyEvent{Key: KeyDown},
				KeyEvent{Key: KeyLeft},
			},
			expected: []Event{
				KeyEvent{Key: KeyUp},
				KeyEvent{Key: KeyDown},
				KeyEvent{Key: KeyLeft},
			},
		},
		"mixed event types": {
			events: []Event{
				KeyEvent{Key: KeyRune, Rune: 'a'},
				ResizeEvent{Width: 80, Height: 24},
				KeyEvent{Key: KeyEscape},
			},
			expected: []Event{
				KeyEvent{Key: KeyRune, Rune: 'a'},
				ResizeEvent{Width: 80, Height: 24},
				KeyEvent{Key: KeyEscape},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			reader := NewMockEventReader(tt.events...)

			for i, expected := range tt.expected {
				event, ok := reader.PollEvent(time.Millisecond)
				if !ok {
					t.Errorf("PollEvent() returned false at index %d, expected event", i)
					continue
				}

				// Type-switch to compare events
				switch exp := expected.(type) {
				case KeyEvent:
					ke, isKey := event.(KeyEvent)
					if !isKey {
						t.Errorf("Event %d: expected KeyEvent, got %T", i, event)
						continue
					}
					if ke.Key != exp.Key || ke.Rune != exp.Rune || ke.Mod != exp.Mod {
						t.Errorf("Event %d: got %+v, want %+v", i, ke, exp)
					}
				case ResizeEvent:
					re, isResize := event.(ResizeEvent)
					if !isResize {
						t.Errorf("Event %d: expected ResizeEvent, got %T", i, event)
						continue
					}
					if re.Width != exp.Width || re.Height != exp.Height {
						t.Errorf("Event %d: got %+v, want %+v", i, re, exp)
					}
				}
			}
		})
	}
}

func TestMockEventReader_ReturnsFalseWhenExhausted(t *testing.T) {
	type tc struct {
		events []Event
	}

	tests := map[string]tc{
		"empty reader": {
			events: []Event{},
		},
		"after consuming all events": {
			events: []Event{
				KeyEvent{Key: KeyEnter},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			reader := NewMockEventReader(tt.events...)

			// Consume all events
			for range tt.events {
				reader.PollEvent(time.Millisecond)
			}

			// Next call should return false
			event, ok := reader.PollEvent(time.Millisecond)
			if ok {
				t.Errorf("PollEvent() returned true, want false")
			}
			if event != nil {
				t.Errorf("PollEvent() returned event %v, want nil", event)
			}
		})
	}
}

func TestMockEventReader_Reset(t *testing.T) {
	events := []Event{
		KeyEvent{Key: KeyEnter},
		KeyEvent{Key: KeyEscape},
	}
	reader := NewMockEventReader(events...)

	// Consume first event
	reader.PollEvent(time.Millisecond)

	// Reset
	reader.Reset()

	// Should return first event again
	event, ok := reader.PollEvent(time.Millisecond)
	if !ok {
		t.Error("PollEvent() returned false after Reset")
	}
	ke, isKey := event.(KeyEvent)
	if !isKey || ke.Key != KeyEnter {
		t.Errorf("After Reset: got %+v, want KeyEvent{Key: KeyEnter}", event)
	}
}

func TestMockEventReader_AddEvents(t *testing.T) {
	reader := NewMockEventReader(KeyEvent{Key: KeyEnter})

	// Add more events
	reader.AddEvents(KeyEvent{Key: KeyEscape})

	// Consume first
	reader.PollEvent(time.Millisecond)

	// Should get the added event
	event, ok := reader.PollEvent(time.Millisecond)
	if !ok {
		t.Error("PollEvent() returned false after AddEvents")
	}
	ke, isKey := event.(KeyEvent)
	if !isKey || ke.Key != KeyEscape {
		t.Errorf("After AddEvents: got %+v, want KeyEvent{Key: KeyEscape}", event)
	}
}

func TestMockEventReader_Remaining(t *testing.T) {
	type tc struct {
		events        []Event
		consumeCount  int
		expectedRemain int
	}

	tests := map[string]tc{
		"all remaining": {
			events:        []Event{KeyEvent{Key: KeyUp}, KeyEvent{Key: KeyDown}},
			consumeCount:  0,
			expectedRemain: 2,
		},
		"some consumed": {
			events:        []Event{KeyEvent{Key: KeyUp}, KeyEvent{Key: KeyDown}},
			consumeCount:  1,
			expectedRemain: 1,
		},
		"all consumed": {
			events:        []Event{KeyEvent{Key: KeyUp}, KeyEvent{Key: KeyDown}},
			consumeCount:  2,
			expectedRemain: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			reader := NewMockEventReader(tt.events...)

			for i := 0; i < tt.consumeCount; i++ {
				reader.PollEvent(time.Millisecond)
			}

			if got := reader.Remaining(); got != tt.expectedRemain {
				t.Errorf("Remaining() = %d, want %d", got, tt.expectedRemain)
			}
		})
	}
}

func TestMockEventReader_Close(t *testing.T) {
	reader := NewMockEventReader(KeyEvent{Key: KeyEnter})

	err := reader.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestStdinReader_NewEventReader(t *testing.T) {
	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	defer r.Close()
	defer w.Close()

	reader, err := NewEventReader(r)
	if err != nil {
		t.Fatalf("NewEventReader() error = %v", err)
	}
	defer reader.Close()

	// Verify reader was created
	if reader == nil {
		t.Error("NewEventReader() returned nil reader")
	}
}

func TestStdinReader_PollEventTimeout(t *testing.T) {
	// Create a pipe with no data to simulate timeout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	defer r.Close()
	defer w.Close()

	reader, err := NewEventReader(r)
	if err != nil {
		t.Fatalf("NewEventReader() error = %v", err)
	}
	defer reader.Close()

	// Poll with very short timeout - should return false
	start := time.Now()
	event, ok := reader.PollEvent(10 * time.Millisecond)
	elapsed := time.Since(start)

	if ok {
		t.Errorf("PollEvent() returned true with no data, got event: %v", event)
	}
	if event != nil {
		t.Errorf("PollEvent() returned non-nil event: %v", event)
	}

	// Should have respected roughly the timeout duration
	if elapsed < 5*time.Millisecond {
		t.Errorf("PollEvent() returned too quickly: %v", elapsed)
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("PollEvent() took too long: %v", elapsed)
	}
}

func TestStdinReader_PollEventWithData(t *testing.T) {
	// Create a pipe and write data
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	defer r.Close()
	defer w.Close()

	reader, err := NewEventReader(r)
	if err != nil {
		t.Fatalf("NewEventReader() error = %v", err)
	}
	defer reader.Close()

	// Write a simple key to the pipe
	_, err = w.Write([]byte{'a'})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Poll for the event
	event, ok := reader.PollEvent(100 * time.Millisecond)
	if !ok {
		t.Error("PollEvent() returned false, expected event")
	}

	ke, isKey := event.(KeyEvent)
	if !isKey {
		t.Errorf("PollEvent() returned %T, expected KeyEvent", event)
	} else if ke.Key != KeyRune || ke.Rune != 'a' {
		t.Errorf("PollEvent() = %+v, want KeyEvent{Key: KeyRune, Rune: 'a'}", ke)
	}
}

func TestStdinReader_PollEventEscapeSequence(t *testing.T) {
	// Create a pipe and write an escape sequence (arrow up)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	defer r.Close()
	defer w.Close()

	reader, err := NewEventReader(r)
	if err != nil {
		t.Fatalf("NewEventReader() error = %v", err)
	}
	defer reader.Close()

	// Write arrow up escape sequence
	_, err = w.Write([]byte("\x1b[A"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Poll for the event
	event, ok := reader.PollEvent(100 * time.Millisecond)
	if !ok {
		t.Error("PollEvent() returned false, expected event")
	}

	ke, isKey := event.(KeyEvent)
	if !isKey {
		t.Errorf("PollEvent() returned %T, expected KeyEvent", event)
	} else if ke.Key != KeyUp {
		t.Errorf("PollEvent() = %+v, want KeyEvent{Key: KeyUp}", ke)
	}
}

func TestFindIncompleteUTF8Suffix(t *testing.T) {
	type tc struct {
		input    []byte
		expected []byte
	}

	tests := map[string]tc{
		"empty": {
			input:    []byte{},
			expected: nil,
		},
		"ascii only": {
			input:    []byte("hello"),
			expected: nil,
		},
		"complete utf8 2 byte": {
			input:    []byte{0xc3, 0xa9}, // é
			expected: nil,
		},
		"incomplete utf8 2 byte start": {
			input:    []byte{0xc3}, // Start of 2-byte sequence
			expected: []byte{0xc3},
		},
		"complete utf8 3 byte": {
			input:    []byte{0xe2, 0x9c, 0x93}, // ✓
			expected: nil,
		},
		"incomplete utf8 3 byte 1 missing": {
			input:    []byte{0xe2, 0x9c}, // Missing last byte
			expected: []byte{0xe2, 0x9c},
		},
		"incomplete utf8 3 byte 2 missing": {
			input:    []byte{0xe2}, // Missing 2 bytes
			expected: []byte{0xe2},
		},
		"complete utf8 4 byte": {
			input:    []byte{0xf0, 0x9f, 0x98, 0x80}, // 😀
			expected: nil,
		},
		"incomplete utf8 4 byte 1 missing": {
			input:    []byte{0xf0, 0x9f, 0x98}, // Missing last byte
			expected: []byte{0xf0, 0x9f, 0x98},
		},
		"ascii followed by incomplete utf8": {
			input:    []byte{'a', 'b', 0xc3},
			expected: []byte{0xc3},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := findIncompleteUTF8Suffix(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("findIncompleteUTF8Suffix(%v) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("findIncompleteUTF8Suffix(%v) = %v, want %v", tt.input, result, tt.expected)
					return
				}
			}
		})
	}
}

func TestParseInputWithRemainder(t *testing.T) {
	type tc struct {
		input          []byte
		expectedEvents int
		expectedRemain int
	}

	tests := map[string]tc{
		"complete input": {
			input:          []byte("abc"),
			expectedEvents: 3,
			expectedRemain: 0,
		},
		"with incomplete utf8": {
			input:          []byte{'a', 0xc3}, // 'a' followed by incomplete 2-byte
			expectedEvents: 1,
			expectedRemain: 1,
		},
		"only incomplete utf8": {
			input:          []byte{0xe2, 0x9c}, // incomplete 3-byte
			expectedEvents: 0,
			expectedRemain: 2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			events, remaining := parseInputWithRemainder(tt.input)
			if len(events) != tt.expectedEvents {
				t.Errorf("parseInputWithRemainder(%v) events count = %d, want %d", tt.input, len(events), tt.expectedEvents)
			}
			if len(remaining) != tt.expectedRemain {
				t.Errorf("parseInputWithRemainder(%v) remaining count = %d, want %d", tt.input, len(remaining), tt.expectedRemain)
			}
		})
	}
}
