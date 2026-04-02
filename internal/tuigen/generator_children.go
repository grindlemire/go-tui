package tuigen

import (
	"fmt"
	"strconv"
	"strings"
)

// generateChildren generates code for element children.
func (g *Generator) generateChildren(parentVar string, children []Node) {
	g.generateChildrenWithRefs(parentVar, children, false, false)
}

// generateChildrenWithRefs generates code for element children with ref context tracking.
func (g *Generator) generateChildrenWithRefs(parentVar string, children []Node, inLoop bool, inConditional bool) {
	for _, child := range children {
		switch c := child.(type) {
		case *Element:
			g.generateElementWithRefs(c, parentVar, inLoop, inConditional)
		case *LetBinding:
			g.generateLetBinding(c, parentVar, inConditional)
		case *ForLoop:
			g.generateForLoopWithRefs(c, parentVar, inLoop, inConditional)
		case *IfStmt:
			g.generateIfStmtWithRefs(c, parentVar, inLoop)
		case *GoExpr:
			// GoExpr as child - create text element with the expression
			varName := g.nextVar()
			g.writef("%s := tui.New(tui.WithText(%s))\n", varName, textExpr(c.Code))
			g.writef("%s.AddChild(%s)\n", parentVar, varName)
		case *TextContent:
			// TextContent as child - create text element
			varName := g.nextVar()
			g.writef("%s := tui.New(tui.WithText(%s))\n", varName, strconv.Quote(c.Text))
			g.writef("%s.AddChild(%s)\n", parentVar, varName)
		case *RawGoExpr:
			// RawGoExpr is a variable reference - add directly
			g.writef("%s.AddChild(%s)\n", parentVar, c.Code)
		case *ComponentCall:
			g.generateComponentCallWithRefs(c, parentVar, inConditional)
		case *ComponentExpr:
			g.generateComponentExpr(c, parentVar)
		case *ChildrenSlot:
			// Expand the children parameter
			// In method templs, children are stored on the receiver struct
			childrenExpr := "children"
			if g.currentReceiver != "" {
				childrenExpr = g.currentReceiver + ".children"
			}
			g.writef("for _, __child := range %s {\n", childrenExpr)
			g.indent++
			g.writef("%s.AddChild(__child)\n", parentVar)
			g.indent--
			g.writeln("}")
		}
	}
}

// generateBodyNodeWithRefs generates code for a node with ref context tracking.
func (g *Generator) generateBodyNodeWithRefs(node Node, parentVar string, inLoop bool, inConditional bool) {
	switch n := node.(type) {
	case *Element:
		g.generateElementWithRefs(n, parentVar, inLoop, inConditional)
	case *LetBinding:
		g.generateLetBinding(n, parentVar, inConditional)
	case *ForLoop:
		g.generateForLoopWithRefs(n, parentVar, inLoop, inConditional)
	case *IfStmt:
		g.generateIfStmtWithRefs(n, parentVar, inLoop)
	case *GoCode:
		g.generateGoCode(n)
	case *GoExpr:
		if parentVar != "" {
			varName := g.nextVar()
			g.writef("%s := tui.New(tui.WithText(%s))\n", varName, textExpr(n.Code))
			g.writef("%s.AddChild(%s)\n", parentVar, varName)
		} else {
			g.writef("%s\n", n.Code)
		}
	case *TextContent:
		if parentVar != "" {
			varName := g.nextVar()
			g.writef("%s := tui.New(tui.WithText(%s))\n", varName, strconv.Quote(n.Text))
			g.writef("%s.AddChild(%s)\n", parentVar, varName)
		}
	case *RawGoExpr:
		if parentVar != "" {
			g.writef("%s.AddChild(%s)\n", parentVar, n.Code)
		}
	case *ComponentCall:
		g.generateComponentCallWithRefs(n, parentVar, inConditional)
	case *ComponentExpr:
		g.generateComponentExpr(n, parentVar)
	case *ChildrenSlot:
		if parentVar != "" {
			// In method templs, children are stored on the receiver struct
			childrenExpr := "children"
			if g.currentReceiver != "" {
				childrenExpr = g.currentReceiver + ".children"
			}
			g.writef("for _, __child := range %s {\n", childrenExpr)
			g.indent++
			g.writef("%s.AddChild(__child)\n", parentVar)
			g.indent--
			g.writeln("}")
		}
	}
}

// generateGoCode generates a raw Go statement.
func (g *Generator) generateGoCode(gc *GoCode) {
	g.writef("%s\n", gc.Code)
}

// generateGoFunc generates a top-level Go function.
func (g *Generator) generateGoFunc(fn *GoFunc) {
	g.generatePassthroughCode(fn.Code, fn.Position)
	g.writeln("")
}

// generateGoDecl generates a top-level Go declaration (type, const, var).
func (g *Generator) generateGoDecl(decl *GoDecl) {
	g.generatePassthroughCode(decl.Code, decl.Position)
	g.writeln("")
}

// generatePassthroughCode generates code that is passed through from .gsx to .go
// and records source map mappings for each line.
func (g *Generator) generatePassthroughCode(code string, pos Position) {
	lines := strings.Split(code, "\n")
	// Convert to 0-indexed. We count newlines in the Code to determine the
	// actual starting line, since Position.Line may point to inside the declaration.
	codeLines := strings.Count(code, "\n") + 1
	// gsxLine is where this code block ENDS, so start = end - count + 1
	// But we use Position.Line as the start, just convert to 0-indexed
	gsxLine := pos.Line - 1 // Convert to 0-indexed

	for i, line := range lines {
		// Record mapping for this line
		if g.sourceMap != nil {
			g.sourceMap.AddMapping(SourceMapping{
				GoLine:  g.currentLine,
				GoCol:   0,
				GsxLine: gsxLine + i,
				GsxCol:  0,
				Length:  len(line),
			})
		}
		g.writeln(line)
	}
	_ = codeLines // unused for now
}

// generateComponentCallWithRefs generates code for a component call.
// Returns the variable name holding the result.
//
// For struct component mounts (IsStructMount=true), generates:
//
//	app.Mount(receiverVar, index, func() tui.Component { return Name(args) })
//
// For function component calls (IsStructMount=false), generates the existing
// view struct pattern: varName := Name(args)
//
// Function templs defined in the same file are always called directly (not mounted),
// even when inside a method templ. They are stateless views — calling them fresh
// each render ensures updated props. Cross-package function templs that are unknown
// to the generator still get mounted, but their view types have UpdateProps so the
// mount system can refresh them.
func (g *Generator) generateComponentCallWithRefs(call *ComponentCall, parentVar string, inConditional bool) string {
	if call.IsStructMount && !g.functionTempls[call.Name] {
		return g.generateStructMount(call, parentVar)
	}
	return g.generateFunctionComponentCall(call, parentVar, inConditional)
}

// returnsElement reports whether a ComponentCall produces a *tui.Element directly
// (true for actual struct mounts) vs a view struct that needs .Root (false for
// function templs, even when IsStructMount is set by parser context).
func (g *Generator) returnsElement(call *ComponentCall) bool {
	return call.IsStructMount && !g.functionTempls[call.Name]
}

// generateStructMount generates an app.Mount() call for struct components.
// Returns the variable name holding the *tui.Element result.
func (g *Generator) generateStructMount(call *ComponentCall, parentVar string) string {
	varName := g.nextVar()
	baseIndex := g.mountIndex
	g.mountIndex++

	// Determine the mount index expression.
	// If we're inside a loop, combine the static base index with runtime loop indices
	// to ensure each iteration gets a unique mount key.
	indexExpr := g.loopIndexExpr(baseIndex)
	if indexExpr == "" {
		// Not in a loop - use static index
		indexExpr = fmt.Sprintf("%d", baseIndex)
	}

	// Build children slice if the component call has children
	childrenVar := ""
	if len(call.Children) > 0 {
		childrenVar = varName + "_children"
		g.writef("%s := []*tui.Element{}\n", childrenVar)

		for _, child := range call.Children {
			switch c := child.(type) {
			case *Element:
				elemVar := g.generateElement(c, "")
				g.writef("%s = append(%s, %s)\n", childrenVar, childrenVar, elemVar)
			case *ComponentCall:
				innerVar := g.generateComponentCallWithRefs(c, "", false)
				if g.returnsElement(c) {
					g.writef("%s = append(%s, %s)\n", childrenVar, childrenVar, innerVar)
				} else {
					g.writef("%s = append(%s, %s.Root)\n", childrenVar, childrenVar, innerVar)
				}
			case *LetBinding:
				g.generateLetBinding(c, "", false)
				g.writef("%s = append(%s, %s)\n", childrenVar, childrenVar, c.Name)
			case *ForLoop:
				g.generateForLoopForSlice(c, childrenVar)
			case *IfStmt:
				g.generateIfStmtForSlice(c, childrenVar)
			case *GoExpr:
				elemVar := g.nextVar()
				g.writef("%s := tui.New(tui.WithText(%s))\n", elemVar, textExpr(c.Code))
				g.writef("%s = append(%s, %s)\n", childrenVar, childrenVar, elemVar)
			case *TextContent:
				elemVar := g.nextVar()
				g.writef("%s := tui.New(tui.WithText(%s))\n", elemVar, strconv.Quote(c.Text))
				g.writef("%s = append(%s, %s)\n", childrenVar, childrenVar, elemVar)
			case *RawGoExpr:
				g.writef("%s = append(%s, %s)\n", childrenVar, childrenVar, c.Code)
			}
		}
	}

	// Generate: varName := app.Mount(receiverVar, indexExpr, func() tui.Component { return Name(args) })
	g.writef("%s := app.Mount(%s, %s, func() tui.Component {\n", varName, g.currentReceiver, indexExpr)
	g.indent++

	// Build the argument list, appending children if present
	args := call.Args
	if childrenVar != "" {
		if args == "" {
			args = childrenVar
		} else {
			args = args + ", " + childrenVar
		}
	}

	if args == "" {
		g.writef("return %s()\n", call.Name)
	} else {
		g.writef("return %s(%s)\n", call.Name, args)
	}
	g.indent--
	g.writeln("})")

	// Add to parent if specified — Mount returns *tui.Element directly
	if parentVar != "" {
		g.writef("%s.AddChild(%s)\n", parentVar, varName)
	}

	return varName
}

// generateFunctionComponentCall generates a function component call (existing behavior).
// Returns the variable name holding the view struct result.
func (g *Generator) generateFunctionComponentCall(call *ComponentCall, parentVar string, inConditional bool) string {
	varName := g.nextVar()

	// When inside a block scope in a function templ, use assignment (=) instead of
	// short declaration (:=) because the variable will be hoisted to function scope.
	// Method templs don't need hoisting since they don't have watcher/bind/unbind
	// wiring that references component vars outside the block.
	decl := ":="
	if inConditional && g.currentReceiver == "" {
		decl = "="
	}

	if len(call.Children) == 0 {
		// No children - simple call, returns view struct
		if call.Args == "" {
			g.writef("%s %s %s()\n", varName, decl, call.Name)
		} else {
			g.writef("%s %s %s(%s)\n", varName, decl, call.Name, call.Args)
		}
	} else {
		// Has children - build children slice first
		childrenVar := g.nextVar() + "_children"
		g.writef("%s := []*tui.Element{}\n", childrenVar)

		// Generate each child and append to slice
		for _, child := range call.Children {
			switch c := child.(type) {
			case *Element:
				elemVar := g.generateElement(c, "")
				g.writef("%s = append(%s, %s)\n", childrenVar, childrenVar, elemVar)
			case *ComponentCall:
				innerVar := g.generateComponentCallWithRefs(c, "", false)
				if g.returnsElement(c) {
					// Struct mount returns *tui.Element directly
					g.writef("%s = append(%s, %s)\n", childrenVar, childrenVar, innerVar)
				} else {
					// Function component returns view struct — extract .Root
					g.writef("%s = append(%s, %s.Root)\n", childrenVar, childrenVar, innerVar)
				}
			case *LetBinding:
				g.generateLetBinding(c, "", false)
				g.writef("%s = append(%s, %s)\n", childrenVar, childrenVar, c.Name)
			case *ForLoop:
				// For loops generate multiple elements - use a temp slice
				g.generateForLoopForSlice(c, childrenVar)
			case *IfStmt:
				// If statements may or may not generate elements
				g.generateIfStmtForSlice(c, childrenVar)
			case *GoExpr:
				// Expression - wrap in text element
				elemVar := g.nextVar()
				g.writef("%s := tui.New(tui.WithText(%s))\n", elemVar, textExpr(c.Code))
				g.writef("%s = append(%s, %s)\n", childrenVar, childrenVar, elemVar)
			case *TextContent:
				elemVar := g.nextVar()
				g.writef("%s := tui.New(tui.WithText(%s))\n", elemVar, strconv.Quote(c.Text))
				g.writef("%s = append(%s, %s)\n", childrenVar, childrenVar, elemVar)
			case *RawGoExpr:
				g.writef("%s = append(%s, %s)\n", childrenVar, childrenVar, c.Code)
			}
		}

		// Call component with children
		if call.Args == "" {
			g.writef("%s %s %s(%s)\n", varName, decl, call.Name, childrenVar)
		} else {
			g.writef("%s %s %s(%s, %s)\n", varName, decl, call.Name, call.Args, childrenVar)
		}
	}

	// Track this component call for watcher aggregation
	g.componentVars = append(g.componentVars, componentVarEntry{
		name:          varName,
		componentName: call.Name,
		inConditional: inConditional,
	})

	// Add to parent if specified - use .Root to get the element from the view struct
	if parentVar != "" {
		g.writef("%s.AddChild(%s.Root)\n", parentVar, varName)
	}

	return varName
}

// generateComponentExpr generates code for a component expression like @c.textarea.
// It calls .Render() on the expression and adds the result to the parent.
func (g *Generator) generateComponentExpr(expr *ComponentExpr, parentVar string) {
	varName := g.nextVar()
	g.writef("%s := %s.Render(app)\n", varName, expr.Expr)
	if parentVar != "" {
		g.writef("%s.AddChild(%s)\n", parentVar, varName)
	}

	// Track receiver field accesses for BindApp generation in method components.
	// e.g., @c.settingsView → track "settingsView" so generateBindApp can bind it.
	g.trackComponentExprField(expr.Expr)
}

// generateForLoopForSlice generates a for loop that appends elements to a slice.
func (g *Generator) generateForLoopForSlice(loop *ForLoop, sliceVar string) {
	// Push the loop index variable for use in struct mount calls
	idxVar := g.pushLoopIndex(loop)
	defer g.popLoopIndex()

	// Build loop header - may need to generate a synthetic index variable
	var loopVars string
	if loop.Index != "" && loop.Index != "_" {
		loopVars = fmt.Sprintf("%s, %s", loop.Index, loop.Value)
	} else if loop.Index == "_" {
		loopVars = fmt.Sprintf("%s, %s", idxVar, loop.Value)
	} else {
		loopVars = fmt.Sprintf("%s, %s", idxVar, loop.Value)
	}

	g.writef("for %s := range %s {\n", loopVars, loop.Iterable)
	g.indent++

	// Silence unused variable warnings
	if loop.Index != "" && loop.Index != "_" {
		g.writef("_ = %s\n", loop.Index)
	} else {
		g.writef("_ = %s\n", idxVar)
	}

	for _, node := range loop.Body {
		switch n := node.(type) {
		case *Element:
			elemVar := g.generateElement(n, "")
			g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, elemVar)
		case *ComponentCall:
			callVar := g.generateComponentCallWithRefs(n, "", true)
			if g.returnsElement(n) {
				g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, callVar)
			} else {
				g.writef("%s = append(%s, %s.Root)\n", sliceVar, sliceVar, callVar)
			}
		case *ComponentExpr:
			elemVar := g.nextVar()
			g.writef("%s := %s.Render(app)\n", elemVar, n.Expr)
			g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, elemVar)
		case *LetBinding:
			g.generateLetBinding(n, "", true)
			g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, n.Name)
		case *ForLoop:
			g.generateForLoopForSlice(n, sliceVar)
		case *IfStmt:
			g.generateIfStmtForSlice(n, sliceVar)
		case *GoCode:
			g.generateGoCode(n)
		case *GoExpr:
			elemVar := g.nextVar()
			g.writef("%s := tui.New(tui.WithText(%s))\n", elemVar, textExpr(n.Code))
			g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, elemVar)
		}
	}

	g.indent--
	g.writeln("}")
}

// generateIfStmtForSlice generates an if statement that appends elements to a slice.
func (g *Generator) generateIfStmtForSlice(stmt *IfStmt, sliceVar string) {
	g.writef("if %s {\n", stmt.Condition)
	g.indent++

	for _, node := range stmt.Then {
		switch n := node.(type) {
		case *Element:
			elemVar := g.generateElement(n, "")
			g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, elemVar)
		case *ComponentCall:
			callVar := g.generateComponentCallWithRefs(n, "", true)
			if g.returnsElement(n) {
				g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, callVar)
			} else {
				g.writef("%s = append(%s, %s.Root)\n", sliceVar, sliceVar, callVar)
			}
		case *ComponentExpr:
			elemVar := g.nextVar()
			g.writef("%s := %s.Render(app)\n", elemVar, n.Expr)
			g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, elemVar)
		case *LetBinding:
			g.generateLetBinding(n, "", true)
			g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, n.Name)
		case *ForLoop:
			g.generateForLoopForSlice(n, sliceVar)
		case *IfStmt:
			g.generateIfStmtForSlice(n, sliceVar)
		case *GoCode:
			g.generateGoCode(n)
		case *GoExpr:
			elemVar := g.nextVar()
			g.writef("%s := tui.New(tui.WithText(%s))\n", elemVar, textExpr(n.Code))
			g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, elemVar)
		}
	}

	g.indent--

	if len(stmt.Else) > 0 {
		g.write("} else ")

		if len(stmt.Else) == 1 {
			if elseIf, ok := stmt.Else[0].(*IfStmt); ok {
				g.generateIfStmtForSlice(elseIf, sliceVar)
				return
			}
		}

		g.writeln("{")
		g.indent++
		for _, node := range stmt.Else {
			switch n := node.(type) {
			case *Element:
				elemVar := g.generateElement(n, "")
				g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, elemVar)
			case *ComponentCall:
				callVar := g.generateComponentCallWithRefs(n, "", true)
				if g.returnsElement(n) {
					g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, callVar)
				} else {
					g.writef("%s = append(%s, %s.Root)\n", sliceVar, sliceVar, callVar)
				}
			case *ComponentExpr:
				elemVar := g.nextVar()
				g.writef("%s := %s.Render(app)\n", elemVar, n.Expr)
				g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, elemVar)
			case *LetBinding:
				g.generateLetBinding(n, "", true)
				g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, n.Name)
			case *ForLoop:
				g.generateForLoopForSlice(n, sliceVar)
			case *IfStmt:
				g.generateIfStmtForSlice(n, sliceVar)
			case *GoCode:
				g.generateGoCode(n)
			case *GoExpr:
				elemVar := g.nextVar()
				g.writef("%s := tui.New(tui.WithText(%s))\n", elemVar, textExpr(n.Code))
				g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, elemVar)
			}
		}
		g.indent--
		g.writeln("}")
	} else {
		g.writeln("}")
	}
}
