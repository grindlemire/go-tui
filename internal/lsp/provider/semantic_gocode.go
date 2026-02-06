package provider

import (
	"strings"

	"github.com/grindlemire/go-tui/internal/lsp/log"
	"github.com/grindlemire/go-tui/internal/tuigen"
)

// collectVariableTokensInCode finds and tokenizes variable references in Go code.
func (s *semanticTokensProvider) collectVariableTokensInCode(code string, pos tuigen.Position, paramNames map[string]bool, localVars map[string]bool, tokens *[]SemanticToken) {
	s.collectTokensInGoCode(code, pos, 1, paramNames, localVars, tokens)
}

// collectTokensInGoCodeDirect tokenizes Go code without any offset adjustment.
func (s *semanticTokensProvider) collectTokensInGoCodeDirect(code string, pos tuigen.Position, paramNames map[string]bool, localVars map[string]bool, tokens *[]SemanticToken) {
	s.collectTokensInGoCode(code, pos, 0, paramNames, localVars, tokens)
}

// collectTokensInGoCode tokenizes Go code for semantic highlighting.
func (s *semanticTokensProvider) collectTokensInGoCode(code string, pos tuigen.Position, startOffset int, paramNames map[string]bool, localVars map[string]bool, tokens *[]SemanticToken) {
	log.Server("collectTokensInGoCode: code=%q pos.Line=%d pos.Column=%d startOffset=%d", code, pos.Line, pos.Column, startOffset)

	// Handle multi-line code by splitting into lines and processing each separately.
	// Without this, all tokens would be placed on the first line.
	if strings.Contains(code, "\n") {
		lines := strings.Split(code, "\n")
		for lineIdx, line := range lines {
			linePos := tuigen.Position{
				Line:   pos.Line + lineIdx,
				Column: pos.Column,
			}
			lineOffset := startOffset
			if lineIdx > 0 {
				linePos.Column = 1
				lineOffset = 0
			}
			s.collectTokensInGoCode(line, linePos, lineOffset, paramNames, localVars, tokens)
		}
		return
	}

	i := 0
	for i < len(code) {
		ch := code[i]

		if ch == ' ' || ch == '\t' || ch == '\r' {
			i++
			continue
		}

		// Block comment /* ... */
		if ch == '/' && i+1 < len(code) && code[i+1] == '*' {
			start := i
			i += 2
			for i+1 < len(code) && !(code[i] == '*' && code[i+1] == '/') {
				i++
			}
			if i+1 < len(code) {
				i += 2
			}
			charPos := pos.Column - 1 + startOffset + start
			*tokens = append(*tokens, SemanticToken{
				Line:      pos.Line - 1,
				StartChar: charPos,
				Length:    i - start,
				TokenType: TokenTypeComment,
				Modifiers: 0,
			})
			continue
		}

		// Line comment // ...
		if ch == '/' && i+1 < len(code) && code[i+1] == '/' {
			start := i
			for i < len(code) && code[i] != '\n' {
				i++
			}
			charPos := pos.Column - 1 + startOffset + start
			*tokens = append(*tokens, SemanticToken{
				Line:      pos.Line - 1,
				StartChar: charPos,
				Length:    i - start,
				TokenType: TokenTypeComment,
				Modifiers: 0,
			})
			continue
		}

		// String literal (double-quoted)
		if ch == '"' {
			start := i
			i++
			for i < len(code) && code[i] != '"' {
				if code[i] == '\\' && i+1 < len(code) {
					i += 2
				} else {
					i++
				}
			}
			if i < len(code) {
				i++
			}
			stringContent := code[start:i]
			charPos := pos.Column - 1 + startOffset + start
			log.Server("Found string in GoCode: code=%q stringContent=%q start=%d pos.Column=%d startOffset=%d charPos=%d",
				code, stringContent, start, pos.Column, startOffset, charPos)
			s.emitStringWithFormatSpecifiers(stringContent, pos.Line-1, charPos, tokens)
			continue
		}

		// Backtick string
		if ch == '`' {
			start := i
			i++
			for i < len(code) && code[i] != '`' {
				i++
			}
			if i < len(code) {
				i++
			}
			stringContent := code[start:i]
			charPos := pos.Column - 1 + startOffset + start
			s.emitStringWithFormatSpecifiers(stringContent, pos.Line-1, charPos, tokens)
			continue
		}

		// Rune literal
		if ch == '\'' {
			start := i
			i++
			for i < len(code) && code[i] != '\'' {
				if code[i] == '\\' && i+1 < len(code) {
					i += 2
				} else {
					i++
				}
			}
			if i < len(code) {
				i++
			}
			charPos := pos.Column - 1 + startOffset + start
			*tokens = append(*tokens, SemanticToken{
				Line:      pos.Line - 1,
				StartChar: charPos,
				Length:    i - start,
				TokenType: TokenTypeString,
				Modifiers: 0,
			})
			continue
		}

		// Number literal
		if isDigit(ch) || (ch == '.' && i+1 < len(code) && isDigit(code[i+1])) {
			start := i
			if ch == '0' && i+1 < len(code) {
				next := code[i+1]
				if next == 'x' || next == 'X' {
					i += 2
					for i < len(code) && isHexDigit(code[i]) {
						i++
					}
				} else if next == 'b' || next == 'B' {
					i += 2
					for i < len(code) && (code[i] == '0' || code[i] == '1') {
						i++
					}
				} else if next == 'o' || next == 'O' {
					i += 2
					for i < len(code) && code[i] >= '0' && code[i] <= '7' {
						i++
					}
				} else {
					for i < len(code) {
						if isDigit(code[i]) || code[i] == '.' || code[i] == 'e' || code[i] == 'E' {
							i++
						} else if (code[i] == '+' || code[i] == '-') && i > 0 && (code[i-1] == 'e' || code[i-1] == 'E') {
							i++
						} else {
							break
						}
					}
				}
			} else {
				for i < len(code) {
					if isDigit(code[i]) || code[i] == '.' || code[i] == 'e' || code[i] == 'E' {
						i++
					} else if (code[i] == '+' || code[i] == '-') && i > 0 && (code[i-1] == 'e' || code[i-1] == 'E') {
						i++
					} else {
						break
					}
				}
			}
			charPos := pos.Column - 1 + startOffset + start
			*tokens = append(*tokens, SemanticToken{
				Line:      pos.Line - 1,
				StartChar: charPos,
				Length:    i - start,
				TokenType: TokenTypeNumber,
				Modifiers: 0,
			})
			continue
		}

		// Identifier or keyword
		if isWordStartChar(ch) {
			start := i
			for i < len(code) && isWordCharByte(code[i]) {
				i++
			}
			ident := code[start:i]
			charPos := pos.Column - 1 + startOffset + start

			// Boolean/nil literals
			if ident == "true" || ident == "false" || ident == "nil" {
				*tokens = append(*tokens, SemanticToken{
					Line:      pos.Line - 1,
					StartChar: charPos,
					Length:    len(ident),
					TokenType: TokenTypeNumber,
					Modifiers: 0,
				})
				continue
			}

			if paramNames[ident] {
				*tokens = append(*tokens, SemanticToken{
					Line:      pos.Line - 1,
					StartChar: charPos,
					Length:    len(ident),
					TokenType: TokenTypeParameter,
					Modifiers: 0,
				})
				continue
			}

			if localVars[ident] {
				*tokens = append(*tokens, SemanticToken{
					Line:      pos.Line - 1,
					StartChar: charPos,
					Length:    len(ident),
					TokenType: TokenTypeVariable,
					Modifiers: 0,
				})
				continue
			}

			// Check what comes after the identifier
			peekIdx := i
			for peekIdx < len(code) && (code[peekIdx] == ' ' || code[peekIdx] == '\t') {
				peekIdx++
			}

			// If followed by '.', this is a package/namespace
			if peekIdx < len(code) && code[peekIdx] == '.' {
				continue
			}

			// If followed by '(' or preceded by '.', this is a function call
			isPrecededByDot := start > 0 && code[start-1] == '.'
			isFollowedByParen := peekIdx < len(code) && code[peekIdx] == '('
			if isPrecededByDot || isFollowedByParen {
				*tokens = append(*tokens, SemanticToken{
					Line:      pos.Line - 1,
					StartChar: charPos,
					Length:    len(ident),
					TokenType: TokenTypeFunction,
					Modifiers: 0,
				})
				continue
			}

			// Check if it's a known function
			if s.fnChecker != nil && s.fnChecker.IsFunctionName(ident) {
				*tokens = append(*tokens, SemanticToken{
					Line:      pos.Line - 1,
					StartChar: charPos,
					Length:    len(ident),
					TokenType: TokenTypeFunction,
					Modifiers: 0,
				})
				continue
			}

			continue
		}

		// := operator
		if ch == ':' && i+1 < len(code) && code[i+1] == '=' {
			charPos := pos.Column - 1 + startOffset + i
			*tokens = append(*tokens, SemanticToken{
				Line:      pos.Line - 1,
				StartChar: charPos,
				Length:    2,
				TokenType: TokenTypeOperator,
				Modifiers: 0,
			})
			i += 2
			continue
		}

		i++
	}
}

// collectTokensFromFuncBody tokenizes the body of a Go function.
func (s *semanticTokensProvider) collectTokensFromFuncBody(code string, pos tuigen.Position, paramNames map[string]bool, tokens *[]SemanticToken) {
	braceIdx := strings.Index(code, "{")
	if braceIdx == -1 {
		return
	}

	lines := strings.Split(code, "\n")
	if len(lines) == 0 {
		return
	}

	localVars := make(map[string]bool)

	charCount := 0
	bodyStartLine := 0
	for lineIdx, line := range lines {
		lineEnd := charCount + len(line)
		if braceIdx >= charCount && braceIdx < lineEnd+1 {
			bodyStartLine = lineIdx
			break
		}
		charCount = lineEnd + 1
	}

	for lineIdx := bodyStartLine; lineIdx < len(lines); lineIdx++ {
		line := lines[lineIdx]
		trimmed := strings.TrimSpace(line)
		if trimmed == "{" || trimmed == "}" {
			continue
		}

		docLine := pos.Line + lineIdx
		lineStartCol := 1
		if lineIdx == bodyStartLine {
			braceInLine := strings.Index(line, "{")
			if braceInLine != -1 {
				line = line[braceInLine+1:]
				lineStartCol = pos.Column + braceInLine + 1
			}
		}

		varDecls := extractVarDeclarationsWithPositions(line)
		for _, decl := range varDecls {
			localVars[decl.name] = true
			*tokens = append(*tokens, SemanticToken{
				Line:      docLine - 1,
				StartChar: lineStartCol - 1 + decl.offset,
				Length:    len(decl.name),
				TokenType: TokenTypeVariable,
				Modifiers: TokenModDeclaration,
			})
		}

		linePos := tuigen.Position{Line: docLine, Column: lineStartCol}
		s.collectTokensInGoCode(line, linePos, 0, paramNames, localVars, tokens)
	}
}

// varDecl represents a variable declaration with its position.
type varDecl struct {
	name   string
	offset int
}

// extractVarDeclarationsWithPositions extracts variable names and their positions from Go code.
func extractVarDeclarationsWithPositions(code string) []varDecl {
	var decls []varDecl

	if idx := strings.Index(code, ":="); idx > 0 {
		lhs := code[:idx]
		parts := strings.Split(lhs, ",")
		pos := 0
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" && trimmed != "_" && isValidIdentifier(trimmed) {
				identStart := strings.Index(lhs[pos:], trimmed)
				if identStart >= 0 {
					decls = append(decls, varDecl{
						name:   trimmed,
						offset: pos + identStart,
					})
				}
			}
			pos += len(part) + 1
		}
		return decls
	}

	trimmed := strings.TrimSpace(code)
	if strings.HasPrefix(trimmed, "var ") {
		varIdx := strings.Index(code, "var ")
		rest := code[varIdx+4:]
		if eqIdx := strings.Index(rest, "="); eqIdx > 0 {
			lhs := rest[:eqIdx]
			parts := strings.Split(lhs, ",")
			pos := 0
			for _, part := range parts {
				part = strings.TrimSpace(part)
				fields := strings.Fields(part)
				if len(fields) > 0 {
					name := fields[0]
					if name != "_" && isValidIdentifier(name) {
						identStart := strings.Index(lhs[pos:], name)
						if identStart >= 0 {
							decls = append(decls, varDecl{
								name:   name,
								offset: varIdx + 4 + pos + identStart,
							})
						}
					}
				}
				pos += len(part) + 1
			}
		}
	}

	return decls
}

// funcParam represents a function parameter for semantic tokenization.
type funcParam struct {
	Name string
	Type string
}

// parseFuncSignatureForTokens extracts function name, receiver, params, and return type from code.
// For methods like "func (s *Type) Name(...) RetType { ... }", receiver will be "s *Type".
// For plain functions, receiver will be "".
func parseFuncSignatureForTokens(code string) (name, receiver string, params []funcParam, returns string) {
	code = strings.TrimSpace(code)
	if !strings.HasPrefix(code, "func ") {
		return "", "", nil, ""
	}
	rest := code[5:] // skip "func "

	// Check for method receiver: starts with '('
	if len(rest) > 0 && rest[0] == '(' {
		// Find matching close paren for receiver
		depth := 0
		closeIdx := -1
		for i, c := range rest {
			if c == '(' {
				depth++
			} else if c == ')' {
				depth--
				if depth == 0 {
					closeIdx = i
					break
				}
			}
		}
		if closeIdx == -1 {
			return "", "", nil, ""
		}
		receiver = strings.TrimSpace(rest[1:closeIdx])
		rest = strings.TrimSpace(rest[closeIdx+1:])
	}

	parenIdx := strings.Index(rest, "(")
	if parenIdx == -1 {
		return "", "", nil, ""
	}
	name = strings.TrimSpace(rest[:parenIdx])
	rest = rest[parenIdx:]

	// Find matching close paren for params
	depth := 0
	closeIdx := -1
	for i, c := range rest {
		if c == '(' {
			depth++
		} else if c == ')' {
			depth--
			if depth == 0 {
				closeIdx = i
				break
			}
		}
	}
	if closeIdx == -1 {
		return name, receiver, nil, ""
	}

	paramStr := rest[1:closeIdx]

	// Parse parameters
	if paramStr != "" {
		for _, p := range strings.Split(paramStr, ",") {
			p = strings.TrimSpace(p)
			parts := strings.SplitN(p, " ", 2)
			if len(parts) == 2 {
				params = append(params, funcParam{Name: parts[0], Type: parts[1]})
			}
		}
	}

	// Return type
	after := strings.TrimSpace(rest[closeIdx+1:])
	if braceIdx := strings.Index(after, "{"); braceIdx > 0 {
		returns = strings.TrimSpace(after[:braceIdx])
	}

	return name, receiver, params, returns
}
