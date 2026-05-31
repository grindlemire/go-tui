package tui

import "testing"

func TestTextArea_SetText_UsesRuneCursorPosition(t *testing.T) {
	ta := NewTextArea()
	ta.BindApp(testApp)
	ta.SetText("a界")

	if got := ta.cursorPos.Get(); got != 2 {
		t.Fatalf("cursorPos = %d, want 2", got)
	}
}

func TestTextArea_Edit_MultibyteRunes(t *testing.T) {
	ta := NewTextArea()
	ta.BindApp(testApp)
	ta.SetText("a界")
	ta.cursorPos.Set(1)

	ta.insertChar(KeyEvent{Key: KeyRune, Rune: '🙂'})
	if got := ta.Text(); got != "a🙂界" {
		t.Fatalf("text after insert = %q, want %q", got, "a🙂界")
	}

	ta.backspace(KeyEvent{Key: KeyBackspace})
	if got := ta.Text(); got != "a界" {
		t.Fatalf("text after backspace = %q, want %q", got, "a界")
	}

	ta.delete(KeyEvent{Key: KeyDelete})
	if got := ta.Text(); got != "a" {
		t.Fatalf("text after delete = %q, want %q", got, "a")
	}
}

func TestTextArea_MoveRight_UsesRuneLength(t *testing.T) {
	ta := NewTextArea()
	ta.BindApp(testApp)
	ta.SetText("é界")
	ta.cursorPos.Set(0)

	ta.moveRight(KeyEvent{Key: KeyRight})
	ta.moveRight(KeyEvent{Key: KeyRight})
	ta.moveRight(KeyEvent{Key: KeyRight})

	if got := ta.cursorPos.Get(); got != 2 {
		t.Fatalf("cursorPos = %d, want 2", got)
	}
}

func TestTextArea_HideVirtualCursor_ReturnsLineUnchanged(t *testing.T) {
	ta := NewTextArea(
		WithTextAreaVirtualCursor(false),
	)
	ta.BindApp(testApp)
	ta.SetText("hello")
	ta.Focus()
	ta.cursorPos.Set(3)

	line := ta.lineWithCursor(0)
	want := "hello"
	if line != want {
		t.Fatalf("lineWithCursor(0) = %q, want %q (unchanged)", line, want)
	}
}

func TestTextArea_HideVirtualCursor_ReturnsSpaceOnEmptyLine(t *testing.T) {
	ta := NewTextArea(
		WithTextAreaVirtualCursor(false),
	)
	ta.BindApp(testApp)
	ta.SetText("")
	ta.Focus()

	line := ta.lineWithCursor(0)
	want := " "
	if line != want {
		t.Fatalf("lineWithCursor(0) on empty = %q, want %q", line, want)
	}
}

func TestTextArea_HideVirtualCursor_LineWidthUnchanged(t *testing.T) {
	ta := NewTextArea(
		WithTextAreaVirtualCursor(false),
	)
	ta.BindApp(testApp)
	ta.SetText("hello world")
	ta.Focus()

	// With virtual cursor on, line would gain an extra character.
	// With it off, line width should match the wrapped line width.
	lines := ta.wrapText()
	for i, line := range lines {
		rendered := ta.lineWithCursor(i)
		if rendered != line && rendered != " " {
			t.Fatalf("lineWithCursor(%d) = %q, want %q (or space for empty)", i, rendered, line)
		}
	}
}
