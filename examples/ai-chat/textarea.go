package main

import (
	"strings"
	"time"

	tui "github.com/grindlemire/go-tui"
	"github.com/grindlemire/go-tui/internal/debug"
)

// TextArea is a multi-line text input with wrapping and cursor management.
type TextArea struct {
	// Configuration
	width     int          // Available width for text (excluding border/padding)
	maxHeight int          // Max height (0 = unlimited)
	onSubmit  func(string) // Called when Enter is pressed

	// State
	text      *tui.State[string]
	cursorPos *tui.State[int] // Cursor position in text (0 = before first char)
	blink     *tui.State[bool]
}

// NewTextArea creates a new text area with the given width.
func NewTextArea(width int, onSubmit func(string)) *TextArea {
	return &TextArea{
		width:     width,
		maxHeight: 0,
		onSubmit:  onSubmit,
		text:      tui.NewState(""),
		cursorPos: tui.NewState(0),
		blink:     tui.NewState(true),
	}
}

// WithMaxHeight sets maximum height (0 = unlimited).
func (t *TextArea) WithMaxHeight(h int) *TextArea {
	t.maxHeight = h
	return t
}

// Text returns the current text content.
func (t *TextArea) Text() string {
	return t.text.Get()
}

// SetText sets the text and moves cursor to end.
func (t *TextArea) SetText(s string) {
	t.text.Set(s)
	t.cursorPos.Set(len(s))
}

// Clear clears the text area.
func (t *TextArea) Clear() {
	t.text.Set("")
	t.cursorPos.Set(0)
}

// KeyMap returns the key bindings for the text area.
func (t *TextArea) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRunesStop(t.insertChar),
		tui.OnKeyStop(tui.KeyBackspace, t.backspace),
		tui.OnKeyStop(tui.KeyDelete, t.delete),
		tui.OnKeyStop(tui.KeyLeft, t.moveLeft),
		tui.OnKeyStop(tui.KeyRight, t.moveRight),
		tui.OnKeyStop(tui.KeyUp, t.moveUp),
		tui.OnKeyStop(tui.KeyDown, t.moveDown),
		tui.OnKeyStop(tui.KeyHome, t.moveHome),
		tui.OnKeyStop(tui.KeyEnd, t.moveEnd),
		tui.OnKeyStop(tui.KeyEnter, t.submit),
		tui.OnKeyStop(tui.KeyCtrlJ, t.insertNewline),
	}
}

// Watchers returns watchers for cursor blink.
func (t *TextArea) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.OnTimer(500*time.Millisecond, func() {
			t.blink.Set(!t.blink.Get())
		}),
	}
}

// insertChar inserts a character at the cursor position.
func (t *TextArea) insertChar(ke tui.KeyEvent) {
	text := t.text.Get()
	pos := t.cursorPos.Get()
	newText := text[:pos] + string(ke.Rune) + text[pos:]
	t.text.Set(newText)
	t.cursorPos.Set(pos + 1)
	t.blink.Set(true) // Reset blink on input
}

// insertNewline inserts a newline character at the cursor position (Ctrl+J).
func (t *TextArea) insertNewline(ke tui.KeyEvent) {
	text := t.text.Get()
	pos := t.cursorPos.Get()
	newText := text[:pos] + "\n" + text[pos:]
	t.text.Set(newText)
	t.cursorPos.Set(pos + 1)
	t.blink.Set(true) // Reset blink on input
}

// backspace deletes the character before the cursor.
func (t *TextArea) backspace(ke tui.KeyEvent) {
	text := t.text.Get()
	pos := t.cursorPos.Get()
	if pos > 0 {
		newText := text[:pos-1] + text[pos:]
		t.text.Set(newText)
		t.cursorPos.Set(pos - 1)
	}
}

// delete deletes the character at the cursor.
func (t *TextArea) delete(ke tui.KeyEvent) {
	text := t.text.Get()
	pos := t.cursorPos.Get()
	if pos < len(text) {
		newText := text[:pos] + text[pos+1:]
		t.text.Set(newText)
	}
}

// moveLeft moves cursor left.
func (t *TextArea) moveLeft(ke tui.KeyEvent) {
	pos := t.cursorPos.Get()
	if pos > 0 {
		t.cursorPos.Set(pos - 1)
		t.blink.Set(true)
	}
}

// moveRight moves cursor right.
func (t *TextArea) moveRight(ke tui.KeyEvent) {
	pos := t.cursorPos.Get()
	if pos < len(t.text.Get()) {
		t.cursorPos.Set(pos + 1)
		t.blink.Set(true)
	}
}

// moveUp moves cursor up one line.
func (t *TextArea) moveUp(ke tui.KeyEvent) {
	lines := t.wrapText()
	row, col := t.cursorRowCol(lines)
	if row > 0 {
		// Move to same column on previous line (or end if shorter)
		prevLine := lines[row-1]
		if col > len(prevLine) {
			col = len(prevLine)
		}
		t.cursorPos.Set(t.posFromRowCol(lines, row-1, col))
		t.blink.Set(true)
	}
}

// moveDown moves cursor down one line.
func (t *TextArea) moveDown(ke tui.KeyEvent) {
	lines := t.wrapText()
	row, col := t.cursorRowCol(lines)
	if row < len(lines)-1 {
		// Move to same column on next line (or end if shorter)
		nextLine := lines[row+1]
		if col > len(nextLine) {
			col = len(nextLine)
		}
		t.cursorPos.Set(t.posFromRowCol(lines, row+1, col))
		t.blink.Set(true)
	}
}

// moveHome moves cursor to start of current line.
func (t *TextArea) moveHome(ke tui.KeyEvent) {
	lines := t.wrapText()
	row, _ := t.cursorRowCol(lines)
	t.cursorPos.Set(t.posFromRowCol(lines, row, 0))
	t.blink.Set(true)
}

// moveEnd moves cursor to end of current line.
func (t *TextArea) moveEnd(ke tui.KeyEvent) {
	lines := t.wrapText()
	row, _ := t.cursorRowCol(lines)
	t.cursorPos.Set(t.posFromRowCol(lines, row, len(lines[row])))
	t.blink.Set(true)
}

// submit calls the onSubmit callback.
func (t *TextArea) submit(ke tui.KeyEvent) {
	if t.onSubmit != nil {
		t.onSubmit(t.text.Get())
	}
}

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
			// Empty paragraph (from consecutive newlines or trailing newline)
			lines = append(lines, "")
			continue
		}

		// Wrap this paragraph to width
		var currentLine strings.Builder
		for _, r := range para {
			if currentLine.Len() >= t.width {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}
			currentLine.WriteRune(r)
		}
		lines = append(lines, currentLine.String())
	}

	return lines
}

// cursorRowCol returns the row and column of the cursor.
// This accounts for both newline characters (hard breaks) and width-based wrapping (soft breaks).
func (t *TextArea) cursorRowCol(lines []string) (row, col int) {
	text := t.text.Get()
	pos := t.cursorPos.Get()

	// Walk through the text, tracking visual row/col
	currentRow := 0
	currentCol := 0
	lineIdx := 0

	for i := 0; i < len(text) && i < pos; i++ {
		if text[i] == '\n' {
			// Hard break: move to next line
			currentRow++
			currentCol = 0
			lineIdx++
		} else {
			currentCol++
			// Check for soft wrap (width exceeded)
			if lineIdx < len(lines) && currentCol > len(lines[lineIdx]) {
				currentRow++
				currentCol = 1 // This character starts the new line
				lineIdx++
			}
		}
	}

	return currentRow, currentCol
}

// posFromRowCol converts row/col back to absolute position.
// This accounts for both newline characters (hard breaks) and width-based wrapping (soft breaks).
func (t *TextArea) posFromRowCol(lines []string, targetRow, targetCol int) int {
	text := t.text.Get()

	currentRow := 0
	currentCol := 0
	lineIdx := 0

	for i := 0; i < len(text); i++ {
		if currentRow == targetRow && currentCol == targetCol {
			return i
		}

		if text[i] == '\n' {
			// Hard break: move to next line
			if currentRow == targetRow {
				// Target is on this line but at/after end, return end of line
				return i
			}
			currentRow++
			currentCol = 0
			lineIdx++
		} else {
			currentCol++
			// Check for soft wrap (width exceeded)
			if lineIdx < len(lines) && currentCol > len(lines[lineIdx]) {
				if currentRow == targetRow {
					// Target is on this line but at/after end, return current position
					return i
				}
				currentRow++
				currentCol = 1
				lineIdx++
			}
		}
	}

	// Cursor at end of text
	return len(text)
}

// Lines returns the wrapped lines for rendering.
func (t *TextArea) Lines() []string {
	lines := t.wrapText()
	debug.Log("TextArea.Lines: %d lines, text=%q", len(lines), t.text.Get())
	return lines
}

// CursorLine returns which line the cursor is on (0-indexed).
func (t *TextArea) CursorLine() int {
	lines := t.wrapText()
	row, _ := t.cursorRowCol(lines)
	return row
}

// LineWithCursor returns a line with the cursor character inserted.
func (t *TextArea) LineWithCursor(lineIdx int) string {
	lines := t.wrapText()
	if lineIdx >= len(lines) {
		debug.Log("TextArea.LineWithCursor: lineIdx=%d >= len(lines)=%d", lineIdx, len(lines))
		return " " // Return space to ensure line takes vertical space
	}

	row, col := t.cursorRowCol(lines)
	line := lines[lineIdx]

	if lineIdx == row {
		// This line has the cursor
		cursor := "▌"
		if !t.blink.Get() {
			cursor = " "
		}
		result := ""
		if col >= len(line) {
			result = line + cursor
		} else {
			result = line[:col] + cursor + line[col:]
		}
		debug.Log("TextArea.LineWithCursor: lineIdx=%d result=%q", lineIdx, result)
		return result
	}

	// Return at least a space for empty lines to ensure they take vertical space
	if line == "" {
		debug.Log("TextArea.LineWithCursor: lineIdx=%d empty line (returning space)", lineIdx)
		return " "
	}
	debug.Log("TextArea.LineWithCursor: lineIdx=%d line=%q (no cursor)", lineIdx, line)
	return line
}

// Height returns the number of lines needed.
func (t *TextArea) Height() int {
	h := len(t.wrapText())
	if t.maxHeight > 0 && h > t.maxHeight {
		return t.maxHeight
	}
	return h
}

// Render returns the element tree for the text area.
func (t *TextArea) Render() *tui.Element {
	lines := t.wrapText()
	height := len(lines) + 2 // +2 for border

	root := tui.New(
		tui.WithBorder(tui.BorderRounded),
		tui.WithHeight(height),
		tui.WithDirection(tui.Column),
	)

	for i := range lines {
		lineEl := tui.New(
			tui.WithText(t.LineWithCursor(i)),
		)
		root.AddChild(lineEl)
	}

	return root
}
