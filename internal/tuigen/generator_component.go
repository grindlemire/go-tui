package tuigen

import (
	"fmt"
	"regexp"
	"sort"
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
	g.componentExprFields = nil
	g.stateVars = nil
	g.stateBindings = nil
	g.eventsVars = nil

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
			varName := g.generateElementWithRefs(n, "", false, false, false)
			if rootVar == "" {
				rootVar = varName
			}
		case *LetBinding:
			g.generateLetBinding(n, "", false, false)
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
			varName := g.generateComponentCallWithRefs(n, "", false, false)
			if rootVar == "" {
				rootVar = varName
			}
		case *ComponentExpr:
			varName := g.nextVar()
			g.writef("%s := %s.Render(app)\n", varName, n.Expr)
			if rootVar == "" {
				rootVar = varName
			}
			g.trackComponentExprField(n.Expr)
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
	g.generateUnbindApp(comp, g.fileDecls)
}

// generateFunctionComponent generates a function component (existing behavior).
// Function components return a ComponentNameView struct.
func (g *Generator) generateFunctionComponent(comp *Component) {
	// Collect refs from this component
	analyzer := NewAnalyzer()
	g.refs = analyzer.CollectRefs(comp)

	// Detect state variables, events variables, and bindings
	g.stateVars = analyzer.DetectStateVars(comp)
	g.eventsVars = analyzer.DetectEventsVars(comp)
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
	g.writef(") *%s {\n", structName)
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

	// Save buffer position before body generation so we can splice in
	// hoisted declarations for conditional component variables afterward.
	bodyStartPos := g.buf.Len()
	bodyStartLine := g.currentLine

	// Generate body nodes
	for _, node := range comp.Body {
		switch n := node.(type) {
		case *Element:
			varName := g.generateElementWithRefs(n, "", false, false, false)
			if rootVar == "" {
				rootVar = varName
			}
		case *LetBinding:
			// Let bindings create elements that are typically used as children
			// They are NOT the root element unless explicitly used
			g.generateLetBinding(n, "", false, false)
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
			varName := g.generateComponentCallWithRefs(n, "", false, false)
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

	// Hoist declarations for component variables declared inside conditional blocks.
	// These variables use = (not :=) inside the if block and need a var declaration
	// at function scope so they're accessible in the watcher/bind/unbind code below.
	g.spliceConditionalComponentHoists(bodyStartPos, bodyStartLine)

	// Emit watcher collection statements (collected during element generation)
	if len(g.componentVars) > 0 {
		g.writeln("")
		// Aggregate watchers from child component calls
		for _, cv := range g.componentVars {
			if cv.inForLoop {
				g.writef("for _, __cv := range %s_views {\n", cv.name)
				g.indent++
				g.writeln("watchers = append(watchers, __cv.GetWatchers()...)")
				g.indent--
				g.writeln("}")
			} else if cv.inConditional {
				g.writef("if %s != nil {\n", cv.name)
				g.indent++
				g.writef("watchers = append(watchers, %s.GetWatchers()...)\n", cv.name)
				g.indent--
				g.writeln("}")
			} else {
				g.writef("watchers = append(watchers, %s.GetWatchers()...)\n", cv.name)
			}
		}
	}

	// Generate state bindings (reactive updates)
	g.generateStateBindings()

	// Generate bindApp closure capturing local state, events, and child views
	g.generateBindAppClosure()

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
	g.writeln("bindApp: __bindApp,")
	g.writeln("unbindApp: __unbindApp,")
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

	g.writeln("return &view")

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
	g.writeln("bindApp  func(*tui.App)")
	g.writeln("unbindApp func()")

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

	// Generate UnbindApp method to implement tui.AppUnbinder
	g.writef("func (v *%s) UnbindApp() {\n", structName)
	g.indent++
	g.writeln("if v.unbindApp != nil {")
	g.indent++
	g.writeln("v.unbindApp()")
	g.indent--
	g.writeln("}")
	g.indent--
	g.writeln("}")
	g.writeln("")

	// All methods use pointer receivers so the mount system can store and
	// mutate cached *XxxView instances via UpdateProps.

	// Generate GetRoot() method to implement tui.Viewable
	g.writef("func (v *%s) GetRoot() *tui.Element { return v.Root }\n", structName)
	g.writeln("")

	// Generate GetWatchers() method to implement tui.Viewable
	g.writef("func (v *%s) GetWatchers() []tui.Watcher { return v.watchers }\n", structName)
	g.writeln("")

	// Generate Render() method to implement tui.Component
	g.writef("func (v *%s) Render(app *tui.App) *tui.Element { return v.Root }\n", structName)
	g.writeln("")

	// Generate BindApp method to implement tui.AppBinder
	g.writef("func (v *%s) BindApp(app *tui.App) {\n", structName)
	g.indent++
	g.writeln("if v.bindApp != nil {")
	g.indent++
	g.writeln("v.bindApp(app)")
	g.indent--
	g.writeln("}")
	g.indent--
	g.writeln("}")
	g.writeln("")

	// Generate UpdateProps method so the mount system can refresh cached views
	// with new props when cross-package function templs are mounted.
	g.writef("func (v *%s) UpdateProps(fresh tui.Component) {\n", structName)
	g.indent++
	g.writef("f, ok := fresh.(*%s)\n", structName)
	g.writeln("if !ok {")
	g.indent++
	g.writeln("return")
	g.indent--
	g.writeln("}")
	g.writeln("v.Root = f.Root")
	g.writeln("v.watchers = f.watchers")
	g.writeln("v.bindApp = f.bindApp")
	g.writeln("v.unbindApp = f.unbindApp")
	for _, ref := range refs {
		g.writef("v.%s = f.%s\n", ref.ExportName, ref.ExportName)
	}
	g.indent--
	g.writeln("}")
	g.writeln("")

	g.writef("var _ tui.AppBinder = (*%s)(nil)\n", structName)
	g.writeln("")
	g.writef("var _ tui.AppUnbinder = (*%s)(nil)\n", structName)
	g.writeln("")
	g.writef("var _ tui.PropsUpdater = (*%s)(nil)\n", structName)
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

	// Channels and functions are runtime state, not props.
	// Copying them in UpdateProps breaks watchers subscribed to the original channel.
	if strings.HasPrefix(fieldType, "chan ") ||
		strings.HasPrefix(fieldType, "<-chan ") ||
		strings.HasPrefix(fieldType, "chan<-") ||
		fieldType == "chan" ||
		strings.HasPrefix(fieldType, "func(") ||
		fieldType == "func" {
		return true
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

// hasUserBindAppMethod returns true when the source file already declares a
// BindApp method on the receiver type.
func hasUserBindAppMethod(decls []*GoDecl, funcs []*GoFunc, receiverType string) bool {
	typeName := strings.TrimPrefix(receiverType, "*")
	pattern := regexp.MustCompile(`func\s*\(\s*\w+\s+\*?` + regexp.QuoteMeta(typeName) + `\s*\)\s*BindApp\s*\(`)

	for _, decl := range decls {
		if decl.Kind == "func" && pattern.MatchString(decl.Code) {
			return true
		}
	}
	for _, fn := range funcs {
		if pattern.MatchString(fn.Code) {
			return true
		}
	}
	return false
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
		"*tui.TextArea",
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
	if hasUserBindAppMethod(decls, g.fileFuncs, comp.ReceiverType) {
		return
	}

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

	// Find *tui.App fields to set directly (c.app = app)
	var appFields []StructField
	for _, f := range fields {
		if f.Type == "*tui.App" {
			appFields = append(appFields, f)
		}
	}

	// Find fields that need BindApp (known types like State, Events, TextArea)
	var bindableFields []StructField
	for _, f := range fields {
		if isAppBindableType(f.Type) {
			bindableFields = append(bindableFields, f)
		}
	}

	// Find component expression fields that may implement AppBinder.
	// These are fields used as @receiver.field in the template (e.g., @c.settingsView).
	// Since the generator can't do full type checking, we use a runtime type assertion.
	componentExprFieldSet := make(map[string]bool)
	for _, name := range g.componentExprFields {
		componentExprFieldSet[name] = true
	}
	// Remove fields already in bindableFields to avoid duplicate binding
	for _, f := range bindableFields {
		delete(componentExprFieldSet, f.Name)
	}
	var componentBindFields []string
	for _, f := range fields {
		if componentExprFieldSet[f.Name] {
			componentBindFields = append(componentBindFields, f.Name)
		}
	}

	if len(appFields) == 0 && len(bindableFields) == 0 && len(componentBindFields) == 0 {
		return
	}

	// Get the receiver type name without pointer
	typeName := strings.TrimPrefix(comp.ReceiverType, "*")

	// Generate BindApp method with nil checks
	g.writef("func (%s) BindApp(app *tui.App) {\n", comp.Receiver)
	g.indent++
	// Set *tui.App fields directly
	for _, f := range appFields {
		g.writef("%s.%s = app\n", comp.ReceiverName, f.Name)
	}
	for _, f := range bindableFields {
		g.writef("if %s.%s != nil {\n", comp.ReceiverName, f.Name)
		g.indent++
		g.writef("%s.%s.BindApp(app)\n", comp.ReceiverName, f.Name)
		g.indent--
		g.writeln("}")
	}
	// Bind component expression fields via type assertion
	for _, name := range componentBindFields {
		g.writef("if binder, ok := any(%s.%s).(tui.AppBinder); ok {\n", comp.ReceiverName, name)
		g.indent++
		g.writeln("binder.BindApp(app)")
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

// generateUnbindApp generates an UnbindApp method for a method component.
// This allows mount sweep to detach topic-based Events subscriptions.
func (g *Generator) generateUnbindApp(comp *Component, decls []*GoDecl) {
	if hasUserBindAppMethod(decls, g.fileFuncs, comp.ReceiverType) {
		return
	}

	structDecl := findStructDecl(decls, comp.ReceiverType)
	if structDecl == nil {
		return
	}

	fields := parseStructFields(structDecl.Code)
	if len(fields) == 0 {
		return
	}

	componentExprFieldSet := make(map[string]bool)
	for _, name := range g.componentExprFields {
		componentExprFieldSet[name] = true
	}

	var unbindFields []StructField
	for _, f := range fields {
		if strings.HasPrefix(f.Type, "*tui.Events[") {
			unbindFields = append(unbindFields, f)
		}
	}
	// Remove fields already handled explicitly
	for _, f := range unbindFields {
		delete(componentExprFieldSet, f.Name)
	}
	var componentUnbindFields []string
	for _, f := range fields {
		if componentExprFieldSet[f.Name] {
			componentUnbindFields = append(componentUnbindFields, f.Name)
		}
	}

	if len(unbindFields) == 0 && len(componentUnbindFields) == 0 {
		return
	}

	typeName := strings.TrimPrefix(comp.ReceiverType, "*")
	g.writef("func (%s) UnbindApp() {\n", comp.Receiver)
	g.indent++
	for _, f := range unbindFields {
		g.writef("if %s.%s != nil {\n", comp.ReceiverName, f.Name)
		g.indent++
		g.writef("%s.%s.UnbindApp()\n", comp.ReceiverName, f.Name)
		g.indent--
		g.writeln("}")
	}
	for _, name := range componentUnbindFields {
		g.writef("if unbinder, ok := any(%s.%s).(tui.AppUnbinder); ok {\n", comp.ReceiverName, name)
		g.indent++
		g.writeln("unbinder.UnbindApp()")
		g.indent--
		g.writeln("}")
	}
	g.indent--
	g.writeln("}")
	g.writeln("")
	g.writef("var _ tui.AppUnbinder = (*%s)(nil)\n", typeName)
	g.writeln("")
}

// interfaceCheck maps a method name pattern to the tui interface it satisfies.
type interfaceCheck struct {
	// methodPattern matches the method signature in Go source code.
	methodPattern string
	// interfaceName is the tui interface (e.g., "tui.KeyListener").
	interfaceName string
}

// optionalInterfaces lists the user-facing component interfaces that are
// discovered via type assertion at runtime. A wrong method signature silently
// fails the assertion, so we emit compile-time checks when we detect these
// methods on component receiver types.
var optionalInterfaces = []interfaceCheck{
	{`\bKeyMap\s*\(`, "tui.KeyListener"},
	{`\bHandleMouse\s*\(`, "tui.MouseListener"},
	{`\bInit\s*\(`, "tui.Initializer"},
	{`\bWatchers\s*\(`, "tui.WatcherProvider"},
}

// generateInterfaceChecks emits compile-time interface satisfaction checks
// for method component receiver types that define optional interface methods.
// This catches signature mismatches (e.g., returning []tui.KeyBinding instead
// of tui.KeyMap) at compile time rather than failing silently at runtime.
func (g *Generator) generateInterfaceChecks(file *File) {
	// Collect receiver types that are method templ components.
	componentTypes := make(map[string]bool)
	for _, comp := range file.Components {
		if comp.ReceiverType != "" {
			typeName := strings.TrimPrefix(comp.ReceiverType, "*")
			componentTypes[typeName] = true
		}
	}
	if len(componentTypes) == 0 {
		return
	}

	// For each optional interface, check if any GoFunc defines a matching
	// method on a component receiver type.
	type check struct {
		typeName      string
		interfaceName string
	}
	var checks []check
	seen := make(map[check]bool)

	for _, iface := range optionalInterfaces {
		for _, typeName := range sortedKeys(componentTypes) {
			pattern := regexp.MustCompile(
				`func\s*\(\s*\w+\s+\*?` + regexp.QuoteMeta(typeName) + `\s*\)\s*` + iface.methodPattern,
			)
			for _, fn := range file.Funcs {
				if pattern.MatchString(fn.Code) {
					c := check{typeName, iface.interfaceName}
					if !seen[c] {
						seen[c] = true
						checks = append(checks, c)
					}
					break
				}
			}
		}
	}

	if len(checks) == 0 {
		return
	}

	g.writeln("// Compile-time interface satisfaction checks.")
	g.writeln("var (")
	g.indent++
	for _, c := range checks {
		g.writef("_ %s = (*%s)(nil)\n", c.interfaceName, c.typeName)
	}
	g.indent--
	g.writeln(")")
	g.writeln("")
}

// sortedKeys returns the keys of a map in sorted order.
func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// trackComponentExprField extracts and tracks receiver field names from
// component expressions (e.g., "c.settingsView" → tracks "settingsView").
// Only tracks when inside a method component (currentReceiver is set).
func (g *Generator) trackComponentExprField(expr string) {
	if g.currentReceiver == "" {
		return
	}
	prefix := g.currentReceiver + "."
	if !strings.HasPrefix(expr, prefix) {
		return
	}
	fieldName := expr[len(prefix):]
	// Only track simple field names (no further dots or method calls)
	if strings.ContainsAny(fieldName, ".()") {
		return
	}
	// Avoid duplicates
	for _, existing := range g.componentExprFields {
		if existing == fieldName {
			return
		}
	}
	g.componentExprFields = append(g.componentExprFields, fieldName)
}

// generateBindAppClosure emits a __bindApp closure for function components.
// The closure captures local state vars, events vars, and child component views,
// binding them all to the app when called.
func (g *Generator) generateBindAppClosure() {
	g.writeln("")
	g.writeln("__bindApp := func(app *tui.App) {")
	g.indent++

	// Bind local state variables
	for _, sv := range g.stateVars {
		if sv.IsParameter {
			continue // Parameters are already bound by the caller
		}
		g.writef("%s.BindApp(app)\n", sv.Name)
	}

	// Bind local events variables
	for _, ev := range g.eventsVars {
		g.writef("%s.BindApp(app)\n", ev.Name)
	}

	// Bind child function component views (they implement AppBinder)
	for _, cv := range g.componentVars {
		if cv.inForLoop {
			g.writef("for _, __cv := range %s_views {\n", cv.name)
			g.indent++
			g.writeln("if binder, ok := any(__cv).(tui.AppBinder); ok {")
			g.indent++
			g.writeln("binder.BindApp(app)")
			g.indent--
			g.writeln("}")
			g.indent--
			g.writeln("}")
		} else if cv.inConditional {
			g.writef("if %s != nil {\n", cv.name)
			g.indent++
			g.writef("if binder, ok := any(%s).(tui.AppBinder); ok {\n", cv.name)
			g.indent++
			g.writeln("binder.BindApp(app)")
			g.indent--
			g.writeln("}")
			g.indent--
			g.writeln("}")
		} else {
			g.writef("if binder, ok := any(%s).(tui.AppBinder); ok {\n", cv.name)
			g.indent++
			g.writeln("binder.BindApp(app)")
			g.indent--
			g.writeln("}")
		}
	}

	g.indent--
	g.writeln("}")

	g.writeln("")
	g.writeln("__unbindApp := func() {")
	g.indent++

	// Unbind local events variables
	for _, ev := range g.eventsVars {
		g.writef("%s.UnbindApp()\n", ev.Name)
	}

	// Unbind child function component views (they may implement AppUnbinder)
	for _, cv := range g.componentVars {
		if cv.inForLoop {
			g.writef("for _, __cv := range %s_views {\n", cv.name)
			g.indent++
			g.writeln("if unbinder, ok := any(__cv).(tui.AppUnbinder); ok {")
			g.indent++
			g.writeln("unbinder.UnbindApp()")
			g.indent--
			g.writeln("}")
			g.indent--
			g.writeln("}")
		} else if cv.inConditional {
			g.writef("if %s != nil {\n", cv.name)
			g.indent++
			g.writef("if unbinder, ok := any(%s).(tui.AppUnbinder); ok {\n", cv.name)
			g.indent++
			g.writeln("unbinder.UnbindApp()")
			g.indent--
			g.writeln("}")
			g.indent--
			g.writeln("}")
		} else {
			g.writef("if unbinder, ok := any(%s).(tui.AppUnbinder); ok {\n", cv.name)
			g.indent++
			g.writeln("unbinder.UnbindApp()")
			g.indent--
			g.writeln("}")
		}
	}

	g.indent--
	g.writeln("}")
}

// spliceConditionalComponentHoists inserts hoisted declarations at bodyStartPos
// for component variables declared inside block scopes.
// For conditional (if/else) vars: "var __tui_N *XxxView" (single pointer, nil-guarded).
// For for-loop vars: "var __tui_N_views []*XxxView" (slice, range-iterated).
func (g *Generator) spliceConditionalComponentHoists(bodyStartPos int, bodyStartLine int) {
	// Collect hoisted declarations
	var hoistLines []string
	for _, cv := range g.componentVars {
		if cv.inForLoop {
			hoistLines = append(hoistLines, fmt.Sprintf("var %s_views []%s", cv.name, viewTypeName(cv.componentName)))
		} else if cv.inConditional {
			hoistLines = append(hoistLines, fmt.Sprintf("var %s %s", cv.name, viewTypeName(cv.componentName)))
		}
	}
	if len(hoistLines) == 0 {
		return
	}

	// Save body bytes written after bodyStartPos
	bodyBytes := make([]byte, g.buf.Len()-bodyStartPos)
	copy(bodyBytes, g.buf.Bytes()[bodyStartPos:])
	g.buf.Truncate(bodyStartPos)

	// Write hoisted declarations at the saved position (uses current indent)
	hoistLineCount := len(hoistLines)
	for _, line := range hoistLines {
		g.writeln(line)
	}

	// Write body bytes back
	g.buf.Write(bodyBytes)

	// Adjust source map entries: all mappings recorded during body generation
	// have line numbers that are now shifted down by hoistLineCount.
	if g.sourceMap != nil {
		for i := range g.sourceMap.Mappings {
			if g.sourceMap.Mappings[i].GoLine >= bodyStartLine {
				g.sourceMap.Mappings[i].GoLine += hoistLineCount
			}
		}
	}
}

// spliceForLoopViewResets inserts slice reset statements at resetPos for any
// for-loop component variables added after prevVarCount. This prevents unbounded
// growth of views slices in reactive for-loop closures that fire on every state change.
func (g *Generator) spliceForLoopViewResets(resetPos int, resetLine int, prevVarCount int) {
	var resetLines []string
	for _, cv := range g.componentVars[prevVarCount:] {
		if cv.inForLoop {
			resetLines = append(resetLines, fmt.Sprintf("%s_views = %s_views[:0]", cv.name, cv.name))
		}
	}
	if len(resetLines) == 0 {
		return
	}

	// Save bytes written after resetPos
	tailBytes := make([]byte, g.buf.Len()-resetPos)
	copy(tailBytes, g.buf.Bytes()[resetPos:])
	g.buf.Truncate(resetPos)

	// Write reset statements at the saved position
	resetLineCount := len(resetLines)
	for _, line := range resetLines {
		g.writeln(line)
	}

	// Write tail bytes back
	g.buf.Write(tailBytes)

	// Adjust source map entries shifted by the spliced lines.
	if g.sourceMap != nil {
		for i := range g.sourceMap.Mappings {
			if g.sourceMap.Mappings[i].GoLine >= resetLine {
				g.sourceMap.Mappings[i].GoLine += resetLineCount
			}
		}
	}
}
