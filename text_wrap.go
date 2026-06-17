package tui

import "strings"

// wrapText wraps text to fit within maxWidth terminal cells using word boundaries.
// It breaks at spaces, falling back to mid-character breaks when a single word
// exceeds maxWidth. Existing newlines in the text are preserved.
func wrapText(text string, maxWidth int) []string {
	if maxWidth < 1 {
		return []string{""}
	}
	if text == "" {
		return []string{""}
	}

	var result []string
	for paragraph := range strings.SplitSeq(text, "\n") {
		result = append(result, wrapParagraph(paragraph, maxWidth)...)
	}
	return result
}

// wrapParagraph wraps a single paragraph (no newlines) to maxWidth.
func wrapParagraph(text string, maxWidth int) []string {
	if text == "" {
		return []string{""}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var buf strings.Builder
	lineWidth := 0

	for _, word := range words {
		ww := stringWidth(word)

		if ww > maxWidth {
			// Word is longer than line — flush then break on cluster boundaries
			// so a grapheme cluster is never split mid-line.
			if lineWidth > 0 {
				lines = append(lines, buf.String())
				buf.Reset()
				lineWidth = 0
			}
			rest := word
			for len(rest) > 0 {
				cluster, cw, size := nextCluster(rest)
				if size == 0 {
					break
				}
				rest = rest[size:]
				if lineWidth+cw > maxWidth && lineWidth > 0 {
					lines = append(lines, buf.String())
					buf.Reset()
					lineWidth = 0
				}
				buf.WriteString(cluster)
				lineWidth += cw
			}
			continue
		}

		if lineWidth == 0 {
			// First word on line
			buf.WriteString(word)
			lineWidth = ww
		} else if lineWidth+1+ww <= maxWidth {
			// Fits with space
			buf.WriteByte(' ')
			buf.WriteString(word)
			lineWidth += 1 + ww
		} else {
			// Doesn't fit — start new line
			lines = append(lines, buf.String())
			buf.Reset()
			buf.WriteString(word)
			lineWidth = ww
		}
	}

	lines = append(lines, buf.String())
	return lines
}

// styledRune is one rune carrying the style and link of its source span.
type styledRune struct {
	r    rune
	st   Style
	link string
}

// styledCluster is one grapheme cluster carrying the style and link of its base
// rune's source span, with the cluster's display width.
type styledCluster struct {
	text  string
	width int
	st    Style
	link  string
}

// segmentStyledRunes segments a styled-rune stream into grapheme clusters. A
// cluster takes its base (first) rune's style and link, so a cluster whose base
// and combining mark fall in adjacent spans stays one styled unit.
func segmentStyledRunes(rs []styledRune) []styledCluster {
	if len(rs) == 0 {
		return nil
	}
	// Build the concatenated text and a parallel byte->index map to recover the
	// base rune's style/link for each cluster.
	var b strings.Builder
	starts := make([]int, 0, len(rs)) // byte offset where each styledRune begins
	for _, sr := range rs {
		starts = append(starts, b.Len())
		b.WriteRune(sr.r)
	}
	s := b.String()

	var out []styledCluster
	pos := 0
	ri := 0
	for pos < len(s) {
		cluster, w, size := nextCluster(s[pos:])
		if size == 0 {
			break
		}
		// Advance ri to the styledRune whose byte offset matches pos (the base).
		for ri < len(starts) && starts[ri] < pos {
			ri++
		}
		base := rs[0]
		if ri < len(rs) {
			base = rs[ri]
		}
		out = append(out, styledCluster{
			text:  cluster,
			width: w,
			st:    base.st,
			link:  base.link,
		})
		pos += size
	}
	return out
}

// segmentLineClusters flattens a wrapped rich-text line into a styled-cluster
// stream. Each cluster's style is the element base merged with its span style.
func segmentLineClusters(line []TextSpan, base Style) []styledCluster {
	var rs []styledRune
	for _, span := range line {
		st := mergeSpanStyle(base, span.Style)
		for _, r := range span.Text {
			rs = append(rs, styledRune{r: r, st: st, link: span.Link})
		}
	}
	return segmentStyledRunes(rs)
}

// wrapSpans wraps styled spans to maxWidth using word boundaries, mirroring
// wrapParagraph. Words are delimited by actual whitespace in the concatenated
// text (a span boundary is NOT a word boundary), so a single word may mix styles
// (e.g. "a**b**c" is one word "abc"). The flattened styled-rune stream is
// segmented into grapheme clusters, so a cluster is never split at a line break
// or at a span boundary; each cluster keeps its base rune's style/link. Adjacent
// same-style clusters on a line are merged into one segment. Newlines start new
// lines.
func wrapSpans(spans []TextSpan, maxWidth int) [][]TextSpan {
	if maxWidth < 1 {
		return [][]TextSpan{{}}
	}

	var lines [][]TextSpan
	var cur []TextSpan // current line segments
	lineWidth := 0

	flush := func() {
		lines = append(lines, cur)
		cur = nil
		lineWidth = 0
	}
	// emit appends styled clusters to cur, merging same-style into the last segment.
	emit := func(cs []styledCluster) {
		for _, sc := range cs {
			if n := len(cur); n > 0 && cur[n-1].Style == sc.st && cur[n-1].Link == sc.link {
				cur[n-1].Text += sc.text
			} else {
				cur = append(cur, TextSpan{Text: sc.text, Style: sc.st, Link: sc.link})
			}
		}
	}
	// emitSpace appends a single separator space. Separators normally carry the
	// default (zero) style so they inherit the element's base style at render time
	// rather than the adjacent word's style (e.g. the space after a bold word is
	// not itself bold). The exception is a separator interior to one link run: it
	// carries that link's style and target so the hyperlink (and its underline)
	// stays continuous across the gap. st/link give the link context, or zero.
	emitSpace := func(st Style, link string) {
		if n := len(cur); n > 0 && cur[n-1].Style == st && cur[n-1].Link == link {
			cur[n-1].Text += " "
		} else {
			cur = append(cur, TextSpan{Text: " ", Style: st, Link: link})
		}
	}

	var word []styledRune
	placeWord := func() {
		if len(word) == 0 {
			return
		}
		clusters := segmentStyledRunes(word)
		wordWidth := 0
		for _, c := range clusters {
			wordWidth += c.width
		}
		if wordWidth > maxWidth {
			// Word longer than the line: flush current, then hard-break by cluster.
			if lineWidth > 0 {
				flush()
			}
			for _, c := range clusters {
				if lineWidth+c.width > maxWidth && lineWidth > 0 {
					flush()
				}
				emit([]styledCluster{c})
				lineWidth += c.width
			}
			word = word[:0]
			return
		}
		switch {
		case lineWidth == 0:
			emit(clusters)
			lineWidth = wordWidth
		case lineWidth+1+wordWidth <= maxWidth:
			// If the separator sits between two words of the same link, give it
			// that link's style+target so the link renders as one continuous run.
			var sepSt Style
			var sepLink string
			if len(clusters) > 0 && clusters[0].link != "" {
				if n := len(cur); n > 0 && cur[n-1].Link == clusters[0].link {
					sepSt = clusters[0].st
					sepLink = clusters[0].link
				}
			}
			emitSpace(sepSt, sepLink)
			emit(clusters)
			lineWidth += 1 + wordWidth
		default:
			flush()
			emit(clusters)
			lineWidth = wordWidth
		}
		word = word[:0]
	}

	for _, sp := range spans {
		for _, r := range sp.Text {
			switch r {
			case '\n':
				placeWord()
				flush()
			case ' ', '\t', '\r', '\v', '\f':
				// All non-newline whitespace separates words but does not break
				// the line; only '\n' starts a new line.
				placeWord()
			default:
				word = append(word, styledRune{r: r, st: sp.Style, link: sp.Link})
			}
		}
	}
	placeWord()
	flush()
	return lines
}
