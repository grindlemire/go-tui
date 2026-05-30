package tui

import (
	"bytes"
	"strings"
	"testing"
)

// RenderFull clears the screen first, so trailing empty cells in each row do not
// need to be painted. Emitting them as spaces marks the cells as written, which
// stops terminals from trimming trailing whitespace when the user copies a
// selection. RenderFull must therefore skip trailing empty cells per row.
func TestRenderFull_TrimsTrailingEmptyCells(t *testing.T) {
	buf := NewBuffer(10, 2)
	buf.SetCell(0, 0, NewCell('h', NewStyle()))
	buf.SetCell(1, 0, NewCell('i', NewStyle()))
	// Row 1 is left entirely empty.

	var out bytes.Buffer
	term := NewANSITerminalWithCaps(&out, nil, Capabilities{Colors: Color16})
	RenderFull(term, buf)
	s := out.String()

	if !strings.Contains(s, "hi") {
		t.Fatalf("content missing from output: %q", s)
	}
	if strings.Contains(s, "hi ") {
		t.Errorf("trailing space written after content: %q", s)
	}
	// A fully empty row must not be painted with spaces.
	if strings.Contains(s, strings.Repeat(" ", 5)) {
		t.Errorf("empty cells painted as spaces: %q", s)
	}
}

// When a row's content shrinks between frames (e.g. scrolling a ragged
// document), the diff must clear the now-empty tail with an erase-to-end-of-line
// instead of writing spaces, so the cleared cells are not marked as written and
// terminals trim them on copy.
func TestDiff_ShrinkClearsTailWithEraseLine(t *testing.T) {
	buf := NewBuffer(20, 1)
	for x, r := range "a long line here!!!!" {
		buf.SetCell(x, 0, NewCell(r, NewStyle()))
	}
	var f1 bytes.Buffer
	t1 := NewANSITerminalWithCaps(&f1, nil, Capabilities{Colors: Color16})
	Render(t1, buf) // frame 1 + swap

	buf.Clear()
	for x, r := range "short" {
		buf.SetCell(x, 0, NewCell(r, NewStyle()))
	}
	var f2 bytes.Buffer
	t2 := NewANSITerminalWithCaps(&f2, nil, Capabilities{Colors: Color16})
	Render(t2, buf) // frame 2
	s := f2.String()

	if !strings.Contains(s, "short") {
		t.Fatalf("new content missing: %q", s)
	}
	if !strings.Contains(s, "\x1b[K") {
		t.Errorf("expected erase-to-end-of-line (ESC[K) to clear the tail: %q", s)
	}
	if strings.Contains(s, "short ") || strings.Contains(s, strings.Repeat(" ", 3)) {
		t.Errorf("tail cleared by writing spaces instead of erase-line: %q", s)
	}
}
