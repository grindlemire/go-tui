package tui

import (
	"bytes"
	"strings"
	"testing"
)

func linkChanges(url string) []CellChange {
	mk := func(x int, r rune) CellChange {
		c := NewCell(r, NewStyle())
		c.Link = url
		return CellChange{X: x, Y: 0, Cell: c}
	}
	return []CellChange{mk(0, 'a'), mk(1, 'b')}
}

func TestFlush_EmitsHyperlinkOnce(t *testing.T) {
	var out bytes.Buffer
	caps := Capabilities{Colors: Color16, Hyperlinks: true}
	term := NewANSITerminalWithCaps(&out, nil, caps)
	term.Flush(linkChanges("https://example.com"))
	s := out.String()
	if strings.Count(s, "\x1b]8;;https://example.com\x1b\\") != 1 {
		t.Errorf("want exactly one open seq, got: %q", s)
	}
	if strings.Count(s, "\x1b]8;;\x1b\\") != 1 {
		t.Errorf("want exactly one close seq, got: %q", s)
	}
	// Open before the text, close after it.
	if !strings.Contains(s, "\x1b\\ab") || !strings.HasSuffix(s, "\x1b]8;;\x1b\\") {
		t.Errorf("link run not wrapped correctly: %q", s)
	}
}

func TestFlush_NoHyperlinkWhenUnsupported(t *testing.T) {
	var out bytes.Buffer
	caps := Capabilities{Colors: Color16, Hyperlinks: false}
	term := NewANSITerminalWithCaps(&out, nil, caps)
	term.Flush(linkChanges("https://example.com"))
	if strings.Contains(out.String(), "]8;;") {
		t.Errorf("must not emit OSC 8 when unsupported: %q", out.String())
	}
}

func TestANSITerminal_AltScroll(t *testing.T) {
	type tc struct {
		fn       func(*ANSITerminal)
		expected string
	}

	tests := map[string]tc{
		"enable": {
			fn:       func(t *ANSITerminal) { t.EnableAltScroll() },
			expected: "\x1b[?1007h",
		},
		"disable": {
			fn:       func(t *ANSITerminal) { t.DisableAltScroll() },
			expected: "\x1b[?1007l",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer
			term := NewANSITerminalWithCaps(&out, nil, Capabilities{Colors: Color16})
			tt.fn(term)
			if out.String() != tt.expected {
				t.Errorf("got %q, want %q", out.String(), tt.expected)
			}
		})
	}
}

func TestRichTextLink_EndToEnd(t *testing.T) {
	buf := NewBuffer(12, 1)
	e := New(
		WithSize(12, 1),
		WithRichText(
			TextSpan{Text: "see "},
			TextSpan{Text: "site", Link: "https://example.com"},
		),
	)
	e.Calculate(12, 1)
	RenderTree(buf, e)

	var out bytes.Buffer
	term := NewANSITerminalWithCaps(&out, nil, Capabilities{Colors: Color16, Hyperlinks: true})
	term.Flush(buf.Diff())
	s := out.String()
	if strings.Count(s, "\x1b]8;;https://example.com\x1b\\") != 1 {
		t.Errorf("want one hyperlink open around \"site\", got: %q", s)
	}
	if strings.Count(s, "\x1b]8;;\x1b\\") != 1 {
		t.Errorf("want one hyperlink close, got: %q", s)
	}
}
