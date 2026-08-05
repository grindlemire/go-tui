package tui

import (
	"testing"
)

// writeDirectRecorder wraps MockTerminal to capture raw WriteDirect payloads.
type writeDirectRecorder struct {
	*MockTerminal
	writes []string
}

func (t *writeDirectRecorder) WriteDirect(b []byte) (int, error) {
	t.writes = append(t.writes, string(b))
	return len(b), nil
}

// TestRenderInline_WrapsFrameInSynchronizedUpdate verifies the whole inline
// frame is wrapped in a DEC 2026 synchronized update so resize drags do not
// flash partially-rendered intermediate states.
func TestRenderInline_WrapsFrameInSynchronizedUpdate(t *testing.T) {
	const width, termHeight, inlineHeight = 80, 24, 6
	rec := &writeDirectRecorder{MockTerminal: NewMockTerminal(width, termHeight)}
	app := &App{
		terminal:       rec,
		buffer:         NewBuffer(width, inlineHeight),
		inlineHeight:   inlineHeight,
		inlineStartRow: 18,
	}
	app.needsFullRedraw = true
	app.renderInline()

	if len(rec.writes) < 2 {
		t.Fatalf("expected at least 2 WriteDirect calls, got %d", len(rec.writes))
	}
	if rec.writes[0] != "\x1b[?2026h" {
		t.Fatalf("first WriteDirect = %q, want synchronized update begin", rec.writes[0])
	}
	if last := rec.writes[len(rec.writes)-1]; last != "\x1b[?2026l" {
		t.Fatalf("last WriteDirect = %q, want synchronized update end", last)
	}
}

// TestForceFullRedraw verifies ForceFullRedraw sets needsFullRedraw and marks
// the app dirty so the next render repaints the whole viewport.
func TestForceFullRedraw(t *testing.T) {
	app := &App{}
	app.ForceFullRedraw()
	if !app.needsFullRedraw {
		t.Fatal("ForceFullRedraw should set needsFullRedraw")
	}
	if !app.dirty.Load() {
		t.Fatal("ForceFullRedraw should mark the app dirty")
	}
}
