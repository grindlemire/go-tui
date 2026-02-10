package tui

import (
	"strings"
	"time"
	"unicode/utf8"
)

// TextArea is a multi-line text input with word wrapping and cursor management.
// It implements Component, KeyListener, WatcherProvider, and Focusable interfaces.
type TextArea struct {
	// Configuration (set via options, immutable after construction)
	width            int
	maxHeight        int
	border           BorderStyle
	textStyle        Style
	placeholder      string
	placeholderStyle Style
	cursorRune       rune
	submitKey        Key
	onSubmit         func(string)

	// Reactive state
	text      *State[string]
	cursorPos *State[int]
	blink     *State[bool]
	focused   *State[bool]
}

// Interface assertions
var (
	_ Component       = (*TextArea)(nil)
	_ KeyListener     = (*TextArea)(nil)
	_ WatcherProvider = (*TextArea)(nil)
	_ Focusable       = (*TextArea)(nil)
)

// NewTextArea creates a new multi-line text input.
func NewTextArea(opts ...TextAreaOption) *TextArea {
	t := &TextArea{
		// Defaults
		width:            40,
		maxHeight:        0, // unlimited
		border:           BorderNone,
		textStyle:        Style{},
		placeholder:      "",
		placeholderStyle: Style{}.Dim(),
		cursorRune:       '▌',
		submitKey:        KeyEnter,

		// State
		text:      NewState(""),
		cursorPos: NewState(0),
		blink:     NewState(true),
		focused:   NewState(false),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// --- State Access ---

// Text returns the current text content.
func (t *TextArea) Text() string {
	return t.text.Get()
}

// SetText sets the text and moves cursor to end.
func (t *TextArea) SetText(s string) {
	t.text.Set(s)
	t.cursorPos.Set(utf8.RuneCountInString(s))
}

// Clear clears the text area.
func (t *TextArea) Clear() {
	t.text.Set("")
	t.cursorPos.Set(0)
}

// Height returns the total rendered height including border.
func (t *TextArea) Height() int {
	lines := t.wrapText()
	height := len(lines)
	if height < 1 {
		height = 1
	}
	if t.maxHeight > 0 && height > t.maxHeight {
		height = t.maxHeight
	}
	if t.border != BorderNone {
		height += 2
	}
	return height
}

// --- Component Interface ---

// Render returns the element tree for the text area.
func (t *TextArea) Render() *Element {
	lines := t.wrapText()
	height := len(lines)
	if height < 1 {
		height = 1
	}
	if t.maxHeight > 0 && height > t.maxHeight {
		height = t.maxHeight
	}

	// Account for border
	totalHeight := height
	if t.border != BorderNone {
		totalHeight += 2
	}

	opts := []Option{
		WithDirection(Column),
		WithHeight(totalHeight),
	}
	if t.border != BorderNone {
		opts = append(opts, WithBorder(t.border))
	}
	root := New(opts...)

	// Render placeholder or content
	if t.text.Get() == "" && t.placeholder != "" && !t.focused.Get() {
		root.AddChild(New(WithText(t.placeholder), WithTextStyle(t.placeholderStyle)))
	} else {
		for i := range lines {
			root.AddChild(New(WithText(t.lineWithCursor(i)), WithTextStyle(t.textStyle)))
		}
	}

	return root
}

// --- Focusable Interface ---

// IsFocusable returns true since TextArea can receive focus.
func (t *TextArea) IsFocusable() bool {
	return true
}

// Focus is called when the text area gains focus.
func (t *TextArea) Focus() {
	t.focused.Set(true)
	t.blink.Set(true)
}

// Blur is called when the text area loses focus.
func (t *TextArea) Blur() {
	t.focused.Set(false)
}

// HandleEvent processes keyboard events.
func (t *TextArea) HandleEvent(e Event) bool {
	ke, ok := e.(KeyEvent)
	if !ok {
		return false
	}

	// Check each binding in our keymap
	for _, binding := range t.KeyMap() {
		if t.matchesPattern(ke, binding.Pattern) {
			binding.Handler(ke)
			return binding.Stop
		}
	}
	return false
}

// matchesPattern checks if a key event matches a pattern.
func (t *TextArea) matchesPattern(ke KeyEvent, p KeyPattern) bool {
	// Check for specific key match
	if p.Key != 0 && ke.Key == p.Key {
		return true
	}
	// Check for specific rune match
	if p.Rune != 0 && ke.Rune == p.Rune {
		return true
	}
	// Check for any rune match
	if p.AnyRune && ke.Rune != 0 {
		return true
	}
	return false
}

// --- KeyListener Interface ---

// KeyMap returns the key bindings for the text area.
func (t *TextArea) KeyMap() KeyMap {
	// Determine which key inserts newline vs submits
	// Default: Enter submits, Ctrl+J inserts newline
	// Alternative (if submitKey is not KeyEnter): submitKey submits, Enter inserts newline
	var newlineKey, submitKeyBinding Key
	if t.submitKey == KeyEnter {
		submitKeyBinding = KeyEnter
		newlineKey = KeyCtrlJ
	} else {
		submitKeyBinding = t.submitKey
		newlineKey = KeyEnter
	}

	return KeyMap{
		// Text input
		OnRunesStop(t.insertChar),

		// Editing
		OnKeyStop(KeyBackspace, t.backspace),
		OnKeyStop(KeyDelete, t.delete),

		// Navigation
		OnKeyStop(KeyLeft, t.moveLeft),
		OnKeyStop(KeyRight, t.moveRight),
		OnKeyStop(KeyUp, t.moveUp),
		OnKeyStop(KeyDown, t.moveDown),
		OnKeyStop(KeyHome, t.moveHome),
		OnKeyStop(KeyEnd, t.moveEnd),

		// Newline and submit
		OnKeyStop(newlineKey, t.insertNewline),
		OnKeyStop(submitKeyBinding, t.submit),
	}
}

// --- WatcherProvider Interface ---

// Watchers returns watchers for cursor blink.
func (t *TextArea) Watchers() []Watcher {
	return []Watcher{
		OnTimer(500*time.Millisecond, func() {
			if t.focused.Get() {
				t.blink.Set(!t.blink.Get())
			}
		}),
	}
}

// --- Key Handlers ---

// insertChar inserts a character at the cursor position.
func (t *TextArea) insertChar(ke KeyEvent) {
	runes := []rune(t.text.Get())
	pos := t.clampCursorPos()
	newRunes := append(runes[:pos], append([]rune{ke.Rune}, runes[pos:]...)...)
	t.text.Set(string(newRunes))
	t.cursorPos.Set(pos + 1)
	t.blink.Set(true)
}

// insertNewline inserts a newline character at the cursor position.
func (t *TextArea) insertNewline(ke KeyEvent) {
	runes := []rune(t.text.Get())
	pos := t.clampCursorPos()
	newRunes := append(runes[:pos], append([]rune{'\n'}, runes[pos:]...)...)
	t.text.Set(string(newRunes))
	t.cursorPos.Set(pos + 1)
	t.blink.Set(true)
}

// backspace deletes the character before the cursor.
func (t *TextArea) backspace(ke KeyEvent) {
	runes := []rune(t.text.Get())
	pos := t.clampCursorPos()
	if pos > 0 {
		newRunes := append(runes[:pos-1], runes[pos:]...)
		t.text.Set(string(newRunes))
		t.cursorPos.Set(pos - 1)
	}
}

// delete deletes the character at the cursor.
func (t *TextArea) delete(ke KeyEvent) {
	runes := []rune(t.text.Get())
	pos := t.clampCursorPos()
	if pos < len(runes) {
		newRunes := append(runes[:pos], runes[pos+1:]...)
		t.text.Set(string(newRunes))
	}
}

// moveLeft moves cursor left.
func (t *TextArea) moveLeft(ke KeyEvent) {
	pos := t.cursorPos.Get()
	if pos > 0 {
		t.cursorPos.Set(pos - 1)
		t.blink.Set(true)
	}
}

// moveRight moves cursor right.
func (t *TextArea) moveRight(ke KeyEvent) {
	pos := t.cursorPos.Get()
	if pos < utf8.RuneCountInString(t.text.Get()) {
		t.cursorPos.Set(pos + 1)
		t.blink.Set(true)
	}
}

// moveUp moves cursor up one line.
func (t *TextArea) moveUp(ke KeyEvent) {
	lines := t.wrapText()
	row, col := t.cursorRowCol(lines)
	if row > 0 {
		prevLine := lines[row-1]
		prevLen := utf8.RuneCountInString(prevLine)
		if col > prevLen {
			col = prevLen
		}
		t.cursorPos.Set(t.posFromRowCol(lines, row-1, col))
		t.blink.Set(true)
	}
}

// moveDown moves cursor down one line.
func (t *TextArea) moveDown(ke KeyEvent) {
	lines := t.wrapText()
	row, col := t.cursorRowCol(lines)
	if row < len(lines)-1 {
		nextLine := lines[row+1]
		nextLen := utf8.RuneCountInString(nextLine)
		if col > nextLen {
			col = nextLen
		}
		t.cursorPos.Set(t.posFromRowCol(lines, row+1, col))
		t.blink.Set(true)
	}
}

// moveHome moves cursor to start of current line.
func (t *TextArea) moveHome(ke KeyEvent) {
	lines := t.wrapText()
	row, _ := t.cursorRowCol(lines)
	t.cursorPos.Set(t.posFromRowCol(lines, row, 0))
	t.blink.Set(true)
}

// moveEnd moves cursor to end of current line.
func (t *TextArea) moveEnd(ke KeyEvent) {
	lines := t.wrapText()
	row, _ := t.cursorRowCol(lines)
	t.cursorPos.Set(t.posFromRowCol(lines, row, utf8.RuneCountInString(lines[row])))
	t.blink.Set(true)
}

// submit calls the onSubmit callback.
func (t *TextArea) submit(ke KeyEvent) {
	if t.onSubmit != nil {
		t.onSubmit(t.text.Get())
	}
}

// --- Text Wrapping and Cursor Position ---

// wrapText wraps the text to fit within width, respecting embedded newlines.
func (t *TextArea) wrapText() []string {
	text := t.text.Get()
	if text == "" {
		return []string{""}
	}

	var lines []string

	// Split on embedded newlines first
	paragraphs := strings.Split(text, "\n")

	for _, para := range paragraphs {
		if para == "" {
			lines = append(lines, "")
			continue
		}

		// Wrap this paragraph to width
		currentLine := make([]rune, 0)
		for _, r := range para {
			if t.width > 0 && len(currentLine) >= t.width {
				lines = append(lines, string(currentLine))
				currentLine = currentLine[:0]
			}
			currentLine = append(currentLine, r)
		}
		lines = append(lines, string(currentLine))
	}

	return lines
}

// cursorRowCol returns the row and column of the cursor.
func (t *TextArea) cursorRowCol(lines []string) (row, col int) {
	text := t.text.Get()
	pos := t.clampCursorPos()
	textRunes := []rune(text)

	currentRow := 0
	currentCol := 0
	lineIdx := 0

	for i := 0; i < len(textRunes) && i < pos; i++ {
		if textRunes[i] == '\n' {
			currentRow++
			currentCol = 0
			lineIdx++
		} else {
			currentCol++
			if t.width > 0 && lineIdx < len(lines) && currentCol > utf8.RuneCountInString(lines[lineIdx]) {
				currentRow++
				currentCol = 1
				lineIdx++
			}
		}
	}

	return currentRow, currentCol
}

// posFromRowCol converts row/col back to absolute position.
func (t *TextArea) posFromRowCol(lines []string, targetRow, targetCol int) int {
	text := t.text.Get()
	textRunes := []rune(text)

	currentRow := 0
	currentCol := 0
	lineIdx := 0

	for i := 0; i < len(textRunes); i++ {
		if currentRow == targetRow && currentCol == targetCol {
			return i
		}

		if textRunes[i] == '\n' {
			if currentRow == targetRow {
				return i
			}
			currentRow++
			currentCol = 0
			lineIdx++
		} else {
			currentCol++
			if t.width > 0 && lineIdx < len(lines) && currentCol > utf8.RuneCountInString(lines[lineIdx]) {
				if currentRow == targetRow {
					return i
				}
				currentRow++
				currentCol = 1
				lineIdx++
			}
		}
	}

	return len(textRunes)
}

// lineWithCursor returns a line with the cursor character inserted.
func (t *TextArea) lineWithCursor(lineIdx int) string {
	lines := t.wrapText()
	if lineIdx >= len(lines) {
		return " "
	}

	row, col := t.cursorRowCol(lines)
	line := lines[lineIdx]

	if lineIdx == row && t.focused.Get() {
		cursor := string(t.cursorRune)
		if !t.blink.Get() {
			cursor = " "
		}
		runes := []rune(line)
		if col >= len(runes) {
			return line + cursor
		}
		withCursor := append(runes[:col], append([]rune{t.cursorRune}, runes[col:]...)...)
		if !t.blink.Get() {
			withCursor[col] = ' '
		}
		return string(withCursor)
	}

	if line == "" {
		return " "
	}
	return line
}

func (t *TextArea) clampCursorPos() int {
	pos := t.cursorPos.Get()
	if pos < 0 {
		return 0
	}
	max := utf8.RuneCountInString(t.text.Get())
	if pos > max {
		return max
	}
	return pos
}
