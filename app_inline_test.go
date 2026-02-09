package tui

import (
	"fmt"
	"testing"
)

// newInlineTestApp creates an App configured for inline mode testing with an
// EmulatorTerminal. The app is NOT fully initialized via NewApp (which requires
// a real terminal) — instead we directly construct the App struct with the
// fields that SetInlineHeight, PrintAboveln, and the scroll helpers need.
func newInlineTestApp(termWidth, termHeight, inlineHeight int) (*App, *EmulatorTerminal) {
	emu := NewEmulatorTerminal(termWidth, termHeight)

	app := &App{
		terminal:       emu,
		inlineHeight:   inlineHeight,
		inlineStartRow: termHeight - inlineHeight,
		buffer:         NewBuffer(termWidth, inlineHeight),
		focus:          NewFocusManager(),
		reader:         NewMockEventReader(),
		eventQueue:     make(chan func(), 256),
		stopCh:         make(chan struct{}),
	}

	return app, emu
}

func TestSetInlineHeight_GrowingWithNoHistory_NoBlankScrollback(t *testing.T) {
	type tc struct {
		startHeight int
		endHeight   int
	}

	tests := map[string]tc{
		"grow by 1": {
			startHeight: 3,
			endHeight:   4,
		},
		"grow by 3": {
			startHeight: 3,
			endHeight:   6,
		},
		"grow by 10": {
			startHeight: 3,
			endHeight:   13,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app, emu := newInlineTestApp(80, 24, tt.startHeight)

			app.SetInlineHeight(tt.endHeight)

			blankCount := emu.BlankScrollbackCount()
			if blankCount > 0 {
				t.Errorf("got %d blank lines in scrollback, want 0\n%s", blankCount, emu.DumpState())
			}

			totalScrollback := len(emu.Scrollback())
			if totalScrollback > 0 {
				t.Errorf("got %d total scrollback lines, want 0 (nothing should scroll when history is empty)\n%s",
					totalScrollback, emu.DumpState())
			}
		})
	}
}

func TestSetInlineHeight_GrowingWithHistory_ContentStaysVisible(t *testing.T) {
	type tc struct {
		historyLines []string
		growBy       int
	}

	tests := map[string]tc{
		"1 history line, grow by 1": {
			historyLines: []string{"hello"},
			growBy:       1,
		},
		"3 history lines, grow by 2": {
			historyLines: []string{"line1", "line2", "line3"},
			growBy:       2,
		},
		"2 history lines, grow by 5": {
			historyLines: []string{"line1", "line2"},
			growBy:       5,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app, emu := newInlineTestApp(80, 24, 3)

			// Print history lines
			for _, line := range tt.historyLines {
				app.printAboveRaw(line + "\n")
			}
			emu.scrollback = nil

			app.SetInlineHeight(3 + tt.growBy)

			// No blank lines in scrollback
			blanks := emu.BlankScrollbackCount()
			if blanks > 0 {
				t.Errorf("got %d blank scrollback lines, want 0\n%s", blanks, emu.DumpState())
			}

			// Content should stay on screen (not be pushed to scrollback)
			// since there are plenty of blank rows to absorb the growth
			for _, line := range tt.historyLines {
				found := false
				for r := 0; r < 24; r++ {
					if emu.ScreenRow(r) == line {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("content %q not found on screen after growth\n%s", line, emu.DumpState())
				}
			}
		})
	}
}

func TestSetInlineHeight_RepeatedGrowth_NoBlankAccumulation(t *testing.T) {
	// Simulates the actual bug: user types multiline content, textarea grows
	// from 3 → 4 → 5 → 6 → ... with no PrintAboveln calls in between.
	app, emu := newInlineTestApp(80, 24, 3)

	// Grow one row at a time, simulating textarea growth
	for h := 4; h <= 12; h++ {
		app.SetInlineHeight(h)
	}

	blanks := emu.BlankScrollbackCount()
	total := len(emu.Scrollback())

	if blanks > 0 {
		t.Errorf("after growing from 3→12 with no history: got %d blank scrollback lines, want 0\n%s",
			blanks, emu.DumpState())
	}
	if total > 0 {
		t.Errorf("after growing from 3→12 with no history: got %d total scrollback lines, want 0\n%s",
			total, emu.DumpState())
	}
}

func TestSetInlineHeight_GrowAfterPrintAboveln(t *testing.T) {
	// Print some content, then grow. Content should stay visible on screen
	// because there are plenty of blank rows to absorb the growth.
	app, emu := newInlineTestApp(80, 24, 3)

	app.printAboveRaw("You: hello\n")
	app.printAboveRaw("Bot: hi there\n")
	app.printAboveRaw("You: how are you?\n")
	emu.scrollback = nil

	app.SetInlineHeight(6)

	blanks := emu.BlankScrollbackCount()
	if blanks > 0 {
		t.Errorf("got %d blank scrollback lines, want 0\n%s", blanks, emu.DumpState())
	}

	// All content should remain visible on screen
	expected := []string{"You: hello", "Bot: hi there", "You: how are you?"}
	for _, line := range expected {
		found := false
		for r := 0; r < 24; r++ {
			if emu.ScreenRow(r) == line {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("content %q not found on screen\n%s", line, emu.DumpState())
		}
	}
}

func TestSetInlineHeight_GrowMoreThanHistory(t *testing.T) {
	// Grow by more rows than there are history lines.
	// Content should stay visible (plenty of blank rows to absorb growth).
	app, emu := newInlineTestApp(80, 24, 3)

	app.printAboveRaw("only one line\n")
	emu.scrollback = nil

	app.SetInlineHeight(8)

	blanks := emu.BlankScrollbackCount()
	if blanks > 0 {
		t.Errorf("got %d blank scrollback lines, want 0\n%s", blanks, emu.DumpState())
	}

	// Content should remain on screen
	found := false
	for r := 0; r < 24; r++ {
		if emu.ScreenRow(r) == "only one line" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("content not found on screen\n%s", emu.DumpState())
	}
}

func TestSetInlineHeight_GrowExceedsBlankSpace_ContentGoesToScrollback(t *testing.T) {
	// Small terminal (10 rows), widget starts at 3, leaving 7 history rows.
	// Fill history completely (7 content rows), then grow by 3.
	// blankRows = 0, so all 3 must come from scrolling content to scrollback.
	app, emu := newInlineTestApp(80, 10, 3)

	// Fill history area completely: 7 rows
	for i := 0; i < 7; i++ {
		app.printAboveRaw(fmt.Sprintf("line%d\n", i))
	}
	emu.scrollback = nil

	// Grow by 3 — no blank rows available, content must scroll
	app.SetInlineHeight(6)

	blanks := emu.BlankScrollbackCount()
	if blanks > 0 {
		t.Errorf("got %d blank scrollback lines, want 0\n%s", blanks, emu.DumpState())
	}

	// 3 oldest content lines should be in scrollback
	nonBlank := emu.NonBlankScrollback()
	if len(nonBlank) != 3 {
		t.Errorf("got %d non-blank scrollback lines, want 3\n  got: %v\n%s",
			len(nonBlank), nonBlank, emu.DumpState())
	}
}

func TestSetInlineHeight_ShrinkThenGrow_NoBlankScrollback(t *testing.T) {
	// Widget grows, user submits (widget shrinks), user types again (grows).
	// The shrink+grow cycle should not introduce blank lines.
	app, emu := newInlineTestApp(80, 24, 3)

	// Phase 1: grow with no history
	app.SetInlineHeight(6)

	// Phase 2: user submits — print content, then shrink
	app.printAboveRaw("You: hello world\n")
	app.SetInlineHeight(3)

	// Phase 3: user starts typing again — grow
	emu.scrollback = nil // reset for this phase
	app.SetInlineHeight(6)

	blanks := emu.BlankScrollbackCount()
	if blanks > 0 {
		t.Errorf("grow after shrink: got %d blank scrollback lines, want 0\n%s",
			blanks, emu.DumpState())
	}
}

func TestSetInlineHeight_InlineHeightAndStartRowCorrect(t *testing.T) {
	type tc struct {
		termHeight  int
		startHeight int
		newHeight   int
		wantHeight  int
		wantStart   int
	}

	tests := map[string]tc{
		"grow": {
			termHeight:  24,
			startHeight: 3,
			newHeight:   5,
			wantHeight:  5,
			wantStart:   19,
		},
		"shrink": {
			termHeight:  24,
			startHeight: 5,
			newHeight:   3,
			wantHeight:  3,
			wantStart:   21,
		},
		"cap to terminal height": {
			termHeight:  10,
			startHeight: 3,
			newHeight:   15,
			wantHeight:  10,
			wantStart:   0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app, _ := newInlineTestApp(80, tt.termHeight, tt.startHeight)

			app.SetInlineHeight(tt.newHeight)

			if app.inlineHeight != tt.wantHeight {
				t.Errorf("inlineHeight = %d, want %d", app.inlineHeight, tt.wantHeight)
			}
			if app.inlineStartRow != tt.wantStart {
				t.Errorf("inlineStartRow = %d, want %d", app.inlineStartRow, tt.wantStart)
			}
		})
	}
}

func TestSetInlineHeight_NoChangeIsNoop(t *testing.T) {
	app, emu := newInlineTestApp(80, 24, 5)

	app.SetInlineHeight(5) // same height

	if len(emu.Scrollback()) > 0 {
		t.Errorf("no-change call should not produce scrollback\n%s", emu.DumpState())
	}
}

func TestPrintAboveRaw_AddsToScreen(t *testing.T) {
	app, emu := newInlineTestApp(80, 24, 3)

	app.printAboveRaw("hello world\n")

	// The text should appear on the screen in the history area
	// (the row just above the widget)
	historyBottom := app.inlineStartRow - 1
	row := emu.ScreenRow(historyBottom)
	if row != "hello world" {
		t.Errorf("history row = %q, want %q\n%s", row, "hello world", emu.DumpState())
	}
}

func TestPrintAboveRaw_TracksHistoryRows(t *testing.T) {
	type tc struct {
		prints    []string
		wantCount int
	}

	tests := map[string]tc{
		"single line": {
			prints:    []string{"hello\n"},
			wantCount: 1,
		},
		"three lines": {
			prints:    []string{"a\n", "b\n", "c\n"},
			wantCount: 3,
		},
		"no trailing newline": {
			prints:    []string{"no newline"},
			wantCount: 1,
		},
		"multi-line content": {
			prints:    []string{"line1\nline2\n"},
			wantCount: 2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app, _ := newInlineTestApp(80, 24, 3)

			for _, content := range tt.prints {
				app.printAboveRaw(content)
			}

			if app.historyRows != tt.wantCount {
				t.Errorf("historyRows = %d, want %d", app.historyRows, tt.wantCount)
			}
		})
	}
}

func TestSetInlineHeight_HistoryRowsCapped(t *testing.T) {
	app, _ := newInlineTestApp(80, 24, 3)

	for i := 0; i < 30; i++ {
		app.printAboveRaw(fmt.Sprintf("line %d\n", i))
	}

	if app.historyRows > app.inlineStartRow {
		t.Errorf("historyRows = %d, exceeds inlineStartRow = %d",
			app.historyRows, app.inlineStartRow)
	}
}

func TestEmulatorTerminal_ScrollRegionUp(t *testing.T) {
	// Verify the emulator correctly implements scroll region behavior
	emu := NewEmulatorTerminal(20, 5)

	// Set up some content
	emu.SetScreenRow(0, "row0")
	emu.SetScreenRow(1, "row1")
	emu.SetScreenRow(2, "row2")
	emu.SetScreenRow(3, "row3")
	emu.SetScreenRow(4, "row4")

	// Set scroll region to rows 1-3 (0-indexed), i.e. ANSI rows 1-4
	// Then position cursor at bottom and emit \n to scroll
	emu.WriteDirect([]byte("\033[1;4r"))  // scroll region rows 1-4 (ANSI 1-indexed)
	emu.WriteDirect([]byte("\033[4;1H"))  // cursor to row 4 (ANSI 1-indexed) = row 3 (0-indexed)
	emu.WriteDirect([]byte("\n"))         // scroll within region

	// Row 0 (outside region above) should be pushed to scrollback
	// Actually wait — scrollTop=0, scrollBottom=3 (rows 0-3)
	// The scroll moves row 0 to scrollback, rows 1-3 shift up, row 3 becomes blank
	if emu.ScreenRow(0) != "row1" {
		t.Errorf("row 0 = %q, want %q", emu.ScreenRow(0), "row1")
	}
	if emu.ScreenRow(1) != "row2" {
		t.Errorf("row 1 = %q, want %q", emu.ScreenRow(1), "row2")
	}
	if emu.ScreenRow(2) != "row3" {
		t.Errorf("row 2 = %q, want %q", emu.ScreenRow(2), "row3")
	}
	if emu.ScreenRow(3) != "" {
		t.Errorf("row 3 = %q, want blank", emu.ScreenRow(3))
	}
	// Row 4 should be untouched
	if emu.ScreenRow(4) != "row4" {
		t.Errorf("row 4 = %q, want %q (should be untouched)", emu.ScreenRow(4), "row4")
	}

	// Scrollback should contain the top row
	sb := emu.Scrollback()
	if len(sb) != 1 || sb[0] != "row0" {
		t.Errorf("scrollback = %v, want [row0]", sb)
	}
}

func TestEmulatorTerminal_ReverseIndex(t *testing.T) {
	emu := NewEmulatorTerminal(20, 5)

	emu.SetScreenRow(0, "row0")
	emu.SetScreenRow(1, "row1")
	emu.SetScreenRow(2, "row2")
	emu.SetScreenRow(3, "row3")
	emu.SetScreenRow(4, "row4")

	// Cursor at top of screen, reverse index inserts blank at top
	emu.WriteDirect([]byte("\033[1;1H")) // cursor to row 1, col 1 (ANSI)
	emu.WriteDirect([]byte("\033M"))     // reverse index

	if emu.ScreenRow(0) != "" {
		t.Errorf("row 0 = %q, want blank", emu.ScreenRow(0))
	}
	if emu.ScreenRow(1) != "row0" {
		t.Errorf("row 1 = %q, want %q", emu.ScreenRow(1), "row0")
	}
	if emu.ScreenRow(2) != "row1" {
		t.Errorf("row 2 = %q, want %q", emu.ScreenRow(2), "row1")
	}
	if emu.ScreenRow(3) != "row2" {
		t.Errorf("row 3 = %q, want %q", emu.ScreenRow(3), "row2")
	}
	// Row 4 had "row3" pushed into it, "row4" fell off
	if emu.ScreenRow(4) != "row3" {
		t.Errorf("row 4 = %q, want %q", emu.ScreenRow(4), "row3")
	}

	// Nothing should go to scrollback (reverse index drops the bottom)
	if len(emu.Scrollback()) > 0 {
		t.Errorf("scrollback = %v, want empty (reverse index drops bottom)", emu.Scrollback())
	}
}

func TestEmulatorTerminal_EraseLine(t *testing.T) {
	emu := NewEmulatorTerminal(10, 3)

	emu.SetScreenRow(0, "0123456789")
	emu.SetScreenRow(1, "abcdefghij")

	// Move to row 1 (ANSI row 2) and clear entire line
	emu.WriteDirect([]byte("\033[2;1H\033[2K"))

	if emu.ScreenRow(0) != "0123456789" {
		t.Errorf("row 0 = %q, want unchanged", emu.ScreenRow(0))
	}
	if emu.ScreenRow(1) != "" {
		t.Errorf("row 1 = %q, want blank", emu.ScreenRow(1))
	}
}
