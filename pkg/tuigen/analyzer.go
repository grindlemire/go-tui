package tuigen

import (
	"strings"
)

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
	"onFocus": true,
	"onBlur":  true,
	"onEvent": true,

	// Scroll
	"scrollable":         true,
	"scrollbarStyle":     true,
	"scrollbarThumbStyle": true,

	// Generic
	"disabled": true,
	"id":       true,

	// Tailwind-style class attribute
	"class": true,
}

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

	// Fourth pass: validate elements and attributes
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
