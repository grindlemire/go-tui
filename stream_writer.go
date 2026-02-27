package tui

import "io"

// StreamWriter wraps an inner stream writer with convenience methods for
// styled and gradient-colored output. It implements io.WriteCloser for
// backward compatibility with code that used the previous io.WriteCloser
// return type of App.StreamAbove().
type StreamWriter struct {
	w     io.WriteCloser // inner writer (inlineStreamWriter or nopStreamWriter)
	app   *App           // app reference for WriteElement
	col   int            // current column for gradient position
	width int            // terminal width (cached at creation)
	esc   escBuilder     // reusable escape sequence builder
	caps  Capabilities   // terminal capabilities
	nop   bool           // true when not in inline mode
}

// Write delegates to the inner writer. Plain Write does not track column
// position since it handles raw bytes that may contain ANSI sequences.
func (sw *StreamWriter) Write(p []byte) (int, error) {
	return sw.w.Write(p)
}

// Close delegates to the inner writer.
func (sw *StreamWriter) Close() error {
	return sw.w.Close()
}

// WriteStyled writes text wrapped in the style's ANSI prefix and a reset
// suffix. Uses escBuilder.SetStyle for capability-aware rendering.
// Advances the internal column counter by the display width of the text.
func (sw *StreamWriter) WriteStyled(text string, style Style) (int, error) {
	if sw.nop {
		return len(text), nil
	}

	sw.esc.Reset()
	sw.esc.SetStyle(style, sw.caps)
	sw.esc.WriteString(text)
	sw.esc.ResetStyle()

	// Track column position.
	for _, r := range text {
		if r == '\n' {
			sw.col = 0
		} else {
			sw.col += RuneWidth(r)
			if sw.width > 0 && sw.col >= sw.width {
				sw.col = 0
			}
		}
	}

	return sw.w.Write(sw.esc.Bytes())
}

// WriteGradient writes each character with an interpolated gradient foreground
// color. An optional base style provides additional attributes (bold, italic,
// etc.) and background color. The gradient color replaces the base style's
// foreground.
//
// Column position is tracked internally and wraps at the terminal width.
// Newlines reset the column to 0. The entire styled output is built in a
// single buffer and written in one call to the inner writer.
func (sw *StreamWriter) WriteGradient(text string, g Gradient, base ...Style) (int, error) {
	if sw.nop {
		return len(text), nil
	}

	var baseStyle Style
	if len(base) > 0 {
		baseStyle = base[0]
	}

	sw.esc.Reset()

	for _, r := range text {
		if r == '\n' {
			sw.esc.ResetStyle()
			sw.esc.WriteRune('\n')
			sw.col = 0
			continue
		}

		// Compute gradient position.
		w := sw.width
		if w < 1 {
			w = 80
		}
		t := float64(sw.col) / float64(w-1)
		if t > 1 {
			t = 1
		}

		charStyle := baseStyle
		charStyle.Fg = g.At(t)
		sw.esc.SetStyle(charStyle, sw.caps)
		sw.esc.WriteRune(r)

		sw.col += RuneWidth(r)
		if sw.width > 0 && sw.col >= sw.width {
			sw.col = 0
		}
	}

	sw.esc.ResetStyle()

	return sw.w.Write(sw.esc.Bytes())
}

// WriteElement renders a Viewable and inserts the resulting rows into the
// inline scrollback mid-stream. Any current partial line is finalized before
// the element rows are inserted. After insertion, the next Write call starts
// a fresh partial line.
// No-op if the writer is in nop mode or the app is nil.
func (sw *StreamWriter) WriteElement(v Viewable) {
	if sw.nop || sw.app == nil {
		return
	}
	sw.app.QueueUpdate(func() {
		sw.app.PrintAboveElement(v)
	})
	sw.col = 0
}
