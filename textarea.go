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
	width             int
	maxHeight         int
	border            BorderStyle
	textStyle         Style
	placeholder       string
	placeholderStyle  Style
	cursorRune        rune
	hideVirtualCursor bool
	focusColor        *Color
	borderGradient    *Gradient
	focusGradient     *Gradient
	autoFocus         bool
	submitKey         Key
	onChange          func(string)
	onSubmit          func(string)

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
	_ AppBinder       = (*TextArea)(nil)
)

// BindApp binds this TextArea's internal States to the given app.
func (t *TextArea) BindApp(app *App) {
	t.text.BindApp(app)
	t.cursorPos.BindApp(app)
	t.blink.BindApp(app)
	t.focused.BindApp(app)
}

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
		// Real terminal cursor is the default; the drawn glyph is opt-in via
		// WithTextAreaVirtualCursor.
		hideVirtualCursor: true,
		submitKey:         KeyEnter,

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
	if t.onChange != nil {
		t.onChange(t.text.Get())
	}
}

// Clear clears the text area.
func (t *TextArea) Clear() {
	t.text.Set("")
	t.cursorPos.Set(0)
	if t.onChange != nil {
		t.onChange(t.text.Get())
	}
}

// CursorPos returns the cursor position as a grapheme-cluster index: the count
// of whole glyphs before the cursor. This is the unit SetCursorPos accepts; it
// is not a byte or rune offset, so do not use it to slice the text directly.
func (t *TextArea) CursorPos() int {
	return runeIndexToClusterIndex(t.text.Get(), t.clampCursorPos())
}

// SetCursorPos moves the cursor to the given grapheme-cluster index. The index
// is clamped to [0, ClusterCount(text)] and always lands on a cluster boundary.
func (t *TextArea) SetCursorPos(pos int) {
	text := t.text.Get()
	if pos < 0 {
		pos = 0
	}
	t.cursorPos.Set(clusterIndexToRuneIndex(text, pos))
	t.blink.Set(true)
}

// InsertText inserts s at the cursor and advances the cursor past it. Insertion
// routes through the same internal path that typing uses, so the bound text and
// the cursor stay cluster-consistent (combining marks join the preceding base,
// and the cursor lands after the final whole cluster).
func (t *TextArea) InsertText(s string) {
	t.insertString(s)
}

// insertString splices s into the text at the cursor's rune index and advances
// the cursor to the end of the inserted content, snapped to a cluster boundary
// via clusterEnd. This is the single insert path shared by insertChar,
// insertNewline, and InsertText.
func (t *TextArea) insertString(s string) {
	if s == "" {
		return
	}
	text := t.text.Get()
	pos := t.clampCursorPos()
	runes := []rune(text)
	insert := []rune(s)
	newRunes := make([]rune, 0, len(runes)+len(insert))
	newRunes = append(newRunes, runes[:pos]...)
	newRunes = append(newRunes, insert...)
	newRunes = append(newRunes, runes[pos:]...)
	newText := string(newRunes)
	t.text.Set(newText)
	// Advance to the end of the cluster that contains the last inserted rune so
	// a trailing combining mark glues to its base and the cursor lands after it.
	t.cursorPos.Set(clusterEnd(newText, pos+len(insert)-1))
	t.blink.Set(true)
	if t.onChange != nil {
		t.onChange(t.text.Get())
	}
}

// contentRows returns the number of content rows to render: the wrapped
// lines plus the phantom cursor row, clamped to maxHeight. Note that when
// content exceeds maxHeight the rows below the clamp (including the cursor's
// row) are clipped; the textarea has no scroll-to-cursor.
func (t *TextArea) contentRows(lines []string) int {
	rows := len(lines)
	if t.phantomCursorRow(lines) {
		rows++
	}
	rows = max(rows, 1)
	if t.maxHeight > 0 && rows > t.maxHeight {
		rows = t.maxHeight
	}
	return rows
}

// Height returns the total rendered height including border.
func (t *TextArea) Height() int {
	height := t.contentRows(t.wrapText())
	if t.border != BorderNone {
		height += 2
	}
	return height
}

// --- Component Interface ---

// Render returns the element tree for the text area.
func (t *TextArea) Render(app *App) *Element {
	lines := t.wrapText()
	rows := t.contentRows(lines)

	// Account for border
	totalHeight := rows
	if t.border != BorderNone {
		totalHeight += 2
	}

	opts := []Option{
		WithDirection(Column),
		WithHeight(totalHeight),
		WithFocusable(true),
		WithAutoFocus(t.autoFocus),
	}
	if t.width > 0 {
		opts = append(opts, WithWidth(t.width))
	}
	if t.border != BorderNone {
		opts = append(opts, WithBorder(t.border))
		if t.focused.Get() {
			if t.focusGradient != nil {
				opts = append(opts, WithBorderGradient(*t.focusGradient))
			} else if t.focusColor != nil {
				opts = append(opts, WithBorderStyle(NewStyle().Foreground(*t.focusColor)))
			}
		} else if t.borderGradient != nil {
			opts = append(opts, WithBorderGradient(*t.borderGradient))
		}
	}
	root := New(opts...)

	// Wire Element focus/blur to component focus/blur
	root.SetOnFocus(func(e *Element) {
		t.Focus()
	})
	root.SetOnBlur(func(e *Element) {
		t.Blur()
	})

	// Drive the real terminal cursor unless the drawn glyph is opted in.
	if t.hideVirtualCursor {
		root.SetCursorSource(t.cursorColRow)
	}

	// Render placeholder or content
	if t.text.Get() == "" && t.placeholder != "" && !t.focused.Get() {
		root.AddChild(New(WithText(t.placeholder), WithTextStyle(t.placeholderStyle)))
	} else {
		for i := range rows {
			root.AddChild(New(WithText(t.lineWithCursor(i)), WithTextStyle(t.textStyle), WithWrap(false)))
		}
	}

	return root
}

// --- Focusable Interface ---

// IsFocusable returns true since TextArea can receive focus.
func (t *TextArea) IsFocusable() bool {
	return true
}

// IsTabStop returns true since TextArea participates in Tab navigation.
func (t *TextArea) IsTabStop() bool {
	return true
}

// Focus is called when the text area gains focus. Idempotent.
func (t *TextArea) Focus() {
	if t.focused.Get() {
		return
	}
	t.focused.Set(true)
	t.blink.Set(true)
}

// Blur is called when the text area loses focus. Idempotent.
func (t *TextArea) Blur() {
	if !t.focused.Get() {
		return
	}
	t.focused.Set(false)
}

// IsFocused returns whether this text area is currently focused.
func (t *TextArea) IsFocused() bool {
	return t.focused.Get()
}

// HandleEvent processes keyboard events.
func (t *TextArea) HandleEvent(e Event) bool {
	ke, ok := e.(KeyEvent)
	if !ok {
		return false
	}

	for _, binding := range t.KeyMap() {
		entry := dispatchEntry{pattern: binding.Pattern}
		if entry.matchesKey(ke) {
			binding.Handler(ke)
			return binding.Stop
		}
	}
	return false
}

// --- KeyListener Interface ---

// KeyMap returns the key bindings for the text area.
func (t *TextArea) KeyMap() KeyMap {
	km := KeyMap{
		OnFocused(AnyRune, t.insertChar),
		OnFocused(KeyBackspace, t.backspace),
		OnFocused(KeyDelete, t.delete),
		OnFocused(KeyLeft, t.moveLeft),
		OnFocused(KeyRight, t.moveRight),
		OnFocused(KeyUp, t.moveUp),
		OnFocused(KeyDown, t.moveDown),
		OnFocused(KeyHome, t.moveHome),
		OnFocused(KeyEnd, t.moveEnd),
	}

	if t.submitKey == KeyEnter {
		km = append(
			km,
			OnFocused(Rune('j').Ctrl(), t.insertNewline),
			OnFocused(KeyEnter, t.submit),
		)
	} else {
		km = append(
			km,
			OnFocused(KeyEnter, t.insertNewline),
			OnFocused(t.submitKey, t.submit),
		)
	}

	km = append(
		km,
		OnFocused(KeyEscape, func(ke KeyEvent) {
			if app := ke.App(); app != nil {
				app.BlurFocused()
			}
		}),
	)

	return km
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
// Cursor advances past the cluster that contains the inserted character,
// so combining marks join the previous base and the cursor lands after
// the combined cluster.
func (t *TextArea) insertChar(ke KeyEvent) {
	t.insertString(string(ke.Rune))
}

// insertNewline inserts a newline character at the cursor position.
func (t *TextArea) insertNewline(ke KeyEvent) {
	t.insertString("\n")
}

// backspace deletes the cluster before the cursor.
func (t *TextArea) backspace(ke KeyEvent) {
	text := t.text.Get()
	pos := t.clampCursorPos()
	if pos > 0 {
		snapped := snapRuneToClusterStart(text, pos-1)
		runes := []rune(text)
		newRunes := append(runes[:snapped], runes[pos:]...)
		t.text.Set(string(newRunes))
		t.cursorPos.Set(snapped)
		if t.onChange != nil {
			t.onChange(t.text.Get())
		}
	}
}

// delete deletes the cluster at the cursor.
func (t *TextArea) delete(ke KeyEvent) {
	text := t.text.Get()
	runes := []rune(text)
	pos := t.clampCursorPos()
	if pos < len(runes) {
		end := clusterEnd(text, pos)
		newRunes := append(runes[:pos], runes[end:]...)
		t.text.Set(string(newRunes))
		if t.onChange != nil {
			t.onChange(t.text.Get())
		}
	}
}

// moveLeft moves cursor left by one grapheme cluster.
func (t *TextArea) moveLeft(ke KeyEvent) {
	text := t.text.Get()
	pos := t.cursorPos.Get()
	if pos > 0 {
		t.cursorPos.Set(snapRuneToClusterStart(text, pos-1))
		t.blink.Set(true)
	}
}

// moveRight moves cursor right by one grapheme cluster.
func (t *TextArea) moveRight(ke KeyEvent) {
	text := t.text.Get()
	pos := t.cursorPos.Get()
	if pos < utf8.RuneCountInString(text) {
		t.cursorPos.Set(clusterEnd(text, pos))
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
	// The cursor can sit on the phantom row one past the last line; End there
	// resolves against the last real line.
	if row >= len(lines) {
		row = len(lines) - 1
	}
	t.cursorPos.Set(t.posFromRowCol(lines, row, utf8.RuneCountInString(lines[row])))
	t.blink.Set(true)
}

// submit calls the onSubmit callback. Does not fire onChange
// because submit does not mutate the text.
func (t *TextArea) submit(ke KeyEvent) {
	if t.onSubmit != nil {
		t.onSubmit(t.text.Get())
	}
}

// --- Text Wrapping and Cursor Position ---

// wrapWidth returns the display columns available for text content. Borders
// are drawn inside the element width, so they reduce the wrap width by one
// column on each side.
func (t *TextArea) wrapWidth() int {
	w := t.width
	if t.border != BorderNone {
		w -= 2
	}
	return w
}

// wrapText wraps the text to fit within the content width, respecting
// embedded newlines. Lines break at display-column boundaries, never
// mid-grapheme-cluster: a wide cluster that does not fit moves to the next
// line. Cursor math stays in rune indices, which remain consistent because
// wrapping only changes where lines split, not their runes.
func (t *TextArea) wrapText() []string {
	text := t.text.Get()
	if text == "" {
		return []string{""}
	}

	var lines []string
	width := t.wrapWidth()

	// Split on embedded newlines first
	paragraphs := strings.SplitSeq(text, "\n")

	for para := range paragraphs {
		if para == "" {
			lines = append(lines, "")
			continue
		}

		// Wrap this paragraph to width
		var currentLine strings.Builder
		currentWidth := 0
		rest := para
		for len(rest) > 0 {
			cluster, cw, size := nextCluster(rest)
			if size == 0 {
				break
			}
			rest = rest[size:]
			// The len guard keeps a cluster wider than the wrap width on a line
			// of its own instead of emitting empty lines before it.
			if width > 0 && currentLine.Len() > 0 && currentWidth+cw > width {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
				currentWidth = 0
			}
			currentLine.WriteString(cluster)
			currentWidth += cw
		}
		lines = append(lines, currentLine.String())
	}

	return lines
}

// cursorRowCol returns the row and column of the cursor.
//
// A cursor at the end of a display-full line has no column left to render in,
// so it moves to the start of the next visual line (downstream affinity). This
// applies at soft wrap boundaries and at end of text, where the reported row
// is one past the last wrapped line (a phantom row that Render and Height
// account for). A full line ended by a hard newline keeps the cursor at its
// end, since the next row starts a different paragraph.
func (t *TextArea) cursorRowCol(lines []string) (row, col int) {
	text := t.text.Get()
	pos := t.clampCursorPos()

	currentRow := 0
	currentRuneCol := 0    // rune index within the current line (returned as col)
	currentDisplayCol := 0 // display column for wrap boundary detection
	lineIdx := 0
	width := t.wrapWidth()
	wrapping := width > 0

	runePos := 0
	rest := text
	for len(rest) > 0 && runePos < pos {
		cluster, cw, size, rc := NextClusterRunes(rest)
		if size == 0 {
			break
		}

		if cluster == "\n" {
			currentRow++
			currentRuneCol = 0
			currentDisplayCol = 0
			lineIdx++
			runePos += rc
			rest = rest[size:]
			continue
		}

		// Check if this cluster crosses the wrap boundary.
		if wrapping && lineIdx < len(lines) && currentDisplayCol > 0 && currentDisplayCol+cw > stringWidth(lines[lineIdx]) {
			currentRow++
			currentRuneCol = 0
			currentDisplayCol = 0
			lineIdx++
		}

		currentRuneCol += rc    // rune index for callers
		currentDisplayCol += cw // display column for wrap detection
		runePos += rc
		rest = rest[size:]
	}

	// Check for the phantom cursor row: cursor at end of a display-full line.
	if wrapping && lineIdx < len(lines) && currentRuneCol > 0 &&
		currentDisplayCol == stringWidth(lines[lineIdx]) && currentDisplayCol >= width &&
		(pos == utf8.RuneCountInString(text) || (len(rest) > 0 && rest[0] != '\n')) {
		return currentRow + 1, 0
	}

	return currentRow, currentRuneCol
}

// posFromRowCol converts row/col back to absolute rune position.
// The col parameter is a rune index within the target row.
func (t *TextArea) posFromRowCol(lines []string, targetRow, targetCol int) int {
	text := t.text.Get()

	currentRow := 0
	currentCol := 0
	lineIdx := 0
	wrapping := t.wrapWidth() > 0

	// Walk clusters, tracking display columns for wrap detection,
	// but keeping currentCol as a rune index so targetCol (a rune index)
	// can be matched.
	displayCol := 0
	runePos := 0
	rest := text
	for len(rest) > 0 {
		if currentRow == targetRow && currentCol == targetCol {
			return runePos
		}

		cluster, cw, size, rc := NextClusterRunes(rest)
		if size == 0 {
			return runePos
		}

		if cluster == "\n" {
			if currentRow == targetRow {
				return runePos
			}
			currentRow++
			currentCol = 0
			displayCol = 0
			lineIdx++
			runePos += rc
			rest = rest[size:]
			continue
		}

		// Check if this cluster crosses the wrap boundary.
		if wrapping && lineIdx < len(lines) && displayCol > 0 && displayCol+cw > stringWidth(lines[lineIdx]) {
			if currentRow == targetRow {
				// The wrap boundary is the start of the next line.
				return runePos
			}
			currentRow++
			currentCol = 0
			displayCol = 0
			lineIdx++
			// Handle targetCol == 0 on a soft-wrapped line.
			if currentRow == targetRow && targetCol == 0 {
				return runePos
			}
		}

		currentCol += rc
		displayCol += cw
		runePos += rc
		rest = rest[size:]
	}

	return runePos
}

// phantomCursorRow reports whether the cursor sits one row past the last
// wrapped line (end of text on a display-full line), which needs an extra
// rendered row to host the cursor.
func (t *TextArea) phantomCursorRow(lines []string) bool {
	if !t.focused.Get() {
		return false
	}
	row, _ := t.cursorRowCol(lines)
	return row >= len(lines)
}

// cursorColRow returns the cursor's content-local position (display column, row)
// within the textarea's content area and whether it is visible. It shares the
// row/column computation (cursorRowCol) with the drawn-glyph path in
// lineWithCursor; the display column is the width of the line prefix up to the
// cursor's rune column, so wide clusters before the cursor advance it correctly.
// The textarea has no internal horizontal scroll; vertical scroll, if any, is
// owned by an enclosing scrollable element and applied by Element.ReportCursor.
func (t *TextArea) cursorColRow() (col, row int, visible bool) {
	if !t.focused.Get() {
		return 0, 0, false
	}
	lines := t.wrapText()
	r, runeCol := t.cursorRowCol(lines)

	// On the phantom row past the last wrapped line the cursor sits at column 0.
	if r >= len(lines) {
		return 0, r, true
	}

	// Convert the rune column within the line to a display column.
	line := lines[r]
	runes := []rune(line)
	if runeCol > len(runes) {
		runeCol = len(runes)
	}
	col = stringWidth(string(runes[:runeCol]))
	return col, r, true
}

// lineWithCursor returns a line with the cursor character inserted.
func (t *TextArea) lineWithCursor(lineIdx int) string {
	lines := t.wrapText()
	row, col := t.cursorRowCol(lines)

	if lineIdx >= len(lines) {
		// Phantom row hosting the cursor past a display-full last line.
		if lineIdx == row && t.focused.Get() && !t.hideVirtualCursor && t.blink.Get() {
			return string(t.cursorRune)
		}
		return " "
	}

	line := lines[lineIdx]

	if lineIdx == row && t.focused.Get() {
		// Skip virtual cursor when hardware cursor mode is enabled
		if t.hideVirtualCursor {
			if line == "" {
				return " "
			}
			return line
		}
		cursor := string(t.cursorRune)
		if !t.blink.Get() {
			cursor = " "
		}
		runes := []rune(line)
		if col >= len(runes) {
			// A display-full line ended by a hard newline keeps the cursor at
			// its end (see cursorRowCol), where an appended cursor would be
			// clipped. Overlay the last cell instead, like a block cursor
			// sitting on the character.
			if w := t.wrapWidth(); w > 0 && stringWidth(line) >= w && len(runes) > 0 {
				if !t.blink.Get() {
					return line
				}
				return string(runes[:snapRuneToClusterStart(line, len(runes)-1)]) + string(t.cursorRune)
			}
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
	text := t.text.Get()
	max := utf8.RuneCountInString(text)
	if pos > max {
		return max
	}
	// Snap to cluster start so cursor cannot sit inside a cluster.
	return snapRuneToClusterStart(text, pos)
}
