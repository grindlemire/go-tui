package lsp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/grindlemire/go-tui/pkg/lsp/gopls"
	"github.com/grindlemire/go-tui/pkg/lsp/log"
	"github.com/grindlemire/go-tui/pkg/tuigen"
)

// HoverParams represents textDocument/hover parameters.
type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// Hover represents the result of a hover request.
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// MarkupContent represents markup content for hover.
type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// handleHover handles textDocument/hover requests.
func (s *Server) handleHover(params json.RawMessage) (any, *Error) {
	var p HoverParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &Error{Code: CodeInvalidParams, Message: err.Error()}
	}

	log.Server("Hover request at %s:%d:%d", p.TextDocument.URI, p.Position.Line, p.Position.Character)

	doc := s.docs.Get(p.TextDocument.URI)
	if doc == nil {
		return nil, nil
	}

	// Check if we're inside a Go expression - use gopls if available
	if s.isInGoExpression(doc, p.Position) {
		hover, err := s.getGoplsHover(doc, p.Position)
		if err != nil {
			log.Server("gopls hover error: %v", err)
			// Fall through to TUI hover
		} else if hover != nil {
			return hover, nil
		}
	}

	// Find what's at the cursor position
	word := s.getWordAtPosition(doc, p.Position)
	if word == "" {
		return nil, nil
	}

	log.Server("Word at hover position: %s", word)

	// Check for TUI keyword hover
	if hover := hoverForKeyword(word); hover != nil {
		return hover, nil
	}

	// Check for component hover
	componentName := word
	if strings.HasPrefix(word, "@") {
		componentName = strings.TrimPrefix(word, "@")
	}

	// Look up component in index
	info, ok := s.index.Lookup(componentName)
	if ok {
		return s.hoverForComponent(info), nil
	}

	// Check if this is a helper function
	if funcInfo, ok := s.index.LookupFunc(word); ok {
		return hoverForFunc(funcInfo), nil
	}

	// Check if this is a component parameter (find which component we're in)
	if componentCtx := s.findComponentAtPosition(doc, p.Position); componentCtx != "" {
		if paramInfo, ok := s.index.LookupParam(componentCtx, word); ok {
			return hoverForParam(paramInfo), nil
		}
	}

	// Check if this is an element tag
	if hover := hoverForElement(word); hover != nil {
		return hover, nil
	}

	// Check for Tailwind class hover
	if hover := s.hoverForClass(doc, p.Position); hover != nil {
		return hover, nil
	}

	// Check if this is an attribute
	attrInfo := s.getAttributeAtPosition(doc, p.Position)
	if attrInfo != nil {
		return s.hoverForAttribute(attrInfo.tag, attrInfo.name), nil
	}

	return nil, nil
}

// hoverForComponent creates hover content for a component.
func (s *Server) hoverForComponent(info *ComponentInfo) *Hover {
	// Build signature
	var params []string
	for _, p := range info.Params {
		params = append(params, fmt.Sprintf("%s %s", p.Name, p.Type))
	}
	sig := fmt.Sprintf("func %s(%s) *element.Element", info.Name, strings.Join(params, ", "))

	markdown := fmt.Sprintf("```go\n%s\n```\n\n**TUI Component**", sig)

	return &Hover{
		Contents: MarkupContent{
			Kind:  "markdown",
			Value: markdown,
		},
	}
}

// hoverForFunc creates hover content for a helper function.
func hoverForFunc(info *FuncInfo) *Hover {
	markdown := fmt.Sprintf("```go\n%s\n```\n\n**Helper Function**", info.Signature)

	return &Hover{
		Contents: MarkupContent{
			Kind:  "markdown",
			Value: markdown,
		},
	}
}

// hoverForParam creates hover content for a component parameter.
func hoverForParam(info *ParamInfo) *Hover {
	markdown := fmt.Sprintf("```go\n%s %s\n```\n\n**Parameter** of component `%s`",
		info.Name, info.Type, info.ComponentName)

	return &Hover{
		Contents: MarkupContent{
			Kind:  "markdown",
			Value: markdown,
		},
	}
}

// findComponentAtPosition finds which component the given position is inside.
func (s *Server) findComponentAtPosition(doc *Document, pos Position) string {
	if doc.AST == nil {
		return ""
	}

	// Convert to 1-indexed for comparison with AST positions
	line := pos.Line + 1

	// Find the last component that starts at or before this line.
	// We iterate backwards so the first match is the correct one.
	for i := len(doc.AST.Components) - 1; i >= 0; i-- {
		comp := doc.AST.Components[i]
		if line >= comp.Position.Line {
			return comp.Name
		}
	}

	return ""
}

// hoverForKeyword returns hover documentation for TUI-specific keywords.
func hoverForKeyword(word string) *Hover {
	var doc string

	switch word {
	case "@component", "component":
		doc = `## @component

Defines a reusable TUI component that compiles to a Go function.

**Syntax:**
` + "```tui" + `
@component Name(param1 Type1, param2 Type2) {
    <div>...</div>
}
` + "```" + `

**Example:**
` + "```tui" + `
@component Button(label string, disabled bool) {
    <button class="p-1 border-rounded" disabled={disabled}>
        {label}
    </button>
}
` + "```" + `

Components are compiled to functions with signature:
` + "```go" + `
func Name(params...) *element.Element
` + "```"

	case "@for", "for":
		doc = `## @for

Iterates over a collection, rendering elements for each item.

**Syntax:**
` + "```tui" + `
@for index, item := range collection {
    <element>...</element>
}
` + "```" + `

**Example:**
` + "```tui" + `
@for i, name := range names {
    <li>{fmt.Sprintf("%d. %s", i+1, name)}</li>
}
` + "```" + `

- Supports standard Go range syntax
- Variables are scoped to the loop body
- Can iterate over slices, arrays, maps, and channels`

	case "@if", "if":
		doc = `## @if

Conditionally renders elements based on a boolean expression.

**Syntax:**
` + "```tui" + `
@if condition {
    <element>...</element>
} @else {
    <element>...</element>
}
` + "```" + `

**Example:**
` + "```tui" + `
@if user.IsAdmin {
    <span class="text-green">Admin</span>
} @else {
    <span class="text-dim">User</span>
}
` + "```" + `

- Condition must be a Go boolean expression
- @else clause is optional`

	case "@else", "else":
		doc = `## @else

The else branch of a conditional @if statement.

**Syntax:**
` + "```tui" + `
@if condition {
    <element>...</element>
} @else {
    <element>...</element>
}
` + "```" + `

- Must follow an @if block
- Contains elements to render when condition is false`

	case "@let", "let":
		doc = `## @let

Creates a local binding within a component body.

**Syntax:**
` + "```tui" + `
@let varName = expression
` + "```" + `

**Examples:**
` + "```tui" + `
@let label = fmt.Sprintf("Count: %d", count)
<span>{label}</span>

@let header = <div class="font-bold">{title}</div>
{header}
` + "```" + `

- Can bind Go expressions or element trees
- Variables are scoped to the component body
- Useful for computed values or reusable sub-elements`

	case "package":
		doc = `## package

Declares the Go package for the generated code.

**Syntax:**
` + "```tui" + `
package mypackage
` + "```" + `

- Must be the first declaration in the file
- Generated Go code will have this package name
- Should match the directory name (Go convention)`

	case "import":
		doc = `## import

Imports Go packages for use in expressions.

**Syntax:**
` + "```tui" + `
import (
    "fmt"
    "strings"

    "mymodule/pkg/utils"
)
` + "```" + `

- Standard Go import syntax
- Imported packages can be used in {} expressions
- The tui package is automatically available`

	default:
		return nil
	}

	return &Hover{
		Contents: MarkupContent{
			Kind:  "markdown",
			Value: doc,
		},
	}
}

// hoverForElement creates hover content for an element tag.
func hoverForElement(tag string) *Hover {
	info := getElementInfo(tag)
	if info == nil {
		return nil
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("## `<%s>`", tag))
	lines = append(lines, "")
	lines = append(lines, info.Description)
	lines = append(lines, "")
	lines = append(lines, "**Available attributes:**")
	for _, attr := range info.Attributes {
		lines = append(lines, fmt.Sprintf("- `%s` (%s): %s", attr.Name, attr.Type, attr.Description))
	}

	return &Hover{
		Contents: MarkupContent{
			Kind:  "markdown",
			Value: strings.Join(lines, "\n"),
		},
	}
}

// ElementInfo describes a built-in element.
type ElementInfo struct {
	Description string
	Attributes  []AttributeDoc
}

// getElementInfo returns documentation for an element tag.
func getElementInfo(tag string) *ElementInfo {
	switch tag {
	case "div":
		return &ElementInfo{
			Description: "A block container with flexbox layout. The primary building block for layouts.",
			Attributes:  commonLayoutAttrs(),
		}
	case "span":
		return &ElementInfo{
			Description: "An inline text container for styling text content.",
			Attributes: append([]AttributeDoc{
				{Name: "id", Type: "string", Description: "Unique identifier for the element"},
				{Name: "class", Type: "string", Description: "Tailwind-style CSS classes"},
			}, commonTextAttrs()...),
		}
	case "p":
		return &ElementInfo{
			Description: "A paragraph element for text blocks.",
			Attributes: append([]AttributeDoc{
				{Name: "id", Type: "string", Description: "Unique identifier for the element"},
				{Name: "class", Type: "string", Description: "Tailwind-style CSS classes"},
			}, commonTextAttrs()...),
		}
	case "ul":
		return &ElementInfo{
			Description: "An unordered list container. Use with `<li>` children.",
			Attributes:  commonLayoutAttrs(),
		}
	case "li":
		return &ElementInfo{
			Description: "A list item. Should be a child of `<ul>`.",
			Attributes:  commonLayoutAttrs(),
		}
	case "button":
		return &ElementInfo{
			Description: "A clickable button element that can receive focus and handle events.",
			Attributes: []AttributeDoc{
				{Name: "id", Type: "string", Description: "Unique identifier for the element"},
				{Name: "class", Type: "string", Description: "Tailwind-style CSS classes"},
				{Name: "disabled", Type: "bool", Description: "Whether button is disabled"},
				{Name: "onEvent", Type: "func", Description: "Event handler function"},
			},
		}
	case "input":
		return &ElementInfo{
			Description: "A text input field for user input.",
			Attributes: []AttributeDoc{
				{Name: "id", Type: "string", Description: "Unique identifier for the element"},
				{Name: "class", Type: "string", Description: "Tailwind-style CSS classes"},
				{Name: "value", Type: "string", Description: "Current input value"},
				{Name: "placeholder", Type: "string", Description: "Placeholder text when empty"},
				{Name: "width", Type: "int", Description: "Input width in characters"},
				{Name: "disabled", Type: "bool", Description: "Whether input is disabled"},
			},
		}
	case "table":
		return &ElementInfo{
			Description: "A table container for tabular data.",
			Attributes:  commonLayoutAttrs(),
		}
	case "progress":
		return &ElementInfo{
			Description: "A progress bar element showing completion status.",
			Attributes: []AttributeDoc{
				{Name: "id", Type: "string", Description: "Unique identifier for the element"},
				{Name: "class", Type: "string", Description: "Tailwind-style CSS classes"},
				{Name: "value", Type: "int", Description: "Current progress value (0 to max)"},
				{Name: "max", Type: "int", Description: "Maximum progress value"},
				{Name: "width", Type: "int", Description: "Progress bar width in characters"},
			},
		}
	default:
		return nil
	}
}

// AttributeDoc describes an attribute for documentation.
type AttributeDoc struct {
	Name        string
	Type        string
	Description string
}

// commonLayoutAttrs returns layout attributes common to container elements.
func commonLayoutAttrs() []AttributeDoc {
	return []AttributeDoc{
		{Name: "id", Type: "string", Description: "Unique identifier for the element"},
		{Name: "class", Type: "string", Description: "Tailwind-style CSS classes"},
		{Name: "padding", Type: "int", Description: "Padding on all sides"},
		{Name: "margin", Type: "int", Description: "Margin on all sides"},
		{Name: "width", Type: "int", Description: "Fixed width"},
		{Name: "widthPercent", Type: "int", Description: "Width as percentage"},
		{Name: "height", Type: "int", Description: "Fixed height"},
		{Name: "heightPercent", Type: "int", Description: "Height as percentage"},
		{Name: "minWidth", Type: "int", Description: "Minimum width"},
		{Name: "minHeight", Type: "int", Description: "Minimum height"},
		{Name: "maxWidth", Type: "int", Description: "Maximum width"},
		{Name: "maxHeight", Type: "int", Description: "Maximum height"},
		{Name: "flexGrow", Type: "float", Description: "Flex grow factor"},
		{Name: "flexShrink", Type: "float", Description: "Flex shrink factor"},
		{Name: "direction", Type: "direction", Description: "Flex direction (row, column)"},
		{Name: "justify", Type: "justify", Description: "Justify content (start, center, end, between, around)"},
		{Name: "align", Type: "align", Description: "Align items (start, center, end, stretch)"},
		{Name: "gap", Type: "int", Description: "Gap between children"},
		{Name: "border", Type: "border", Description: "Border style (none, single, double, rounded, thick)"},
		{Name: "background", Type: "color", Description: "Background color"},
	}
}

// commonTextAttrs returns text styling attributes.
func commonTextAttrs() []AttributeDoc {
	return []AttributeDoc{
		{Name: "text", Type: "string", Description: "Text content"},
		{Name: "textStyle", Type: "style", Description: "Text styling"},
		{Name: "textAlign", Type: "string", Description: "Text alignment"},
	}
}

// getElementAttributes returns documentation for element attributes (for attribute hover).
func getElementAttributes(tag string) []AttributeDoc {
	info := getElementInfo(tag)
	if info == nil {
		return nil
	}
	return info.Attributes
}

// isElementTag returns true if the word is a known element tag.
func isElementTag(word string) bool {
	return getElementInfo(word) != nil
}

// hoverForClass returns hover documentation for Tailwind-style classes.
func (s *Server) hoverForClass(doc *Document, pos Position) *Hover {
	// Find if we're inside a class attribute string
	offset := PositionToOffset(doc.Content, pos)

	// Search backwards for class="
	searchStart := offset - 100
	if searchStart < 0 {
		searchStart = 0
	}

	content := doc.Content[searchStart:offset]
	classIdx := strings.LastIndex(content, `class="`)
	if classIdx == -1 {
		return nil
	}

	// Check we haven't passed the closing quote
	afterClass := content[classIdx+7:]
	if strings.Contains(afterClass, `"`) {
		return nil
	}

	// Find the class name at cursor
	classStart := searchStart + classIdx + 7
	classContent := doc.Content[classStart:offset]

	// Find word boundaries within the class string
	lastSpace := strings.LastIndex(classContent, " ")
	var className string
	if lastSpace == -1 {
		className = classContent
	} else {
		className = classContent[lastSpace+1:]
	}

	// Extend forward to get full class name
	endOffset := offset
	for endOffset < len(doc.Content) && doc.Content[endOffset] != ' ' && doc.Content[endOffset] != '"' {
		endOffset++
	}
	if endOffset > offset {
		className += doc.Content[offset:endOffset]
	}

	className = strings.TrimSpace(className)
	if className == "" {
		return nil
	}

	return hoverForTailwindClass(className)
}

// hoverForTailwindClass returns documentation for a Tailwind-style class.
func hoverForTailwindClass(class string) *Hover {
	classDoc := getTailwindClassDoc(class)
	if classDoc == "" {
		return nil
	}

	return &Hover{
		Contents: MarkupContent{
			Kind:  "markdown",
			Value: fmt.Sprintf("**`%s`**\n\n%s", class, classDoc),
		},
	}
}

// getTailwindClassDoc returns documentation for a Tailwind-style class.
func getTailwindClassDoc(class string) string {
	// Layout classes
	switch class {
	case "flex":
		return "Display as flexbox with row direction."
	case "flex-col":
		return "Display as flexbox with column direction."
	case "flex-row":
		return "Display as flexbox with row direction (default)."
	}

	// Alignment
	switch class {
	case "items-start":
		return "Align items to the start of the cross axis."
	case "items-center":
		return "Align items to the center of the cross axis."
	case "items-end":
		return "Align items to the end of the cross axis."
	case "items-stretch":
		return "Stretch items to fill the cross axis."
	case "justify-start":
		return "Justify content to the start of the main axis."
	case "justify-center":
		return "Justify content to the center of the main axis."
	case "justify-end":
		return "Justify content to the end of the main axis."
	case "justify-between":
		return "Distribute items with space between them."
	case "justify-around":
		return "Distribute items with space around them."
	}

	// Border styles
	switch class {
	case "border-none":
		return "No border."
	case "border-single":
		return "Single line border: `┌─┐│└─┘`"
	case "border-double":
		return "Double line border: `╔═╗║╚═╝`"
	case "border-rounded":
		return "Rounded border: `╭─╮│╰─╯`"
	case "border-thick":
		return "Thick border: `┏━┓┃┗━┛`"
	}

	// Font styles
	switch class {
	case "font-bold":
		return "Bold text style."
	case "font-dim":
		return "Dim/faint text style."
	case "font-italic":
		return "Italic text style."
	case "font-underline":
		return "Underlined text style."
	case "font-blink":
		return "Blinking text style."
	case "font-reverse":
		return "Reverse video (swap foreground/background)."
	case "font-strikethrough":
		return "Strikethrough text style."
	}

	// Check for parameterized classes
	if strings.HasPrefix(class, "gap-") {
		n := strings.TrimPrefix(class, "gap-")
		return fmt.Sprintf("Set gap between children to %s characters.", n)
	}
	if strings.HasPrefix(class, "p-") {
		n := strings.TrimPrefix(class, "p-")
		return fmt.Sprintf("Set padding to %s on all sides.", n)
	}
	if strings.HasPrefix(class, "px-") {
		n := strings.TrimPrefix(class, "px-")
		return fmt.Sprintf("Set horizontal padding (left and right) to %s.", n)
	}
	if strings.HasPrefix(class, "py-") {
		n := strings.TrimPrefix(class, "py-")
		return fmt.Sprintf("Set vertical padding (top and bottom) to %s.", n)
	}
	if strings.HasPrefix(class, "pt-") {
		n := strings.TrimPrefix(class, "pt-")
		return fmt.Sprintf("Set top padding to %s.", n)
	}
	if strings.HasPrefix(class, "pb-") {
		n := strings.TrimPrefix(class, "pb-")
		return fmt.Sprintf("Set bottom padding to %s.", n)
	}
	if strings.HasPrefix(class, "pl-") {
		n := strings.TrimPrefix(class, "pl-")
		return fmt.Sprintf("Set left padding to %s.", n)
	}
	if strings.HasPrefix(class, "pr-") {
		n := strings.TrimPrefix(class, "pr-")
		return fmt.Sprintf("Set right padding to %s.", n)
	}
	if strings.HasPrefix(class, "m-") {
		n := strings.TrimPrefix(class, "m-")
		return fmt.Sprintf("Set margin to %s on all sides.", n)
	}
	if strings.HasPrefix(class, "mx-") {
		n := strings.TrimPrefix(class, "mx-")
		return fmt.Sprintf("Set horizontal margin (left and right) to %s.", n)
	}
	if strings.HasPrefix(class, "my-") {
		n := strings.TrimPrefix(class, "my-")
		return fmt.Sprintf("Set vertical margin (top and bottom) to %s.", n)
	}
	if strings.HasPrefix(class, "mt-") {
		n := strings.TrimPrefix(class, "mt-")
		return fmt.Sprintf("Set top margin to %s.", n)
	}
	if strings.HasPrefix(class, "mb-") {
		n := strings.TrimPrefix(class, "mb-")
		return fmt.Sprintf("Set bottom margin to %s.", n)
	}
	if strings.HasPrefix(class, "ml-") {
		n := strings.TrimPrefix(class, "ml-")
		return fmt.Sprintf("Set left margin to %s.", n)
	}
	if strings.HasPrefix(class, "mr-") {
		n := strings.TrimPrefix(class, "mr-")
		return fmt.Sprintf("Set right margin to %s.", n)
	}

	// Text colors
	if strings.HasPrefix(class, "text-") {
		color := strings.TrimPrefix(class, "text-")
		return fmt.Sprintf("Set text color to **%s**.", color)
	}

	// Background colors
	if strings.HasPrefix(class, "bg-") {
		color := strings.TrimPrefix(class, "bg-")
		return fmt.Sprintf("Set background color to **%s**.", color)
	}

	// Width/height
	if strings.HasPrefix(class, "w-") {
		val := strings.TrimPrefix(class, "w-")
		if strings.HasSuffix(val, "%") {
			return fmt.Sprintf("Set width to %s of parent.", val)
		}
		return fmt.Sprintf("Set width to %s characters.", val)
	}
	if strings.HasPrefix(class, "h-") {
		val := strings.TrimPrefix(class, "h-")
		if strings.HasSuffix(val, "%") {
			return fmt.Sprintf("Set height to %s of parent.", val)
		}
		return fmt.Sprintf("Set height to %s rows.", val)
	}
	if strings.HasPrefix(class, "min-w-") {
		return fmt.Sprintf("Set minimum width to %s.", strings.TrimPrefix(class, "min-w-"))
	}
	if strings.HasPrefix(class, "min-h-") {
		return fmt.Sprintf("Set minimum height to %s.", strings.TrimPrefix(class, "min-h-"))
	}
	if strings.HasPrefix(class, "max-w-") {
		return fmt.Sprintf("Set maximum width to %s.", strings.TrimPrefix(class, "max-w-"))
	}
	if strings.HasPrefix(class, "max-h-") {
		return fmt.Sprintf("Set maximum height to %s.", strings.TrimPrefix(class, "max-h-"))
	}

	// Flex grow/shrink
	if strings.HasPrefix(class, "grow-") {
		return fmt.Sprintf("Set flex grow factor to %s.", strings.TrimPrefix(class, "grow-"))
	}
	if class == "grow" {
		return "Set flex grow factor to 1."
	}
	if strings.HasPrefix(class, "shrink-") {
		return fmt.Sprintf("Set flex shrink factor to %s.", strings.TrimPrefix(class, "shrink-"))
	}
	if class == "shrink" {
		return "Set flex shrink factor to 1."
	}
	if class == "shrink-0" {
		return "Prevent element from shrinking."
	}

	return ""
}

// attributeInfo holds information about an attribute at a position.
type attributeInfo struct {
	tag  string
	name string
}

// getAttributeAtPosition finds attribute information at the given position.
func (s *Server) getAttributeAtPosition(doc *Document, pos Position) *attributeInfo {
	if doc.AST == nil {
		return nil
	}

	offset := PositionToOffset(doc.Content, pos)

	// Walk the AST to find the element and attribute
	for _, comp := range doc.AST.Components {
		if info := s.findAttributeInNodes(comp.Body, offset, doc.Content); info != nil {
			return info
		}
	}

	return nil
}

// findAttributeInNodes recursively searches for an attribute at the given offset.
func (s *Server) findAttributeInNodes(nodes []tuigen.Node, offset int, content string) *attributeInfo {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.Element:
			// Check if offset is within an attribute
			for _, attr := range n.Attributes {
				attrOffset := PositionToOffset(content, Position{
					Line:      attr.Position.Line - 1,
					Character: attr.Position.Column - 1,
				})
				attrEnd := attrOffset + len(attr.Name)
				if offset >= attrOffset && offset <= attrEnd {
					return &attributeInfo{tag: n.Tag, name: attr.Name}
				}
			}
			// Check children
			if info := s.findAttributeInNodes(n.Children, offset, content); info != nil {
				return info
			}
		case *tuigen.ForLoop:
			if info := s.findAttributeInNodes(n.Body, offset, content); info != nil {
				return info
			}
		case *tuigen.IfStmt:
			if info := s.findAttributeInNodes(n.Then, offset, content); info != nil {
				return info
			}
			if info := s.findAttributeInNodes(n.Else, offset, content); info != nil {
				return info
			}
		case *tuigen.LetBinding:
			if n.Element != nil {
				if info := s.findAttributeInNodes([]tuigen.Node{n.Element}, offset, content); info != nil {
					return info
				}
			}
		}
	}
	return nil
}

// hoverForAttribute creates hover content for an attribute.
func (s *Server) hoverForAttribute(tag, attrName string) *Hover {
	attrs := getElementAttributes(tag)
	for _, attr := range attrs {
		if attr.Name == attrName {
			return &Hover{
				Contents: MarkupContent{
					Kind:  "markdown",
					Value: fmt.Sprintf("**%s** (`%s`)\n\n%s", attr.Name, attr.Type, attr.Description),
				},
			}
		}
	}
	return nil
}

// getGoplsHover gets hover information from gopls for Go expressions.
func (s *Server) getGoplsHover(doc *Document, pos Position) (*Hover, error) {
	if s.goplsProxy == nil {
		return nil, nil
	}

	// Get the cached virtual file
	cached := s.virtualFiles.Get(doc.URI)
	if cached == nil || cached.SourceMap == nil {
		return nil, nil
	}

	// Translate position from .tui to .go
	goLine, goCol, found := cached.SourceMap.TuiToGo(pos.Line, pos.Character)
	if !found {
		log.Server("No mapping found for hover position %d:%d", pos.Line, pos.Character)
		return nil, nil
	}

	log.Server("Translated hover position %d:%d -> %d:%d", pos.Line, pos.Character, goLine, goCol)

	// Call gopls for hover
	goplsHover, err := s.goplsProxy.Hover(cached.GoURI, gopls.Position{
		Line:      goLine,
		Character: goCol,
	})
	if err != nil {
		return nil, err
	}

	if goplsHover == nil {
		return nil, nil
	}

	// Convert gopls hover to our Hover format
	hover := &Hover{
		Contents: MarkupContent{
			Kind:  goplsHover.Contents.Kind,
			Value: goplsHover.Contents.Value,
		},
	}

	// Translate range back to .tui positions if present
	// Only include range if we can successfully map both start and end positions
	if goplsHover.Range != nil {
		log.Server("gopls hover range: (%d:%d)-(%d:%d)",
			goplsHover.Range.Start.Line, goplsHover.Range.Start.Character,
			goplsHover.Range.End.Line, goplsHover.Range.End.Character)
		tuiStartLine, tuiStartCol, startFound := cached.SourceMap.GoToTui(goplsHover.Range.Start.Line, goplsHover.Range.Start.Character)
		tuiEndLine, tuiEndCol, endFound := cached.SourceMap.GoToTui(goplsHover.Range.End.Line, goplsHover.Range.End.Character)
		log.Server("GoToTui hover translation: start(%d:%d->%d:%d, found=%v) end(%d:%d->%d:%d, found=%v)",
			goplsHover.Range.Start.Line, goplsHover.Range.Start.Character, tuiStartLine, tuiStartCol, startFound,
			goplsHover.Range.End.Line, goplsHover.Range.End.Character, tuiEndLine, tuiEndCol, endFound)
		if startFound && endFound {
			hover.Range = &Range{
				Start: Position{Line: tuiStartLine, Character: tuiStartCol},
				End:   Position{Line: tuiEndLine, Character: tuiEndCol},
			}
			log.Server("Final hover range: (%d:%d)-(%d:%d)", tuiStartLine, tuiStartCol, tuiEndLine, tuiEndCol)
		}
		// If mapping not found, omit range - the editor will determine it from position
	}

	return hover, nil
}
