package tui

import "testing"

func TestBufferRowToANSI(t *testing.T) {
	type tc struct {
		name     string
		setup    func(buf *Buffer)
		row      int
		contains []string
		notEmpty bool
	}

	defaultCaps := Capabilities{Colors: ColorTrue, TrueColor: true}

	tests := map[string]tc{
		"plain text": {
			setup: func(buf *Buffer) {
				buf.SetString(0, 0, "hello", NewStyle())
			},
			row:      0,
			contains: []string{"hello"},
			notEmpty: true,
		},
		"styled text emits ANSI": {
			setup: func(buf *Buffer) {
				buf.SetString(0, 0, "red", NewStyle().Foreground(Red))
			},
			row:      0,
			contains: []string{"\x1b[", "red", "\x1b[0m"},
			notEmpty: true,
		},
		"trailing empty cells trimmed": {
			setup: func(buf *Buffer) {
				buf.SetString(0, 0, "hi", NewStyle())
				// rest of the 20-wide buffer is empty
			},
			row:      0,
			contains: []string{"hi"},
			notEmpty: true,
		},
		"wide character skips continuation": {
			setup: func(buf *Buffer) {
				// CJK character '中' is 2 cells wide
				buf.SetRune(0, 0, '中', NewStyle())
			},
			row:      0,
			contains: []string{"中"},
			notEmpty: true,
		},
		"empty row returns empty string": {
			setup:    func(buf *Buffer) {},
			row:      0,
			notEmpty: false,
		},
		"style transition mid-row": {
			setup: func(buf *Buffer) {
				buf.SetString(0, 0, "ab", NewStyle().Bold())
				buf.SetString(2, 0, "cd", NewStyle().Italic())
			},
			row: 0,
			// Both style changes must appear plus the text
			contains: []string{"ab", "cd", "\x1b[0m"},
			notEmpty: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buf := NewBuffer(20, 2)
			tt.setup(buf)

			esc := newEscBuilder(128)
			result := bufferRowToANSI(buf, tt.row, esc, defaultCaps)

			if tt.notEmpty && result == "" {
				t.Fatalf("expected non-empty result, got empty")
			}
			if !tt.notEmpty && result != "" {
				t.Fatalf("expected empty result, got %q", result)
			}

			for _, sub := range tt.contains {
				if !containsStr(result, sub) {
					t.Errorf("result %q does not contain %q", result, sub)
				}
			}
		})
	}
}

// containsStr is a simple substring check helper.
func containsStr(s, sub string) bool {
	return len(sub) <= len(s) && searchStr(s, sub)
}

func searchStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestRenderElementToBuffer(t *testing.T) {
	type tc struct {
		el         *Element
		width      int
		wantHeight int // expected height; 0 means nil buffer
	}

	defaultCaps := Capabilities{Colors: ColorTrue, TrueColor: true}

	tests := map[string]tc{
		"simple text span": {
			el:         New(WithText("hello")),
			width:      20,
			wantHeight: 1,
		},
		"text with border": {
			el:         New(WithText("hi"), WithBorder(BorderSingle)),
			width:      20,
			wantHeight: 3, // 1 text + 2 border rows
		},
		"empty element": {
			el:         New(),
			width:      20,
			wantHeight: 0,
		},
		"column with children": {
			el: func() *Element {
				parent := New(WithDirection(Column))
				parent.AddChild(
					New(WithText("line 1")),
					New(WithText("line 2")),
				)
				return parent
			}(),
			width:      20,
			wantHeight: 2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buf, height := renderElementToBuffer(tt.el, tt.width, defaultCaps)

			if height != tt.wantHeight {
				t.Fatalf("height = %d, want %d", height, tt.wantHeight)
			}
			if tt.wantHeight == 0 {
				if buf != nil {
					t.Fatalf("expected nil buffer for zero height")
				}
				return
			}
			if buf == nil {
				t.Fatalf("expected non-nil buffer for height %d", tt.wantHeight)
			}
			if buf.Width() != tt.width {
				t.Fatalf("buffer width = %d, want %d", buf.Width(), tt.width)
			}
			if buf.Height() != tt.wantHeight {
				t.Fatalf("buffer height = %d, want %d", buf.Height(), tt.wantHeight)
			}
		})
	}
}
