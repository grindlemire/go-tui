package tuigen

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// StateVar tracks information about a state variable declaration.
type StateVar struct {
	Name        string   // Variable name (e.g., "count")
	Type        string   // Go type (e.g., "int", "string", "[]Item")
	InitExpr    string   // Initialization expression (empty for parameters)
	IsParameter bool     // True if state is passed as component parameter
	Position    Position // Position of the declaration
}

// StateBinding tracks a binding between state variables and an element.
type StateBinding struct {
	StateVars   []string // State variables referenced in expression
	Element     *Element // Element that uses this expression
	ElementName string   // Generated variable name for the element (e.g., "__tui_0")
	Attribute   string   // Which attribute ("text", "class", etc.)
	Expr        string   // The expression (e.g., "fmt.Sprintf(...)")
	ExplicitDeps bool    // True if deps={...} was used
}

// NamedRef tracks information about a named element reference.
type NamedRef struct {
	Name          string
	Element       *Element
	InLoop        bool   // true = generate slice or map type
	InConditional bool   // true = may be nil at runtime
	KeyExpr       string // if set, generate map[KeyType]*element.Element
	KeyType       string // inferred type of key expression (e.g., "string", "int")
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

	// Track @let bindings for unused variable detection
	letBindings map[string]bool // name -> used

	// Track component definitions for children validation
	componentDefs map[string]bool // name -> accepts children
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
	"progress": true,
	"hr":       true,
	"br":       true,
}

// voidElements lists elements that cannot have children.
var voidElements = map[string]bool{
	"hr":    true,
	"br":    true,
	"input": true,
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
	"direction": true,
	"justify":   true,
	"align":     true,
	"gap":       true,

	// Flex item
	"flexGrow":   true,
	"flexShrink": true,
	"alignSelf":  true,

	// Spacing
	"padding": true,
	"margin":  true,

	// Visual
	"border":      true,
	"borderStyle": true,
	"background":  true,

	// Text
	"text":      true,
	"textStyle": true,
	"textAlign": true,

	// Focus
	"onFocus":   true,
	"onBlur":    true,
	"onEvent":   true,
	"focusable": true,

	// Event handlers (no bool return)
	"onKeyPress": true,
	"onClick":    true,

	// Watchers
	"onChannel": true,
	"onTimer":   true,

	// Scroll
	"scrollable":         true,
	"scrollbarStyle":     true,
	"scrollbarThumbStyle": true,

	// Generic
	"disabled": true,
	"id":       true,

	// Tailwind-style class attribute
	"class": true,

	// State reactive bindings
	"deps": true, // explicit dependencies for reactive bindings
}

// stateNewStateRegex matches tui.NewState(...) declarations.
// It captures the variable name and the initializer expression.
var stateNewStateRegex = regexp.MustCompile(`(\w+)\s*:=\s*tui\.NewState\((.+)\)`)

// stateGetRegex matches state.Get() calls to detect state usage in expressions.
// This pattern handles:
// - Simple: count.Get()
// - Dereferenced pointer: (*count).Get()
var stateGetRegex = regexp.MustCompile(`(?:\(\*(\w+)\)|(\w+))\.Get\(\)`)

// stateParamRegex matches *tui.State[T] parameter types.
// Uses greedy .+ but anchored to end of string with $, which works because
// parameter type strings don't have trailing content after the closing bracket.
var stateParamRegex = regexp.MustCompile(`\*tui\.State\[(.+)\]$`)

// attributeSimilar maps common typos to correct attribute names.
var attributeSimilar = map[string]string{
	"colour":     "color",
	"color":      "background",
	"onclick":    "onEvent",
	"onfocus":    "onFocus",
	"onblur":     "onBlur",
	"flexgrow":   "flexGrow",
	"flexshrink": "flexShrink",
	"textstyle":  "textStyle",
	"textalign":  "textAlign",
	"alignself":  "alignSelf",
	"borderstyle": "borderStyle",
}

// Analyze performs semantic analysis on a parsed file.
// Returns a list of errors/warnings found during analysis.
// Also modifies the file to add missing imports and transform element references.
func (a *Analyzer) Analyze(file *File) error {
	a.errors = NewErrorList()
	a.file = file
	a.letBindings = make(map[string]bool)
	a.componentDefs = make(map[string]bool)
	a.usesElement = false
	a.usesLayout = false
	a.usesTUI = false

	// First pass: scan components for {children...} and collect definitions
	for _, comp := range file.Components {
		comp.AcceptsChildren = a.containsChildrenSlot(comp.Body)
		a.componentDefs[comp.Name] = comp.AcceptsChildren
	}

	// Second pass: collect @let binding names from all components
	for _, comp := range file.Components {
		a.collectLetBindings(comp.Body)
	}

	// Third pass: transform GoExpr references to @let bindings into RawGoExpr
	for _, comp := range file.Components {
		comp.Body = a.transformElementRefs(comp.Body)
	}

	// Fourth pass: validate named refs
	for _, comp := range file.Components {
		a.validateNamedRefs(comp)
	}

	// Fifth pass: validate elements and attributes
	for _, comp := range file.Components {
		a.analyzeComponent(comp)
	}

	// Check for unused @let bindings
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

// validateNamedRefs validates named element references in a component.
// It checks for:
// - Valid Go identifiers (PascalCase required)
// - Reserved name 'Root'
// - Unique names within the component
// - key attribute only valid inside @for loops
func (a *Analyzer) validateNamedRefs(comp *Component) []NamedRef {
	names := make(map[string]Position)
	var refs []NamedRef

	var check func(nodes []Node, inLoop, inConditional bool)
	check = func(nodes []Node, inLoop, inConditional bool) {
		for _, node := range nodes {
			switch n := node.(type) {
			case *Element:
				if n.NamedRef != "" {
					// Must be valid Go identifier starting with uppercase
					if !isValidRefName(n.NamedRef) {
						a.errors.AddErrorf(n.Position,
							"invalid ref name %q - must be valid Go identifier starting with uppercase letter",
							n.NamedRef)
					}
					// Reserved name check
					if n.NamedRef == "Root" {
						a.errors.AddErrorf(n.Position, "ref name 'Root' is reserved")
					}
					// Must be unique
					if prev, exists := names[n.NamedRef]; exists {
						a.errors.AddErrorf(n.Position,
							"duplicate ref name %q (first defined at %s)",
							n.NamedRef, prev)
					}
					names[n.NamedRef] = n.Position

					ref := NamedRef{
						Name:          n.NamedRef,
						Element:       n,
						InLoop:        inLoop,
						InConditional: inConditional,
						Position:      n.Position,
					}

					// Check for key attribute (for map-based refs)
					if n.RefKey != nil {
						if !inLoop {
							a.errors.AddErrorf(n.Position,
								"key attribute on ref %q only valid inside @for loop",
								n.NamedRef)
						}
						ref.KeyExpr = n.RefKey.Code
						ref.KeyType = a.inferKeyType(n.RefKey.Code)
					}

					refs = append(refs, ref)
				}
				check(n.Children, inLoop, inConditional)

			case *LetBinding:
				check(n.Element.Children, inLoop, inConditional)

			case *ForLoop:
				// Refs inside loops get slice type
				check(n.Body, true, inConditional)

			case *IfStmt:
				// Refs inside conditionals may be nil
				check(n.Then, inLoop, true)
				check(n.Else, inLoop, true)

			case *ComponentCall:
				check(n.Children, inLoop, inConditional)
			}
		}
	}

	check(comp.Body, false, false)
	return refs
}

// isValidRefName checks if a name is a valid Go identifier starting with uppercase.
func isValidRefName(name string) bool {
	if len(name) == 0 {
		return false
	}
	// First character must be uppercase letter
	first := rune(name[0])
	if !unicode.IsUpper(first) {
		return false
	}
	// Rest must be letters, digits, or underscores
	for _, ch := range name[1:] {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			return false
		}
	}
	return true
}

// inferKeyType attempts to infer the type of a key expression.
// For now, this returns a simple heuristic based on the expression.
func (a *Analyzer) inferKeyType(expr string) string {
	// Simple heuristics for common patterns
	// In a real implementation, we'd need type information from the Go type checker
	if strings.HasSuffix(expr, ".ID") || strings.HasSuffix(expr, ".Id") {
		return "string" // Common pattern: item.ID
	}
	if strings.Contains(expr, "int") || strings.Contains(expr, "Int") {
		return "int"
	}
	// Default to string which is the most flexible
	return "string"
}

// CollectNamedRefs collects all named refs from a component.
// This is used by the generator to determine struct fields.
func (a *Analyzer) CollectNamedRefs(comp *Component) []NamedRef {
	return a.validateNamedRefs(comp)
}

// collectLetBindings traverses nodes to collect all @let binding names.
func (a *Analyzer) collectLetBindings(nodes []Node) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *LetBinding:
			a.letBindings[n.Name] = false
			a.collectLetBindings(n.Element.Children)
		case *Element:
			a.collectLetBindings(n.Children)
		case *ForLoop:
			a.collectLetBindings(n.Body)
		case *IfStmt:
			a.collectLetBindings(n.Then)
			a.collectLetBindings(n.Else)
		case *ComponentCall:
			a.collectLetBindings(n.Children)
		}
	}
}

// containsChildrenSlot recursively checks if nodes contain a {children...} slot.
func (a *Analyzer) containsChildrenSlot(nodes []Node) bool {
	for _, node := range nodes {
		switch n := node.(type) {
		case *ChildrenSlot:
			return true
		case *Element:
			if a.containsChildrenSlot(n.Children) {
				return true
			}
		case *LetBinding:
			if a.containsChildrenSlot(n.Element.Children) {
				return true
			}
		case *ForLoop:
			if a.containsChildrenSlot(n.Body) {
				return true
			}
		case *IfStmt:
			if a.containsChildrenSlot(n.Then) || a.containsChildrenSlot(n.Else) {
				return true
			}
		case *ComponentCall:
			if a.containsChildrenSlot(n.Children) {
				return true
			}
		}
	}
	return false
}

// transformElementRefs transforms GoExpr nodes that reference @let bindings into RawGoExpr.
func (a *Analyzer) transformElementRefs(nodes []Node) []Node {
	result := make([]Node, len(nodes))
	for i, node := range nodes {
		result[i] = a.transformNode(node)
	}
	return result
}

// transformNode transforms a single node, recursively processing children.
func (a *Analyzer) transformNode(node Node) Node {
	switch n := node.(type) {
	case *GoExpr:
		// Check if this is a simple identifier that matches a @let binding
		if isSimpleIdentifier(n.Code) {
			if _, ok := a.letBindings[n.Code]; ok {
				a.letBindings[n.Code] = true
				return &RawGoExpr{Code: n.Code, Position: n.Position}
			}
		}
		return n
	case *Element:
		n.Children = a.transformElementRefs(n.Children)
		return n
	case *LetBinding:
		n.Element.Children = a.transformElementRefs(n.Element.Children)
		return n
	case *ForLoop:
		n.Body = a.transformElementRefs(n.Body)
		return n
	case *IfStmt:
		n.Then = a.transformElementRefs(n.Then)
		n.Else = a.transformElementRefs(n.Else)
		return n
	case *ComponentCall:
		n.Children = a.transformElementRefs(n.Children)
		return n
	default:
		return node
	}
}

// isSimpleIdentifier returns true if the string is a valid Go identifier (no dots, parens, etc.)
func isSimpleIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	// First character must be letter or underscore
	ch := rune(s[0])
	if !isIdentLetter(ch) {
		return false
	}
	// Rest must be letters, digits, or underscores
	for _, ch := range s[1:] {
		if !isIdentLetter(ch) && !isIdentDigit(ch) {
			return false
		}
	}
	return true
}

// isIdentLetter checks if a rune is a letter or underscore (for identifier checking).
func isIdentLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// isIdentDigit checks if a rune is a digit (for identifier checking).
func isIdentDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

// Errors returns the errors found during analysis.
func (a *Analyzer) Errors() *ErrorList {
	return a.errors
}

// analyzeComponent validates a single component.
func (a *Analyzer) analyzeComponent(comp *Component) {
	// Track that we use elements
	a.usesElement = true

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
	case *ChildrenSlot:
		// ChildrenSlot is valid - no additional validation needed
	}
}

// analyzeElement validates an element and its children.
func (a *Analyzer) analyzeElement(elem *Element) {
	// Check if tag is known
	if !knownTags[elem.Tag] {
		a.errors.AddErrorf(elem.Position, "unknown element tag <%s>", elem.Tag)
	}

	// Check for children on void elements
	if voidElements[elem.Tag] && len(elem.Children) > 0 {
		a.errors.AddErrorf(elem.Position,
			"<%s> is a void element and cannot have children", elem.Tag)
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

	// Check if class attribute uses Tailwind classes that need imports
	if attr.Name == "class" {
		if v, ok := attr.Value.(*StringLit); ok {
			result := ParseTailwindClasses(v.Value)
			if result.NeedsImports["layout"] {
				a.usesLayout = true
			}
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

	// Analyze the element
	a.analyzeElement(let.Element)
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

// analyzeGoExpr validates a Go expression.
func (a *Analyzer) analyzeGoExpr(expr *GoExpr) {
	// Check if expression references layout or tui packages
	if strings.Contains(expr.Code, "layout.") {
		a.usesLayout = true
	}
	if strings.Contains(expr.Code, "tui.") {
		a.usesTUI = true
	}

	// Check if expression references a @let binding
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

	// Check if code references a @let binding
	for name := range a.letBindings {
		if strings.Contains(code.Code, name) {
			a.letBindings[name] = true
		}
	}
}

// addMissingImports adds required imports that are missing from the file.
func (a *Analyzer) addMissingImports() {
	// Check which imports are already present
	hasElement := false
	hasLayout := false
	hasTUI := false

	for _, imp := range a.file.Imports {
		switch imp.Path {
		case "github.com/grindlemire/go-tui/pkg/tui/element":
			hasElement = true
		case "github.com/grindlemire/go-tui/pkg/layout":
			hasLayout = true
		case "github.com/grindlemire/go-tui/pkg/tui":
			hasTUI = true
		}
	}

	// Add missing imports
	if a.usesElement && !hasElement {
		a.file.Imports = append(a.file.Imports, Import{
			Path: "github.com/grindlemire/go-tui/pkg/tui/element",
		})
	}

	if a.usesLayout && !hasLayout {
		a.file.Imports = append(a.file.Imports, Import{
			Path: "github.com/grindlemire/go-tui/pkg/layout",
		})
	}

	if a.usesTUI && !hasTUI {
		a.file.Imports = append(a.file.Imports, Import{
			Path: "github.com/grindlemire/go-tui/pkg/tui",
		})
	}
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

// DetectStateVars scans a component for tui.NewState declarations and state parameters.
// It returns a list of all detected state variables.
func (a *Analyzer) DetectStateVars(comp *Component) []StateVar {
	var stateVars []StateVar

	// First, detect state parameters in component signature
	for _, param := range comp.Params {
		matches := stateParamRegex.FindStringSubmatch(param.Type)
		if matches != nil {
			stateVars = append(stateVars, StateVar{
				Name:        param.Name,
				Type:        matches[1], // Type inside State[T]
				IsParameter: true,
				Position:    param.Position,
			})
		}
	}

	// Then, scan component body for tui.NewState declarations
	// We need to look for GoCode nodes that contain state declarations
	for _, node := range comp.Body {
		if goCode, ok := node.(*GoCode); ok {
			stateVars = append(stateVars, a.parseStateDeclarations(goCode)...)
		}
	}

	return stateVars
}

// parseStateDeclarations extracts tui.NewState declarations from Go code.
func (a *Analyzer) parseStateDeclarations(code *GoCode) []StateVar {
	var stateVars []StateVar

	// Find all matches in the code
	matches := stateNewStateRegex.FindAllStringSubmatch(code.Code, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			varName := match[1]
			initExpr := match[2]
			stateType := inferTypeFromExpr(initExpr)

			stateVars = append(stateVars, StateVar{
				Name:     varName,
				Type:     stateType,
				InitExpr: initExpr,
				Position: code.Position,
			})
		}
	}

	return stateVars
}

// inferTypeFromExpr attempts to infer the Go type from an expression.
// This uses heuristics for common patterns.
func inferTypeFromExpr(expr string) string {
	expr = strings.TrimSpace(expr)

	// Integer literal (positive or negative)
	if regexp.MustCompile(`^-?\d+$`).MatchString(expr) {
		return "int"
	}

	// Float literal (positive or negative)
	if regexp.MustCompile(`^-?\d+\.\d+$`).MatchString(expr) {
		return "float64"
	}

	// Boolean literal
	if expr == "true" || expr == "false" {
		return "bool"
	}

	// String literal
	if (strings.HasPrefix(expr, `"`) && strings.HasSuffix(expr, `"`)) ||
		(strings.HasPrefix(expr, "`") && strings.HasSuffix(expr, "`")) {
		return "string"
	}

	// Nil pointer
	if expr == "nil" {
		return "any" // Can't infer specific type from nil
	}

	// Slice literal []Type{...}
	if sliceMatch := regexp.MustCompile(`^\[\](\w+(?:\.\w+)?)\{`).FindStringSubmatch(expr); sliceMatch != nil {
		return "[]" + sliceMatch[1]
	}

	// Map literal map[K]V{...}
	if mapMatch := regexp.MustCompile(`^map\[(\w+)\](\w+(?:\.\w+)?)\{`).FindStringSubmatch(expr); mapMatch != nil {
		return "map[" + mapMatch[1] + "]" + mapMatch[2]
	}

	// Pointer to struct &Type{...}
	if ptrMatch := regexp.MustCompile(`^&(\w+(?:\.\w+)?)\{`).FindStringSubmatch(expr); ptrMatch != nil {
		return "*" + ptrMatch[1]
	}

	// Struct literal Type{...}
	if structMatch := regexp.MustCompile(`^(\w+(?:\.\w+)?)\{`).FindStringSubmatch(expr); structMatch != nil {
		return structMatch[1]
	}

	// Default to any if we can't infer
	return "any"
}

// DetectStateBindings scans elements for state usage and returns bindings.
// This detects both explicit deps={...} attributes and auto-detected .Get() calls.
//
// The elementIndex counter assigns names like "__tui_0", "__tui_1", etc. to unnamed
// elements. This must match the generator's naming scheme in generator.go. Named
// refs (#Name) use their ref name instead.
//
// Note: Elements inside @for loops are skipped for binding generation because their
// generated variable names are scoped to the loop and cannot be referenced from outside.
func (a *Analyzer) DetectStateBindings(comp *Component, stateVars []StateVar) []StateBinding {
	// Build a set of state variable names for quick lookup
	stateNames := make(map[string]bool)
	for _, sv := range stateVars {
		stateNames[sv.Name] = true
	}

	var bindings []StateBinding
	elementIndex := 0

	var scan func(nodes []Node, inLoop bool)
	scan = func(nodes []Node, inLoop bool) {
		for _, node := range nodes {
			switch n := node.(type) {
			case *Element:
				// Check for explicit deps attribute first
				var explicitDeps []string
				for _, attr := range n.Attributes {
					if attr.Name == "deps" {
						explicitDeps = a.parseExplicitDeps(attr, stateNames)
						break
					}
				}

				// Generate element name (same pattern as generator)
				// Named refs outside loops use the ref name directly without incrementing the counter.
				// Named refs inside loops or conditionals still use the counter (generator creates temp var).
				elementName := "__tui_" + strconv.Itoa(elementIndex)
				usesCounter := true
				if n.NamedRef != "" && !inLoop {
					elementName = n.NamedRef
					usesCounter = false
				}

				// Skip creating bindings for elements inside for loops - their variable
				// names are scoped to the loop and can't be referenced from the binding code
				if !inLoop {
					// Check text content in children for state usage
					for _, child := range n.Children {
						if goExpr, ok := child.(*GoExpr); ok {
							var deps []string
							isExplicit := false

							if explicitDeps != nil {
								deps = explicitDeps
								isExplicit = true
							} else {
								deps = a.detectGetCalls(goExpr.Code, stateNames)
							}

							if len(deps) > 0 {
								bindings = append(bindings, StateBinding{
									StateVars:    deps,
									Element:      n,
									ElementName:  elementName,
									Attribute:    "text",
									Expr:         goExpr.Code,
									ExplicitDeps: isExplicit,
								})
							}
						}
					}

					// Check for dynamic class attribute
					for _, attr := range n.Attributes {
						if attr.Name == "class" {
							if goExpr, ok := attr.Value.(*GoExpr); ok {
								var deps []string
								isExplicit := false

								if explicitDeps != nil {
									deps = explicitDeps
									isExplicit = true
								} else {
									deps = a.detectGetCalls(goExpr.Code, stateNames)
								}

								if len(deps) > 0 {
									bindings = append(bindings, StateBinding{
										StateVars:    deps,
										Element:      n,
										ElementName:  elementName,
										Attribute:    "class",
										Expr:         goExpr.Code,
										ExplicitDeps: isExplicit,
									})
								}
							}
						}
					}
				}

				// Only increment counter if the element uses it (matches generator behavior)
				if usesCounter {
					elementIndex++
				}

				// Count GoExpr and TextContent children that become separate text elements.
				// This matches the generator's behavior where it creates new elements for these.
				// Exception: span/p elements with a single text child put it in WithText, not as child element.
				skipChildren := (n.Tag == "span" || n.Tag == "p") && len(n.Children) == 1
				if !skipChildren {
					for _, child := range n.Children {
						switch child.(type) {
						case *GoExpr, *TextContent:
							elementIndex++
						}
					}
				}

				scan(n.Children, inLoop)

			case *LetBinding:
				// LetBindings wrap elements; recursively scan the wrapped element's children.
				// The wrapped element itself is handled when we encounter the Element node.
				scan(n.Element.Children, inLoop)

			case *ForLoop:
				// Elements inside for loops have loop-scoped variable names
				scan(n.Body, true)

			case *IfStmt:
				scan(n.Then, inLoop)
				scan(n.Else, inLoop)

			case *ComponentCall:
				scan(n.Children, inLoop)
			}
		}
	}

	scan(comp.Body, false)
	return bindings
}

// parseExplicitDeps extracts state variable names from a deps={[state1, state2]} attribute.
// It validates that each name exists in the known state variables.
// Returns nil if the attribute is not a valid deps specification.
func (a *Analyzer) parseExplicitDeps(attr *Attribute, stateNames map[string]bool) []string {
	goExpr, ok := attr.Value.(*GoExpr)
	if !ok {
		// deps attribute must use Go expression syntax: deps={[state1, state2]}
		// String literals like deps="..." are not valid
		a.errors.AddErrorf(attr.Position, "deps attribute must use expression syntax: deps={[state1, state2]}")
		return nil
	}

	code := strings.TrimSpace(goExpr.Code)

	// Parse [state1, state2] format - must have brackets
	if !strings.HasPrefix(code, "[") || !strings.HasSuffix(code, "]") {
		a.errors.AddErrorf(attr.Position, "deps attribute must be an array literal: deps={[state1, state2]}")
		return nil
	}

	// Extract the contents between [ and ]
	inner := strings.TrimSpace(code[1 : len(code)-1])
	if inner == "" {
		// Empty deps like deps={[]} - warn but don't treat as error
		a.errors.AddErrorf(attr.Position, "empty deps attribute has no effect")
		return nil
	}

	// Split by comma and validate each name
	var deps []string
	for _, part := range strings.Split(inner, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		if !stateNames[name] {
			a.errors.AddErrorf(attr.Position, "unknown state variable %q in deps", name)
			continue
		}
		deps = append(deps, name)
	}

	return deps
}

// detectGetCalls finds all state.Get() calls in an expression and returns the state variable names.
// It handles both simple calls (count.Get()) and dereferenced pointers ((*count).Get()).
func (a *Analyzer) detectGetCalls(expr string, stateNames map[string]bool) []string {
	matches := stateGetRegex.FindAllStringSubmatch(expr, -1)

	// Use a map to deduplicate
	seen := make(map[string]bool)
	var deps []string

	for _, match := range matches {
		// The regex has two capture groups:
		// - match[1] captures from (*name) pattern
		// - match[2] captures from simple name pattern
		// Only one will be non-empty for each match
		var name string
		if match[1] != "" {
			name = match[1] // (*name).Get()
		} else if len(match) > 2 && match[2] != "" {
			name = match[2] // name.Get()
		}

		if name != "" && stateNames[name] && !seen[name] {
			seen[name] = true
			deps = append(deps, name)
		}
	}

	return deps
}
