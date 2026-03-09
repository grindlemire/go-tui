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

// WithTextAreaCursor sets the cursor character (defaults to '▌').
func WithTextAreaCursor(r rune) TextAreaOption {
	return func(t *TextArea) {
		t.cursorRune = r
	}
}

// WithTextAreaFocusColor sets the border color when focused (defaults to Cyan).
func WithTextAreaFocusColor(c Color) TextAreaOption {
	return func(t *TextArea) {
		t.focusColor = c
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
// For long-form text, use a different key like KeyCtrlS (Ctrl+S submits, Enter inserts newline).
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

// WithTextAreaOnSubmit sets the callback called when the submit key is pressed.
func WithTextAreaOnSubmit(fn func(string)) TextAreaOption {
	return func(t *TextArea) {
		t.onSubmit = fn
	}
}
