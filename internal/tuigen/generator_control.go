package tuigen

import (
	"fmt"
)

// generateLetBinding generates code for a @let binding.
func (g *Generator) generateLetBinding(let *LetBinding, parentVar string) {
	// Generate the element with a specific variable name
	elemOpts := g.buildElementOptions(let.Element)

	if len(elemOpts.options) == 0 {
		g.writef("%s := tui.New()\n", let.Name)
	} else {
		g.writef("%s := tui.New(\n", let.Name)
		g.indent++
		for _, opt := range elemOpts.options {
			g.writef("%s,\n", opt)
		}
		g.indent--
		g.writeln(")")
	}

	// Defer watcher attachment until after all elements are created
	for _, watcher := range elemOpts.watchers {
		g.deferredWatchers = append(g.deferredWatchers, deferredWatcher{
			elementVar:  let.Name,
			watcherExpr: watcher,
		})
	}

	// Generate children for the let-bound element - skip if text element already has content in WithText
	if !skipTextChildren(let.Element) {
		g.generateChildren(let.Name, let.Element.Children)
	}

	// Add to parent if specified
	if parentVar != "" {
		g.writef("%s.AddChild(%s)\n", parentVar, let.Name)
	}
}

// generateForLoop generates code for a @for loop.
func (g *Generator) generateForLoop(loop *ForLoop, parentVar string) {
	g.generateForLoopWithRefs(loop, parentVar, false)
}

// generateForLoopWithRefs generates code for a @for loop with ref context tracking.
func (g *Generator) generateForLoopWithRefs(loop *ForLoop, parentVar string, inConditional bool) {
	// Build loop header
	var loopVars string
	if loop.Index != "" {
		loopVars = fmt.Sprintf("%s, %s", loop.Index, loop.Value)
	} else {
		loopVars = loop.Value
	}

	g.writef("for %s := range %s {\n", loopVars, loop.Iterable)
	g.indent++

	// Silence unused variable warnings if index is not used
	if loop.Index != "" && loop.Index != "_" {
		g.writef("_ = %s\n", loop.Index)
	}

	// Generate loop body - now inside a loop context
	for _, node := range loop.Body {
		switch n := node.(type) {
		case *Element:
			g.generateElementWithRefs(n, parentVar, true, inConditional)
		case *LetBinding:
			g.generateLetBinding(n, parentVar)
		case *ForLoop:
			g.generateForLoopWithRefs(n, parentVar, inConditional)
		case *IfStmt:
			g.generateIfStmtWithRefs(n, parentVar, true) // now in loop context
		case *GoCode:
			g.generateGoCode(n)
		case *GoExpr:
			// Bare expression in loop body
			if parentVar != "" {
				varName := g.nextVar()
				g.writef("%s := tui.New(tui.WithText(%s))\n", varName, n.Code)
				g.writef("%s.AddChild(%s)\n", parentVar, varName)
			} else {
				g.writef("%s\n", n.Code)
			}
		case *ComponentCall:
			g.generateComponentCallWithRefs(n, parentVar)
		case *ChildrenSlot:
			if parentVar != "" {
				g.writeln("for _, __child := range children {")
				g.indent++
				g.writef("%s.AddChild(__child)\n", parentVar)
				g.indent--
				g.writeln("}")
			}
		}
	}

	g.indent--
	g.writeln("}")
}

// generateIfStmt generates code for an @if statement.
func (g *Generator) generateIfStmt(stmt *IfStmt, parentVar string) {
	g.generateIfStmtWithRefs(stmt, parentVar, false)
}

// generateIfStmtWithRefs generates code for an @if statement with ref context tracking.
func (g *Generator) generateIfStmtWithRefs(stmt *IfStmt, parentVar string, inLoop bool) {
	g.writef("if %s {\n", stmt.Condition)
	g.indent++

	// Generate then body - now inside conditional context
	for _, node := range stmt.Then {
		g.generateBodyNodeWithRefs(node, parentVar, inLoop, true)
	}

	g.indent--

	// Generate else branch if present
	if len(stmt.Else) > 0 {
		g.write("} else ")

		// Check if else contains a single IfStmt (else-if chain)
		if len(stmt.Else) == 1 {
			if elseIf, ok := stmt.Else[0].(*IfStmt); ok {
				g.generateIfStmtWithRefs(elseIf, parentVar, inLoop)
				return
			}
		}

		g.writeln("{")
		g.indent++
		for _, node := range stmt.Else {
			g.generateBodyNodeWithRefs(node, parentVar, inLoop, true)
		}
		g.indent--
		g.writeln("}")
	} else {
		g.writeln("}")
	}
}
