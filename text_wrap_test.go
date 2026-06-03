package tui

import (
	"strings"
	"testing"
)

func TestWrapText(t *testing.T) {
	type tc struct {
		text     string
		maxWidth int
		want     []string
	}

	tests := map[string]tc{
		"empty string": {
			text:     "",
			maxWidth: 10,
			want:     []string{""},
		},
		"fits on one line": {
			text:     "hello",
			maxWidth: 10,
			want:     []string{"hello"},
		},
		"wraps at word boundary": {
			text:     "hello world",
			maxWidth: 7,
			want:     []string{"hello", "world"},
		},
		"multiple wraps": {
			text:     "the quick brown fox",
			maxWidth: 10,
			want:     []string{"the quick", "brown fox"},
		},
		"long word breaks mid-word": {
			text:     "abcdefghij",
			maxWidth: 5,
			want:     []string{"abcde", "fghij"},
		},
		"long word after short word": {
			text:     "hi abcdefghij",
			maxWidth: 5,
			want:     []string{"hi", "abcde", "fghij"},
		},
		"preserves newlines": {
			text:     "line1\nline2",
			maxWidth: 20,
			want:     []string{"line1", "line2"},
		},
		"wraps within newline sections": {
			text:     "hello world\nfoo bar",
			maxWidth: 7,
			want:     []string{"hello", "world", "foo bar"},
		},
		"zero width": {
			text:     "hello",
			maxWidth: 0,
			want:     []string{""},
		},
		"width of 1": {
			text:     "hi",
			maxWidth: 1,
			want:     []string{"h", "i"},
		},
		"exact fit": {
			text:     "hello",
			maxWidth: 5,
			want:     []string{"hello"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := wrapText(tt.text, tt.maxWidth)
			if len(got) != len(tt.want) {
				t.Fatalf("wrapText(%q, %d) = %v (len %d), want %v (len %d)",
					tt.text, tt.maxWidth, got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line %d: got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestWrapSpans_KeepsStyleAcrossWrap(t *testing.T) {
	bold := NewStyle().Bold()
	spans := []TextSpan{
		{Text: "aa "},
		{Text: "bbbb cccc", Style: bold}, // two bold words
		{Text: " dd"},
	}
	// Width 6 forces a break inside the bold run.
	lines := wrapSpans(spans, 6)
	if len(lines) < 2 {
		t.Fatalf("expected multiple lines, got %d: %+v", len(lines), lines)
	}
	// Every segment whose text is a bold word must carry bold on every line.
	for li, line := range lines {
		for _, seg := range line {
			if seg.Text == "bbbb" || seg.Text == "cccc" {
				if seg.Style.Attrs&AttrBold == 0 {
					t.Errorf("line %d: %q lost bold", li, seg.Text)
				}
			}
		}
	}
}

func TestWrapSpans_MergesAdjacentSameStyle(t *testing.T) {
	// Two plain spans with words that fit on one line should merge.
	spans := []TextSpan{{Text: "foo "}, {Text: "bar"}}
	lines := wrapSpans(spans, 40)
	if len(lines) != 1 {
		t.Fatalf("want 1 line, got %d", len(lines))
	}
	if len(lines[0]) != 1 {
		t.Errorf("adjacent same-style segments should merge into 1, got %d: %+v", len(lines[0]), lines[0])
	}
	if lines[0][0].Text != "foo bar" {
		t.Errorf("merged text = %q, want \"foo bar\"", lines[0][0].Text)
	}
}

func TestWrapSpans_Empty(t *testing.T) {
	if got := wrapSpans(nil, 10); len(got) != 1 || len(got[0]) != 0 {
		t.Errorf("empty spans should give one empty line, got %+v", got)
	}
}

// joinSpanLine concatenates a wrapped line's segment texts into one string.
func joinSpanLine(line []TextSpan) string {
	var s strings.Builder
	for _, seg := range line {
		s.WriteString(seg.Text)
	}
	return s.String()
}

// hasControlRune reports whether any segment text contains a control rune that
// should never survive wrapping (CR, VT, FF, or a literal newline).
func hasControlRune(lines [][]TextSpan) bool {
	for _, line := range lines {
		for _, seg := range line {
			for _, r := range seg.Text {
				switch r {
				case '\r', '\v', '\f', '\n':
					return true
				}
			}
		}
	}
	return false
}

func TestWrapSpans_ExoticWhitespaceDoesNotLeak(t *testing.T) {
	// '\r' is a word separator (like a space), '\n' is a hard line break.
	// "a\r\nb" must wrap to clean ["a"] ["b"] with no stray control runes.
	lines := wrapSpans([]TextSpan{{Text: "a\r\nb"}}, 10)
	if hasControlRune(lines) {
		t.Fatalf("control rune leaked into output: %+v", lines)
	}
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d: %+v", len(lines), lines)
	}
	if len(lines[0]) != 1 || lines[0][0].Text != "a" {
		t.Errorf("line 0 = %+v, want single segment \"a\"", lines[0])
	}
	if len(lines[1]) != 1 || lines[1][0].Text != "b" {
		t.Errorf("line 1 = %+v, want single segment \"b\"", lines[1])
	}
}

func TestWrapSpans_VerticalTabAndFormFeedAreSeparators(t *testing.T) {
	// '\v' and '\f' separate words but do NOT break the line.
	lines := wrapSpans([]TextSpan{{Text: "a\vb\fc"}}, 20)
	if hasControlRune(lines) {
		t.Fatalf("control rune leaked into output: %+v", lines)
	}
	if len(lines) != 1 {
		t.Fatalf("want 1 line (no line break), got %d: %+v", len(lines), lines)
	}
	// Three words joined by single separator spaces on one line.
	if got := spanLineWidth(lines[0]); got != 5 { // "a b c"
		t.Errorf("line width = %d, want 5", got)
	}
}

func TestWrapSpans_SpanBoundaryIsNotWordBoundary_HardBreak(t *testing.T) {
	// "ab"+"cd" is the single logical word "abcd". At width 3 it must hard-break
	// by rune (["abc"] ["d"]), NOT split at the span seam (["ab"] ["cd"]).
	lines := wrapSpans([]TextSpan{{Text: "ab"}, {Text: "cd"}}, 3)
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d: %+v", len(lines), lines)
	}
	if joinSpanLine(lines[0]) != "abc" || joinSpanLine(lines[1]) != "d" {
		t.Errorf("got %q / %q, want \"abc\" / \"d\" (hard break, not seam split)",
			joinSpanLine(lines[0]), joinSpanLine(lines[1]))
	}
}

func TestWrapSpans_StylePreservedAcrossHardBreak(t *testing.T) {
	// "go"(plain)+"lang"(bold) is one 6-wide word "golang". At width 4 it
	// hard-breaks mid-word ("gola" / "ng") with bold preserved on every bold rune.
	bold := NewStyle().Bold()
	lines := wrapSpans([]TextSpan{{Text: "go"}, {Text: "lang", Style: bold}}, 4)
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d: %+v", len(lines), lines)
	}
	if joinSpanLine(lines[0]) != "gola" || joinSpanLine(lines[1]) != "ng" {
		t.Fatalf("got %q / %q, want \"gola\" / \"ng\"", joinSpanLine(lines[0]), joinSpanLine(lines[1]))
	}
	// Check style by segment (the rune 'g' appears in both the plain "go" and the
	// bold "lang", so we must assert per segment, not per rune):
	//   line 0: [{"go", plain}, {"la", bold}]
	//   line 1: [{"ng", bold}]
	if len(lines[0]) != 2 || lines[0][0].Text != "go" || lines[0][0].Style.Attrs&AttrBold != 0 {
		t.Errorf("line 0 seg 0 should be plain \"go\", got %+v", lines[0])
	}
	if len(lines[0]) != 2 || lines[0][1].Text != "la" || lines[0][1].Style.Attrs&AttrBold == 0 {
		t.Errorf("line 0 seg 1 should be bold \"la\", got %+v", lines[0])
	}
	if len(lines[1]) != 1 || lines[1][0].Text != "ng" || lines[1][0].Style.Attrs&AttrBold == 0 {
		t.Errorf("line 1 should be bold \"ng\", got %+v", lines[1])
	}
}

func TestWrapSpans_WideRuneHardBreakDoesNotOverflow(t *testing.T) {
	if RuneWidth('世') != 2 {
		t.Fatalf("precondition: expected '世' to be width 2, got %d", RuneWidth('世'))
	}
	// "世界世" is one 6-wide word. At width 3 no line may exceed maxWidth, and a
	// width-2 rune must not be split across lines.
	lines := wrapSpans([]TextSpan{{Text: "世界世"}}, 3)
	for i, line := range lines {
		if w := spanLineWidth(line); w > 3 {
			t.Errorf("line %d width = %d, overflows maxWidth 3: %+v", i, w, line)
		}
	}
}

func TestWrapSpans_PreservesLinkAndSplitsOnLinkChange(t *testing.T) {
	// "ab"(link X) + "cd"(link Y), no whitespace = one logical word "abcd" that
	// must split into two segments at the link boundary, each keeping its link.
	lines := wrapSpans([]TextSpan{
		{Text: "ab", Link: "X"},
		{Text: "cd", Link: "Y"},
	}, 40)
	if len(lines) != 1 || len(lines[0]) != 2 {
		t.Fatalf("want one line of two segments, got %+v", lines)
	}
	if lines[0][0].Text != "ab" || lines[0][0].Link != "X" {
		t.Errorf("seg 0 = %+v, want {ab, X}", lines[0][0])
	}
	if lines[0][1].Text != "cd" || lines[0][1].Link != "Y" {
		t.Errorf("seg 1 = %+v, want {cd, Y}", lines[0][1])
	}
}

func TestWrapSpans_LinkSpacesStayLinkedAndStyled(t *testing.T) {
	// A multi-word link must render as one continuous run: the spaces between its
	// words carry the link target and the link's style (e.g. underline), so the
	// hyperlink and its underline are not broken at the gaps.
	link := NewStyle().Underline()
	lines := wrapSpans([]TextSpan{
		{Text: "go to site", Style: link, Link: "http://x"},
	}, 40)
	if len(lines) != 1 {
		t.Fatalf("want one line, got %d", len(lines))
	}
	for i, seg := range lines[0] {
		if seg.Link != "http://x" {
			t.Errorf("segment %d %q lost the link: %+v", i, seg.Text, seg)
		}
		if seg.Style.Attrs&AttrUnderline == 0 {
			t.Errorf("segment %d %q lost the underline: %+v", i, seg.Text, seg)
		}
	}
}

func TestWrapSpans_NonLinkSeparatorStaysNeutral(t *testing.T) {
	// The deliberate neutral-separator behavior is preserved for non-link runs:
	// the space after a bold word is not itself bold.
	lines := wrapSpans([]TextSpan{
		{Text: "bold words", Style: NewStyle().Bold()},
	}, 40)
	if len(lines) != 1 {
		t.Fatalf("want one line, got %d", len(lines))
	}
	var joined strings.Builder
	for _, seg := range lines[0] {
		if seg.Text == " " && seg.Style.Attrs&AttrBold != 0 {
			t.Errorf("separator space should not be bold: %+v", seg)
		}
		joined.WriteString(seg.Text)
	}
	if joined.String() != "bold words" {
		t.Errorf("joined = %q, want %q", joined.String(), "bold words")
	}
}
