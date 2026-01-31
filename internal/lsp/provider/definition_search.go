package provider

import (
	"strings"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

// --- Shared AST traversal helpers ---

// findLetBindingInNodes finds a @let binding by name in AST nodes.
func findLetBindingInNodes(nodes []tuigen.Node, name string) *tuigen.LetBinding {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.LetBinding:
			if n != nil && n.Name == name {
				return n
			}
		case *tuigen.Element:
			if n != nil {
				if found := findLetBindingInNodes(n.Children, name); found != nil {
					return found
				}
			}
		case *tuigen.ForLoop:
			if n != nil {
				if found := findLetBindingInNodes(n.Body, name); found != nil {
					return found
				}
			}
		case *tuigen.IfStmt:
			if n != nil {
				if found := findLetBindingInNodes(n.Then, name); found != nil {
					return found
				}
				if found := findLetBindingInNodes(n.Else, name); found != nil {
					return found
				}
			}
		case *tuigen.ComponentCall:
			if n != nil {
				if found := findLetBindingInNodes(n.Children, name); found != nil {
					return found
				}
			}
		}
	}
	return nil
}

// findForLoopWithVariable finds a for loop that declares the given variable.
func findForLoopWithVariable(nodes []tuigen.Node, varName string) *tuigen.ForLoop {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.ForLoop:
			if n != nil && (n.Index == varName || n.Value == varName) {
				return n
			}
			if n != nil {
				if found := findForLoopWithVariable(n.Body, varName); found != nil {
					return found
				}
			}
		case *tuigen.Element:
			if n != nil {
				if found := findForLoopWithVariable(n.Children, varName); found != nil {
					return found
				}
			}
		case *tuigen.IfStmt:
			if n != nil {
				if found := findForLoopWithVariable(n.Then, varName); found != nil {
					return found
				}
				if found := findForLoopWithVariable(n.Else, varName); found != nil {
					return found
				}
			}
		case *tuigen.ComponentCall:
			if n != nil {
				if found := findForLoopWithVariable(n.Children, varName); found != nil {
					return found
				}
			}
		}
	}
	return nil
}

// findGoCodeWithVariable finds a GoCode node that declares the given variable.
func findGoCodeWithVariable(nodes []tuigen.Node, varName string) *tuigen.GoCode {
	for _, node := range nodes {
		switch n := node.(type) {
		case *tuigen.GoCode:
			if n != nil && containsVarDecl(n.Code, varName) {
				return n
			}
		case *tuigen.Element:
			if n != nil {
				if found := findGoCodeWithVariable(n.Children, varName); found != nil {
					return found
				}
			}
		case *tuigen.ForLoop:
			if n != nil {
				if found := findGoCodeWithVariable(n.Body, varName); found != nil {
					return found
				}
			}
		case *tuigen.IfStmt:
			if n != nil {
				if found := findGoCodeWithVariable(n.Then, varName); found != nil {
					return found
				}
				if found := findGoCodeWithVariable(n.Else, varName); found != nil {
					return found
				}
			}
		case *tuigen.ComponentCall:
			if n != nil {
				if found := findGoCodeWithVariable(n.Children, varName); found != nil {
					return found
				}
			}
		case *tuigen.LetBinding:
			if n != nil && n.Element != nil {
				if found := findGoCodeWithVariable(n.Element.Children, varName); found != nil {
					return found
				}
			}
		}
	}
	return nil
}

// containsVarDecl checks if code declares the given variable.
// Uses exact identifier matching (not substring) for each declared name.
func containsVarDecl(code, varName string) bool {
	if idx := strings.Index(code, ":="); idx > 0 {
		lhs := code[:idx]
		parts := strings.Split(lhs, ",")
		for _, part := range parts {
			if strings.TrimSpace(part) == varName {
				return true
			}
		}
	}

	if strings.HasPrefix(strings.TrimSpace(code), "var ") {
		rest := strings.TrimPrefix(strings.TrimSpace(code), "var ")
		if idx := strings.Index(rest, "="); idx > 0 {
			lhs := rest[:idx]
			parts := strings.Split(lhs, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				fields := strings.Fields(part)
				if len(fields) > 0 && fields[0] == varName {
					return true
				}
			}
		}
	}

	return false
}

// isWordBoundary checks if position i in s is at a word boundary for the given word length.
func isWordBoundary(s string, i, wordLen int) bool {
	before := i == 0 || !IsWordChar(s[i-1])
	after := i+wordLen >= len(s) || !IsWordChar(s[i+wordLen])
	return before && after
}

// findVarDeclPosition finds the position of a variable declaration in code.
// Uses word-boundary-aware search so "count" doesn't match inside "accountCount".
func findVarDeclPosition(code, varName string) int {
	if idx := strings.Index(code, ":="); idx > 0 {
		lhs := code[:idx]
		parts := strings.Split(lhs, ",")
		pos := 0
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed == varName {
				partStart := strings.Index(lhs[pos:], trimmed)
				if partStart >= 0 {
					return pos + partStart
				}
			}
			pos += len(part) + 1
		}
	}

	if strings.HasPrefix(strings.TrimSpace(code), "var ") {
		varIdx := strings.Index(code, "var ")
		rest := code[varIdx+4:]
		if idx := strings.Index(rest, "="); idx > 0 {
			lhs := rest[:idx]
			// Use word-boundary search within the LHS
			offset := indexWholeWordIn(lhs, varName)
			if offset >= 0 {
				return varIdx + 4 + offset
			}
		}
	}

	return -1
}

// indexWholeWordIn finds the first whole-word occurrence of word in s.
func indexWholeWordIn(s, word string) int {
	idx := 0
	for {
		i := strings.Index(s[idx:], word)
		if i < 0 {
			return -1
		}
		absIdx := idx + i
		if isWordBoundary(s, absIdx, len(word)) {
			return absIdx
		}
		idx = absIdx + len(word)
	}
}

// parseFuncName extracts the function name from a Go function definition.
// Handles both plain functions ("func Name(") and methods ("func (r *T) Name(").
func parseFuncName(code string) string {
	idx := strings.Index(code, "func ")
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(code[idx+5:])

	// Skip receiver type: func (r *Receiver) Name(...)
	if len(rest) > 0 && rest[0] == '(' {
		closeIdx := strings.Index(rest, ")")
		if closeIdx < 0 {
			return ""
		}
		rest = strings.TrimSpace(rest[closeIdx+1:])
	}

	parenIdx := strings.Index(rest, "(")
	if parenIdx < 0 {
		return ""
	}
	return strings.TrimSpace(rest[:parenIdx])
}
