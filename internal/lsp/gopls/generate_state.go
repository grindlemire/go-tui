package gopls

import (
	"fmt"
	"strings"

	"github.com/grindlemire/go-tui/internal/lsp/log"
	"github.com/grindlemire/go-tui/internal/tuigen"
)

// emitStateVarDeclarations scans a component body for tui.NewState(...) declarations
// and emits corresponding Go variable declarations so gopls can resolve state types.
func (g *generator) emitStateVarDeclarations(comp *tuigen.Component) {
	for _, node := range comp.Body {
		goCode, ok := node.(*tuigen.GoCode)
		if !ok || goCode == nil {
			continue
		}
		matches := stateNewStateRegex.FindStringSubmatch(goCode.Code)
		if len(matches) < 3 {
			continue
		}
		varName := matches[1]
		initExpr := matches[2]

		// Map the variable name position from .gsx to .go
		tuiLine := goCode.Position.Line - 1
		tuiCol := goCode.Position.Column - 1
		// Find the variable name offset in the original code
		varIdx := strings.Index(goCode.Code, varName)
		if varIdx >= 0 {
			tuiCol += varIdx
		}

		goVarStartCol := 1 + 0 // "\t" + start of varName in "varName := tui.NewState(...)"
		g.sourceMap.AddMapping(Mapping{
			TuiLine: tuiLine,
			TuiCol:  tuiCol,
			GoLine:  g.goLine,
			GoCol:   goVarStartCol,
			Length:  len(varName),
		})
		log.Generate("STATE mapping: %s := tui.NewState(%s) -> TuiLine=%d TuiCol=%d GoLine=%d",
			varName, initExpr, tuiLine, tuiCol, g.goLine)

		// Emit: varName := tui.NewState(initExpr)
		g.writeLine(fmt.Sprintf("\t%s := tui.NewState(%s)", varName, initExpr))
	}
}

// emitNamedRefDeclarations scans a component body for elements with #Name refs
// and emits Go variable declarations so gopls understands ref types.
func (g *generator) emitNamedRefDeclarations(comp *tuigen.Component) {
	g.emitNamedRefFromNodes(comp.Body, false, false)
}

// emitNamedRefFromNodes recursively finds named refs in AST nodes.
func (g *generator) emitNamedRefFromNodes(nodes []tuigen.Node, inLoop, inConditional bool) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.Element:
			if n == nil {
				continue
			}
			if n.NamedRef != "" {
				refType := "var %s *element.Element"
				if inLoop {
					if n.RefKey != nil {
						refType = "var %s map[string]*element.Element"
					} else {
						refType = "var %s []*element.Element"
					}
				}

				tuiLine := n.Position.Line - 1
				// Map to the #Name position, not the element tag position
				// Position.Column-1 is the '<' char; skip <, tag name, and space to reach '#'
				tuiCol := n.Position.Column - 1 + 1 + len(n.Tag) + 1

				goVarStartCol := 1 + len("var ") // "\t" + "var "
				g.sourceMap.AddMapping(Mapping{
					TuiLine: tuiLine,
					TuiCol:  tuiCol,
					GoLine:  g.goLine,
					GoCol:   goVarStartCol,
					Length:  len(n.NamedRef),
				})
				log.Generate("REF mapping: #%s -> TuiLine=%d TuiCol=%d GoLine=%d",
					n.NamedRef, tuiLine, tuiCol, g.goLine)

				g.writeLine(fmt.Sprintf("\t"+refType, n.NamedRef))
			}
			g.emitNamedRefFromNodes(n.Children, inLoop, inConditional)
		case *tuigen.ForLoop:
			if n != nil {
				g.emitNamedRefFromNodes(n.Body, true, inConditional)
			}
		case *tuigen.IfStmt:
			if n != nil {
				g.emitNamedRefFromNodes(n.Then, inLoop, true)
				g.emitNamedRefFromNodes(n.Else, inLoop, true)
			}
		case *tuigen.LetBinding:
			if n != nil && n.Element != nil {
				g.emitNamedRefFromNodes([]tuigen.Node{n.Element}, inLoop, inConditional)
			}
		case *tuigen.ComponentCall:
			if n != nil {
				g.emitNamedRefFromNodes(n.Children, inLoop, inConditional)
			}
		}
	}
}
