package formatter

import (
	"github.com/grindlemire/go-tui/internal/tuigen"
)

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

	// Emit ref={expr} if present (extracted from attributes during parsing)
	if let.Element.RefExpr != nil {
		p.buf.WriteString(" ref={")
		p.buf.WriteString(let.Element.RefExpr.Code)
		p.buf.WriteString("}")
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
