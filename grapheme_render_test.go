package tui

import (
	"strings"
	"testing"
)

// TestBuffer_SetString_Cluster verifies the buffer path: a multi-rune
// grapheme cluster is stored whole in one leading cell (width 2) with a single
// continuation cell, and survives round-trip through the buffer's text emission.
func TestBuffer_SetString_Cluster(t *testing.T) {
	b := NewBuffer(10, 1)
	b.SetString(0, 0, emojiFlagUS+"x", NewStyle())

	lead := b.Cell(0, 0)
	if cellGlyph(lead) != emojiFlagUS {
		t.Errorf("cell(0,0) glyph = %q, want the whole flag %q", cellGlyph(lead), emojiFlagUS)
	}
	if lead.Width != 2 {
		t.Errorf("cell(0,0).Width = %d, want 2", lead.Width)
	}
	if !b.Cell(1, 0).IsContinuation() {
		t.Errorf("cell(1,0) should be a continuation cell")
	}
	// The trailing 'x' lands at column 2, not 4: the flag consumed two columns.
	if got := b.Cell(2, 0).Rune; got != 'x' {
		t.Errorf("cell(2,0).Rune = %q, want 'x' (flag is 2 cols wide, not 4)", got)
	}

	// Emission writes the cluster's full bytes back out.
	if line := b.String(); !strings.Contains(line, emojiFlagUS) {
		t.Errorf("buffer text %q does not contain the flag cluster", line)
	}
}

// TestBuffer_SetStringGradient_Cluster verifies the gradient writer is cluster
// aware: a flag occupies one width-2 cell plus a continuation, not two cells.
func TestBuffer_SetStringGradient_Cluster(t *testing.T) {
	b := NewBuffer(10, 1)
	g := NewGradient(Red, Blue)
	w := b.SetStringGradient(0, 0, emojiFlagUS+"x", g, NewStyle())

	if w != 3 {
		t.Errorf("SetStringGradient returned width %d, want 3 (flag 2 + x 1)", w)
	}
	lead := b.Cell(0, 0)
	if cellGlyph(lead) != emojiFlagUS || lead.Width != 2 {
		t.Errorf("cell(0,0) = {%q, w%d}, want the flag at width 2", cellGlyph(lead), lead.Width)
	}
	if !b.Cell(1, 0).IsContinuation() {
		t.Errorf("cell(1,0) should be a continuation cell")
	}
	if got := b.Cell(2, 0).Rune; got != 'x' {
		t.Errorf("cell(2,0).Rune = %q, want 'x'", got)
	}
	if lead.Style.Fg.IsDefault() {
		t.Errorf("gradient color was not applied to the flag cell")
	}
}

// TestWrapText_KeepsClustersWhole verifies wrapping measures by cluster width and
// never breaks inside a cluster. This is the wrap symptom from issue #95.
func TestWrapText_KeepsClustersWhole(t *testing.T) {
	type tc struct {
		text  string
		width int
		want  []string
	}

	tests := map[string]tc{
		// flag(2) + "abcd"(4) = 6 columns: fits a width-6 line. A per-code-point
		// sum would score the flag as 4 and wrap to two lines.
		"flag plus text fits one line": {text: emojiFlagUS + "abcd", width: 6, want: []string{emojiFlagUS + "abcd"}},
		// A hard break falls between clusters, never inside the flag.
		"break between clusters": {text: "a" + emojiFlagUS + "b", width: 2, want: []string{"a", emojiFlagUS, "b"}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := wrapText(tt.text, tt.width)
			if len(got) != len(tt.want) {
				t.Fatalf("wrapText(%q, %d) = %q, want %q", tt.text, tt.width, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestRenderTree_ClusterBoxNoWrap is the issue #95 reproduction turned regression
// test: a width-8 single-border box (content width 6) holding a flag plus four
// ASCII chars renders on one content row, because the flag is two columns wide.
func TestRenderTree_ClusterBoxNoWrap(t *testing.T) {
	buf := NewBuffer(8, 5)
	box := New(
		WithWidth(8),
		WithBorder(BorderSingle),
		WithText(emojiFlagUS+"abcd"),
		WithWrap(false),
	)
	Calculate(box, 8, 3)
	RenderTree(buf, box)

	// The content area is 6 wide. A per-code-point sum would say "US flag = 4
	// columns + abcd = 4 = 8 > 6" and force a second line. Grapheme-aware: flag
	// = 2 + abcd = 4 = 6, so it fits one line.
	line := buf.StringTrimmed()
	if !strings.Contains(line, emojiFlagUS+"abcd") {
		t.Errorf("box content %q should contain the full text on one line", line)
	}
}

func TestCellGlyph(t *testing.T) {
	// Single-rune cell
	c1 := NewCell('a', NewStyle())
	if g := cellGlyph(c1); g != "a" {
		t.Errorf("cellGlyph(single rune) = %q, want %q", g, "a")
	}

	// Multi-rune cluster cell
	c2 := newClusterCell(emojiFlagUS, 2, NewStyle(), "")
	g2 := cellGlyph(c2)
	if g2 != emojiFlagUS {
		t.Errorf("cellGlyph(cluster) = %q, want %q", g2, emojiFlagUS)
	}

	// Empty cell
	c3 := Cell{}
	if g3 := cellGlyph(c3); g3 != "" {
		t.Errorf("cellGlyph(empty) = %q, want empty string", g3)
	}
}

// TestWrapInlineStyledRows_ASCII verifies the ANSI-styled wrapping path
// handles plain ASCII text correctly (no ANSI sequences).
func TestWrapInlineStyledRows_ASCII(t *testing.T) {
	type tc struct {
		name  string
		text  string
		width int
		want  []string
	}

	tests := []tc{
		{name: "empty", text: "", width: 10, want: []string{""}},
		{name: "fits on one line", text: "hello", width: 10, want: []string{"hello"}},
		// wrapInlineStyledRows wraps by visual column, not word boundary.
		// The space is written before the flush check triggers, so it
		// becomes the first character of the next line.
		{name: "wraps at boundary", text: "hello world", width: 5, want: []string{"hello", " worl", "d"}},
		{name: "newline breaks", text: "ab\ncd", width: 10, want: []string{"ab", "cd"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapInlineStyledRows(tt.text, tt.width)
			if len(got) != len(tt.want) {
				t.Fatalf("wrapInlineStyledRows = %q, want %q", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("row[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestWrapInlineStyledRows_ANSI verifies ANSI sequences pass through unchanged
// and don't count toward column width.
func TestWrapInlineStyledRows_ANSI(t *testing.T) {
	type tc struct {
		name  string
		text  string
		width int
		want  []string
	}

	tests := []tc{
		// ANSI around plain text
		{name: "ansi bold", text: "\x1b[1mhello\x1b[0m", width: 10, want: []string{"\x1b[1mhello\x1b[0m"}},
		// wrapInlineStyledRows wraps by visual column, not word boundary.
		// The space is written before the flush check fires.
		{name: "ansi wraps", text: "\x1b[31mhello world\x1b[0m", width: 5, want: []string{"\x1b[31mhello", " worl", "d\x1b[0m"}},
		// ANSI at wrap boundary
		{name: "ansi at boundary", text: "ab\x1b[1mcdef\x1b[0m", width: 4, want: []string{"ab\x1b[1mcd", "ef\x1b[0m"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapInlineStyledRows(tt.text, tt.width)
			if len(got) != len(tt.want) {
				t.Fatalf("wrapInlineStyledRows = %q, want %q", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("row[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestWrapInlineStyledRows_Cluster verifies that multi-rune grapheme clusters
// (flags, decomposed accents, CJK) are not split by the ANSI-styled path.
func TestWrapInlineStyledRows_Cluster(t *testing.T) {
	type tc struct {
		name  string
		text  string
		width int
		want  []string
	}

	tests := []tc{
		// Flag (2 cols) + ascii: fits one line at width 6
		{name: "flag plus ascii", text: emojiFlagUS + "abcd", width: 6, want: []string{emojiFlagUS + "abcd"}},
		// Flag wraps as atomic unit
		{name: "flag wraps atomic", text: "a" + emojiFlagUS + "b", width: 2, want: []string{"a", emojiFlagUS, "b"}},
		// CJK (2 cols each)
		{name: "cjk wraps", text: "你好世界", width: 4, want: []string{"你好", "世界"}},
		// ANSI + flag (realistic: styled flag emoji)
		{name: "ansi flag", text: "\x1b[33m" + emojiFlagUS + "\x1b[0m", width: 4, want: []string{"\x1b[33m" + emojiFlagUS + "\x1b[0m"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapInlineStyledRows(tt.text, tt.width)
			if len(got) != len(tt.want) {
				t.Fatalf("wrapInlineStyledRows = %q, want %q", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("row[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestWrapInlineStyledRows_Oversized verifies that a cluster wider than the
// entire line is replaced with "?" without exceeding the width.
func TestWrapInlineStyledRows_Oversized(t *testing.T) {
	type tc struct {
		name  string
		text  string
		width int
		want  []string
	}

	tests := []tc{
		// Single CJK at width 1: too wide, replaced with ?
		{name: "cjk at width 1", text: "你", width: 1, want: []string{"?"}},
		// CJK at width 2: fits exactly
		{name: "cjk at width 2", text: "你", width: 2, want: []string{"你"}},
		// Flag at width 1: replaced with ?
		{name: "flag at width 1", text: emojiFlagUS, width: 1, want: []string{"?"}},
		// ASCII 'a' then CJK at width 1. 'a' fills line, CJK too wide → ? on new line
		{name: "ascii then cjk at width 1", text: "a你", width: 1, want: []string{"a", "?"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapInlineStyledRows(tt.text, tt.width)
			if len(got) != len(tt.want) {
				t.Fatalf("wrapInlineStyledRows = %q, want %q", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("row[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
