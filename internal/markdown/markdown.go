// Package markdown parses a small, well-scoped subset of markdown into a block
// tree for terminal rendering. It has no dependency on the tui package.
package markdown

import "strings"

type BlockKind int

const (
	KindParagraph BlockKind = iota
	KindHeading
	KindCodeFence
	KindTable
	KindList
	KindListItem
	KindBlockquote
)

// Inline is a styled run of text. Code spans set Code; links set Link (with Text
// as the label). Bold/Italic may combine.
type Inline struct {
	Text   string
	Bold   bool
	Italic bool
	Code   bool
	Link   string // non-empty => hyperlink target
}

// TableCell holds one cell's inline content.
type TableCell struct {
	Inline []Inline
}

// Block is one node in the document tree.
type Block struct {
	Kind     BlockKind
	Level    int           // heading level (1-6); unused otherwise
	Ordered  bool          // ordered list
	Lang     string        // code-fence info string
	Inline   []Inline      // leaf inline content (heading, paragraph, list item)
	Lines    []string      // raw code-fence lines
	Rows     [][]TableCell // table rows; row 0 is the header
	Children []Block       // nested blocks (list items, blockquote/list contents)
}

// Parse parses markdown source into a block tree.
func Parse(src string) []Block {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	return parseBlocks(lines)
}

func parseBlocks(lines []string) []Block {
	var blocks []Block
	for i := 0; i < len(lines); {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}
		switch {
		case isFence(line):
			b, next := parseFence(lines, i)
			blocks, i = append(blocks, b), next
		case isATXHeading(line):
			blocks, i = append(blocks, parseATX(line)), i+1
		case isBlockquote(line):
			b, next := parseBlockquote(lines, i)
			blocks, i = append(blocks, b), next
		case isListLine(line):
			b, next := parseList(lines, i, listIndent(line))
			blocks, i = append(blocks, b), next
		case isTableStart(lines, i):
			b, next := parseTable(lines, i)
			blocks, i = append(blocks, b), next
		default:
			b, next := parseParagraphOrSetext(lines, i)
			blocks, i = append(blocks, b), next
		}
	}
	return blocks
}

func isFence(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "```")
}

func parseFence(lines []string, i int) (Block, int) {
	info := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[i]), "```"))
	var body []string
	j := i + 1
	for j < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[j]), "```") {
		body = append(body, lines[j])
		j++
	}
	if j < len(lines) {
		j++ // consume closing fence
	}
	return Block{Kind: KindCodeFence, Lang: info, Lines: body}, j
}

func isATXHeading(line string) bool {
	t := strings.TrimLeft(line, " ")
	n := 0
	for n < len(t) && t[n] == '#' {
		n++
	}
	return n >= 1 && n <= 6 && n < len(t) && t[n] == ' '
}

func parseATX(line string) Block {
	t := strings.TrimLeft(line, " ")
	n := 0
	for n < len(t) && t[n] == '#' {
		n++
	}
	text := strings.TrimSpace(strings.TrimRight(strings.TrimSpace(t[n:]), "#"))
	return Block{Kind: KindHeading, Level: n, Inline: parseInline(text)}
}

func isSetextUnderline(s string, ch byte) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] != ch {
			return false
		}
	}
	return true
}

func parseParagraphOrSetext(lines []string, i int) (Block, int) {
	// Single-line setext: a text line followed by an all-'=' or all-'-' underline.
	if i+1 < len(lines) {
		if isSetextUnderline(lines[i+1], '=') {
			return Block{Kind: KindHeading, Level: 1, Inline: parseInline(strings.TrimSpace(lines[i]))}, i + 2
		}
		if isSetextUnderline(lines[i+1], '-') {
			return Block{Kind: KindHeading, Level: 2, Inline: parseInline(strings.TrimSpace(lines[i]))}, i + 2
		}
	}
	var parts []string
	j := i
	for j < len(lines) {
		l := lines[j]
		if strings.TrimSpace(l) == "" || isFence(l) || isATXHeading(l) || isBlockquote(l) || isListLine(l) {
			break
		}
		parts = append(parts, strings.TrimSpace(l))
		j++
	}
	return Block{Kind: KindParagraph, Inline: parseInline(strings.Join(parts, " "))}, j
}

func isTableStart(lines []string, i int) bool {
	return i+1 < len(lines) && strings.Contains(lines[i], "|") && isTableSeparator(lines[i+1])
}

func isTableSeparator(line string) bool {
	s := strings.TrimSpace(line)
	if s == "" || !strings.Contains(s, "-") {
		return false
	}
	for _, r := range s {
		if r != '-' && r != '|' && r != ':' && r != ' ' {
			return false
		}
	}
	return true
}

func parseTable(lines []string, i int) (Block, int) {
	rows := [][]TableCell{splitRow(lines[i])}
	j := i + 2 // skip header + separator
	for j < len(lines) && strings.TrimSpace(lines[j]) != "" && strings.Contains(lines[j], "|") {
		rows = append(rows, splitRow(lines[j]))
		j++
	}
	return Block{Kind: KindTable, Rows: rows}, j
}

func splitRow(line string) []TableCell {
	s := strings.TrimSpace(line)
	s = strings.TrimSuffix(strings.TrimPrefix(s, "|"), "|")
	parts := strings.Split(s, "|")
	cells := make([]TableCell, 0, len(parts))
	for _, p := range parts {
		cells = append(cells, TableCell{Inline: parseInline(strings.TrimSpace(p))})
	}
	return cells
}
