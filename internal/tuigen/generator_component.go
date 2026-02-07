package tuigen

// generateComponent generates a Go function from a component.
// Dispatches to generateMethodComponent for method templs (has receiver)
// or generateFunctionComponent for function templs (no receiver).
func (g *Generator) generateComponent(comp *Component) {
	// Reset variable counter and watcher tracking for each component
	g.varCounter = 0
	g.condCounter = 0
	g.loopCounter = 0
	g.mountIndex = 0
	g.currentReceiver = ""
	g.componentVars = nil
	g.stateVars = nil
	g.stateBindings = nil

	if comp.Receiver != "" {
		g.generateMethodComponent(comp)
	} else {
		g.generateFunctionComponent(comp)
	}
}

// generateMethodComponent generates a Render() method on a struct receiver.
// Method components return *tui.Element directly — no view struct, no watcher
// aggregation. The receiver variable is available for expressions in the template.
//
// Generated form:
//
//	func (s *sidebar) Render() *tui.Element { ... return __tui_0 }
func (g *Generator) generateMethodComponent(comp *Component) {
	g.currentReceiver = comp.ReceiverName
	defer func() { g.currentReceiver = "" }()

	// Method signature: func (recv) Render() *tui.Element
	g.writef("func (%s) Render() *tui.Element {\n", comp.Receiver)
	g.indent++

	// Track the root element variable name
	var rootVar string

	// Generate body nodes
	for _, node := range comp.Body {
		switch n := node.(type) {
		case *Element:
			varName := g.generateElementWithRefs(n, "", false, false)
			if rootVar == "" {
				rootVar = varName
			}
		case *LetBinding:
			g.generateLetBinding(n, "")
		case *ForLoop:
			g.generateForLoopWithRefs(n, "", false, false)
		case *IfStmt:
			g.generateIfStmtWithRefs(n, "", false)
		case *GoCode:
			g.generateGoCode(n)
		case *GoExpr:
			g.writef("%s\n", n.Code)
		case *ComponentCall:
			varName := g.generateComponentCallWithRefs(n, "")
			if rootVar == "" {
				rootVar = varName
			}
		}
	}

	// Return the root element directly
	g.writeln("")
	if rootVar != "" {
		g.writef("return %s\n", rootVar)
	} else {
		g.writeln("return nil")
	}

	g.indent--
	g.writeln("}")
	g.writeln("")
}

// generateFunctionComponent generates a function component (existing behavior).
// Function components return a ComponentNameView struct.
func (g *Generator) generateFunctionComponent(comp *Component) {
	// Collect refs from this component
	analyzer := NewAnalyzer()
	g.refs = analyzer.CollectRefs(comp)

	// Detect state variables and bindings
	g.stateVars = analyzer.DetectStateVars(comp)
	g.stateBindings = analyzer.DetectStateBindings(comp, g.stateVars)

	// Generate view struct for this component (always generated)
	structName := comp.Name + "View"
	g.generateViewStruct(comp.Name, g.refs)

	// Generate function signature - always returns struct
	g.writef("func %s(", comp.Name)
	for i, param := range comp.Params {
		if i > 0 {
			g.write(", ")
		}
		g.writef("%s %s", param.Name, param.Type)
	}
	// Add children parameter if component accepts children
	if comp.AcceptsChildren {
		if len(comp.Params) > 0 {
			g.write(", ")
		}
		g.write("children []*tui.Element")
	}
	g.writef(") %s {\n", structName)
	g.indent++

	// Pre-declare view variable so closures can capture it
	g.writef("var view %s\n", structName)
	g.writeln("var watchers []tui.Watcher")
	g.writeln("")

	// No forward declarations needed — refs are user-declared Go variables
	// (e.g., content := tui.NewRef()) written in the component body Go code

	// Track the root element variable name
	// The root is the first top-level Element (not LetBinding, which is typically a child reference)
	var rootVar string
	var rootIsComponent bool // Whether root is a component call (needs .Root accessor)

	// Generate body nodes
	for _, node := range comp.Body {
		switch n := node.(type) {
		case *Element:
			varName := g.generateElementWithRefs(n, "", false, false)
			if rootVar == "" {
				rootVar = varName
			}
		case *LetBinding:
			// @let bindings create elements that are typically used as children
			// They are NOT the root element unless explicitly used
			g.generateLetBinding(n, "")
		case *ForLoop:
			g.generateForLoopWithRefs(n, "", false, false)
		case *IfStmt:
			g.generateIfStmtWithRefs(n, "", false)
		case *GoCode:
			g.generateGoCode(n)
		case *GoExpr:
			// A bare expression in component body - treat as statement
			g.writef("%s\n", n.Code)
		case *ComponentCall:
			varName := g.generateComponentCallWithRefs(n, "")
			if rootVar == "" {
				rootVar = varName
				rootIsComponent = true
			}
		}
	}

	// Emit watcher collection statements (collected during element generation)
	if len(g.componentVars) > 0 {
		g.writeln("")
		// Aggregate watchers from child component calls
		for _, compVar := range g.componentVars {
			g.writef("watchers = append(watchers, %s.GetWatchers()...)\n", compVar)
		}
	}

	// Generate state bindings (reactive updates)
	g.generateStateBindings()

	// Populate view struct before returning
	g.writeln("")
	g.writef("view = %s{\n", structName)
	g.indent++
	if rootVar != "" {
		if rootIsComponent {
			g.writef("Root: %s.Root,\n", rootVar)
		} else {
			g.writef("Root: %s,\n", rootVar)
		}
	} else {
		g.writeln("Root: nil,")
	}
	g.writeln("watchers: watchers,")
	for _, ref := range g.refs {
		// View struct exposes *tui.Element (not ref types)
		switch ref.RefKind {
		case RefSingle:
			g.writef("%s: %s.El(),\n", ref.ExportName, ref.Name)
		case RefList:
			g.writef("%s: %s.All(),\n", ref.ExportName, ref.Name)
		case RefMap:
			g.writef("%s: %s.All(),\n", ref.ExportName, ref.Name)
		}
	}
	g.indent--
	g.writeln("}")

	g.writeln("return view")

	g.indent--
	g.writeln("}")
	g.writeln("")
}

// generateViewStruct generates the ComponentNameView struct definition.
func (g *Generator) generateViewStruct(compName string, refs []RefInfo) {
	structName := compName + "View"

	g.writef("type %s struct {\n", structName)
	g.indent++
	g.writeln("Root     *tui.Element")
	g.writeln("watchers []tui.Watcher")

	for _, ref := range refs {
		switch ref.RefKind {
		case RefSingle:
			if ref.InConditional {
				g.writef("%s *tui.Element // may be nil\n", ref.ExportName)
			} else {
				g.writef("%s *tui.Element\n", ref.ExportName)
			}
		case RefList:
			g.writef("%s []*tui.Element\n", ref.ExportName)
		case RefMap:
			g.writef("%s map[%s]*tui.Element\n", ref.ExportName, ref.KeyType)
		}
	}

	g.indent--
	g.writeln("}")
	g.writeln("")

	// Generate GetRoot() method to implement tui.Viewable
	g.writef("func (v %s) GetRoot() tui.Renderable { return v.Root }\n", structName)
	g.writeln("")

	// Generate GetWatchers() method to implement tui.Viewable
	g.writef("func (v %s) GetWatchers() []tui.Watcher { return v.watchers }\n", structName)
	g.writeln("")
}
