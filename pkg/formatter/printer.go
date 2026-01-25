package formatter

import (
	"fmt"
	"strings"

	"github.com/grindlemire/go-tui/pkg/tuigen"
)

// printer generates formatted .tui source code from an AST.
type printer struct {
	indent       string
	maxLineWidth int
	depth        int
	buf          strings.Builder
}

// newPrinter creates a new printer with the given settings.
func newPrinter(indent string, maxLineWidth int) *printer {
	return &printer{
		indent:       indent,
		maxLineWidth: maxLineWidth,
	}
}

// PrintFile formats an entire .tui file.
func (p *printer) PrintFile(file *tuigen.File) string {
	p.buf.Reset()

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
		p.newline()
	}

	p.depth--
	p.write(")")
	p.newline()
}

// printComponent outputs a component declaration.
func (p *printer) printComponent(comp *tuigen.Component) {
	p.write("@component ")
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
	p.newline()

	// Body
	p.depth++
	p.printBody(comp.Body)
	p.depth--

	p.write("}")
	p.newline()
}

// printBody outputs a list of body nodes.
func (p *printer) printBody(nodes []tuigen.Node) {
	for _, node := range nodes {
		p.printNode(node)
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
		p.writeIndent()
		p.write("{")
		p.write(n.Code)
		p.write("}")
		p.newline()
	case *tuigen.GoCode:
		p.writeIndent()
		p.write(n.Code)
		p.newline()
	case *tuigen.TextContent:
		p.writeIndent()
		p.write(n.Text)
		p.newline()
	case *tuigen.ChildrenSlot:
		p.writeIndent()
		p.write("{children...}")
		p.newline()
	}
}

// printElement outputs an element with its attributes and children.
func (p *printer) printElement(elem *tuigen.Element) {
	p.writeIndent()
	p.write("<")
	p.write(elem.Tag)

	// Attributes
	if len(elem.Attributes) > 0 {
		// Calculate approximate line length with all attrs inline
		inlineLen := p.currentIndentLen() + 1 + len(elem.Tag)
		for _, attr := range elem.Attributes {
			inlineLen += 1 + len(attr.Name) + 1 + p.attrValueLen(attr.Value)
		}

		// If closing > or /> adds to length
		if elem.SelfClose {
			inlineLen += 3 // " />"
		} else {
			inlineLen += 1 // ">"
		}

		// Decide whether to break attributes across lines
		multiLine := inlineLen > p.maxLineWidth && len(elem.Attributes) > 1

		for i, attr := range elem.Attributes {
			if multiLine && i > 0 {
				p.newline()
				p.writeIndent()
				// Align with first attribute
				p.write(strings.Repeat(" ", len(elem.Tag)+1))
			}
			p.write(" ")
			p.printAttribute(attr)
		}
	}

	if elem.SelfClose {
		p.write(" />")
		p.newline()
		return
	}

	p.write(">")

	// Check if we can render children inline (single simple child)
	if p.canInlineChildren(elem.Children) {
		p.printChildrenInline(elem.Children)
		p.write("</")
		p.write(elem.Tag)
		p.write(">")
		p.newline()
		return
	}

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
		p.write(v.Code)
		p.write("}")
	}
}

// attrValueLen returns the approximate rendered length of an attribute value.
func (p *printer) attrValueLen(val tuigen.Node) int {
	switch v := val.(type) {
	case *tuigen.StringLit:
		return len(v.Value) + 2 // quotes
	case *tuigen.IntLit:
		return len(fmt.Sprintf("{%d}", v.Value))
	case *tuigen.FloatLit:
		return len(fmt.Sprintf("{%g}", v.Value))
	case *tuigen.BoolLit:
		return len(fmt.Sprintf("{%t}", v.Value))
	case *tuigen.GoExpr:
		return len(v.Code) + 2 // braces
	}
	return 0
}

// canInlineChildren returns true if children can be rendered on one line.
func (p *printer) canInlineChildren(children []tuigen.Node) bool {
	// Empty children can be inlined (e.g., <box></box>)
	if len(children) == 0 {
		return true
	}

	if len(children) != 1 {
		return false
	}

	switch c := children[0].(type) {
	case *tuigen.GoExpr:
		// Inline simple expressions
		return !strings.Contains(c.Code, "\n") && len(c.Code) < 60
	case *tuigen.TextContent:
		return !strings.Contains(c.Text, "\n") && len(c.Text) < 60
	}

	return false
}

// printChildrenInline prints children on the same line (no newline).
func (p *printer) printChildrenInline(children []tuigen.Node) {
	for _, child := range children {
		switch c := child.(type) {
		case *tuigen.GoExpr:
			p.write("{")
			p.write(c.Code)
			p.write("}")
		case *tuigen.TextContent:
			p.write(c.Text)
		}
	}
}

// printForLoop outputs a @for loop.
func (p *printer) printForLoop(f *tuigen.ForLoop) {
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
	p.newline()

	// Body
	p.depth++
	p.printBody(f.Body)
	p.depth--

	p.writeIndent()
	p.write("}")
	p.newline()
}

// printIfStmt outputs an @if statement.
func (p *printer) printIfStmt(stmt *tuigen.IfStmt) {
	p.writeIndent()
	p.write("@if ")
	p.write(stmt.Condition)
	p.write(" {")
	p.newline()

	// Then branch
	p.depth++
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
				p.newline()

				p.depth++
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
			p.newline()

			p.depth++
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

	// Check if children can be inline
	if p.canInlineChildren(let.Element.Children) {
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
	p.writeIndent()
	p.write("@")
	p.write(call.Name)
	p.write("(")
	p.write(call.Args)
	p.write(")")

	if len(call.Children) > 0 {
		p.write(" {")
		p.newline()

		p.depth++
		p.printBody(call.Children)
		p.depth--

		p.writeIndent()
		p.write("}")
	}
	p.newline()
}

// printGoFunc outputs a top-level Go function.
func (p *printer) printGoFunc(fn *tuigen.GoFunc) {
	// Go functions are printed as-is since they're raw Go code
	p.write(fn.Code)
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

func (p *printer) currentIndentLen() int {
	return p.depth * len(p.indent)
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
