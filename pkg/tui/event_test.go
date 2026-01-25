package tui

import "testing"

func TestKeyEvent_IsRune(t *testing.T) {
	type tc struct {
		event    KeyEvent
		expected bool
	}

	tests := map[string]tc{
		"rune event":       {event: KeyEvent{Key: KeyRune, Rune: 'a'}, expected: true},
		"enter event":      {event: KeyEvent{Key: KeyEnter}, expected: false},
		"escape event":     {event: KeyEvent{Key: KeyEscape}, expected: false},
		"arrow event":      {event: KeyEvent{Key: KeyUp}, expected: false},
		"function event":   {event: KeyEvent{Key: KeyF1}, expected: false},
		"ctrl event":       {event: KeyEvent{Key: KeyCtrlA}, expected: false},
		"rune with mod":    {event: KeyEvent{Key: KeyRune, Rune: 'x', Mod: ModCtrl}, expected: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.event.IsRune()
			if got != tt.expected {
				t.Errorf("KeyEvent{Key: %v}.IsRune() = %v, want %v", tt.event.Key, got, tt.expected)
			}
		})
	}
}

func TestKeyEvent_Is(t *testing.T) {
	type tc struct {
		event    KeyEvent
		key      Key
		mods     []Modifier
		expected bool
	}

	tests := map[string]tc{
		"enter matches enter":              {event: KeyEvent{Key: KeyEnter}, key: KeyEnter, expected: true},
		"enter does not match escape":      {event: KeyEvent{Key: KeyEnter}, key: KeyEscape, expected: false},
		"rune matches rune":                {event: KeyEvent{Key: KeyRune, Rune: 'a'}, key: KeyRune, expected: true},
		"ctrl+a with ctrl matches":         {event: KeyEvent{Key: KeyRune, Rune: 'a', Mod: ModCtrl}, key: KeyRune, mods: []Modifier{ModCtrl}, expected: true},
		"ctrl+a without ctrl no match":     {event: KeyEvent{Key: KeyRune, Rune: 'a', Mod: ModCtrl}, key: KeyRune, expected: true},
		"no mod with ctrl no match":        {event: KeyEvent{Key: KeyRune, Rune: 'a'}, key: KeyRune, mods: []Modifier{ModCtrl}, expected: false},
		"ctrl+alt matches both":            {event: KeyEvent{Key: KeyRune, Rune: 'a', Mod: ModCtrl | ModAlt}, key: KeyRune, mods: []Modifier{ModCtrl, ModAlt}, expected: true},
		"ctrl+alt only ctrl no match":      {event: KeyEvent{Key: KeyRune, Rune: 'a', Mod: ModCtrl | ModAlt}, key: KeyRune, mods: []Modifier{ModCtrl}, expected: false},
		"shift matches shift":              {event: KeyEvent{Key: KeyUp, Mod: ModShift}, key: KeyUp, mods: []Modifier{ModShift}, expected: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.event.Is(tt.key, tt.mods...)
			if got != tt.expected {
				t.Errorf("KeyEvent.Is(%v, %v) = %v, want %v", tt.key, tt.mods, got, tt.expected)
			}
		})
	}
}

func TestKeyEvent_Char(t *testing.T) {
	type tc struct {
		event    KeyEvent
		expected rune
	}

	tests := map[string]tc{
		"rune a":        {event: KeyEvent{Key: KeyRune, Rune: 'a'}, expected: 'a'},
		"rune space":    {event: KeyEvent{Key: KeyRune, Rune: ' '}, expected: ' '},
		"rune unicode":  {event: KeyEvent{Key: KeyRune, Rune: '日'}, expected: '日'},
		"enter":         {event: KeyEvent{Key: KeyEnter}, expected: 0},
		"escape":        {event: KeyEvent{Key: KeyEscape}, expected: 0},
		"arrow":         {event: KeyEvent{Key: KeyUp}, expected: 0},
		"function key":  {event: KeyEvent{Key: KeyF1}, expected: 0},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.event.Char()
			if got != tt.expected {
				t.Errorf("KeyEvent{Key: %v, Rune: %q}.Char() = %q, want %q", tt.event.Key, tt.event.Rune, got, tt.expected)
			}
		})
	}
}

func TestEvent_TypeAssertion(t *testing.T) {
	type tc struct {
		event   Event
		isKey   bool
		isResize bool
	}

	tests := map[string]tc{
		"key event":    {event: KeyEvent{Key: KeyEnter}, isKey: true, isResize: false},
		"resize event": {event: ResizeEvent{Width: 80, Height: 24}, isKey: false, isResize: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, isKey := tt.event.(KeyEvent)
			_, isResize := tt.event.(ResizeEvent)
			if isKey != tt.isKey {
				t.Errorf("type assertion to KeyEvent = %v, want %v", isKey, tt.isKey)
			}
			if isResize != tt.isResize {
				t.Errorf("type assertion to ResizeEvent = %v, want %v", isResize, tt.isResize)
			}
		})
	}
}

func TestResizeEvent(t *testing.T) {
	type tc struct {
		width  int
		height int
	}

	tests := map[string]tc{
		"standard terminal": {width: 80, height: 24},
		"large terminal":    {width: 200, height: 50},
		"small terminal":    {width: 40, height: 10},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := ResizeEvent{Width: tt.width, Height: tt.height}
			if e.Width != tt.width {
				t.Errorf("ResizeEvent.Width = %d, want %d", e.Width, tt.width)
			}
			if e.Height != tt.height {
				t.Errorf("ResizeEvent.Height = %d, want %d", e.Height, tt.height)
			}
		})
	}
}
