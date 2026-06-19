package tui

import (
	"strings"
	"testing"
)

// TestViewportText_ClusterBoundary tests viewportText starting at column
// boundaries with cursor set to not trigger scroll adjustment.
func TestViewportText_ClusterBoundary(t *testing.T) {
	clusters := []displayCluster{
		{text: "a", width: 1},
		{text: emojiFlagUS, width: 2},
		{text: "b", width: 1},
	}

	cases := []struct {
		name      string
		clusters  []displayCluster
		scroll    int
		cursorCol int
		visible   int
		want      string
	}{
		{name: "scroll 0 cursor inside", clusters: clusters, scroll: 0, cursorCol: 2, visible: 4, want: "a" + emojiFlagUS + "b"},
		{name: "scroll 1 cursor at 3", clusters: clusters, scroll: 1, cursorCol: 3, visible: 4, want: emojiFlagUS + "b"},
		{name: "scroll 3 cursor at 4", clusters: clusters, scroll: 3, cursorCol: 4, visible: 4, want: "b"},
		{name: "scroll beyond all", clusters: []displayCluster{{text: "a", width: 1}}, scroll: 5, cursorCol: 5, visible: 5, want: " "},
		{name: "empty clusters", clusters: nil, scroll: 0, cursorCol: 0, visible: 5, want: " "},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			inp := newTestInput(WithInputWidth(tt.visible))
			var sb strings.Builder
			for _, c := range tt.clusters {
				sb.WriteString(c.text)
			}
			inp.text.Set(sb.String())
			// Ensure cursor is at or to the right of scroll so scroll isn't adjusted
			// backward. We use cursorCol as the approximate rune index.
			inp.cursorPos.Set(tt.cursorCol)
			inp.scrollPos.Set(tt.scroll)
			got := inp.viewportText(tt.clusters, 0, tt.visible)
			if got != tt.want {
				t.Errorf("viewportText(scroll=%d, visible=%d) = %q, want %q", tt.scroll, tt.visible, got, tt.want)
			}
		})
	}
}

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
		{name: "single cjk", input: "\u4E16", wantCl: "\u4E16", wantW: 2, wantSz: 3},
		{name: "cjk then ascii", input: "\u4E16a", wantCl: "\u4E16", wantW: 2, wantSz: 3},
		{name: "zwj two emoji", input: "\U0001F468\u200d\U0001F469", wantCl: "\U0001F468\u200d\U0001F469", wantW: 2, wantSz: 11},
		{name: "lone RI", input: "\U0001F1FAx", wantCl: "\U0001F1FA", wantW: 2, wantSz: 4},
		{name: "leading combining", input: "\u0301a", wantCl: "\u0301", wantW: 1, wantSz: 2},
		{name: "zwnj between ascii", input: "a\u200cb", wantCl: "a\u200c", wantW: 1, wantSz: 4},
		{name: "skin tone", input: "\U0001F44D\U0001F3FD", wantCl: "\U0001F44D\U0001F3FD", wantW: 2, wantSz: 8},
		{name: "vs15", input: "\u261D\U0000FE0Ex", wantCl: "\u261D\U0000FE0E", wantW: 1, wantSz: 6},
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
		{name: "cjk", s: "\u4F60\u597D\u4E16\u754C", want: 4},
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
	b.SetString(0, 0, "\u4F60\u597D", NewStyle())
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
	for x := 0; x < total; x++ {
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
	inp.HandleEvent(KeyEvent{Key: KeyRune, Rune: '\u0301'})
	if text := inp.Text(); text != "a\u0301" {
		t.Errorf("text = %q, want %q", text, "a\u0301")
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
	inp.HandleEvent(KeyEvent{Key: KeyRune, Rune: '\u200D'})
	inp.HandleEvent(KeyEvent{Key: KeyRune, Rune: '\U0001F469'})

	want := "a\u200d\U0001F469"
	// Check that the man+ZWJ+woman formed a single cluster
	// The text should be: 'a' + man + ZWJ + woman
	if text := inp.Text(); strings.Count(text, "\u200d") != 1 {
		t.Errorf("text = %q, should contain one ZWJ", text)
	}
	if text := inp.Text(); !strings.Contains(text, "\U0001F468") || !strings.Contains(text, "\U0001F469") {
		t.Errorf("text = %q should contain both emoji", text)
	}
	_ = want
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
		{"cjk", "\u4F60\u597D\u4E16\u754C"},
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
	if clusters[0].text != "a\u0301" {
		t.Errorf("cluster text = %q, want %q", clusters[0].text, "a\u0301")
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
	c = newClusterCell("\u4F60", 2, NewStyle(), "")
	if c.Rune != '\u4F60' || c.Combining != "" {
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
		{name: "cjk 0", s: "\u4F60\u597D", idx: 0, want: 0},
		{name: "cjk 1", s: "\u4F60\u597D", idx: 1, want: 2},
		{name: "cjk 2", s: "\u4F60\u597D", idx: 2, want: 4},
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
