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

// WithInputCursorRune sets the drawn cursor glyph used in virtual-cursor mode
// (defaults to '▌'). Only takes effect together with WithInputVirtualCursor; the
// default real-cursor mode draws no glyph.
func WithInputCursorRune(r rune) InputOption {
	return func(inp *Input) {
		inp.cursorRune = r
	}
}

// WithInputCursor sets the drawn cursor glyph.
//
// Deprecated: use WithInputCursorRune. Retained as an alias so existing call
// sites keep compiling.
func WithInputCursor(r rune) InputOption {
	return WithInputCursorRune(r)
}

// WithInputVirtualCursor switches the input to the drawn '▌' cursor glyph
// instead of the framework-driven real terminal cursor. Presence enables it; the
// glyph is customizable via WithInputCursorRune. By default (option absent) the
// real terminal cursor is used and no glyph is drawn.
func WithInputVirtualCursor() InputOption {
	return func(inp *Input) {
		inp.hideVirtualCursor = false
	}
}

// WithInputValue binds the Input to an external State for its text content.
// The Input reads from and writes to this state directly, enabling reactive
// two-way binding between the Input and the parent component.
func WithInputValue(state *State[string]) InputOption {
	return func(inp *Input) {
		inp.text = state
		inp.cursorPos = NewState(len([]rune(state.Get())))
		inp.scrollPos = NewState(0)
	}
}

// WithInputFocusColor sets the border color when focused.
func WithInputFocusColor(c Color) InputOption {
	return func(inp *Input) {
		inp.focusColor = &c
	}
}

// WithInputBorderGradient sets a gradient for the border color when unfocused.
func WithInputBorderGradient(g Gradient) InputOption {
	return func(inp *Input) {
		inp.borderGradient = &g
	}
}

// WithInputFocusGradient sets a gradient for the border color when focused.
// Takes priority over focusColor when set.
func WithInputFocusGradient(g Gradient) InputOption {
	return func(inp *Input) {
		inp.focusGradient = &g
	}
}

// WithInputAutoFocus sets whether the input should automatically receive
// focus when the element tree is first applied.
func WithInputAutoFocus(auto bool) InputOption {
	return func(inp *Input) {
		inp.autoFocus = auto
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
