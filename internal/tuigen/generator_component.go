package tuigen

import (
	"regexp"
	"strings"
)

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

// generateMethodComponent generates a Render(app *tui.App) method on a struct receiver.
// Method components return *tui.Element directly — no view struct, no watcher
// aggregation. The receiver variable is available for expressions in the template.
//
// Generated form:
//
//	func (s *sidebar) Render(app *tui.App) *tui.Element { ... return __tui_0 }
func (g *Generator) generateMethodComponent(comp *Component) {
	g.currentReceiver = comp.ReceiverName
	defer func() { g.currentReceiver = "" }()

	// Method signature: func (recv) Render(app *tui.App) *tui.Element
	g.writef("func (%s) Render(app *tui.App) *tui.Element {\n", comp.Receiver)
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
			if rootVar == "" {
				rootVar = g.nextVar()
				g.writef("var %s *tui.Element\n", rootVar)
			}
			g.generateForLoopToRoot(n, rootVar, false)
		case *IfStmt:
			if rootVar == "" {
				rootVar = g.nextVar()
				g.writef("var %s *tui.Element\n", rootVar)
			}
			g.generateIfStmtToRoot(n, rootVar, false)
		case *GoCode:
			g.generateGoCode(n)
		case *GoExpr:
			g.writef("%s\n", n.Code)
		case *ComponentCall:
			varName := g.generateComponentCallWithRefs(n, "")
			if rootVar == "" {
				rootVar = varName
			}
		case *ComponentExpr:
			varName := g.nextVar()
			g.writef("%s := %s.Render(app)\n", varName, n.Expr)
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

	// Generate UpdateProps method for prop updates on cached components
	g.generateUpdateProps(comp, g.fileDecls)

	// Generate BindApp method for app binding on State/Events fields
	g.generateBindApp(comp, g.fileDecls)
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
			if rootVar == "" {
				rootVar = g.nextVar()
				g.writef("var %s *tui.Element\n", rootVar)
			}
			g.generateForLoopToRoot(n, rootVar, false)
		case *IfStmt:
			if rootVar == "" {
				rootVar = g.nextVar()
				g.writef("var %s *tui.Element\n", rootVar)
			}
			g.generateIfStmtToRoot(n, rootVar, false)
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
		case *ComponentExpr:
			varName := g.nextVar()
			g.writef("%s := %s.Render(app)\n", varName, n.Expr)
			if rootVar == "" {
				rootVar = varName
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

// StructField represents a parsed field from a struct definition.
type StructField struct {
	Name string
	Type string
}

// parseStructFields extracts field names and types from a struct definition.
// Input: "type foo struct {\n    field1 Type1\n    field2 Type2\n}"
// Returns: [{Name: "field1", Type: "Type1"}, {Name: "field2", Type: "Type2"}]
func parseStructFields(structCode string) []StructField {
	var fields []StructField

	// Find the struct body between { and }
	start := strings.Index(structCode, "{")
	end := strings.LastIndex(structCode, "}")
	if start == -1 || end == -1 || start >= end {
		return fields
	}

	body := structCode[start+1 : end]
	lines := strings.Split(body, "\n")

	// Pattern to match field declarations: name type or name, name2 type
	// Handles: "field Type", "field *Type", "field Type // comment"
	fieldPattern := regexp.MustCompile(`^\s*(\w+)\s+(\S+.*)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Remove trailing comments
		if idx := strings.Index(line, "//"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		matches := fieldPattern.FindStringSubmatch(line)
		if len(matches) >= 3 {
			fields = append(fields, StructField{
				Name: matches[1],
				Type: strings.TrimSpace(matches[2]),
			})
		}
	}

	return fields
}

// isInternalStateType returns true if the type is an internal state type
// that should NOT be updated via UpdateProps.
func isInternalStateType(fieldType string) bool {
	// These types are internal state that should be preserved across re-renders
	internalTypes := []string{
		"*tui.Ref",
		"*tui.RefList",
		"*tui.RefMap",
		"*tui.State",
		"tui.Ref",
		"tui.RefList",
		"tui.RefMap",
		"tui.State",
	}

	for _, t := range internalTypes {
		if strings.HasPrefix(fieldType, t) {
			return true
		}
	}

	return false
}

// findStructDecl finds the struct declaration for a given type name in the file's declarations.
func findStructDecl(decls []*GoDecl, typeName string) *GoDecl {
	// Remove pointer prefix if present
	typeName = strings.TrimPrefix(typeName, "*")

	// Look for "type TypeName struct"
	pattern := regexp.MustCompile(`type\s+` + regexp.QuoteMeta(typeName) + `\s+struct\s*\{`)

	for _, decl := range decls {
		if decl.Kind == "type" && pattern.MatchString(decl.Code) {
			return decl
		}
	}
	return nil
}

// generateUpdateProps generates an UpdateProps method for a method component.
// This allows Mount to update cached component instances with fresh props.
func (g *Generator) generateUpdateProps(comp *Component, decls []*GoDecl) {
	// Find the struct declaration for this component's receiver type
	structDecl := findStructDecl(decls, comp.ReceiverType)
	if structDecl == nil {
		return // No struct found, skip generating UpdateProps
	}

	// Parse the struct fields
	fields := parseStructFields(structDecl.Code)
	if len(fields) == 0 {
		return
	}

	// Find prop fields (non-internal-state types)
	var propFields []StructField
	for _, f := range fields {
		if !isInternalStateType(f.Type) {
			propFields = append(propFields, f)
		}
	}

	if len(propFields) == 0 {
		return // No props to update
	}

	// Get the receiver type name without pointer
	typeName := strings.TrimPrefix(comp.ReceiverType, "*")

	// Generate UpdateProps method
	g.writef("func (%s) UpdateProps(fresh tui.Component) {\n", comp.Receiver)
	g.indent++
	g.writef("f, ok := fresh.(%s)\n", comp.ReceiverType)
	g.writeln("if !ok {")
	g.indent++
	g.writeln("return")
	g.indent--
	g.writeln("}")

	// Copy each prop field
	for _, f := range propFields {
		g.writef("%s.%s = f.%s\n", comp.ReceiverName, f.Name, f.Name)
	}

	g.indent--
	g.writeln("}")
	g.writeln("")

	// Add a compile-time check that the type implements PropsUpdater
	g.writef("var _ tui.PropsUpdater = (*%s)(nil)\n", typeName)
	g.writeln("")
}

// isAppBindableType returns true if the field type has a BindApp method
// (i.e., *tui.State[...] or *tui.Events[...]).
func isAppBindableType(fieldType string) bool {
	bindableTypes := []string{
		"*tui.State[",
		"*tui.Events[",
	}
	for _, t := range bindableTypes {
		if strings.HasPrefix(fieldType, t) {
			return true
		}
	}
	return false
}

// generateBindApp generates a BindApp method for a method component.
// This allows the mount system to bind the app to State/Events fields.
func (g *Generator) generateBindApp(comp *Component, decls []*GoDecl) {
	// Find the struct declaration for this component's receiver type
	structDecl := findStructDecl(decls, comp.ReceiverType)
	if structDecl == nil {
		return
	}

	// Parse the struct fields
	fields := parseStructFields(structDecl.Code)
	if len(fields) == 0 {
		return
	}

	// Find fields that need BindApp
	var bindableFields []StructField
	for _, f := range fields {
		if isAppBindableType(f.Type) {
			bindableFields = append(bindableFields, f)
		}
	}

	if len(bindableFields) == 0 {
		return
	}

	// Get the receiver type name without pointer
	typeName := strings.TrimPrefix(comp.ReceiverType, "*")

	// Generate BindApp method with nil checks
	g.writef("func (%s) BindApp(app *tui.App) {\n", comp.Receiver)
	g.indent++
	for _, f := range bindableFields {
		g.writef("if %s.%s != nil {\n", comp.ReceiverName, f.Name)
		g.indent++
		g.writef("%s.%s.BindApp(app)\n", comp.ReceiverName, f.Name)
		g.indent--
		g.writeln("}")
	}
	g.indent--
	g.writeln("}")
	g.writeln("")

	// Add a compile-time check that the type implements AppBinder
	g.writef("var _ tui.AppBinder = (*%s)(nil)\n", typeName)
	g.writeln("")
}
