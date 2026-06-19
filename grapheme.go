package tui

import (
	"unicode"
	"unicode/utf8"
)

// nextCluster returns the next grapheme cluster at the start of s, its display
// width in terminal cells, and the number of bytes it consumed. The returned
// cluster is a slice of s (no allocation); callers that persist it must clone it.
//
// The cluster profile targets terminal text: combining marks, ZWJ emoji
// sequences, regional-indicator pairs, emoji modifiers (skin tones), and
// variation selectors. It is not full UAX #29: Hangul jamo composition, Prepend,
// Indic conjuncts, and emoji tag sequences (subdivision flags) are not handled.
// Width 0 is never returned: the buffer model reserves 0 for continuation cells,
// and a defective leading combining mark becomes a width-1 cluster.
func nextCluster(s string) (cluster string, width, size int) {
	if len(s) == 0 {
		return "", 0, 0
	}

	// ASCII fast path: only when the next byte is also ASCII (or we are at the
	// end). A non-ASCII next byte (leading byte >= 0x80, e.g. 0xCC starting a
	// combining accent) is NOT a continuation byte but could still attach to
	// this base, so it forces the full decode below.
	if s[0] < 0x80 {
		if len(s) == 1 || s[1] < 0x80 {
			return s[:1], 1, 1
		}
	}

	w, size, _ := clusterAdvance(decodeStringStep(s))
	return s[:size], w, size
}

// nextClusterBytes is the []byte variant for the inline byte scanner. It shares
// the cluster state machine with nextCluster (one implementation, two decode
// steps). It reports the cluster's display width, the number of bytes consumed,
// and the cluster's base rune.
func nextClusterBytes(b []byte) (width, size int, base rune) {
	if len(b) == 0 {
		return 0, 0, 0
	}

	if b[0] < 0x80 {
		if len(b) == 1 || b[1] < 0x80 {
			return 1, 1, rune(b[0])
		}
	}

	return clusterAdvance(decodeBytesStep(b))
}

// decodeStep decodes one rune at a byte offset. Returns the rune, its byte
// size, and the total length of the underlying buffer. A decode at or past the
// end returns size 0.
type decodeStep func(pos int) (r rune, size, length int)

func decodeStringStep(s string) decodeStep {
	return func(pos int) (rune, int, int) {
		if pos >= len(s) {
			return 0, 0, len(s)
		}
		r, size := utf8.DecodeRuneInString(s[pos:])
		return r, size, len(s)
	}
}

func decodeBytesStep(b []byte) decodeStep {
	return func(pos int) (rune, int, int) {
		if pos >= len(b) {
			return 0, 0, len(b)
		}
		r, size := utf8.DecodeRune(b[pos:])
		return r, size, len(b)
	}
}

// clusterAdvance runs the shared grapheme state machine over a decode step,
// starting at offset 0. It returns the cluster's display width, the number of
// bytes consumed, and the base rune.
func clusterAdvance(decode decodeStep) (width, size int, base rune) {
	r0, s0, length := decode(0)
	if s0 == 0 {
		return 0, 0, 0
	}
	pos := s0
	w := baseRuneWidth(r0)

	if regionalIndicator(r0) {
		if pos < length {
			r1, s1, _ := decode(pos)
			if s1 > 0 && regionalIndicator(r1) {
				pos += s1
				return 2, pos, r0
			}
		}
		// Lone RI: w is already 2 (RI is in emojiWideRanges). Fall through so a
		// trailing combining mark or ZWJ sequence still attaches.
	}

	lastWasZWJ := false
	for pos < length {
		r, sz, _ := decode(pos)
		if sz == 0 {
			break
		}
		if graphemeExtend(r) {
			pos += sz
			if isVS16(r) {
				w = 2 // emoji presentation forces wide
			}
			lastWasZWJ = isZWJ(r)
			continue
		}
		if lastWasZWJ {
			// Emoji (or any base) after a ZWJ joins the cluster.
			pos += sz
			w = 2
			lastWasZWJ = false
			continue
		}
		break
	}

	return w, pos, r0
}

// baseRuneWidth returns the display width of a base rune ignoring the
// zero-width override: 2 for East Asian wide / emoji-wide ranges, else 1.
func baseRuneWidth(r rune) int {
	if r < 0 || r > unicode.MaxRune {
		return 1
	}
	// C0 and C1 controls render narrow.
	if r < 0x20 || (r >= 0x7F && r < 0xA0) {
		return 1
	}
	if inRuneRanges(r, eastAsianWideRanges) || inRuneRanges(r, emojiWideRanges) {
		return 2
	}
	return 1
}

// graphemeExtend reports whether r attaches to the preceding base, contributing
// zero width. This covers combining marks (Mn/Me/Mc), the join controls (ZWNJ
// and ZWJ, exactly U+200C and U+200D), variation selectors, and the emoji
// modifier range U+1F3FB..U+1F3FF (skin tones).
//
// It deliberately does NOT include the whole unicode.Cf category: most format
// characters (bidi controls, ZWSP U+200B) have UAX #29 GCB=Control and must
// break, not glue. The cost is that emoji tag sequences (Cf tag chars) do not
// cluster.
//
// Keep this deliberately narrower than isZeroWidthRune in cell.go (which uses the
// broad unicode.Cf, fine for per-rune width). Do not unify the two predicates, or
// format controls would wrongly glue into grapheme clusters.
func graphemeExtend(r rune) bool {
	if r >= 0x1F3FB && r <= 0x1F3FF {
		return true
	}
	return unicode.In(r, unicode.Mn, unicode.Me, unicode.Mc, unicode.Join_Control, unicode.Variation_Selector)
}

// isZWJ reports whether r is the ZERO WIDTH JOINER.
func isZWJ(r rune) bool {
	return r == 0x200D
}

// isVS16 reports whether r is VARIATION SELECTOR-16 (emoji presentation).
func isVS16(r rune) bool {
	return r == 0xFE0F
}

// regionalIndicator reports whether r is a regional indicator symbol (used in
// pairs to form flag emoji).
func regionalIndicator(r rune) bool {
	return r >= 0x1F1E6 && r <= 0x1F1FF
}

// clusterCount returns the number of grapheme clusters in s.
func clusterCount(s string) int {
	n := 0
	for len(s) > 0 {
		_, _, size := nextCluster(s)
		if size == 0 {
			break
		}
		n++
		s = s[size:]
	}
	return n
}

// clusterRuneStarts returns the rune index where each grapheme cluster begins,
// followed by the total rune count. For "ab" it returns [0,1,2]; for the family
// emoji (7 runes, one cluster) it returns [0,7]. Used by editable widgets to
// snap a cursor rune index onto a cluster boundary.
func clusterRuneStarts(s string) []int {
	starts := []int{0}
	runeIdx := 0
	for len(s) > 0 {
		_, _, size := nextCluster(s)
		if size == 0 {
			break
		}
		// Count runes in this cluster.
		for i := 0; i < size; {
			_, rs := utf8.DecodeRuneInString(s[i:])
			if rs == 0 {
				break
			}
			i += rs
			runeIdx++
		}
		starts = append(starts, runeIdx)
		s = s[size:]
	}
	return starts
}

// clusterEndAfterInsert returns the rune index at the end of the cluster that
// contains the rune at insertedRuneIdx in s. It re-segments from the cluster
// start at/under insertedRuneIdx, so a combining mark inserted after a base lands
// the cursor after the combined cluster, and a base char lands one cluster on.
func clusterEndAfterInsert(s string, insertedRuneIdx int) int {
	starts := clusterRuneStarts(s)
	// Find the cluster whose [start, end) range contains insertedRuneIdx.
	for i := 0; i+1 < len(starts); i++ {
		if insertedRuneIdx >= starts[i] && insertedRuneIdx < starts[i+1] {
			return starts[i+1]
		}
	}
	// At or past the end (e.g. appended at the very end): cursor at total.
	return starts[len(starts)-1]
}

// snapRuneToClusterStart snaps a rune index down to the nearest cluster start at
// or before it (clamped to [0, total]).
func snapRuneToClusterStart(s string, runeIdx int) int {
	starts := clusterRuneStarts(s)
	if runeIdx <= 0 {
		return 0
	}
	last := starts[len(starts)-1]
	if runeIdx >= last {
		return last
	}
	snapped := 0
	for _, st := range starts {
		if st <= runeIdx {
			snapped = st
		} else {
			break
		}
	}
	return snapped
}

// runeIndexToDisplayCol returns the display column at the given rune index,
// summing cluster widths from the start of s. The index is snapped to a cluster
// boundary first so a position inside a cluster reports that cluster's start col.
func runeIndexToDisplayCol(s string, runeIdx int) int {
	col := 0
	runeAt := 0
	for len(s) > 0 {
		_, w, size := nextCluster(s)
		if size == 0 {
			break
		}
		clusterRunes := utf8.RuneCountInString(s[:size])
		if runeAt+clusterRunes > runeIdx {
			// runeIdx is at the boundary between runes within or before
			// this cluster. If this cluster is multi-rune we cannot position
			// the cursor inside it, so snap to the cluster's start column.
			// For a single-rune cluster, runeAt == runeIdx means we are at
			// the boundary before it (already accumulated); runeAt < runeIdx
			// means we are inside the cluster (snap to start).
			if clusterRunes > 1 && runeAt+clusterRunes >= runeIdx && runeAt < runeIdx {
				// Inside a multi-rune cluster: snap to the cluster's start.
				return col
			}
			// At a single-rune cluster boundary or past end: col is already
			// at the boundary.
			break
		}
		col += w
		runeAt += clusterRunes
		s = s[size:]
	}
	return col
}
