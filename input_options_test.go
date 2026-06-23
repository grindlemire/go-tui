package tui

import (
	"testing"
)

func TestInputOptions_ConfigureFields(t *testing.T) {
	boldStyle := NewStyle().Bold()
	italicStyle := NewStyle().Italic()
	borderGrad := NewGradient(Red, Blue)
	focusGrad := NewGradient(Green, Yellow)

	type tc struct {
		opt   InputOption
		check func(t *testing.T, inp *Input)
	}

	tests := map[string]tc{
		"WithInputWidth sets width": {
			opt: WithInputWidth(42),
			check: func(t *testing.T, inp *Input) {
				if inp.width != 42 {
					t.Errorf("width = %d, want 42", inp.width)
				}
			},
		},
		"WithInputBorder sets border style": {
			opt: WithInputBorder(BorderDouble),
			check: func(t *testing.T, inp *Input) {
				if inp.border != BorderDouble {
					t.Errorf("border = %v, want BorderDouble", inp.border)
				}
			},
		},
		"WithInputTextStyle sets text style": {
			opt: WithInputTextStyle(boldStyle),
			check: func(t *testing.T, inp *Input) {
				if inp.textStyle != boldStyle {
					t.Errorf("textStyle = %+v, want %+v", inp.textStyle, boldStyle)
				}
			},
		},
		"WithInputPlaceholder sets placeholder text": {
			opt: WithInputPlaceholder("type here"),
			check: func(t *testing.T, inp *Input) {
				if inp.placeholder != "type here" {
					t.Errorf("placeholder = %q, want %q", inp.placeholder, "type here")
				}
			},
		},
		"WithInputPlaceholderStyle overrides dim default": {
			opt: WithInputPlaceholderStyle(italicStyle),
			check: func(t *testing.T, inp *Input) {
				if inp.placeholderStyle != italicStyle {
					t.Errorf("placeholderStyle = %+v, want %+v", inp.placeholderStyle, italicStyle)
				}
			},
		},
		"WithInputCursorRune sets cursor rune": {
			opt: WithInputCursorRune('_'),
			check: func(t *testing.T, inp *Input) {
				if inp.cursorRune != '_' {
					t.Errorf("cursorRune = %q, want '_'", inp.cursorRune)
				}
			},
		},
		"WithInputCursor alias sets cursor rune": {
			opt: WithInputCursor('#'),
			check: func(t *testing.T, inp *Input) {
				if inp.cursorRune != '#' {
					t.Errorf("cursorRune = %q, want '#'", inp.cursorRune)
				}
			},
		},
		"WithInputVirtualCursor enables drawn glyph": {
			opt: WithInputVirtualCursor(),
			check: func(t *testing.T, inp *Input) {
				if inp.hideVirtualCursor {
					t.Error("expected hideVirtualCursor to be false after WithInputVirtualCursor()")
				}
			},
		},
		"WithInputFocusColor sets focus color": {
			opt: WithInputFocusColor(Cyan),
			check: func(t *testing.T, inp *Input) {
				if inp.focusColor == nil || *inp.focusColor != Cyan {
					t.Errorf("focusColor = %+v, want Cyan", inp.focusColor)
				}
			},
		},
		"WithInputBorderGradient sets unfocused gradient": {
			opt: WithInputBorderGradient(borderGrad),
			check: func(t *testing.T, inp *Input) {
				if inp.borderGradient == nil || *inp.borderGradient != borderGrad {
					t.Errorf("borderGradient = %+v, want %+v", inp.borderGradient, borderGrad)
				}
			},
		},
		"WithInputFocusGradient sets focused gradient": {
			opt: WithInputFocusGradient(focusGrad),
			check: func(t *testing.T, inp *Input) {
				if inp.focusGradient == nil || *inp.focusGradient != focusGrad {
					t.Errorf("focusGradient = %+v, want %+v", inp.focusGradient, focusGrad)
				}
			},
		},
		"WithInputAutoFocus enables auto focus": {
			opt: WithInputAutoFocus(true),
			check: func(t *testing.T, inp *Input) {
				if !inp.autoFocus {
					t.Error("autoFocus = false, want true")
				}
				root := inp.Render(testApp)
				if !root.IsAutoFocus() {
					t.Error("rendered element should request auto focus")
				}
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inp := NewInput(tt.opt)
			inp.BindApp(testApp)
			tt.check(t, inp)
		})
	}
}

func TestInputOptions_WithInputValue(t *testing.T) {
	type tc struct {
		initial    string
		wantCursor int
	}

	tests := map[string]tc{
		"empty state starts cursor at zero":  {initial: "", wantCursor: 0},
		"prefilled state puts cursor at end": {initial: "hello", wantCursor: 5},
		"multibyte state counts runes":       {initial: "a界🙂", wantCursor: 3},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			external := NewState(tt.initial)
			inp := NewInput(WithInputValue(external))
			inp.BindApp(testApp)

			if inp.text != external {
				t.Fatal("input should use the external state instance directly")
			}
			if got := inp.Text(); got != tt.initial {
				t.Errorf("Text() = %q, want %q", got, tt.initial)
			}
			if got := inp.cursorPos.Get(); got != tt.wantCursor {
				t.Errorf("cursorPos = %d, want %d", got, tt.wantCursor)
			}
			if got := inp.scrollPos.Get(); got != 0 {
				t.Errorf("scrollPos = %d, want 0", got)
			}

			// Two-way binding: typing into the input updates the external state.
			inp.HandleEvent(KeyEvent{Key: KeyRune, Rune: '!'})
			if got := external.Get(); got != tt.initial+"!" {
				t.Errorf("external state = %q, want %q", got, tt.initial+"!")
			}
		})
	}
}

func TestInputOptions_Callbacks(t *testing.T) {
	type tc struct {
		makeOpt   func(record *string) InputOption
		event     KeyEvent
		preset    string
		wantValue string
	}

	tests := map[string]tc{
		"WithInputOnSubmit fires on enter": {
			makeOpt: func(record *string) InputOption {
				return WithInputOnSubmit(func(s string) { *record = "submit:" + s })
			},
			preset:    "done",
			event:     KeyEvent{Key: KeyEnter},
			wantValue: "submit:done",
		},
		"WithInputOnChange fires on insert": {
			makeOpt: func(record *string) InputOption {
				return WithInputOnChange(func(s string) { *record = "change:" + s })
			},
			preset:    "a",
			event:     KeyEvent{Key: KeyRune, Rune: 'b'},
			wantValue: "change:ab",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var record string
			inp := NewInput(tt.makeOpt(&record))
			inp.BindApp(testApp)
			inp.SetText(tt.preset)

			inp.HandleEvent(tt.event)
			if record != tt.wantValue {
				t.Errorf("callback recorded %q, want %q", record, tt.wantValue)
			}
		})
	}
}
