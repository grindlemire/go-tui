package tui

import (
	"strings"
	"testing"
	"unicode/utf8"
)

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

func TestTextArea_WrapText_DisplayWidth(t *testing.T) {
	type tc struct {
		width  int
		border BorderStyle
		text   string
		want   []string
	}

	tests := map[string]tc{
		"ascii wraps at width": {
			width: 10,
			text:  "abcdefghijklmnop",
			want:  []string{"abcdefghij", "klmnop"},
		},
		"cjk wraps at display columns not rune count": {
			width: 10,
			text:  "一二三四五六七八九十",
			want:  []string{"一二三四五", "六七八九十"},
		},
		"mixed ascii and cjk": {
			width: 10,
			text:  "ab界cd界ef界",
			want:  []string{"ab界cd界ef", "界"},
		},
		"wide char that does not fit moves to next line": {
			width: 5,
			text:  "ab界界",
			want:  []string{"ab界", "界"},
		},
		"border reduces wrap width by two": {
			width:  10,
			border: BorderSingle,
			text:   "abcdefghijklmnop",
			want:   []string{"abcdefgh", "ijklmnop"},
		},
		"embedded newlines preserved": {
			width: 10,
			text:  "一二三\n\nab",
			want:  []string{"一二三", "", "ab"},
		},
		"emoji wraps at display columns": {
			width: 5,
			text:  "🎉🎉🎉",
			want:  []string{"🎉🎉", "🎉"},
		},
		"rune wider than wrap width gets its own line": {
			width:  3,
			border: BorderSingle,
			text:   "界界",
			want:   []string{"界", "界"},
		},
		"zero width disables wrapping": {
			width: 0,
			text:  "abcdef\ngh",
			want:  []string{"abcdef", "gh"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			opts := []TextAreaOption{WithTextAreaWidth(tt.width)}
			if tt.border != BorderNone {
				opts = append(opts, WithTextAreaBorder(tt.border))
			}
			ta := NewTextArea(opts...)
			ta.BindApp(testApp)
			ta.SetText(tt.text)

			lines := ta.wrapText()
			if len(lines) != len(tt.want) {
				t.Fatalf("wrapText() = %q, want %q", lines, tt.want)
			}
			for i := range lines {
				if lines[i] != tt.want[i] {
					t.Fatalf("wrapText()[%d] = %q, want %q", i, lines[i], tt.want[i])
				}
			}
		})
	}
}

func TestTextArea_CursorRowCol_WideChars(t *testing.T) {
	type tc struct {
		pos     int
		wantRow int
		wantCol int
	}

	// width 10, "一二三四五六七八九十" wraps to ["一二三四五", "六七八九十"]
	tests := map[string]tc{
		"start of text":              {pos: 0, wantRow: 0, wantCol: 0},
		"end of first wrapped line":  {pos: 5, wantRow: 0, wantCol: 5},
		"after first rune of second": {pos: 6, wantRow: 1, wantCol: 1},
		"end of text":                {pos: 10, wantRow: 1, wantCol: 5},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ta := NewTextArea(WithTextAreaWidth(10))
			ta.BindApp(testApp)
			ta.SetText("一二三四五六七八九十")
			ta.cursorPos.Set(tt.pos)

			row, col := ta.cursorRowCol(ta.wrapText())
			if row != tt.wantRow || col != tt.wantCol {
				t.Fatalf("cursorRowCol() = (%d, %d), want (%d, %d)", row, col, tt.wantRow, tt.wantCol)
			}
		})
	}
}

// renderedRows renders a component into a standalone buffer and returns each
// row as a plain string (continuation cells skipped, empty cells as spaces).
func renderedRows(t *testing.T, c Component, width int) []string {
	t.Helper()
	buf, height := renderElementToBuffer(c.Render(testApp), width, Capabilities{})
	if buf == nil {
		t.Fatal("renderElementToBuffer returned nil buffer")
	}
	rows := make([]string, 0, height)
	for y := range height {
		var sb strings.Builder
		for x := range width {
			cell := buf.Cell(x, y)
			if cell.IsContinuation() {
				continue
			}
			if cell.Rune == 0 {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(cell.Rune)
			}
		}
		rows = append(rows, sb.String())
	}
	return rows
}

func TestTextArea_Render_NoClippedContent(t *testing.T) {
	type tc struct {
		width  int
		border BorderStyle
		text   string
	}

	tests := map[string]tc{
		"cjk without border": {width: 10, text: "一二三四五六七八九十"},
		"cjk with border":    {width: 10, border: BorderSingle, text: "一二三四五六七八"},
		"ascii with border":  {width: 10, border: BorderSingle, text: "abcdefghijklmnop"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			opts := []TextAreaOption{WithTextAreaWidth(tt.width)}
			if tt.border != BorderNone {
				opts = append(opts, WithTextAreaBorder(tt.border))
			}
			ta := NewTextArea(opts...)
			ta.BindApp(testApp)
			ta.SetText(tt.text)

			rendered := strings.Join(renderedRows(t, ta, tt.width), "\n")
			for _, r := range tt.text {
				if !strings.ContainsRune(rendered, r) {
					t.Errorf("rune %q clipped out of rendered output:\n%s", r, rendered)
				}
			}
		})
	}
}

func TestTextArea_HideVirtualCursor(t *testing.T) {
	type tc struct {
		text    string
		cursor  int
		wantLen int
	}

	tests := map[string]tc{
		"returns line unchanged":          {text: "hello", cursor: 3, wantLen: 5},
		"returns space on empty line":     {text: "", cursor: 0, wantLen: 1},
		"line width matches wrapped text": {text: "hello world", cursor: 5, wantLen: 11},
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
				if utf8.RuneCountInString(rendered) != tt.wantLen {
					t.Fatalf("lineWithCursor(%d) = %q (len=%d), want len=%d", i, rendered, utf8.RuneCountInString(rendered), tt.wantLen)
				}
			}
		})
	}
}
