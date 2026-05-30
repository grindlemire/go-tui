package markdown

// parseInline scans text into styled segments: **bold**/__bold__, *italic*/_italic_,
// `code`, and [label](url). Unmatched markers are literal. Code spans and link
// labels are not further parsed.
func parseInline(s string) []Inline {
	var out []Inline
	var buf []rune
	bold, italic := false, false

	flush := func() {
		if len(buf) > 0 {
			out = append(out, Inline{Text: string(buf), Bold: bold, Italic: italic})
			buf = buf[:0]
		}
	}

	rs := []rune(s)
	for i := 0; i < len(rs); {
		switch r := rs[i]; {
		case r == '`':
			j := i + 1
			for j < len(rs) && rs[j] != '`' {
				j++
			}
			if j < len(rs) {
				flush()
				out = append(out, Inline{Text: string(rs[i+1 : j]), Code: true})
				i = j + 1
				continue
			}
			buf = append(buf, r)
			i++
		case r == '[':
			if label, url, n, ok := parseLink(rs[i:]); ok {
				flush()
				out = append(out, Inline{Text: label, Link: url})
				i += n
				continue
			}
			buf = append(buf, r)
			i++
		case r == '*' || r == '_':
			// Only toggle emphasis when a matching closer exists ahead; an
			// unmatched delimiter (e.g. "see **docs") stays literal so a stray
			// marker doesn't style the rest of the text.
			if i+1 < len(rs) && rs[i+1] == r {
				if bold { // closing an open bold run
					flush()
					bold = false
					i += 2
					continue
				}
				if hasDelimCloser(rs, i+2, r, true) {
					flush()
					bold = true
					i += 2
					continue
				}
				buf = append(buf, r, r)
				i += 2
				continue
			}
			if italic { // closing an open italic run
				flush()
				italic = false
				i++
				continue
			}
			if hasDelimCloser(rs, i+1, r, false) {
				flush()
				italic = true
				i++
				continue
			}
			buf = append(buf, r)
			i++
		default:
			buf = append(buf, r)
			i++
		}
	}
	flush()
	return out
}

// hasDelimCloser reports whether a matching emphasis delimiter appears at or
// after start: a doubled run (e.g. "**") when double is true, otherwise a single
// occurrence of r.
func hasDelimCloser(rs []rune, start int, r rune, double bool) bool {
	for j := start; j < len(rs); j++ {
		if rs[j] != r {
			continue
		}
		if double {
			if j+1 < len(rs) && rs[j+1] == r {
				return true
			}
			continue
		}
		// Single-rune closer: a doubled run (e.g. "**") is not a single
		// closer, and neither of its two characters may serve as one. Skip
		// both so a closing "**" can't be mistaken for an italic closer.
		if j+1 < len(rs) && rs[j+1] == r {
			j++
			continue
		}
		return true
	}
	return false
}

// parseLink parses [label](url) at rs[0]=='['. Returns label, url, runes consumed, ok.
func parseLink(rs []rune) (label, url string, n int, ok bool) {
	if len(rs) == 0 || rs[0] != '[' {
		return "", "", 0, false
	}
	closeIdx := -1
	for i := 1; i < len(rs); i++ {
		if rs[i] == ']' {
			closeIdx = i
			break
		}
	}
	if closeIdx < 0 || closeIdx+1 >= len(rs) || rs[closeIdx+1] != '(' {
		return "", "", 0, false
	}
	parenIdx := -1
	for i := closeIdx + 2; i < len(rs); i++ {
		if rs[i] == ')' {
			parenIdx = i
			break
		}
	}
	if parenIdx < 0 {
		return "", "", 0, false
	}
	return string(rs[1:closeIdx]), string(rs[closeIdx+2 : parenIdx]), parenIdx + 1, true
}
