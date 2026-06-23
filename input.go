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
	scrollPos *State[int] // horizontal scroll offset (first visible display column)
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
	inp.ensureCursorVisible()
}

// Clear clears the input.
func (inp *Input) Clear() {
	inp.text.Set("")
	inp.cursorPos.Set(0)
	inp.scrollPos.Set(0)
}

// CursorPos returns the cursor position as a grapheme-cluster index: the count
// of whole glyphs before the cursor. This is the unit SetCursorPos accepts; it
// is not a byte or rune offset, so do not use it to slice the text directly.
func (inp *Input) CursorPos() int {
	return runeIndexToClusterIndex(inp.text.Get(), inp.clampCursorPos())
}

// SetCursorPos moves the cursor to the given grapheme-cluster index. The index
// is clamped to [0, ClusterCount(text)] and always lands on a cluster boundary.
func (inp *Input) SetCursorPos(pos int) {
	text := inp.text.Get()
	if pos < 0 {
		pos = 0
	}
	inp.cursorPos.Set(clusterIndexToRuneIndex(text, pos))
	inp.blink.Set(true)
	inp.ensureCursorVisible()
}

// InsertText inserts s at the cursor and advances the cursor past it. Insertion
// routes through the same internal path that typing uses, so the bound text and
// the cursor stay cluster-consistent (combining marks join the preceding base,
// and the cursor lands after the final whole cluster).
func (inp *Input) InsertText(s string) {
	inp.insertString(s)
}

// insertString splices s into the text at the cursor's rune index and advances
// the cursor to the end of the inserted content, snapped to a cluster boundary
// via clusterEnd. This is the single insert path shared by insertChar and
// InsertText; it keeps scroll and onChange consistent.
func (inp *Input) insertString(s string) {
	if s == "" {
		return
	}
	text := inp.text.Get()
	pos := inp.clampCursorPos()
	runes := []rune(text)
	insert := []rune(s)
	newRunes := make([]rune, 0, len(runes)+len(insert))
	newRunes = append(newRunes, runes[:pos]...)
	newRunes = append(newRunes, insert...)
	newRunes = append(newRunes, runes[pos:]...)
	newText := string(newRunes)

	inp.text.Set(newText)
	// Advance to the end of the cluster that contains the last inserted rune so
	// a trailing combining mark glues to its base and the cursor lands after it.
	inp.cursorPos.Set(clusterEnd(newText, pos+len(insert)-1))
	inp.blink.Set(true)
	inp.ensureCursorVisible()
	if inp.onChange != nil {
		inp.onChange(inp.text.Get())
	}
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

// ensureCursorVisible adjusts scrollPos so the cursor is within the visible
// window. scrollPos stores a display column; the cursor's display column
// is computed from its rune index via runeIndexToDisplayCol.
func (inp *Input) ensureCursorVisible() {
	text := inp.text.Get()
	cursorCol := runeIndexToDisplayCol(text, inp.clampCursorPos())
	scroll := inp.scrollPos.Get()
	visible := inp.visibleWidth()
	if visible <= 0 {
		inp.scrollPos.Set(0)
		return
	}

	// Cursor is left of the visible window: scroll back so cursor is at col 0.
	if cursorCol < scroll {
		inp.scrollPos.Set(cursorCol)
		return
	}

	// Cursor is right of the visible window (reserve 1 column for the cursor
	// glyph itself): scroll forward so cursor is at the rightmost column.
	// Snap the new scroll to a cluster-start column so that viewportText
	// doesn't land in the middle of a wide cluster and push the cursor
	// out of the visible window.
	if cursorCol >= scroll+visible {
		want := cursorCol - visible + 1
		inp.scrollPos.Set(snapColToNextClusterBoundary(text, want))
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

// insertChar inserts a character at the cursor position.
// Cursor advances past the cluster that contains the inserted character,
// so combining marks join the previous base and the cursor lands after
// the combined cluster.
func (inp *Input) insertChar(ke KeyEvent) {
	inp.insertString(string(ke.Rune))
}

// backspace deletes the character before the cursor.
// The cursor snaps to the previous cluster boundary, so a multi-rune
// cluster is deleted as one unit.
func (inp *Input) backspace(ke KeyEvent) {
	text := inp.text.Get()
	pos := inp.clampCursorPos()
	if pos > 0 {
		// Snap to the start of the cluster that contains pos-1.
		snapped := snapRuneToClusterStart(text, pos-1)
		runes := []rune(text)
		newRunes := append(runes[:snapped], runes[pos:]...)
		newText := string(newRunes)
		inp.text.Set(newText)
		inp.cursorPos.Set(snapped)
		// Re-compute scroll so the cursor is visible.
		inp.ensureCursorVisible()
		if inp.onChange != nil {
			inp.onChange(inp.text.Get())
		}
	}
}

// delete deletes the character at the cursor.
// The cursor stays at the same cluster boundary, so a multi-rune cluster
// is deleted as one unit.
func (inp *Input) delete(ke KeyEvent) {
	text := inp.text.Get()
	runes := []rune(text)
	pos := inp.clampCursorPos()
	if pos < len(runes) {
		// Find the end of the cluster at pos.
		end := clusterEnd(text, pos)
		newRunes := append(runes[:pos], runes[end:]...)
		inp.text.Set(string(newRunes))
		// Re-compute scroll so the cursor is visible.
		inp.ensureCursorVisible()
		if inp.onChange != nil {
			inp.onChange(inp.text.Get())
		}
	}
}

// moveLeft moves cursor left by one grapheme cluster.
func (inp *Input) moveLeft(ke KeyEvent) {
	text := inp.text.Get()
	pos := inp.cursorPos.Get()
	if pos > 0 {
		inp.cursorPos.Set(snapRuneToClusterStart(text, pos-1))
		inp.blink.Set(true)
		inp.ensureCursorVisible()
	}
}

// moveRight moves cursor right by one grapheme cluster.
func (inp *Input) moveRight(ke KeyEvent) {
	text := inp.text.Get()
	pos := inp.cursorPos.Get()
	if pos < utf8.RuneCountInString(text) {
		inp.cursorPos.Set(clusterEnd(text, pos))
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

// snapColToNextClusterBoundary advances col to the next cluster boundary at or after
// col. For ASCII text every column is a boundary, so this is a no-op.
// For CJK/emoji where clusters are 2 columns wide, this ensures the scroll
// position never splits a cluster.
func snapColToNextClusterBoundary(s string, col int) int {
	if col <= 0 {
		return 0
	}
	cur := 0
	for len(s) > 0 {
		_, w, size := nextCluster(s)
		if size == 0 {
			break
		}
		s = s[size:]
		next := cur + w
		if next >= col {
			// col falls at or before the boundary after this cluster.
			// If col is already at cur (a boundary), return it.
			// If col is inside (cur < col < next), snap up to next.
			if col == cur {
				return col
			}
			return next
		}
		cur = next
	}
	return col
}

// --- Display ---

// displayCluster represents one grapheme cluster in the display stream.
type displayCluster struct {
	text  string
	width int
}

// textToClusters segments s into grapheme clusters and returns them along with
// the total display width.
func textToClusters(s string) (clusters []displayCluster, totalWidth int) {
	for len(s) > 0 {
		cl, w, size := nextCluster(s)
		if size == 0 {
			break
		}
		clusters = append(clusters, displayCluster{text: cl, width: w})
		totalWidth += w
		s = s[size:]
	}
	return
}

// displayText returns a viewport-clamped slice of the text with cursor overlay.
func (inp *Input) displayText() string {
	text := inp.text.Get()
	pos := inp.clampCursorPos()
	visible := inp.visibleWidth()

	// Build the cluster list from the full text.
	allClusters, _ := textToClusters(text)

	if !inp.focused.Get() {
		if len(allClusters) == 0 {
			return " "
		}
		// In the unfocused path, scroll adjustment preserves the previously-set
		// scroll position. ensureCursorVisible is only called in the focused path
		// so that manually-scrolled text doesn't jump when the user blurs.
		return inp.viewportText(allClusters, visible)
	}

	// Build a cluster list with the cursor inserted as an extra cluster.
	cursorR := inp.cursorRune
	if !inp.blink.Get() {
		cursorR = ' '
	}
	cursorCluster := displayCluster{text: string(cursorR), width: 1}

	// Insert the cursor cluster at the cursor's rune position within the
	// cluster list. The cursor is inserted before the cluster that contains
	// the rune at pos. For a combining mark inserted after a base (pos inside
	// a multi-rune cluster), the cursor goes after that cluster.
	insertIdx := len(allClusters)
	runeCount := 0
	for i, c := range allClusters {
		clusterRunes := utf8.RuneCountInString(c.text)
		if runeCount == pos {
			// Cursor is at the start of this cluster: insert before it.
			insertIdx = i
			break
		}
		if runeCount+clusterRunes > pos {
			// Cursor is inside this cluster (e.g. after a combining mark).
			// Insert after this cluster.
			insertIdx = i + 1
			break
		}
		runeCount += clusterRunes
	}

	withCursor := make([]displayCluster, 0, len(allClusters)+1)
	withCursor = append(withCursor, allClusters[:insertIdx]...)
	withCursor = append(withCursor, cursorCluster)
	withCursor = append(withCursor, allClusters[insertIdx:]...)

	inp.ensureCursorVisible()
	return inp.viewportText(withCursor, visible)
}

// viewportText returns the visible slice of clusters starting at scrollPos
// (display column), filling at most visible display columns. The cursor is
// expected to already be inserted into the cluster list.
// Callers must call ensureCursorVisible first (focused path) or decide
// whether scroll adjustment is desired (unfocused path skips it).
func (inp *Input) viewportText(clusters []displayCluster, visible int) string {
	scroll := max(inp.scrollPos.Get(), 0)
	if visible <= 0 {
		return ""
	}

	// Find the first cluster whose column range starts at or after scrollPos.
	// If scroll falls inside a cluster (between its start and end), that
	// cluster is skipped — we can't show half a cluster. The next cluster
	// whose start column is >= scroll becomes the first visible one.
	startIdx := len(clusters)
	col := 0
	for i, c := range clusters {
		if col >= scroll {
			startIdx = i
			break
		}
		col += c.width
		if col > scroll {
			// scroll fell inside this cluster: skip it.
			startIdx = i + 1
			break
		}
	}

	// Build the visible string up to visible columns.
	var sb strings.Builder
	col = 0
	for _, c := range clusters[startIdx:] {
		if col+c.width > visible {
			break
		}
		sb.WriteString(c.text)
		col += c.width
	}

	// Ensure we return at least one space for an empty viewport.
	if sb.Len() == 0 {
		sb.WriteRune(' ')
	}
	return sb.String()
}

func (inp *Input) clampCursorPos() int {
	pos := inp.cursorPos.Get()
	if pos < 0 {
		return 0
	}
	text := inp.text.Get()
	max := utf8.RuneCountInString(text)
	if pos > max {
		return max
	}
	// Snap to cluster start so cursor cannot sit inside a cluster.
	return snapRuneToClusterStart(text, pos)
}
