package provider

import (
	"regexp"
	"sort"

	"github.com/grindlemire/go-tui/internal/lsp/log"
)

// formatSpecifierRegex matches Go format specifiers like %s, %d, %v, %.2f, %#x, etc.
var formatSpecifierRegex = regexp.MustCompile(`%[-+# 0]*(\*|\d+)?(\.\*|\.\d+)?[vTtbcdoqxXUeEfFgGsp%]`)

// stateNewStateRegex matches state declarations like: count := tui.NewState(0)
var stateNewStateRegex = regexp.MustCompile(`(\w+)\s*:=\s*tui\.NewState\((.+)\)`)

// Semantic token types (must match the order in SemanticTokensLegend.TokenTypes).
const (
	TokenTypeNamespace = 0  // package
	TokenTypeType      = 1  // types
	TokenTypeClass     = 2  // components
	TokenTypeFunction  = 3  // functions
	TokenTypeParameter = 4  // parameters
	TokenTypeVariable  = 5  // variables
	TokenTypeProperty  = 6  // attributes
	TokenTypeKeyword   = 7  // keywords
	TokenTypeString    = 8  // strings
	TokenTypeNumber    = 9  // numbers
	TokenTypeOperator  = 10 // operators
	TokenTypeDecorator = 11 // @ prefix
	TokenTypeRegexp    = 12 // format specifiers (often purple)
	TokenTypeComment   = 13 // comments
	TokenTypeLabel         = 14 // named refs (#Name)
	TokenTypeTypeParameter = 15 // generic type arguments
)

// Semantic token modifiers (bit flags).
const (
	TokenModDeclaration  = 1 << 0 // where defined
	TokenModDefinition   = 1 << 1 // where defined
	TokenModReadonly     = 1 << 2 // const/let
	TokenModModification = 1 << 3 // where modified
)

// SemanticTokens represents the result of a semantic tokens request.
type SemanticTokens struct {
	Data []int `json:"data"`
}

// SemanticToken represents a single semantic token before encoding.
type SemanticToken struct {
	Line      int
	StartChar int
	Length    int
	TokenType int
	Modifiers int
}

// FunctionNameChecker checks if an identifier is a known function.
type FunctionNameChecker interface {
	IsFunctionName(name string) bool
}

// semanticTokensProvider implements SemanticTokensProvider.
type semanticTokensProvider struct {
	fnChecker  FunctionNameChecker
	docs       DocumentAccessor // optional, for accurate position lookups
	currentURI string          // set during SemanticTokensFull call
}

// NewSemanticTokensProvider creates a new semantic tokens provider.
func NewSemanticTokensProvider(fnChecker FunctionNameChecker, docs DocumentAccessor) SemanticTokensProvider {
	return &semanticTokensProvider{fnChecker: fnChecker, docs: docs}
}

func (s *semanticTokensProvider) SemanticTokensFull(doc *Document) (*SemanticTokens, error) {
	s.currentURI = doc.URI
	log.Server("=== SemanticTokens provider for %s ===", doc.URI)

	if doc.AST == nil {
		return &SemanticTokens{Data: []int{}}, nil
	}

	tokens := s.collectSemanticTokens(doc)

	// Sort tokens by position
	sort.Slice(tokens, func(i, j int) bool {
		if tokens[i].Line != tokens[j].Line {
			return tokens[i].Line < tokens[j].Line
		}
		return tokens[i].StartChar < tokens[j].StartChar
	})

	// Encode tokens into delta format
	data := EncodeSemanticTokens(tokens)

	if len(data) > 0 {
		log.Server("Encoded %d tokens (%d ints). First 25 values: %v", len(tokens), len(data), data[:min(25, len(data))])
	}

	return &SemanticTokens{Data: data}, nil
}

// collectSemanticTokens collects all semantic tokens from a document.
func (s *semanticTokensProvider) collectSemanticTokens(doc *Document) []SemanticToken {
	var tokens []SemanticToken
	ast := doc.AST

	// Collect comment tokens
	s.collectAllCommentTokens(ast, &tokens)

	// Collect component-related tokens
	for _, comp := range ast.Components {
		// Component keyword (templ)
		tokens = append(tokens, SemanticToken{
			Line:      comp.Position.Line - 1,
			StartChar: comp.Position.Column - 1,
			Length:    len("templ"),
			TokenType: TokenTypeKeyword,
			Modifiers: 0,
		})

		// Component name (declaration)
		nameStart := comp.Position.Column - 1 + len("templ ")
		tokens = append(tokens, SemanticToken{
			Line:      comp.Position.Line - 1,
			StartChar: nameStart,
			Length:    len(comp.Name),
			TokenType: TokenTypeClass,
			Modifiers: TokenModDeclaration | TokenModDefinition,
		})

		// Parameters (declarations) with type coloring
		for _, param := range comp.Params {
			tokens = append(tokens, SemanticToken{
				Line:      param.Position.Line - 1,
				StartChar: param.Position.Column - 1,
				Length:    len(param.Name),
				TokenType: TokenTypeParameter,
				Modifiers: TokenModDeclaration,
			})
			// Parameter type
			typeStart := param.Position.Column - 1 + len(param.Name) + 1 // +1 for space
			emitGoTypeTokens(param.Type, param.Position.Line-1, typeStart, &tokens)
		}

		// Collect tokens from body
		s.collectTokensFromNodes(comp.Body, comp.Params, &tokens)
	}

	// Collect function-related tokens
	for _, fn := range ast.Funcs {
		// func keyword
		tokens = append(tokens, SemanticToken{
			Line:      fn.Position.Line - 1,
			StartChar: fn.Position.Column - 1,
			Length:    len("func"),
			TokenType: TokenTypeKeyword,
			Modifiers: 0,
		})

		// Function name
		name, _, params, returns := parseFuncSignatureForTokens(fn.Code)
		if name != "" {
			line := fn.Position.Line - 1
			nameStart := fn.Position.Column - 1 + len("func ")
			tokens = append(tokens, SemanticToken{
				Line:      line,
				StartChar: nameStart,
				Length:    len(name),
				TokenType: TokenTypeFunction,
				Modifiers: TokenModDeclaration | TokenModDefinition,
			})

			// Function parameters — names and types
			paramStart := nameStart + len(name) + 1 // +1 for '('
			for _, p := range params {
				// Parameter name
				tokens = append(tokens, SemanticToken{
					Line:      line,
					StartChar: paramStart,
					Length:    len(p.Name),
					TokenType: TokenTypeParameter,
					Modifiers: TokenModDeclaration,
				})
				// Parameter type
				typeStart := paramStart + len(p.Name) + 1 // +1 for space
				emitGoTypeTokens(p.Type, line, typeStart, &tokens)
				paramStart += len(p.Name) + 1 + len(p.Type) + 2 // +2 for ", "
			}

			// Return type
			if returns != "" {
				// Return type starts after closing paren + space
				// paramStart is past the last param; it's at the position after "lastType, "
				// but we need position after ')'
				returnStart := nameStart + len(name) + 1 // "name("
				if len(params) > 0 {
					// Calculate total param string length
					for i, p := range params {
						returnStart += len(p.Name) + 1 + len(p.Type)
						if i < len(params)-1 {
							returnStart += 2 // ", "
						}
					}
				}
				returnStart += 2 // ") " — close paren + space
				emitGoTypeTokens(returns, line, returnStart, &tokens)
			}

			// Build parameter names map for body tokenization
			paramNames := make(map[string]bool)
			for _, p := range params {
				paramNames[p.Name] = true
			}

			// Tokenize function body
			s.collectTokensFromFuncBody(fn.Code, fn.Position, paramNames, &tokens)
		}
	}

	return tokens
}

// --- Helper functions ---

// EncodeSemanticTokens encodes tokens into the LSP delta format.
func EncodeSemanticTokens(tokens []SemanticToken) []int {
	if len(tokens) == 0 {
		return []int{}
	}
	data := make([]int, 0, len(tokens)*5)
	prevLine := 0
	prevChar := 0
	for _, t := range tokens {
		deltaLine := t.Line - prevLine
		deltaChar := t.StartChar
		if deltaLine == 0 {
			deltaChar = t.StartChar - prevChar
		}
		data = append(data, deltaLine, deltaChar, t.Length, t.TokenType, t.Modifiers)
		prevLine = t.Line
		prevChar = t.StartChar
	}
	return data
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isHexDigit(c byte) bool {
	return isDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func isWordStartChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

// isWordCharByte checks if a byte is a valid word character (for identifiers).
func isWordCharByte(c byte) bool {
	return isWordStartChar(c) || isDigit(c)
}

// isValidIdentifier checks if a string is a valid Go identifier.
func isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, c := range s {
		if i == 0 {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_') {
				return false
			}
		} else {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
				return false
			}
		}
	}
	return true
}
