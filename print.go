package tui

import (
	"io"
	"os"
	"strings"
)

// PrintOption configures single-frame rendering.
type PrintOption func(*printConfig)

type printConfig struct {
	width int // 0 means auto-detect
}

// WithPrintWidth sets an explicit width in characters.
// Default: auto-detect terminal width, falling back to 80 columns.
func WithPrintWidth(w int) PrintOption {
	return func(c *printConfig) {
		c.width = w
	}
}

// Print renders an element tree to stdout with ANSI styling.
// Width is auto-detected from the terminal unless overridden with WithPrintWidth.
func Print(el *Element, opts ...PrintOption) {
	Fprint(os.Stdout, el, opts...)
}

// Sprint renders an element tree to a string with ANSI escape codes.
// Width is auto-detected from the terminal unless overridden with WithPrintWidth.
func Sprint(el *Element, opts ...PrintOption) string {
	cfg := printConfig{}
	for _, o := range opts {
		o(&cfg)
	}

	width := cfg.width
	if width <= 0 {
		w, _, err := getTerminalSize(int(os.Stdout.Fd()))
		if err != nil || w <= 0 {
			width = 80
		} else {
			width = w
		}
	}

	caps := DetectCapabilities()
	buf, height := renderElementToBuffer(el, width, caps)
	if height == 0 {
		return ""
	}

	var sb strings.Builder
	esc := newEscBuilder(256)
	for row := 0; row < height; row++ {
		if row > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(bufferRowToANSI(buf, row, esc, caps))
	}
	return sb.String()
}

// Fprint renders an element tree to the given writer with ANSI styling.
// Width is auto-detected from the terminal unless overridden with WithPrintWidth.
func Fprint(w io.Writer, el *Element, opts ...PrintOption) {
	s := Sprint(el, opts...)
	if s != "" {
		io.WriteString(w, s)
		io.WriteString(w, "\n")
	}
}
