package tui

import (
	"bytes"
	"io"
	"os"
)

// ANSITerminal implements Terminal using ANSI escape sequences.
// It works with any terminal emulator that supports ANSI codes.
type ANSITerminal struct {
	out       io.Writer     // Output destination (usually os.Stdout)
	in        io.Reader     // Input source (usually os.Stdin)
	caps      Capabilities  // Terminal capabilities
	lastStyle Style         // Last emitted style (for optimization)
	esc       *escBuilder   // Escape sequence builder
	inFd      uintptr       // File descriptor for input (needed for raw mode)
	outFd     uintptr       // File descriptor for output (needed for size query)
	rawState  *rawModeState // Platform-specific raw mode state
}

// NewANSITerminal creates a new ANSI terminal with auto-detected capabilities.
// The output writer is typically os.Stdout and the input reader is os.Stdin.
func NewANSITerminal(out io.Writer, in io.Reader) (*ANSITerminal, error) {
	caps := defaultCapabilities()

	t := &ANSITerminal{
		out:  out,
		in:   in,
		caps: caps,
		esc:  newEscBuilder(4096),
	}

	// Try to get file descriptors for size queries and raw mode
	if f, ok := out.(*os.File); ok {
		t.outFd = f.Fd()
	}
	if f, ok := in.(*os.File); ok {
		t.inFd = f.Fd()
	}

	return t, nil
}

// NewANSITerminalWithCaps creates a new ANSI terminal with explicit capabilities.
// Use this when you want to override auto-detection.
func NewANSITerminalWithCaps(out io.Writer, in io.Reader, caps Capabilities) *ANSITerminal {
	t := &ANSITerminal{
		out:  out,
		in:   in,
		caps: caps,
		esc:  newEscBuilder(4096),
	}

	if f, ok := out.(*os.File); ok {
		t.outFd = f.Fd()
	}
	if f, ok := in.(*os.File); ok {
		t.inFd = f.Fd()
	}

	return t
}

// defaultCapabilities returns conservative default capabilities.
func defaultCapabilities() Capabilities {
	return Capabilities{
		Colors:    Color16,
		Unicode:   true,
		TrueColor: false,
		AltScreen: true,
	}
}

// Size returns the terminal dimensions.
// Returns a default of 80x24 if the size cannot be determined.
func (t *ANSITerminal) Size() (width, height int) {
	w, h, err := getTerminalSize(int(t.outFd))
	if err != nil {
		return 80, 24 // Sensible default
	}
	return w, h
}

// Flush writes the given cell changes to the terminal.
// It optimizes cursor movement and style changes for efficiency.
func (t *ANSITerminal) Flush(changes []CellChange) {
	if len(changes) == 0 {
		return
	}

	t.esc.Reset()
	lastX, lastY := -1, -1

	for _, ch := range changes {
		// Skip continuation cells entirely - they represent the second column
		// of a wide character, which was already rendered by the primary cell.
		// Processing them would incorrectly move the cursor backwards.
		if ch.Cell.IsContinuation() {
			continue
		}
		// Optimize cursor movement
		needsMove := false
		if ch.Y != lastY {
			needsMove = true
		} else if ch.X != lastX+1 {
			// Not sequential on the same row
			needsMove = true
		}

		if needsMove {
			t.esc.MoveTo(ch.X, ch.Y)
		}

		// Only emit style changes when style differs
		if !ch.Cell.Style.Equal(t.lastStyle) {
			t.esc.SetStyle(ch.Cell.Style, t.caps)
			t.lastStyle = ch.Cell.Style
		}

		// Write the character (skip continuation cells)
		if !ch.Cell.IsContinuation() {
			if ch.Cell.Rune != 0 {
				t.esc.WriteRune(ch.Cell.Rune)
			} else {
				t.esc.WriteRune(' ')
			}
		}

		lastX = ch.X
		if !ch.Cell.IsContinuation() && ch.Cell.Width > 1 {
			// Wide character advances cursor by its width
			lastX = ch.X + int(ch.Cell.Width) - 1
		}
		lastY = ch.Y
	}

	t.out.Write(t.esc.Bytes())
}

// Clear clears the entire terminal screen.
func (t *ANSITerminal) Clear() {
	t.esc.Reset()
	t.esc.ResetStyle()
	t.esc.MoveTo(0, 0)     // Home first
	t.esc.ClearScreen()    // ESC[2J - clear visible screen
	t.esc.ClearScrollback() // ESC[3J - also clear scrollback (helps with resize)
	t.esc.MoveTo(0, 0)     // Ensure cursor at home after clear
	t.out.Write(t.esc.Bytes())
	t.lastStyle = NewStyle()
}

// SetCursor moves the cursor to the specified position (0-indexed).
func (t *ANSITerminal) SetCursor(x, y int) {
	t.esc.Reset()
	t.esc.MoveTo(x, y)
	t.out.Write(t.esc.Bytes())
}

// HideCursor makes the cursor invisible.
func (t *ANSITerminal) HideCursor() {
	t.esc.Reset()
	t.esc.HideCursor()
	t.out.Write(t.esc.Bytes())
}

// ShowCursor makes the cursor visible.
func (t *ANSITerminal) ShowCursor() {
	t.esc.Reset()
	t.esc.ShowCursor()
	t.out.Write(t.esc.Bytes())
}

// EnterRawMode puts the terminal into raw mode.
// This is implemented in platform-specific files.
func (t *ANSITerminal) EnterRawMode() error {
	state, err := enableRawMode(int(t.inFd))
	if err != nil {
		return err
	}
	t.rawState = state
	return nil
}

// ExitRawMode restores the terminal to its previous mode.
// This is implemented in platform-specific files.
func (t *ANSITerminal) ExitRawMode() error {
	if t.rawState == nil {
		return nil
	}
	err := disableRawMode(t.rawState)
	t.rawState = nil
	return err
}

// EnterAltScreen switches to the alternate screen buffer.
func (t *ANSITerminal) EnterAltScreen() {
	t.esc.Reset()
	t.esc.EnterAltScreen()
	t.out.Write(t.esc.Bytes())
}

// ExitAltScreen switches back to the main screen buffer.
func (t *ANSITerminal) ExitAltScreen() {
	t.esc.Reset()
	t.esc.ExitAltScreen()
	t.out.Write(t.esc.Bytes())
}

// BeginSyncUpdate starts a synchronized update block.
// Output is buffered until EndSyncUpdate, then displayed atomically.
func (t *ANSITerminal) BeginSyncUpdate() {
	t.esc.Reset()
	t.esc.BeginSyncUpdate()
	t.out.Write(t.esc.Bytes())
}

// EndSyncUpdate ends a synchronized update block.
func (t *ANSITerminal) EndSyncUpdate() {
	t.esc.Reset()
	t.esc.EndSyncUpdate()
	t.out.Write(t.esc.Bytes())
}

// Caps returns the terminal's capabilities.
func (t *ANSITerminal) Caps() Capabilities {
	return t.caps
}

// SetCaps updates the terminal's capabilities.
// This is useful after detecting capabilities at runtime.
func (t *ANSITerminal) SetCaps(caps Capabilities) {
	t.caps = caps
}

// ResetStyle resets the style tracking, forcing the next Flush to emit style codes.
func (t *ANSITerminal) ResetStyle() {
	t.lastStyle = Style{Fg: RGBColor(255, 255, 255)} // Use something that won't match
}

// Writer returns the underlying writer for direct output.
// Use with caution as it bypasses the terminal's buffering.
func (t *ANSITerminal) Writer() io.Writer {
	return t.out
}

// WriteDirect writes raw bytes directly to the terminal output.
// Use for escape sequences or content that doesn't need processing.
func (t *ANSITerminal) WriteDirect(b []byte) (int, error) {
	return t.out.Write(b)
}

// BufferedWriter provides a buffered writer for efficient batch writes.
type BufferedWriter struct {
	buf bytes.Buffer
	out io.Writer
}

// NewBufferedWriter creates a buffered writer wrapping the given writer.
func NewBufferedWriter(out io.Writer) *BufferedWriter {
	return &BufferedWriter{out: out}
}

// Write writes bytes to the buffer.
func (w *BufferedWriter) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

// Flush writes the buffer contents to the underlying writer and clears the buffer.
func (w *BufferedWriter) Flush() error {
	_, err := w.out.Write(w.buf.Bytes())
	w.buf.Reset()
	return err
}
