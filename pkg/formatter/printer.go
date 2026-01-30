package formatter

import (
	"fmt"
	"strings"

	"github.com/grindlemire/go-tui/pkg/tuigen"
)

// printer generates formatted .gsx source code from an AST.
type printer struct {
	indent string
	depth  int
	buf    strings.Builder
}

// newPrinter creates a new printer with the given settings.
func newPrinter(indent string) *printer {
	return &printer{
		indent: indent,
	}
}

// PrintFile formats an entire .gsx file.
func (p *printer) PrintFile(file *tuigen.File) string {
	p.buf.Reset()

	// Leading comments before package declaration
	p.printLeadingComments(file.LeadingComments)

	// Package declaration
	p.printPackage(file.Package)
	p.newline()

	// Imports
	if len(file.Imports) > 0 {
		p.printImports(file.Imports)
		p.newline()
	}

	// Components and functions
	for i, comp := range file.Components {
		if i > 0 {
			p.newline()
		}
		p.printComponent(comp)
	}

	for i, fn := range file.Funcs {
		if i > 0 || len(file.Components) > 0 {
			p.newline()
		}
		p.printGoFunc(fn)
	}

	// Orphan comments at end of file
	p.printOrphanComments(file.OrphanComments)

	return p.buf.String()
}

// printPackage outputs the package declaration.
func (p *printer) printPackage(name string) {
	p.write("package ")
	p.write(name)
	p.newline()
}

// printImports outputs import declarations.
func (p *printer) printImports(imports []tuigen.Import) {
	if len(imports) == 1 {
		// Single import - use inline form
		imp := imports[0]
		p.write("import ")
		if imp.Alias != "" {
			p.write(imp.Alias)
			p.write(" ")
		}
		p.write(`"`)
		p.write(imp.Path)
		p.write(`"`)
		p.printTrailingComment(imp.TrailingComments)
		p.newline()
		return
	}

	// Multiple imports - use grouped form
	p.write("import (")
	p.newline()
	p.depth++

	for _, imp := range imports {
		p.writeIndent()
		if imp.Alias != "" {
			p.write(imp.Alias)
			p.write(" ")
		}
		p.write(`"`)
		p.write(imp.Path)
		p.write(`"`)
		p.printTrailingComment(imp.TrailingComments)
		p.newline()
	}

	p.depth--
	p.write(")")
	p.newline()
}

// printComponent outputs a component declaration.
// Components are templ functions with no return type: templ Name(params) { ... }
func (p *printer) printComponent(comp *tuigen.Component) {
	// Leading comments (doc comments)
	p.printLeadingComments(comp.LeadingComments)

	p.write("templ ")
	p.write(comp.Name)
	p.write("(")

	// Parameters
	for i, param := range comp.Params {
		if i > 0 {
			p.write(", ")
		}
		p.write(param.Name)
		p.write(" ")
		p.write(param.Type)
	}

	p.write(") {")
	p.printTrailingComment(comp.TrailingComments)
	p.newline()

	// Body
	p.depth++
	p.printOrphanComments(comp.OrphanComments)
	p.printBody(comp.Body)
	p.depth--

	p.write("}")
	p.newline()
}

// printBody outputs a list of body nodes, preserving blank lines between them.
func (p *printer) printBody(nodes []tuigen.Node) {
	for _, node := range nodes {
		if hasBlankLineBefore(node) {
			p.newline()
		}
		p.printNode(node)
	}
}

// hasBlankLineBefore returns true if the node had a blank line before it in the source.
func hasBlankLineBefore(node tuigen.Node) bool {
	switch n := node.(type) {
	case *tuigen.Element:
		return n.BlankLineBefore
	case *tuigen.GoExpr:
		return n.BlankLineBefore
	case *tuigen.TextContent:
		return n.BlankLineBefore
	case *tuigen.LetBinding:
		return n.BlankLineBefore
	case *tuigen.ForLoop:
		return n.BlankLineBefore
	case *tuigen.IfStmt:
		return n.BlankLineBefore
	case *tuigen.ComponentCall:
		return n.BlankLineBefore
	case *tuigen.ChildrenSlot:
		return n.BlankLineBefore
	default:
		return false
	}
}

// printNode outputs a single AST node.
func (p *printer) printNode(node tuigen.Node) {
	switch n := node.(type) {
	case *tuigen.Element:
		p.printElement(n)
	case *tuigen.ForLoop:
		p.printForLoop(n)
	case *tuigen.IfStmt:
		p.printIfStmt(n)
	case *tuigen.LetBinding:
		p.printLetBinding(n)
	case *tuigen.ComponentCall:
		p.printComponentCall(n)
	case *tuigen.GoExpr:
		p.printLeadingComments(n.LeadingComments)
		p.writeIndent()
		p.write("{")
		p.write(formatInlineBlockComments(n.Code))
		p.write("}")
		p.printTrailingComment(n.TrailingComments)
		p.newline()
	case *tuigen.GoCode:
		p.printLeadingComments(n.LeadingComments)
		p.writeIndent()
		p.write(formatInlineBlockComments(n.Code))
		p.printTrailingComment(n.TrailingComments)
		p.newline()
	case *tuigen.TextContent:
		p.writeIndent()
		p.write(n.Text)
		p.newline()
	case *tuigen.ChildrenSlot:
		p.printLeadingComments(n.LeadingComments)
		p.writeIndent()
		p.write("{children...}")
		p.printTrailingComment(n.TrailingComments)
		p.newline()
	}
}

// printElement outputs an element with its attributes and children.
func (p *printer) printElement(elem *tuigen.Element) {
	// Leading comments
	p.printLeadingComments(elem.LeadingComments)

	p.writeIndent()
	p.write("<")
	p.write(elem.Tag)

	// Determine multi-line: either user had attrs on separate lines, or > was on its own line
	multiLine := elem.MultiLineAttrs || elem.ClosingBracketNewLine

	if multiLine {
		// Multi-line: each attr/ref on its own line, indented one tab deeper than element
		p.depth++
		if elem.NamedRef != "" {
			p.newline()
			p.writeIndent()
			p.write("#")
			p.write(elem.NamedRef)
		}
		for _, attr := range elem.Attributes {
			p.newline()
			p.writeIndent()
			p.printAttribute(attr)
		}
		p.depth--
	} else {
		// Single-line: all attrs/ref on same line
		if elem.NamedRef != "" {
			p.write(" #")
			p.write(elem.NamedRef)
		}
		for _, attr := range elem.Attributes {
			p.write(" ")
			p.printAttribute(attr)
		}
	}

	if elem.SelfClose {
		if elem.ClosingBracketNewLine {
			p.newline()
			p.writeIndent()
			p.write("/>")
		} else {
			p.write(" />")
		}
		p.printTrailingComment(elem.TrailingComments)
		p.newline()
		return
	}

	if elem.ClosingBracketNewLine {
		p.newline()
		p.writeIndent()
		p.write(">")
	} else {
		p.write(">")
	}

	// Check if children should be rendered inline (preserving user's source layout)
	if elem.InlineChildren && p.canStructurallyInline(elem.Children) {
		p.printChildrenInline(elem.Children)
		p.write("</")
		p.write(elem.Tag)
		p.write(">")
		p.printTrailingComment(elem.TrailingComments)
		p.newline()
		return
	}

	// Trailing comment after opening tag
	p.printTrailingComment(elem.TrailingComments)

	// Multi-line children
	p.newline()
	p.depth++
	p.printBody(elem.Children)
	p.depth--
	p.writeIndent()
	p.write("</")
	p.write(elem.Tag)
	p.write(">")
	p.newline()
}

// printAttribute outputs a single attribute.
func (p *printer) printAttribute(attr *tuigen.Attribute) {
	p.write(attr.Name)

	// Check for boolean shorthand (attr with true value and no explicit =true)
	if bl, ok := attr.Value.(*tuigen.BoolLit); ok && bl.Value {
		// Could be shorthand boolean - check if it was originally shorthand
		// For now, always use explicit form for consistency
		p.write("={true}")
		return
	}

	p.write("=")
	p.printAttrValue(attr.Value)
}

// printAttrValue outputs an attribute value.
func (p *printer) printAttrValue(val tuigen.Node) {
	switch v := val.(type) {
	case *tuigen.StringLit:
		p.write(`"`)
		p.write(escapeString(v.Value))
		p.write(`"`)
	case *tuigen.IntLit:
		p.write(fmt.Sprintf("{%d}", v.Value))
	case *tuigen.FloatLit:
		p.write(fmt.Sprintf("{%g}", v.Value))
	case *tuigen.BoolLit:
		p.write(fmt.Sprintf("{%t}", v.Value))
	case *tuigen.GoExpr:
		p.write("{")
		p.write(formatInlineBlockComments(v.Code))
		p.write("}")
	}
}

// canStructurallyInline returns true if the children types support inline rendering.
// This checks structural compatibility (types and content), not user layout preference.
func (p *printer) canStructurallyInline(children []tuigen.Node) bool {
	if len(children) == 0 {
		return true
	}

	for _, child := range children {
		switch c := child.(type) {
		case *tuigen.GoExpr:
			if strings.Contains(c.Code, "\n") {
				return false
			}
		case *tuigen.TextContent:
			if strings.Contains(c.Text, "\n") {
				return false
			}
		default:
			return false
		}
	}

	return true
}

// printChildrenInline prints children on the same line (no newline).
func (p *printer) printChildrenInline(children []tuigen.Node) {
	for _, child := range children {
		switch c := child.(type) {
		case *tuigen.GoExpr:
			p.write("{")
			p.write(formatInlineBlockComments(c.Code))
			p.write("}")
		case *tuigen.TextContent:
			p.write(c.Text)
		}
	}
}

// printForLoop outputs a @for loop.
func (p *printer) printForLoop(f *tuigen.ForLoop) {
	// Leading comments
	p.printLeadingComments(f.LeadingComments)

	p.writeIndent()
	p.write("@for ")

	// Loop variables
	if f.Index != "" {
		p.write(f.Index)
		p.write(", ")
	}
	p.write(f.Value)
	p.write(" := range ")
	p.write(f.Iterable)
	p.write(" {")
	p.printTrailingComment(f.TrailingComments)
	p.newline()

	// Body
	p.depth++
	p.printOrphanComments(f.OrphanComments)
	p.printBody(f.Body)
	p.depth--

	p.writeIndent()
	p.write("}")
	p.newline()
}

// printIfStmt outputs an @if statement.
func (p *printer) printIfStmt(stmt *tuigen.IfStmt) {
	// Leading comments
	p.printLeadingComments(stmt.LeadingComments)

	p.writeIndent()
	p.write("@if ")
	p.write(stmt.Condition)
	p.write(" {")
	p.printTrailingComment(stmt.TrailingComments)
	p.newline()

	// Then branch
	p.depth++
	p.printOrphanComments(stmt.OrphanComments)
	p.printBody(stmt.Then)
	p.depth--

	// Else branch
	if len(stmt.Else) > 0 {
		p.writeIndent()
		p.write("} @else ")

		// Check for else-if chain
		if len(stmt.Else) == 1 {
			if elseIf, ok := stmt.Else[0].(*tuigen.IfStmt); ok {
				// Print else-if without extra indent
				p.write("@if ")
				p.write(elseIf.Condition)
				p.write(" {")
				p.printTrailingComment(elseIf.TrailingComments)
				p.newline()

				p.depth++
				p.printOrphanComments(elseIf.OrphanComments)
				p.printBody(elseIf.Then)
				p.depth--

				if len(elseIf.Else) > 0 {
					p.printElseBranch(elseIf.Else)
				} else {
					p.writeIndent()
					p.write("}")
					p.newline()
				}
				return
			}
		}

		// Regular else
		p.write("{")
		p.newline()
		p.depth++
		p.printBody(stmt.Else)
		p.depth--
		p.writeIndent()
		p.write("}")
		p.newline()
	} else {
		p.writeIndent()
		p.write("}")
		p.newline()
	}
}

// printElseBranch handles recursive else-if chains.
func (p *printer) printElseBranch(nodes []tuigen.Node) {
	p.writeIndent()
	p.write("} @else ")

	if len(nodes) == 1 {
		if elseIf, ok := nodes[0].(*tuigen.IfStmt); ok {
			p.write("@if ")
			p.write(elseIf.Condition)
			p.write(" {")
			p.printTrailingComment(elseIf.TrailingComments)
			p.newline()

			p.depth++
			p.printOrphanComments(elseIf.OrphanComments)
			p.printBody(elseIf.Then)
			p.depth--

			if len(elseIf.Else) > 0 {
				p.printElseBranch(elseIf.Else)
			} else {
				p.writeIndent()
				p.write("}")
				p.newline()
			}
			return
		}
	}

	p.write("{")
	p.newline()
	p.depth++
	p.printBody(nodes)
	p.depth--
	p.writeIndent()
	p.write("}")
	p.newline()
}

// printLetBinding outputs a @let binding.
func (p *printer) printLetBinding(let *tuigen.LetBinding) {
	// Leading comments
	p.printLeadingComments(let.LeadingComments)

	p.writeIndent()
	p.write("@let ")
	p.write(let.Name)
	p.write(" = ")

	// Reset indent for the element since we're already indented
	savedDepth := p.depth
	p.depth = 0

	// Print element inline or multi-line based on complexity
	p.buf.WriteString("<")
	p.buf.WriteString(let.Element.Tag)

	// Named ref (e.g., #Content)
	if let.Element.NamedRef != "" {
		p.buf.WriteString(" #")
		p.buf.WriteString(let.Element.NamedRef)
	}

	for _, attr := range let.Element.Attributes {
		p.buf.WriteString(" ")
		p.printAttribute(attr)
	}

	if let.Element.SelfClose {
		p.buf.WriteString(" />")
		p.newline()
		p.depth = savedDepth
		return
	}

	p.buf.WriteString(">")

	// Check if children should be rendered inline (preserving user's source layout)
	if let.Element.InlineChildren && p.canStructurallyInline(let.Element.Children) {
		p.printChildrenInline(let.Element.Children)
		p.buf.WriteString("</")
		p.buf.WriteString(let.Element.Tag)
		p.buf.WriteString(">")
		p.newline()
		p.depth = savedDepth
		return
	}

	// Multi-line children
	p.newline()
	p.depth = savedDepth + 1
	p.printBody(let.Element.Children)
	p.depth = savedDepth
	p.writeIndent()
	p.buf.WriteString("</")
	p.buf.WriteString(let.Element.Tag)
	p.buf.WriteString(">")
	p.newline()
}

// printComponentCall outputs a component call.
func (p *printer) printComponentCall(call *tuigen.ComponentCall) {
	// Leading comments
	p.printLeadingComments(call.LeadingComments)

	p.writeIndent()
	p.write("@")
	p.write(call.Name)
	p.write("(")
	p.write(formatInlineBlockComments(call.Args))
	p.write(")")

	if len(call.Children) > 0 {
		p.write(" {")
		p.printTrailingComment(call.TrailingComments)
		p.newline()

		p.depth++
		p.printBody(call.Children)
		p.depth--

		p.writeIndent()
		p.write("}")
	} else {
		p.printTrailingComment(call.TrailingComments)
	}
	p.newline()
}

// printGoFunc outputs a top-level Go function.
func (p *printer) printGoFunc(fn *tuigen.GoFunc) {
	// Leading comments
	p.printLeadingComments(fn.LeadingComments)

	// Go functions are printed as-is since they're raw Go code
	p.write(fn.Code)
	p.printTrailingComment(fn.TrailingComments)
	p.newline()
}

// Helper methods

func (p *printer) write(s string) {
	p.buf.WriteString(s)
}

func (p *printer) newline() {
	p.buf.WriteByte('\n')
}

func (p *printer) writeIndent() {
	for i := 0; i < p.depth; i++ {
		p.buf.WriteString(p.indent)
	}
}


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
