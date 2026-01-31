package provider

import (
	"strings"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

// --- Workspace search ---

func (r *referencesProvider) searchWorkspaceForComponentRefs(name string, refs *[]Location) {
	for uri, ast := range r.workspace.AllWorkspaceASTs() {
		// Skip if file is already open (already searched above)
		if r.docs.GetDocument(uri) != nil {
			continue
		}
		if ast == nil {
			continue
		}
		for _, comp := range ast.Components {
			findComponentCallsInNodes(comp.Body, name, uri, refs)
		}
	}
}

func (r *referencesProvider) searchWorkspaceForFunctionRefs(name string, refs *[]Location) {
	for uri, ast := range r.workspace.AllWorkspaceASTs() {
		// Skip if file is already open (already searched above)
		if r.docs.GetDocument(uri) != nil {
			continue
		}
		if ast == nil {
			continue
		}
		for _, comp := range ast.Components {
			findFunctionCallsInNodes(comp.Body, name, uri, refs)
		}
	}
}

// --- AST traversal helpers ---

// findComponentCallsInNodes finds component calls recursively.
func findComponentCallsInNodes(nodes []tuigen.Node, name string, uri string, refs *[]Location) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.ComponentCall:
			if n != nil && n.Name == name {
				*refs = append(*refs, Location{
					URI: uri,
					Range: Range{
						Start: Position{Line: n.Position.Line - 1, Character: n.Position.Column - 1},
						End:   Position{Line: n.Position.Line - 1, Character: n.Position.Column - 1 + len("@") + len(name)},
					},
				})
			}
			if n != nil {
				findComponentCallsInNodes(n.Children, name, uri, refs)
			}
		case *tuigen.Element:
			if n != nil {
				findComponentCallsInNodes(n.Children, name, uri, refs)
			}
		case *tuigen.ForLoop:
			if n != nil {
				findComponentCallsInNodes(n.Body, name, uri, refs)
			}
		case *tuigen.IfStmt:
			if n != nil {
				findComponentCallsInNodes(n.Then, name, uri, refs)
				findComponentCallsInNodes(n.Else, name, uri, refs)
			}
		case *tuigen.LetBinding:
			if n != nil && n.Element != nil {
				findComponentCallsInNodes(n.Element.Children, name, uri, refs)
			}
		}
	}
}

// findFunctionCallsInNodes finds function calls in Go expressions.
func findFunctionCallsInNodes(nodes []tuigen.Node, name string, uri string, refs *[]Location) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.GoCode:
			if n != nil {
				findFuncCallInCode(n.Code, name, n.Position.Line-1, n.Position.Column-1, uri, refs)
			}
		case *tuigen.GoExpr:
			if n != nil {
				findFuncCallInCode(n.Code, name, n.Position.Line-1, n.Position.Column, uri, refs)
			}
		case *tuigen.Element:
			if n != nil {
				for _, attr := range n.Attributes {
					if expr, ok := attr.Value.(*tuigen.GoExpr); ok && expr != nil {
						findFuncCallInCode(expr.Code, name, expr.Position.Line-1, expr.Position.Column, uri, refs)
					}
				}
				findFunctionCallsInNodes(n.Children, name, uri, refs)
			}
		case *tuigen.ForLoop:
			if n != nil {
				findFunctionCallsInNodes(n.Body, name, uri, refs)
			}
		case *tuigen.IfStmt:
			if n != nil {
				findFunctionCallsInNodes(n.Then, name, uri, refs)
				findFunctionCallsInNodes(n.Else, name, uri, refs)
			}
		case *tuigen.ComponentCall:
			if n != nil {
				colOffset := n.Position.Column + len("@") + len(n.Name) + 1
				findFuncCallInCode(n.Args, name, n.Position.Line-1, colOffset, uri, refs)
				findFunctionCallsInNodes(n.Children, name, uri, refs)
			}
		case *tuigen.LetBinding:
			if n != nil && n.Element != nil {
				findFunctionCallsInNodes(n.Element.Children, name, uri, refs)
			}
		}
	}
}

// findFuncCallInCode finds all occurrences of name+"(" in code with word-boundary checks.
func findFuncCallInCode(code, name string, line, colBase int, uri string, refs *[]Location) {
	searchTarget := name + "("
	idx := 0
	for {
		i := strings.Index(code[idx:], searchTarget)
		if i < 0 {
			break
		}
		absIdx := idx + i

		// Check word boundary before the match
		before := absIdx == 0 || !IsWordChar(code[absIdx-1])
		if before {
			*refs = append(*refs, Location{
				URI: uri,
				Range: Range{
					Start: Position{Line: line, Character: colBase + absIdx},
					End:   Position{Line: line, Character: colBase + absIdx + len(name)},
				},
			})
		}

		idx = absIdx + len(searchTarget)
	}
}

// findVariableUsagesInNodes finds usages of a variable in AST nodes.
func findVariableUsagesInNodes(nodes []tuigen.Node, varName string, uri string, refs *[]Location) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.GoExpr:
			if n != nil {
				findVariableInCode(n.Code, varName, n.Position, 1, uri, refs)
			}
		case *tuigen.RawGoExpr:
			if n != nil {
				findVariableInCode(n.Code, varName, n.Position, 1, uri, refs)
			}
		case *tuigen.GoCode:
			if n != nil {
				findVariableInCode(n.Code, varName, n.Position, 0, uri, refs)
			}
		case *tuigen.Element:
			if n != nil {
				for _, attr := range n.Attributes {
					if expr, ok := attr.Value.(*tuigen.GoExpr); ok && expr != nil {
						findVariableInCode(expr.Code, varName, expr.Position, 1, uri, refs)
					}
				}
				findVariableUsagesInNodes(n.Children, varName, uri, refs)
			}
		case *tuigen.ForLoop:
			if n != nil {
				iterableOffset := len("@for ") + len(n.Index) + len(", ") + len(n.Value) + len(" := range ")
				findVariableInCode(n.Iterable, varName, n.Position, iterableOffset, uri, refs)
				findVariableUsagesInNodes(n.Body, varName, uri, refs)
			}
		case *tuigen.IfStmt:
			if n != nil {
				findVariableInCode(n.Condition, varName, n.Position, len("@if "), uri, refs)
				findVariableUsagesInNodes(n.Then, varName, uri, refs)
				findVariableUsagesInNodes(n.Else, varName, uri, refs)
			}
		case *tuigen.ComponentCall:
			if n != nil {
				argsOffset := len("@") + len(n.Name) + 1
				findVariableInCode(n.Args, varName, n.Position, argsOffset, uri, refs)
				findVariableUsagesInNodes(n.Children, varName, uri, refs)
			}
		case *tuigen.LetBinding:
			if n != nil && n.Element != nil {
				findVariableUsagesInNodes(n.Element.Children, varName, uri, refs)
			}
		}
	}
}

// findVariableUsagesInNodesExcluding finds usages excluding a specific location.
func findVariableUsagesInNodesExcluding(nodes []tuigen.Node, varName string, uri string, exclLine, exclCharStart, exclCharEnd int, refs *[]Location) {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.GoExpr:
			if n != nil {
				findVariableInCodeExcluding(n.Code, varName, n.Position, 1, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.RawGoExpr:
			if n != nil {
				findVariableInCodeExcluding(n.Code, varName, n.Position, 1, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.GoCode:
			if n != nil {
				findVariableInCodeExcluding(n.Code, varName, n.Position, 0, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.Element:
			if n != nil {
				for _, attr := range n.Attributes {
					if expr, ok := attr.Value.(*tuigen.GoExpr); ok && expr != nil {
						findVariableInCodeExcluding(expr.Code, varName, expr.Position, 1, uri, exclLine, exclCharStart, exclCharEnd, refs)
					}
				}
				findVariableUsagesInNodesExcluding(n.Children, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.ForLoop:
			if n != nil {
				iterableOffset := len("@for ") + len(n.Index) + len(", ") + len(n.Value) + len(" := range ")
				findVariableInCodeExcluding(n.Iterable, varName, n.Position, iterableOffset, uri, exclLine, exclCharStart, exclCharEnd, refs)
				findVariableUsagesInNodesExcluding(n.Body, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.IfStmt:
			if n != nil {
				findVariableInCodeExcluding(n.Condition, varName, n.Position, len("@if "), uri, exclLine, exclCharStart, exclCharEnd, refs)
				findVariableUsagesInNodesExcluding(n.Then, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
				findVariableUsagesInNodesExcluding(n.Else, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.ComponentCall:
			if n != nil {
				argsOffset := len("@") + len(n.Name) + 1
				findVariableInCodeExcluding(n.Args, varName, n.Position, argsOffset, uri, exclLine, exclCharStart, exclCharEnd, refs)
				findVariableUsagesInNodesExcluding(n.Children, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		case *tuigen.LetBinding:
			if n != nil && n.Element != nil {
				findVariableUsagesInNodesExcluding(n.Element.Children, varName, uri, exclLine, exclCharStart, exclCharEnd, refs)
			}
		}
	}
}

// findVariableInCode finds variable occurrences with a custom offset.
func findVariableInCode(code, varName string, pos tuigen.Position, startOffset int, uri string, refs *[]Location) {
	findVariableInCodeExcluding(code, varName, pos, startOffset, uri, -1, -1, -1, refs)
}

// findVariableInCodeExcluding finds variable occurrences, excluding a specific location.
// Handles multi-line code blocks by splitting on newlines and tracking line offsets.
func findVariableInCodeExcluding(code, varName string, pos tuigen.Position, startOffset int, uri string, exclLine, exclCharStart, exclCharEnd int, refs *[]Location) {
	lines := strings.Split(code, "\n")
	for lineIdx, lineContent := range lines {
		idx := 0
		for {
			i := strings.Index(lineContent[idx:], varName)
			if i < 0 {
				break
			}
			absIdx := idx + i

			before := absIdx == 0 || !IsWordChar(lineContent[absIdx-1])
			after := absIdx+len(varName) >= len(lineContent) || !IsWordChar(lineContent[absIdx+len(varName)])

			if before && after {
				line := pos.Line - 1 + lineIdx
				charPos := absIdx
				if lineIdx == 0 {
					charPos = pos.Column - 1 + startOffset + absIdx
				}

				if line == exclLine && charPos == exclCharStart && charPos+len(varName) == exclCharEnd {
					idx = absIdx + len(varName)
					continue
				}

				*refs = append(*refs, Location{
					URI: uri,
					Range: Range{
						Start: Position{Line: line, Character: charPos},
						End:   Position{Line: line, Character: charPos + len(varName)},
					},
				})
			}

			idx = absIdx + len(varName)
		}
	}
}

// indexWholeWord finds the first occurrence of word in s with word boundary checks.
// Returns -1 if not found as a whole word.
func indexWholeWord(s, word string) int {
	idx := 0
	for {
		i := strings.Index(s[idx:], word)
		if i < 0 {
			return -1
		}
		absIdx := idx + i
		before := absIdx == 0 || !IsWordChar(s[absIdx-1])
		after := absIdx+len(word) >= len(s) || !IsWordChar(s[absIdx+len(word)])
		if before && after {
			return absIdx
		}
		idx = absIdx + len(word)
	}
}
