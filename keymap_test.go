package tui

import "testing"

func TestKeyMap_On_Key(t *testing.T) {
	type tc struct {
		key      Key
		wantKey  Key
		wantStop bool
	}

	tests := map[string]tc{
		"On sets key and broadcast": {
			key:      KeyEscape,
			wantKey:  KeyEscape,
			wantStop: false,
		},
		"On with function key": {
			key:      KeyF1,
			wantKey:  KeyF1,
			wantStop: false,
		},
		"On with enter": {
			key:      KeyEnter,
			wantKey:  KeyEnter,
			wantStop: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := func(ke KeyEvent) {}
			binding := On(tt.key, handler)
			if binding.Pattern.Key != tt.wantKey {
				t.Errorf("On(%v).Pattern.Key = %v, want %v", tt.key, binding.Pattern.Key, tt.wantKey)
			}
			if binding.Stop != tt.wantStop {
				t.Errorf("On(%v).Stop = %v, want %v", tt.key, binding.Stop, tt.wantStop)
			}
			if binding.Handler == nil {
				t.Error("On().Handler should not be nil")
			}
		})
	}
}

func TestKeyMap_OnStop_Key(t *testing.T) {
	type tc struct {
		key      Key
		wantKey  Key
		wantStop bool
	}

	tests := map[string]tc{
		"OnStop sets key and stop": {
			key:      KeyBackspace,
			wantKey:  KeyBackspace,
			wantStop: true,
		},
		"OnStop with escape": {
			key:      KeyEscape,
			wantKey:  KeyEscape,
			wantStop: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := func(ke KeyEvent) {}
			binding := OnStop(tt.key, handler)
			if binding.Pattern.Key != tt.wantKey {
				t.Errorf("OnStop(%v).Pattern.Key = %v, want %v", tt.key, binding.Pattern.Key, tt.wantKey)
			}
			if binding.Stop != tt.wantStop {
				t.Errorf("OnStop(%v).Stop = %v, want %v", tt.key, binding.Stop, tt.wantStop)
			}
		})
	}
}

func TestKeyMap_On_Rune(t *testing.T) {
	type tc struct {
		r        rune
		wantRune rune
		wantStop bool
	}

	tests := map[string]tc{
		"On(Rune) sets rune and broadcast": {
			r:        'q',
			wantRune: 'q',
			wantStop: false,
		},
		"On(Rune) with slash": {
			r:        '/',
			wantRune: '/',
			wantStop: false,
		},
		"On(Rune) with unicode": {
			r:        '日',
			wantRune: '日',
			wantStop: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := func(ke KeyEvent) {}
			binding := On(Rune(tt.r), handler)
			if binding.Pattern.Rune != tt.wantRune {
				t.Errorf("On(Rune(%q)).Pattern.Rune = %q, want %q", tt.r, binding.Pattern.Rune, tt.wantRune)
			}
			if binding.Stop != tt.wantStop {
				t.Errorf("On(Rune(%q)).Stop = %v, want %v", tt.r, binding.Stop, tt.wantStop)
			}
			if binding.Pattern.AnyRune {
				t.Errorf("On(Rune(%q)).Pattern.AnyRune should be false", tt.r)
			}
		})
	}
}

func TestKeyMap_OnStop_Rune(t *testing.T) {
	type tc struct {
		r        rune
		wantRune rune
		wantStop bool
	}

	tests := map[string]tc{
		"OnStop(Rune) sets rune and stop": {
			r:        'x',
			wantRune: 'x',
			wantStop: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := func(ke KeyEvent) {}
			binding := OnStop(Rune(tt.r), handler)
			if binding.Pattern.Rune != tt.wantRune {
				t.Errorf("OnStop(Rune(%q)).Pattern.Rune = %q, want %q", tt.r, binding.Pattern.Rune, tt.wantRune)
			}
			if binding.Stop != tt.wantStop {
				t.Errorf("OnStop(Rune(%q)).Stop = %v, want %v", tt.r, binding.Stop, tt.wantStop)
			}
			if binding.Pattern.AnyRune {
				t.Errorf("OnStop(Rune(%q)).Pattern.AnyRune should be false", tt.r)
			}
		})
	}
}

func TestKeyMap_On_AnyRune(t *testing.T) {
	type tc struct {
		wantAnyRune bool
		wantStop    bool
	}

	tests := map[string]tc{
		"On(AnyRune) sets AnyRune and broadcast": {
			wantAnyRune: true,
			wantStop:    false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := func(ke KeyEvent) {}
			binding := On(AnyRune, handler)
			if binding.Pattern.AnyRune != tt.wantAnyRune {
				t.Errorf("On(AnyRune).Pattern.AnyRune = %v, want %v", binding.Pattern.AnyRune, tt.wantAnyRune)
			}
			if binding.Stop != tt.wantStop {
				t.Errorf("On(AnyRune).Stop = %v, want %v", binding.Stop, tt.wantStop)
			}
			if binding.Pattern.Key != KeyNone {
				t.Errorf("On(AnyRune).Pattern.Key = %v, want KeyNone", binding.Pattern.Key)
			}
			if binding.Pattern.Rune != 0 {
				t.Errorf("On(AnyRune).Pattern.Rune = %q, want 0", binding.Pattern.Rune)
			}
		})
	}
}

func TestKeyMap_OnStop_AnyRune(t *testing.T) {
	type tc struct {
		wantAnyRune bool
		wantStop    bool
	}

	tests := map[string]tc{
		"OnStop(AnyRune) sets AnyRune and stop": {
			wantAnyRune: true,
			wantStop:    true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := func(ke KeyEvent) {}
			binding := OnStop(AnyRune, handler)
			if binding.Pattern.AnyRune != tt.wantAnyRune {
				t.Errorf("OnStop(AnyRune).Pattern.AnyRune = %v, want %v", binding.Pattern.AnyRune, tt.wantAnyRune)
			}
			if binding.Stop != tt.wantStop {
				t.Errorf("OnStop(AnyRune).Stop = %v, want %v", binding.Stop, tt.wantStop)
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
		"On(Key) only sets Key field": {
			binding:     On(KeyF2, func(ke KeyEvent) {}),
			wantKey:     KeyF2,
			wantRune:    0,
			wantAnyRune: false,
			wantMods:    ModNone,
			wantStop:    false,
		},
		"On(Rune) only sets Rune field": {
			binding:     On(Rune('/'), func(ke KeyEvent) {}),
			wantKey:     KeyNone,
			wantRune:    '/',
			wantAnyRune: false,
			wantMods:    ModNone,
			wantStop:    false,
		},
		"On(AnyRune) only sets AnyRune field": {
			binding:     On(AnyRune, func(ke KeyEvent) {}),
			wantKey:     KeyNone,
			wantRune:    0,
			wantAnyRune: true,
			wantMods:    ModNone,
			wantStop:    false,
		},
		"OnStop(Key) sets Key and Stop": {
			binding:     OnStop(KeyEscape, func(ke KeyEvent) {}),
			wantKey:     KeyEscape,
			wantRune:    0,
			wantAnyRune: false,
			wantMods:    ModNone,
			wantStop:    true,
		},
		"OnStop(Rune) sets Rune and Stop": {
			binding:     OnStop(Rune('q'), func(ke KeyEvent) {}),
			wantKey:     KeyNone,
			wantRune:    'q',
			wantAnyRune: false,
			wantMods:    ModNone,
			wantStop:    true,
		},
		"OnStop(AnyRune) sets AnyRune and Stop": {
			binding:     OnStop(AnyRune, func(ke KeyEvent) {}),
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
				On(Rune('c').Ctrl(), func(ke KeyEvent) {}),
			},
			want: 1,
		},
		"multiple bindings": {
			km: KeyMap{
				On(Rune('c').Ctrl(), func(ke KeyEvent) {}),
				On(Rune('/'), func(ke KeyEvent) {}),
				OnStop(AnyRune, func(ke KeyEvent) {}),
				OnStop(KeyEscape, func(ke KeyEvent) {}),
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

func TestKeyMap_On_WithModifier(t *testing.T) {
	// Without modifier: ExcludeMods set
	binding := On(KeyTab, func(ke KeyEvent) {})
	if binding.Pattern.ExcludeMods != ModCtrl|ModAlt|ModShift {
		t.Errorf("On(KeyTab): ExcludeMods = %v, want all", binding.Pattern.ExcludeMods)
	}
	if binding.Pattern.Mod != ModNone {
		t.Errorf("On(KeyTab): Mod = %v, want None", binding.Pattern.Mod)
	}

	// With modifier via Key.Shift()
	binding = On(KeyTab.Shift(), func(ke KeyEvent) {})
	if binding.Pattern.Mod != ModShift {
		t.Errorf("On(KeyTab.Shift()): Mod = %v, want Shift", binding.Pattern.Mod)
	}
	if binding.Pattern.ExcludeMods != ModNone {
		t.Errorf("On(KeyTab.Shift()): ExcludeMods = %v, want None", binding.Pattern.ExcludeMods)
	}
}

func TestKeyMap_On_RuneWithModifier(t *testing.T) {
	// Without modifier: ExcludeMods set to Ctrl|Alt
	binding := On(Rune('a'), func(ke KeyEvent) {})
	if binding.Pattern.ExcludeMods != ModCtrl|ModAlt {
		t.Errorf("On(Rune('a')): ExcludeMods = %v, want Ctrl|Alt", binding.Pattern.ExcludeMods)
	}

	// With modifier via Rune.Ctrl()
	binding = On(Rune('a').Ctrl(), func(ke KeyEvent) {})
	if binding.Pattern.Mod != ModCtrl {
		t.Errorf("On(Rune('a').Ctrl()): Mod = %v, want Ctrl", binding.Pattern.Mod)
	}
	if binding.Pattern.ExcludeMods != ModNone {
		t.Errorf("On(Rune('a').Ctrl()): ExcludeMods = %v, want None", binding.Pattern.ExcludeMods)
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
		"ExcludeMods patterns are equal": {
			a:    KeyPattern{Key: KeyTab, ExcludeMods: ModCtrl | ModAlt | ModShift},
			b:    KeyPattern{Key: KeyTab, ExcludeMods: ModCtrl | ModAlt | ModShift},
			want: true,
		},
		"ExcludeMods vs Mod not equal": {
			a:    KeyPattern{Key: KeyTab, ExcludeMods: ModCtrl | ModAlt | ModShift},
			b:    KeyPattern{Key: KeyTab, Mod: ModShift},
			want: false,
		},
		"same ExcludeMods rune patterns are equal": {
			a:    KeyPattern{Rune: 'a', ExcludeMods: ModCtrl | ModAlt},
			b:    KeyPattern{Rune: 'a', ExcludeMods: ModCtrl | ModAlt},
			want: true,
		},
		"different ExcludeMods are not equal": {
			a:    KeyPattern{Rune: 'a', ExcludeMods: ModCtrl | ModAlt},
			b:    KeyPattern{Rune: 'a', ExcludeMods: ModCtrl},
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
