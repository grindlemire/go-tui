package tui

// InputOption configures an Input.
type InputOption func(*Input)

// WithInputWidth sets the input width in characters.
func WithInputWidth(cells int) InputOption {
	return func(inp *Input) {
		inp.width = cells
	}
}

// WithInputBorder sets the border style.
func WithInputBorder(b BorderStyle) InputOption {
	return func(inp *Input) {
		inp.border = b
	}
}

// WithInputTextStyle sets the text style.
func WithInputTextStyle(s Style) InputOption {
	return func(inp *Input) {
		inp.textStyle = s
	}
}

// WithInputPlaceholder sets placeholder text shown when empty and unfocused.
func WithInputPlaceholder(text string) InputOption {
	return func(inp *Input) {
		inp.placeholder = text
	}
}

// WithInputPlaceholderStyle sets the placeholder text style (defaults to dim).
func WithInputPlaceholderStyle(s Style) InputOption {
	return func(inp *Input) {
		inp.placeholderStyle = s
	}
}

// WithInputCursor sets the cursor character (defaults to '▌').
func WithInputCursor(r rune) InputOption {
	return func(inp *Input) {
		inp.cursorRune = r
	}
}

// WithInputValue binds the Input to an external State for its text content.
// The Input reads from and writes to this state directly, enabling reactive
// two-way binding between the Input and the parent component.
func WithInputValue(state *State[string]) InputOption {
	return func(inp *Input) {
		inp.text = state
		inp.cursorPos = NewState(len([]rune(state.Get())))
	}
}

// WithInputFocusColor sets the border color when focused (defaults to Cyan).
func WithInputFocusColor(c Color) InputOption {
	return func(inp *Input) {
		inp.focusColor = c
	}
}

// WithInputOnSubmit sets the callback called when Enter is pressed.
func WithInputOnSubmit(fn func(string)) InputOption {
	return func(inp *Input) {
		inp.onSubmit = fn
	}
}

// WithInputOnChange sets the callback called when the text changes.
func WithInputOnChange(fn func(string)) InputOption {
	return func(inp *Input) {
		inp.onChange = fn
	}
}
