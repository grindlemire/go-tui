# Developer Tooling Overhaul Specification

**Status:** Planned\
**Version:** 1.0\
**Last Updated:** 2025-01-30

---

## 1. Overview

### Purpose

The developer tooling (LSP server, VSCode extension, tree-sitter grammar) has fallen behind the current state of the `.gsx` language after the addition of named references (`#Name`), reactive state (`tui.NewState`), event handlers, and the rename from the old syntax to `templ`-based declarations. This overhaul rearchitects the LSP for maintainability and extensibility, brings all tooling up to date with current language features, and fixes stale references across the VSCode extension.

### Goals

- Rearchitect the LSP with a Provider + CursorContext pattern so adding new language constructs doesn't require touching 8+ files
- Centralize language construct knowledge (elements, attributes, keywords, tailwind classes) into a shared schema
- Add full LSP awareness of named refs, reactive state, event handlers, and watchers
- Update the VSCode extension (TextMate grammar, language config, README) for current syntax
- Update the tree-sitter grammar and queries for completeness
- Maintain all existing LSP capabilities (diagnostics, completion, hover, definition, references, symbols, formatting, semantic tokens)

### Non-Goals

- Adding new LSP features that don't exist today (code actions, rename, signature help)
- Switching VSCode from TextMate to tree-sitter (VSCode doesn't support it natively)
- Rewriting the gopls proxy from scratch (it works well; we'll extend it)
- Changing the parser/analyzer/generator in `pkg/tuigen/` (those are out of scope)

---

## 2. Architecture

### Current Architecture Problems

The existing LSP (~7,400 lines across 16 files) is feature-rich but has accumulated structural debt:

1. **Scattered construct knowledge** — Each feature file (hover.go, completion.go, semantic_tokens.go, etc.) independently reimplements "what is the cursor pointing at?" with its own position math and AST traversal. Adding a new language construct means updating 6-8 files.

2. **Hardcoded schemas** — Element definitions, attribute lists, tailwind classes, and keyword documentation are hardcoded inline within each feature file. The same element list appears in completion.go, hover.go, and semantic_tokens.go with different levels of detail.

3. **No construct awareness for new features** — Named refs (`#Name`), reactive state (`tui.NewState()`, `.Get()`, `.Set()`, `.Bind()`), event handler attributes (`onClick`, `onFocus`, etc.), and watchers are not represented in hover, completion, definition, references, or semantic tokens.

4. **Stale naming** — The VSCode extension README still documents `<box>`/`<text>` elements and `@component` keyword. The language-configuration.json folding patterns reference `@component`. Named ref TextMate scopes use `.tui` suffix instead of `.gsx`.

5. **gopls virtual file gaps** — The virtual Go generator (`gopls/generate.go`) doesn't emit state variable declarations or named ref variables, limiting Go code intelligence for expressions that use them.

### Proposed Architecture

```
pkg/lsp/
├── server.go              # Core server, JSON-RPC I/O, lifecycle
├── router.go              # Method routing with provider dispatch
├── context.go             # CursorContext resolver
├── document.go            # Document manager (parse, sync, position conversion)
├── index.go               # Workspace symbol index
│
├── schema/                # Centralized language knowledge
│   ├── schema.go          # Elements, attributes, type definitions
│   ├── keywords.go        # DSL keywords and documentation
│   └── tailwind.go        # Tailwind class definitions and docs
│
├── provider/              # LSP feature providers
│   ├── provider.go        # Provider interfaces and registry
│   ├── hover.go           # Hover provider
│   ├── completion.go      # Completion provider
│   ├── definition.go      # Definition provider
│   ├── references.go      # References provider
│   ├── symbols.go         # Document + workspace symbol providers
│   ├── diagnostics.go     # Diagnostics provider
│   ├── formatting.go      # Formatting provider
│   └── semantic.go        # Semantic tokens provider
│
├── gopls/                 # Gopls integration (extended)
│   ├── proxy.go           # Subprocess communication
│   ├── generate.go        # Virtual Go generation (updated for refs/state)
│   └── mapping.go         # Source position mapping
│
└── log/
    └── log.go             # Logging (unchanged)
```

### Key Abstractions

#### CursorContext

Centralized resolution of "what is under the cursor." Every provider receives this instead of raw positions:

```go
type CursorContext struct {
    Document    *Document
    Position    protocol.Position
    Offset      int

    // Resolved AST information
    Node        tuigen.Node       // The AST node at the cursor
    NodeKind    NodeKind          // Classification (element, ref, state, etc.)
    Scope       *Scope            // Enclosing component/function scope
    ParentChain []tuigen.Node     // Path from root to current node

    // Convenience fields
    Word        string            // Word under cursor
    Line        string            // Full line text
    InGoExpr    bool              // Inside a Go expression {…}
    InClassAttr bool              // Inside class="…"
    InElement   bool              // Inside an element tag
}

type NodeKind int
const (
    NodeKindUnknown NodeKind = iota
    NodeKindComponent
    NodeKindElement
    NodeKindAttribute
    NodeKindNamedRef
    NodeKindGoExpr
    NodeKindForLoop
    NodeKindIfStmt
    NodeKindLetBinding
    NodeKindStateDecl
    NodeKindStateAccess     // .Get(), .Set()
    NodeKindParameter
    NodeKindFunction
    NodeKindComponentCall
    NodeKindEventHandler    // onClick, onFocus, etc.
    NodeKindText
    NodeKindKeyword
    NodeKindTailwindClass
)

type Scope struct {
    Component  *tuigen.Component
    Function   *tuigen.Function
    ForLoop    *tuigen.ForLoop    // nil if not in a loop
    IfStmt     *tuigen.IfStmt     // nil if not in conditional
    StateVars  []tuigen.StateVar  // State variables in scope
    NamedRefs  []tuigen.NamedRef  // Named refs in scope
    LetBinds   []tuigen.LetBinding
    Params     []tuigen.Param
}
```

#### Provider Interfaces

Each LSP feature implements a focused interface. The router dispatches to providers:

```go
type HoverProvider interface {
    Hover(ctx *CursorContext) (*protocol.Hover, error)
}

type CompletionProvider interface {
    Complete(ctx *CursorContext) (*protocol.CompletionList, error)
}

type DefinitionProvider interface {
    Definition(ctx *CursorContext) ([]protocol.Location, error)
}

type ReferencesProvider interface {
    References(ctx *CursorContext, includeDecl bool) ([]protocol.Location, error)
}

type DocumentSymbolProvider interface {
    DocumentSymbols(doc *Document) ([]protocol.DocumentSymbol, error)
}

type WorkspaceSymbolProvider interface {
    WorkspaceSymbols(query string) ([]protocol.SymbolInformation, error)
}

type DiagnosticsProvider interface {
    Diagnose(doc *Document) ([]protocol.Diagnostic, error)
}

type FormattingProvider interface {
    Format(doc *Document, opts protocol.FormattingOptions) ([]protocol.TextEdit, error)
}

type SemanticTokensProvider interface {
    SemanticTokensFull(doc *Document) (*protocol.SemanticTokens, error)
}
```

#### Language Schema

Centralized definitions that all providers reference:

```go
// Element schema - single source of truth
type ElementDef struct {
    Tag         string
    Description string
    Attributes  []AttributeDef
    SelfClosing bool
    Category    string           // "container", "text", "input", "display"
}

type AttributeDef struct {
    Name        string
    Type        string           // "string", "int", "bool", "expression", etc.
    Description string
    Category    string           // "layout", "visual", "event", "ref"
}

// All providers import schema.Elements, schema.Keywords, schema.TailwindClasses
// instead of each having their own hardcoded lists
```

### Flow Diagram

```
┌──────────────────────────────────────────────────────────┐
│                    Editor (VSCode)                        │
│  ┌──────────────┐  ┌───────────────────────────────────┐ │
│  │ TextMate     │  │ LSP Client (extension.ts)         │ │
│  │ Grammar      │  │   → spawns tui lsp subprocess     │ │
│  └──────────────┘  └───────────────┬───────────────────┘ │
└────────────────────────────────────┼─────────────────────┘
                                     │ JSON-RPC (stdio)
                                     ▼
┌──────────────────────────────────────────────────────────┐
│                    LSP Server                             │
│  ┌─────────┐    ┌─────────┐    ┌──────────────────────┐ │
│  │ Router  │───►│ Context │───►│ Provider Registry    │ │
│  │         │    │ Resolver│    │  ├─ HoverProvider     │ │
│  │ routes  │    │         │    │  ├─ CompletionProv.   │ │
│  │ request │    │ builds  │    │  ├─ DefinitionProv.   │ │
│  │ to      │    │ Cursor  │    │  ├─ ReferencesProv.   │ │
│  │ provider│    │ Context │    │  ├─ SymbolsProv.      │ │
│  └─────────┘    └────┬────┘    │  ├─ DiagnosticsProv.  │ │
│                      │         │  ├─ FormattingProv.   │ │
│                      │         │  └─ SemanticProv.     │ │
│                      │         └──────────┬───────────┘ │
│                      │                    │              │
│         ┌────────────┼────────────────────┘              │
│         ▼            ▼                                   │
│  ┌──────────┐  ┌──────────┐  ┌────────────────────────┐ │
│  │ Language │  │ Document │  │ Gopls Proxy            │ │
│  │ Schema   │  │ Manager  │  │  ├─ generate.go        │ │
│  │ (shared) │  │ (parse,  │  │  │  (refs + state)     │ │
│  │          │  │  index)  │  │  ├─ mapping.go         │ │
│  └──────────┘  └──────────┘  │  └─ proxy.go           │ │
│                              └────────────────────────┘ │
└──────────────────────────────────────────────────────────┘
```

---

## 3. Core Entities

### New Constructs the LSP Must Understand

#### Named Refs

```gsx
<div #Header class="border p-1">...</div>           // Simple ref
@for _, item := range items {
    <span #Items>{item}</span>                       // Loop ref (slice)
    <span #Users key={user.ID}>{user.Name}</span>    // Keyed ref (map)
}
@if showWarning {
    <div #Warning>...</div>                          // Conditional ref (nullable)
}
```

The LSP must provide:
- **Hover** on `#Header` → show type (`*element.Element`), context (simple/loop/keyed/conditional)
- **Completion** after `#` → suggest existing ref patterns
- **Definition** on `Header` usage in Go expression → jump to `#Header` declaration
- **References** on `#Header` → find all usages in expressions and handlers
- **Semantic tokens** → highlight `#` as punctuation, ref name as a declaration

#### Reactive State

```gsx
templ Counter() {
    count := tui.NewState(0)
    <span>{fmt.Sprintf("%d", count.Get())}</span>
    <button onClick={increment(count)}>+</button>
}
```

The LSP must provide:
- **Hover** on `count` → show `*tui.State[int]`, initial value
- **Hover** on `.Get()` → show state accessor documentation
- **Completion** after state var → suggest `.Get()`, `.Set()`, `.Update()`, `.Bind()`, `.Batch()`
- **Definition** on `count.Get()` → jump to `count := tui.NewState(0)` declaration
- **References** on state var → find all `.Get()`, `.Set()`, handler usages
- **Semantic tokens** → highlight state declarations distinctly
- **Diagnostics** → warn on state vars declared but never bound

#### Event Handlers

```gsx
<div onClick={handleClick} onFocus={handleFocus} onKeyPress={handleKey}>
<div focusable={true} onBlur={handleBlur}>
```

The LSP must provide:
- **Completion** for event attributes → suggest `onClick`, `onFocus`, `onBlur`, `onKeyPress`, `onEvent`
- **Hover** on event attributes → show expected handler signature (`func()`)
- **Semantic tokens** → highlight event attributes distinctly from layout/visual attributes

### Gopls Virtual File Updates

The virtual Go generator must emit:

```go
// Current: only component params and basic expressions
// Updated: include state vars and ref vars

func Counter() *element.Element {
    count := tui.NewState(0)                        // State declaration
    var Header *element.Element                     // Simple ref
    var Items []*element.Element                    // Loop ref
    var Users map[string]*element.Element           // Keyed ref

    _ = fmt.Sprintf("%d", count.Get())              // Go expression
    _ = increment(count)                            // Handler expression
    return nil
}
```

This enables gopls to provide accurate completions and type information for Go expressions that reference state variables and named refs.

---

## 4. User Experience

### LSP Features After Overhaul

All existing features continue to work. New construct awareness is added:

#### Hover Examples

```
Hovering #Header:
┌──────────────────────────────────┐
│ **Named Ref** `Header`           │
│ Type: `*element.Element`         │
│ Context: Simple (direct access)  │
│                                  │
│ Access via view struct:          │
│ `view.Header`                    │
└──────────────────────────────────┘

Hovering count (state var):
┌──────────────────────────────────┐
│ **State Variable** `count`       │
│ Type: `*tui.State[int]`          │
│ Initial: `0`                     │
│                                  │
│ Methods: Get(), Set(), Update(), │
│          Bind(), Batch()         │
└──────────────────────────────────┘

Hovering onClick:
┌──────────────────────────────────┐
│ **Event Handler** `onClick`      │
│ Type: `func()`                   │
│                                  │
│ Called when the element is        │
│ clicked or activated.            │
└──────────────────────────────────┘
```

#### Completion Examples

```
Typing count. in Go expression:
  count.Get()     - Get current value
  count.Set(v)    - Set new value
  count.Update(fn)- Update with function
  count.Bind(fn)  - Register change callback
  count.Batch(fn) - Batch multiple updates

Typing on (in element attribute position):
  onClick={}      - Click/activation handler
  onFocus={}      - Focus gained handler
  onBlur={}       - Focus lost handler
  onKeyPress={}   - Key press handler
  onEvent={}      - Generic event handler

Typing # (in element tag):
  (suggests existing ref name patterns)
```

### VSCode Extension After Overhaul

- README documents current syntax (`<div>`, `<span>`, `templ`, `#Name`, state)
- TextMate grammar correctly scopes all constructs with `.gsx` suffix
- Language configuration folds on `templ` (not `@component`)
- All test files use current syntax

---

## 5. Complexity Assessment

| Size | Phases | When to Use |
|------|--------|-------------|
| Small | 1-2 | Single component, bug fix, minor enhancement |
| Medium | 3-4 | New feature touching multiple files/components |
| **Large** | **5-6** | **Cross-cutting feature, new subsystem** |

**Assessed Size:** Large\
**Recommended Phases:** 6\
**Rationale:** This is a full rearchitecture of the LSP server (~7,400 lines), introducing new abstractions (CursorContext, Provider interfaces, Schema), migrating all 8 existing features to the new pattern, adding awareness of 3 new language constructs (refs, state, events) across all features, updating the gopls virtual file generator, overhauling the VSCode extension (grammar, config, README), and updating the tree-sitter grammar. The work spans the LSP Go code, TypeScript extension, tree-sitter JavaScript grammar, and JSON configurations across ~20 files.

> **IMPORTANT:** User must approve the complexity assessment before proceeding to implementation plan. The plan MUST use the approved number of phases.

---

## 6. Success Criteria

1. **All existing LSP features work** — Diagnostics, completion, hover, definition, references, document/workspace symbols, formatting, and semantic tokens continue to function for all previously supported constructs
2. **Named refs are fully supported** — Hover shows ref type/context, definition jumps to `#Name` declaration, references finds all usages, semantic tokens highlights refs, completion suggests ref patterns
3. **Reactive state is fully supported** — Hover shows state type/initial value, completion suggests state methods after `.`, definition jumps to `tui.NewState()` declaration, references finds all `.Get()`/`.Set()` usages
4. **Event handlers are recognized** — Completion suggests event attributes, hover shows handler documentation, semantic tokens distinguishes event attributes
5. **gopls integration works with new constructs** — Go expressions referencing state vars and named refs get accurate completions and type information from gopls
6. **Provider architecture is clean** — Adding a new language construct requires updating the schema + CursorContext resolver + relevant providers (not all files). Adding a new LSP feature requires implementing one provider interface.
7. **VSCode extension is current** — README documents current syntax, TextMate grammar uses `.gsx` scopes throughout, language config folds on `templ`, test files use current syntax
8. **Tree-sitter grammar covers all constructs** — State declarations, event handler attributes, and any missing syntax are properly parsed
9. **All tests pass** — Existing tests continue to pass, new tests cover refs/state/events for each provider

---

## 7. Open Questions

1. Should semantic tokens have a dedicated token type for state variables (distinct from regular variables)? → **Yes**, use the existing `variable` type with a `readonly` modifier for state, since state is accessed through methods rather than directly assigned.

2. Should the schema include attribute validation rules (e.g., `onClick` only valid on elements, `focusable` must be set for focus events)? → **Deferred to v2**. For now, schema provides documentation/completion but not enforcement.

3. Should the LSP warn about unreachable state variables (declared but never `.Get()`'d in a binding)? → **Deferred to v2**. Focus on making existing diagnostics work correctly first.

4. Should we add state variable tracking to the ComponentIndex for cross-file state awareness? → **No for now**. State variables are component-local. Cross-file tracking is only needed for components and functions.
