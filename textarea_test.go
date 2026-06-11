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

	// width 10, "一二三四五六七八九十" wraps to ["一二三四五", "六七八九十"].
	// Both wrapped lines are display-full, so boundary positions move to the
	// next visual line (downstream affinity); end of text lands on a phantom
	// row past the last line.
	tests := map[string]tc{
		"start of text":              {pos: 0, wantRow: 0, wantCol: 0},
		"soft wrap boundary":         {pos: 5, wantRow: 1, wantCol: 0},
		"after first rune of second": {pos: 6, wantRow: 1, wantCol: 1},
		"end of text":                {pos: 10, wantRow: 2, wantCol: 0},
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

func TestTextArea_CursorRowCol_WrapBoundaryAffinity(t *testing.T) {
	type tc struct {
		text    string
		width   int
		pos     int
		wantRow int
		wantCol int
	}

	tests := map[string]tc{
		"soft boundary on full line moves to next line start": {
			text: "abcdefgh", width: 4, pos: 4, wantRow: 1, wantCol: 0,
		},
		"soft boundary on non-full line stays at line end": {
			// "ab界" is 4 columns in width 5, so the cursor still fits there
			text: "ab界界", width: 5, pos: 3, wantRow: 0, wantCol: 3,
		},
		"hard newline after full line stays at line end": {
			text: "abcd\nef", width: 4, pos: 4, wantRow: 0, wantCol: 4,
		},
		"end of text on full last line moves to phantom row": {
			text: "abcd", width: 4, pos: 4, wantRow: 1, wantCol: 0,
		},
		"end of text on full cjk line moves to phantom row": {
			text: "一二", width: 4, pos: 2, wantRow: 1, wantCol: 0,
		},
		"end of text on non-full line stays at line end": {
			text: "abc", width: 4, pos: 3, wantRow: 0, wantCol: 3,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ta := NewTextArea(WithTextAreaWidth(tt.width))
			ta.BindApp(testApp)
			ta.SetText(tt.text)
			ta.cursorPos.Set(tt.pos)

			row, col := ta.cursorRowCol(ta.wrapText())
			if row != tt.wantRow || col != tt.wantCol {
				t.Fatalf("cursorRowCol() = (%d, %d), want (%d, %d)", row, col, tt.wantRow, tt.wantCol)
			}
		})
	}
}

func TestTextArea_Move_WrapBoundary(t *testing.T) {
	type tc struct {
		text    string
		width   int
		pos     int
		move    func(*TextArea)
		wantPos int
	}

	tests := map[string]tc{
		// posFromRowCol could not resolve (row, 0) targets on soft-wrapped
		// lines, so moving down from column 0 jumped past the target line.
		"down from column zero crosses soft wrap": {
			text: "abcdef", width: 4, pos: 0,
			move:    func(ta *TextArea) { ta.moveDown(KeyEvent{Key: KeyDown}) },
			wantPos: 4,
		},
		// The cursor at a full-line soft boundary displays on the next line,
		// so moving up from there lands on the line above it.
		"up from wrap boundary lands on previous line": {
			text: "abcdef", width: 4, pos: 4,
			move:    func(ta *TextArea) { ta.moveUp(KeyEvent{Key: KeyUp}) },
			wantPos: 0,
		},
		// End with the cursor on the phantom row resolves against the last
		// real line instead of indexing past the lines slice.
		"end on phantom row stays at end of text": {
			text: "abcd", width: 4, pos: 4,
			move:    func(ta *TextArea) { ta.moveEnd(KeyEvent{Key: KeyEnd}) },
			wantPos: 4,
		},
		"home on phantom row stays at boundary": {
			text: "abcd", width: 4, pos: 4,
			move:    func(ta *TextArea) { ta.moveHome(KeyEvent{Key: KeyHome}) },
			wantPos: 4,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ta := NewTextArea(WithTextAreaWidth(tt.width))
			ta.BindApp(testApp)
			ta.SetText(tt.text)
			ta.cursorPos.Set(tt.pos)

			tt.move(ta)
			if got := ta.cursorPos.Get(); got != tt.wantPos {
				t.Fatalf("cursorPos after move = %d, want %d", got, tt.wantPos)
			}
		})
	}
}

func TestTextArea_CursorVisibleAtWrapBoundary(t *testing.T) {
	type tc struct {
		text       string
		width      int
		pos        int
		wantHeight int
	}

	tests := map[string]tc{
		"soft boundary on full line":      {text: "abcdefgh", width: 4, pos: 4, wantHeight: 2},
		"end of text on full last line":   {text: "abcd", width: 4, pos: 4, wantHeight: 2},
		"end of text on full cjk line":    {text: "一二三四五", width: 10, pos: 5, wantHeight: 2},
		"end of text on non-full line":    {text: "abc", width: 4, pos: 3, wantHeight: 1},
		"end of text after hard newline":  {text: "abcd\n", width: 4, pos: 5, wantHeight: 2},
		"mid-line cursor away from edges": {text: "abcdef", width: 4, pos: 2, wantHeight: 2},
		// The cursor overlays the last cell when a display-full line ends in
		// a hard newline, since there is no continuation line to move it to.
		"full line before hard newline": {text: "abcd\nef", width: 4, pos: 4, wantHeight: 2},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ta := NewTextArea(WithTextAreaWidth(tt.width))
			ta.BindApp(testApp)
			ta.SetText(tt.text)
			ta.Focus()
			ta.cursorPos.Set(tt.pos)
			ta.blink.Set(true)

			rows := renderedRows(t, ta, tt.width)
			if len(rows) != tt.wantHeight {
				t.Fatalf("rendered height = %d, want %d\n%s", len(rows), tt.wantHeight, strings.Join(rows, "\n"))
			}
			if !strings.ContainsRune(strings.Join(rows, "\n"), ta.cursorRune) {
				t.Fatalf("cursor not visible in rendered output:\n%s", strings.Join(rows, "\n"))
			}
		})
	}
}

func TestTextArea_Height_PhantomCursorRow(t *testing.T) {
	ta := NewTextArea(WithTextAreaWidth(4))
	ta.BindApp(testApp)
	ta.SetText("abcd")

	if got := ta.Height(); got != 1 {
		t.Fatalf("unfocused Height() = %d, want 1", got)
	}
	ta.Focus()
	if got := ta.Height(); got != 2 {
		t.Fatalf("focused Height() = %d, want 2", got)
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
