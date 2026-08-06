package tui

import (
	"strings"
	"testing"
)

// newInlineResizeTestApp constructs an inline-mode App on a MockTerminal.
// MockTerminal.Resize preserves content top-anchored (rows keep their
// positions, blank rows appear at the bottom), matching how terminals behave
// when the window grows and the inline widget's rows stay where they were.
func newInlineResizeTestApp(width, termHeight, inlineHeight int) (*App, *MockTerminal) {
	term := NewMockTerminal(width, termHeight)
	app := &App{
		terminal:       term,
		inlineHeight:   inlineHeight,
		inlineStartRow: termHeight - inlineHeight,
		inlineLayout:   newInlineLayoutState(termHeight - inlineHeight),
		buffer:         NewBuffer(width, inlineHeight),
		focus:          newFocusManager(),
		mounts:         newMountState(),
	}
	return app, term
}

// rowText returns the trimmed text content of a terminal row.
func rowText(term *MockTerminal, row int) string {
	lines := strings.Split(term.String(), "\n")
	if row < 0 || row >= len(lines) {
		return ""
	}
	return strings.TrimRight(lines[row], " ")
}

// TestInlineResize_TerminalGrowth_ClearsOldWidgetRows reproduces the resize
// ghost bug from issue #119: when the terminal grows taller, the full redraw
// after the ResizeEvent must erase the widget rows at the old (higher) start
// row, not just repaint at the new one.
func TestInlineResize_TerminalGrowth_ClearsOldWidgetRows(t *testing.T) {
	app, term := newInlineResizeTestApp(80, 30, 6)
	app.root = New(WithText("COMPOSER"))

	// Initial paint: widget occupies rows 24-29.
	app.needsFullRedraw = true
	app.renderFrame()
	if got := rowText(term, 24); got != "COMPOSER" {
		t.Fatalf("precondition: row 24 = %q, want %q", got, "COMPOSER")
	}

	// Terminal grows to 40 rows. Content stays top-anchored, so the old
	// widget band is still visible at rows 24-29.
	term.Resize(80, 40)
	app.Dispatch(ResizeEvent{Width: 80, Height: 40})
	app.renderFrame()

	// The widget must now be at rows 34-39.
	if got := rowText(term, 34); got != "COMPOSER" {
		t.Fatalf("row 34 = %q, want %q", got, "COMPOSER")
	}
	// The old band (rows 24-29) and the gap below it must be blank.
	for row := 24; row < 34; row++ {
		if got := rowText(term, row); got != "" {
			t.Fatalf("ghost row %d = %q, want blank", row, got)
		}
	}
}

// TestInlineResize_TerminalGrowth_RepeatedSteps mimics a resize drag: several
// growth steps between renders must clear every stale band, down to the
// earliest start row painted since the last full redraw.
func TestInlineResize_TerminalGrowth_RepeatedSteps(t *testing.T) {
	app, term := newInlineResizeTestApp(80, 30, 6)
	app.root = New(WithText("COMPOSER"))

	app.needsFullRedraw = true
	app.renderFrame()

	// Two growth events arrive before the next render (drag).
	term.Resize(80, 34)
	app.Dispatch(ResizeEvent{Width: 80, Height: 34})
	term.Resize(80, 42)
	app.Dispatch(ResizeEvent{Width: 80, Height: 42})
	app.renderFrame()

	if got := rowText(term, 36); got != "COMPOSER" {
		t.Fatalf("row 36 = %q, want %q", got, "COMPOSER")
	}
	for row := 24; row < 36; row++ {
		if got := rowText(term, row); got != "" {
			t.Fatalf("ghost row %d = %q, want blank", row, got)
		}
	}
}

// TestInlineResize_TerminalShrink_UnchangedBehavior pins the shrink path: the
// new start row is above the old one, so clearing from the new start row
// already covers the old band. No extra clearing should occur above it.
func TestInlineResize_TerminalShrink_UnchangedBehavior(t *testing.T) {
	app, term := newInlineResizeTestApp(80, 40, 6)
	app.root = New(WithText("COMPOSER"))

	app.needsFullRedraw = true
	app.renderFrame()
	if got := rowText(term, 34); got != "COMPOSER" {
		t.Fatalf("precondition: row 34 = %q, want %q", got, "COMPOSER")
	}

	// History line above the shrunken viewport must survive the redraw.
	term.SetCell(0, 23, NewCell('H', NewStyle()))

	term.Resize(80, 30)
	app.Dispatch(ResizeEvent{Width: 80, Height: 30})
	app.renderFrame()

	if got := rowText(term, 24); got != "COMPOSER" {
		t.Fatalf("row 24 = %q, want %q", got, "COMPOSER")
	}
	if got := rowText(term, 23); got != "H" {
		t.Fatalf("history row 23 = %q, want %q (must not be cleared)", got, "H")
	}
}
