package lsp

import (
	"sort"
	"strings"

	"github.com/grindlemire/go-tui/internal/lsp/provider"
	"github.com/grindlemire/go-tui/internal/tuigen"
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

// CompletionList, CompletionItem, and CompletionItemKind are type aliases for
// the canonical definitions in the provider package, eliminating duplicate type
// definitions.
type CompletionList = provider.CompletionList
type CompletionItem = provider.CompletionItem
type CompletionItemKind = provider.CompletionItemKind

// Re-export CompletionItemKind constants so existing lsp package code compiles unchanged.
const (
	CompletionItemKindText          = provider.CompletionItemKindText
	CompletionItemKindMethod        = provider.CompletionItemKindMethod
	CompletionItemKindFunction      = provider.CompletionItemKindFunction
	CompletionItemKindConstructor   = provider.CompletionItemKindConstructor
	CompletionItemKindField         = provider.CompletionItemKindField
	CompletionItemKindVariable      = provider.CompletionItemKindVariable
	CompletionItemKindClass         = provider.CompletionItemKindClass
	CompletionItemKindInterface     = provider.CompletionItemKindInterface
	CompletionItemKindModule        = provider.CompletionItemKindModule
	CompletionItemKindProperty      = provider.CompletionItemKindProperty
	CompletionItemKindUnit          = provider.CompletionItemKindUnit
	CompletionItemKindValue         = provider.CompletionItemKindValue
	CompletionItemKindEnum          = provider.CompletionItemKindEnum
	CompletionItemKindKeyword       = provider.CompletionItemKindKeyword
	CompletionItemKindSnippet       = provider.CompletionItemKindSnippet
	CompletionItemKindColor         = provider.CompletionItemKindColor
	CompletionItemKindFile          = provider.CompletionItemKindFile
	CompletionItemKindReference     = provider.CompletionItemKindReference
	CompletionItemKindFolder        = provider.CompletionItemKindFolder
	CompletionItemKindEnumMember    = provider.CompletionItemKindEnumMember
	CompletionItemKindConstant      = provider.CompletionItemKindConstant
	CompletionItemKindStruct        = provider.CompletionItemKindStruct
	CompletionItemKindEvent         = provider.CompletionItemKindEvent
	CompletionItemKindOperator      = provider.CompletionItemKindOperator
	CompletionItemKindTypeParameter = provider.CompletionItemKindTypeParameter
)

// isInClassAttribute checks if the cursor is inside a class="" attribute value.
// Returns (isInClass, partialPrefix) where partialPrefix is what the user has typed
// for the current class name (text after the last space before cursor).
// Retained for use by features_test.go; the provider package has its own implementation.
func (s *Server) isInClassAttribute(doc *Document, pos Position) (bool, string) {
	offset := PositionToOffset(doc.Content, pos)
	if offset <= 0 {
		return false, ""
	}

	content := doc.Content

	// Search backwards for class=" or class='
	// We need to find the start of the class attribute value
	classAttrStart := -1
	quoteChar := byte(0)

	for i := offset - 1; i >= 0; i-- {
		// If we hit a newline, stop searching (class attribute should be on same line)
		if content[i] == '\n' {
			break
		}

		// Check if we're hitting a quote that starts the class value
		if content[i] == '"' || content[i] == '\'' {
			// Check if this is the opening quote of class="..."
			// Look backwards for "class="
			if i >= 6 {
				before := string(content[i-6 : i+1])
				if before == `class="` || before == `class='` {
					classAttrStart = i + 1 // Position right after the quote
					quoteChar = content[i]
					break
				}
			}
			// If we hit a quote that's not a class attribute opening, check if it's closing
			// by looking for the matching opening class=" before it
			// This handles the case where cursor is past the class attribute
			break
		}
	}

	if classAttrStart == -1 {
		return false, ""
	}

	// Now check if we're still inside the attribute (before the closing quote)
	for i := offset; i < len(content); i++ {
		if content[i] == quoteChar {
			// Found closing quote - we are inside the class attribute
			break
		}
		if content[i] == '\n' {
			// Newline before closing quote - malformed, but still inside
			break
		}
	}

	// Extract the partial prefix (text after last space before cursor)
	valueContent := string(content[classAttrStart:offset])
	prefix := ""

	// Find the last space to get the current partial class name
	lastSpace := strings.LastIndex(valueContent, " ")
	if lastSpace == -1 {
		prefix = valueContent
	} else {
		prefix = valueContent[lastSpace+1:]
	}

	return true, prefix
}

// getTailwindCompletions returns completion items for Tailwind classes.
// Retained for use by features_test.go; the provider package has its own implementation.
func (s *Server) getTailwindCompletions(prefix string) []CompletionItem {
	allClasses := tuigen.AllTailwindClasses()

	var items []CompletionItem
	for _, classInfo := range allClasses {
		// Filter by prefix if provided
		if prefix != "" && !strings.HasPrefix(classInfo.Name, prefix) {
			continue
		}

		// Build documentation with example
		docValue := classInfo.Description
		if classInfo.Example != "" {
			docValue += "\n\n**Example:**\n```tui\n" + classInfo.Example + "\n```"
		}

		items = append(items, CompletionItem{
			Label:      classInfo.Name,
			Kind:       CompletionItemKindConstant,
			Detail:     classInfo.Category,
			InsertText: classInfo.Name,
			FilterText: classInfo.Name,
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: docValue,
			},
		})
	}

	// Sort by category priority, then alphabetically within category
	sortLegacyCompletionsByCategory(items)

	return items
}

// sortLegacyCompletionsByCategory sorts completion items by category priority then name.
// Retained for use by getTailwindCompletions above (used by features_test.go).
func sortLegacyCompletionsByCategory(items []CompletionItem) {
	categoryOrder := map[string]int{
		"layout":     1,
		"flex":       2,
		"spacing":    3,
		"typography": 4,
		"visual":     5,
	}

	sort.Slice(items, func(i, j int) bool {
		orderI := categoryOrder[items[i].Detail]
		orderJ := categoryOrder[items[j].Detail]
		if orderI == 0 {
			orderI = 100
		}
		if orderJ == 0 {
			orderJ = 100
		}
		if orderI != orderJ {
			return orderI < orderJ
		}
		return items[i].Label < items[j].Label
	})
}
