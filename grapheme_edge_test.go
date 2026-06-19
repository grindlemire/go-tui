package tui

import (
	"strings"
	"testing"
)

// TestNextCluster_EdgeCases covers edge cases in the cluster state machine.
func TestNextCluster_EdgeCases(t *testing.T) {
	type tc struct {
		name   string
		input  string
		wantCl string
		wantW  int
		wantSz int
	}

	cases := []tc{
		{name: "empty", input: "", wantCl: "", wantW: 0, wantSz: 0},
		{name: "single ascii", input: "a", wantCl: "a", wantW: 1, wantSz: 1},
		{name: "two ascii", input: "ab", wantCl: "a", wantW: 1, wantSz: 1},
		{name: "ascii then emoji", input: "a\U0001F680", wantCl: "a", wantW: 1, wantSz: 1},
		{name: "single cjk", input: "世", wantCl: "世", wantW: 2, wantSz: 3},
		{name: "cjk then ascii", input: "世a", wantCl: "世", wantW: 2, wantSz: 3},
		{name: "zwj two emoji", input: "\U0001F468‍\U0001F469", wantCl: "\U0001F468‍\U0001F469", wantW: 2, wantSz: 11},
		{name: "lone RI", input: "\U0001F1FAx", wantCl: "\U0001F1FA", wantW: 2, wantSz: 4},
		{name: "leading combining", input: "́a", wantCl: "́", wantW: 1, wantSz: 2},
		{name: "zwnj between ascii", input: "a‌b", wantCl: "a‌", wantW: 1, wantSz: 4},
		{name: "skin tone", input: "\U0001F44D\U0001F3FD", wantCl: "\U0001F44D\U0001F3FD", wantW: 2, wantSz: 8},
		{name: "vs15", input: "☝\U0000FE0Ex", wantCl: "☝\U0000FE0E", wantW: 1, wantSz: 6},
		// VS15 on a base that is emoji-wide by default (U+1F600, width 2) must
		// narrow the cluster to width 1. This exercises the narrowing branch of
		// clusterExtendUpdateWidth, which the "☝" case above does not (U+261D is
		// already width 1).
		{name: "vs15 narrows wide emoji", input: "\U0001F600\U0000FE0E", wantCl: "\U0001F600\U0000FE0E", wantW: 1, wantSz: 7},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cl, w, sz := nextCluster(tt.input)
			if cl != tt.wantCl || w != tt.wantW || sz != tt.wantSz {
				t.Errorf("nextCluster(%q) = (%q, %d, %d), want (%q, %d, %d)",
					tt.input, cl, w, sz, tt.wantCl, tt.wantW, tt.wantSz)
			}
		})
	}
}

// TestNextClusterBytes_EdgeCases covers []byte variant edge cases.
func TestNextClusterBytes_EdgeCases(t *testing.T) {
	w, sz, base := nextClusterBytes(nil)
	if w != 0 || sz != 0 || base != 0 {
		t.Errorf("nextClusterBytes(nil) = (%d, %d, %d), want (0, 0, 0)", w, sz, base)
	}

	w, sz, base = nextClusterBytes([]byte{'a'})
	if w != 1 || sz != 1 || base != 'a' {
		t.Errorf("nextClusterBytes('a') = (%d, %d, %c), want (1, 1, a)", w, sz, base)
	}

	w, sz, base = nextClusterBytes([]byte("ab"))
	if w != 1 || sz != 1 || base != 'a' {
		t.Errorf("nextClusterBytes('ab') = (%d, %d, %c), want (1, 1, a)", w, sz, base)
	}

	w, sz, base = nextClusterBytes([]byte(emojiFlagUS))
	if w != 2 || sz != 8 || base != 0x1F1FA {
		t.Errorf("nextClusterBytes(flag) = (%d, %d, %x), want (2, 8, 1F1FA)", w, sz, base)
	}
}

// TestClusterCount_EdgeCases tests cluster counting.
func TestClusterCount_EdgeCases(t *testing.T) {
	type tc struct {
		name string
		s    string
		want int
	}

	cases := []tc{
		{name: "empty", s: "", want: 0},
		{name: "ascii", s: "hello", want: 5},
		{name: "flag", s: emojiFlagUS, want: 1},
		{name: "flag x flag", s: emojiFlagUS + "x" + emojiFlagUS, want: 3},
		{name: "family", s: emojiFamily, want: 1},
		{name: "mixed", s: "a" + emojiFamily + "b" + emojiFlagUS, want: 4},
		{name: "accent", s: accentE, want: 1},
		{name: "cjk", s: "你好世界", want: 4},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := clusterCount(tt.s); got != tt.want {
				t.Errorf("clusterCount(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

// TestBuffer_SetString_Overlap tests setCluster properly handles overlapping
// wide characters.
func TestBuffer_SetString_Overlap(t *testing.T) {
	b := NewBuffer(10, 1)
	b.SetString(0, 0, "你好", NewStyle())
	b.SetString(1, 0, "x", NewStyle())
	if got := b.Cell(1, 0).Rune; got != 'x' {
		t.Errorf("cell(1,0).Rune = %q, want 'x'", got)
	}
}

// TestBuffer_SetStringGradient_ClusterMulti tests gradient with multiple cluster types.
func TestBuffer_SetStringGradient_ClusterMulti(t *testing.T) {
	b := NewBuffer(20, 1)
	g := NewGradient(Red, Blue)
	text := "a" + emojiFlagUS + "b" + accentE + "c" + emojiFamily + "d"
	total := b.SetStringGradient(0, 0, text, g, NewStyle())
	if total != 9 {
		t.Errorf("SetStringGradient returned width %d, want 9", total)
	}
	for x := range total {
		c := b.Cell(x, 0)
		if c.IsContinuation() || c.Rune == 0 {
			continue
		}
		if c.Style.Fg.IsDefault() {
			t.Errorf("cell(%d,0).Fg is default for rune %q", x, c.Rune)
		}
	}
}

// TestInput_CombiningInsert tests editing with combining marks.
func TestInput_CombiningInsert(t *testing.T) {
	inp := newTestInput(WithInputWidth(10))
	inp.text.Set("a")
	inp.cursorPos.Set(1)
	inp.focused.Set(true)
	inp.blink.Set(true)
	inp.HandleEvent(KeyEvent{Key: KeyRune, Rune: '́'})
	if text := inp.Text(); text != "á" {
		t.Errorf("text = %q, want %q", text, "á")
	}
	if got := inp.cursorPos.Get(); got != 2 {
		t.Errorf("cursorPos = %d, want 2", got)
	}
}

// TestInput_ZWJInsert tests editing with ZWJ sequences.
func TestInput_ZWJInsert(t *testing.T) {
	inp := newTestInput(WithInputWidth(10))
	inp.SetText("a")
	inp.focused.Set(true)
	inp.blink.Set(true)

	inp.HandleEvent(KeyEvent{Key: KeyRune, Rune: '\U0001F468'})
	inp.HandleEvent(KeyEvent{Key: KeyRune, Rune: '‍'})
	inp.HandleEvent(KeyEvent{Key: KeyRune, Rune: '\U0001F469'})

	// 'a' + man (U+1F468) + ZWJ (U+200D) + woman (U+1F469)
	want := "a\U0001F468‍\U0001F469"
	if text := inp.Text(); text != want {
		t.Errorf("text = %q, want %q", text, want)
	}
}

// TestBuffer_String_Combining verifies String/StringTrimmed preserve clusters.
func TestBuffer_String_Combining(t *testing.T) {
	inputs := []struct {
		name string
		text string
	}{
		{"flag", emojiFlagUS},
		{"family", emojiFamily},
		{"accent", accentE},
		{"mixed", "a" + emojiFlagUS + "b" + accentE + "c" + emojiFamily + "d"},
		{"cjk", "你好世界"},
		{"keycap", emojiKeycap1},
		{"heart+vs16", emojiHeart},
	}

	for _, tt := range inputs {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuffer(30, 1)
			b.SetString(0, 0, tt.text, NewStyle())
			if s := b.String(); !strings.Contains(s, tt.text) {
				t.Errorf("String() = %q, should contain %q", s, tt.text)
			}
			if st := b.StringTrimmed(); !strings.Contains(st, tt.text) {
				t.Errorf("StringTrimmed() = %q, should contain %q", st, tt.text)
			}
		})
	}
}

// TestSegmentStyledRunes_MultiRune tests multi-rune clusters across spans.
func TestSegmentStyledRunes_MultiRune(t *testing.T) {
	styleA := NewStyle().Bold()
	styleB := NewStyle().Italic()
	rs := []styledRune{{r: 'a', st: styleA}, {r: 0x0301, st: styleB}}
	clusters := segmentStyledRunes(rs)
	if len(clusters) != 1 {
		t.Fatalf("segmentStyledRunes returned %d clusters, want 1", len(clusters))
	}
	if clusters[0].st != styleA {
		t.Errorf("cluster style should be base rune's style")
	}
	if clusters[0].text != "á" {
		t.Errorf("cluster text = %q, want %q", clusters[0].text, "á")
	}
	if clusters[0].width != 1 {
		t.Errorf("cluster width = %d, want 1", clusters[0].width)
	}
}

// TestSegmentStyledRunes_Flag tests flag emoji as styled runes.
func TestSegmentStyledRunes_Flag(t *testing.T) {
	styleA := NewStyle().Bold()
	rs := []styledRune{{r: 0x1F1FA, st: styleA}, {r: 0x1F1F8, st: styleA}}
	clusters := segmentStyledRunes(rs)
	if len(clusters) != 1 {
		t.Fatalf("segmentStyledRunes returned %d clusters, want 1", len(clusters))
	}
	if clusters[0].text != emojiFlagUS {
		t.Errorf("cluster text = %q, want US flag", clusters[0].text)
	}
	if clusters[0].width != 2 {
		t.Errorf("cluster width = %d, want 2", clusters[0].width)
	}
}

// TestNewClusterCell_EdgeCases covers newClusterCell.
func TestNewClusterCell_EdgeCases(t *testing.T) {
	c := newClusterCell("", 1, NewStyle(), "")
	if c.Rune != 0 || c.Combining != "" {
		t.Errorf("empty cell: Rune=%d Combining=%q", c.Rune, c.Combining)
	}
	c = newClusterCell("a", 0, NewStyle(), "")
	if c.Rune != 0 || c.Combining != "" {
		t.Errorf("width-0 cell should be empty")
	}
	c = newClusterCell("你", 2, NewStyle(), "")
	if c.Rune != '你' || c.Combining != "" {
		t.Errorf("single CJK: Rune=%c Combining=%q", c.Rune, c.Combining)
	}
}

// TestSnapRuneToClusterStart_AllCases tests snapRuneToClusterStart.
func TestSnapRuneToClusterStart_AllCases(t *testing.T) {
	type tc struct {
		name string
		s    string
		idx  int
		want int
	}
	cases := []tc{
		{name: "ascii 0", s: "abc", idx: 0, want: 0},
		{name: "ascii 1", s: "abc", idx: 1, want: 1},
		{name: "ascii 2", s: "abc", idx: 2, want: 2},
		{name: "ascii 3", s: "abc", idx: 3, want: 3},
		{name: "inside flag", s: "x" + emojiFlagUS + "y", idx: 2, want: 1},
		{name: "at flag start", s: "x" + emojiFlagUS + "y", idx: 1, want: 1},
		{name: "at flag end", s: "x" + emojiFlagUS + "y", idx: 3, want: 3},
		{name: "inside accent", s: "caf" + accentE, idx: 4, want: 3},
		{name: "negative", s: "abc", idx: -1, want: 0},
		{name: "past end", s: "abc", idx: 99, want: 3},
		{name: "empty", s: "", idx: 0, want: 0},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := snapRuneToClusterStart(tt.s, tt.idx); got != tt.want {
				t.Errorf("snapRuneToClusterStart(%q, %d) = %d, want %d", tt.s, tt.idx, got, tt.want)
			}
		})
	}
}

// TestRuneIndexToDisplayCol_AllCases covers runeIndexToDisplayCol.
func TestRuneIndexToDisplayCol_AllCases(t *testing.T) {
	type tc struct {
		name string
		s    string
		idx  int
		want int
	}
	cases := []tc{
		{name: "ascii 0", s: "abc", idx: 0, want: 0},
		{name: "ascii 1", s: "abc", idx: 1, want: 1},
		{name: "ascii 3", s: "abc", idx: 3, want: 3},
		{name: "cjk 0", s: "你好", idx: 0, want: 0},
		{name: "cjk 1", s: "你好", idx: 1, want: 2},
		{name: "cjk 2", s: "你好", idx: 2, want: 4},
		{name: "flag 0", s: emojiFlagUS + "x", idx: 0, want: 0},
		{name: "flag inside", s: emojiFlagUS + "x", idx: 1, want: 0},
		{name: "flag after", s: emojiFlagUS + "x", idx: 2, want: 2},
		{name: "flag past", s: emojiFlagUS + "x", idx: 3, want: 3},
		{name: "zwj 0", s: emojiFamily + "x", idx: 0, want: 0},
		{name: "zwj inside", s: emojiFamily + "x", idx: 3, want: 0},
		{name: "zwj at boundary", s: emojiFamily + "x", idx: 7, want: 2},
		{name: "zwj after", s: emojiFamily + "x", idx: 8, want: 3},
		{name: "accent inside", s: "caf" + accentE, idx: 4, want: 3},
		{name: "accent at end", s: "caf" + accentE, idx: 5, want: 4},
		{name: "empty", s: "", idx: 0, want: 0},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := runeIndexToDisplayCol(tt.s, tt.idx); got != tt.want {
				t.Errorf("runeIndexToDisplayCol(%q, %d) = %d, want %d", tt.s, tt.idx, got, tt.want)
			}
		})
	}
}

// TestCellGlyph verifies cellGlyph reconstructs the cell's glyph from its base
// rune plus combining tail, including the empty-cell case.
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
