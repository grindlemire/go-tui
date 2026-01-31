package provider

import (
	"fmt"
	"strings"

	"github.com/grindlemire/go-tui/internal/lsp/log"
	"github.com/grindlemire/go-tui/internal/tuigen"
)

// documentSymbolProvider implements DocumentSymbolProvider.
type documentSymbolProvider struct{}

// NewDocumentSymbolProvider creates a new document symbol provider.
func NewDocumentSymbolProvider() DocumentSymbolProvider {
	return &documentSymbolProvider{}
}

func (s *documentSymbolProvider) DocumentSymbols(doc *Document) ([]DocumentSymbol, error) {
	log.Server("DocumentSymbol provider for %s", doc.URI)

	if doc.AST == nil {
		return []DocumentSymbol{}, nil
	}

	var symbols []DocumentSymbol

	for _, comp := range doc.AST.Components {
		symbol := componentToSymbol(comp, doc.Content)
		symbols = append(symbols, symbol)
	}

	for _, fn := range doc.AST.Funcs {
		symbol := funcToSymbol(fn, doc.Content)
		symbols = append(symbols, symbol)
	}

	return symbols, nil
}

// componentToSymbol converts a component AST node to a DocumentSymbol.
func componentToSymbol(comp *tuigen.Component, content string) DocumentSymbol {
	// Build detail from parameters
	var params []string
	for _, p := range comp.Params {
		params = append(params, fmt.Sprintf("%s %s", p.Name, p.Type))
	}
	detail := fmt.Sprintf("(%s)", strings.Join(params, ", "))

	startPos := Position{
		Line:      comp.Position.Line - 1,
		Character: comp.Position.Column - 1,
	}

	nameEndPos := Position{
		Line:      comp.Position.Line - 1,
		Character: comp.Position.Column - 1 + len("templ") + 1 + len(comp.Name),
	}

	endPos := findComponentEnd(comp, content)

	fullRange := Range{Start: startPos, End: endPos}
	selRange := clampSelectionRange(fullRange, Range{Start: startPos, End: nameEndPos})

	symbol := DocumentSymbol{
		Name:           comp.Name,
		Detail:         detail,
		Kind:           SymbolKindFunction,
		Range:          fullRange,
		SelectionRange: selRange,
	}

	// Add child symbols (let bindings, elements with IDs)
	for _, node := range comp.Body {
		if child := nodeToSymbol(node); child != nil {
			symbol.Children = append(symbol.Children, *child)
		}
	}

	return symbol
}

// funcToSymbol converts a Go function to a DocumentSymbol.
func funcToSymbol(fn *tuigen.GoFunc, content string) DocumentSymbol {
	name := extractFuncName(fn.Code)

	startPos := Position{
		Line:      fn.Position.Line - 1,
		Character: fn.Position.Column - 1,
	}

	// Find end of function by searching for matching close brace
	endPos := startPos
	offset := PositionToOffset(content, startPos)
	if offset < len(content) {
		braceCount := 0
		started := false
		for i := offset; i < len(content); i++ {
			if content[i] == '{' {
				braceCount++
				started = true
			} else if content[i] == '}' {
				braceCount--
				if started && braceCount == 0 {
					endPos = offsetToPosition(content, i+1)
					break
				}
			}
		}
	}

	fullRange := Range{Start: startPos, End: endPos}
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
	code = strings.TrimPrefix(strings.TrimSpace(code), "func ")

	idx := strings.Index(code, "(")
	if idx == -1 {
		return "unknown"
	}

	name := strings.TrimSpace(code[:idx])

	// Handle methods: (receiver) Name
	if strings.HasPrefix(name, "(") {
		closeIdx := strings.Index(name, ")")
		if closeIdx != -1 {
			name = strings.TrimSpace(name[closeIdx+1:])
		}
	}

	return name
}

// nodeToSymbol converts an AST node to a DocumentSymbol if applicable.
func nodeToSymbol(node tuigen.Node) *DocumentSymbol {
	switch n := node.(type) {
	case *tuigen.LetBinding:
		fullRange := tuigenPosToRange(n.Position, len(n.Name)+5) // "@let " + name
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
					fullRange := tuigenPosToRange(n.Position, len(n.Tag)+2) // "<tag>"
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

// --- Workspace symbol provider ---

// workspaceSymbolProvider implements WorkspaceSymbolProvider.
type workspaceSymbolProvider struct {
	index ComponentIndex
}

// NewWorkspaceSymbolProvider creates a new workspace symbol provider.
func NewWorkspaceSymbolProvider(index ComponentIndex) WorkspaceSymbolProvider {
	return &workspaceSymbolProvider{index: index}
}

func (w *workspaceSymbolProvider) WorkspaceSymbols(query string) ([]SymbolInformation, error) {
	log.Server("WorkspaceSymbol provider: query=%q", query)

	q := strings.ToLower(query)
	var symbols []SymbolInformation

	// Search all indexed components
	for _, name := range w.index.All() {
		if q == "" || strings.Contains(strings.ToLower(name), q) {
			info, ok := w.index.Lookup(name)
			if ok && info != nil {
				symbols = append(symbols, SymbolInformation{
					Name:     name,
					Kind:     SymbolKindFunction,
					Location: info.Location,
				})
			}
		}
	}

	// Search all indexed functions
	for _, name := range w.index.AllFunctions() {
		if q == "" || strings.Contains(strings.ToLower(name), q) {
			info, ok := w.index.LookupFunc(name)
			if ok && info != nil {
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

// --- Range helpers ---

func findComponentEnd(comp *tuigen.Component, content string) Position {
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
				return offsetToPosition(content, i+1)
			}
		}
	}

	// Fallback
	return Position{Line: comp.Position.Line - 1 + 5, Character: 0}
}

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
	if positionBefore(selectionRange.Start, fullRange.Start) {
		selectionRange.Start = fullRange.Start
	}
	if positionAfter(selectionRange.End, fullRange.End) {
		selectionRange.End = fullRange.End
	}
	if positionAfter(selectionRange.Start, selectionRange.End) {
		selectionRange.Start = selectionRange.End
	}
	return selectionRange
}

// tuigenPosToRange converts a tuigen.Position (1-indexed) to a provider.Range (0-indexed).
func tuigenPosToRange(pos tuigen.Position, length int) Range {
	return Range{
		Start: Position{Line: pos.Line - 1, Character: pos.Column - 1},
		End:   Position{Line: pos.Line - 1, Character: pos.Column - 1 + length},
	}
}

// offsetToPosition converts a byte offset to a 0-indexed Position.
func offsetToPosition(content string, offset int) Position {
	line := 0
	col := 0
	for i := 0; i < offset && i < len(content); i++ {
		if content[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	return Position{Line: line, Character: col}
}
