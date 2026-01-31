package formatter

import (
	"strings"

	"github.com/grindlemire/go-tui/internal/tuigen"
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
