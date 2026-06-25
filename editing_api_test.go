package tui

import "testing"

// Grapheme fixtures: each is a single user-perceived glyph made of multiple runes.
// Built with explicit escapes so the ZWJ joiners and combining marks are visible.
var (
	flagUS    = "\U0001F1FA\U0001F1F8"             // US flag (2 runes, 1 cluster)
	zwjFamily = "\U0001F468‍\U0001F469‍\U0001F467" // man+woman+girl ZWJ family (5 runes, 1 cluster)
	skinWave  = "\U0001F44B\U0001F3FD"             // waving hand + medium skin tone (2 runes, 1 cluster)
	comboE    = "é"                               // e + combining acute (2 runes, 1 cluster)
)

func TestTextArea_InsertText_KeepsClustersWhole(t *testing.T) {
	type tc struct {
		insert        string
		wantClusters  int // ClusterCount of the inserted glyph (always 1)
		wantCursorPos int // cluster index after insert
	}

	tests := map[string]tc{
		"flag":       {insert: flagUS, wantClusters: 1, wantCursorPos: 1},
		"zwj family": {insert: zwjFamily, wantClusters: 1, wantCursorPos: 1},
		"skin tone":  {insert: skinWave, wantClusters: 1, wantCursorPos: 1},
		"combining":  {insert: comboE, wantClusters: 1, wantCursorPos: 1},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ta := NewTextArea()
			ta.BindApp(testApp)
			ta.InsertText(tt.insert)

			if got := ta.Text(); got != tt.insert {
				t.Fatalf("text = %q, want %q", got, tt.insert)
			}
			if got := ClusterCount(ta.Text()); got != tt.wantClusters {
				t.Fatalf("ClusterCount = %d, want %d", got, tt.wantClusters)
			}
			if got := ta.CursorPos(); got != tt.wantCursorPos {
				t.Fatalf("CursorPos = %d, want %d", got, tt.wantCursorPos)
			}
		})
	}
}

func TestTextArea_InsertText_RoutesThroughInsertPath(t *testing.T) {
	ta := NewTextArea()
	ta.BindApp(testApp)
	// Place a combining accent after a base by inserting the base first, then
	// the mark. InsertText must glue them and land the cursor after the cluster.
	ta.InsertText("e")
	ta.InsertText("́") // combining acute
	if got := ta.Text(); got != comboE {
		t.Fatalf("text = %q, want %q", got, comboE)
	}
	// One whole cluster, cursor after it.
	if got := ClusterCount(ta.Text()); got != 1 {
		t.Fatalf("ClusterCount = %d, want 1", got)
	}
	if got := ta.CursorPos(); got != 1 {
		t.Fatalf("CursorPos = %d, want 1", got)
	}
	// The cursor rune index must sit at the end of the multi-rune cluster.
	if got := ta.cursorPos.Get(); got != 2 {
		t.Fatalf("internal cursorPos = %d, want 2", got)
	}
}

func TestTextArea_CursorPos_RoundTripsThroughSetCursorPos(t *testing.T) {
	type tc struct {
		text string
		want int // ClusterCount(text)
	}

	tests := map[string]tc{
		"ascii":          {text: "hello", want: 5},
		"flags":          {text: flagUS + flagUS, want: 2},
		"mixed clusters": {text: "a" + flagUS + comboE + "z", want: 4},
		"zwj family":     {text: zwjFamily, want: 1},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ta := NewTextArea()
			ta.BindApp(testApp)
			ta.SetText(tt.text)

			if got := ClusterCount(tt.text); got != tt.want {
				t.Fatalf("ClusterCount(%q) = %d, want %d", tt.text, got, tt.want)
			}

			// Every cluster index round-trips through SetCursorPos/CursorPos.
			for pos := 0; pos <= tt.want; pos++ {
				ta.SetCursorPos(pos)
				if got := ta.CursorPos(); got != pos {
					t.Fatalf("SetCursorPos(%d) -> CursorPos() = %d", pos, got)
				}
			}
		})
	}
}

func TestTextArea_SetCursorPos_ClampsAndSnaps(t *testing.T) {
	ta := NewTextArea()
	ta.BindApp(testApp)
	ta.SetText("a" + flagUS) // 2 clusters, 3 runes

	// Past the end clamps to the last cluster boundary.
	ta.SetCursorPos(99)
	if got := ta.CursorPos(); got != 2 {
		t.Fatalf("CursorPos after over-clamp = %d, want 2", got)
	}
	if got := ta.cursorPos.Get(); got != 3 {
		t.Fatalf("internal cursorPos after over-clamp = %d, want 3", got)
	}

	// Negative clamps to 0.
	ta.SetCursorPos(-5)
	if got := ta.CursorPos(); got != 0 {
		t.Fatalf("CursorPos after negative = %d, want 0", got)
	}
}

func TestTextArea_InsertText_AtCursor_ClusterIndexSemantics(t *testing.T) {
	ta := NewTextArea()
	ta.BindApp(testApp)
	ta.SetText(flagUS + flagUS) // 2 clusters
	ta.SetCursorPos(1)          // between the two flags

	ta.InsertText("X")
	if got := ta.Text(); got != flagUS+"X"+flagUS {
		t.Fatalf("text = %q, want %q", got, flagUS+"X"+flagUS)
	}
	// Cursor now sits after the inserted X (cluster index 2).
	if got := ta.CursorPos(); got != 2 {
		t.Fatalf("CursorPos = %d, want 2", got)
	}
}

func TestInput_InsertText_KeepsClustersWhole(t *testing.T) {
	type tc struct {
		insert       string
		wantClusters int
	}

	tests := map[string]tc{
		"flag":       {insert: flagUS, wantClusters: 1},
		"zwj family": {insert: zwjFamily, wantClusters: 1},
		"skin tone":  {insert: skinWave, wantClusters: 1},
		"combining":  {insert: comboE, wantClusters: 1},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inp := NewInput(WithInputWidth(20))
			inp.BindApp(testApp)
			inp.InsertText(tt.insert)

			if got := inp.Text(); got != tt.insert {
				t.Fatalf("text = %q, want %q", got, tt.insert)
			}
			if got := ClusterCount(inp.Text()); got != tt.wantClusters {
				t.Fatalf("ClusterCount = %d, want %d", got, tt.wantClusters)
			}
			if got := inp.CursorPos(); got != tt.wantClusters {
				t.Fatalf("CursorPos = %d, want %d", got, tt.wantClusters)
			}
		})
	}
}

func TestInput_InsertText_FiresOnChangeAndAdvancesCursor(t *testing.T) {
	var changes []string
	inp := NewInput(WithInputWidth(20), WithInputOnChange(func(s string) {
		changes = append(changes, s)
	}))
	inp.BindApp(testApp)

	inp.InsertText("ab")
	inp.InsertText(flagUS)

	if got := inp.Text(); got != "ab"+flagUS {
		t.Fatalf("text = %q, want %q", got, "ab"+flagUS)
	}
	// 3 clusters: a, b, flag.
	if got := inp.CursorPos(); got != 3 {
		t.Fatalf("CursorPos = %d, want 3", got)
	}
	if len(changes) != 2 {
		t.Fatalf("onChange called %d times, want 2", len(changes))
	}
}

func TestInput_CursorPos_RoundTripsThroughSetCursorPos(t *testing.T) {
	inp := NewInput(WithInputWidth(20))
	inp.BindApp(testApp)
	text := "a" + flagUS + comboE + "z" // 4 clusters
	inp.SetText(text)

	want := ClusterCount(text)
	if want != 4 {
		t.Fatalf("ClusterCount = %d, want 4", want)
	}
	for pos := 0; pos <= want; pos++ {
		inp.SetCursorPos(pos)
		if got := inp.CursorPos(); got != pos {
			t.Fatalf("SetCursorPos(%d) -> CursorPos() = %d", pos, got)
		}
	}
}

func TestInput_SetCursorPos_Clamps(t *testing.T) {
	inp := NewInput(WithInputWidth(20))
	inp.BindApp(testApp)
	inp.SetText(flagUS) // 1 cluster, 2 runes

	inp.SetCursorPos(99)
	if got := inp.CursorPos(); got != 1 {
		t.Fatalf("CursorPos after over-clamp = %d, want 1", got)
	}
	if got := inp.cursorPos.Get(); got != 2 {
		t.Fatalf("internal cursorPos = %d, want 2", got)
	}

	inp.SetCursorPos(-3)
	if got := inp.CursorPos(); got != 0 {
		t.Fatalf("CursorPos after negative = %d, want 0", got)
	}
}
