package tui

// TextAreaOption configures a TextArea.
type TextAreaOption func(*TextArea)

// --- Sizing Options ---

// WithTextAreaWidth sets the text area width in characters.
func WithTextAreaWidth(cells int) TextAreaOption {
	return func(t *TextArea) {
		t.width = cells
	}
}

// WithTextAreaMaxHeight sets the maximum height in rows (0 = unlimited).
func WithTextAreaMaxHeight(rows int) TextAreaOption {
	return func(t *TextArea) {
		t.maxHeight = rows
	}
}

// --- Visual Options ---

// WithTextAreaBorder sets the border style.
func WithTextAreaBorder(b BorderStyle) TextAreaOption {
	return func(t *TextArea) {
		t.border = b
	}
}

// WithTextAreaTextStyle sets the text style.
func WithTextAreaTextStyle(s Style) TextAreaOption {
	return func(t *TextArea) {
		t.textStyle = s
	}
}

// WithTextAreaPlaceholder sets placeholder text shown when empty and unfocused.
func WithTextAreaPlaceholder(text string) TextAreaOption {
	return func(t *TextArea) {
		t.placeholder = text
	}
}

// WithTextAreaPlaceholderStyle sets the placeholder text style (defaults to dim).
func WithTextAreaPlaceholderStyle(s Style) TextAreaOption {
	return func(t *TextArea) {
		t.placeholderStyle = s
	}
}

// WithTextAreaCursorRune sets the drawn cursor glyph used in virtual-cursor
// mode (defaults to '▌'). Only takes effect together with
// WithTextAreaVirtualCursor; the default real-cursor mode draws no glyph.
func WithTextAreaCursorRune(r rune) TextAreaOption {
	return func(t *TextArea) {
		t.cursorRune = r
	}
}

// WithTextAreaCursor sets the drawn cursor glyph.
//
// Deprecated: use WithTextAreaCursorRune. Retained as an alias so existing call
// sites keep compiling.
func WithTextAreaCursor(r rune) TextAreaOption {
	return WithTextAreaCursorRune(r)
}

// WithTextAreaVirtualCursor switches the text area to the drawn '▌' cursor glyph
// instead of the framework-driven real terminal cursor. Presence enables it; the
// glyph is customizable via WithTextAreaCursorRune. By default (option absent)
// the real terminal cursor is used and no glyph is drawn.
func WithTextAreaVirtualCursor() TextAreaOption {
	return func(t *TextArea) {
		t.hideVirtualCursor = false
	}
}

// WithTextAreaFocusColor sets the border color when focused.
func WithTextAreaFocusColor(c Color) TextAreaOption {
	return func(t *TextArea) {
		t.focusColor = &c
	}
}

// WithTextAreaBorderGradient sets a gradient for the border color when unfocused.
func WithTextAreaBorderGradient(g Gradient) TextAreaOption {
	return func(t *TextArea) {
		t.borderGradient = &g
	}
}

// WithTextAreaFocusGradient sets a gradient for the border color when focused.
// Takes priority over focusColor when set.
func WithTextAreaFocusGradient(g Gradient) TextAreaOption {
	return func(t *TextArea) {
		t.focusGradient = &g
	}
}

// --- Behavior Options ---

// WithTextAreaSubmitKey sets the key that triggers submit.
// Default is KeyEnter (Enter submits, Ctrl+J inserts newline).
// For long-form text, use a different key (e.g. a function key) so Enter inserts newline.
func WithTextAreaSubmitKey(k Key) TextAreaOption {
	return func(t *TextArea) {
		t.submitKey = k
	}
}

// WithTextAreaValue binds the TextArea to an external State for its text content.
// The TextArea reads from and writes to this state directly, enabling reactive
// two-way binding between the TextArea and the parent component.
func WithTextAreaValue(state *State[string]) TextAreaOption {
	return func(t *TextArea) {
		t.text = state
		t.cursorPos = NewState(len([]rune(state.Get())))
	}
}

// WithTextAreaAutoFocus sets whether the text area should automatically
// receive focus when the element tree is first applied.
func WithTextAreaAutoFocus(auto bool) TextAreaOption {
	return func(t *TextArea) {
		t.autoFocus = auto
	}
}

// WithTextAreaOnSubmit sets the callback called when the submit key is pressed.
func WithTextAreaOnSubmit(fn func(string)) TextAreaOption {
	return func(t *TextArea) {
		t.onSubmit = fn
	}
}
