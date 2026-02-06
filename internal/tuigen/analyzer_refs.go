package tuigen

import (
	"strings"
	"unicode"
)

// validateRefs validates element references declared via ref={} in a component.
// It checks for:
// - Valid Go identifiers for function components (needed for view struct generation)
// - Any valid Go expression for struct components (e.g., s.content)
// - Reserved export name 'Root' (capitalized ref name must not be "Root")
// - Unique names within the component (for function components only)
// - key attribute only valid inside @for loops
// - Determines ref kind from context (single, list, or map)
func (a *Analyzer) validateRefs(comp *Component) []RefInfo {
	names := make(map[string]Position)
	var refs []RefInfo

	// Struct components (with receiver) can use any Go expression for refs
	// Function components need simple identifiers for view struct generation
	isStructComponent := comp.Receiver != ""

	var check func(nodes []Node, inLoop, inConditional bool)
	check = func(nodes []Node, inLoop, inConditional bool) {
		for _, node := range nodes {
			switch n := node.(type) {
			case *Element:
				if n.RefExpr != nil {
					refName := n.RefExpr.Code

					// For function components, must be a simple identifier
					// For struct components, any Go expression is allowed
					if !isStructComponent && !isSimpleIdentifier(refName) {
						a.errors.AddErrorf(n.Position,
							"ref expression must be a simple identifier, got %q",
							refName)
					}

					// Generate export name by capitalizing first letter
					// (only meaningful for function components)
					exportName := capitalizeFirst(refName)

					// Reserved name check (only for function components)
					if !isStructComponent && exportName == "Root" {
						a.errors.AddErrorf(n.Position, "ref name %q is reserved (capitalizes to 'Root')", refName)
					}

					// Must be unique (only for function components)
					if !isStructComponent {
						if prev, exists := names[refName]; exists {
							a.errors.AddErrorf(n.Position,
								"duplicate ref name %q (first defined at %s)",
								refName, prev)
						}
						names[refName] = n.Position
					}

					// Determine ref kind from context
					kind := RefSingle
					if inLoop {
						kind = RefList
					}

					ref := RefInfo{
						Name:          refName,
						ExportName:    exportName,
						Element:       n,
						InLoop:        inLoop,
						InConditional: inConditional,
						RefKind:       kind,
						Position:      n.Position,
					}

					// Check for key attribute (for map-based refs)
					if n.RefKey != nil {
						if !inLoop {
							a.errors.AddErrorf(n.Position,
								"key attribute on ref %q only valid inside @for loop",
								refName)
						}
						ref.KeyExpr = n.RefKey.Code
						ref.KeyType = a.inferKeyType(n.RefKey.Code)
						ref.RefKind = RefMap
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

// capitalizeFirst returns the string with its first letter capitalized.
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
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

// CollectRefs collects all refs from a component.
// This is used by the generator to determine struct fields and ref bindings.
func (a *Analyzer) CollectRefs(comp *Component) []RefInfo {
	return a.validateRefs(comp)
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
