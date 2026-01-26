package lsp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/grindlemire/go-tui/pkg/lsp/gopls"
	"github.com/grindlemire/go-tui/pkg/lsp/log"
	"github.com/grindlemire/go-tui/pkg/tuigen"
)

// CompletionParams represents textDocument/completion parameters.
type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      *CompletionContext     `json:"context,omitempty"`
}

// CompletionContext contains additional information about the context.
type CompletionContext struct {
	TriggerKind      int    `json:"triggerKind"`
	TriggerCharacter string `json:"triggerCharacter,omitempty"`
}

// CompletionList represents a list of completion items.
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// CompletionItem represents a completion suggestion.
type CompletionItem struct {
	Label         string             `json:"label"`
	Kind          CompletionItemKind `json:"kind,omitempty"`
	Detail        string             `json:"detail,omitempty"`
	Documentation *MarkupContent     `json:"documentation,omitempty"`
	InsertText    string             `json:"insertText,omitempty"`
	FilterText    string             `json:"filterText,omitempty"`
}

// CompletionItemKind represents the kind of completion item.
type CompletionItemKind int

const (
	CompletionItemKindText          CompletionItemKind = 1
	CompletionItemKindMethod        CompletionItemKind = 2
	CompletionItemKindFunction      CompletionItemKind = 3
	CompletionItemKindConstructor   CompletionItemKind = 4
	CompletionItemKindField         CompletionItemKind = 5
	CompletionItemKindVariable      CompletionItemKind = 6
	CompletionItemKindClass         CompletionItemKind = 7
	CompletionItemKindInterface     CompletionItemKind = 8
	CompletionItemKindModule        CompletionItemKind = 9
	CompletionItemKindProperty      CompletionItemKind = 10
	CompletionItemKindUnit          CompletionItemKind = 11
	CompletionItemKindValue         CompletionItemKind = 12
	CompletionItemKindEnum          CompletionItemKind = 13
	CompletionItemKindKeyword       CompletionItemKind = 14
	CompletionItemKindSnippet       CompletionItemKind = 15
	CompletionItemKindColor         CompletionItemKind = 16
	CompletionItemKindFile          CompletionItemKind = 17
	CompletionItemKindReference     CompletionItemKind = 18
	CompletionItemKindFolder        CompletionItemKind = 19
	CompletionItemKindEnumMember    CompletionItemKind = 20
	CompletionItemKindConstant      CompletionItemKind = 21
	CompletionItemKindStruct        CompletionItemKind = 22
	CompletionItemKindEvent         CompletionItemKind = 23
	CompletionItemKindOperator      CompletionItemKind = 24
	CompletionItemKindTypeParameter CompletionItemKind = 25
)

// handleCompletion handles textDocument/completion requests.
func (s *Server) handleCompletion(params json.RawMessage) (any, *Error) {
	var p CompletionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &Error{Code: CodeInvalidParams, Message: err.Error()}
	}

	log.Server("Completion request at %s:%d:%d", p.TextDocument.URI, p.Position.Line, p.Position.Character)

	doc := s.docs.Get(p.TextDocument.URI)
	if doc == nil {
		return CompletionList{Items: []CompletionItem{}}, nil
	}

	// Check if we're inside a Go expression - use gopls if available
	if s.isInGoExpression(doc, p.Position) {
		items, err := s.getGoplsCompletions(doc, p.Position)
		if err != nil {
			log.Server("gopls completion error: %v", err)
			// Fall through to TUI completions
		} else if len(items) > 0 {
			return CompletionList{
				IsIncomplete: false,
				Items:        items,
			}, nil
		}
	}

	// Determine context from trigger character or position
	trigger := ""
	if p.Context != nil && p.Context.TriggerCharacter != "" {
		trigger = p.Context.TriggerCharacter
	} else {
		// Look at character before cursor
		trigger = s.getCharBeforePosition(doc, p.Position)
	}

	log.Server("Completion trigger: %q", trigger)

	var items []CompletionItem

	switch trigger {
	case "@":
		// Component call or DSL keyword
		items = append(items, s.getComponentCompletions()...)
		items = append(items, s.getDSLKeywordCompletions()...)
	case "<":
		// Element tag
		items = append(items, s.getElementCompletions()...)
	case "{":
		// Start of Go expression - try gopls
		goplsItems, err := s.getGoplsCompletions(doc, p.Position)
		if err == nil && len(goplsItems) > 0 {
			items = append(items, goplsItems...)
		}
	default:
		// Context-based completion
		items = append(items, s.getContextualCompletions(doc, p.Position)...)
	}

	return CompletionList{
		IsIncomplete: false,
		Items:        items,
	}, nil
}

// getCharBeforePosition returns the character immediately before the cursor.
func (s *Server) getCharBeforePosition(doc *Document, pos Position) string {
	offset := PositionToOffset(doc.Content, pos)
	if offset <= 0 {
		return ""
	}
	return string(doc.Content[offset-1])
}

// getComponentCompletions returns completions for component calls.
func (s *Server) getComponentCompletions() []CompletionItem {
	var items []CompletionItem

	for _, name := range s.index.All() {
		info := s.index.GetInfo(name)
		if info == nil {
			continue
		}

		// Build parameter string
		var params []string
		for _, p := range info.Params {
			params = append(params, fmt.Sprintf("%s %s", p.Name, p.Type))
		}
		detail := fmt.Sprintf("(%s)", strings.Join(params, ", "))

		items = append(items, CompletionItem{
			Label:      name,
			Kind:       CompletionItemKindFunction,
			Detail:     detail,
			InsertText: name + "()",
			FilterText: "@" + name,
		})
	}

	return items
}

// getDSLKeywordCompletions returns completions for DSL keywords.
func (s *Server) getDSLKeywordCompletions() []CompletionItem {
	return []CompletionItem{
		{
			Label:      "component",
			Kind:       CompletionItemKindKeyword,
			Detail:     "Define a new component",
			InsertText: "component ${1:Name}(${2:params}) {\n\t$0\n}",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "Define a new TUI component.\n\n```tui\n@component MyComponent(title string) {\n    <div>{title}</div>\n}\n```",
			},
		},
		{
			Label:      "for",
			Kind:       CompletionItemKindKeyword,
			Detail:     "Loop over items",
			InsertText: "for ${1:i}, ${2:item} := range ${3:items} {\n\t$0\n}",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "Loop over a collection.\n\n```tui\n@for i, item := range items {\n    <span>{item}</span>\n}\n```",
			},
		},
		{
			Label:      "if",
			Kind:       CompletionItemKindKeyword,
			Detail:     "Conditional rendering",
			InsertText: "if ${1:condition} {\n\t$0\n}",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "Conditionally render content.\n\n```tui\n@if showHeader {\n    <span>Header</span>\n}\n```",
			},
		},
		{
			Label:      "let",
			Kind:       CompletionItemKindKeyword,
			Detail:     "Bind element to variable",
			InsertText: "let ${1:name} = ",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "Bind an element to a variable for later reference.\n\n```tui\n@let header = <div>Header</div>\n```",
			},
		},
	}
}

// getElementCompletions returns completions for element tags.
func (s *Server) getElementCompletions() []CompletionItem {
	return []CompletionItem{
		{
			Label:      "div",
			Kind:       CompletionItemKindClass,
			Detail:     "Block container",
			InsertText: "div>$0</div>",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "A flexbox container for layout.\n\n```tui\n<div class=\"flex-col gap-1 p-2\">\n    <span>Child</span>\n</div>\n```",
			},
		},
		{
			Label:      "span",
			Kind:       CompletionItemKindClass,
			Detail:     "Inline text container",
			InsertText: "span>$0</span>",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "Display inline text content.\n\n```tui\n<span class=\"font-bold text-cyan\">Hello, World!</span>\n```",
			},
		},
		{
			Label:      "p",
			Kind:       CompletionItemKindClass,
			Detail:     "Paragraph",
			InsertText: "p>$0</p>",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "A paragraph of text.\n\n```tui\n<p>Some paragraph text here.</p>\n```",
			},
		},
		{
			Label:      "ul",
			Kind:       CompletionItemKindClass,
			Detail:     "Unordered list",
			InsertText: "ul>\n\t<li>$0</li>\n</ul>",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "An unordered list container.\n\n```tui\n<ul>\n    <li>Item 1</li>\n    <li>Item 2</li>\n</ul>\n```",
			},
		},
		{
			Label:      "li",
			Kind:       CompletionItemKindClass,
			Detail:     "List item",
			InsertText: "li>$0</li>",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "A list item.\n\n```tui\n<li>Item content</li>\n```",
			},
		},
		{
			Label:      "button",
			Kind:       CompletionItemKindClass,
			Detail:     "Clickable button",
			InsertText: "button>$0</button>",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "A clickable button element.\n\n```tui\n<button onEvent={handleClick}>Click me</button>\n```",
			},
		},
		{
			Label:      "input",
			Kind:       CompletionItemKindClass,
			Detail:     "Text input",
			InsertText: "input value={$1} />",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "A text input field.\n\n```tui\n<input value={inputValue} placeholder=\"Enter text...\" />\n```",
			},
		},
		{
			Label:      "table",
			Kind:       CompletionItemKindClass,
			Detail:     "Table container",
			InsertText: "table>$0</table>",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "A table container for tabular data.\n\n```tui\n<table>\n    ...\n</table>\n```",
			},
		},
		{
			Label:      "progress",
			Kind:       CompletionItemKindClass,
			Detail:     "Progress bar",
			InsertText: "progress value={$1} max={100} />",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "A progress bar.\n\n```tui\n<progress value={50} max={100} />\n```",
			},
		},
		{
			Label:      "hr",
			Kind:       CompletionItemKindClass,
			Detail:     "Horizontal rule",
			InsertText: "hr/>",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "A horizontal dividing line.\n\n```tui\n<hr/>\n<hr class=\"border-double text-cyan\"/>\n```",
			},
		},
		{
			Label:      "br",
			Kind:       CompletionItemKindClass,
			Detail:     "Line break",
			InsertText: "br/>",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "An empty line break.\n\n```tui\n<span>Line 1</span>\n<br/>\n<span>Line 2</span>\n```",
			},
		},
	}
}

// getContextualCompletions returns completions based on cursor context.
func (s *Server) getContextualCompletions(doc *Document, pos Position) []CompletionItem {
	// Check if we're inside an element tag (for attributes)
	tag := s.getEnclosingTagAtPosition(doc, pos)
	if tag != "" {
		return s.getAttributeCompletions(tag)
	}

	// Default: offer all completions
	var items []CompletionItem
	items = append(items, s.getComponentCompletions()...)
	items = append(items, s.getDSLKeywordCompletions()...)
	items = append(items, s.getElementCompletions()...)
	return items
}

// getEnclosingTagAtPosition finds the element tag that contains the cursor.
func (s *Server) getEnclosingTagAtPosition(doc *Document, pos Position) string {
	offset := PositionToOffset(doc.Content, pos)

	// Search backwards for < to find opening tag
	for i := offset - 1; i >= 0; i-- {
		if doc.Content[i] == '<' {
			// Found opening tag, extract tag name
			j := i + 1
			for j < len(doc.Content) && isWordChar(doc.Content[j]) {
				j++
			}
			if j > i+1 {
				tagName := doc.Content[i+1 : j]
				// Make sure we're still inside the tag (haven't hit >)
				for k := j; k < offset; k++ {
					if doc.Content[k] == '>' {
						return "" // Past the opening tag
					}
				}
				return tagName
			}
			break
		}
		if doc.Content[i] == '>' {
			// Hit a closing bracket, we're not in an opening tag
			break
		}
	}
	return ""
}

// getAttributeCompletions returns completions for attributes of an element.
func (s *Server) getAttributeCompletions(tag string) []CompletionItem {
	attrs := getElementAttributes(tag)
	var items []CompletionItem

	for _, attr := range attrs {
		insertText := attr.Name + "="
		if attr.Type == "bool" {
			insertText = attr.Name // Boolean attributes don't need =value
		} else if attr.Type == "string" {
			insertText = attr.Name + `="${1}"`
		} else {
			insertText = attr.Name + "={$1}"
		}

		items = append(items, CompletionItem{
			Label:      attr.Name,
			Kind:       CompletionItemKindProperty,
			Detail:     attr.Type,
			InsertText: insertText,
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: attr.Description,
			},
		})
	}

	return items
}

// isInGoExpression checks if the cursor position is inside a Go expression.
// This includes: {} expressions, helper function bodies, and GoCode statements.
func (s *Server) isInGoExpression(doc *Document, pos Position) bool {
	// First, check if we have a source map mapping for this position
	// (this covers all Go expressions that were mapped during generation)
	cached := s.virtualFiles.Get(doc.URI)
	if cached != nil && cached.SourceMap != nil {
		if cached.SourceMap.IsInGoExpression(pos.Line, pos.Character) {
			return true
		}
	}

	// Check if we're inside a {} expression (curly brace heuristic)
	offset := PositionToOffset(doc.Content, pos)

	// Search backwards for { or }
	braceDepth := 0
	for i := offset - 1; i >= 0; i-- {
		switch doc.Content[i] {
		case '{':
			if braceDepth == 0 {
				return true // Inside a { block
			}
			braceDepth--
		case '}':
			braceDepth++
		case '\n':
			// Don't search past line boundaries for simple heuristic
			// unless we're in a nested brace
			if braceDepth == 0 {
				break
			}
		}
	}

	// Check if we're inside a helper function body (func ... { ... })
	if doc.AST != nil {
		for _, fn := range doc.AST.Funcs {
			// Convert position to 1-indexed for AST comparison
			line := pos.Line + 1
			fnEndLine := fn.Position.Line + countLines(fn.Code) - 1

			if line >= fn.Position.Line && line <= fnEndLine {
				return true
			}
		}

		// Check if we're on a GoCode line (like shouldShowHeader := true)
		for _, comp := range doc.AST.Components {
			if goCode := s.findGoCodeAtLine(comp.Body, pos.Line+1); goCode != nil {
				return true
			}
		}
	}

	return false
}

// countLines counts the number of lines in a string.
func countLines(s string) int {
	if s == "" {
		return 0
	}
	count := 1
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}

// findGoCodeAtLine finds a GoCode node at the given line.
func (s *Server) findGoCodeAtLine(nodes []tuigen.Node, line int) *tuigen.GoCode {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.GoCode:
			if n != nil && n.Position.Line == line {
				return n
			}
		case *tuigen.Element:
			if n != nil {
				if found := s.findGoCodeAtLine(n.Children, line); found != nil {
					return found
				}
			}
		case *tuigen.ForLoop:
			if n != nil {
				if found := s.findGoCodeAtLine(n.Body, line); found != nil {
					return found
				}
			}
		case *tuigen.IfStmt:
			if n != nil {
				if found := s.findGoCodeAtLine(n.Then, line); found != nil {
					return found
				}
				if found := s.findGoCodeAtLine(n.Else, line); found != nil {
					return found
				}
			}
		case *tuigen.LetBinding:
			if n != nil && n.Element != nil {
				if found := s.findGoCodeAtLine(n.Element.Children, line); found != nil {
					return found
				}
			}
		}
	}
	return nil
}

// getGoplsCompletions gets completion items from gopls for Go expressions.
func (s *Server) getGoplsCompletions(doc *Document, pos Position) ([]CompletionItem, error) {
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
		log.Server("No mapping found for position %d:%d", pos.Line, pos.Character)
		return nil, nil
	}

	log.Server("Translated position %d:%d -> %d:%d", pos.Line, pos.Character, goLine, goCol)

	// Call gopls for completions
	goplsItems, err := s.goplsProxy.Completion(cached.GoURI, gopls.Position{
		Line:      goLine,
		Character: goCol,
	})
	if err != nil {
		return nil, err
	}

	// Convert gopls items to our CompletionItem format
	var items []CompletionItem
	for _, gi := range goplsItems {
		item := CompletionItem{
			Label:      gi.Label,
			Kind:       CompletionItemKind(gi.Kind),
			Detail:     gi.Detail,
			InsertText: gi.InsertText,
			FilterText: gi.FilterText,
		}
		if gi.Documentation != nil {
			item.Documentation = &MarkupContent{
				Kind:  gi.Documentation.Kind,
				Value: gi.Documentation.Value,
			}
		}
		items = append(items, item)
	}

	return items, nil
}
