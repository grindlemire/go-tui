package tui

import (
	"strings"
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
	focusColor       *Color
	borderGradient   *Gradient
	focusGradient    *Gradient
	autoFocus        bool
	onSubmit         func(string)
	onChange         func(string)

	// Reactive state
	text      *State[string]
	cursorPos *State[int]
	scrollPos *State[int] // horizontal scroll offset (first visible display column, snapped to a cluster boundary)
	blink     *State[bool]
	focused   *State[bool]
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

// ensureCursorVisible adjusts scrollPos (a display column, always snapped to a
// cluster boundary) so the cursor column is within the visible window.
func (inp *Input) ensureCursorVisible() {
	text := inp.text.Get()
	pos := inp.clampCursorPos()
	cursorCol := runeIndexToDisplayCol(text, pos)
	scroll := inp.scrollPos.Get()
	visible := inp.visibleWidth()
	if visible <= 0 {
		return
	}

	// Cursor is left of the visible window: scroll so the cursor's column is the
	// first visible column (already a cluster boundary).
	if cursorCol < scroll {
		inp.scrollPos.Set(cursorCol)
		return
	}

	// Cursor is right of the visible window. Reserve 1 column for the cursor
	// character itself, then snap the new scroll column down to a cluster start.
	if cursorCol >= scroll+visible {
		want := cursorCol - visible + 1
		inp.scrollPos.Set(inp.snapColToClusterStart(want))
	}
}

// snapColToClusterStart snaps a display column down to the nearest cluster-start
// column at or before it.
func (inp *Input) snapColToClusterStart(col int) int {
	if col <= 0 {
		return 0
	}
	text := inp.text.Get()
	cur := 0
	for len(text) > 0 {
		_, w, size := nextCluster(text)
		if size == 0 {
			break
		}
		if cur+w > col {
			return cur
		}
		cur += w
		text = text[size:]
		if cur == col {
			return cur
		}
	}
	return cur
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
		WithAutoFocus(inp.autoFocus),
	}
	if inp.width > 0 {
		opts = append(opts, WithWidth(inp.width))
	}
	if inp.border != BorderNone {
		opts = append(opts, WithBorder(inp.border))
		if inp.focused.Get() {
			if inp.focusGradient != nil {
				opts = append(opts, WithBorderGradient(*inp.focusGradient))
			} else if inp.focusColor != nil {
				opts = append(opts, WithBorderStyle(NewStyle().Foreground(*inp.focusColor)))
			}
		} else if inp.borderGradient != nil {
			opts = append(opts, WithBorderGradient(*inp.borderGradient))
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
		entry := dispatchEntry{pattern: binding.Pattern}
		if entry.matchesKey(ke) {
			binding.Handler(ke)
			return binding.Stop
		}
	}
	return false
}

// --- KeyListener Interface ---

// KeyMap returns the key bindings for the input.
func (inp *Input) KeyMap() KeyMap {
	return KeyMap{
		OnFocused(AnyRune, inp.insertChar),
		OnFocused(KeyBackspace, inp.backspace),
		OnFocused(KeyDelete, inp.delete),
		OnFocused(KeyLeft, inp.moveLeft),
		OnFocused(KeyRight, inp.moveRight),
		OnFocused(KeyHome, inp.moveHome),
		OnFocused(KeyEnd, inp.moveEnd),
		OnFocused(KeyEnter, inp.submit),
		OnFocused(KeyEscape, func(ke KeyEvent) {
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

// insertChar inserts a character at the cursor position. After inserting, the
// cursor lands on the cluster boundary immediately following the resulting
// cluster: a base char advances one cluster; a combining mark typed after a base
// merges into the preceding cluster, and the cursor lands after that combined
// cluster.
func (inp *Input) insertChar(ke KeyEvent) {
	runes := []rune(inp.text.Get())
	pos := inp.clampCursorPos()
	newRunes := make([]rune, 0, len(runes)+1)
	newRunes = append(newRunes, runes[:pos]...)
	newRunes = append(newRunes, ke.Rune)
	newRunes = append(newRunes, runes[pos:]...)
	newText := string(newRunes)
	inp.text.Set(newText)

	// Re-segment from the previous cluster boundary at/under the old position and
	// move to the end of the cluster that now contains the inserted rune.
	inp.cursorPos.Set(clusterEndAfterInsert(newText, pos))
	inp.blink.Set(true)
	inp.ensureCursorVisible()
	if inp.onChange != nil {
		inp.onChange(inp.text.Get())
	}
}

// backspace deletes the cluster before the cursor.
func (inp *Input) backspace(ke KeyEvent) {
	runes := []rune(inp.text.Get())
	pos := inp.clampCursorPos()
	if pos > 0 {
		prev := inp.prevClusterBoundary(pos)
		newRunes := append(runes[:prev], runes[pos:]...)
		inp.text.Set(string(newRunes))
		inp.cursorPos.Set(prev)
		inp.ensureCursorVisible()
		if inp.onChange != nil {
			inp.onChange(inp.text.Get())
		}
	}
}

// delete deletes the cluster at the cursor.
func (inp *Input) delete(ke KeyEvent) {
	runes := []rune(inp.text.Get())
	pos := inp.clampCursorPos()
	if pos < len(runes) {
		next := inp.nextClusterBoundary(pos)
		newRunes := append(runes[:pos], runes[next:]...)
		inp.text.Set(string(newRunes))
		inp.ensureCursorVisible()
		if inp.onChange != nil {
			inp.onChange(inp.text.Get())
		}
	}
}

// moveLeft moves cursor to the previous cluster boundary.
func (inp *Input) moveLeft(ke KeyEvent) {
	pos := inp.clampCursorPos()
	if pos > 0 {
		inp.cursorPos.Set(inp.prevClusterBoundary(pos))
		inp.blink.Set(true)
		inp.ensureCursorVisible()
	}
}

// moveRight moves cursor to the next cluster boundary.
func (inp *Input) moveRight(ke KeyEvent) {
	pos := inp.clampCursorPos()
	if pos < utf8.RuneCountInString(inp.text.Get()) {
		inp.cursorPos.Set(inp.nextClusterBoundary(pos))
		inp.blink.Set(true)
		inp.ensureCursorVisible()
	}
}

// prevClusterBoundary returns the rune index of the cluster boundary immediately
// before the given (cluster-aligned) rune position.
func (inp *Input) prevClusterBoundary(pos int) int {
	starts := clusterRuneStarts(inp.text.Get())
	prev := 0
	for _, st := range starts {
		if st >= pos {
			break
		}
		prev = st
	}
	return prev
}

// nextClusterBoundary returns the rune index of the cluster boundary immediately
// after the given (cluster-aligned) rune position.
func (inp *Input) nextClusterBoundary(pos int) int {
	starts := clusterRuneStarts(inp.text.Get())
	for _, st := range starts {
		if st > pos {
			return st
		}
	}
	return starts[len(starts)-1]
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

// displayText returns a viewport-clamped slice of the text (by display column)
// with the cursor glyph overlaid at the cursor's display column. It walks
// clusters, emits those inside the column window, and pads a single space for a
// wide cluster whose left half is scrolled off so columns stay aligned.
func (inp *Input) displayText() string {
	text := inp.text.Get()
	pos := inp.clampCursorPos()
	visible := inp.visibleWidth()
	focused := inp.focused.Get()

	inp.ensureCursorVisible()
	scroll := max(inp.scrollPos.Get(), 0)

	cursorCol := runeIndexToDisplayCol(text, pos)
	cursor := inp.cursorRune
	if !inp.blink.Get() {
		cursor = ' '
	}

	// Build the column-space stream of glyph units. When focused, a one-column
	// cursor unit is inserted at the cursor's column, which shifts later glyphs one
	// column to the right (with-cursor space).
	type unit struct {
		s string
		w int
	}
	var units []unit
	inserted := false
	col := 0 // running text column
	insertCursor := func() {
		if focused && !inserted && col == cursorCol {
			units = append(units, unit{string(cursor), 1})
			inserted = true
		}
	}
	rest := text
	for len(rest) > 0 {
		insertCursor()
		cluster, w, size := nextCluster(rest)
		if size == 0 {
			break
		}
		rest = rest[size:]
		units = append(units, unit{cluster, w})
		col += w
	}
	insertCursor() // cursor at end of text

	// The window is `visible` text columns wide, plus one column for the inserted
	// cursor when focused. scroll is a text column; shift it into with-cursor space
	// when it sits past the cursor (the cursor added a column before the window).
	windowWidth := visible
	wcScroll := scroll
	if focused {
		windowWidth = visible + 1
		if scroll > cursorCol {
			wcScroll = scroll + 1
		}
	}
	windowEnd := wcScroll + windowWidth

	var b strings.Builder
	wc := 0 // running with-cursor column
	for _, u := range units {
		start, end := wc, wc+u.w
		wc = end
		if end <= wcScroll {
			continue
		}
		if start >= windowEnd {
			break
		}
		if start < wcScroll {
			// Wide glyph straddling the left edge: pad its hidden half so columns
			// stay aligned.
			b.WriteRune(' ')
			continue
		}
		if end > windowEnd {
			// Wide glyph straddling the right edge: pad the visible columns instead
			// of writing the whole cluster, which would overflow the viewport and
			// then be clip-dropped (the glyph would silently vanish).
			for range windowEnd - start {
				b.WriteRune(' ')
			}
			break
		}
		b.WriteString(u.s)
	}
	if b.Len() == 0 {
		return " "
	}
	return b.String()
}

func (inp *Input) clampCursorPos() int {
	pos := inp.cursorPos.Get()
	text := inp.text.Get()
	if pos < 0 {
		return 0
	}
	max := utf8.RuneCountInString(text)
	if pos > max {
		return max
	}
	// Snap to a cluster boundary so the cursor never sits inside a cluster.
	return snapRuneToClusterStart(text, pos)
}
