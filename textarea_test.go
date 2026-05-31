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

func TestTextArea_HideVirtualCursor(t *testing.T) {
	type tc struct {
		text    string
		cursor  int
		wantLen int
	}

	tests := map[string]tc{
		"returns line unchanged":            {text: "hello", cursor: 3, wantLen: 5},
		"returns space on empty line":       {text: "", cursor: 0, wantLen: 1},
		"line width matches wrapped text":   {text: "hello world", cursor: 5, wantLen: 11},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ta := NewTextArea(
				WithTextAreaVirtualCursor(false),
			)
			ta.BindApp(testApp)
			ta.SetText(tt.text)
			ta.Focus()
			ta.cursorPos.Set(tt.cursor)

			lines := ta.wrapText()
			for i := range lines {
				rendered := ta.lineWithCursor(i)
				if len(rendered) != tt.wantLen {
					t.Fatalf("lineWithCursor(%d) = %q (len=%d), want len=%d", i, rendered, len(rendered), tt.wantLen)
				}
			}
		})
	}
}
