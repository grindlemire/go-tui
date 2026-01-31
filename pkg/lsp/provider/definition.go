package provider

import (
	"strings"

	"github.com/grindlemire/go-tui/pkg/lsp/gopls"
	"github.com/grindlemire/go-tui/pkg/lsp/log"
	"github.com/grindlemire/go-tui/pkg/tuigen"
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

	// First, check local function definitions before gopls (prevents gopls from
	// returning generated .go files instead of .gsx sources).
	if word != "" {
		if funcInfo, ok := d.index.LookupFunc(word); ok {
			log.Server("Found local function %s at %s (before gopls)", word, funcInfo.Location.URI)
			return []Location{funcInfo.Location}, nil
		}

		// Check for loop variables before gopls (gopls doesn't understand .gsx for loops)
		if ctx.Scope != nil && ctx.Scope.Component != nil {
			if loc := d.findLoopVariableDefinition(ctx, word); loc != nil {
				return []Location{*loc}, nil
			}
		}
	}

	// For Go expressions, try gopls
	if ctx.InGoExpr {
		locs, err := d.getGoplsDefinition(ctx)
		if err != nil {
			log.Server("gopls definition error: %v", err)
		} else if len(locs) > 0 {
			return locs, nil
		}
	}

	if word == "" {
		return nil, nil
	}

	// Dispatch based on node kind
	switch ctx.NodeKind {
	case NodeKindComponentCall:
		return d.definitionComponentCall(ctx)
	case NodeKindNamedRef:
		return d.definitionNamedRef(ctx)
	case NodeKindStateAccess, NodeKindStateDecl:
		return d.definitionStateVar(ctx)
	case NodeKindEventHandler:
		return d.definitionEventHandler(ctx)
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
	if ctx.Scope != nil && ctx.Scope.Component != nil {
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

func (d *definitionProvider) definitionNamedRef(ctx *CursorContext) ([]Location, error) {
	elem, ok := ctx.Node.(*tuigen.Element)
	if !ok || elem == nil {
		return nil, nil
	}

	// The definition of a named ref is the element itself (where #Name is declared)
	return []Location{{
		URI: ctx.Document.URI,
		Range: Range{
			Start: Position{Line: elem.Position.Line - 1, Character: elem.Position.Column - 1},
			End:   Position{Line: elem.Position.Line - 1, Character: elem.Position.Column - 1 + len(elem.Tag)},
		},
	}}, nil
}

func (d *definitionProvider) definitionStateVar(ctx *CursorContext) ([]Location, error) {
	if ctx.Scope == nil {
		return nil, nil
	}

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

// --- Local variable definitions ---

func (d *definitionProvider) findLetBindingDefinition(ctx *CursorContext, varName string) *Location {
	if ctx.Document.AST == nil || ctx.Scope == nil || ctx.Scope.Component == nil {
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
	if ctx.Document.AST == nil || ctx.Scope == nil || ctx.Scope.Component == nil {
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
	if ctx.Document.AST == nil || ctx.Scope == nil || ctx.Scope.Component == nil {
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
				if startFound || endFound {
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

// findVarDeclPosition finds the position of a variable declaration in code.
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
			partStart := strings.Index(lhs, varName)
			if partStart >= 0 {
				return varIdx + 4 + partStart
			}
		}
	}

	return -1
}

// parseFuncName extracts the function name from a Go function definition.
func parseFuncName(code string) string {
	// Simple extraction: "func Name(" -> "Name"
	idx := strings.Index(code, "func ")
	if idx < 0 {
		return ""
	}
	rest := code[idx+5:]
	parenIdx := strings.Index(rest, "(")
	if parenIdx < 0 {
		return ""
	}
	return strings.TrimSpace(rest[:parenIdx])
}
