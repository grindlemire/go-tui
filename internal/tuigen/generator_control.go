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
	g.generateForLoopWithRefs(loop, parentVar, false, false)
}

// generateForLoopWithRefs generates code for a @for loop with ref context tracking.
// When the loop body references state variables (and we're not already in a loop/reactive context),
// generates a reactive wrapper that rebuilds loop children when state changes.
func (g *Generator) generateForLoopWithRefs(loop *ForLoop, parentVar string, inLoop bool, inConditional bool) {
	// Check if this loop body references state and we're not in a loop context
	if !inLoop && parentVar != "" {
		deps := collectForLoopDeps(loop, g.stateNameSet())
		if len(deps) > 0 {
			g.generateReactiveForLoop(loop, parentVar, deps)
			return
		}
	}

	// Push the loop index variable for use in struct mount calls.
	// This ensures each loop iteration gets a unique mount key.
	idxVar := g.pushLoopIndex(loop)
	defer g.popLoopIndex()

	// Build loop header - may need to generate a synthetic index variable
	var loopVars string
	if loop.Index != "" && loop.Index != "_" {
		// User provided a usable index variable
		loopVars = fmt.Sprintf("%s, %s", loop.Index, loop.Value)
	} else if loop.Index == "_" {
		// User explicitly ignored index, but we need one for mount keys
		loopVars = fmt.Sprintf("%s, %s", idxVar, loop.Value)
	} else {
		// No index in original, need to add one for mount keys
		loopVars = fmt.Sprintf("%s, %s", idxVar, loop.Value)
	}

	g.writef("for %s := range %s {\n", loopVars, loop.Iterable)
	g.indent++

	// Silence unused variable warnings if index is not used elsewhere
	// (it will be used in mount calls, but Go doesn't know that at compile time
	// if there are no struct component calls in the loop body)
	if loop.Index != "" && loop.Index != "_" {
		g.writef("_ = %s\n", loop.Index)
	} else {
		g.writef("_ = %s\n", idxVar)
	}

	// Generate loop body - now inside a loop context
	for _, node := range loop.Body {
		switch n := node.(type) {
		case *Element:
			g.generateElementWithRefs(n, parentVar, true, inConditional)
		case *LetBinding:
			g.generateLetBinding(n, parentVar)
		case *ForLoop:
			g.generateForLoopWithRefs(n, parentVar, true, inConditional) // nested loop inside loop context
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
// When the condition references state variables (and we're not in a loop), generates
// a reactive wrapper that rebuilds its children when state changes.
func (g *Generator) generateIfStmtWithRefs(stmt *IfStmt, parentVar string, inLoop bool) {
	// Check if this condition references state and we're not in a loop context
	if !inLoop && parentVar != "" {
		deps := collectAllIfStmtDeps(stmt, g.stateNameSet())
		if len(deps) > 0 {
			g.generateReactiveIfStmt(stmt, parentVar, deps)
			return
		}
	}

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

// generateReactiveIfStmt generates a reactive @if block that rebuilds when state changes.
// It creates a wrapper element, an update closure, and state bindings.
func (g *Generator) generateReactiveIfStmt(stmt *IfStmt, parentVar string, deps []string) {
	condVar := g.nextCondVar()
	updateFn := fmt.Sprintf("__update_%s", condVar)

	// Create wrapper element and add to parent
	g.writef("%s := tui.New()\n", condVar)
	g.writef("%s.AddChild(%s)\n", parentVar, condVar)

	// Generate update function
	g.writef("%s := func() {\n", updateFn)
	g.indent++
	g.writef("%s.RemoveAllChildren()\n", condVar)

	// Generate the if/else structure inside the closure with wrapper as parent.
	// Use inLoop=true to prevent nested reactive handling and skip text bindings.
	g.generateIfStmtWithRefs(stmt, condVar, true)

	g.indent--
	g.writeln("}")

	// Call initially
	g.writef("%s()\n", updateFn)

	// Bind to all referenced state variables
	stateTypes := make(map[string]string)
	for _, sv := range g.stateVars {
		stateTypes[sv.Name] = sv.Type
	}
	for _, dep := range deps {
		stateType := stateTypes[dep]
		g.writef("%s.Bind(func(_ %s) { %s() })\n", dep, stateType, updateFn)
	}
}

// collectAllIfStmtDeps collects all state variable dependencies from an @if statement,
// including the condition, else-if conditions, and all expressions in the body.
// This ensures the reactive update fires for any state change that could affect the block.
func collectAllIfStmtDeps(stmt *IfStmt, stateNames map[string]bool) []string {
	seen := make(map[string]bool)
	var deps []string

	addDeps := func(expr string) {
		for _, d := range detectGetCallsInExpr(expr, stateNames) {
			if !seen[d] {
				seen[d] = true
				deps = append(deps, d)
			}
		}
	}

	var collectFromNodes func(nodes []Node)
	collectFromNodes = func(nodes []Node) {
		for _, node := range nodes {
			switch n := node.(type) {
			case *Element:
				for _, attr := range n.Attributes {
					if goExpr, ok := attr.Value.(*GoExpr); ok {
						addDeps(goExpr.Code)
					}
				}
				collectFromNodes(n.Children)
			case *GoExpr:
				addDeps(n.Code)
			case *IfStmt:
				addDeps(n.Condition)
				collectFromNodes(n.Then)
				collectFromNodes(n.Else)
			case *ForLoop:
				addDeps(n.Iterable)
				collectFromNodes(n.Body)
			case *ComponentCall:
				addDeps(n.Args)
				collectFromNodes(n.Children)
			}
		}
	}

	// Collect from the top-level condition
	addDeps(stmt.Condition)

	// Collect from then/else bodies
	collectFromNodes(stmt.Then)
	collectFromNodes(stmt.Else)

	return deps
}

// generateReactiveForLoop generates a reactive @for loop that rebuilds when state changes.
// It creates a wrapper element, an update closure, and state bindings.
func (g *Generator) generateReactiveForLoop(loop *ForLoop, parentVar string, deps []string) {
	loopVar := g.nextLoopVar()
	updateFn := fmt.Sprintf("__update_%s", loopVar)

	// Create wrapper element that inherits parent's container layout, and add to parent.
	// This ensures loop children are laid out the same way they would be as direct children.
	parentStyle := fmt.Sprintf("%s_style", loopVar)
	g.writef("%s := %s.LayoutStyle()\n", parentStyle, parentVar)
	g.writef("%s := tui.New(tui.WithDirection(%s.Direction), tui.WithGap(%s.Gap))\n", loopVar, parentStyle, parentStyle)
	g.writef("%s.AddChild(%s)\n", parentVar, loopVar)

	// Generate update function
	g.writef("%s := func() {\n", updateFn)
	g.indent++
	g.writef("%s.RemoveAllChildren()\n", loopVar)

	// Generate the for loop inside the closure with wrapper as parent.
	// Use inLoop=true to prevent nested reactive handling.
	g.generateForLoopWithRefs(loop, loopVar, true, false)

	g.indent--
	g.writeln("}")

	// Call initially
	g.writef("%s()\n", updateFn)

	// Bind to all referenced state variables
	stateTypes := make(map[string]string)
	for _, sv := range g.stateVars {
		stateTypes[sv.Name] = sv.Type
	}
	for _, dep := range deps {
		stateType := stateTypes[dep]
		g.writef("%s.Bind(func(_ %s) { %s() })\n", dep, stateType, updateFn)
	}
}

// collectForLoopDeps collects all state variable dependencies from a @for loop body.
// This scans the iterable expression and all nested nodes for state .Get() calls.
func collectForLoopDeps(loop *ForLoop, stateNames map[string]bool) []string {
	seen := make(map[string]bool)
	var deps []string

	addDeps := func(expr string) {
		for _, d := range detectGetCallsInExpr(expr, stateNames) {
			if !seen[d] {
				seen[d] = true
				deps = append(deps, d)
			}
		}
	}

	// Check the iterable expression
	addDeps(loop.Iterable)

	// Collect from body nodes
	var collectFromNodes func(nodes []Node)
	collectFromNodes = func(nodes []Node) {
		for _, node := range nodes {
			switch n := node.(type) {
			case *Element:
				for _, attr := range n.Attributes {
					if goExpr, ok := attr.Value.(*GoExpr); ok {
						addDeps(goExpr.Code)
					}
				}
				collectFromNodes(n.Children)
			case *GoExpr:
				addDeps(n.Code)
			case *IfStmt:
				addDeps(n.Condition)
				collectFromNodes(n.Then)
				collectFromNodes(n.Else)
			case *ForLoop:
				addDeps(n.Iterable)
				collectFromNodes(n.Body)
			case *ComponentCall:
				addDeps(n.Args)
				collectFromNodes(n.Children)
			}
		}
	}

	collectFromNodes(loop.Body)
	return deps
}
