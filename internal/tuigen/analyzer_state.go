package tuigen

import (
	"regexp"
	"strconv"
	"strings"
)

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
				// All elements use auto-generated variable names now (__tui_N).
				// Refs are user-declared variables, bound after element creation.
				elementName := "__tui_" + strconv.Itoa(elementIndex)
				usesCounter := true

				// Skip creating bindings for elements inside for loops - their variable
				// names are scoped to the loop and can't be referenced from the binding code
				if !inLoop {
					// Check text content in children for state usage.
					// We need to calculate the correct child element index for the binding.
					// Exception: span/p with single text/expr child puts it in WithText, not as child element.
					skipChildElements := (n.Tag == "span" || n.Tag == "p") && len(n.Children) == 1

					// Calculate starting index for child elements.
					// If parent uses counter, children start at elementIndex + 1.
					// If parent is a named ref, children start at elementIndex (counter wasn't incremented for parent).
					childElementIndex := elementIndex
					if usesCounter {
						childElementIndex++ // Parent element uses one slot
					}

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
								// Determine which element the binding should target.
								// If children become separate elements, use the child's index.
								// If the parent element has the text (single child span/p), use parent.
								bindingElementName := elementName
								if !skipChildElements {
									bindingElementName = "__tui_" + strconv.Itoa(childElementIndex)
								}

								bindings = append(bindings, StateBinding{
									StateVars:    deps,
									Element:      n,
									ElementName:  bindingElementName,
									Attribute:    "text",
									Expr:         goExpr.Code,
									ExplicitDeps: isExplicit,
								})
							}
						}

						// Increment child element index for each child that creates an element
						if !skipChildElements {
							switch child.(type) {
							case *GoExpr, *TextContent:
								childElementIndex++
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
