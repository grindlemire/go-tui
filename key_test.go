package tui

import "testing"

func TestModifier_Has(t *testing.T) {
	type tc struct {
		mod      Modifier
		check    Modifier
		expected bool
	}

	tests := map[string]tc{
		"none has none":           {mod: ModNone, check: ModNone, expected: false},
		"ctrl has ctrl":           {mod: ModCtrl, check: ModCtrl, expected: true},
		"ctrl has alt":            {mod: ModCtrl, check: ModAlt, expected: false},
		"ctrl+alt has ctrl":       {mod: ModCtrl | ModAlt, check: ModCtrl, expected: true},
		"ctrl+alt has alt":        {mod: ModCtrl | ModAlt, check: ModAlt, expected: true},
		"ctrl+alt has shift":      {mod: ModCtrl | ModAlt, check: ModShift, expected: false},
		"all has ctrl":            {mod: ModCtrl | ModAlt | ModShift, check: ModCtrl, expected: true},
		"all has alt":             {mod: ModCtrl | ModAlt | ModShift, check: ModAlt, expected: true},
		"all has shift":           {mod: ModCtrl | ModAlt | ModShift, check: ModShift, expected: true},
		"shift alone has shift":   {mod: ModShift, check: ModShift, expected: true},
		"shift alone has ctrl":    {mod: ModShift, check: ModCtrl, expected: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.mod.Has(tt.check)
			if got != tt.expected {
				t.Errorf("Modifier(%d).Has(%d) = %v, want %v", tt.mod, tt.check, got, tt.expected)
			}
		})
	}
}

func TestKey_String(t *testing.T) {
	type tc struct {
		key      Key
		expected string
	}

	tests := map[string]tc{
		"none":      {key: KeyNone, expected: "None"},
		"rune":      {key: KeyRune, expected: "Rune"},
		"escape":    {key: KeyEscape, expected: "Escape"},
		"enter":     {key: KeyEnter, expected: "Enter"},
		"tab":       {key: KeyTab, expected: "Tab"},
		"backspace": {key: KeyBackspace, expected: "Backspace"},
		"delete":    {key: KeyDelete, expected: "Delete"},
		"insert":    {key: KeyInsert, expected: "Insert"},
		"up":        {key: KeyUp, expected: "Up"},
		"down":      {key: KeyDown, expected: "Down"},
		"left":      {key: KeyLeft, expected: "Left"},
		"right":     {key: KeyRight, expected: "Right"},
		"home":      {key: KeyHome, expected: "Home"},
		"end":       {key: KeyEnd, expected: "End"},
		"pageup":    {key: KeyPageUp, expected: "PageUp"},
		"pagedown":  {key: KeyPageDown, expected: "PageDown"},
		"f1":        {key: KeyF1, expected: "F1"},
		"f2":        {key: KeyF2, expected: "F2"},
		"f3":        {key: KeyF3, expected: "F3"},
		"f4":        {key: KeyF4, expected: "F4"},
		"f5":        {key: KeyF5, expected: "F5"},
		"f6":        {key: KeyF6, expected: "F6"},
		"f7":        {key: KeyF7, expected: "F7"},
		"f8":        {key: KeyF8, expected: "F8"},
		"f9":        {key: KeyF9, expected: "F9"},
		"f10":       {key: KeyF10, expected: "F10"},
		"f11":       {key: KeyF11, expected: "F11"},
		"f12":       {key: KeyF12, expected: "F12"},
		"ctrl+a":    {key: KeyCtrlA, expected: "Ctrl+A"},
		"ctrl+b":    {key: KeyCtrlB, expected: "Ctrl+B"},
		"ctrl+c":    {key: KeyCtrlC, expected: "Ctrl+C"},
		"ctrl+z":    {key: KeyCtrlZ, expected: "Ctrl+Z"},
		"ctrl+space": {key: KeyCtrlSpace, expected: "Ctrl+Space"},
		"unknown":   {key: Key(9999), expected: "Unknown"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.key.String()
			if got != tt.expected {
				t.Errorf("Key(%d).String() = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}
}

func TestModifier_String(t *testing.T) {
	type tc struct {
		mod      Modifier
		expected string
	}

	tests := map[string]tc{
		"none":              {mod: ModNone, expected: "None"},
		"ctrl":              {mod: ModCtrl, expected: "Ctrl"},
		"alt":               {mod: ModAlt, expected: "Alt"},
		"shift":             {mod: ModShift, expected: "Shift"},
		"ctrl+alt":          {mod: ModCtrl | ModAlt, expected: "Ctrl+Alt"},
		"ctrl+shift":        {mod: ModCtrl | ModShift, expected: "Ctrl+Shift"},
		"alt+shift":         {mod: ModAlt | ModShift, expected: "Alt+Shift"},
		"ctrl+alt+shift":    {mod: ModCtrl | ModAlt | ModShift, expected: "Ctrl+Alt+Shift"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.mod.String()
			if got != tt.expected {
				t.Errorf("Modifier(%d).String() = %q, want %q", tt.mod, got, tt.expected)
			}
		})
	}
}
