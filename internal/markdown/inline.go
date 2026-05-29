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
			flush()
			if i+1 < len(rs) && rs[i+1] == r {
				bold = !bold
				i += 2
			} else {
				italic = !italic
				i++
			}
		default:
			buf = append(buf, r)
			i++
		}
	}
	flush()
	return out
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
