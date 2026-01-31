package provider

import (
	"strings"

	"github.com/grindlemire/go-tui/internal/lsp/gopls"
	"github.com/grindlemire/go-tui/internal/lsp/log"
	"github.com/grindlemire/go-tui/internal/tuigen"
)

// definitionProvider implements DefinitionProvider.
type definitionProvider struct {
	index        ComponentIndex
	goplsProxy   GoplsProxyAccessor
	virtualFiles VirtualFileAccessor
	docs         DocumentAccessor
}

// NewDefinitionProvider creates a new definition provider.
func NewDefinitionProvider(index ComponentIndex, proxy GoplsProxyAccessor, vf VirtualFileAccessor, docs DocumentAccessor) DefinitionProvider {
	return &definitionProvider{
		index:        index,
		goplsProxy:   proxy,
		virtualFiles: vf,
		docs:         docs,
	}
}

func (d *definitionProvider) Definition(ctx *CursorContext) ([]Location, error) {
	log.Server("Definition provider: NodeKind=%s, Word=%q, InGoExpr=%v", ctx.NodeKind, ctx.Word, ctx.InGoExpr)

	word := ctx.Word

	// Check local function definitions first (prevents gopls from
	// returning generated .go files instead of .gsx sources).
	if word != "" {
		if funcInfo, ok := d.index.LookupFunc(word); ok {
			log.Server("Found local function %s at %s (before gopls)", word, funcInfo.Location.URI)
			return []Location{funcInfo.Location}, nil
		}
	}

	if word == "" {
		// Without a word, only gopls can resolve (it works by position).
		if ctx.InGoExpr {
			locs, err := d.getGoplsDefinition(ctx)
			if err == nil && len(locs) > 0 {
				return locs, nil
			}
		}
		return nil, nil
	}

	// Dispatch based on node kind. Each case resolves using local knowledge
	// first; unresolved Go expressions fall through to the gopls fallback.
	switch ctx.NodeKind {
	case NodeKindComponentCall:
		return d.definitionComponentCall(ctx)
	case NodeKindNamedRef:
		return d.definitionNamedRef(ctx)
	case NodeKindEventHandler:
		return d.definitionEventHandler(ctx)
	case NodeKindParameter:
		return d.definitionParameter(ctx)
	case NodeKindStateAccess, NodeKindStateDecl:
		locs, err := d.definitionStateVar(ctx)
		if err == nil && len(locs) > 0 {
			return locs, nil
		}
		// Fall through to gopls for unresolved state references
	case NodeKindGoExpr:
		// Check named refs in scope before deferring to gopls (which would
		// return the element tag position from generated code, not #Name).
		if locs := d.definitionNamedRefFromScope(ctx); len(locs) > 0 {
			return locs, nil
		}
		// Fall through to gopls
	case NodeKindFunction:
		locs, err := d.getGoplsDefinition(ctx)
		if err == nil && len(locs) > 0 {
			return locs, nil
		}
	case NodeKindComponent:
		locs, err := d.getGoplsDefinition(ctx)
		if err == nil && len(locs) > 0 {
			return locs, nil
		}
	}

	// Gopls fallback for Go expressions not resolved by handlers above.
	if ctx.InGoExpr {
		locs, err := d.getGoplsDefinition(ctx)
		if err != nil {
			log.Server("gopls definition error: %v", err)
		} else if len(locs) > 0 {
			return locs, nil
		}
	}

	// Word-based fallbacks
	componentName := strings.TrimPrefix(word, "@")

	// Look up component in index
	if info, ok := d.index.Lookup(componentName); ok {
		return []Location{info.Location}, nil
	}

	// Check if it's a function
	if funcInfo, ok := d.index.LookupFunc(word); ok {
		return []Location{funcInfo.Location}, nil
	}

	// Check within component scope
	if ctx.Scope.Component != nil {
		compName := ctx.Scope.Component.Name

		// Parameter
		if paramInfo, ok := d.index.LookupParam(compName, word); ok {
			return []Location{paramInfo.Location}, nil
		}

		// Let binding
		if loc := d.findLetBindingDefinition(ctx, word); loc != nil {
			return []Location{*loc}, nil
		}

		// For loop variable
		if loc := d.findLoopVariableDefinition(ctx, word); loc != nil {
			return []Location{*loc}, nil
		}

		// GoCode variable
		if loc := d.findGoCodeVariableDefinition(ctx, word); loc != nil {
			return []Location{*loc}, nil
		}
	}

	// AST-based component/function definition within current file
	if ctx.Document.AST != nil {
		if loc := d.findComponentInAST(ctx.Document.AST, componentName, ctx.Document.URI); loc != nil {
			return []Location{*loc}, nil
		}
		if loc := d.findFuncInAST(ctx.Document.AST, word, ctx.Document.URI); loc != nil {
			return []Location{*loc}, nil
		}
	}

	return nil, nil
}

// --- Component and function definition ---

func (d *definitionProvider) definitionComponentCall(ctx *CursorContext) ([]Location, error) {
	call, ok := ctx.Node.(*tuigen.ComponentCall)
	if !ok || call == nil {
		return nil, nil
	}

	if info, ok := d.index.Lookup(call.Name); ok {
		return []Location{info.Location}, nil
	}
	return nil, nil
}

// definitionNamedRefFromScope checks if the word under the cursor matches a
// named ref in the component scope and returns its #Name definition position.
func (d *definitionProvider) definitionNamedRefFromScope(ctx *CursorContext) []Location {
	if ctx.Scope.Component == nil {
		return nil
	}
	for _, ref := range ctx.Scope.NamedRefs {
		if ref.Name == ctx.Word && ref.Element != nil {
			lineIdx, charIdx, found := findNamedRefPosition(ctx.Document.Content, ref.Element)
			if found {
				hashRef := "#" + ref.Name
				return []Location{{
					URI: ctx.Document.URI,
					Range: Range{
						Start: Position{Line: lineIdx, Character: charIdx},
						End:   Position{Line: lineIdx, Character: charIdx + len(hashRef)},
					},
				}}
			}
		}
	}
	return nil
}

func (d *definitionProvider) definitionNamedRef(ctx *CursorContext) ([]Location, error) {
	elem, ok := ctx.Node.(*tuigen.Element)
	if !ok || elem == nil || elem.NamedRef == "" {
		return nil, nil
	}

	lineIdx, charIdx, found := findNamedRefPosition(ctx.Document.Content, elem)
	if !found {
		// Fallback to element tag position
		lineIdx = elem.Position.Line - 1
		charIdx = elem.Position.Column - 1
	}

	hashRef := "#" + elem.NamedRef
	return []Location{{
		URI: ctx.Document.URI,
		Range: Range{
			Start: Position{Line: lineIdx, Character: charIdx},
			End:   Position{Line: lineIdx, Character: charIdx + len(hashRef)},
		},
	}}, nil
}

// findNamedRefPosition finds the source position of #Name for an element.
// Searches from the element's tag line through subsequent lines to handle
// multiline elements where #Name is on its own line.
// Returns 0-indexed line and column.
func findNamedRefPosition(content string, elem *tuigen.Element) (line, col int, found bool) {
	if elem == nil || elem.NamedRef == "" {
		return 0, 0, false
	}

	hashRef := "#" + elem.NamedRef
	lines := strings.Split(content, "\n")
	startLine := elem.Position.Line - 1 // 0-indexed

	maxSearch := startLine + 20
	if maxSearch > len(lines) {
		maxSearch = len(lines)
	}

	for lineIdx := startLine; lineIdx < maxSearch; lineIdx++ {
		idx := strings.Index(lines[lineIdx], hashRef)
		if idx >= 0 {
			return lineIdx, idx, true
		}
	}

	return 0, 0, false
}

func (d *definitionProvider) definitionStateVar(ctx *CursorContext) ([]Location, error) {
	// Find the state variable declaration in scope, matching by name
	for _, sv := range ctx.Scope.StateVars {
		if sv.Name == ctx.Word {
			return []Location{{
				URI: ctx.Document.URI,
				Range: Range{
					Start: Position{Line: sv.Position.Line - 1, Character: sv.Position.Column - 1},
					End:   Position{Line: sv.Position.Line - 1, Character: sv.Position.Column - 1 + len(sv.Name)},
				},
			}}, nil
		}
	}
	return nil, nil
}

func (d *definitionProvider) definitionEventHandler(ctx *CursorContext) ([]Location, error) {
	// Event handler attributes — try to find the handler function in Go expressions
	// For now, fall back to gopls for Go expression resolution
	if ctx.InGoExpr {
		return d.getGoplsDefinition(ctx)
	}
	return nil, nil
}

func (d *definitionProvider) definitionParameter(ctx *CursorContext) ([]Location, error) {
	word := ctx.Word

	// Function parameter
	if ctx.Scope.Function != nil {
		funcName := parseFuncName(ctx.Scope.Function.Code)
		if paramInfo, ok := d.index.LookupFuncParam(funcName, word); ok {
			return []Location{paramInfo.Location}, nil
		}
	}

	// Component parameter
	if ctx.Scope.Component != nil {
		if paramInfo, ok := d.index.LookupParam(ctx.Scope.Component.Name, word); ok {
			return []Location{paramInfo.Location}, nil
		}
	}

	return nil, nil
}

// --- Local variable definitions ---

func (d *definitionProvider) findLetBindingDefinition(ctx *CursorContext, varName string) *Location {
	if ctx.Document.AST == nil || ctx.Scope.Component == nil {
		return nil
	}

	for _, comp := range ctx.Document.AST.Components {
		if comp.Name != ctx.Scope.Component.Name {
			continue
		}

		if binding := findLetBindingInNodes(comp.Body, varName); binding != nil {
			return &Location{
				URI: ctx.Document.URI,
				Range: Range{
					Start: Position{
						Line:      binding.Position.Line - 1,
						Character: binding.Position.Column - 1 + len("@let "),
					},
					End: Position{
						Line:      binding.Position.Line - 1,
						Character: binding.Position.Column - 1 + len("@let ") + len(varName),
					},
				},
			}
		}
	}
	return nil
}

func (d *definitionProvider) findLoopVariableDefinition(ctx *CursorContext, varName string) *Location {
	if ctx.Document.AST == nil || ctx.Scope.Component == nil {
		return nil
	}

	for _, comp := range ctx.Document.AST.Components {
		if comp.Name != ctx.Scope.Component.Name {
			continue
		}

		loop := findForLoopWithVariable(comp.Body, varName)
		if loop == nil {
			return nil
		}

		if loop.Index == varName {
			return &Location{
				URI: ctx.Document.URI,
				Range: Range{
					Start: Position{
						Line:      loop.Position.Line - 1,
						Character: loop.Position.Column - 1 + len("@for "),
					},
					End: Position{
						Line:      loop.Position.Line - 1,
						Character: loop.Position.Column - 1 + len("@for ") + len(varName),
					},
				},
			}
		} else if loop.Value == varName {
			offset := len("@for ")
			if loop.Index != "" {
				offset += len(loop.Index) + 2 // ", "
			}
			return &Location{
				URI: ctx.Document.URI,
				Range: Range{
					Start: Position{
						Line:      loop.Position.Line - 1,
						Character: loop.Position.Column - 1 + offset,
					},
					End: Position{
						Line:      loop.Position.Line - 1,
						Character: loop.Position.Column - 1 + offset + len(varName),
					},
				},
			}
		}
	}
	return nil
}

func (d *definitionProvider) findGoCodeVariableDefinition(ctx *CursorContext, varName string) *Location {
	if ctx.Document.AST == nil || ctx.Scope.Component == nil {
		return nil
	}

	for _, comp := range ctx.Document.AST.Components {
		if comp.Name != ctx.Scope.Component.Name {
			continue
		}

		goCode := findGoCodeWithVariable(comp.Body, varName)
		if goCode == nil {
			return nil
		}

		idx := findVarDeclPosition(goCode.Code, varName)
		if idx >= 0 {
			return &Location{
				URI: ctx.Document.URI,
				Range: Range{
					Start: Position{
						Line:      goCode.Position.Line - 1,
						Character: goCode.Position.Column - 1 + idx,
					},
					End: Position{
						Line:      goCode.Position.Line - 1,
						Character: goCode.Position.Column - 1 + idx + len(varName),
					},
				},
			}
		}
	}
	return nil
}

// --- AST lookup helpers ---

func (d *definitionProvider) findComponentInAST(ast *tuigen.File, name string, uri string) *Location {
	for _, comp := range ast.Components {
		if comp.Name == name {
			return &Location{
				URI: uri,
				Range: Range{
					Start: Position{Line: comp.Position.Line - 1, Character: comp.Position.Column - 1},
					End:   Position{Line: comp.Position.Line - 1, Character: comp.Position.Column - 1 + len("templ") + 1 + len(comp.Name)},
				},
			}
		}
	}
	return nil
}

func (d *definitionProvider) findFuncInAST(ast *tuigen.File, name string, uri string) *Location {
	for _, fn := range ast.Funcs {
		fnName := parseFuncName(fn.Code)
		if fnName == name {
			return &Location{
				URI: uri,
				Range: Range{
					Start: Position{Line: fn.Position.Line - 1, Character: fn.Position.Column - 1},
					End:   Position{Line: fn.Position.Line - 1, Character: fn.Position.Column - 1 + len("func") + 1 + len(fnName)},
				},
			}
		}
	}
	return nil
}

// --- gopls definition delegation ---

func (d *definitionProvider) getGoplsDefinition(ctx *CursorContext) ([]Location, error) {
	proxy := d.goplsProxy.GetProxy()
	if proxy == nil {
		return nil, nil
	}

	cached := d.virtualFiles.GetVirtualFile(ctx.Document.URI)
	if cached == nil || cached.SourceMap == nil {
		return nil, nil
	}

	goLine, goCol, found := cached.SourceMap.TuiToGo(ctx.Position.Line, ctx.Position.Character)
	if !found {
		return nil, nil
	}

	goplsLocs, err := proxy.Definition(cached.GoURI, gopls.Position{
		Line:      goLine,
		Character: goCol,
	})
	if err != nil {
		return nil, err
	}

	if len(goplsLocs) == 0 {
		return nil, nil
	}

	var locs []Location
	for _, gl := range goplsLocs {
		// Check if this is a virtual file — translate back to .gsx
		if gopls.IsVirtualGoFile(gl.URI) {
			tuiURI := gopls.GoURIToTuiURI(gl.URI)
			cachedFile := d.virtualFiles.GetVirtualFile(tuiURI)
			if cachedFile != nil && cachedFile.SourceMap != nil {
				tuiStartLine, tuiStartCol, startFound := cachedFile.SourceMap.GoToTui(gl.Range.Start.Line, gl.Range.Start.Character)
				tuiEndLine, tuiEndCol, endFound := cachedFile.SourceMap.GoToTui(gl.Range.End.Line, gl.Range.End.Character)
				if startFound && endFound {
					locs = append(locs, Location{
						URI: tuiURI,
						Range: Range{
							Start: Position{Line: tuiStartLine, Character: tuiStartCol},
							End:   Position{Line: tuiEndLine, Character: tuiEndCol},
						},
					})
					continue
				}
			}
		}

		// External file (standard library, etc.) — return as-is
		locs = append(locs, Location{
			URI: gl.URI,
			Range: Range{
				Start: Position{Line: gl.Range.Start.Line, Character: gl.Range.Start.Character},
				End:   Position{Line: gl.Range.End.Line, Character: gl.Range.End.Character},
			},
		})
	}

	return locs, nil
}
