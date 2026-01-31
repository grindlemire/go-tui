package provider

import (
	"strings"

	"github.com/grindlemire/go-tui/internal/lsp/log"
)

// --- Completion types ---

// CompletionList represents a list of completion items.
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// CompletionItem represents a completion suggestion.
type CompletionItem struct {
	Label         string              `json:"label"`
	Kind          CompletionItemKind  `json:"kind,omitempty"`
	Detail        string              `json:"detail,omitempty"`
	Documentation *MarkupContent      `json:"documentation,omitempty"`
	InsertText    string              `json:"insertText,omitempty"`
	FilterText    string              `json:"filterText,omitempty"`
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

// CompletionContext contains additional completion context from the editor.
type CompletionContext struct {
	TriggerKind      int    `json:"triggerKind"`
	TriggerCharacter string `json:"triggerCharacter,omitempty"`
}

// CompletionProvider interface is defined in provider.go.

// completionProvider implements CompletionProvider.
type completionProvider struct {
	index        ComponentIndex
	goplsProxy   GoplsProxyAccessor
	virtualFiles VirtualFileAccessor
}

// NewCompletionProvider creates a new completion provider.
func NewCompletionProvider(index ComponentIndex, proxy GoplsProxyAccessor, vf VirtualFileAccessor) CompletionProvider {
	return &completionProvider{
		index:        index,
		goplsProxy:   proxy,
		virtualFiles: vf,
	}
}

func (c *completionProvider) Complete(ctx *CursorContext) (*CompletionList, error) {
	log.Server("Completion provider: NodeKind=%s, Word=%q, InGoExpr=%v, InClassAttr=%v, InElement=%v",
		ctx.NodeKind, ctx.Word, ctx.InGoExpr, ctx.InClassAttr, ctx.InElement)

	// Tailwind class completions inside class="..."
	if ctx.InClassAttr {
		prefix := classPrefix(ctx)
		log.Server("In class attribute with prefix: %q", prefix)
		items := c.getTailwindCompletions(prefix)
		return &CompletionList{Items: items}, nil
	}

	// State method completions (count. → Get, Set, Update, Bind, Batch)
	if ctx.InGoExpr {
		stateItems := c.getStateMethodCompletions(ctx)
		if len(stateItems) > 0 {
			return &CompletionList{Items: stateItems}, nil
		}
	}

	// Go expression completions via gopls
	if ctx.InGoExpr {
		items, err := c.getGoplsCompletions(ctx)
		if err != nil {
			log.Server("gopls completion error: %v", err)
		} else if len(items) > 0 {
			return &CompletionList{Items: items}, nil
		}
	}

	// Determine trigger character
	trigger := triggerChar(ctx)
	log.Server("Completion trigger: %q", trigger)

	var items []CompletionItem

	switch trigger {
	case "@":
		// Component call or DSL keyword
		items = append(items, c.getComponentCompletions()...)
		items = append(items, c.getDSLKeywordCompletions()...)
	case "<":
		// Element tag
		items = append(items, c.getElementCompletions()...)
	case "{":
		// Start of Go expression — try gopls
		goplsItems, err := c.getGoplsCompletions(ctx)
		if err == nil && len(goplsItems) > 0 {
			items = append(items, goplsItems...)
		}
	default:
		// Context-based completion
		items = append(items, c.getContextualCompletions(ctx)...)
	}

	return &CompletionList{Items: items}, nil
}

// --- Context helpers ---

// triggerChar returns the character that triggered completion.
func triggerChar(ctx *CursorContext) string {
	offset := PositionToOffset(ctx.Document.Content, ctx.Position)
	if offset <= 0 {
		return ""
	}
	return string(ctx.Document.Content[offset-1])
}

// maxClassAttrSearchDistance is the maximum number of bytes to search backwards
// when looking for a class="..." attribute opening. Must match the value in
// context.go's isOffsetInClassAttr for consistent behavior.
const maxClassAttrSearchDistance = 500

// classPrefix extracts the partial class name the user is typing inside class="...".
func classPrefix(ctx *CursorContext) string {
	offset := PositionToOffset(ctx.Document.Content, ctx.Position)
	content := ctx.Document.Content

	// Search backwards for class="
	searchStart := offset - maxClassAttrSearchDistance
	if searchStart < 0 {
		searchStart = 0
	}

	segment := content[searchStart:offset]
	classIdx := strings.LastIndex(segment, `class="`)
	if classIdx == -1 {
		return ""
	}

	// Extract text after class=" up to cursor
	attrValueStart := searchStart + classIdx + 7
	valueContent := content[attrValueStart:offset]

	// Find the last space to get the current partial class name
	lastSpace := strings.LastIndex(valueContent, " ")
	if lastSpace == -1 {
		return valueContent
	}
	return valueContent[lastSpace+1:]
}

