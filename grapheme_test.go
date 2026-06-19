package tui

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// Test data uses escapes so the intent is unambiguous in source:
//
//	\U0001F680                          rocket
//	\U0001F1FA\U0001F1F8                US flag (regional-indicator pair)
//	\U0001F44D\U0001F3FD                thumbs up + medium skin tone
//	\U0001F468‍\U0001F469...\U0001F466       man ZWJ woman ZWJ girl ZWJ boy (family)
//	❤️                        heavy heart + VS16 (emoji presentation)
//	1️⃣                       keycap digit one
//	é                             "e" + combining acute ("é", decomposed)
//	a​b                            "a" + zero-width space + "b"
const (
	emojiRocket  = "\U0001F680"
	emojiFlagUS  = "\U0001F1FA\U0001F1F8"
	emojiThumbs  = "\U0001F44D\U0001F3FD"
	emojiFamily  = "\U0001F468\u200d\U0001F469\u200d\U0001F467\u200d\U0001F466"
	emojiHeart   = "\u2764\U0000FE0F"
	emojiKeycap1 = "1\U0000FE0F\u20E3"
	accentE      = "e\u0301"
)

func TestNextCluster_SegmentationAndWidth(t *testing.T) {
	type tc struct {
		in       string
		clusters []string
		width    int
	}

	tests := map[string]tc{
		"ascii":        {in: "ab", clusters: []string{"a", "b"}, width: 2},
		"cjk":          {in: "你好", clusters: []string{"你", "好"}, width: 4},
		"single emoji": {in: emojiRocket, clusters: []string{emojiRocket}, width: 2},
		// Regional-indicator pair, skin-tone, ZWJ family, heart+VS16, and keycap are
		// each one cluster two columns wide, not the sum of their code points.
		"flag":       {in: emojiFlagUS, clusters: []string{emojiFlagUS}, width: 2},
		"skin tone":  {in: emojiThumbs, clusters: []string{emojiThumbs}, width: 2},
		"zwj family": {in: emojiFamily, clusters: []string{emojiFamily}, width: 2},
		"heart vs16": {in: emojiHeart, clusters: []string{emojiHeart}, width: 2},
		"keycap":     {in: emojiKeycap1, clusters: []string{emojiKeycap1}, width: 2},
		// ASCII fast path must not split a base from a following combining mark:
		// the byte after 'e' is 0xCC (a leading byte), not ASCII, so it decodes.
		"decomposed accent":  {in: accentE, clusters: []string{accentE}, width: 1},
		"decomposed in word": {in: "caf" + accentE, clusters: []string{"c", "a", "f", accentE}, width: 4},
		// A format control with GCB=Control (ZWSP) must NOT glue to its neighbours.
		"zwsp does not glue": {in: "a\u200bb", clusters: []string{"a", "\u200b", "b"}, width: 3},
		// Regional indicators pair at most two at a time; a third stands alone.
		"three regional indicators": {
			in:       emojiFlagUS + "\U0001F1EB",
			clusters: []string{emojiFlagUS, "\U0001F1EB"},
			width:    4,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var got []string
			total := 0
			for rest := tt.in; len(rest) > 0; {
				cl, w, size := nextCluster(rest)
				if size == 0 {
					t.Fatalf("nextCluster(%q) returned size 0", rest)
				}
				got = append(got, cl)
				total += w
				rest = rest[size:]
			}

			if len(got) != len(tt.clusters) {
				t.Fatalf("clusters = %q, want %q", got, tt.clusters)
			}
			for i := range got {
				if got[i] != tt.clusters[i] {
					t.Errorf("cluster[%d] = %q, want %q", i, got[i], tt.clusters[i])
				}
			}
			if strings.Join(got, "") != tt.in {
				t.Errorf("clusters %q do not reconstruct input %q", got, tt.in)
			}
			if total != tt.width {
				t.Errorf("summed cluster width = %d, want %d", total, tt.width)
			}
			if sw := stringWidth(tt.in); sw != tt.width {
				t.Errorf("stringWidth(%q) = %d, want %d", tt.in, sw, tt.width)
			}
		})
	}
}

// TestStringWidth_Clusters documents the exact divergence from issue #95: a
// per-code-point sum over-counts these, the grapheme-aware width does not.
func TestStringWidth_Clusters(t *testing.T) {
	type tc struct {
		in    string
		width int
	}

	tests := map[string]tc{
		"flag":         {in: emojiFlagUS, width: 2},
		"skin tone":    {in: emojiThumbs, width: 2},
		"zwj family":   {in: emojiFamily, width: 2},
		"decomposed":   {in: "caf" + accentE, width: 4},
		"flag in text": {in: "x" + emojiFlagUS + "y", width: 4},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := stringWidth(tt.in); got != tt.width {
				t.Errorf("stringWidth(%q) = %d, want %d", tt.in, got, tt.width)
			}
		})
	}
}

// TestNextClusterBytes_MatchesNextCluster pins that the []byte scanner variant
// segments identically to the string variant (same shared state machine).
func TestNextClusterBytes_MatchesNextCluster(t *testing.T) {
	inputs := map[string]string{
		"ascii":      "ab",
		"cjk":        "你好",
		"rocket":     emojiRocket,
		"flag":       emojiFlagUS,
		"skin tone":  emojiThumbs,
		"zwj family": emojiFamily,
		"heart vs16": emojiHeart,
		"keycap":     emojiKeycap1,
		"accent":     accentE,
		"zwsp":       "a\u200bb",
	}

	for name, in := range inputs {
		t.Run(name, func(t *testing.T) {
			for rest := in; len(rest) > 0; {
				cl, w, size := nextCluster(rest)
				bw, bsize, base := nextClusterBytes([]byte(rest))
				if bw != w || bsize != size {
					t.Fatalf("nextClusterBytes = (w=%d size=%d), nextCluster = (w=%d size=%d)", bw, bsize, w, size)
				}
				wantBase, _ := utf8.DecodeRuneInString(cl)
				if base != wantBase {
					t.Errorf("base rune = %q, want %q", base, wantBase)
				}
				rest = rest[size:]
			}
		})
	}
}

func TestClusterRuneStarts(t *testing.T) {
	type tc struct {
		in   string
		want []int
	}

	tests := map[string]tc{
		"ascii":                 {in: "ab", want: []int{0, 1, 2}},
		"family is one cluster": {in: emojiFamily, want: []int{0, 7}},
		// c, a, f are one rune each; the final "é" is two runes (e + combining).
		"decomposed in word": {in: "caf" + accentE, want: []int{0, 1, 2, 3, 5}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := clusterRuneStarts(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("clusterRuneStarts(%q) = %v, want %v", tt.in, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("clusterRuneStarts(%q) = %v, want %v", tt.in, got, tt.want)
				}
			}
		})
	}
}
