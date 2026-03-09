package tui

import (
	"time"
	"unicode/utf8"
)

// Input is a single-line text input with cursor management.
// It implements Component, KeyListener, WatcherProvider, and Focusable interfaces.
type Input struct {
	// Configuration (set via options, immutable after construction)
	width            int
	border           BorderStyle
	textStyle        Style
	placeholder      string
	placeholderStyle Style
	cursorRune       rune
	focusColor       Color
	onSubmit         func(string)
	onChange         func(string)

	// Reactive state
	text       *State[string]
	cursorPos  *State[int]
	scrollPos  *State[int] // horizontal scroll offset (first visible rune index)
	blink      *State[bool]
	focused    *State[bool]
}

// Interface assertions
var (
	_ Component       = (*Input)(nil)
	_ KeyListener     = (*Input)(nil)
	_ WatcherProvider = (*Input)(nil)
	_ Focusable       = (*Input)(nil)
	_ AppBinder       = (*Input)(nil)
)

// BindApp binds this Input's internal States to the given app.
func (inp *Input) BindApp(app *App) {
	inp.text.BindApp(app)
	inp.cursorPos.BindApp(app)
	inp.scrollPos.BindApp(app)
	inp.blink.BindApp(app)
	inp.focused.BindApp(app)
}

// NewInput creates a new single-line text input.
func NewInput(opts ...InputOption) *Input {
	inp := &Input{
		// Defaults
		width:            20,
		border:           BorderNone,
		textStyle:        Style{},
		placeholder:      "",
		placeholderStyle: Style{}.Dim(),
		cursorRune:       '▌',
		focusColor:       Cyan,

		// State
		text:      NewState(""),
		cursorPos: NewState(0),
		scrollPos: NewState(0),
		blink:     NewState(true),
		focused:   NewState(false),
	}
	for _, opt := range opts {
		opt(inp)
	}
	return inp
}

// --- State Access ---

// Text returns the current text content.
func (inp *Input) Text() string {
	return inp.text.Get()
}

// SetText sets the text and moves cursor to end.
func (inp *Input) SetText(s string) {
	inp.text.Set(s)
	inp.cursorPos.Set(utf8.RuneCountInString(s))
}

// Clear clears the input.
func (inp *Input) Clear() {
	inp.text.Set("")
	inp.cursorPos.Set(0)
	inp.scrollPos.Set(0)
}

// --- Component Interface ---

// visibleWidth returns the number of characters visible inside the input.
// Accounts for border taking 1 char on each side.
func (inp *Input) visibleWidth() int {
	w := inp.width
	if inp.border != BorderNone {
		// Border chars are drawn inside the element width, reducing text space
		return w - 2
	}
	return w
}

// ensureCursorVisible adjusts scrollPos so the cursor is within the visible window.
func (inp *Input) ensureCursorVisible() {
	pos := inp.clampCursorPos()
	scroll := inp.scrollPos.Get()
	visible := inp.visibleWidth()
	if visible <= 0 {
		return
	}

	// Cursor is left of the visible window
	if pos < scroll {
		inp.scrollPos.Set(pos)
		return
	}

	// Cursor is right of the visible window.
	// Reserve 1 column for the cursor character itself.
	if pos >= scroll+visible {
		inp.scrollPos.Set(pos - visible + 1)
	}
}

// Render returns the element tree for the input.
func (inp *Input) Render(app *App) *Element {
	totalHeight := 1
	if inp.border != BorderNone {
		totalHeight += 2
	}

	opts := []Option{
		WithDirection(Row),
		WithHeight(totalHeight),
		WithFocusable(true),
	}
	if inp.width > 0 {
		opts = append(opts, WithWidth(inp.width))
	}
	if inp.border != BorderNone {
		opts = append(opts, WithBorder(inp.border))
		if inp.focused.Get() {
			opts = append(opts, WithBorderStyle(NewStyle().Foreground(inp.focusColor)))
		}
	}
	root := New(opts...)

	// Wire Element focus/blur to component focus/blur
	root.SetOnFocus(func(e *Element) {
		inp.Focus()
	})
	root.SetOnBlur(func(e *Element) {
		inp.Blur()
	})

	// Render placeholder or content
	if inp.text.Get() == "" && inp.placeholder != "" && !inp.focused.Get() {
		root.AddChild(New(WithText(inp.placeholder), WithTextStyle(inp.placeholderStyle)))
	} else {
		root.AddChild(New(WithText(inp.displayText()), WithTextStyle(inp.textStyle)))
	}

	return root
}

// --- Focusable Interface ---

// IsFocusable returns true since Input can receive focus.
func (inp *Input) IsFocusable() bool {
	return true
}

// IsTabStop returns true since Input participates in Tab navigation.
func (inp *Input) IsTabStop() bool {
	return true
}

// Focus is called when the input gains focus. Idempotent.
func (inp *Input) Focus() {
	if inp.focused.Get() {
		return
	}
	inp.focused.Set(true)
	inp.blink.Set(true)
}

// Blur is called when the input loses focus. Idempotent.
func (inp *Input) Blur() {
	if !inp.focused.Get() {
		return
	}
	inp.focused.Set(false)
}

// IsFocused returns whether this input is currently focused.
func (inp *Input) IsFocused() bool {
	return inp.focused.Get()
}

// HandleEvent processes keyboard events.
func (inp *Input) HandleEvent(e Event) bool {
	ke, ok := e.(KeyEvent)
	if !ok {
		return false
	}

	for _, binding := range inp.KeyMap() {
		if inp.matchesPattern(ke, binding.Pattern) {
			binding.Handler(ke)
			return binding.Stop
		}
	}
	return false
}

// matchesPattern checks if a key event matches a pattern.
func (inp *Input) matchesPattern(ke KeyEvent, p KeyPattern) bool {
	if p.Key != 0 && ke.Key == p.Key {
		return true
	}
	if p.Rune != 0 && ke.Rune == p.Rune {
		return true
	}
	if p.AnyRune && ke.Rune != 0 {
		return true
	}
	return false
}

// --- KeyListener Interface ---

// KeyMap returns the key bindings for the input.
func (inp *Input) KeyMap() KeyMap {
	return KeyMap{
		// Text input (focus-gated)
		OnRunesFocused(inp.insertChar),

		// Editing (focus-gated)
		OnKeyFocused(KeyBackspace, inp.backspace),
		OnKeyFocused(KeyDelete, inp.delete),

		// Navigation (focus-gated)
		OnKeyFocused(KeyLeft, inp.moveLeft),
		OnKeyFocused(KeyRight, inp.moveRight),
		OnKeyFocused(KeyHome, inp.moveHome),
		OnKeyFocused(KeyEnd, inp.moveEnd),

		// Submit (focus-gated)
		OnKeyFocused(KeyEnter, inp.submit),

		// Blur on Escape (focus-gated)
		OnKeyFocused(KeyEscape, func(ke KeyEvent) {
			if app := ke.App(); app != nil {
				app.BlurFocused()
			}
		}),
	}
}

// --- WatcherProvider Interface ---

// Watchers returns watchers for cursor blink.
func (inp *Input) Watchers() []Watcher {
	return []Watcher{
		OnTimer(500*time.Millisecond, func() {
			if inp.focused.Get() {
				inp.blink.Set(!inp.blink.Get())
			}
		}),
	}
}

// --- Key Handlers ---

// insertChar inserts a character at the cursor position.
func (inp *Input) insertChar(ke KeyEvent) {
	runes := []rune(inp.text.Get())
	pos := inp.clampCursorPos()
	newRunes := append(runes[:pos], append([]rune{ke.Rune}, runes[pos:]...)...)
	inp.text.Set(string(newRunes))
	inp.cursorPos.Set(pos + 1)
	inp.blink.Set(true)
	inp.ensureCursorVisible()
	if inp.onChange != nil {
		inp.onChange(inp.text.Get())
	}
}

// backspace deletes the character before the cursor.
func (inp *Input) backspace(ke KeyEvent) {
	runes := []rune(inp.text.Get())
	pos := inp.clampCursorPos()
	if pos > 0 {
		newRunes := append(runes[:pos-1], runes[pos:]...)
		inp.text.Set(string(newRunes))
		inp.cursorPos.Set(pos - 1)
		inp.ensureCursorVisible()
		if inp.onChange != nil {
			inp.onChange(inp.text.Get())
		}
	}
}

// delete deletes the character at the cursor.
func (inp *Input) delete(ke KeyEvent) {
	runes := []rune(inp.text.Get())
	pos := inp.clampCursorPos()
	if pos < len(runes) {
		newRunes := append(runes[:pos], runes[pos+1:]...)
		inp.text.Set(string(newRunes))
		inp.ensureCursorVisible()
		if inp.onChange != nil {
			inp.onChange(inp.text.Get())
		}
	}
}

// moveLeft moves cursor left.
func (inp *Input) moveLeft(ke KeyEvent) {
	pos := inp.cursorPos.Get()
	if pos > 0 {
		inp.cursorPos.Set(pos - 1)
		inp.blink.Set(true)
		inp.ensureCursorVisible()
	}
}

// moveRight moves cursor right.
func (inp *Input) moveRight(ke KeyEvent) {
	pos := inp.cursorPos.Get()
	if pos < utf8.RuneCountInString(inp.text.Get()) {
		inp.cursorPos.Set(pos + 1)
		inp.blink.Set(true)
		inp.ensureCursorVisible()
	}
}

// moveHome moves cursor to start.
func (inp *Input) moveHome(ke KeyEvent) {
	inp.cursorPos.Set(0)
	inp.blink.Set(true)
	inp.ensureCursorVisible()
}

// moveEnd moves cursor to end.
func (inp *Input) moveEnd(ke KeyEvent) {
	inp.cursorPos.Set(utf8.RuneCountInString(inp.text.Get()))
	inp.blink.Set(true)
	inp.ensureCursorVisible()
}

// submit calls the onSubmit callback.
func (inp *Input) submit(ke KeyEvent) {
	if inp.onSubmit != nil {
		inp.onSubmit(inp.text.Get())
	}
}

// --- Display ---

// displayText returns a viewport-clamped slice of the text with cursor overlay.
func (inp *Input) displayText() string {
	text := inp.text.Get()
	runes := []rune(text)
	pos := inp.clampCursorPos()
	visible := inp.visibleWidth()

	inp.ensureCursorVisible()
	scroll := inp.scrollPos.Get()

	// Clamp scroll to valid range
	if scroll < 0 {
		scroll = 0
	}
	if scroll > len(runes) {
		scroll = len(runes)
	}

	if !inp.focused.Get() {
		if len(runes) == 0 {
			return " "
		}
		// Show viewport slice
		end := scroll + visible
		if end > len(runes) {
			end = len(runes)
		}
		return string(runes[scroll:end])
	}

	// Build the visible slice with cursor inserted
	cursor := inp.cursorRune
	if !inp.blink.Get() {
		cursor = ' '
	}

	// Insert cursor into the full rune slice at pos
	withCursor := make([]rune, 0, len(runes)+1)
	withCursor = append(withCursor, runes[:pos]...)
	withCursor = append(withCursor, cursor)
	withCursor = append(withCursor, runes[pos:]...)

	// The cursor insertion shifts indices after pos by 1.
	// Adjust scroll start: characters before cursor are unshifted,
	// characters at/after cursor position are shifted by 1.
	viewStart := scroll
	if scroll > pos {
		viewStart = scroll + 1
	}

	// visible+1 because the cursor character takes a column
	viewEnd := viewStart + visible + 1
	if viewEnd > len(withCursor) {
		viewEnd = len(withCursor)
	}

	return string(withCursor[viewStart:viewEnd])
}

func (inp *Input) clampCursorPos() int {
	pos := inp.cursorPos.Get()
	if pos < 0 {
		return 0
	}
	max := utf8.RuneCountInString(inp.text.Get())
	if pos > max {
		return max
	}
	return pos
}
