package tui

import "testing"

func TestTextAreaOptions(t *testing.T) {
	boldStyle := NewStyle().Bold()
	italicStyle := NewStyle().Italic()
	focusColor := Cyan
	borderGrad := NewGradient(Red, Blue)
	focusGrad := NewGradient(Green, Yellow).WithDirection(GradientVertical)
	boundState := NewState("héllo")

	submitted := ""
	changed := ""
	onSubmit := func(s string) { submitted = s }
	onChange := func(s string) { changed = s }

	type tc struct {
		opts   []TextAreaOption
		assert func(t *testing.T, ta *TextArea)
	}

	tests := map[string]tc{
		"defaults without options": {
			opts: nil,
			assert: func(t *testing.T, ta *TextArea) {
				if ta.width != 40 {
					t.Fatalf("width = %d, want 40", ta.width)
				}
				if ta.maxHeight != 0 {
					t.Fatalf("maxHeight = %d, want 0", ta.maxHeight)
				}
				if ta.border != BorderNone {
					t.Fatalf("border = %v, want BorderNone", ta.border)
				}
				if ta.placeholder != "" {
					t.Fatalf("placeholder = %q, want empty", ta.placeholder)
				}
				if ta.placeholderStyle != (Style{}.Dim()) {
					t.Fatal("placeholderStyle should default to dim")
				}
				if ta.cursorRune != '▌' {
					t.Fatalf("cursorRune = %q, want '▌'", ta.cursorRune)
				}
				if ta.submitKey != KeyEnter {
					t.Fatalf("submitKey = %v, want KeyEnter", ta.submitKey)
				}
				if !ta.hideVirtualCursor {
					t.Fatal("hideVirtualCursor should default to true (real cursor is the default)")
				}
				if ta.autoFocus {
					t.Fatal("autoFocus should default to false")
				}
				if ta.focusColor != nil || ta.borderGradient != nil || ta.focusGradient != nil {
					t.Fatal("color and gradient pointers should default to nil")
				}
				if ta.onSubmit != nil {
					t.Fatal("onSubmit should default to nil")
				}
				if ta.Text() != "" {
					t.Fatalf("text = %q, want empty", ta.Text())
				}
				if ta.cursorPos.Get() != 0 {
					t.Fatalf("cursorPos = %d, want 0", ta.cursorPos.Get())
				}
			},
		},
		"WithTextAreaWidth sets width": {
			opts: []TextAreaOption{WithTextAreaWidth(25)},
			assert: func(t *testing.T, ta *TextArea) {
				if ta.width != 25 {
					t.Fatalf("width = %d, want 25", ta.width)
				}
			},
		},
		"WithTextAreaMaxHeight sets max height": {
			opts: []TextAreaOption{WithTextAreaMaxHeight(5)},
			assert: func(t *testing.T, ta *TextArea) {
				if ta.maxHeight != 5 {
					t.Fatalf("maxHeight = %d, want 5", ta.maxHeight)
				}
			},
		},
		"WithTextAreaBorder sets border style": {
			opts: []TextAreaOption{WithTextAreaBorder(BorderRounded)},
			assert: func(t *testing.T, ta *TextArea) {
				if ta.border != BorderRounded {
					t.Fatalf("border = %v, want BorderRounded", ta.border)
				}
			},
		},
		"WithTextAreaTextStyle sets text style": {
			opts: []TextAreaOption{WithTextAreaTextStyle(boldStyle)},
			assert: func(t *testing.T, ta *TextArea) {
				if ta.textStyle != boldStyle {
					t.Fatalf("textStyle = %v, want bold", ta.textStyle)
				}
			},
		},
		"WithTextAreaPlaceholder sets placeholder text": {
			opts: []TextAreaOption{WithTextAreaPlaceholder("type here")},
			assert: func(t *testing.T, ta *TextArea) {
				if ta.placeholder != "type here" {
					t.Fatalf("placeholder = %q, want %q", ta.placeholder, "type here")
				}
			},
		},
		"WithTextAreaPlaceholderStyle overrides default dim": {
			opts: []TextAreaOption{WithTextAreaPlaceholderStyle(italicStyle)},
			assert: func(t *testing.T, ta *TextArea) {
				if ta.placeholderStyle != italicStyle {
					t.Fatalf("placeholderStyle = %v, want italic", ta.placeholderStyle)
				}
			},
		},
		"WithTextAreaCursorRune sets cursor rune": {
			opts: []TextAreaOption{WithTextAreaCursorRune('_')},
			assert: func(t *testing.T, ta *TextArea) {
				if ta.cursorRune != '_' {
					t.Fatalf("cursorRune = %q, want '_'", ta.cursorRune)
				}
			},
		},
		"WithTextAreaCursor alias sets cursor rune": {
			opts: []TextAreaOption{WithTextAreaCursor('#')},
			assert: func(t *testing.T, ta *TextArea) {
				if ta.cursorRune != '#' {
					t.Fatalf("cursorRune = %q, want '#'", ta.cursorRune)
				}
			},
		},
		"WithTextAreaVirtualCursor enables drawn glyph": {
			opts: []TextAreaOption{WithTextAreaVirtualCursor()},
			assert: func(t *testing.T, ta *TextArea) {
				if ta.hideVirtualCursor {
					t.Fatal("expected hideVirtualCursor to be false after WithTextAreaVirtualCursor()")
				}
			},
		},
		"WithTextAreaFocusColor sets focus color": {
			opts: []TextAreaOption{WithTextAreaFocusColor(focusColor)},
			assert: func(t *testing.T, ta *TextArea) {
				if ta.focusColor == nil {
					t.Fatal("expected focusColor to be set")
				}
				if *ta.focusColor != focusColor {
					t.Fatalf("focusColor = %v, want %v", *ta.focusColor, focusColor)
				}
			},
		},
		"WithTextAreaBorderGradient sets border gradient": {
			opts: []TextAreaOption{WithTextAreaBorderGradient(borderGrad)},
			assert: func(t *testing.T, ta *TextArea) {
				if ta.borderGradient == nil {
					t.Fatal("expected borderGradient to be set")
				}
				if *ta.borderGradient != borderGrad {
					t.Fatalf("borderGradient = %v, want %v", *ta.borderGradient, borderGrad)
				}
			},
		},
		"WithTextAreaFocusGradient sets focus gradient": {
			opts: []TextAreaOption{WithTextAreaFocusGradient(focusGrad)},
			assert: func(t *testing.T, ta *TextArea) {
				if ta.focusGradient == nil {
					t.Fatal("expected focusGradient to be set")
				}
				if *ta.focusGradient != focusGrad {
					t.Fatalf("focusGradient = %v, want %v", *ta.focusGradient, focusGrad)
				}
			},
		},
		"WithTextAreaSubmitKey overrides default": {
			opts: []TextAreaOption{WithTextAreaSubmitKey(KeyTab)},
			assert: func(t *testing.T, ta *TextArea) {
				if ta.submitKey != KeyTab {
					t.Fatalf("submitKey = %v, want KeyTab", ta.submitKey)
				}
			},
		},
		"WithTextAreaValue binds external state and places cursor at end": {
			opts: []TextAreaOption{WithTextAreaValue(boundState)},
			assert: func(t *testing.T, ta *TextArea) {
				if ta.text != boundState {
					t.Fatal("expected text state to be the bound state")
				}
				if ta.Text() != "héllo" {
					t.Fatalf("text = %q, want %q", ta.Text(), "héllo")
				}
				// Cursor lands at the end, counted in runes not bytes.
				if got := ta.cursorPos.Get(); got != 5 {
					t.Fatalf("cursorPos = %d, want 5", got)
				}
			},
		},
		"WithTextAreaAutoFocus enables auto focus": {
			opts: []TextAreaOption{WithTextAreaAutoFocus(true)},
			assert: func(t *testing.T, ta *TextArea) {
				if !ta.autoFocus {
					t.Fatal("expected autoFocus to be true")
				}
			},
		},
		"WithTextAreaOnSubmit stores a working callback": {
			opts: []TextAreaOption{WithTextAreaOnSubmit(onSubmit)},
			assert: func(t *testing.T, ta *TextArea) {
				if ta.onSubmit == nil {
					t.Fatal("expected onSubmit to be set")
				}
				ta.onSubmit("submitted text")
				if submitted != "submitted text" {
					t.Fatalf("submitted = %q, want %q", submitted, "submitted text")
				}
			},
		},
		"WithTextAreaOnChange stores callback and fires on text mutation": {
			opts: []TextAreaOption{WithTextAreaOnChange(onChange)},
			assert: func(t *testing.T, ta *TextArea) {
				if ta.onChange == nil {
					t.Fatal("expected onChange to be set")
				}
				// SetText fires onChange.
				changed = ""
				ta.SetText("hello")
				if changed != "hello" {
					t.Fatalf("after SetText: changed = %q, want %q", changed, "hello")
				}
				// Clear fires onChange.
				changed = "SENTINEL"
				ta.Clear()
				if changed != "" {
					t.Fatalf("after Clear: changed = %q, want ''", changed)
				}
				// insertChar fires onChange via insertString.
				changed = ""
				ta.SetText("ab")
				changed = ""
				// Simulate typing 'c' at the end (cluster index 2, rune index 2).
				ta.insertChar(KeyEvent{Rune: 'c'})
				if changed != "abc" {
					t.Fatalf("after insertChar: changed = %q, want %q", changed, "abc")
				}
				// Backspace fires onChange.
				changed = ""
				ta.SetText("abc")
				ta.cursorPos.Set(3) // at end
				changed = ""
				ta.backspace(KeyEvent{})
				if changed != "ab" {
					t.Fatalf("after backspace: changed = %q, want %q", changed, "ab")
				}
				// Delete fires onChange.
				changed = ""
				ta.SetText("abc")
				ta.cursorPos.Set(0)
				changed = ""
				ta.delete(KeyEvent{})
				if changed != "bc" {
					t.Fatalf("after delete: changed = %q, want %q", changed, "bc")
				}
				// Submit does NOT fire onChange.
				changed = ""
				ta.SetText("hello")
				changed = ""
				ta.submit(KeyEvent{})
				if changed != "" {
					t.Fatalf("after submit: onChange fired (changed = %q), want no call", changed)
				}
			},
		},
		"multiple options compose": {
			opts: []TextAreaOption{
				WithTextAreaWidth(30),
				WithTextAreaMaxHeight(4),
				WithTextAreaBorder(BorderDouble),
				WithTextAreaPlaceholder("compose"),
			},
			assert: func(t *testing.T, ta *TextArea) {
				if ta.width != 30 || ta.maxHeight != 4 {
					t.Fatalf("size = (%d, %d), want (30, 4)", ta.width, ta.maxHeight)
				}
				if ta.border != BorderDouble {
					t.Fatalf("border = %v, want BorderDouble", ta.border)
				}
				if ta.placeholder != "compose" {
					t.Fatalf("placeholder = %q, want %q", ta.placeholder, "compose")
				}
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ta := NewTextArea(tt.opts...)
			tt.assert(t, ta)
		})
	}
}
