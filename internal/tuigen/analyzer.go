package tuigen

import (
	"fmt"
	"regexp"
	"strings"
)

// StateVar tracks information about a state variable declaration.
type StateVar struct {
	Name        string   // Variable name (e.g., "count")
	Type        string   // Go type (e.g., "int", "string", "[]Item")
	InitExpr    string   // Initialization expression (empty for parameters)
	IsParameter bool     // True if state is passed as component parameter
	Position    Position // Position of the declaration
}

// EventsVar tracks information about an events variable declaration.
type EventsVar struct {
	Name     string   // Variable name (e.g., "events")
	Position Position // Position of the declaration
}

// StateBinding tracks a binding between state variables and an element.
type StateBinding struct {
	StateVars    []string // State variables referenced in expression
	Element      *Element // Element that uses this expression
	ElementName  string   // Generated variable name for the element (e.g., "__tui_0")
	Attribute    string   // Which attribute ("text", "class", etc.)
	Expr         string   // The expression (e.g., "fmt.Sprintf(...)")
	ExplicitDeps bool     // True if deps={...} was used
}

// RefKind describes how a ref should be generated.
type RefKind int

const (
	RefSingle RefKind = iota // Single ref: tui.NewRef()
	RefList                  // Loop ref without key: tui.NewRefList()
	RefMap                   // Loop ref with key: tui.NewRefMap[K]()
)

// RefInfo tracks information about an element reference declared via ref={}.
type RefInfo struct {
	Name          string // Variable name from ref={name} (e.g., "content")
	ExportName    string // Capitalized for View struct (e.g., "Content")
	Element       *Element
	InLoop        bool    // true = generate slice or map type
	InConditional bool    // true = may be nil at runtime
	KeyExpr       string  // if set, generate map[KeyType]*element.Element
	KeyType       string  // inferred type of key expression (e.g., "string", "int")
	RefKind       RefKind // RefSingle, RefList, RefMap
	Position      Position
}

// Analyzer performs semantic analysis on parsed .tui ASTs.
// It validates element tags, attributes, and ensures required imports are present.
type Analyzer struct {
	errors *ErrorList
	file   *File

	// Track used features to determine required imports
	usesElement bool
	usesLayout  bool
	usesTUI     bool

	// Track := bindings for unused variable detection
	letBindings map[string]bool // name -> used

	// Track component definitions for children validation
	componentDefs map[string]bool // name -> accepts children

	// currentComponent is the component whose body is being walked, used to
	// reject component elements in function templs (no receiver to mount against).
	currentComponent *Component

	// structComponentFactories holds the names of local functions that return a
	// struct-component type (a `func Name(...) *T` where T has a method templ).
	// Calling one via @Name() in a function templ generates broken code, so the
	// set lets analyzeComponentCall reject it.
	structComponentFactories map[string]bool
}

// NewAnalyzer creates a new semantic analyzer.
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		errors:        NewErrorList(),
		letBindings:   make(map[string]bool),
		componentDefs: make(map[string]bool),
	}
}

// knownTags lists all supported element tags (HTML-style).
var knownTags = map[string]bool{
	"div":      true,
	"span":     true,
	"p":        true,
	"ul":       true,
	"li":       true,
	"button":   true,
	"input":    true,
	"table":    true,
	"tr":       true,
	"td":       true,
	"th":       true,
	"progress": true,
	"textarea": true,
	"hr":       true,
	"br":       true,
	"modal":    true,
	"markdown": true,
}

// voidElements lists elements that cannot have children.
var voidElements = map[string]bool{
	"hr":       true,
	"br":       true,
	"input":    true,
	"textarea": true,
	"markdown": true,
}

// knownAttributes lists all supported element attributes.
var knownAttributes = map[string]bool{
	// Dimensions
	"width":         true,
	"widthPercent":  true,
	"height":        true,
	"heightPercent": true,
	"minWidth":      true,
	"minHeight":     true,
	"maxWidth":      true,
	"maxHeight":     true,

	// Flex container
	"direction":    true,
	"justify":      true,
	"align":        true,
	"gap":          true,
	"flexWrap":     true,
	"alignContent": true,

	// Flex item
	"flexGrow":   true,
	"flexShrink": true,
	"alignSelf":  true,

	// Spacing
	"padding": true,
	"margin":  true,

	// Visual
	"border":             true,
	"borderStyle":        true,
	"borderTitle":        true,
	"background":         true,
	"backgroundGradient": true,

	// Text
	"text":      true,
	"textStyle": true,
	"textAlign": true,

	// Focus
	"onFocus":   true,
	"onBlur":    true,
	"focusable": true,
	"autoFocus": true,

	// Scroll
	"scrollable":          true,
	"scrollOffset":        true,
	"scrollbarStyle":      true,
	"scrollbarThumbStyle": true,
	"hideScrollbar":       true,

	// Generic
	"disabled": true,
	"id":       true,

	// Tailwind-style class attribute
	"class": true,

	// State reactive bindings
	"deps": true, // explicit dependencies for reactive bindings

	// Element refs
	"ref": true, // ref={varName} for element references
	"key": true, // key={expr}: identity for mounts (own, or descendants' when on a container); RefMap key with ref

	// Modal
	"open":                 true,
	"backdrop":             true,
	"closeOnEscape":        true,
	"closeOnBackdropClick": true,
	"trapFocus":            true,
	"keyMap":               true,

	// Activation
	"onActivate": true,

	// Textarea / Input
	"placeholder":      true,
	"placeholderStyle": true,
	"cursor":           true,
	"focusColor":       true,
	"borderGradient":   true,
	"focusGradient":    true,
	"submitKey":        true,
	"onSubmit":         true,
	"value":            true,
	"onChange":         true,

	// Markdown
	"source": true,
	"state":  true,
	"theme":  true,
}

// stateGetRegex matches state.Get() calls to detect state usage in expressions.
// This pattern handles:
// - Simple: count.Get()
// - Dereferenced pointer: (*count).Get()
var stateGetRegex = regexp.MustCompile(`(?:\(\*(\w+)\)|(\w+))\.Get\(\)`)

// attributeSimilar maps common typos to correct attribute names.
var attributeSimilar = map[string]string{
	"colour":       "color",
	"color":        "background",
	"onfocus":      "onFocus",
	"onblur":       "onBlur",
	"flexgrow":     "flexGrow",
	"flexshrink":   "flexShrink",
	"textstyle":    "textStyle",
	"textalign":    "textAlign",
	"alignself":    "alignSelf",
	"flexwrap":     "flexWrap",
	"aligncontent": "alignContent",
	"borderstyle":  "borderStyle",
}

// Analyze performs semantic analysis on a parsed file.
// Returns a list of errors/warnings found during analysis.
// Also modifies the file to add missing imports and transform element references.
func (a *Analyzer) Analyze(file *File) error {
	a.errors = NewErrorList()
	a.file = file
	a.letBindings = make(map[string]bool)
	a.componentDefs = make(map[string]bool)
	a.currentComponent = nil
	a.usesElement = false
	a.usesLayout = false
	a.usesTUI = false

	// First pass: scan components for {children...} and collect definitions
	for _, comp := range file.Components {
		comp.AcceptsChildren = a.containsChildrenSlot(comp.Body)
		a.componentDefs[comp.Name] = comp.AcceptsChildren
	}

	// Resolve which local factory functions return a struct component, so
	// analyzeComponentCall can reject @Factory() calls in function templs.
	a.structComponentFactories = collectStructComponentFactories(file)

	// Validate method templs using {children...} have a children field on their struct
	for _, comp := range file.Components {
		if comp.Receiver != "" && comp.AcceptsChildren {
			a.validateChildrenField(comp, file.Decls)
		}
	}

	// Second pass: collect := binding names from all components
	for _, comp := range file.Components {
		a.collectLetBindings(comp.Body)
	}

	// Third pass: transform GoExpr references to := bindings into RawGoExpr
	for _, comp := range file.Components {
		comp.Body = a.transformElementRefs(comp.Body)
	}

	// Fourth pass: validate refs
	for _, comp := range file.Components {
		a.validateRefs(comp)
	}

	// Fifth pass: validate elements and attributes
	for _, comp := range file.Components {
		a.analyzeComponent(comp)
	}

	// Check for unused := bindings
	for name, used := range a.letBindings {
		if !used {
			// This is a warning, not an error - but we'll still report it
			// For now, we'll skip this as it might have false positives
			_ = name
		}
	}

	// Add missing imports
	a.addMissingImports()

	return a.errors.Err()
}

// Errors returns the errors found during analysis.
func (a *Analyzer) Errors() *ErrorList {
	return a.errors
}

// analyzeComponent validates a single component.
func (a *Analyzer) analyzeComponent(comp *Component) {
	// Track that we use elements
	a.usesElement = true

	// Record the enclosing component so element validation can tell whether it
	// is a function templ (no receiver) or a struct component.
	a.currentComponent = comp

	// Analyze body nodes
	for _, node := range comp.Body {
		a.analyzeNode(node)
	}
}

// analyzeNode validates an AST node.
func (a *Analyzer) analyzeNode(node Node) {
	switch n := node.(type) {
	case *Element:
		a.analyzeElement(n)
	case *LetBinding:
		a.analyzeLetBinding(n)
	case *ForLoop:
		a.analyzeForLoop(n)
	case *IfStmt:
		a.analyzeIfStmt(n)
	case *GoExpr:
		a.analyzeGoExpr(n)
	case *GoCode:
		a.analyzeGoCode(n)
	case *ComponentCall:
		a.analyzeComponentCall(n)
	case *ComponentExpr:
		a.analyzeComponentExpr(n)
	case *ChildrenSlot:
		// ChildrenSlot is valid - no additional validation needed
	}
}

// analyzeComponentExpr validates a component expression (@expr). Rendering it
// lowers to expr.Render(app), which a function templ cannot satisfy: its body
// has no app in scope. Only struct components (with a Render(app) receiver) can.
func (a *Analyzer) analyzeComponentExpr(expr *ComponentExpr) {
	a.rejectComponentExprInFunctionTempl(expr.Expr, expr.Position)
}

// rejectComponentExprInFunctionTempl reports the shared error for a component
// expression used where no receiver is available. It also covers the
// `name := @expr` let-binding form, which the generator lowers the same way.
func (a *Analyzer) rejectComponentExprInFunctionTempl(exprText string, pos Position) {
	if a.currentComponent != nil && a.currentComponent.Receiver == "" {
		a.errors.Add(NewErrorWithHint(pos,
			fmt.Sprintf("component expression @%s can only be used inside a struct component", exprText),
			fmt.Sprintf("component expressions render against a receiver; give %s a receiver, e.g. templ (c *T) Render() { ... }", a.currentComponent.Name)))
	}
}

// analyzeElement validates an element and its children.
func (a *Analyzer) analyzeElement(elem *Element) {
	// Check if tag is known
	if !knownTags[elem.Tag] {
		a.errors.AddErrorf(elem.Position, "unknown element tag <%s>", elem.Tag)
	}

	// Component elements mount via app.Mount against the host component's
	// receiver. A function templ has no receiver, so this cannot be generated.
	if isComponentElement(elem.Tag) && a.currentComponent != nil && a.currentComponent.Receiver == "" {
		a.errors.Add(NewErrorWithHint(elem.Position,
			fmt.Sprintf("<%s> can only be used inside a struct component", elem.Tag),
			fmt.Sprintf("component elements mount against a receiver; give %s a receiver, e.g. templ (c *T) Render() { ... }", a.currentComponent.Name)))
	}

	// Check for children on void elements
	if voidElements[elem.Tag] && len(elem.Children) > 0 {
		a.errors.AddErrorf(elem.Position,
			"<%s> is a void element and cannot have children", elem.Tag)
	}

	// Validate table element hierarchy
	switch elem.Tag {
	case "tr":
		for _, child := range elem.Children {
			if childElem, ok := child.(*Element); ok {
				if childElem.Tag != "td" && childElem.Tag != "th" {
					a.errors.AddErrorf(childElem.Position,
						"<tr> can only contain <td> or <th> children, found <%s>", childElem.Tag)
				}
			}
		}
	case "table":
		for _, child := range elem.Children {
			if childElem, ok := child.(*Element); ok {
				if childElem.Tag != "tr" && childElem.Tag != "hr" {
					a.errors.AddErrorf(childElem.Position,
						"<table> can only contain <tr> and <hr> children, found <%s>", childElem.Tag)
				}
			}
		}
	case "td", "th":
		// Validated from the tr parent above
	}

	// Check attributes
	for _, attr := range elem.Attributes {
		a.analyzeAttribute(attr, elem.Tag)
	}

	// Analyze children
	for _, child := range elem.Children {
		a.analyzeNode(child)
	}
}

// analyzeAttribute validates an element attribute.
func (a *Analyzer) analyzeAttribute(attr *Attribute, tagName string) {
	if !knownAttributes[attr.Name] {
		err := NewError(attr.Position, "unknown attribute "+attr.Name)

		// Check for similar attribute name (typo)
		if similar, ok := attributeSimilar[strings.ToLower(attr.Name)]; ok {
			err.Hint = "did you mean " + similar + "?"
		}

		a.errors.Add(err)
		return
	}

	// The parser lifts expression-valued key/ref out of Attributes, so any
	// surviving here hold a literal the generator would silently ignore.
	if attr.Name == "key" || attr.Name == "ref" {
		err := NewError(attr.Position, attr.Name+" must be an expression")
		err.Hint = "use " + attr.Name + "={...} with a Go expression, not a literal"
		a.errors.Add(err)
		return
	}

	// Check if class attribute uses Tailwind classes that need imports
	if attr.Name == "class" {
		if v, ok := attr.Value.(*StringLit); ok {
			result := ParseTailwindClasses(v.Value)
			if result.NeedsImports["tui"] {
				a.usesTUI = true
			}

			// Validate individual Tailwind classes and report errors
			classesWithPos := ParseTailwindClassesWithPositions(v.Value, 0)
			for _, cwp := range classesWithPos {
				if !cwp.Valid {
					// Calculate the position of this specific class within the attribute value
					// attr.ValuePosition is the start of the string content (after the opening quote)
					classPos := Position{
						File:   attr.ValuePosition.File,
						Line:   attr.ValuePosition.Line,
						Column: attr.ValuePosition.Column + cwp.StartCol,
					}
					classEndPos := Position{
						File:   attr.ValuePosition.File,
						Line:   attr.ValuePosition.Line,
						Column: attr.ValuePosition.Column + cwp.EndCol,
					}

					msg := "unknown Tailwind class \"" + cwp.Class + "\""
					var err *Error
					if cwp.Suggestion != "" {
						err = NewErrorWithRangeAndHint(classPos, classEndPos, msg, "did you mean \""+cwp.Suggestion+"\"?")
					} else {
						err = NewErrorWithRange(classPos, classEndPos, msg)
					}
					a.errors.Add(err)
				}
			}
		}
		return
	}

	// Check if attribute value uses layout package
	if v, ok := attr.Value.(*GoExpr); ok {
		if strings.Contains(v.Code, "layout.") {
			a.usesLayout = true
		}
		if strings.Contains(v.Code, "tui.") {
			a.usesTUI = true
		}
	}
}

// analyzeLetBinding validates a let binding.
func (a *Analyzer) analyzeLetBinding(let *LetBinding) {
	// Register the binding
	a.letBindings[let.Name] = false

	if let.Element != nil {
		a.analyzeElement(let.Element)
	}
	if let.Call != nil {
		a.analyzeComponentCall(let.Call)
	}
	if let.Expr != "" {
		// A let binding's Expr is only set for a `name := @expr` component
		// expression, so it needs the same receiver check as a bare @expr.
		a.rejectComponentExprInFunctionTempl(let.Expr, let.Position)
		if strings.Contains(let.Expr, "tui.") {
			a.usesTUI = true
		}
	}
}

// analyzeForLoop validates a for loop.
func (a *Analyzer) analyzeForLoop(loop *ForLoop) {
	// Analyze body
	for _, node := range loop.Body {
		a.analyzeNode(node)
	}
}

// analyzeIfStmt validates an if statement.
func (a *Analyzer) analyzeIfStmt(stmt *IfStmt) {
	// Analyze then branch
	for _, node := range stmt.Then {
		a.analyzeNode(node)
	}

	// Analyze else branch
	for _, node := range stmt.Else {
		a.analyzeNode(node)
	}
}

// analyzeComponentCall validates a component call.
func (a *Analyzer) analyzeComponentCall(call *ComponentCall) {
	// A @Factory() that returns a struct component mounts via app.Mount against
	// the caller's receiver. In a function templ there is no receiver, so the
	// generator falls back to a plain call and accesses .Root on a type that has
	// none. Reject it here with a clear message instead.
	if a.currentComponent != nil && a.currentComponent.Receiver == "" && a.structComponentFactories[call.Name] {
		a.errors.Add(NewErrorWithHint(call.Position,
			fmt.Sprintf("@%s() mounts a struct component and can only be used inside a struct component", call.Name),
			fmt.Sprintf("%s returns a struct component; give %s a receiver, e.g. templ (c *T) Render() { ... }", call.Name, a.currentComponent.Name)))
	}

	// Check if component is defined in this file
	acceptsChildren, defined := a.componentDefs[call.Name]

	if defined {
		// Validate children usage
		if len(call.Children) > 0 && !acceptsChildren {
			a.errors.AddErrorf(call.Position,
				"component %s does not accept children (no {children...} slot in definition)",
				call.Name)
		}
	}
	// Note: if component is not defined in this file, it might be imported
	// We let the Go compiler catch undefined references

	// Check if args reference layout or tui packages
	if strings.Contains(call.Args, "layout.") {
		a.usesLayout = true
	}
	if strings.Contains(call.Args, "tui.") {
		a.usesTUI = true
	}

	// Analyze children recursively
	for _, child := range call.Children {
		a.analyzeNode(child)
	}
}

// factoryReturnPattern matches a top-level factory function and captures its
// name and single return type: `func Widget(...) *widget {`. The lazy param
// group backtracks past nested parens in the parameter list to reach the real
// return type. Tuple returns and generic result types do not match and stay
// unflagged, which keeps the check free of false positives on ordinary code.
var factoryReturnPattern = regexp.MustCompile(`^func\s+(\w+)\s*\([^{]*?\)\s*(\*?[\w.]+)\s*\{`)

// collectStructComponentFactories returns the names of local functions whose
// return type is a struct-component receiver type (a type with a method templ).
// These are the @Name() calls that only work inside a struct component.
func collectStructComponentFactories(file *File) map[string]bool {
	structTypes := make(map[string]bool)
	for _, comp := range file.Components {
		if comp.Receiver != "" {
			structTypes[strings.TrimPrefix(comp.ReceiverType, "*")] = true
		}
	}

	factories := make(map[string]bool)
	if len(structTypes) == 0 {
		return factories
	}
	for _, fn := range file.Funcs {
		m := factoryReturnPattern.FindStringSubmatch(strings.TrimSpace(fn.Code))
		if m == nil {
			continue
		}
		if structTypes[strings.TrimPrefix(m[2], "*")] {
			factories[m[1]] = true
		}
	}
	return factories
}

// analyzeGoExpr validates a Go expression.
func (a *Analyzer) analyzeGoExpr(expr *GoExpr) {
	// Check if expression references layout or tui packages
	if strings.Contains(expr.Code, "layout.") {
		a.usesLayout = true
	}
	if strings.Contains(expr.Code, "tui.") {
		a.usesTUI = true
	}

	// Check if expression references a := binding
	for name := range a.letBindings {
		if strings.Contains(expr.Code, name) {
			a.letBindings[name] = true
		}
	}
}

// analyzeGoCode validates raw Go code.
func (a *Analyzer) analyzeGoCode(code *GoCode) {
	// Check if code references layout or tui packages
	if strings.Contains(code.Code, "layout.") {
		a.usesLayout = true
	}
	if strings.Contains(code.Code, "tui.") {
		a.usesTUI = true
	}

	// Check if code references a := binding
	for name := range a.letBindings {
		if strings.Contains(code.Code, name) {
			a.letBindings[name] = true
		}
	}
}

// addMissingImports adds required imports that are missing from the file.
func (a *Analyzer) addMissingImports() {
	// Check if root tui package is already imported
	hasTUI := false

	for _, imp := range a.file.Imports {
		if imp.Path == "github.com/grindlemire/go-tui" {
			hasTUI = true
		}
	}

	// Add root import if any tui features are used and not already imported
	if (a.usesElement || a.usesLayout || a.usesTUI) && !hasTUI {
		a.file.Imports = append(a.file.Imports, Import{
			Alias: "tui",
			Path:  "github.com/grindlemire/go-tui",
		})
	}
}

// getTUIAlias returns the import alias for github.com/grindlemire/go-tui in the current file.
func (a *Analyzer) getTUIAlias() string {
	if a.file == nil {
		return "tui"
	}
	for _, imp := range a.file.Imports {
		if imp.Path == "github.com/grindlemire/go-tui" {
			if imp.Alias != "" {
				return imp.Alias
			}
			return "tui"
		}
	}
	return "tui"
}

// validateChildrenField checks that a method templ using {children...} has a
// `children []*tui.Element` field on its receiver struct.
func (a *Analyzer) validateChildrenField(comp *Component, decls []*GoDecl) {
	structDecl := findStructDecl(decls, comp.ReceiverType)
	if structDecl == nil {
		a.errors.AddErrorf(comp.Position,
			"method templ %s uses {children...} but no struct definition found for %s",
			comp.Name, strings.TrimPrefix(comp.ReceiverType, "*"))
		return
	}

	fields := parseStructFields(structDecl.Code)
	for _, f := range fields {
		if f.Name == "children" {
			return // Found the children field
		}
	}

	typeName := strings.TrimPrefix(comp.ReceiverType, "*")
	a.errors.AddErrorf(comp.Position,
		"method templ %s uses {children...} but struct %s has no `children` field; add `children []*tui.Element` to the struct",
		comp.Name, typeName)
}

// AnalyzeFile is a convenience function that parses and analyzes a .tui file.
func AnalyzeFile(filename, source string) (*File, error) {
	lexer := NewLexer(filename, source)
	parser := NewParser(lexer)

	file, err := parser.ParseFile()
	if err != nil {
		return nil, err
	}

	analyzer := NewAnalyzer()
	if err := analyzer.Analyze(file); err != nil {
		return file, err
	}

	return file, nil
}

// ValidateElement checks if an element tag is known.
func ValidateElement(tag string) bool {
	return knownTags[tag]
}

// ValidateAttribute checks if an attribute name is known.
func ValidateAttribute(name string) bool {
	return knownAttributes[name]
}

// SuggestAttribute returns a suggestion for a misspelled attribute, or empty string.
func SuggestAttribute(name string) string {
	if similar, ok := attributeSimilar[strings.ToLower(name)]; ok {
		return similar
	}
	return ""
}
