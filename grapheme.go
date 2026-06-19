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
//
// NextCluster is the exported wrapper. It behaves identically.
func NextCluster(s string) (cluster string, width, size int) { return nextCluster(s) }
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
			w = clusterExtendUpdateWidth(r, r0, w)
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

// clusterExtendUpdateWidth returns the updated width when a grapheme-extend
// rune attaches to the base rune. VS16 forces wide (2), VS15 can narrow an
// emoji base (but not CJK) to 1, and all other extenders preserve the width.
//
// This is shared between clusterAdvance and nextClusterWidth so both
// state machines agree on width after a combining mark, VS16, or VS15.
func clusterExtendUpdateWidth(extendRune, baseRune rune, currentWidth int) int {
	if isVS16(extendRune) {
		return 2
	}
	if extendRune == 0xFE0E && baseRuneWidth(baseRune) == 2 && inRuneRanges(baseRune, emojiWideRanges) {
		return 1
	}
	return currentWidth
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

// ClusterCount returns the number of user-perceived characters (grapheme
// clusters) in s. This accounts for multi-rune clusters — flags, ZWJ emoji
// families, and decomposed accents each count as one cluster.
func ClusterCount(s string) int { return clusterCount(s) }

// clusterEnd returns the rune index at the end of the cluster that contains the
// rune at clusterStartRuneIdx in s. The target may be at a cluster boundary
// (normal case after clampCursorPos) or inside a multi-rune cluster (insertChar
// after inserting a combining mark).
// Unlike the O(N) clusterRuneStarts-based approach, this walks clusters
// incrementally and stops as soon as it passes the target, making it O(pos).
func clusterEnd(s string, clusterStartRuneIdx int) int {
	// Return the rune index at the end of the cluster that contains the target
	// rune position. The target may be at a cluster boundary (normal case after
	// clampCursorPos) or inside a multi-rune cluster (insertChar after inserting
	// a combining mark).
	runeAt := 0
	for len(s) > 0 {
		_, _, size := nextCluster(s)
		if size == 0 {
			break
		}
		clusterRunes := utf8.RuneCountInString(s[:size])
		if runeAt+clusterRunes > clusterStartRuneIdx {
			// This cluster contains or starts at the target. Return its end.
			return runeAt + clusterRunes
		}
		runeAt += clusterRunes
		s = s[size:]
	}
	return runeAt
}

// snapRuneToClusterStart snaps a rune index down to the nearest cluster start at
// or before it (clamped to [0, total]). Walks clusters incrementally and stops
// as soon as it passes the target, making it O(pos) instead of O(N).
func snapRuneToClusterStart(s string, runeIdx int) int {
	if len(s) == 0 || runeIdx <= 0 {
		return 0
	}
	runeAt := 0
	lastStart := 0
	for len(s) > 0 {
		_, _, size := nextCluster(s)
		if size == 0 {
			break
		}
		clusterRunes := utf8.RuneCountInString(s[:size])
		if runeAt+clusterRunes > runeIdx {
			// runeIdx is inside this cluster: snap to its start.
			return lastStart
		}
		runeAt += clusterRunes
		lastStart = runeAt
		s = s[size:]
	}
	return runeAt
}

// runeIndexToDisplayCol returns the display column at the given rune index,
// summing cluster widths from the start of s. The index is snapped to a cluster
// boundary first so a position inside a cluster reports that cluster's start col.
// Walks clusters incrementally and stops at the target, making it O(pos).
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
			// runeIdx is at or inside this cluster. Whether it is exactly at
			// the cluster start (single-rune cluster) or somewhere in the
			// interior (multi-rune cluster), col is already the cluster's start
			// column — snap to it.
			break
		}
		col += w
		runeAt += clusterRunes
		s = s[size:]
	}
	return col
}
