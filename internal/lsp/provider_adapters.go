package lsp

import (
	"github.com/grindlemire/go-tui/internal/lsp/gopls"
	"github.com/grindlemire/go-tui/internal/lsp/provider"
	"github.com/grindlemire/go-tui/internal/tuigen"
)

// --- Adapters that implement provider.* interfaces using lsp.Server internals ---

// componentIndexAdapter adapts *ComponentIndex to provider.ComponentIndex.
type componentIndexAdapter struct {
	index *ComponentIndex
}

func (a *componentIndexAdapter) Lookup(name string) (*provider.ComponentInfo, bool) {
	info, ok := a.index.Lookup(name)
	if !ok || info == nil {
		return nil, false
	}
	return &provider.ComponentInfo{
		Name:     info.Name,
		Params:   info.Params,
		Location: info.Location,
	}, true
}

func (a *componentIndexAdapter) LookupFunc(name string) (*provider.FuncInfo, bool) {
	info, ok := a.index.LookupFunc(name)
	if !ok || info == nil {
		return nil, false
	}
	return &provider.FuncInfo{
		Name:      info.Name,
		Signature: info.Signature,
		Returns:   info.Returns,
		Location:  info.Location,
	}, true
}

func (a *componentIndexAdapter) LookupParam(componentName, paramName string) (*provider.ParamInfo, bool) {
	info, ok := a.index.LookupParam(componentName, paramName)
	if !ok || info == nil {
		return nil, false
	}
	return &provider.ParamInfo{
		Name:          info.Name,
		Type:          info.Type,
		ComponentName: info.ComponentName,
		Location:      info.Location,
	}, true
}

func (a *componentIndexAdapter) LookupFuncParam(funcName, paramName string) (*provider.FuncParamInfo, bool) {
	param, uri, ok := a.index.LookupFuncParam(funcName, paramName)
	if !ok || param == nil {
		return nil, false
	}
	return &provider.FuncParamInfo{
		Name:     param.Name,
		Type:     param.Type,
		FuncName: funcName,
		Location: provider.Location{
			URI: uri,
			Range: provider.Range{
				Start: param.Position,
				End:   provider.Position{Line: param.Position.Line, Character: param.Position.Character + len(param.Name)},
			},
		},
	}, true
}

func (a *componentIndexAdapter) All() []string {
	return a.index.All()
}

func (a *componentIndexAdapter) AllFunctions() []string {
	return a.index.AllFunctions()
}

// goplsProxyAdapter adapts *Server to provider.GoplsProxyAccessor.
type goplsProxyAdapter struct {
	server *Server
}

func (a *goplsProxyAdapter) GetProxy() *gopls.GoplsProxy {
	return a.server.goplsProxy
}

// virtualFileAdapter adapts *Server to provider.VirtualFileAccessor.
type virtualFileAdapter struct {
	server *Server
}

func (a *virtualFileAdapter) GetVirtualFile(uri string) *gopls.CachedVirtualFile {
	return a.server.virtualFiles.Get(uri)
}

// documentAdapter adapts *Server to provider.DocumentAccessor.
type documentAdapter struct {
	server *Server
}

func (a *documentAdapter) GetDocument(uri string) *provider.Document {
	doc := a.server.docs.Get(uri)
	if doc == nil {
		return nil
	}
	return convertDocument(doc)
}

func (a *documentAdapter) AllDocuments() []*provider.Document {
	docs := a.server.docs.All()
	result := make([]*provider.Document, len(docs))
	for i, d := range docs {
		result[i] = convertDocument(d)
	}
	return result
}

// workspaceASTAdapter adapts *Server to provider.WorkspaceASTAccessor.
type workspaceASTAdapter struct {
	server *Server
}

func (a *workspaceASTAdapter) GetWorkspaceAST(uri string) *tuigen.File {
	a.server.workspaceASTsMu.RLock()
	defer a.server.workspaceASTsMu.RUnlock()
	return a.server.workspaceASTs[uri]
}

func (a *workspaceASTAdapter) AllWorkspaceASTs() map[string]*tuigen.File {
	a.server.workspaceASTsMu.RLock()
	defer a.server.workspaceASTsMu.RUnlock()

	// Return a copy to avoid holding the lock
	result := make(map[string]*tuigen.File, len(a.server.workspaceASTs))
	for k, v := range a.server.workspaceASTs {
		result[k] = v
	}
	return result
}

// --- Conversion helpers ---
// Only Document and CursorContext need conversion since they are separate types
// in the lsp and provider packages (different roles). Protocol types (Position,
// Range, Location, Hover, etc.) are type aliases and need no conversion.

func convertDocument(d *Document) *provider.Document {
	return &provider.Document{
		URI:     d.URI,
		Content: d.Content,
		Version: d.Version,
		AST:     d.AST,
		Errors:  d.Errors,
	}
}

// CursorContextToProvider converts an lsp.CursorContext to a provider.CursorContext.
func CursorContextToProvider(ctx *CursorContext) *provider.CursorContext {
	var scope *provider.Scope
	if ctx.Scope != nil {
		scope = &provider.Scope{
			Component: ctx.Scope.Component,
			Function:  ctx.Scope.Function,
			ForLoop:   ctx.Scope.ForLoop,
			IfStmt:    ctx.Scope.IfStmt,
			Refs:      ctx.Scope.Refs,
			StateVars: ctx.Scope.StateVars,
			LetBinds:  ctx.Scope.LetBinds,
			Params:    ctx.Scope.Params,
		}
	}

	return &provider.CursorContext{
		Document:    convertDocument(ctx.Document),
		Position:    ctx.Position,
		Offset:      ctx.Offset,
		Node:        ctx.Node,
		NodeKind:    provider.NodeKind(ctx.NodeKind),
		Scope:       scope,
		ParentChain: ctx.ParentChain,
		Word:        ctx.Word,
		Line:        ctx.Line,
		InGoExpr:    ctx.InGoExpr,
		InClassAttr: ctx.InClassAttr,
		InElement:   ctx.InElement,
		AttrTag:     ctx.AttrTag,
		AttrName:    ctx.AttrName,
	}
}

// --- Provider construction ---

// CreateProviderRegistry creates the provider registry with all implemented providers.
func (s *Server) CreateProviderRegistry() *Registry {
	indexAdapter := &componentIndexAdapter{index: s.index}
	proxyAdapter := &goplsProxyAdapter{server: s}
	vfAdapter := &virtualFileAdapter{server: s}
	docsAdapter := &documentAdapter{server: s}
	workspaceAdapter := &workspaceASTAdapter{server: s}

	return &Registry{
		Hover:           newHoverProviderAdapter(provider.NewHoverProvider(indexAdapter, proxyAdapter, vfAdapter)),
		Completion:      newCompletionProviderAdapter(provider.NewCompletionProvider(indexAdapter, proxyAdapter, vfAdapter)),
		Definition:      newDefinitionProviderAdapter(provider.NewDefinitionProvider(indexAdapter, proxyAdapter, vfAdapter, docsAdapter)),
		References:      newReferencesProviderAdapter(provider.NewReferencesProvider(indexAdapter, docsAdapter, workspaceAdapter)),
		DocumentSymbol:  newDocumentSymbolProviderAdapter(provider.NewDocumentSymbolProvider()),
		WorkspaceSymbol: newWorkspaceSymbolProviderAdapter(provider.NewWorkspaceSymbolProvider(indexAdapter)),
		Diagnostics:     newDiagnosticsProviderAdapter(provider.NewDiagnosticsProvider()),
		Formatting:      newFormattingProviderAdapter(provider.NewFormattingProvider()),
		SemanticTokens:  newSemanticTokensProviderAdapter(provider.NewSemanticTokensProvider(&functionNameCheckerAdapter{server: s}, docsAdapter)),
	}
}

// --- Provider wrapper adapters ---
// These wrap provider.* implementations to satisfy lsp.* interfaces,
// handling the CursorContext and Document conversions. Protocol types
// (Position, Range, Location, Hover, etc.) are type aliases, so no
// conversion is needed for them.

type hoverProviderAdapter struct {
	inner provider.HoverProvider
}

func newHoverProviderAdapter(p provider.HoverProvider) HoverProvider {
	return &hoverProviderAdapter{inner: p}
}

func (a *hoverProviderAdapter) Hover(ctx *CursorContext) (*Hover, error) {
	pCtx := CursorContextToProvider(ctx)
	return a.inner.Hover(pCtx)
}

type definitionProviderAdapter struct {
	inner provider.DefinitionProvider
}

func newDefinitionProviderAdapter(p provider.DefinitionProvider) DefinitionProvider {
	return &definitionProviderAdapter{inner: p}
}

func (a *definitionProviderAdapter) Definition(ctx *CursorContext) ([]Location, error) {
	pCtx := CursorContextToProvider(ctx)
	return a.inner.Definition(pCtx)
}

type referencesProviderAdapter struct {
	inner provider.ReferencesProvider
}

func newReferencesProviderAdapter(p provider.ReferencesProvider) ReferencesProvider {
	return &referencesProviderAdapter{inner: p}
}

func (a *referencesProviderAdapter) References(ctx *CursorContext, includeDecl bool) ([]Location, error) {
	pCtx := CursorContextToProvider(ctx)
	return a.inner.References(pCtx, includeDecl)
}

type completionProviderAdapter struct {
	inner provider.CompletionProvider
}

func newCompletionProviderAdapter(p provider.CompletionProvider) CompletionProvider {
	return &completionProviderAdapter{inner: p}
}

func (a *completionProviderAdapter) Complete(ctx *CursorContext) (*CompletionList, error) {
	pCtx := CursorContextToProvider(ctx)
	return a.inner.Complete(pCtx)
}

type documentSymbolProviderAdapter struct {
	inner provider.DocumentSymbolProvider
}

func newDocumentSymbolProviderAdapter(p provider.DocumentSymbolProvider) DocumentSymbolProvider {
	return &documentSymbolProviderAdapter{inner: p}
}

func (a *documentSymbolProviderAdapter) DocumentSymbols(doc *Document) ([]DocumentSymbol, error) {
	pDoc := convertDocument(doc)
	return a.inner.DocumentSymbols(pDoc)
}

type workspaceSymbolProviderAdapter struct {
	inner provider.WorkspaceSymbolProvider
}

func newWorkspaceSymbolProviderAdapter(p provider.WorkspaceSymbolProvider) WorkspaceSymbolProvider {
	return &workspaceSymbolProviderAdapter{inner: p}
}

func (a *workspaceSymbolProviderAdapter) WorkspaceSymbols(query string) ([]SymbolInformation, error) {
	return a.inner.WorkspaceSymbols(query)
}

// --- Diagnostics provider adapter ---

type diagnosticsProviderAdapter struct {
	inner provider.DiagnosticsProvider
}

func newDiagnosticsProviderAdapter(p provider.DiagnosticsProvider) DiagnosticsProvider {
	return &diagnosticsProviderAdapter{inner: p}
}

func (a *diagnosticsProviderAdapter) Diagnose(doc *Document) ([]Diagnostic, error) {
	pDoc := convertDocument(doc)
	return a.inner.Diagnose(pDoc)
}

// --- Formatting provider adapter ---

type formattingProviderAdapter struct {
	inner provider.FormattingProvider
}

func newFormattingProviderAdapter(p provider.FormattingProvider) FormattingProvider {
	return &formattingProviderAdapter{inner: p}
}

func (a *formattingProviderAdapter) Format(doc *Document, opts FormattingOptions) ([]TextEdit, error) {
	pDoc := convertDocument(doc)
	return a.inner.Format(pDoc, opts)
}

// --- Semantic tokens provider adapter ---

type semanticTokensProviderAdapter struct {
	inner provider.SemanticTokensProvider
}

func newSemanticTokensProviderAdapter(p provider.SemanticTokensProvider) SemanticTokensProvider {
	return &semanticTokensProviderAdapter{inner: p}
}

func (a *semanticTokensProviderAdapter) SemanticTokensFull(doc *Document) (*SemanticTokens, error) {
	pDoc := convertDocument(doc)
	result, err := a.inner.SemanticTokensFull(pDoc)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return &SemanticTokens{Data: []int{}}, nil
	}
	return result, nil
}

// --- FunctionNameChecker adapter ---
// Bridges provider.FunctionNameChecker to the server's ComponentIndex.

type functionNameCheckerAdapter struct {
	server *Server
}

// goBuiltinFunctions is the set of Go built-in function names used for
// semantic token classification. Defined at package level to avoid
// reconstructing the map on every IsFunctionName call.
var goBuiltinFunctions = map[string]bool{
	"len": true, "cap": true, "make": true, "new": true,
	"append": true, "copy": true, "delete": true,
	"close": true, "panic": true, "recover": true,
	"print": true, "println": true,
	"real": true, "imag": true, "complex": true,
}

func (a *functionNameCheckerAdapter) IsFunctionName(name string) bool {
	// Check indexed functions
	if _, ok := a.server.index.LookupFunc(name); ok {
		return true
	}

	return goBuiltinFunctions[name]
}
