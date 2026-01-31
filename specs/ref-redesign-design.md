# Explicit Refs & Handler Self-Inject Specification

**Status:** Planned\
**Version:** 2.0\
**Last Updated:** 2025-01-31

---

## 1. Overview

### Purpose

Replace the magic `#Name` syntax for element references with two complementary features: **handler self-inject** (handlers automatically receive their element as the first parameter) and **explicit `ref={}` attribute** (cross-element access via declared ref variables). This eliminates undiscoverable syntax, reduces closure boilerplate, and makes the origin of every variable visible at its declaration site.

### Goals

- Remove the `#Name` syntax entirely from the GSX language
- Handlers that only need their own element require zero ceremony (self-inject)
- Cross-element references are explicit, traceable Go variables (`tui.NewRef()` + `ref={}`)
- View struct API remains compatible (exposes `*tui.Element`, not ref types)
- Support single refs, loop refs (slice), and keyed loop refs (map) matching current capabilities
- Clean error messages when old `#Name` syntax is used

### Non-Goals

- Ref types for non-Element values (future work if needed)
- Changes to watcher signatures (`tui.Watch`, `tui.OnTimer` unchanged)
- Changes to `onRender` or `onChildAdded` callbacks (already receive `*Element`)
- Changes to `onUpdate` callback (pre-render hook, no element needed)

---

## 2. Architecture

### Files Modified

```
tui/                                  # Root package
├── ref.go                            # NEW: Ref, RefList, RefMap types
├── ref_test.go                       # NEW: Ref type tests
├── element.go                        # MODIFY: handler field signatures
├── element_options.go                # MODIFY: WithOn* option signatures
├── element_focus.go                  # MODIFY: Set* methods, HandleEvent dispatch
│
├── internal/tuigen/
│   ├── token.go                      # MODIFY: remove TokenHash
│   ├── lexer.go                      # MODIFY: remove '#' handling
│   ├── ast.go                        # MODIFY: NamedRef → RefExpr on Element
│   ├── parser_element.go             # MODIFY: remove #Name parsing
│   ├── analyzer.go                   # MODIFY: NamedRef struct → RefInfo
│   ├── analyzer_refs.go              # MODIFY: ref validation rewrite
│   ├── generator.go                  # MODIFY: deferred handler types
│   ├── generator_element.go          # MODIFY: ref binding, handler-as-option
│   ├── generator_component.go        # MODIFY: forward decls → ref binding, view struct
│   └── (test files)                  # MODIFY: update all ref/handler tests
│
├── internal/formatter/
│   └── printer_elements.go           # MODIFY: remove #Name formatting
│
├── internal/lsp/
│   ├── context.go                    # MODIFY: remove NodeKindNamedRef
│   ├── context_resolve.go            # MODIFY: detect ref={} instead of #Name
│   ├── provider/definition.go        # MODIFY: ref definition navigation
│   ├── provider/references.go        # MODIFY: ref reference finding
│   ├── provider/hover.go             # MODIFY: ref hover info
│   ├── provider/semantic_nodes.go    # MODIFY: remove #Name tokens
│   ├── provider_adapters.go          # MODIFY: update named ref handling
│   └── schema/schema.go             # MODIFY: add "ref" attribute
│
├── editor/
│   ├── tree-sitter-gsx/grammar.js    # MODIFY: remove named_ref rule
│   ├── tree-sitter-gsx/queries/highlights.scm  # MODIFY: highlighting
│   ├── tree-sitter-gsx/test/corpus/basic.txt    # MODIFY: test cases
│   └── vscode/syntaxes/gsx.tmLanguage.json      # MODIFY: remove #Name pattern
│
└── examples/                         # MODIFY: all 6 examples using #Name
    ├── 08-focus/focus.gsx
    ├── 09-scrollable/scrollable.gsx
    ├── 10-refs/refs.gsx
    ├── 11-streaming/streaming.gsx
    ├── refs-demo/refs.gsx
    └── streaming-dsl/streaming.gsx
```

### Component Overview

| Component | Changes |
|-----------|---------|
| `tui/ref.go` | New ref types: `Ref`, `RefList`, `RefMap[K]` |
| `tui/element.go` | Handler fields gain `*Element` first param |
| `tui/element_options.go` | `WithOn*` options gain `*Element` first param |
| `tui/element_focus.go` | `Set*` methods and `HandleEvent` dispatch pass self |
| `internal/tuigen/` | Full compiler pipeline: remove `#`, add `ref={}`, handler-as-option |
| `internal/formatter/` | Remove `#Name` printing logic |
| `internal/lsp/` | Context, completions, references, definition, hover, semantics |
| `editor/tree-sitter-gsx/` | Grammar, highlights, test corpus |
| `editor/vscode/` | TextMate grammar updates |
| `examples/` | Update all 6 examples + regenerate |

### Flow Diagram

```
                        OLD FLOW
┌──────────────┐    ┌───────────────┐    ┌──────────────────┐
│ #Content     │───►│ NamedRef AST  │───►│ var Content *El  │
│ (magic)      │    │ field         │    │ Content = New()  │
└──────────────┘    └───────────────┘    │ Content.SetOn*() │
                                         └──────────────────┘

                        NEW FLOW
┌──────────────┐    ┌───────────────┐    ┌──────────────────┐
│ ref={content} │───►│ RefExpr AST   │───►│ el := New(       │
│ (explicit)   │    │ field         │    │   WithOnKey*(..) │
└──────────────┘    └───────────────┘    │ )                │
                                         │ content.Set(el)  │
                                         └──────────────────┘
```

---

## 3. Core Entities

### 3.1 Ref Types (new: `tui/ref.go`)

All ref types live in the root `tui` package alongside `Element`.

```go
// Ref is a reference to an Element, set during construction
// and accessed later in handlers. Thread-safe.
type Ref struct {
    mu    sync.RWMutex
    value *Element
}

func NewRef() *Ref {
    return &Ref{}
}

func (r *Ref) Set(v *Element) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.value = v
}

func (r *Ref) El() *Element {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.value
}

func (r *Ref) IsSet() bool {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.value != nil
}
```

No generic type parameter needed since refs always hold `*Element` in this framework. This keeps the API simple: `tui.NewRef()` instead of `tui.NewRef[tui.Element]()`.

```go
// RefList holds references to multiple elements created in a loop.
type RefList struct {
    mu    sync.RWMutex
    elems []*Element
}

func NewRefList() *RefList              { return &RefList{} }
func (r *RefList) Append(el *Element)   { /* lock + append */ }
func (r *RefList) All() []*Element      { /* lock + return copy */ }
func (r *RefList) At(i int) *Element    { /* lock + bounds check */ }
func (r *RefList) Len() int             { /* lock + return len */ }
```

```go
// RefMap holds keyed references to elements created in a loop.
type RefMap[K comparable] struct {
    mu    sync.RWMutex
    elems map[K]*Element
}

func NewRefMap[K comparable]() *RefMap[K]       { return &RefMap[K]{elems: make(map[K]*Element)} }
func (r *RefMap[K]) Put(key K, el *Element)     { /* lock + assign */ }
func (r *RefMap[K]) Get(key K) *Element         { /* lock + lookup */ }
func (r *RefMap[K]) All() map[K]*Element        { /* lock + return */ }
func (r *RefMap[K]) Len() int                   { /* lock + return len */ }
```

### 3.2 Changed Handler Signatures

All element handler types gain `*Element` as their first parameter:

| Handler | Current Signature | New Signature |
|---------|------------------|---------------|
| `onKeyPress` | `func(KeyEvent)` | `func(*Element, KeyEvent)` |
| `onClick` | `func()` | `func(*Element)` |
| `onEvent` | `func(Event) bool` | `func(*Element, Event) bool` |
| `onFocus` | `func()` | `func(*Element)` |
| `onBlur` | `func()` | `func(*Element)` |

**Unchanged handlers** (already have `*Element` or don't need it):
- `onRender func(*Element, *Buffer)` — already has `*Element`
- `onChildAdded func(*Element)` — internal callback
- `onFocusableAdded func(Focusable)` — internal callback
- `onUpdate func()` — pre-render hook, no element context needed

### 3.3 AST Changes (`internal/tuigen/ast.go`)

```go
type Element struct {
    Tag        string
    // NamedRef   string   // REMOVE
    RefExpr    *GoExpr    // NEW: expression from ref={expr}
    RefKey     *GoExpr    // KEEP: key expression for map-based refs
    Attributes []*Attribute
    Children   []Node
    // ... rest unchanged
}
```

### 3.4 Analyzer RefInfo (`internal/tuigen/analyzer.go`)

```go
// RefInfo replaces NamedRef with richer type information.
type RefInfo struct {
    Name          string   // Variable name from ref={name} (e.g., "content")
    ExportName    string   // Capitalized for View struct (e.g., "Content")
    Element       *Element
    InLoop        bool
    InConditional bool
    KeyExpr       string   // if set, generates map ref
    KeyType       string   // inferred type of key expression
    RefKind       RefKind  // RefSingle, RefList, RefMap
    Position      Position
}

type RefKind int
const (
    RefSingle RefKind = iota
    RefList
    RefMap
)
```

Ref kind is determined by context:
- Not in loop → `RefSingle`
- In loop, no key → `RefList`
- In loop, with key → `RefMap`

---

## 4. User Experience

### GSX Usage — Before (Current)

```gsx
templ StreamApp(dataCh <-chan string) {
    lineCount := tui.NewState(0)
    <div onChannel={tui.Watch(dataCh, addLine(lineCount, Content))}>
        <div #Content
            scrollable={tui.ScrollVertical}
            onKeyPress={handleScrollKeys(Content)}
            onEvent={handleEvent(Content)}></div>
    </div>
}

// Closure-returning-closure pattern (verbose)
func handleScrollKeys(content *tui.Element) func(tui.KeyEvent) {
    return func(e tui.KeyEvent) {
        switch e.Rune {
        case 'j': content.ScrollBy(0, 1)
        case 'k': content.ScrollBy(0, -1)
        }
    }
}
```

### GSX Usage — After (New)

```gsx
templ StreamApp(dataCh <-chan string) {
    lineCount := tui.NewState(0)
    content := tui.NewRef()
    <div onChannel={tui.Watch(dataCh, addLine(lineCount, content))}>
        <div ref={content}
            scrollable={tui.ScrollVertical}
            onKeyPress={handleScrollKeys}
            onEvent={handleEvent}></div>
    </div>
}

// Self-inject: plain functions, no closures needed
func handleScrollKeys(el *tui.Element, e tui.KeyEvent) {
    switch e.Rune {
    case 'j': el.ScrollBy(0, 1)
    case 'k': el.ScrollBy(0, -1)
    }
}

// Cross-element: closure captures ref variable (pointer)
func addLine(lineCount *tui.State[int], content *tui.Ref) func(string) {
    return func(line string) {
        lineCount.Set(lineCount.Get() + 1)
        el := content.El()
        el.AddChild(lineElem)
    }
}
```

### Ref Types in Loops

```gsx
// Single ref
content := tui.NewRef()
<div ref={content}></div>

// List ref (loop, no key)
items := tui.NewRefList()
@for _, item := range data {
    <span ref={items}>{item}</span>
}

// Map ref (loop, with key)
users := tui.NewRefMap[string]()
@for _, user := range userData {
    <span ref={users} key={user.ID}>{user.Name}</span>
}
```

### Generated Code (Key Differences)

```go
// OLD generated:
var Content *tui.Element
Content = tui.New(...)
Content.SetOnKeyPress(handleScrollKeys(Content))  // deferred

// NEW generated:
__tui_3 := tui.New(
    tui.WithOnKeyPress(handleScrollKeys),  // inline option, self-inject
)
content.Set(__tui_3)  // bind ref after creation
```

View struct construction:
```go
// OLD:
view = StreamAppView{Root: __tui_0, Content: Content}

// NEW:
view = StreamAppView{Root: __tui_0, Content: content.El()}
```

---

## 5. Complexity Assessment

| Size | Phases | When to Use |
|------|--------|-------------|
| Small | 1-2 | Single component, bug fix, minor enhancement |
| Medium | 3-4 | New feature touching multiple files/components |
| **Large** | **5-6** | **Cross-cutting feature, new subsystem** |

**Assessed Size:** Large\
**Recommended Phases:** 5\
**Rationale:** This feature touches every layer of the stack — runtime types (root `tui` package), the entire compiler pipeline (lexer, parser, AST, analyzer, generator, formatter in `internal/tuigen`), the language server (`internal/lsp` — 8+ provider files), editor integrations (tree-sitter grammar, VSCode syntax), and all 6 examples. The breadth across ~40+ files and 5 distinct subsystems makes this unambiguously large.

> **IMPORTANT:** User must approve the complexity assessment before proceeding to implementation plan. The plan MUST use the approved number of phases.

---

## 6. Detailed Design

### 6.1 Lexer Changes (`internal/tuigen/token.go`, `lexer.go`)

- **Remove** `TokenHash` from the token type enum
- **Remove** the `case '#'` branch in the lexer's scan switch
- The `#` character becomes a syntax error in element position, giving clear feedback

### 6.2 Parser Changes (`internal/tuigen/parser_element.go`)

- **Remove** the `#Name` detection block that checks for `TokenHash` after the tag name
- **No new parser changes needed** for `ref={}` — it is parsed as a regular attribute via the existing `parseAttribute` path, just like `key={}` is already extracted

### 6.3 Analyzer Changes (`internal/tuigen/analyzer_refs.go`)

Replace `validateNamedRefs` with `validateRefs`:

1. Scan element attributes for `ref={expr}`
2. Extract the expression and store in `elem.RefExpr`
3. Remove `ref` from the attribute list (like `key` is removed today)
4. Validate the ref expression is a simple identifier (not a complex expression)
5. Detect context (in loop? has key?) and determine ref kind
6. Capitalize variable name for View struct export name
7. Validate no duplicate ref names and no reserved names ("Root")

**Key change:** Remove `isValidRefName` uppercase-first-letter requirement. Ref names are now regular Go variable names (lowercase). The export name for the View struct is generated by capitalizing the first letter.

Add `"ref"` to the `knownAttributes` map alongside `"key"`.

### 6.4 Generator Changes (`internal/tuigen/generator_component.go`, `generator_element.go`)

This is the largest change area:

**A. Remove forward-declared ref variables** — Refs are now user-declared Go variables (`content := tui.NewRef()`), so the generator no longer emits `var Content *tui.Element`.

**B. Handle `ref={}` binding** — After element creation, emit the appropriate binding call:

```go
// RefSingle:
content.Set(__tui_3)

// RefList (no key):
items.Append(__tui_5)

// RefMap (with key):
users.Put(user.ID, __tui_5)
```

**C. Handlers as inline options** — Instead of deferred `Set*` calls, emit handlers as `With*` options during element creation:

```go
// OLD: deferred
Content.SetOnKeyPress(handleScrollKeys(Content))

// NEW: inline option
__tui_3 := tui.New(
    tui.WithOnKeyPress(handleScrollKeys),
)
```

Update `handlerAttributes` map to use `With*` options:
```go
var handlerAttributes = map[string]string{
    "onKeyPress": "WithOnKeyPress",
    "onClick":    "WithOnClick",
    "onEvent":    "WithOnEvent",
    "onFocus":    "WithOnFocus",
    "onBlur":     "WithOnBlur",
}
```

**D. Remove deferred handler infrastructure** — Delete `deferredHandler` struct and the deferred handler emission block. Keep deferred watcher emission (watchers still use `AddWatcher` on the parent).

**E. View struct generation** — Change view struct field resolution:

```go
// OLD: direct variable
Content: Content,

// NEW: resolve ref
Content: content.El(),

// Loop refs:
Items: items.All(),
Users: users.All(),
```

The View struct continues to expose `*tui.Element` (and `[]*tui.Element`, `map[K]*tui.Element`) — not ref types — maintaining API compatibility for external consumers.

### 6.5 Formatter Changes (`internal/formatter/printer_elements.go`)

Remove `NamedRef` handling in element printing. Since `ref={expr}` is a regular attribute, the formatter handles it automatically with no special logic needed.

### 6.6 LSP Changes

| Provider | Changes |
|----------|---------|
| `context.go` | Remove `NodeKindNamedRef` enum value |
| `context_resolve.go` | Detect `ref={ident}` in attributes instead of `#Name` |
| `schema/schema.go` | Add `"ref"` to generic element attributes |
| `provider/definition.go` | Navigate from `ref={content}` → variable declaration |
| `provider/references.go` | Find all usages of a ref variable |
| `provider/hover.go` | Show ref info on `ref={content}` |
| `provider/semantic_nodes.go` | Remove `#Name` token, add `ref` attribute value token |
| `provider_adapters.go` | Update named ref scope handling |

### 6.7 Editor Changes

**Tree-sitter (`editor/tree-sitter-gsx/grammar.js`):**
- Remove `named_ref` rule: `named_ref: $ => seq('#', $.identifier)`
- Remove `optional(field('named_ref', $.named_ref))` from element rules
- `ref={expr}` handled by existing `attribute` rule automatically

**Tree-sitter highlights (`queries/highlights.scm`):**
- Remove `named_ref` highlighting (the `#` + identifier pattern)
- Optionally add `ref` attribute name highlight

**VSCode (`editor/vscode/syntaxes/gsx.tmLanguage.json`):**
- Remove `named-ref` pattern and its include from element open tag
- Add `ref` attribute pattern to highlight the value as a variable reference

---

## 7. Success Criteria

1. All `#Name` syntax removed — using `#Name` in a `.gsx` file produces a clear parse error
2. `ref={}` attribute correctly binds single refs, list refs, and map refs
3. Self-inject handlers work — `onKeyPress={handleScrollKeys}` where `handleScrollKeys(el *tui.Element, e tui.KeyEvent)` receives the element automatically
4. Cross-element access works via `tui.Ref` — closures capture ref pointer, resolve at runtime
5. View struct continues to expose `*tui.Element` fields, not ref types
6. All 6 existing examples updated and functional with new syntax
7. LSP provides completions, hover, go-to-definition, and references for `ref={}` attributes
8. Tree-sitter grammar and VSCode syntax highlighting handle `ref={}` correctly
9. All existing tests pass (updated for new signatures)
10. `go test ./...` passes with no failures

---

## 8. Open Questions

1. ~~Should `Ref` be generic `Ref[T]`?~~ → No. Refs always hold `*Element` in this framework. A non-generic `Ref` is simpler. If needed for other types later, a generic version can be added.
2. ~~Should `RefList`/`RefMap` be deferred to v2?~~ → No. The current `#Name` system supports loop refs. Dropping them would be a regression.
3. ~~Should watchers be un-deferred?~~ → Keep watchers deferred for safety. Ref pointers are valid from declaration, but deferring costs nothing and prevents edge cases.
4. ~~Package location for ref types?~~ → Root `tui` package. No `element` sub-package exists; no circular import concern.
