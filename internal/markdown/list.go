package markdown

import "strings"

func isBlockquote(line string) bool {
	return strings.HasPrefix(strings.TrimLeft(line, " "), ">")
}

func parseBlockquote(lines []string, i int) (Block, int) {
	var inner []string
	j := i
	for j < len(lines) && isBlockquote(lines[j]) {
		l := strings.TrimPrefix(strings.TrimLeft(lines[j], " "), ">")
		inner = append(inner, strings.TrimPrefix(l, " "))
		j++
	}
	return Block{Kind: KindBlockquote, Children: parseBlocks(inner)}, j
}

func listIndent(line string) int {
	n := 0
	for n < len(line) && line[n] == ' ' {
		n++
	}
	return n
}

func isListLine(line string) bool {
	t := strings.TrimLeft(line, " ")
	if strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ") || strings.HasPrefix(t, "+ ") {
		return true
	}
	k := 0
	for k < len(t) && t[k] >= '0' && t[k] <= '9' {
		k++
	}
	return k > 0 && k+1 < len(t) && t[k] == '.' && t[k+1] == ' '
}

// listMarkerInfo reports whether the item is ordered and the byte offset where
// its content starts (after the marker and one space).
func listMarkerInfo(line string) (ordered bool, contentOffset int) {
	t := strings.TrimLeft(line, " ")
	lead := len(line) - len(t)
	if strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ") || strings.HasPrefix(t, "+ ") {
		return false, lead + 2
	}
	k := 0
	for k < len(t) && t[k] >= '0' && t[k] <= '9' {
		k++
	}
	return true, lead + k + 2 // digits + ". "
}

// parseList consumes consecutive list lines at the given indent into a List of
// ListItems. Lines indented further attach as a nested List on the previous item.
func parseList(lines []string, i, indent int) (Block, int) {
	ordered, _ := listMarkerInfo(lines[i])
	list := Block{Kind: KindList, Ordered: ordered}
	j := i
	for j < len(lines) {
		if strings.TrimSpace(lines[j]) == "" || !isListLine(lines[j]) {
			break
		}
		ind := listIndent(lines[j])
		if ind < indent {
			break // belongs to an outer list
		}
		if ind > indent {
			child, next := parseList(lines, j, ind)
			if n := len(list.Children); n > 0 {
				list.Children[n-1].Children = append(list.Children[n-1].Children, child)
			} else {
				list.Children = append(list.Children, Block{Kind: KindListItem, Children: []Block{child}})
			}
			j = next
			continue
		}
		_, off := listMarkerInfo(lines[j])
		text := strings.TrimSpace(lines[j][off:])
		list.Children = append(list.Children, Block{Kind: KindListItem, Inline: parseInline(text)})
		j++
	}
	return list, j
}
