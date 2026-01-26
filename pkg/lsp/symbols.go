package lsp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/grindlemire/go-tui/pkg/lsp/log"
	"github.com/grindlemire/go-tui/pkg/tuigen"
)

// DocumentSymbolParams represents textDocument/documentSymbol parameters.
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// WorkspaceSymbolParams represents workspace/symbol parameters.
type WorkspaceSymbolParams struct {
	Query string `json:"query"`
}

// DocumentSymbol represents a symbol found in a document.
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           SymbolKind       `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// SymbolInformation represents a symbol for workspace symbols.
type SymbolInformation struct {
	Name          string     `json:"name"`
	Kind          SymbolKind `json:"kind"`
	Location      Location   `json:"location"`
	ContainerName string     `json:"containerName,omitempty"`
}

// SymbolKind represents the kind of symbol.
type SymbolKind int

const (
	SymbolKindFile          SymbolKind = 1
	SymbolKindModule        SymbolKind = 2
	SymbolKindNamespace     SymbolKind = 3
	SymbolKindPackage       SymbolKind = 4
	SymbolKindClass         SymbolKind = 5
	SymbolKindMethod        SymbolKind = 6
	SymbolKindProperty      SymbolKind = 7
	SymbolKindField         SymbolKind = 8
	SymbolKindConstructor   SymbolKind = 9
	SymbolKindEnum          SymbolKind = 10
	SymbolKindInterface     SymbolKind = 11
	SymbolKindFunction      SymbolKind = 12
	SymbolKindVariable      SymbolKind = 13
	SymbolKindConstant      SymbolKind = 14
	SymbolKindString        SymbolKind = 15
	SymbolKindNumber        SymbolKind = 16
	SymbolKindBoolean       SymbolKind = 17
	SymbolKindArray         SymbolKind = 18
	SymbolKindObject        SymbolKind = 19
	SymbolKindKey           SymbolKind = 20
	SymbolKindNull          SymbolKind = 21
	SymbolKindEnumMember    SymbolKind = 22
	SymbolKindStruct        SymbolKind = 23
	SymbolKindEvent         SymbolKind = 24
	SymbolKindOperator      SymbolKind = 25
	SymbolKindTypeParameter SymbolKind = 26
)

// positionBefore returns true if a is before b.
func positionBefore(a, b Position) bool {
	if a.Line != b.Line {
		return a.Line < b.Line
	}
	return a.Character < b.Character
}

// positionAfter returns true if a is after b.
func positionAfter(a, b Position) bool {
	if a.Line != b.Line {
		return a.Line > b.Line
	}
	return a.Character > b.Character
}

// clampSelectionRange ensures selectionRange is contained within fullRange.
func clampSelectionRange(fullRange, selectionRange Range) Range {
	// Clamp start
	if positionBefore(selectionRange.Start, fullRange.Start) {
		selectionRange.Start = fullRange.Start
	}
	// Clamp end
	if positionAfter(selectionRange.End, fullRange.End) {
		selectionRange.End = fullRange.End
	}
	// Ensure start <= end
	if positionAfter(selectionRange.Start, selectionRange.End) {
		selectionRange.Start = selectionRange.End
	}
	return selectionRange
}

// handleDocumentSymbol handles textDocument/documentSymbol requests.
func (s *Server) handleDocumentSymbol(params json.RawMessage) (any, *Error) {
	var p DocumentSymbolParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &Error{Code: CodeInvalidParams, Message: err.Error()}
	}

	log.Server("DocumentSymbol request for %s", p.TextDocument.URI)

	doc := s.docs.Get(p.TextDocument.URI)
	if doc == nil || doc.AST == nil {
		return []DocumentSymbol{}, nil
	}

	return s.getDocumentSymbols(doc), nil
}

// getDocumentSymbols extracts symbols from a document.
func (s *Server) getDocumentSymbols(doc *Document) []DocumentSymbol {
	var symbols []DocumentSymbol

	for _, comp := range doc.AST.Components {
		symbol := s.componentToSymbol(comp, doc.Content)
		symbols = append(symbols, symbol)
	}

	for _, fn := range doc.AST.Funcs {
		symbol := s.funcToSymbol(fn, doc.Content)
		symbols = append(symbols, symbol)
	}

	return symbols
}

// componentToSymbol converts a component to a document symbol.
func (s *Server) componentToSymbol(comp *tuigen.Component, content string) DocumentSymbol {
	// Build detail string from parameters
	var params []string
	for _, p := range comp.Params {
		params = append(params, fmt.Sprintf("%s %s", p.Name, p.Type))
	}
	detail := fmt.Sprintf("(%s)", strings.Join(params, ", "))

	// Calculate range - from @component to closing brace
	// For simplicity, use the name position for selection range
	startPos := Position{
		Line:      comp.Position.Line - 1,
		Character: comp.Position.Column - 1,
	}

	nameEndPos := Position{
		Line:      comp.Position.Line - 1,
		Character: comp.Position.Column - 1 + len("@component") + 1 + len(comp.Name),
	}

	// Find end of component (rough estimate - find matching brace)
	endPos := s.findComponentEnd(comp, content)

	fullRange := Range{
		Start: startPos,
		End:   endPos,
	}
	selRange := clampSelectionRange(fullRange, Range{
		Start: startPos,
		End:   nameEndPos,
	})

	symbol := DocumentSymbol{
		Name:           comp.Name,
		Detail:         detail,
		Kind:           SymbolKindFunction,
		Range:          fullRange,
		SelectionRange: selRange,
	}

	// Add child symbols (let bindings, nested elements with IDs)
	for _, node := range comp.Body {
		if child := s.nodeToSymbol(node); child != nil {
			symbol.Children = append(symbol.Children, *child)
		}
	}

	return symbol
}

// funcToSymbol converts a Go function to a document symbol.
func (s *Server) funcToSymbol(fn *tuigen.GoFunc, content string) DocumentSymbol {
	// Extract function name from code
	name := extractFuncName(fn.Code)

	startPos := Position{
		Line:      fn.Position.Line - 1,
		Character: fn.Position.Column - 1,
	}

	// Find end of function
	endPos := startPos
	offset := PositionToOffset(content, startPos)
	if offset < len(content) {
		// Find the matching closing brace
		braceCount := 0
		started := false
		for i := offset; i < len(content); i++ {
			if content[i] == '{' {
				braceCount++
				started = true
			} else if content[i] == '}' {
				braceCount--
				if started && braceCount == 0 {
					endPos = OffsetToPosition(content, i+1)
					break
				}
			}
		}
	}

	fullRange := Range{
		Start: startPos,
		End:   endPos,
	}
	selRange := clampSelectionRange(fullRange, Range{
		Start: startPos,
		End:   Position{Line: startPos.Line, Character: startPos.Character + len("func") + 1 + len(name)},
	})

	return DocumentSymbol{
		Name:           name,
		Detail:         "func",
		Kind:           SymbolKindFunction,
		Range:          fullRange,
		SelectionRange: selRange,
	}
}

// extractFuncName extracts the function name from Go function code.
func extractFuncName(code string) string {
	// Skip "func "
	code = strings.TrimPrefix(strings.TrimSpace(code), "func ")

	// Find opening paren
	idx := strings.Index(code, "(")
	if idx == -1 {
		return "unknown"
	}

	name := strings.TrimSpace(code[:idx])

	// Handle methods: (receiver) Name
	if strings.HasPrefix(name, "(") {
		// Find closing paren of receiver
		closeIdx := strings.Index(name, ")")
		if closeIdx != -1 {
			name = strings.TrimSpace(name[closeIdx+1:])
		}
	}

	return name
}

// nodeToSymbol converts an AST node to a symbol if applicable.
func (s *Server) nodeToSymbol(node tuigen.Node) *DocumentSymbol {
	switch n := node.(type) {
	case *tuigen.LetBinding:
		fullRange := TuigenPosToRange(n.Position, len(n.Name)+5) // "@let " + name
		selRange := clampSelectionRange(fullRange, Range{
			Start: Position{Line: n.Position.Line - 1, Character: n.Position.Column - 1 + 5},
			End:   Position{Line: n.Position.Line - 1, Character: n.Position.Column - 1 + 5 + len(n.Name)},
		})
		return &DocumentSymbol{
			Name:           n.Name,
			Detail:         "let binding",
			Kind:           SymbolKindVariable,
			Range:          fullRange,
			SelectionRange: selRange,
		}
	case *tuigen.Element:
		// Only create symbol for elements with an ID attribute
		for _, attr := range n.Attributes {
			if attr.Name == "id" {
				if str, ok := attr.Value.(*tuigen.StringLit); ok {
					fullRange := TuigenPosToRange(n.Position, len(n.Tag)+2) // "<tag>"
					selRange := clampSelectionRange(fullRange, Range{
						Start: Position{Line: attr.Position.Line - 1, Character: attr.Position.Column - 1},
						End:   Position{Line: attr.Position.Line - 1, Character: attr.Position.Column - 1 + len(attr.Name) + 2 + len(str.Value)},
					})
					return &DocumentSymbol{
						Name:           str.Value,
						Detail:         fmt.Sprintf("<%s>", n.Tag),
						Kind:           SymbolKindField,
						Range:          fullRange,
						SelectionRange: selRange,
					}
				}
			}
		}
	}
	return nil
}

// findComponentEnd finds the end position of a component.
func (s *Server) findComponentEnd(comp *tuigen.Component, content string) Position {
	startOffset := PositionToOffset(content, Position{
		Line:      comp.Position.Line - 1,
		Character: comp.Position.Column - 1,
	})

	braceCount := 0
	started := false
	for i := startOffset; i < len(content); i++ {
		if content[i] == '{' {
			braceCount++
			started = true
		} else if content[i] == '}' {
			braceCount--
			if started && braceCount == 0 {
				return OffsetToPosition(content, i+1)
			}
		}
	}

	// Fallback to a few lines after start
	return Position{Line: comp.Position.Line - 1 + 5, Character: 0}
}

// handleWorkspaceSymbol handles workspace/symbol requests.
func (s *Server) handleWorkspaceSymbol(params json.RawMessage) (any, *Error) {
	var p WorkspaceSymbolParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &Error{Code: CodeInvalidParams, Message: err.Error()}
	}

	log.Server("WorkspaceSymbol request: %q", p.Query)

	query := strings.ToLower(p.Query)
	var symbols []SymbolInformation

	// Search all indexed components
	for _, name := range s.index.All() {
		if query == "" || strings.Contains(strings.ToLower(name), query) {
			info := s.index.GetInfo(name)
			if info != nil {
				symbols = append(symbols, SymbolInformation{
					Name:     name,
					Kind:     SymbolKindFunction,
					Location: info.Location,
				})
			}
		}
	}

	return symbols, nil
}
