package tuigen

import (
	"fmt"
	"strconv"
)

// generateElement generates code for an element and returns the variable name.
// If parentVar is non-empty, adds this element as a child.
func (g *Generator) generateElement(elem *Element, parentVar string) string {
	return g.generateElementWithRefs(elem, parentVar, false, false)
}

// generateElementWithRefs generates code for an element with named ref handling.
// inLoop and inConditional track the context for proper variable handling.
func (g *Generator) generateElementWithRefs(elem *Element, parentVar string, inLoop bool, inConditional bool) string {
	// Determine variable name and whether we need := or =
	var varName string
	useAssignment := false // true means use "=", false means use ":="

	if elem.NamedRef != "" {
		if inLoop {
			// In loops, use temp var then append/assign to the ref slice/map
			varName = g.nextVar()
		} else {
			// Outside loops, the ref is forward-declared, use assignment
			varName = elem.NamedRef
			useAssignment = true
		}
	} else {
		varName = g.nextVar()
	}

	// Build options from attributes and tag
	elemOpts := g.buildElementOptions(elem)

	// Generate element creation - use = for forward-declared refs, := otherwise
	assignOp := ":="
	if useAssignment {
		assignOp = "="
	}

	if len(elemOpts.options) == 0 {
		g.writef("%s %s tui.New()\n", varName, assignOp)
	} else {
		g.writef("%s %s tui.New(\n", varName, assignOp)
		g.indent++
		for _, opt := range elemOpts.options {
			g.writef("%s,\n", opt)
		}
		g.indent--
		g.writeln(")")
	}

	// Defer watcher attachment until after all elements are created
	// This ensures forward-declared refs are assigned before handlers reference them
	for _, watcher := range elemOpts.watchers {
		g.deferredWatchers = append(g.deferredWatchers, deferredWatcher{
			elementVar:  varName,
			watcherExpr: watcher,
		})
	}

	// Defer handler attachment (onKeyPress, onClick) for the same reason
	for _, h := range elemOpts.handlers {
		g.deferredHandlers = append(g.deferredHandlers, deferredHandler{
			elementVar: varName,
			setter:     h.setter,
			handlerExp: h.expr,
		})
	}

	// Handle named ref assignment in loops (append to slice or assign to map)
	if elem.NamedRef != "" && inLoop {
		if elem.RefKey != nil {
			// Map-based ref: assign to map with key
			g.writef("%s[%s] = %s\n", elem.NamedRef, elem.RefKey.Code, varName)
		} else {
			// Slice-based ref: append to slice
			g.writef("%s = append(%s, %s)\n", elem.NamedRef, elem.NamedRef, varName)
		}
	}

	// Generate children - skip if text element already has content in WithText
	if !skipTextChildren(elem) {
		g.generateChildrenWithRefs(varName, elem.Children, inLoop, inConditional)
	}

	// Add to parent if specified
	if parentVar != "" {
		g.writef("%s.AddChild(%s)\n", parentVar, varName)
	}

	return varName
}

// elementOptions holds options, watchers, and deferred handlers for an element.
type elementOptions struct {
	options  []string
	watchers []string
	handlers []struct { // Handlers to defer (onKeyPress, onClick)
		setter string // e.g., "SetOnKeyPress"
		expr   string // the handler expression
	}
}

// buildElementOptions generates option expressions for an element.
// Returns both element options and any watcher expressions found.
func (g *Generator) buildElementOptions(elem *Element) elementOptions {
	var result elementOptions

	// Handle tag-specific options
	switch elem.Tag {
	case "hr":
		result.options = append(result.options, "tui.WithHR()")
	case "br":
		result.options = append(result.options, "tui.WithWidth(0)")
		result.options = append(result.options, "tui.WithHeight(1)")
	case "span", "p":
		// If text element has children that are text content, add WithText
		textContent := g.extractTextContent(elem.Children)
		if textContent != "" {
			result.options = append(result.options, fmt.Sprintf("tui.WithText(%s)", textContent))
		}
	}

	// Track text style methods from class attribute separately
	var classTextMethods []string

	// Generate options from attributes
	for _, attr := range elem.Attributes {
		// Handle class attribute specially - parse Tailwind classes
		if attr.Name == "class" {
			classValue := g.getClassAttributeValue(attr)
			if classValue != "" {
				twResult := ParseTailwindClasses(classValue)
				// Add direct options
				result.options = append(result.options, twResult.Options...)
				// Collect text style methods for combining later
				classTextMethods = append(classTextMethods, twResult.TextMethods...)
			}
			continue
		}

		// Handle watcher attributes (onChannel, onTimer) - they create watchers, not element options
		if watcherAttributes[attr.Name] {
			watcherExpr := g.generateAttributeValue(attr.Value)
			if watcherExpr != "" {
				result.watchers = append(result.watchers, watcherExpr)
			}
			continue
		}

		// Handle handler attributes - defer them so forward-declared refs are assigned first
		if setter, isHandler := handlerAttributes[attr.Name]; isHandler {
			handlerExpr := g.generateAttributeValue(attr.Value)
			if handlerExpr != "" {
				result.handlers = append(result.handlers, struct {
					setter string
					expr   string
				}{setter: setter, expr: handlerExpr})
			}
			continue
		}

		opt := g.generateAttributeOption(attr)
		if opt != "" {
			result.options = append(result.options, opt)
		}
	}

	// Build combined text style from class attribute if any
	if len(classTextMethods) > 0 {
		textStyleOpt := BuildTextStyleOption(classTextMethods)
		if textStyleOpt != "" {
			result.options = append(result.options, textStyleOpt)
		}
	}

	return result
}

// getClassAttributeValue extracts the string value from a class attribute.
func (g *Generator) getClassAttributeValue(attr *Attribute) string {
	switch v := attr.Value.(type) {
	case *StringLit:
		return v.Value
	default:
		// class attribute only supports string literals for now
		return ""
	}
}

// extractTextContent extracts text from element children for WithText.
// Returns empty string if children contain non-text content.
func (g *Generator) extractTextContent(children []Node) string {
	if len(children) == 0 {
		return ""
	}

	// If single GoExpr child, return the expression
	if len(children) == 1 {
		if expr, ok := children[0].(*GoExpr); ok {
			return expr.Code
		}
		if text, ok := children[0].(*TextContent); ok {
			return strconv.Quote(text.Text)
		}
	}

	// Multiple children or complex content - handled separately in generateChildren
	return ""
}

// handlerAttributes maps handler/callback attribute names to their setter methods.
// These attributes take function values that might capture forward-declared refs,
// so they are always deferred until after all elements are created.
// When adding a new handler attribute, add it here AND add the Set* method to element.go.
var handlerAttributes = map[string]string{
	"onKeyPress": "SetOnKeyPress",
	"onClick":    "SetOnClick",
	"onEvent":    "SetOnEvent",
	"onFocus":    "SetOnFocus",
	"onBlur":     "SetOnBlur",
}

// watcherAttributes are special attributes that create watchers, not element options.
// They are deferred and attached via AddWatcher after all elements are created.
var watcherAttributes = map[string]bool{
	"onChannel": true,
	"onTimer":   true,
}

// attributeToOption maps DSL attribute names to tui.With* functions.
// NOTE: Handler attributes (onKeyPress, onClick, etc.) are NOT in this map -
// they are in handlerAttributes and are deferred so refs are assigned first.
var attributeToOption = map[string]string{
	// Dimensions
	"width":         "tui.WithWidth(%s)",
	"widthPercent":  "tui.WithWidthPercent(%s)",
	"height":        "tui.WithHeight(%s)",
	"heightPercent": "tui.WithHeightPercent(%s)",
	"minWidth":      "tui.WithMinWidth(%s)",
	"minHeight":     "tui.WithMinHeight(%s)",
	"maxWidth":      "tui.WithMaxWidth(%s)",
	"maxHeight":     "tui.WithMaxHeight(%s)",

	// Flex container
	"direction": "tui.WithDirection(%s)",
	"justify":   "tui.WithJustify(%s)",
	"align":     "tui.WithAlign(%s)",
	"gap":       "tui.WithGap(%s)",

	// Flex item
	"flexGrow":   "tui.WithFlexGrow(%s)",
	"flexShrink": "tui.WithFlexShrink(%s)",
	"alignSelf":  "tui.WithAlignSelf(%s)",

	// Spacing
	"padding": "tui.WithPadding(%s)",
	"margin":  "tui.WithMargin(%s)",

	// Visual
	"border":      "tui.WithBorder(%s)",
	"borderStyle": "tui.WithBorderStyle(%s)",
	"background":  "tui.WithBackground(%s)",

	// Text
	"text":      "tui.WithText(%s)",
	"textStyle": "tui.WithTextStyle(%s)",
	"textAlign": "tui.WithTextAlign(%s)",

	// Focus (non-handler attributes only)
	"focusable": "tui.WithFocusable(%s)",

	// Scroll
	"scrollable": "tui.WithScrollable(%s)",
}

// generateAttributeOption generates an option expression from an attribute.
func (g *Generator) generateAttributeOption(attr *Attribute) string {
	template, ok := attributeToOption[attr.Name]
	if !ok {
		// Unknown attribute - skip with no error (analyzer should catch this)
		return ""
	}

	value := g.generateAttributeValue(attr.Value)
	return fmt.Sprintf(template, value)
}

// generateAttributeValue generates a Go expression from an attribute value.
func (g *Generator) generateAttributeValue(value Node) string {
	switch v := value.(type) {
	case *StringLit:
		return strconv.Quote(v.Value)
	case *IntLit:
		return strconv.FormatInt(v.Value, 10)
	case *FloatLit:
		return strconv.FormatFloat(v.Value, 'f', -1, 64)
	case *BoolLit:
		if v.Value {
			return "true"
		}
		return "false"
	case *GoExpr:
		return v.Code
	case *RawGoExpr:
		return v.Code
	default:
		return ""
	}
}
