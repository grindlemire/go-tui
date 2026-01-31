package tui

import (
	"strings"
	"testing"
)

// TestIntegration_RenderPipeline tests the full render pipeline:
// Buffer → Diff → MockTerminal → verify output
func TestIntegration_RenderPipeline(t *testing.T) {
	buf := NewBuffer(20, 5)
	term := NewMockTerminal(20, 5)
	style := NewStyle()

	// Write some content
	buf.SetString(5, 2, "Hello", style)

	// Render to terminal
	Render(term, buf)

	// Verify the content is in the mock terminal
	for i, r := range "Hello" {
		cell := term.CellAt(5+i, 2)
		if cell.Rune != r {
			t.Errorf("Cell(%d, 2).Rune = %q, want %q", 5+i, cell.Rune, r)
		}
	}

	// Verify surrounding cells are spaces
	if term.CellAt(4, 2).Rune != ' ' {
		t.Error("Cell before text should be space")
	}
	if term.CellAt(10, 2).Rune != ' ' {
		t.Error("Cell after text should be space")
	}
}

func TestIntegration_BorderedBoxWithTitle(t *testing.T) {
	buf := NewBuffer(20, 6)
	term := NewMockTerminal(20, 6)
	style := NewStyle()

	// Draw a bordered box with title
	DrawBoxWithTitle(buf, NewRect(2, 1, 15, 4), BorderSingle, "Title", style)

	// Render to terminal
	Render(term, buf)

	// Expected output (trimmed):
	// "  ┌────Title────┐"
	// "  │             │"
	// "  │             │"
	// "  └─────────────┘"

	// Check corners
	if term.CellAt(2, 1).Rune != '┌' {
		t.Errorf("TopLeft = %q, want '┌'", term.CellAt(2, 1).Rune)
	}
	if term.CellAt(16, 1).Rune != '┐' {
		t.Errorf("TopRight = %q, want '┐'", term.CellAt(16, 1).Rune)
	}
	if term.CellAt(2, 4).Rune != '└' {
		t.Errorf("BottomLeft = %q, want '└'", term.CellAt(2, 4).Rune)
	}
	if term.CellAt(16, 4).Rune != '┘' {
		t.Errorf("BottomRight = %q, want '┘'", term.CellAt(16, 4).Rune)
	}

	// Check that "Title" appears in the top border
	output := term.StringTrimmed()
	if !strings.Contains(output, "Title") {
		t.Errorf("Output should contain 'Title', got:\n%s", output)
	}
}

func TestIntegration_StyledTextInBox(t *testing.T) {
	buf := NewBuffer(25, 7)
	term := NewMockTerminal(25, 7)

	// Draw a box
	boxStyle := NewStyle().Foreground(Blue)
	DrawBox(buf, NewRect(1, 1, 20, 5), BorderRounded, boxStyle)

	// Draw styled text inside
	textStyle := NewStyle().Bold().Foreground(Red)
	buf.SetString(3, 3, "Hello, World!", textStyle)

	// Render
	Render(term, buf)

	// Verify box corners (rounded)
	if term.CellAt(1, 1).Rune != '╭' {
		t.Errorf("TopLeft = %q, want '╭'", term.CellAt(1, 1).Rune)
	}
	if term.CellAt(20, 1).Rune != '╮' {
		t.Errorf("TopRight = %q, want '╮'", term.CellAt(20, 1).Rune)
	}

	// Verify text
	expected := "Hello, World!"
	for i, r := range expected {
		cell := term.CellAt(3+i, 3)
		if cell.Rune != r {
			t.Errorf("Text char %d = %q, want %q", i, cell.Rune, r)
		}
		if !cell.Style.HasAttr(AttrBold) {
			t.Errorf("Text char %d should be bold", i)
		}
	}

	// Verify box style
	cornerCell := term.CellAt(1, 1)
	if !cornerCell.Style.Fg.Equal(Blue) {
		t.Error("Box should have blue foreground")
	}
}

func TestIntegration_WideCharacters(t *testing.T) {
	buf := NewBuffer(20, 3)
	term := NewMockTerminal(20, 3)
	style := NewStyle()

	// Write wide characters
	buf.SetString(2, 1, "你好世界", style)

	// Render
	Render(term, buf)

	// Verify wide characters
	// "你好世界" = 4 wide chars = 8 columns
	if term.CellAt(2, 1).Rune != '你' {
		t.Errorf("Position 2 = %q, want '你'", term.CellAt(2, 1).Rune)
	}
	if !term.CellAt(3, 1).IsContinuation() {
		t.Error("Position 3 should be continuation")
	}
	if term.CellAt(4, 1).Rune != '好' {
		t.Errorf("Position 4 = %q, want '好'", term.CellAt(4, 1).Rune)
	}
	if !term.CellAt(5, 1).IsContinuation() {
		t.Error("Position 5 should be continuation")
	}
	if term.CellAt(6, 1).Rune != '世' {
		t.Errorf("Position 6 = %q, want '世'", term.CellAt(6, 1).Rune)
	}
	if term.CellAt(8, 1).Rune != '界' {
		t.Errorf("Position 8 = %q, want '界'", term.CellAt(8, 1).Rune)
	}
}

func TestIntegration_ResizeAndRerender(t *testing.T) {
	buf := NewBuffer(10, 5)
	term := NewMockTerminal(10, 5)
	style := NewStyle()

	// Initial content
	buf.SetString(0, 0, "Hello", style)
	Render(term, buf)

	// Verify initial render
	if term.CellAt(0, 0).Rune != 'H' {
		t.Fatal("Initial render failed")
	}

	// Resize buffer and terminal
	buf.Resize(15, 8)
	term.Resize(15, 8)

	// Add new content
	buf.SetString(0, 0, "Hello", style) // Re-add (resize clears diff state)
	buf.SetString(10, 6, "New!", style)

	// Re-render
	Render(term, buf)

	// Verify both old and new content
	if term.CellAt(0, 0).Rune != 'H' {
		t.Error("Original content should be preserved")
	}
	if term.CellAt(10, 6).Rune != 'N' {
		t.Error("New content should be visible")
	}
}

func TestIntegration_DiffMinimalChanges(t *testing.T) {
	buf := NewBuffer(10, 5)
	term := NewMockTerminal(10, 5)
	style := NewStyle()

	// Initial render
	buf.SetString(0, 0, "AAAA", style)
	Render(term, buf)

	// Make a small change
	buf.SetRune(1, 0, 'B', style)

	// Get diff
	changes := buf.Diff()

	// Should only have one change
	if len(changes) != 1 {
		t.Errorf("Diff() returned %d changes, want 1", len(changes))
	}
	if len(changes) > 0 && changes[0].Cell.Rune != 'B' {
		t.Errorf("Change rune = %q, want 'B'", changes[0].Cell.Rune)
	}
}

func TestIntegration_CapabilityColorFallback(t *testing.T) {
	// Test that color fallback works in the pipeline
	caps256 := Capabilities{Colors: Color256, TrueColor: false}
	capsNone := Capabilities{Colors: ColorNone}

	// RGB color
	rgbColor := RGBColor(255, 0, 0)

	// With 256 colors, should fall back to ANSI
	effective256 := caps256.EffectiveColor(rgbColor)
	if effective256.Type() != ColorANSI {
		t.Errorf("With Color256, RGB should fall back to ANSI, got %v", effective256.Type())
	}

	// With no colors, should fall back to default
	effectiveNone := capsNone.EffectiveColor(rgbColor)
	if effectiveNone.Type() != ColorDefault {
		t.Errorf("With ColorNone, RGB should fall back to Default, got %v", effectiveNone.Type())
	}
}

func TestIntegration_RenderFull(t *testing.T) {
	buf := NewBuffer(10, 5)
	term := NewMockTerminal(10, 5)
	style := NewStyle()

	// Set some content
	buf.SetString(2, 2, "Test", style)

	// Use RenderFull instead of Render
	RenderFull(term, buf)

	// Verify content is present
	if term.CellAt(2, 2).Rune != 'T' {
		t.Error("RenderFull should render all content")
	}
	if term.CellAt(3, 2).Rune != 'e' {
		t.Error("RenderFull should render all content")
	}
}

func TestIntegration_ClearAndRedraw(t *testing.T) {
	buf := NewBuffer(10, 5)
	term := NewMockTerminal(10, 5)
	style := NewStyle()

	// Draw initial content
	buf.SetString(0, 0, "Hello", style)
	Render(term, buf)

	// Clear and redraw different content
	buf.Clear()
	buf.SetString(0, 0, "World", style)
	Render(term, buf)

	// Verify new content
	if term.CellAt(0, 0).Rune != 'W' {
		t.Errorf("After clear, Cell(0, 0) = %q, want 'W'", term.CellAt(0, 0).Rune)
	}

	// Verify old content is gone (position 4 was 'o' in "Hello", now 'd' in "World")
	if term.CellAt(4, 0).Rune != 'd' {
		t.Errorf("After clear, Cell(4, 0) = %q, want 'd'", term.CellAt(4, 0).Rune)
	}
}

func TestIntegration_MultipleBorders(t *testing.T) {
	buf := NewBuffer(30, 10)
	term := NewMockTerminal(30, 10)
	style := NewStyle()

	// Draw multiple boxes with different styles
	DrawBox(buf, NewRect(0, 0, 10, 5), BorderSingle, style)
	DrawBox(buf, NewRect(12, 0, 10, 5), BorderDouble, style)
	DrawBox(buf, NewRect(0, 5, 10, 5), BorderRounded, style)
	DrawBox(buf, NewRect(12, 5, 10, 5), BorderThick, style)

	Render(term, buf)

	// Verify each box type
	// Single
	if term.CellAt(0, 0).Rune != '┌' {
		t.Errorf("Single top-left = %q, want '┌'", term.CellAt(0, 0).Rune)
	}
	// Double
	if term.CellAt(12, 0).Rune != '╔' {
		t.Errorf("Double top-left = %q, want '╔'", term.CellAt(12, 0).Rune)
	}
	// Rounded
	if term.CellAt(0, 5).Rune != '╭' {
		t.Errorf("Rounded top-left = %q, want '╭'", term.CellAt(0, 5).Rune)
	}
	// Thick
	if term.CellAt(12, 5).Rune != '┏' {
		t.Errorf("Thick top-left = %q, want '┏'", term.CellAt(12, 5).Rune)
	}
}

func TestIntegration_NestedBoxes(t *testing.T) {
	buf := NewBuffer(20, 10)
	term := NewMockTerminal(20, 10)
	style := NewStyle()

	// Draw outer box
	DrawBox(buf, NewRect(0, 0, 18, 8), BorderDouble, style)

	// Draw inner box
	DrawBox(buf, NewRect(2, 1, 14, 6), BorderSingle, style)

	// Add text in inner box
	buf.SetString(4, 4, "Content", style)

	Render(term, buf)

	// Verify outer box
	if term.CellAt(0, 0).Rune != '╔' {
		t.Errorf("Outer top-left = %q, want '╔'", term.CellAt(0, 0).Rune)
	}

	// Verify inner box
	if term.CellAt(2, 1).Rune != '┌' {
		t.Errorf("Inner top-left = %q, want '┌'", term.CellAt(2, 1).Rune)
	}

	// Verify content
	if term.CellAt(4, 4).Rune != 'C' {
		t.Errorf("Content = %q, want 'C'", term.CellAt(4, 4).Rune)
	}
}

func TestIntegration_TerminalStateManagement(t *testing.T) {
	term := NewMockTerminal(20, 10)

	// Test alt screen
	if term.IsInAltScreen() {
		t.Error("Should not be in alt screen initially")
	}
	term.EnterAltScreen()
	if !term.IsInAltScreen() {
		t.Error("Should be in alt screen after EnterAltScreen")
	}
	term.ExitAltScreen()
	if term.IsInAltScreen() {
		t.Error("Should not be in alt screen after ExitAltScreen")
	}

	// Test cursor visibility
	if term.IsCursorHidden() {
		t.Error("Cursor should be visible initially")
	}
	term.HideCursor()
	if !term.IsCursorHidden() {
		t.Error("Cursor should be hidden after HideCursor")
	}
	term.ShowCursor()
	if term.IsCursorHidden() {
		t.Error("Cursor should be visible after ShowCursor")
	}

	// Test raw mode
	if term.IsInRawMode() {
		t.Error("Should not be in raw mode initially")
	}
	if err := term.EnterRawMode(); err != nil {
		t.Errorf("EnterRawMode failed: %v", err)
	}
	if !term.IsInRawMode() {
		t.Error("Should be in raw mode after EnterRawMode")
	}
	if err := term.ExitRawMode(); err != nil {
		t.Errorf("ExitRawMode failed: %v", err)
	}
	if term.IsInRawMode() {
		t.Error("Should not be in raw mode after ExitRawMode")
	}
}

func TestIntegration_SnapshotTest(t *testing.T) {
	buf := NewBuffer(22, 7)
	term := NewMockTerminal(22, 7)
	style := NewStyle()

	// Create a simple UI
	DrawBoxWithTitle(buf, NewRect(1, 1, 20, 5), BorderSingle, "Demo", style)
	buf.SetString(3, 3, "Hello, World!", style)

	Render(term, buf)

	// Get string representation
	output := term.StringTrimmed()

	// Verify structure (not exact whitespace, just key elements)
	lines := strings.Split(output, "\n")

	if len(lines) < 7 {
		t.Fatalf("Expected at least 7 lines, got %d", len(lines))
	}

	// Check top border contains title
	if !strings.Contains(lines[1], "Demo") {
		t.Errorf("Line 1 should contain 'Demo', got: %q", lines[1])
	}

	// Check content line
	if !strings.Contains(lines[3], "Hello, World!") {
		t.Errorf("Line 3 should contain 'Hello, World!', got: %q", lines[3])
	}

	// Check bottom border exists
	if !strings.Contains(lines[5], "└") || !strings.Contains(lines[5], "┘") {
		t.Errorf("Line 5 should contain bottom corners, got: %q", lines[5])
	}
}
