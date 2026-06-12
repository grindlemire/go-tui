package tui

import "testing"

func TestKey_KeyPattern(t *testing.T) {
	type tc struct {
		key  Key
		want KeyPattern
	}

	tests := map[string]tc{
		"bare key excludes all modifiers": {
			key:  KeyEnter,
			want: KeyPattern{Key: KeyEnter, ExcludeMods: ModCtrl | ModAlt | ModShift},
		},
		"escape excludes all modifiers": {
			key:  KeyEscape,
			want: KeyPattern{Key: KeyEscape, ExcludeMods: ModCtrl | ModAlt | ModShift},
		},
		"function key excludes all modifiers": {
			key:  KeyF5,
			want: KeyPattern{Key: KeyF5, ExcludeMods: ModCtrl | ModAlt | ModShift},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.key.keyPattern()
			if got != tt.want {
				t.Errorf("Key(%v).keyPattern() = %+v, want %+v", tt.key, got, tt.want)
			}
		})
	}
}

func TestKey_ModifierHelpers(t *testing.T) {
	type tc struct {
		spec KeySpec
		want KeyPattern
	}

	tests := map[string]tc{
		"Key.Ctrl requires ctrl": {
			spec: KeyUp.Ctrl(),
			want: KeyPattern{Key: KeyUp, Mod: ModCtrl},
		},
		"Key.Alt requires alt": {
			spec: KeyLeft.Alt(),
			want: KeyPattern{Key: KeyLeft, Mod: ModAlt},
		},
		"Key.Shift requires shift": {
			spec: KeyTab.Shift(),
			want: KeyPattern{Key: KeyTab, Mod: ModShift},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.spec.keyPattern()
			if got != tt.want {
				t.Errorf("keyPattern() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestKeySpec_ModifiersAccumulate(t *testing.T) {
	type tc struct {
		spec KeySpec
		want KeyPattern
	}

	tests := map[string]tc{
		"ctrl then alt": {
			spec: KeyUp.Ctrl().Alt(),
			want: KeyPattern{Key: KeyUp, Mod: ModCtrl | ModAlt},
		},
		"ctrl then shift": {
			spec: KeyEnter.Ctrl().Shift(),
			want: KeyPattern{Key: KeyEnter, Mod: ModCtrl | ModShift},
		},
		"alt then shift": {
			spec: KeyDown.Alt().Shift(),
			want: KeyPattern{Key: KeyDown, Mod: ModAlt | ModShift},
		},
		"all three modifiers": {
			spec: KeyRight.Ctrl().Alt().Shift(),
			want: KeyPattern{Key: KeyRight, Mod: ModCtrl | ModAlt | ModShift},
		},
		"repeated modifier is idempotent": {
			spec: KeyHome.Ctrl().Ctrl(),
			want: KeyPattern{Key: KeyHome, Mod: ModCtrl},
		},
		"zero-modifier spec excludes all modifiers": {
			spec: KeySpec{key: KeyEnd},
			want: KeyPattern{Key: KeyEnd, ExcludeMods: ModCtrl | ModAlt | ModShift},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.spec.keyPattern()
			if got != tt.want {
				t.Errorf("keyPattern() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestRuneSpec_KeyPattern(t *testing.T) {
	type tc struct {
		spec RuneSpec
		want KeyPattern
	}

	tests := map[string]tc{
		"bare rune excludes ctrl and alt but allows shift": {
			spec: Rune('a'),
			want: KeyPattern{Rune: 'a', ExcludeMods: ModCtrl | ModAlt},
		},
		"unicode rune": {
			spec: Rune('日'),
			want: KeyPattern{Rune: '日', ExcludeMods: ModCtrl | ModAlt},
		},
		"ctrl modifier": {
			spec: Rune('c').Ctrl(),
			want: KeyPattern{Rune: 'c', Mod: ModCtrl},
		},
		"alt modifier": {
			spec: Rune('x').Alt(),
			want: KeyPattern{Rune: 'x', Mod: ModAlt},
		},
		"shift modifier": {
			spec: Rune('s').Shift(),
			want: KeyPattern{Rune: 's', Mod: ModShift},
		},
		"ctrl and alt accumulate": {
			spec: Rune('k').Ctrl().Alt(),
			want: KeyPattern{Rune: 'k', Mod: ModCtrl | ModAlt},
		},
		"all three modifiers accumulate": {
			spec: Rune('z').Ctrl().Alt().Shift(),
			want: KeyPattern{Rune: 'z', Mod: ModCtrl | ModAlt | ModShift},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.spec.keyPattern()
			if got != tt.want {
				t.Errorf("keyPattern() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestCtrlLetterHelpers(t *testing.T) {
	type tc struct {
		spec RuneSpec
		want KeyPattern
	}

	tests := map[string]tc{
		"KeyCtrlA": {
			spec: KeyCtrlA,
			want: KeyPattern{Rune: 'a', Mod: ModCtrl},
		},
		"KeyCtrlZ": {
			spec: KeyCtrlZ,
			want: KeyPattern{Rune: 'z', Mod: ModCtrl},
		},
		"KeyCtrlSpace": {
			spec: KeyCtrlSpace,
			want: KeyPattern{Rune: ' ', Mod: ModCtrl},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.spec.keyPattern()
			if got != tt.want {
				t.Errorf("keyPattern() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAnyMatchers_KeyPattern(t *testing.T) {
	type tc struct {
		matcher KeyMatcher
		want    KeyPattern
	}

	tests := map[string]tc{
		"AnyRune matches printables but excludes ctrl and alt": {
			matcher: AnyRune,
			want:    KeyPattern{AnyRune: true, ExcludeMods: ModCtrl | ModAlt},
		},
		"AnyKey matches everything": {
			matcher: AnyKey,
			want:    KeyPattern{AnyKey: true},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.matcher.keyPattern()
			if got != tt.want {
				t.Errorf("keyPattern() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
