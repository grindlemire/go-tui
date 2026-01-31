package tui

import "testing"

func TestParseMouseSGR(t *testing.T) {
	type tc struct {
		input            []byte
		expectedEvent    MouseEvent
		expectedConsumed int
	}

	tests := map[string]tc{
		"left press at 1,1": {
			input:            []byte("\x1b[<0;1;1M"),
			expectedEvent:    MouseEvent{Button: MouseLeft, Action: MousePress, X: 0, Y: 0},
			expectedConsumed: 9,
		},
		"left release at 1,1": {
			input:            []byte("\x1b[<0;1;1m"),
			expectedEvent:    MouseEvent{Button: MouseLeft, Action: MouseRelease, X: 0, Y: 0},
			expectedConsumed: 9,
		},
		"middle press at 10,20": {
			input:            []byte("\x1b[<1;10;20M"),
			expectedEvent:    MouseEvent{Button: MouseMiddle, Action: MousePress, X: 9, Y: 19},
			expectedConsumed: 11,
		},
		"right press at 5,5": {
			input:            []byte("\x1b[<2;5;5M"),
			expectedEvent:    MouseEvent{Button: MouseRight, Action: MousePress, X: 4, Y: 4},
			expectedConsumed: 9,
		},
		"wheel up": {
			input:            []byte("\x1b[<64;10;10M"),
			expectedEvent:    MouseEvent{Button: MouseWheelUp, Action: MousePress, X: 9, Y: 9},
			expectedConsumed: 12,
		},
		"wheel down": {
			input:            []byte("\x1b[<65;10;10M"),
			expectedEvent:    MouseEvent{Button: MouseWheelDown, Action: MousePress, X: 9, Y: 9},
			expectedConsumed: 12,
		},
		"left drag": {
			input:            []byte("\x1b[<32;15;25M"),
			expectedEvent:    MouseEvent{Button: MouseLeft, Action: MouseDrag, X: 14, Y: 24},
			expectedConsumed: 12,
		},
		"shift+left click": {
			input:            []byte("\x1b[<4;5;5M"),
			expectedEvent:    MouseEvent{Button: MouseLeft, Action: MousePress, X: 4, Y: 4, Mod: ModShift},
			expectedConsumed: 9,
		},
		"alt+left click": {
			input:            []byte("\x1b[<8;5;5M"),
			expectedEvent:    MouseEvent{Button: MouseLeft, Action: MousePress, X: 4, Y: 4, Mod: ModAlt},
			expectedConsumed: 9,
		},
		"ctrl+left click": {
			input:            []byte("\x1b[<16;5;5M"),
			expectedEvent:    MouseEvent{Button: MouseLeft, Action: MousePress, X: 4, Y: 4, Mod: ModCtrl},
			expectedConsumed: 10,
		},
		"ctrl+shift+alt+left click": {
			input:            []byte("\x1b[<28;5;5M"),
			expectedEvent:    MouseEvent{Button: MouseLeft, Action: MousePress, X: 4, Y: 4, Mod: ModCtrl | ModShift | ModAlt},
			expectedConsumed: 10,
		},
		"large coordinates": {
			input:            []byte("\x1b[<0;200;100M"),
			expectedEvent:    MouseEvent{Button: MouseLeft, Action: MousePress, X: 199, Y: 99},
			expectedConsumed: 13,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			event, consumed := parseMouseSGR(tt.input)
			if consumed != tt.expectedConsumed {
				t.Errorf("parseMouseSGR(%q) consumed %d bytes, want %d", tt.input, consumed, tt.expectedConsumed)
			}
			if event.Button != tt.expectedEvent.Button {
				t.Errorf("parseMouseSGR(%q) button = %v, want %v", tt.input, event.Button, tt.expectedEvent.Button)
			}
			if event.Action != tt.expectedEvent.Action {
				t.Errorf("parseMouseSGR(%q) action = %v, want %v", tt.input, event.Action, tt.expectedEvent.Action)
			}
			if event.X != tt.expectedEvent.X {
				t.Errorf("parseMouseSGR(%q) X = %d, want %d", tt.input, event.X, tt.expectedEvent.X)
			}
			if event.Y != tt.expectedEvent.Y {
				t.Errorf("parseMouseSGR(%q) Y = %d, want %d", tt.input, event.Y, tt.expectedEvent.Y)
			}
			if event.Mod != tt.expectedEvent.Mod {
				t.Errorf("parseMouseSGR(%q) Mod = %v, want %v", tt.input, event.Mod, tt.expectedEvent.Mod)
			}
		})
	}
}

func TestParseMouseSGR_Invalid(t *testing.T) {
	type tc struct {
		input []byte
	}

	tests := map[string]tc{
		"empty":                     {input: []byte{}},
		"too short":                 {input: []byte("\x1b[<")},
		"missing M":                 {input: []byte("\x1b[<0;1;1")},
		"wrong prefix":              {input: []byte("\x1b[0;1;1M")},
		"missing x":                 {input: []byte("\x1b[<0;;1M")},
		"missing y":                 {input: []byte("\x1b[<0;1;M")},
		"non-numeric button":        {input: []byte("\x1b[<a;1;1M")},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, consumed := parseMouseSGR(tt.input)
			if consumed != 0 {
				t.Errorf("parseMouseSGR(%q) consumed %d bytes, want 0 for invalid input", tt.input, consumed)
			}
		})
	}
}

func TestParseInput_MouseEvents(t *testing.T) {
	type tc struct {
		input    []byte
		expected MouseEvent
	}

	tests := map[string]tc{
		"left click": {
			input:    []byte("\x1b[<0;10;20M"),
			expected: MouseEvent{Button: MouseLeft, Action: MousePress, X: 9, Y: 19},
		},
		"right click": {
			input:    []byte("\x1b[<2;5;5M"),
			expected: MouseEvent{Button: MouseRight, Action: MousePress, X: 4, Y: 4},
		},
		"wheel up": {
			input:    []byte("\x1b[<64;1;1M"),
			expected: MouseEvent{Button: MouseWheelUp, Action: MousePress, X: 0, Y: 0},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			events := parseInput(tt.input)
			if len(events) != 1 {
				t.Fatalf("parseInput(%q) returned %d events, want 1", tt.input, len(events))
			}
			me, ok := events[0].(MouseEvent)
			if !ok {
				t.Fatalf("event is not MouseEvent, got %T", events[0])
			}
			if me.Button != tt.expected.Button {
				t.Errorf("parseInput(%q) button = %v, want %v", tt.input, me.Button, tt.expected.Button)
			}
			if me.Action != tt.expected.Action {
				t.Errorf("parseInput(%q) action = %v, want %v", tt.input, me.Action, tt.expected.Action)
			}
			if me.X != tt.expected.X || me.Y != tt.expected.Y {
				t.Errorf("parseInput(%q) pos = (%d,%d), want (%d,%d)", tt.input, me.X, me.Y, tt.expected.X, tt.expected.Y)
			}
		})
	}
}
