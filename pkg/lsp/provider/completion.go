package provider

import (
	"fmt"
	"sort"
	"strings"

	"github.com/grindlemire/go-tui/pkg/lsp/gopls"
	"github.com/grindlemire/go-tui/pkg/lsp/log"
	"github.com/grindlemire/go-tui/pkg/lsp/schema"
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
	if ctx.InGoExpr && ctx.Scope != nil {
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

// classPrefix extracts the partial class name the user is typing inside class="...".
func classPrefix(ctx *CursorContext) string {
	offset := PositionToOffset(ctx.Document.Content, ctx.Position)
	content := ctx.Document.Content

	// Search backwards for class="
	searchStart := offset - 100
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

// --- Completion generators ---

func (c *completionProvider) getComponentCompletions() []CompletionItem {
	var items []CompletionItem
	for _, name := range c.index.All() {
		info, ok := c.index.Lookup(name)
		if !ok || info == nil {
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

func (c *completionProvider) getDSLKeywordCompletions() []CompletionItem {
	return []CompletionItem{
		{
			Label:      "for",
			Kind:       CompletionItemKindKeyword,
			Detail:     "Loop over items",
			InsertText: "for ${1:i}, ${2:item} := range ${3:items} {\n\t$0\n}",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "Loop over a collection.\n\n```gsx\n@for i, item := range items {\n    <span>{item}</span>\n}\n```",
			},
		},
		{
			Label:      "if",
			Kind:       CompletionItemKindKeyword,
			Detail:     "Conditional rendering",
			InsertText: "if ${1:condition} {\n\t$0\n}",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "Conditionally render content.\n\n```gsx\n@if showHeader {\n    <span>Header</span>\n}\n```",
			},
		},
		{
			Label:      "let",
			Kind:       CompletionItemKindKeyword,
			Detail:     "Bind element to variable",
			InsertText: "let ${1:name} = ",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "Bind an element to a variable for later reference.\n\n```gsx\n@let header = <div>Header</div>\n```",
			},
		},
	}
}

func (c *completionProvider) getElementCompletions() []CompletionItem {
	var items []CompletionItem
	for _, tag := range schema.AllElementTags() {
		elem := schema.GetElement(tag)
		if elem == nil {
			continue
		}

		insertText := tag + ">$0</" + tag + ">"
		if elem.SelfClosing {
			insertText = tag + " />"
		}

		items = append(items, CompletionItem{
			Label:      tag,
			Kind:       CompletionItemKindClass,
			Detail:     elem.Category,
			InsertText: insertText,
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: elem.Description,
			},
		})
	}
	return items
}

func (c *completionProvider) getAttributeCompletions(tag string) []CompletionItem {
	elem := schema.GetElement(tag)
	if elem == nil {
		return nil
	}

	var items []CompletionItem
	for _, attr := range elem.Attributes {
		insertText := attr.Name + "="
		if attr.Type == "bool" {
			insertText = attr.Name
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

	// Also offer event handler attributes that aren't already in the element's schema
	for _, handlerName := range schema.AllEventHandlerNames() {
		handler := schema.GetEventHandler(handlerName)
		if handler == nil {
			continue
		}
		// Skip if already in the element's attributes
		if schema.GetAttribute(tag, handlerName) != nil {
			continue
		}
		items = append(items, CompletionItem{
			Label:      handlerName,
			Kind:       CompletionItemKindEvent,
			Detail:     handler.Signature,
			InsertText: handlerName + "={$1}",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: handler.Description,
			},
		})
	}

	return items
}

func (c *completionProvider) getContextualCompletions(ctx *CursorContext) []CompletionItem {
	// If inside an element tag, offer attribute completions
	if ctx.InElement && ctx.AttrTag != "" {
		return c.getAttributeCompletions(ctx.AttrTag)
	}

	// Check if we can detect the enclosing tag from text
	tag := enclosingTagFromText(ctx)
	if tag != "" {
		return c.getAttributeCompletions(tag)
	}

	// Default: offer all top-level completions
	var items []CompletionItem
	items = append(items, c.getComponentCompletions()...)
	items = append(items, c.getDSLKeywordCompletions()...)
	items = append(items, c.getElementCompletions()...)
	return items
}

// enclosingTagFromText searches backwards from cursor for an unclosed < to find the tag name.
func enclosingTagFromText(ctx *CursorContext) string {
	offset := PositionToOffset(ctx.Document.Content, ctx.Position)
	content := ctx.Document.Content

	for i := offset - 1; i >= 0; i-- {
		if content[i] == '<' {
			// Extract tag name
			j := i + 1
			for j < len(content) && IsWordChar(content[j]) {
				j++
			}
			if j > i+1 {
				tagName := content[i+1 : j]
				// Check we haven't passed a > before cursor
				for k := j; k < offset; k++ {
					if content[k] == '>' {
						return "" // Past the tag
					}
				}
				return tagName
			}
			break
		}
		if content[i] == '>' {
			break // Hit a closing bracket
		}
	}
	return ""
}

// --- State method completions ---

// getStateMethodCompletions returns state method completions when the user types
// a state variable name followed by a dot (e.g., "count.").
func (c *completionProvider) getStateMethodCompletions(ctx *CursorContext) []CompletionItem {
	if ctx.Scope == nil || len(ctx.Scope.StateVars) == 0 {
		return nil
	}

	// Check if the text before cursor looks like "stateVarName."
	offset := PositionToOffset(ctx.Document.Content, ctx.Position)
	if offset <= 1 {
		return nil
	}

	content := ctx.Document.Content
	// Look for a dot just before cursor
	dotPos := offset - 1
	// Skip back past whitespace to the dot
	for dotPos > 0 && (content[dotPos] == ' ' || content[dotPos] == '\t') {
		dotPos--
	}
	if dotPos < 0 || content[dotPos] != '.' {
		return nil
	}

	// Extract the word before the dot
	wordEnd := dotPos
	wordStart := wordEnd - 1
	for wordStart >= 0 && IsWordChar(content[wordStart]) {
		wordStart--
	}
	wordStart++
	if wordStart >= wordEnd {
		return nil
	}
	varName := content[wordStart:wordEnd]

	// Check if it matches a state variable
	isStateVar := false
	for _, sv := range ctx.Scope.StateVars {
		if sv.Name == varName {
			isStateVar = true
			break
		}
	}
	if !isStateVar {
		return nil
	}

	log.Server("State method completion for %q", varName)

	return []CompletionItem{
		{
			Label:      "Get()",
			Kind:       CompletionItemKindMethod,
			Detail:     "Get current value",
			InsertText: "Get()",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "Returns the current value of the state variable.",
			},
		},
		{
			Label:      "Set(value)",
			Kind:       CompletionItemKindMethod,
			Detail:     "Set new value",
			InsertText: "Set(${1:value})",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "Sets a new value for the state variable, triggering a re-render.",
			},
		},
		{
			Label:      "Update(fn)",
			Kind:       CompletionItemKindMethod,
			Detail:     "Update with function",
			InsertText: "Update(func(current ${1:T}) ${2:T} {\n\t${3:return current}\n})",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "Updates the state using a function that receives the current value and returns the new value.",
			},
		},
		{
			Label:      "Bind(fn)",
			Kind:       CompletionItemKindMethod,
			Detail:     "Register change callback",
			InsertText: "Bind(${1:callback})",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "Registers a callback that is called when the state changes.",
			},
		},
		{
			Label:      "Batch(fn)",
			Kind:       CompletionItemKindMethod,
			Detail:     "Batch multiple updates",
			InsertText: "Batch(func() {\n\t$0\n})",
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: "Batches multiple state updates into a single re-render.",
			},
		},
	}
}

// --- Tailwind class completions ---

func (c *completionProvider) getTailwindCompletions(prefix string) []CompletionItem {
	matches := schema.MatchClasses(prefix)

	var items []CompletionItem
	for _, cls := range matches {
		docValue := cls.Description
		items = append(items, CompletionItem{
			Label:      cls.Name,
			Kind:       CompletionItemKindConstant,
			Detail:     cls.Category,
			InsertText: cls.Name,
			FilterText: cls.Name,
			Documentation: &MarkupContent{
				Kind:  "markdown",
				Value: docValue,
			},
		})
	}

	sortCompletionsByCategory(items)
	return items
}

// sortCompletionsByCategory sorts completion items by category priority then name.
func sortCompletionsByCategory(items []CompletionItem) {
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

// --- gopls completion delegation ---

func (c *completionProvider) getGoplsCompletions(ctx *CursorContext) ([]CompletionItem, error) {
	proxy := c.goplsProxy.GetProxy()
	if proxy == nil {
		return nil, nil
	}

	cached := c.virtualFiles.GetVirtualFile(ctx.Document.URI)
	if cached == nil || cached.SourceMap == nil {
		return nil, nil
	}

	goLine, goCol, found := cached.SourceMap.TuiToGo(ctx.Position.Line, ctx.Position.Character)
	if !found {
		log.Server("No mapping found for completion position %d:%d", ctx.Position.Line, ctx.Position.Character)
		return nil, nil
	}

	goplsItems, err := proxy.Completion(cached.GoURI, gopls.Position{
		Line:      goLine,
		Character: goCol,
	})
	if err != nil {
		return nil, err
	}

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
