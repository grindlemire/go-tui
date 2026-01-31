package provider

import (
	"github.com/grindlemire/go-tui/internal/lsp/gopls"
	"github.com/grindlemire/go-tui/internal/tuigen"
)

// --- Shared test helpers ---

func parseTestDoc(src string) *Document {
	doc := &Document{
		URI:     "file:///test.gsx",
		Content: src,
		Version: 1,
	}
	lexer := tuigen.NewLexer("test.gsx", src)
	parser := tuigen.NewParser(lexer)
	ast, _ := parser.ParseFile()
	doc.AST = ast
	return doc
}

func makeCtx(doc *Document, nodeKind NodeKind, word string) *CursorContext {
	return &CursorContext{
		Document: doc,
		NodeKind: nodeKind,
		Word:     word,
		Scope:    &Scope{},
	}
}

// stubIndex implements ComponentIndex for testing.
type stubIndex struct {
	components map[string]*ComponentInfo
	functions  map[string]*FuncInfo
	params     map[string]*ParamInfo
}

func newStubIndex() *stubIndex {
	return &stubIndex{
		components: make(map[string]*ComponentInfo),
		functions:  make(map[string]*FuncInfo),
		params:     make(map[string]*ParamInfo),
	}
}

func (s *stubIndex) Lookup(name string) (*ComponentInfo, bool) {
	info, ok := s.components[name]
	return info, ok
}

func (s *stubIndex) LookupFunc(name string) (*FuncInfo, bool) {
	info, ok := s.functions[name]
	return info, ok
}

func (s *stubIndex) LookupParam(componentName, paramName string) (*ParamInfo, bool) {
	key := componentName + "." + paramName
	info, ok := s.params[key]
	return info, ok
}

func (s *stubIndex) LookupFuncParam(funcName, paramName string) (*FuncParamInfo, bool) {
	return nil, false
}

func (s *stubIndex) All() []string {
	names := make([]string, 0, len(s.components))
	for name := range s.components {
		names = append(names, name)
	}
	return names
}

func (s *stubIndex) AllFunctions() []string {
	names := make([]string, 0, len(s.functions))
	for name := range s.functions {
		names = append(names, name)
	}
	return names
}

// nilGoplsProxy implements GoplsProxyAccessor returning nil.
type nilGoplsProxy struct{}

func (n *nilGoplsProxy) GetProxy() *gopls.GoplsProxy { return nil }

// nilVirtualFiles implements VirtualFileAccessor returning nil.
type nilVirtualFiles struct{}

func (n *nilVirtualFiles) GetVirtualFile(uri string) *gopls.CachedVirtualFile { return nil }

// stubDocAccessor implements DocumentAccessor for testing.
type stubDocAccessor struct {
	docs []*Document
}

func (s *stubDocAccessor) GetDocument(uri string) *Document {
	for _, d := range s.docs {
		if d.URI == uri {
			return d
		}
	}
	return nil
}

func (s *stubDocAccessor) AllDocuments() []*Document {
	return s.docs
}

// stubWorkspaceAST implements WorkspaceASTAccessor for testing.
type stubWorkspaceAST struct {
	asts map[string]*tuigen.File
}

func (s *stubWorkspaceAST) GetWorkspaceAST(uri string) *tuigen.File {
	return s.asts[uri]
}

func (s *stubWorkspaceAST) AllWorkspaceASTs() map[string]*tuigen.File {
	return s.asts
}
