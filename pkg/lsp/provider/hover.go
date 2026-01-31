// Package provider contains LSP feature implementations organized by capability.
package provider

import (
	"fmt"
	"strings"

	"github.com/grindlemire/go-tui/pkg/lsp/gopls"
	"github.com/grindlemire/go-tui/pkg/lsp/log"
	"github.com/grindlemire/go-tui/pkg/lsp/schema"
	"github.com/grindlemire/go-tui/pkg/tuigen"
)

// hoverProvider implements HoverProvider.
type hoverProvider struct {
	index        ComponentIndex
	goplsProxy   GoplsProxyAccessor
	virtualFiles VirtualFileAccessor
}

// NewHoverProvider creates a new hover provider.
func NewHoverProvider(index ComponentIndex, proxy GoplsProxyAccessor, vf VirtualFileAccessor) HoverProvider {
	return &hoverProvider{
		index:        index,
		goplsProxy:   proxy,
		virtualFiles: vf,
	}
}

func (h *hoverProvider) Hover(ctx *CursorContext) (*Hover, error) {
	log.Server("Hover provider: NodeKind=%s, Word=%q, InGoExpr=%v, InClassAttr=%v",
		ctx.NodeKind, ctx.Word, ctx.InGoExpr, ctx.InClassAttr)

	// For Go expressions, try gopls first
	if ctx.InGoExpr {
		hover, err := h.getGoplsHover(ctx)
		if err != nil {
			log.Server("gopls hover error: %v", err)
		} else if hover != nil {
			return hover, nil
		}
	}

	// Dispatch based on node kind
	switch ctx.NodeKind {
	case NodeKindComponent:
		// Try gopls for types in component parameter signatures
		hover, err := h.getGoplsHover(ctx)
		if err == nil && hover != nil {
			return hover, nil
		}
		return h.hoverComponent(ctx)
	case NodeKindElement:
		return h.hoverElement(ctx)
	case NodeKindAttribute:
		return h.hoverAttribute(ctx)
	case NodeKindEventHandler:
		return h.hoverEventHandler(ctx)
	case NodeKindParameter:
		return h.hoverParameter(ctx)
	case NodeKindKeyword:
		return h.hoverKeyword(ctx)
	case NodeKindForLoop:
		return h.hoverKeyword(ctx)
	case NodeKindIfStmt:
		return h.hoverKeyword(ctx)
	case NodeKindLetBinding:
		return h.hoverKeyword(ctx)
	case NodeKindFunction:
		// Try gopls for types in function signatures (e.g., *tui.State[int])
		hover, err := h.getGoplsHover(ctx)
		if err == nil && hover != nil {
			return hover, nil
		}
		return h.hoverFunction(ctx)
	case NodeKindComponentCall:
		return h.hoverComponentCall(ctx)
	case NodeKindNamedRef:
		return h.hoverNamedRef(ctx)
	case NodeKindStateDecl:
		return h.hoverStateDecl(ctx)
	case NodeKindStateAccess:
		return h.hoverStateAccess(ctx)
	case NodeKindTailwindClass:
		return h.hoverTailwindClass(ctx)
	case NodeKindGoExpr:
		// Already tried gopls above; fall through to word-based checks
	}

	// Word-based fallbacks
	word := ctx.Word
	if word == "" {
		return nil, nil
	}

	// Check if word is a keyword
	if hover := h.hoverForKeyword(word); hover != nil {
		return hover, nil
	}

	// Check if word is a component call (@Name or Name)
	componentName := strings.TrimPrefix(word, "@")
	if info, ok := h.index.Lookup(componentName); ok {
		return hoverForComponentInfo(info), nil
	}

	// Check if word is a function
	if funcInfo, ok := h.index.LookupFunc(word); ok {
		return hoverForFuncInfo(funcInfo), nil
	}

	// Check if word is a parameter in the current component
	if ctx.Scope.Component != nil {
		if paramInfo, ok := h.index.LookupParam(ctx.Scope.Component.Name, word); ok {
			return hoverForParamInfo(paramInfo), nil
		}
	}

	// Check if word is an element tag
	if elem := schema.GetElement(word); elem != nil {
		return hoverForElement(elem), nil
	}

	// Check for tailwind class (when InClassAttr but no AST resolution)
	if ctx.InClassAttr {
		return h.hoverForTailwindWord(ctx)
	}

	// Check for attribute (when AST didn't resolve but we're in an element)
	if ctx.InElement && ctx.AttrTag != "" {
		if attr := schema.GetAttribute(ctx.AttrTag, word); attr != nil {
			return hoverForAttributeDef(ctx.AttrTag, attr), nil
		}
	}

	return nil, nil
}

// --- Node-kind-specific hover functions ---

func (h *hoverProvider) hoverComponent(ctx *CursorContext) (*Hover, error) {
	comp, ok := ctx.Node.(*tuigen.Component)
	if !ok || comp == nil {
		return nil, nil
	}

	if info, ok := h.index.Lookup(comp.Name); ok {
		return hoverForComponentInfo(info), nil
	}

	// Fallback: build from AST directly
	var params []string
	for _, p := range comp.Params {
		params = append(params, fmt.Sprintf("%s %s", p.Name, p.Type))
	}
	sig := fmt.Sprintf("func %s(%s) *element.Element", comp.Name, strings.Join(params, ", "))
	md := fmt.Sprintf("```go\n%s\n```\n\n**TUI Component**", sig)
	return markdownHover(md), nil
}

func (h *hoverProvider) hoverElement(ctx *CursorContext) (*Hover, error) {
	elem, ok := ctx.Node.(*tuigen.Element)
	if !ok || elem == nil {
		return nil, nil
	}

	def := schema.GetElement(elem.Tag)
	if def == nil {
		return nil, nil
	}
	return hoverForElement(def), nil
}

func (h *hoverProvider) hoverAttribute(ctx *CursorContext) (*Hover, error) {
	if ctx.AttrTag == "" || ctx.AttrName == "" {
		return nil, nil
	}

	attr := schema.GetAttribute(ctx.AttrTag, ctx.AttrName)
	if attr != nil {
		return hoverForAttributeDef(ctx.AttrTag, attr), nil
	}

	// Fallback for unknown attributes
	return markdownHover(fmt.Sprintf("**%s** attribute on `<%s>`", ctx.AttrName, ctx.AttrTag)), nil
}

func (h *hoverProvider) hoverEventHandler(ctx *CursorContext) (*Hover, error) {
	handler := schema.GetEventHandler(ctx.AttrName)
	if handler != nil {
		md := fmt.Sprintf("**Event Handler** `%s`\n\nType: `%s`\n\n%s",
			handler.Name, handler.Signature, handler.Description)
		return markdownHover(md), nil
	}
	return nil, nil
}

func (h *hoverProvider) hoverParameter(ctx *CursorContext) (*Hover, error) {
	param, ok := ctx.Node.(*tuigen.Param)
	if !ok || param == nil {
		return nil, nil
	}

	compName := ""
	if ctx.Scope.Component != nil {
		compName = ctx.Scope.Component.Name
	}

	md := fmt.Sprintf("```go\n%s %s\n```\n\n**Parameter** of component `%s`",
		param.Name, param.Type, compName)
	return markdownHover(md), nil
}

func (h *hoverProvider) hoverKeyword(ctx *CursorContext) (*Hover, error) {
	return h.hoverForKeyword(ctx.Word), nil
}

func (h *hoverProvider) hoverFunction(ctx *CursorContext) (*Hover, error) {
	fn, ok := ctx.Node.(*tuigen.GoFunc)
	if !ok || fn == nil {
		return nil, nil
	}

	word := ctx.Word
	if funcInfo, ok := h.index.LookupFunc(word); ok {
		return hoverForFuncInfo(funcInfo), nil
	}
	return nil, nil
}

func (h *hoverProvider) hoverComponentCall(ctx *CursorContext) (*Hover, error) {
	call, ok := ctx.Node.(*tuigen.ComponentCall)
	if !ok || call == nil {
		return nil, nil
	}

	if info, ok := h.index.Lookup(call.Name); ok {
		return hoverForComponentInfo(info), nil
	}
	return nil, nil
}

func (h *hoverProvider) hoverNamedRef(ctx *CursorContext) (*Hover, error) {
	elem, ok := ctx.Node.(*tuigen.Element)
	if !ok || elem == nil {
		return nil, nil
	}

	refType := "`*element.Element`"
	refContext := "Simple (direct access)"
	accessPattern := fmt.Sprintf("`view.%s`", elem.NamedRef)

	// Check scope for richer context
	{
		for _, ref := range ctx.Scope.NamedRefs {
			if ref.Name == elem.NamedRef {
				if ref.InLoop {
					if ref.KeyExpr != "" {
						refType = "`map[KeyType]*element.Element`"
						refContext = "Keyed (map access)"
						accessPattern = fmt.Sprintf("`view.%s[key]`", elem.NamedRef)
					} else {
						refType = "`[]*element.Element`"
						refContext = "Loop (slice access)"
						accessPattern = fmt.Sprintf("`view.%s[i]`", elem.NamedRef)
					}
				}
				if ref.InConditional {
					refContext += " (nullable)"
				}
				break
			}
		}
	}

	md := fmt.Sprintf("**Named Ref** `%s`\n\nType: %s\n\nContext: %s\n\nAccess via view struct: %s",
		elem.NamedRef, refType, refContext, accessPattern)
	return markdownHover(md), nil
}

func (h *hoverProvider) hoverStateDecl(ctx *CursorContext) (*Hover, error) {
	// Try to find the state variable info from scope, matching by name
	for _, sv := range ctx.Scope.StateVars {
		if sv.Name == ctx.Word {
			md := fmt.Sprintf("**State Variable** `%s`\n\nType: `*tui.State[%s]`\n\nInitial: `%s`\n\nMethods: Get(), Set(), Update(), Bind(), Batch()",
				sv.Name, sv.Type, sv.InitExpr)
			return markdownHover(md), nil
		}
	}

	md := "**State Declaration** (`tui.NewState`)\n\nCreates a reactive state variable."
	return markdownHover(md), nil
}

func (h *hoverProvider) hoverStateAccess(ctx *CursorContext) (*Hover, error) {
	word := ctx.Word
	if strings.HasSuffix(word, "Get") || word == "Get" {
		return markdownHover("**State.Get()** — Returns the current value of the state variable."), nil
	}
	if strings.HasSuffix(word, "Set") || word == "Set" {
		return markdownHover("**State.Set(value)** — Sets a new value for the state variable."), nil
	}
	if strings.HasSuffix(word, "Update") || word == "Update" {
		return markdownHover("**State.Update(fn)** — Updates the state using a function that receives the current value."), nil
	}
	if strings.HasSuffix(word, "Bind") || word == "Bind" {
		return markdownHover("**State.Bind(fn)** — Registers a callback that is called when the state changes."), nil
	}
	if strings.HasSuffix(word, "Batch") || word == "Batch" {
		return markdownHover("**State.Batch(fn)** — Batches multiple state updates into a single re-render."), nil
	}
	return markdownHover("**State Access** — Reactive state method call."), nil
}

func (h *hoverProvider) hoverTailwindClass(ctx *CursorContext) (*Hover, error) {
	return h.hoverForTailwindWord(ctx)
}

// hoverForTailwindWord extracts the class name at the cursor and returns hover docs.
func (h *hoverProvider) hoverForTailwindWord(ctx *CursorContext) (*Hover, error) {
	offset := ctx.Offset
	content := ctx.Document.Content

	// Search backwards for class="
	searchStart := offset - maxClassAttrSearchDistance
	if searchStart < 0 {
		searchStart = 0
	}

	segment := content[searchStart:offset]
	classIdx := strings.LastIndex(segment, `class="`)
	if classIdx == -1 {
		return nil, nil
	}

	// Check we haven't passed the closing quote
	afterClass := segment[classIdx+7:]
	if strings.Contains(afterClass, `"`) {
		return nil, nil
	}

	// Find the class name at cursor
	classStart := searchStart + classIdx + 7
	classContent := content[classStart:offset]

	lastSpace := strings.LastIndex(classContent, " ")
	var className string
	if lastSpace == -1 {
		className = classContent
	} else {
		className = classContent[lastSpace+1:]
	}

	// Extend forward for full class name
	endOffset := offset
	for endOffset < len(content) && content[endOffset] != ' ' && content[endOffset] != '"' {
		endOffset++
	}
	if endOffset > offset {
		className += content[offset:endOffset]
	}

	className = strings.TrimSpace(className)
	if className == "" {
		return nil, nil
	}

	doc := schema.GetClassDoc(className)
	if doc == "" {
		return nil, nil
	}

	return markdownHover(fmt.Sprintf("**`%s`**\n\n%s", className, doc)), nil
}

// hoverForKeyword returns hover for a keyword word.
func (h *hoverProvider) hoverForKeyword(word string) *Hover {
	kw := schema.GetKeyword(word)
	if kw == nil {
		return nil
	}
	return markdownHover(kw.Documentation)
}

// --- Hover formatting helpers ---

func hoverForComponentInfo(info *ComponentInfo) *Hover {
	var params []string
	for _, p := range info.Params {
		params = append(params, fmt.Sprintf("%s %s", p.Name, p.Type))
	}
	sig := fmt.Sprintf("func %s(%s) *element.Element", info.Name, strings.Join(params, ", "))
	md := fmt.Sprintf("```go\n%s\n```\n\n**TUI Component**", sig)
	return markdownHover(md)
}

func hoverForFuncInfo(info *FuncInfo) *Hover {
	md := fmt.Sprintf("```go\n%s\n```\n\n**Helper Function**", info.Signature)
	return markdownHover(md)
}

func hoverForParamInfo(info *ParamInfo) *Hover {
	md := fmt.Sprintf("```go\n%s %s\n```\n\n**Parameter** of component `%s`",
		info.Name, info.Type, info.ComponentName)
	return markdownHover(md)
}

func hoverForElement(def *schema.ElementDef) *Hover {
	var lines []string
	lines = append(lines, fmt.Sprintf("## `<%s>`", def.Tag))
	lines = append(lines, "")
	lines = append(lines, def.Description)
	lines = append(lines, "")
	lines = append(lines, "**Available attributes:**")
	for _, attr := range def.Attributes {
		lines = append(lines, fmt.Sprintf("- `%s` (%s): %s", attr.Name, attr.Type, attr.Description))
	}
	return markdownHover(strings.Join(lines, "\n"))
}

func hoverForAttributeDef(tag string, attr *schema.AttributeDef) *Hover {
	md := fmt.Sprintf("**%s** (`%s`)\n\n%s", attr.Name, attr.Type, attr.Description)
	return markdownHover(md)
}

func markdownHover(content string) *Hover {
	return &Hover{
		Contents: MarkupContent{
			Kind:  "markdown",
			Value: content,
		},
	}
}

// --- gopls hover delegation ---

func (h *hoverProvider) getGoplsHover(ctx *CursorContext) (*Hover, error) {
	proxy := h.goplsProxy.GetProxy()
	if proxy == nil {
		return nil, nil
	}

	cached := h.virtualFiles.GetVirtualFile(ctx.Document.URI)
	if cached == nil || cached.SourceMap == nil {
		return nil, nil
	}

	goLine, goCol, found := cached.SourceMap.TuiToGo(ctx.Position.Line, ctx.Position.Character)
	if !found {
		return nil, nil
	}

	goplsHover, err := proxy.Hover(cached.GoURI, gopls.Position{
		Line:      goLine,
		Character: goCol,
	})
	if err != nil {
		return nil, err
	}

	if goplsHover == nil {
		return nil, nil
	}

	hover := &Hover{
		Contents: MarkupContent{
			Kind:  goplsHover.Contents.Kind,
			Value: goplsHover.Contents.Value,
		},
	}

	if goplsHover.Range != nil {
		tuiStartLine, tuiStartCol, startFound := cached.SourceMap.GoToTui(goplsHover.Range.Start.Line, goplsHover.Range.Start.Character)
		tuiEndLine, tuiEndCol, endFound := cached.SourceMap.GoToTui(goplsHover.Range.End.Line, goplsHover.Range.End.Character)
		if startFound && endFound {
			hover.Range = &Range{
				Start: Position{Line: tuiStartLine, Character: tuiStartCol},
				End:   Position{Line: tuiEndLine, Character: tuiEndCol},
			}
		}
	}

	return hover, nil
}
