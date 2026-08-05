package tui

import (
	"testing"
)

// TestRenderInline_FullRedrawClearsUnionOfViewports is a regression test for
// the inline-mode resize ghost bug: after the terminal grows, a full redraw
// must clear from the union of the previous and current viewport start rows
// (min), otherwise the old composer band stays above the new viewport.
func TestRenderInline_FullRedrawClearsUnionOfViewports(t *testing.T) {
	const width, termHeight, inlineHeight = 80, 40, 6
	term := NewMockTerminal(width, termHeight)
	app := &App{
		terminal:           term,
		buffer:             NewBuffer(width, inlineHeight),
		inlineHeight:       inlineHeight,
		inlineStartRow:     34, // current viewport top (terminal grew from 36 to 40 rows)
		prevInlineStartRow: 30, // previous viewport top
	}
	app.needsFullRedraw = true
	app.renderInline()

	// Clear must start at min(30, 34) = 30.
	if _, cy := term.Cursor(); cy != 30 {
		t.Fatalf("full redraw clear start row = %d, want 30 (union of old/new viewport)", cy)
	}
	if app.needsFullRedraw {
		t.Fatal("needsFullRedraw should be consumed by the full redraw")
	}
	if app.prevInlineStartRow != 34 {
		t.Fatalf("prevInlineStartRow = %d, want 34", app.prevInlineStartRow)
	}
}
// TestRenderInline_FirstFullRedrawUsesCurrentStartRow is a regression test:
// on the very first full redraw (prevInlineStartRow = -1) the clear must start
// at the current viewport top, never degrade to row 0 and wipe the log area.
func TestRenderInline_FirstFullRedrawUsesCurrentStartRow(t *testing.T) {
	const width, termHeight, inlineHeight = 80, 40, 6
	term := NewMockTerminal(width, termHeight)
	app := &App{
		terminal:           term,
		buffer:             NewBuffer(width, inlineHeight),
		inlineHeight:       inlineHeight,
		inlineStartRow:     34,
		prevInlineStartRow: -1,
	}
	app.needsFullRedraw = true
	app.renderInline()

	if _, cy := term.Cursor(); cy != 34 {
		t.Fatalf("first full redraw clear start row = %d, want 34", cy)
	}
}
// TestRenderFrame_ResyncsInlineGeometryOnHeightChange is a regression test:
// when the terminal height changes (even before the ResizeEvent is dispatched),
// renderFrame must immediately re-sync inlineStartRow to the live size and force
// a full redraw; otherwise the composer is drawn at a stale row, leaving the old
// band behind (resize ghosts).
func TestRenderFrame_ResyncsInlineGeometryOnHeightChange(t *testing.T) {
	term := NewMockTerminal(80, 24)
	app := &App{
		terminal:           term,
		buffer:             NewBuffer(80, 6),
		inlineHeight:       6,
		inlineStartRow:     18,
		prevInlineStartRow: 18,
		inlineLayout:       newInlineLayoutState(18),
		focus:              newFocusManager(),
	}
	// Simulate the terminal growing to 30 rows with no ResizeEvent dispatched.
	term.Resize(80, 30)

	app.renderFrame()

	if app.inlineStartRow != 24 {
		t.Fatalf("inlineStartRow = %d, want 24 (30-6)", app.inlineStartRow)
	}
	// The full redraw ran and recorded the new start row.
	if app.prevInlineStartRow != 24 {
		t.Fatalf("prevInlineStartRow = %d, want 24", app.prevInlineStartRow)
	}
}

