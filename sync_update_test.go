package tui

import (
	"testing"
)

// syncRecordingTerminal embeds MockTerminal and records the order of sync
// update markers relative to output operations.
type syncRecordingTerminal struct {
	*MockTerminal
	ops []string
}

func newSyncRecordingTerminal(width, height int) *syncRecordingTerminal {
	return &syncRecordingTerminal{MockTerminal: NewMockTerminal(width, height)}
}

func (s *syncRecordingTerminal) BeginSyncUpdate() {
	s.ops = append(s.ops, "begin")
}

func (s *syncRecordingTerminal) EndSyncUpdate() {
	s.ops = append(s.ops, "end")
}

func (s *syncRecordingTerminal) Flush(changes []CellChange) {
	s.ops = append(s.ops, "flush")
	s.MockTerminal.Flush(changes)
}

func (s *syncRecordingTerminal) SetCursor(x, y int) {
	s.ops = append(s.ops, "cursor")
	s.MockTerminal.SetCursor(x, y)
}

func (s *syncRecordingTerminal) Clear() {
	s.ops = append(s.ops, "clear")
	s.MockTerminal.Clear()
}

func (s *syncRecordingTerminal) ClearToEnd() {
	s.ops = append(s.ops, "clearToEnd")
	s.MockTerminal.ClearToEnd()
}

// assertWrapped checks that ops contain exactly one begin and one end, that
// begin comes first, end comes last, and at least one output op sits between.
func assertWrapped(t *testing.T, ops []string) {
	t.Helper()
	if len(ops) < 3 {
		t.Fatalf("ops = %v, want begin + output + end", ops)
	}
	if ops[0] != "begin" {
		t.Fatalf("first op = %q, want %q (ops: %v)", ops[0], "begin", ops)
	}
	if ops[len(ops)-1] != "end" {
		t.Fatalf("last op = %q, want %q (ops: %v)", ops[len(ops)-1], "end", ops)
	}
	begins, ends := 0, 0
	for _, op := range ops {
		switch op {
		case "begin":
			begins++
		case "end":
			ends++
		}
	}
	if begins != 1 || ends != 1 {
		t.Fatalf("begin/end counts = %d/%d, want 1/1 (ops: %v)", begins, ends, ops)
	}
}

func TestRenderFrame_FullScreen_WrapsOutputInSyncUpdate(t *testing.T) {
	term := newSyncRecordingTerminal(80, 24)
	app := &App{
		terminal: term,
		buffer:   NewBuffer(80, 24),
		focus:    newFocusManager(),
		mounts:   newMountState(),
		root:     New(WithText("hello")),
	}
	app.needsFullRedraw = true

	app.renderFrame()

	assertWrapped(t, term.ops)
}

func TestRenderFrame_InlineMode_WrapsOutputInSyncUpdate(t *testing.T) {
	term := newSyncRecordingTerminal(80, 24)
	app := &App{
		terminal:       term,
		buffer:         NewBuffer(80, 6),
		inlineHeight:   6,
		inlineStartRow: 18,
		inlineLayout:   newInlineLayoutState(18),
		focus:          newFocusManager(),
		mounts:         newMountState(),
		root:           New(WithText("hello")),
	}
	app.needsFullRedraw = true

	app.renderFrame()

	assertWrapped(t, term.ops)
}

func TestAppRenderFull_WrapsOutputInSyncUpdate(t *testing.T) {
	term := newSyncRecordingTerminal(80, 24)
	app := &App{
		terminal: term,
		buffer:   NewBuffer(80, 24),
		focus:    newFocusManager(),
		mounts:   newMountState(),
		root:     New(WithText("hello")),
	}

	app.RenderFull()

	assertWrapped(t, term.ops)
}

// TestRenderFrame_Resize_AllOutputInsideSyncWindow pins that a resize frame
// (buffer/terminal size mismatch) emits every terminal write inside the sync
// window. The screen clear on resize used to run before the window opened,
// flashing blank on exactly the frame type synchronized updates target.
func TestRenderFrame_Resize_AllOutputInsideSyncWindow(t *testing.T) {
	term := newSyncRecordingTerminal(100, 30)
	app := &App{
		terminal: term,
		buffer:   NewBuffer(80, 24), // stale size: forces the resize branch
		focus:    newFocusManager(),
		mounts:   newMountState(),
		root:     New(WithText("hello")),
	}

	app.renderFrame()

	assertWrapped(t, term.ops)
}

// TestRenderFrame_SyncUpdateEndsOnPanic pins the panic-safety contract: a
// panicking postRenderHook must not leave the terminal stuck in a
// synchronized update block, or the screen freezes on supporting terminals.
func TestRenderFrame_SyncUpdateEndsOnPanic(t *testing.T) {
	term := newSyncRecordingTerminal(80, 24)
	app := &App{
		terminal: term,
		buffer:   NewBuffer(80, 24),
		focus:    newFocusManager(),
		mounts:   newMountState(),
		root:     New(WithText("hello")),
		postRenderHook: func() {
			panic("hook failure")
		},
	}
	app.needsFullRedraw = true

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected the hook panic to propagate")
			}
		}()
		app.renderFrame()
	}()

	if len(term.ops) == 0 || term.ops[len(term.ops)-1] != "end" {
		t.Fatalf("last op = %v, want trailing %q even on panic", term.ops, "end")
	}
}

// TestRenderFrame_PlainTerminal_NoSyncUpdate pins that terminals without the
// sync update methods (like MockTerminal itself) render without them.
func TestRenderFrame_PlainTerminal_NoSyncUpdate(t *testing.T) {
	term := NewMockTerminal(80, 24)
	app := &App{
		terminal: term,
		buffer:   NewBuffer(80, 24),
		focus:    newFocusManager(),
		mounts:   newMountState(),
		root:     New(WithText("hello")),
	}
	app.needsFullRedraw = true

	app.renderFrame()

	if got := rowText(term, 0); got != "hello" {
		t.Fatalf("row 0 = %q, want %q", got, "hello")
	}
}
