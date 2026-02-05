package tui

import "testing"

func TestKeyMap_OnKey(t *testing.T) {
	type tc struct {
		key      Key
		wantKey  Key
		wantStop bool
	}

	tests := map[string]tc{
		"OnKey sets key and broadcast": {
			key:      KeyEscape,
			wantKey:  KeyEscape,
			wantStop: false,
		},
		"OnKey with ctrl key": {
			key:      KeyCtrlC,
			wantKey:  KeyCtrlC,
			wantStop: false,
		},
		"OnKey with enter": {
			key:      KeyEnter,
			wantKey:  KeyEnter,
			wantStop: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := func(ke KeyEvent) {}
			binding := OnKey(tt.key, handler)
			if binding.Pattern.Key != tt.wantKey {
				t.Errorf("OnKey(%v).Pattern.Key = %v, want %v", tt.key, binding.Pattern.Key, tt.wantKey)
			}
			if binding.Stop != tt.wantStop {
				t.Errorf("OnKey(%v).Stop = %v, want %v", tt.key, binding.Stop, tt.wantStop)
			}
			if binding.Handler == nil {
				t.Error("OnKey().Handler should not be nil")
			}
		})
	}
}

func TestKeyMap_OnKeyStop(t *testing.T) {
	type tc struct {
		key      Key
		wantKey  Key
		wantStop bool
	}

	tests := map[string]tc{
		"OnKeyStop sets key and stop": {
			key:      KeyBackspace,
			wantKey:  KeyBackspace,
			wantStop: true,
		},
		"OnKeyStop with escape": {
			key:      KeyEscape,
			wantKey:  KeyEscape,
			wantStop: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := func(ke KeyEvent) {}
			binding := OnKeyStop(tt.key, handler)
			if binding.Pattern.Key != tt.wantKey {
				t.Errorf("OnKeyStop(%v).Pattern.Key = %v, want %v", tt.key, binding.Pattern.Key, tt.wantKey)
			}
			if binding.Stop != tt.wantStop {
				t.Errorf("OnKeyStop(%v).Stop = %v, want %v", tt.key, binding.Stop, tt.wantStop)
			}
		})
	}
}

func TestKeyMap_OnRune(t *testing.T) {
	type tc struct {
		r        rune
		wantRune rune
		wantStop bool
	}

	tests := map[string]tc{
		"OnRune sets rune and broadcast": {
			r:        'q',
			wantRune: 'q',
			wantStop: false,
		},
		"OnRune with slash": {
			r:        '/',
			wantRune: '/',
			wantStop: false,
		},
		"OnRune with unicode": {
			r:        '日',
			wantRune: '日',
			wantStop: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := func(ke KeyEvent) {}
			binding := OnRune(tt.r, handler)
			if binding.Pattern.Rune != tt.wantRune {
				t.Errorf("OnRune(%q).Pattern.Rune = %q, want %q", tt.r, binding.Pattern.Rune, tt.wantRune)
			}
			if binding.Stop != tt.wantStop {
				t.Errorf("OnRune(%q).Stop = %v, want %v", tt.r, binding.Stop, tt.wantStop)
			}
			if binding.Pattern.AnyRune {
				t.Errorf("OnRune(%q).Pattern.AnyRune should be false", tt.r)
			}
		})
	}
}

func TestKeyMap_OnRuneStop(t *testing.T) {
	type tc struct {
		r        rune
		wantRune rune
		wantStop bool
	}

	tests := map[string]tc{
		"OnRuneStop sets rune and stop": {
			r:        'x',
			wantRune: 'x',
			wantStop: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := func(ke KeyEvent) {}
			binding := OnRuneStop(tt.r, handler)
			if binding.Pattern.Rune != tt.wantRune {
				t.Errorf("OnRuneStop(%q).Pattern.Rune = %q, want %q", tt.r, binding.Pattern.Rune, tt.wantRune)
			}
			if binding.Stop != tt.wantStop {
				t.Errorf("OnRuneStop(%q).Stop = %v, want %v", tt.r, binding.Stop, tt.wantStop)
			}
			if binding.Pattern.AnyRune {
				t.Errorf("OnRuneStop(%q).Pattern.AnyRune should be false", tt.r)
			}
		})
	}
}

func TestKeyMap_OnRunes(t *testing.T) {
	type tc struct {
		wantAnyRune bool
		wantStop    bool
	}

	tests := map[string]tc{
		"OnRunes sets AnyRune and broadcast": {
			wantAnyRune: true,
			wantStop:    false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := func(ke KeyEvent) {}
			binding := OnRunes(handler)
			if binding.Pattern.AnyRune != tt.wantAnyRune {
				t.Errorf("OnRunes().Pattern.AnyRune = %v, want %v", binding.Pattern.AnyRune, tt.wantAnyRune)
			}
			if binding.Stop != tt.wantStop {
				t.Errorf("OnRunes().Stop = %v, want %v", binding.Stop, tt.wantStop)
			}
			if binding.Pattern.Key != KeyNone {
				t.Errorf("OnRunes().Pattern.Key = %v, want KeyNone", binding.Pattern.Key)
			}
			if binding.Pattern.Rune != 0 {
				t.Errorf("OnRunes().Pattern.Rune = %q, want 0", binding.Pattern.Rune)
			}
		})
	}
}

func TestKeyMap_OnRunesStop(t *testing.T) {
	type tc struct {
		wantAnyRune bool
		wantStop    bool
	}

	tests := map[string]tc{
		"OnRunesStop sets AnyRune and stop": {
			wantAnyRune: true,
			wantStop:    true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := func(ke KeyEvent) {}
			binding := OnRunesStop(handler)
			if binding.Pattern.AnyRune != tt.wantAnyRune {
				t.Errorf("OnRunesStop().Pattern.AnyRune = %v, want %v", binding.Pattern.AnyRune, tt.wantAnyRune)
			}
			if binding.Stop != tt.wantStop {
				t.Errorf("OnRunesStop().Stop = %v, want %v", binding.Stop, tt.wantStop)
			}
		})
	}
}

func TestKeyMap_PatternFields(t *testing.T) {
	type tc struct {
		binding     KeyBinding
		wantKey     Key
		wantRune    rune
		wantAnyRune bool
		wantMods    Modifier
		wantStop    bool
	}

	tests := map[string]tc{
		"OnKey only sets Key field": {
			binding:     OnKey(KeyCtrlB, func(ke KeyEvent) {}),
			wantKey:     KeyCtrlB,
			wantRune:    0,
			wantAnyRune: false,
			wantMods:    ModNone,
			wantStop:    false,
		},
		"OnRune only sets Rune field": {
			binding:     OnRune('/', func(ke KeyEvent) {}),
			wantKey:     KeyNone,
			wantRune:    '/',
			wantAnyRune: false,
			wantMods:    ModNone,
			wantStop:    false,
		},
		"OnRunes only sets AnyRune field": {
			binding:     OnRunes(func(ke KeyEvent) {}),
			wantKey:     KeyNone,
			wantRune:    0,
			wantAnyRune: true,
			wantMods:    ModNone,
			wantStop:    false,
		},
		"OnKeyStop sets Key and Stop": {
			binding:     OnKeyStop(KeyEscape, func(ke KeyEvent) {}),
			wantKey:     KeyEscape,
			wantRune:    0,
			wantAnyRune: false,
			wantMods:    ModNone,
			wantStop:    true,
		},
		"OnRuneStop sets Rune and Stop": {
			binding:     OnRuneStop('q', func(ke KeyEvent) {}),
			wantKey:     KeyNone,
			wantRune:    'q',
			wantAnyRune: false,
			wantMods:    ModNone,
			wantStop:    true,
		},
		"OnRunesStop sets AnyRune and Stop": {
			binding:     OnRunesStop(func(ke KeyEvent) {}),
			wantKey:     KeyNone,
			wantRune:    0,
			wantAnyRune: true,
			wantMods:    ModNone,
			wantStop:    true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			p := tt.binding.Pattern
			if p.Key != tt.wantKey {
				t.Errorf("Pattern.Key = %v, want %v", p.Key, tt.wantKey)
			}
			if p.Rune != tt.wantRune {
				t.Errorf("Pattern.Rune = %q, want %q", p.Rune, tt.wantRune)
			}
			if p.AnyRune != tt.wantAnyRune {
				t.Errorf("Pattern.AnyRune = %v, want %v", p.AnyRune, tt.wantAnyRune)
			}
			if p.Mod != tt.wantMods {
				t.Errorf("Pattern.Mod = %v, want %v", p.Mod, tt.wantMods)
			}
			if tt.binding.Stop != tt.wantStop {
				t.Errorf("Stop = %v, want %v", tt.binding.Stop, tt.wantStop)
			}
			if tt.binding.Handler == nil {
				t.Error("Handler should not be nil")
			}
		})
	}
}

func TestKeyMap_SliceType(t *testing.T) {
	type tc struct {
		km   KeyMap
		want int
	}

	tests := map[string]tc{
		"nil keymap": {
			km:   nil,
			want: 0,
		},
		"empty keymap": {
			km:   KeyMap{},
			want: 0,
		},
		"single binding": {
			km: KeyMap{
				OnKey(KeyCtrlC, func(ke KeyEvent) {}),
			},
			want: 1,
		},
		"multiple bindings": {
			km: KeyMap{
				OnKey(KeyCtrlC, func(ke KeyEvent) {}),
				OnRune('/', func(ke KeyEvent) {}),
				OnRunesStop(func(ke KeyEvent) {}),
				OnKeyStop(KeyEscape, func(ke KeyEvent) {}),
			},
			want: 4,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if len(tt.km) != tt.want {
				t.Errorf("len(KeyMap) = %d, want %d", len(tt.km), tt.want)
			}
		})
	}
}

func TestKeyMap_KeyPatternEquality(t *testing.T) {
	type tc struct {
		a    KeyPattern
		b    KeyPattern
		want bool
	}

	tests := map[string]tc{
		"same key patterns are equal": {
			a:    KeyPattern{Key: KeyEscape},
			b:    KeyPattern{Key: KeyEscape},
			want: true,
		},
		"different key patterns are not equal": {
			a:    KeyPattern{Key: KeyEscape},
			b:    KeyPattern{Key: KeyEnter},
			want: false,
		},
		"same rune patterns are equal": {
			a:    KeyPattern{Rune: 'q'},
			b:    KeyPattern{Rune: 'q'},
			want: true,
		},
		"different rune patterns are not equal": {
			a:    KeyPattern{Rune: 'q'},
			b:    KeyPattern{Rune: 'w'},
			want: false,
		},
		"same AnyRune patterns are equal": {
			a:    KeyPattern{AnyRune: true},
			b:    KeyPattern{AnyRune: true},
			want: true,
		},
		"AnyRune vs specific rune not equal": {
			a:    KeyPattern{AnyRune: true},
			b:    KeyPattern{Rune: 'q'},
			want: false,
		},
		"same mod are equal": {
			a:    KeyPattern{Key: KeyTab, Mod: ModShift},
			b:    KeyPattern{Key: KeyTab, Mod: ModShift},
			want: true,
		},
		"different mod are not equal": {
			a:    KeyPattern{Key: KeyTab, Mod: ModShift},
			b:    KeyPattern{Key: KeyTab, Mod: ModCtrl},
			want: false,
		},
		"RequireNoMods patterns are equal": {
			a:    KeyPattern{Key: KeyTab, RequireNoMods: true},
			b:    KeyPattern{Key: KeyTab, RequireNoMods: true},
			want: true,
		},
		"RequireNoMods vs Mod not equal": {
			a:    KeyPattern{Key: KeyTab, RequireNoMods: true},
			b:    KeyPattern{Key: KeyTab, Mod: ModShift},
			want: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.a == tt.b
			if got != tt.want {
				t.Errorf("(%+v == %+v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
