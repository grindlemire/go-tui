package tui

import (
	"bytes"
	"strings"
	"testing"
)

func TestSprint(t *testing.T) {
	type tc struct {
		el       *Element
		opts     []PrintOption
		contains []string
		empty    bool
	}

	tests := map[string]tc{
		"basic text element": {
			el:       New(WithText("hello world")),
			opts:     []PrintOption{WithPrintWidth(40)},
			contains: []string{"hello world"},
		},
		"explicit width": {
			el: func() *Element {
				parent := New(WithDirection(Row))
				parent.AddChild(New(WithText("left")))
				return parent
			}(),
			opts:     []PrintOption{WithPrintWidth(20)},
			contains: []string{"left"},
		},
		"nested with borders": {
			el:       New(WithText("boxed"), WithBorder(BorderSingle)),
			opts:     []PrintOption{WithPrintWidth(20)},
			contains: []string{"boxed", "┌", "└"},
		},
		"styled text": {
			el:       New(WithText("red"), WithTextStyle(NewStyle().Foreground(Red))),
			opts:     []PrintOption{WithPrintWidth(20)},
			contains: []string{"\x1b[", "red", "\x1b[0m"},
		},
		"empty element": {
			el:    New(),
			opts:  []PrintOption{WithPrintWidth(20)},
			empty: true,
		},
		"multiline content": {
			el: func() *Element {
				parent := New(WithDirection(Column))
				parent.AddChild(
					New(WithText("line 1")),
					New(WithText("line 2")),
				)
				return parent
			}(),
			opts:     []PrintOption{WithPrintWidth(30)},
			contains: []string{"line 1", "line 2"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := Sprint(tt.el, tt.opts...)

			if tt.empty {
				if result != "" {
					t.Fatalf("expected empty result, got %q", result)
				}
				return
			}
			if result == "" {
				t.Fatal("expected non-empty result, got empty")
			}

			for _, sub := range tt.contains {
				if !strings.Contains(result, sub) {
					t.Errorf("result does not contain %q\ngot: %q", sub, result)
				}
			}
		})
	}
}

func TestFprint(t *testing.T) {
	type tc struct {
		el       *Element
		contains []string
		empty    bool
	}

	tests := map[string]tc{
		"writes to buffer": {
			el:       New(WithText("hello")),
			contains: []string{"hello"},
		},
		"empty element writes nothing": {
			el:    New(),
			empty: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			Fprint(&buf, tt.el, WithPrintWidth(30))
			result := buf.String()

			if tt.empty {
				if result != "" {
					t.Fatalf("expected empty output, got %q", result)
				}
				return
			}
			if result == "" {
				t.Fatal("expected non-empty output, got empty")
			}

			for _, sub := range tt.contains {
				if !strings.Contains(result, sub) {
					t.Errorf("output does not contain %q\ngot: %q", sub, result)
				}
			}
		})
	}
}

func TestFprintTrailingNewline(t *testing.T) {
	var buf bytes.Buffer
	Fprint(&buf, New(WithText("hi")), WithPrintWidth(10))
	result := buf.String()
	if !strings.HasSuffix(result, "\n") {
		t.Errorf("Fprint output should end with newline, got %q", result)
	}
}

func TestSprintNoTrailingNewline(t *testing.T) {
	result := Sprint(New(WithText("hi")), WithPrintWidth(10))
	if strings.HasSuffix(result, "\n") {
		t.Errorf("Sprint output should NOT end with newline, got %q", result)
	}
}
