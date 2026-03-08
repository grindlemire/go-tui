package tui

import "testing"

func TestDispatchTable_FocusRequired(t *testing.T) {
	type tc struct {
		focusedID      string // which mock to focus ("" = none)
		pressRune      rune
		expectAppQuit  bool
		expectInserted bool
	}

	tests := map[string]tc{
		"unfocused input: app handler fires, input skipped": {
			focusedID:      "",
			pressRune:      'q',
			expectAppQuit:  true,
			expectInserted: false,
		},
		"focused input: both fire, broadcast before focus-gated": {
			focusedID:      "input",
			pressRune:      'q',
			expectAppQuit:  true,  // broadcast handler fires first (position 0)
			expectInserted: true,  // focus-gated handler also fires (position 1, then stops)
		},
		"focused input: non-matching app key still works": {
			focusedID:      "input",
			pressRune:      0, // will use KeyEscape instead
			expectAppQuit:  true,
			expectInserted: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			appQuit := false
			charInserted := false

			fm := newFocusManager()
			inputMock := newMockFocusable("input", true)
			fm.Register(inputMock)

			if tt.focusedID == "input" {
				fm.SetFocus(inputMock)
			}

			// Build entries manually to test dispatch logic
			table := &dispatchTable{}

			// App-level broadcast binding: 'q' quits (no FocusRequired)
			table.entries = append(table.entries, dispatchEntry{
				pattern:  KeyPattern{Rune: 'q'},
				handler:  func(ke KeyEvent) { appQuit = true },
				stop:     false,
				position: 0,
			})

			// App-level broadcast binding: Escape quits (no FocusRequired)
			table.entries = append(table.entries, dispatchEntry{
				pattern:  KeyPattern{Key: KeyEscape, RequireNoMods: true},
				handler:  func(ke KeyEvent) { appQuit = true },
				stop:     false,
				position: 0,
			})

			// Input focus-gated binding: any rune inserts (FocusRequired)
			table.entries = append(table.entries, dispatchEntry{
				pattern:   KeyPattern{AnyRune: true, FocusRequired: true},
				handler:   func(ke KeyEvent) { charInserted = true },
				stop:      true,
				position:  1,
				focusable: inputMock,
			})

			var ke KeyEvent
			if tt.pressRune != 0 {
				ke = KeyEvent{Key: KeyRune, Rune: tt.pressRune}
			} else {
				ke = KeyEvent{Key: KeyEscape}
			}

			table.dispatch(ke, fm)

			if appQuit != tt.expectAppQuit {
				t.Errorf("appQuit = %v, want %v", appQuit, tt.expectAppQuit)
			}
			if charInserted != tt.expectInserted {
				t.Errorf("charInserted = %v, want %v", charInserted, tt.expectInserted)
			}
		})
	}
}

func TestDispatchTable_ValidateFocusRequired(t *testing.T) {
	type tc struct {
		entries   []dispatchEntry
		expectErr bool
	}

	tests := map[string]tc{
		"two focus-gated stop handlers do not conflict": {
			entries: []dispatchEntry{
				{pattern: KeyPattern{AnyRune: true, FocusRequired: true}, stop: true, position: 0, focusable: newMockFocusable("a", true)},
				{pattern: KeyPattern{AnyRune: true, FocusRequired: true}, stop: true, position: 1, focusable: newMockFocusable("b", true)},
			},
			expectErr: false,
		},
		"focus-gated and broadcast stop handlers do not conflict": {
			entries: []dispatchEntry{
				{pattern: KeyPattern{AnyRune: true, FocusRequired: true}, stop: true, position: 0, focusable: newMockFocusable("a", true)},
				{pattern: KeyPattern{AnyRune: true}, stop: true, position: 1},
			},
			expectErr: false,
		},
		"two broadcast stop handlers still conflict": {
			entries: []dispatchEntry{
				{pattern: KeyPattern{AnyRune: true}, stop: true, position: 0},
				{pattern: KeyPattern{AnyRune: true}, stop: true, position: 1},
			},
			expectErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			table := &dispatchTable{entries: tt.entries}
			err := table.validate()
			if tt.expectErr && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}
