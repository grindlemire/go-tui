package formatter

import (
	"fmt"
	"strings"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

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
		// Multi-line: each attr on its own line, indented one tab deeper than element
		p.depth++
		// Emit ref={expr} first if present (extracted from attributes during parsing)
		if elem.RefExpr != nil {
			p.newline()
			p.writeIndent()
			p.write("ref={")
			p.write(elem.RefExpr.Code)
			p.write("}")
		}
		for _, attr := range elem.Attributes {
			p.newline()
			p.writeIndent()
			p.printAttribute(attr)
		}
		p.depth--
	} else {
		// Single-line: all attrs on same line
		// Emit ref={expr} first if present (extracted from attributes during parsing)
		if elem.RefExpr != nil {
			p.write(" ref={")
			p.write(elem.RefExpr.Code)
			p.write("}")
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
