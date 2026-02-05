package tuigen

import (
	"fmt"
	"strconv"
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
			g.generateLetBinding(c, parentVar)
		case *ForLoop:
			g.generateForLoopWithRefs(c, parentVar, inLoop, inConditional)
		case *IfStmt:
			g.generateIfStmtWithRefs(c, parentVar, inLoop)
		case *GoExpr:
			// GoExpr as child - create text element with the expression
			varName := g.nextVar()
			g.writef("%s := tui.New(tui.WithText(%s))\n", varName, c.Code)
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
			g.generateComponentCallWithRefs(c, parentVar)
		case *ChildrenSlot:
			// Expand the children parameter
			g.writeln("for _, __child := range children {")
			g.indent++
			g.writef("%s.AddChild(__child)\n", parentVar)
			g.indent--
			g.writeln("}")
		}
	}
}

// generateBodyNode generates code for a node in a component/control flow body.
func (g *Generator) generateBodyNode(node Node, parentVar string) {
	g.generateBodyNodeWithRefs(node, parentVar, false, false)
}

// generateBodyNodeWithRefs generates code for a node with ref context tracking.
func (g *Generator) generateBodyNodeWithRefs(node Node, parentVar string, inLoop bool, inConditional bool) {
	switch n := node.(type) {
	case *Element:
		g.generateElementWithRefs(n, parentVar, inLoop, inConditional)
	case *LetBinding:
		g.generateLetBinding(n, parentVar)
	case *ForLoop:
		g.generateForLoopWithRefs(n, parentVar, inLoop, inConditional)
	case *IfStmt:
		g.generateIfStmtWithRefs(n, parentVar, inLoop)
	case *GoCode:
		g.generateGoCode(n)
	case *GoExpr:
		if parentVar != "" {
			varName := g.nextVar()
			g.writef("%s := tui.New(tui.WithText(%s))\n", varName, n.Code)
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

// generateGoCode generates a raw Go statement.
func (g *Generator) generateGoCode(gc *GoCode) {
	g.writef("%s\n", gc.Code)
}

// generateGoFunc generates a top-level Go function.
func (g *Generator) generateGoFunc(fn *GoFunc) {
	g.writef("%s\n\n", fn.Code)
}

// generateGoDecl generates a top-level Go declaration (type, const, var).
func (g *Generator) generateGoDecl(decl *GoDecl) {
	g.writef("%s\n\n", decl.Code)
}

// generateComponentCall generates code for a component call.
// Returns the variable name holding the result.
func (g *Generator) generateComponentCall(call *ComponentCall, parentVar string) string {
	return g.generateComponentCallWithRefs(call, parentVar)
}

// generateComponentCallWithRefs generates code for a component call.
// Returns the variable name holding the result.
//
// For struct component mounts (IsStructMount=true), generates:
//
//	tui.Mount(receiverVar, index, func() tui.Component { return Name(args) })
//
// For function component calls (IsStructMount=false), generates the existing
// view struct pattern: varName := Name(args)
func (g *Generator) generateComponentCallWithRefs(call *ComponentCall, parentVar string) string {
	if call.IsStructMount {
		return g.generateStructMount(call, parentVar)
	}
	return g.generateFunctionComponentCall(call, parentVar)
}

// generateStructMount generates a tui.Mount() call for struct components.
// Returns the variable name holding the *tui.Element result.
func (g *Generator) generateStructMount(call *ComponentCall, parentVar string) string {
	varName := g.nextVar()
	index := g.mountIndex
	g.mountIndex++

	// Generate: varName := tui.Mount(receiverVar, index, func() tui.Component { return Name(args) })
	g.writef("%s := tui.Mount(%s, %d, func() tui.Component {\n", varName, g.currentReceiver, index)
	g.indent++
	if call.Args == "" {
		g.writef("return %s()\n", call.Name)
	} else {
		g.writef("return %s(%s)\n", call.Name, call.Args)
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
func (g *Generator) generateFunctionComponentCall(call *ComponentCall, parentVar string) string {
	varName := g.nextVar()

	if len(call.Children) == 0 {
		// No children - simple call, returns view struct
		if call.Args == "" {
			g.writef("%s := %s()\n", varName, call.Name)
		} else {
			g.writef("%s := %s(%s)\n", varName, call.Name, call.Args)
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
				innerVar := g.generateComponentCallWithRefs(c, "")
				if c.IsStructMount {
					// Struct mount returns *tui.Element directly
					g.writef("%s = append(%s, %s)\n", childrenVar, childrenVar, innerVar)
				} else {
					// Function component returns view struct — extract .Root
					g.writef("%s = append(%s, %s.Root)\n", childrenVar, childrenVar, innerVar)
				}
			case *LetBinding:
				g.generateLetBinding(c, "")
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
				g.writef("%s := tui.New(tui.WithText(%s))\n", elemVar, c.Code)
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
			g.writef("%s := %s(%s)\n", varName, call.Name, childrenVar)
		} else {
			g.writef("%s := %s(%s, %s)\n", varName, call.Name, call.Args, childrenVar)
		}
	}

	// Track this component call for watcher aggregation
	g.componentVars = append(g.componentVars, varName)

	// Add to parent if specified - use .Root to get the element from the view struct
	if parentVar != "" {
		g.writef("%s.AddChild(%s.Root)\n", parentVar, varName)
	}

	return varName
}

// generateForLoopForSlice generates a for loop that appends elements to a slice.
func (g *Generator) generateForLoopForSlice(loop *ForLoop, sliceVar string) {
	var loopVars string
	if loop.Index != "" {
		loopVars = fmt.Sprintf("%s, %s", loop.Index, loop.Value)
	} else {
		loopVars = loop.Value
	}

	g.writef("for %s := range %s {\n", loopVars, loop.Iterable)
	g.indent++

	if loop.Index != "" && loop.Index != "_" {
		g.writef("_ = %s\n", loop.Index)
	}

	for _, node := range loop.Body {
		switch n := node.(type) {
		case *Element:
			elemVar := g.generateElement(n, "")
			g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, elemVar)
		case *ComponentCall:
			callVar := g.generateComponentCallWithRefs(n, "")
			if n.IsStructMount {
				g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, callVar)
			} else {
				g.writef("%s = append(%s, %s.Root)\n", sliceVar, sliceVar, callVar)
			}
		case *LetBinding:
			g.generateLetBinding(n, "")
			g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, n.Name)
		case *ForLoop:
			g.generateForLoopForSlice(n, sliceVar)
		case *IfStmt:
			g.generateIfStmtForSlice(n, sliceVar)
		case *GoCode:
			g.generateGoCode(n)
		case *GoExpr:
			elemVar := g.nextVar()
			g.writef("%s := tui.New(tui.WithText(%s))\n", elemVar, n.Code)
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
			callVar := g.generateComponentCallWithRefs(n, "")
			if n.IsStructMount {
				g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, callVar)
			} else {
				g.writef("%s = append(%s, %s.Root)\n", sliceVar, sliceVar, callVar)
			}
		case *LetBinding:
			g.generateLetBinding(n, "")
			g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, n.Name)
		case *ForLoop:
			g.generateForLoopForSlice(n, sliceVar)
		case *IfStmt:
			g.generateIfStmtForSlice(n, sliceVar)
		case *GoCode:
			g.generateGoCode(n)
		case *GoExpr:
			elemVar := g.nextVar()
			g.writef("%s := tui.New(tui.WithText(%s))\n", elemVar, n.Code)
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
				callVar := g.generateComponentCallWithRefs(n, "")
				if n.IsStructMount {
					g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, callVar)
				} else {
					g.writef("%s = append(%s, %s.Root)\n", sliceVar, sliceVar, callVar)
				}
			case *LetBinding:
				g.generateLetBinding(n, "")
				g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, n.Name)
			case *ForLoop:
				g.generateForLoopForSlice(n, sliceVar)
			case *IfStmt:
				g.generateIfStmtForSlice(n, sliceVar)
			case *GoCode:
				g.generateGoCode(n)
			case *GoExpr:
				elemVar := g.nextVar()
				g.writef("%s := tui.New(tui.WithText(%s))\n", elemVar, n.Code)
				g.writef("%s = append(%s, %s)\n", sliceVar, sliceVar, elemVar)
			}
		}
		g.indent--
		g.writeln("}")
	} else {
		g.writeln("}")
	}
}
