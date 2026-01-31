package formatter

import (
	"strings"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

// escapeString escapes special characters in a string for output.
func escapeString(s string) string {
	var buf strings.Builder
	for _, r := range s {
		switch r {
		case '\n':
			buf.WriteString(`\n`)
		case '\t':
			buf.WriteString(`\t`)
		case '\r':
			buf.WriteString(`\r`)
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// Comment printing helpers

// formatBlockComment formats a block comment with proper spacing.
// Single-line: /* text */ -> /* text */ (ensures spaces around text)
// Multi-line: formats with /* and */ on their own lines
func formatBlockComment(text string) string {
	// Must start with /* and end with */
	if !strings.HasPrefix(text, "/*") || !strings.HasSuffix(text, "*/") {
		return text
	}

	// Extract the content between /* and */
	content := text[2 : len(text)-2]

	// Collect all non-empty content lines
	var contentLines []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			contentLines = append(contentLines, trimmed)
		}
	}

	// Empty comment
	if len(contentLines) == 0 {
		return "/* */"
	}

	// Single line of content: use inline format
	if len(contentLines) == 1 {
		return "/* " + contentLines[0] + " */"
	}

	// Multi-line content: format with /* and */ on their own lines
	var result strings.Builder
	result.WriteString("/*\n")
	for _, line := range contentLines {
		result.WriteString(line)
		result.WriteString("\n")
	}
	result.WriteString("*/")
	return result.String()
}

// formatLineComment formats a line comment with proper spacing.
// Ensures a space after // if not already present.
func formatLineComment(text string) string {
	if !strings.HasPrefix(text, "//") {
		return text
	}

	// Get content after //
	content := text[2:]

	// If empty or already starts with space, return as-is
	if content == "" || content[0] == ' ' || content[0] == '\t' {
		return text
	}

	// Add space after //
	return "// " + content
}

// formatComment formats a comment, handling both line and block comments.
func formatComment(c *tuigen.Comment) string {
	if c.IsBlock {
		return formatBlockComment(c.Text)
	}
	return formatLineComment(c.Text)
}

// formatInlineBlockComments formats any block comments within Go code.
// This handles cases like: fmt.Sprintf("> %s", /* ItemList item */ item)
func formatInlineBlockComments(code string) string {
	var result strings.Builder
	i := 0

	for i < len(code) {
		// Check for block comment start
		if i+1 < len(code) && code[i] == '/' && code[i+1] == '*' {
			// Find the end of the block comment
			start := i
			i += 2
			for i+1 < len(code) && !(code[i] == '*' && code[i+1] == '/') {
				i++
			}
			if i+1 < len(code) {
				i += 2 // skip */
			}

			// Extract and format the block comment
			commentText := code[start:i]
			result.WriteString(formatBlockComment(commentText))
			continue
		}

		// Check for string literal (skip to avoid formatting comments inside strings)
		if code[i] == '"' {
			result.WriteByte(code[i])
			i++
			for i < len(code) && code[i] != '"' {
				if code[i] == '\\' && i+1 < len(code) {
					result.WriteByte(code[i])
					i++
				}
				if i < len(code) {
					result.WriteByte(code[i])
					i++
				}
			}
			if i < len(code) {
				result.WriteByte(code[i])
				i++
			}
			continue
		}

		// Check for raw string literal
		if code[i] == '`' {
			result.WriteByte(code[i])
			i++
			for i < len(code) && code[i] != '`' {
				result.WriteByte(code[i])
				i++
			}
			if i < len(code) {
				result.WriteByte(code[i])
				i++
			}
			continue
		}

		// Regular character
		result.WriteByte(code[i])
		i++
	}

	return result.String()
}

// printCommentGroup outputs a comment group with proper indentation.
// Each comment in the group is printed on its own line.
// Respects BlankLineBefore to preserve blank line separation between comment groups.
func (p *printer) printCommentGroup(cg *tuigen.CommentGroup) {
	if cg == nil || len(cg.List) == 0 {
		return
	}
	for _, c := range cg.List {
		if c.BlankLineBefore {
			p.newline()
		}
		p.writeIndent()
		p.write(formatComment(c))
		p.newline()
	}
}

// printLeadingComments outputs leading comments (before a node).
// Comments are printed with proper indentation, each on its own line.
// Respects BlankLineBefore to preserve blank line separation between comment groups.
func (p *printer) printLeadingComments(cg *tuigen.CommentGroup) {
	if cg == nil || len(cg.List) == 0 {
		return
	}
	for _, c := range cg.List {
		if c.BlankLineBefore {
			p.newline()
		}
		p.writeIndent()
		p.write(formatComment(c))
		p.newline()
	}
}

// printTrailingComment outputs a trailing comment (on same line as node).
// Prints with leading spaces, no newline (caller handles newline).
func (p *printer) printTrailingComment(cg *tuigen.CommentGroup) {
	if cg == nil || len(cg.List) == 0 {
		return
	}
	// Only print the first comment as trailing (others would be on next lines)
	p.write("  ")
	p.write(formatComment(cg.List[0]))
}

// printOrphanComments outputs orphan comments (not attached to any node).
// Each comment group is printed with proper indentation, with blank lines between groups.
func (p *printer) printOrphanComments(groups []*tuigen.CommentGroup) {
	if len(groups) == 0 {
		return
	}
	for i, cg := range groups {
		if i > 0 {
			// Add blank line between comment groups
			p.newline()
		}
		p.printCommentGroup(cg)
	}
}
