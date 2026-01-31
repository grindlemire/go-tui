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

// emitRefDeclarations scans a component body for elements with ref={} attributes
// and emits Go variable declarations so gopls understands ref types.
func (g *generator) emitRefDeclarations(comp *tuigen.Component) {
	g.emitRefFromNodes(comp.Body, false, false)
}

// emitRefFromNodes recursively finds ref attributes in AST nodes.
func (g *generator) emitRefFromNodes(nodes []tuigen.Node, inLoop, inConditional bool) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.Element:
			if n == nil {
				continue
			}
			if n.RefExpr != nil {
				refName := n.RefExpr.Code
				refType := "var %s *element.Element"
				if inLoop {
					if n.RefKey != nil {
						refType = "var %s map[string]*element.Element"
					} else {
						refType = "var %s []*element.Element"
					}
				}

				tuiLine := n.RefExpr.Position.Line - 1
				tuiCol := n.RefExpr.Position.Column - 1

				goVarStartCol := 1 + len("var ") // "\t" + "var "
				g.sourceMap.AddMapping(Mapping{
					TuiLine: tuiLine,
					TuiCol:  tuiCol,
					GoLine:  g.goLine,
					GoCol:   goVarStartCol,
					Length:  len(refName),
				})
				log.Generate("REF mapping: ref={%s} -> TuiLine=%d TuiCol=%d GoLine=%d",
					refName, tuiLine, tuiCol, g.goLine)

				g.writeLine(fmt.Sprintf("\t"+refType, refName))
			}
			g.emitRefFromNodes(n.Children, inLoop, inConditional)
		case *tuigen.ForLoop:
			if n != nil {
				g.emitRefFromNodes(n.Body, true, inConditional)
			}
		case *tuigen.IfStmt:
			if n != nil {
				g.emitRefFromNodes(n.Then, inLoop, true)
				g.emitRefFromNodes(n.Else, inLoop, true)
			}
		case *tuigen.LetBinding:
			if n != nil && n.Element != nil {
				g.emitRefFromNodes([]tuigen.Node{n.Element}, inLoop, inConditional)
			}
		case *tuigen.ComponentCall:
			if n != nil {
				g.emitRefFromNodes(n.Children, inLoop, inConditional)
			}
		}
	}
}
