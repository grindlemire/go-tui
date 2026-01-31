# GSX Language Server

The LSP implementation for `.gsx` files, providing real-time editor intelligence including hover, completion, go-to-definition, find-references, diagnostics, formatting, semantic tokens, and document/workspace symbols.

## Architecture Overview

The LSP is organized as a two-package system connected by adapters:

```
Editor (VSCode, etc.)
    │
    │  JSON-RPC 2.0 over stdio
    ▼
┌──────────────────────────────────────────────────┐
│  pkg/lsp/  (server shell)                        │
│                                                  │
│  server.go     ─ Server struct, Run() loop,      │
│                  readMessage/writeMessage         │
│  router.go     ─ Route() dispatch, CursorContext │
│                  resolution, provider delegation  │
│  handler.go    ─ Lifecycle (init/shutdown) and   │
│                  document sync (open/change/close)│
│  document.go   ─ DocumentManager, parse on edit  │
│  context.go    ─ ResolveCursorContext, AST walk,  │
│                  scope collection, text heuristics│
│  providers.go  ─ Provider interfaces + Registry  │
│  provider_adapters.go ─ Bridges lsp ↔ provider   │
│  index.go      ─ ComponentIndex (workspace-wide)  │
│                                                  │
│  gopls/        ─ gopls proxy + virtual file gen  │
│  schema/       ─ Element/attribute/keyword defs  │
│  log/          ─ Structured debug logging        │
└──────────────┬───────────────────────────────────┘
               │ adapters convert
               │ CursorContext & Document
               ▼
┌──────────────────────────────────────────────────┐
│  pkg/lsp/provider/  (feature logic)              │
│                                                  │
│  provider.go    ─ Types, interfaces, NodeKind    │
│  hover.go       ─ Hover documentation            │
│  completion.go  ─ Completion suggestions         │
│  definition.go  ─ Go-to-definition               │
│  references.go  ─ Find all references            │
│  diagnostics.go ─ Error/warning diagnostics      │
│  semantic.go    ─ Semantic token highlighting    │
│  symbols.go     ─ Document & workspace symbols   │
│  formatting.go  ─ Document formatting            │
└──────────────────────────────────────────────────┘
```

## Request Lifecycle

Every LSP request flows through 6 phases:

### Phase 1: Startup and Capability Negotiation

```
Editor                          Server
  │                               │
  │─── initialize ───────────────>│  Store rootURI
  │<── capabilities ──────────────│  Advertise: hover, completion,
  │                               │  definition, references, symbols,
  │─── initialized ──────────────>│  formatting, semanticTokens
  │                               │
  │                               ├── go indexWorkspace()
  │                               │   Walk rootURI for *.gsx files,
  │                               │   parse each, populate ComponentIndex
  │                               │
  │                               └── go InitGopls()
  │                                   Spawn gopls subprocess,
  │                                   initialize over JSON-RPC
```

The server advertises full-document sync (`TextDocumentSyncKindFull`). On `initialized`, two background goroutines start: workspace indexing (for cross-file component/function lookups) and gopls proxy initialization (for Go expression intelligence).

### Phase 2: Document Lifecycle

```
didOpen / didChange / didSave
         │
         ▼
  ┌──────────────┐
  │ DocumentManager │
  │   .Open()       │  Full content stored
  │   .Update()     │  Re-parsed on every change
  │   .Close()      │
  └──────┬─────────┘
         │
         ├──> Lexer → Parser → AST (tuigen.File)
         │    Parse errors stored on Document.Errors
         │
         ├──> Analyzer.Analyze(ast)
         │    Semantic errors (e.g. invalid Tailwind classes)
         │    appended to Document.Errors
         │
         ├──> ComponentIndex.IndexDocument(uri, ast)
         │    Components, functions, params registered
         │    for workspace-wide lookup
         │
         ├──> UpdateVirtualFile(doc)
         │    GenerateVirtualGo(ast) → .go content + SourceMap
         │    Notify gopls of the virtual file
         │
         └──> publishDiagnostics(doc)
              DiagnosticsProvider.Diagnose() → editor
```

Every keystroke triggers a full re-parse, re-index, virtual file regeneration, and diagnostic publish. The `DocumentManager` holds a `map[string]*Document` keyed by URI. When a file is closed, its AST moves to `workspaceASTs` so cross-file lookups still work.

### Phase 3: Request Routing

```
JSON-RPC message
     │
     ▼
  Server.handleMessage()
     │ json.Unmarshal → Request
     ▼
  Router.Route(req)
     │
     ├── Lifecycle methods → Server.handle*() directly
     │   (initialize, initialized, shutdown, exit)
     │
     ├── Document sync → Server.handle*() directly
     │   (didOpen, didChange, didClose, didSave)
     │
     └── Language features → Provider dispatch
         (hover, completion, definition, references,
          documentSymbol, workspaceSymbol, formatting,
          semanticTokens/full)
```

The `Router` splits requests into three categories. Lifecycle and document sync are handled by the `Server` directly. Language feature requests go through provider dispatch, which involves resolving a `CursorContext` first.

### Phase 4: CursorContext Resolution

This is the core of the LSP. Every position-based request (hover, completion, definition, references) resolves a `CursorContext` before any provider logic runs.

```
ResolveCursorContext(doc, position)
     │
     ├── 1. Compute byte offset from line:character
     │
     ├── 2. Extract line text and word under cursor
     │      (includes hyphens for Tailwind, @ for keywords, # for refs)
     │
     ├── 3. Check text-level context flags:
     │      InGoExpr    ─ backwards brace counting for {...}
     │      InClassAttr ─ backwards search for class="..."
     │      InElement   ─ backwards search for < vs >
     │
     ├── 4. If no AST → classifyFromText() and return
     │
     └── 5. Walk AST (resolveFromAST):
           │
           ├── Check component declaration lines
           │   (component name → NodeKindComponent,
           │    param name → NodeKindParameter)
           │
           ├── Find enclosing component by position
           │   Verify cursor is inside component body
           │   Set Scope.Component, Scope.Params
           │
           ├── collectScopeFromBody():
           │   Walk body recursively collecting:
           │   ├── NamedRefs (with InLoop/InConditional flags)
           │   ├── StateVars (via DetectStateVars on first GoCode w/ tui.NewState)
           │   ├── LetBindings
           │   └── ForLoop/IfStmt nesting
           │
           ├── resolveInNodes() → resolveInNode() → resolveInNodeInner()
           │   Dispatch on AST node type:
           │   ├── Element     → tag, #ref, attributes, event handlers
           │   ├── ForLoop     → loop header, body children
           │   ├── IfStmt      → condition, then/else branches
           │   ├── LetBinding  → variable name, element children
           │   ├── ComponentCall → call name, children
           │   ├── GoExpr      → classifyGoExpr() for StateAccess
           │   ├── GoCode      → classifyGoCode() for StateDecl/StateAccess
           │   └── TextContent → NodeKindText
           │
           └── Fallback: check functions, then classifyFromText()
```

The `CursorContext` struct contains everything a provider needs:

| Field | Description |
|-------|-------------|
| `Document` | The open document (content, AST, errors) |
| `Position` | 0-indexed line:character |
| `Offset` | Byte offset in content |
| `Node` | The resolved AST node (may be nil) |
| `NodeKind` | One of 17 classifications (see below) |
| `Scope` | Enclosing component, function, for loop, if stmt, named refs, state vars, let bindings, params |
| `ParentChain` | Path from root to current node |
| `Word` | Word under cursor (hyphen-aware) |
| `Line` | Full line text |
| `InGoExpr` | Inside a `{...}` Go expression |
| `InClassAttr` | Inside `class="..."` |
| `InElement` | Inside an element tag `<...>` |
| `AttrTag` | Element tag when on an attribute |
| `AttrName` | Attribute name when on an attribute |

### NodeKind Classifications

Every cursor position is classified into one of these kinds, which drives dispatch in every provider:

| NodeKind | What it represents |
|----------|--------------------|
| `Component` | `templ Name(...)` declaration line, on the name |
| `Element` | HTML-like element tag (`<div>`, `<span>`, etc.) |
| `Attribute` | Element attribute name (`class`, `id`, etc.) |
| `NamedRef` | `#Name` reference on an element |
| `GoExpr` | Go expression inside `{...}` |
| `ForLoop` | `@for` loop header |
| `IfStmt` | `@if` conditional header |
| `LetBinding` | `@let` variable binding |
| `StateDecl` | `tui.NewState(...)` declaration |
| `StateAccess` | `.Get()`, `.Set()`, `.Update()`, `.Bind()`, `.Batch()` |
| `Parameter` | Component parameter on the declaration line |
| `Function` | `func` declaration line |
| `ComponentCall` | `@Component(args)` call |
| `EventHandler` | Event handler attribute (e.g. `onClick`) |
| `Text` | Plain text content |
| `Keyword` | Language keywords (`templ`, `@for`, `@if`, `@else`, `@let`) |
| `TailwindClass` | Class name inside `class="..."` |

### Phase 5: Provider Decision Logic

Each provider receives the `CursorContext` and switches on `NodeKind`:

**Hover Provider** (`hover.go`):
```
NodeKind → Action
─────────────────────────────────────────────────────
Component     → Show component signature (func Name(params) *element.Element)
Element       → Show element description + available attributes from schema
Attribute     → Show attribute type + description from schema
EventHandler  → Show event handler signature + description
Parameter     → Show parameter name, type, and owning component
Keyword       → Show keyword documentation from schema
ForLoop       → Show keyword documentation
IfStmt        → Show keyword documentation
LetBinding    → Show keyword documentation
Function      → Show function signature from index
ComponentCall → Show component signature from index
NamedRef      → Show ref type (simple, slice, map) + access pattern
StateDecl     → Show state variable type, initial value, available methods
StateAccess   → Show specific state method documentation
TailwindClass → Show class documentation from schema
GoExpr        → Delegate to gopls via virtual file + SourceMap
```

**Completion Provider** (`completion.go`):
```
Context → Completions offered
─────────────────────────────────────────────────────
InClassAttr   → Tailwind class names matching prefix
InGoExpr      → gopls completions via virtual file, state var methods
InElement     → Attribute names for current tag from schema
@ prefix      → Component names from index + keywords
< prefix      → Element tags from schema
Default       → Components, functions, keywords
```

**Definition Provider** (`definition.go`):
```
NodeKind → Jump target
─────────────────────────────────────────────────────
ComponentCall → Component declaration location (from index)
Component     → Self (declaration line)
Parameter     → Parameter position on declaration line
Function      → Function declaration location
NamedRef      → Element with the #Name ref
LetBinding    → Let declaration line
StateDecl     → State variable declaration line
GoExpr        → gopls definition via virtual file + SourceMap
```

**References Provider** (`references.go`):
```
NodeKind → Search scope
─────────────────────────────────────────────────────
Component/ComponentCall → All @Name calls across open docs + workspace ASTs
Parameter               → Parameter declaration + usages in component body
NamedRef                → #Name declaration + view.Name usages
StateDecl/StateAccess   → Declaration + .Get()/.Set()/etc. usages
LetBinding              → Declaration + usages in component body
Function                → Function declaration + calls across workspace
```

### Phase 6: gopls Bridge

For Go expressions inside `{...}`, the LSP delegates to a real gopls instance:

```
.gsx file                    Virtual .go file              gopls
─────────                    ────────────────              ─────
templ Counter(n int) {       func Counter(n int)           Hover/Complete/
  <span>{n + 1}</span>  ──>   *element.Element {          Definition on
}                              _ = n + 1          ──────> generated Go
                               return nil
                             }

Position translation:
.gsx line:col ──[SourceMap.TuiToGo()]──> .go line:col ──> gopls
.gsx line:col <──[SourceMap.GoToTui()]── .go line:col <── gopls result
```

**How virtual files are generated** (`gopls/generate.go`):

1. Package declaration and imports are copied as-is
2. Each `templ` component becomes a Go function with the same signature
3. State declarations (`tui.NewState(...)`) are emitted as Go variable declarations
4. Named refs (`#Name`) are emitted as typed variable declarations
5. Go expressions (`{expr}`) become `_ = expr` assignments
6. For loops, if statements, and let bindings map to their Go equivalents
7. Component calls become `_ = Name(args)` assignments

Every generated construct has a `SourceMap` entry recording the bidirectional mapping between `.gsx` and `.go` positions. The `SourceMap` uses `Mapping` structs:

```go
type Mapping struct {
    TuiLine, TuiCol int  // 0-indexed position in .gsx
    GoLine,  GoCol  int  // 0-indexed position in .go
    Length          int  // length of the mapped region
}
```

## Package Dependency Graph

```
pkg/lsp/provider/
    │
    ├── depends on ──> pkg/lsp/gopls/     (GoplsProxy, CachedVirtualFile, SourceMap)
    ├── depends on ──> pkg/lsp/schema/    (elements, attributes, keywords, tailwind)
    ├── depends on ──> pkg/tuigen/        (AST node types)
    └── NO dependency on pkg/lsp/ (avoids circular imports)

pkg/lsp/
    │
    ├── depends on ──> pkg/lsp/provider/  (provider interfaces, type aliases)
    ├── depends on ──> pkg/lsp/gopls/     (proxy, virtual file cache)
    ├── depends on ──> pkg/lsp/log/       (debug logging)
    └── depends on ──> pkg/tuigen/        (lexer, parser, AST, analyzer)
```

The two packages share structurally identical types (`CursorContext`, `Document`, `Scope`, `NodeKind`) to avoid circular imports. The adapter layer in `provider_adapters.go` converts between them:

```
lsp.CursorContext ──[CursorContextToProvider()]──> provider.CursorContext
lsp.Document      ──[convertDocument()]──────────> provider.Document
```

Protocol types (`Position`, `Range`, `Location`, `Hover`, etc.) are defined once in `provider/provider.go` and aliased in `lsp/document.go`, so no conversion is needed for these.

## Sub-packages

### `gopls/`

- **`proxy.go`** - Manages a gopls subprocess over JSON-RPC. Provides `Hover()`, `Complete()`, `Definition()` methods. Spawns gopls with `cmd/gopls serve`, communicates over stdin/stdout.
- **`generate.go`** - Transforms `.gsx` ASTs into valid Go source files that gopls can analyze. Handles parameter mapping, state variable declarations, named ref declarations, expression mapping, control flow mapping.
- **`mapping.go`** - `SourceMap` for bidirectional `.gsx` <-> `.go` position translation. `VirtualFileCache` stores generated content and maps keyed by `.gsx` URI.

### `schema/`

- **`schema.go`** - Element definitions (`<div>`, `<span>`, `<button>`, etc.), attribute definitions (per-element and global), event handler definitions.
- **`keywords.go`** - Language keyword definitions (`templ`, `@for`, `@if`, `@else`, `@let`) with documentation.
- **`tailwind.go`** - Tailwind-style class definitions, parameterized patterns (e.g. `p-N`, `text-COLOR`), class documentation for hover, class matching for completion.

### `log/`

- **`log.go`** - Structured logging with categories (Server, Generate, Mapping) that can be enabled/disabled. Writes to a configurable output file for debugging.

## Key Design Decisions

1. **Full-document sync**: Every edit sends the complete document content. Simpler than incremental sync, and `.gsx` files are small enough that re-parsing the entire file on each keystroke is fast.

2. **CursorContext as universal currency**: All provider logic receives a pre-resolved context instead of raw positions. This centralizes the complex "what's under the cursor" logic and ensures consistency across all features.

3. **Separate provider package**: Feature logic lives in `provider/` with no dependency on the server. This enables unit testing providers with mock contexts without needing a running server.

4. **Non-fatal gopls**: The gopls proxy is optional. If it fails to start or crashes, all GSX-native features continue working. Only Go expression intelligence is lost.

5. **Workspace AST caching**: When a file is closed in the editor, its last-known AST moves to `workspaceASTs` so cross-file lookups (find references, workspace symbols) still work.
