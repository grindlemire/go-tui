package tui

import (
	"strings"
	"testing"
)

// TestBuffer_SetString_Cluster verifies the Tier-2 buffer path: a multi-rune
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
	)
	box.Calculate(8, 5)
	RenderTree(buf, box)

	// Content row 1 holds the whole string; content row 2 must be blank (no wrap).
	row1 := contentRow(buf, 1)
	if !strings.Contains(row1, emojiFlagUS) || !strings.Contains(row1, "abcd") {
		t.Fatalf("content row 1 = %q, want it to hold the flag and \"abcd\" on one line", row1)
	}
	if row2 := strings.TrimSpace(contentRow(buf, 2)); row2 != "" {
		t.Errorf("content row 2 = %q, want empty (text must not wrap)", row2)
	}
}

// TestTextArea_BlockCursorKeepsClusterWhole guards the block-cursor overlay used
// when a display-full line ends in a hard newline: it must drop the whole final
// grapheme cluster for the cursor, not just the last rune (which would split a
// flag / ZWJ family / decomposed accent).
func TestTextArea_BlockCursorKeepsClusterWhole(t *testing.T) {
	ta := NewTextArea(WithTextAreaWidth(2))
	ta.BindApp(testApp)
	ta.SetText(emojiFlagUS + "\nx") // flag fills the width-2 row, then a hard break
	ta.Focus()
	ta.cursorPos.Set(2) // end of row 0 (after the flag, before '\n')
	ta.blink.Set(true)

	got := ta.lineWithCursor(0)
	if strings.ContainsRune(got, '\U0001F1FA') || strings.ContainsRune(got, '\U0001F1F8') {
		t.Fatalf("lineWithCursor(0) = %q, want the flag cluster dropped whole (no lone regional indicator)", got)
	}
	if got != string(ta.cursorRune) {
		t.Errorf("lineWithCursor(0) = %q, want %q", got, string(ta.cursorRune))
	}
}

// contentRow returns the interior of a single-border box row as text (border
// columns and continuation cells skipped).
func contentRow(b *Buffer, y int) string {
	var sb strings.Builder
	for x := 1; x < b.Width()-1; x++ {
		cell := b.Cell(x, y)
		if cell.IsContinuation() {
			continue
		}
		r := cell.Rune
		if r == 0 {
			r = ' '
		}
		sb.WriteRune(r)
		if cell.Combining != "" {
			sb.WriteString(cell.Combining)
		}
	}
	return sb.String()
}
