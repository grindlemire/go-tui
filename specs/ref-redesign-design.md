# Explicit Refs & Handler Self-Inject Design

## Motivation

The current `#Name` syntax for element references is too magical:

```gsx
// Current: #Content creates a variable out of thin air
<div onChannel={tui.Watch(dataCh, addLine(lineCount, Content))}>
    <div #Content onKeyPress={handleScrollKeys(Content)}></div>
</div>
```

Problems:
1. `Content` appears in expressions but is never visibly declared — it's conjured by `#Content`
2. The connection between `#Content` on the element and `Content` in handler args requires framework knowledge
3. Self-referencing handlers (element referencing itself) require naming the element and passing that name back in — redundant ceremony
4. Not discoverable — you need to know the `#` convention

## Design Overview

Two complementary features replace `#Name`:

1. **Handler self-inject**: Element handlers automatically receive the element they're attached to as the first parameter. Eliminates refs entirely for the 80% case where an element references itself.

2. **Explicit `ref={}` attribute**: For cross-element access, declare a ref variable with `element.NewRef()` and bind it with `ref={varName}`. The variable's origin is visible at its declaration site.

```gsx
// New: explicit, traceable, no magic
content := element.NewRef()

<div onChannel={tui.Watch(dataCh, addLine(lineCount, content))}>
    <div ref={content}
        onKeyPress={handleScrollKeys}
        onEvent={handleEvent}></div>
</div>
```

```go
// Self-inject: handler receives its element automatically
func handleScrollKeys(el *element.Element, e tui.KeyEvent) {
    switch e.Rune {
    case 'j': el.ScrollBy(0, 1)
    case 'k': el.ScrollBy(0, -1)
    }
}

// Cross-element: explicit ref, closure captures ref
func addLine(lineCount *tui.State[int], content *element.Ref) func(string) {
    return func(line string) {
        lineCount.Set(lineCount.Get() + 1)
        el := content.El()
        el.AddChild(lineElem)
    }
}
```

---

## Part 1: New Ref Types

### `tui.Ref[T]` — Generic Ref (in `pkg/tui`)

```go
// pkg/tui/ref.go

// Ref is a typed reference to a value, set during construction
// and accessed later in handlers. Thread-safe.
type Ref[T any] struct {
    mu    sync.RWMutex
    value *T
}

func NewRef[T any]() *Ref[T] {
    return &Ref[T]{}
}

func (r *Ref[T]) Set(v *T) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.value = v
}

func (r *Ref[T]) El() *T {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.value
}

func (r *Ref[T]) IsSet() bool {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.value != nil
}
```

The generic type lives in `pkg/tui` to avoid circular imports (`element` imports `tui`, so `tui` cannot import `element`). The type parameter is supplied at the usage site.

### `element.Ref` — Convenience Alias (in `pkg/tui/element`)

```go
// pkg/tui/element/ref.go

// Ref is a reference to an Element. Declare with NewRef() and bind with ref={} in GSX.
type Ref = tui.Ref[Element]

// NewRef creates an unbound element reference.
func NewRef() *Ref {
    return tui.NewRef[Element]()
}
```

### `element.RefList` — Loop Refs (unkeyed)

```go
// pkg/tui/element/ref.go

// RefList holds references to multiple elements created in a loop.
type RefList struct {
    mu    sync.RWMutex
    elems []*Element
}

func NewRefList() *RefList {
    return &RefList{}
}

func (r *RefList) Append(el *Element) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.elems = append(r.elems, el)
}

func (r *RefList) All() []*Element {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.elems
}

func (r *RefList) At(i int) *Element {
    r.mu.RLock()
    defer r.mu.RUnlock()
    if i < 0 || i >= len(r.elems) {
        return nil
    }
    return r.elems[i]
}

func (r *RefList) Len() int {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return len(r.elems)
}
```

### `element.RefMap[K]` — Keyed Loop Refs

```go
// pkg/tui/element/ref.go

// RefMap holds keyed references to elements created in a loop.
type RefMap[K comparable] struct {
    mu    sync.RWMutex
    elems map[K]*Element
}

func NewRefMap[K comparable]() *RefMap[K] {
    return &RefMap[K]{elems: make(map[K]*Element)}
}

func (r *RefMap[K]) Put(key K, el *Element) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.elems[key] = el
}

func (r *RefMap[K]) Get(key K) *Element {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.elems[key]
}

func (r *RefMap[K]) All() map[K]*Element {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.elems
}

func (r *RefMap[K]) Len() int {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return len(r.elems)
}
```

### Usage in GSX

```gsx
// Single ref
content := element.NewRef()
<div ref={content}></div>
// Handler: content.El() → *element.Element

// List ref (loop, no key)
items := element.NewRefList()
@for _, item := range data {
    <span ref={items}>{item}</span>
}
// Handler: items.All() → []*element.Element, items.At(i) → *element.Element

// Map ref (loop, with key)
users := element.NewRefMap[string]()
@for _, user := range userData {
    <span ref={users} key={user.ID}>{user.Name}</span>
}
// Handler: users.Get("id") → *element.Element, users.All() → map[string]*element.Element
```

---

## Part 2: Handler Self-Inject

### Changed Handler Signatures

All element handler types gain `*element.Element` as their first parameter:

| Handler | Old Signature | New Signature |
|---------|--------------|---------------|
| `onKeyPress` | `func(tui.KeyEvent)` | `func(*element.Element, tui.KeyEvent)` |
| `onClick` | `func()` | `func(*element.Element)` |
| `onEvent` | `func(tui.Event) bool` | `func(*element.Element, tui.Event) bool` |
| `onFocus` | `func()` | `func(*element.Element)` |
| `onBlur` | `func()` | `func(*element.Element)` |

### Element struct changes (`pkg/tui/element/element.go`)

```go
// Field types change:
type Element struct {
    // ...
    onKeyPress func(*Element, tui.KeyEvent)   // was func(tui.KeyEvent)
    onClick    func(*Element)                  // was func()
    onEvent    func(*Element, tui.Event) bool  // was func(tui.Event) bool
    onFocus    func(*Element)                  // was func()
    onBlur     func(*Element)                  // was func()
}
```

### Setter changes

```go
func (e *Element) SetOnKeyPress(fn func(*Element, tui.KeyEvent)) {
    e.focusable = true
    e.onKeyPress = fn
}

func (e *Element) SetOnClick(fn func(*Element)) {
    e.focusable = true
    e.onClick = fn
}

func (e *Element) SetOnEvent(fn func(*Element, tui.Event) bool) {
    e.focusable = true
    e.onEvent = fn
}

func (e *Element) SetOnFocus(fn func(*Element)) {
    e.focusable = true
    e.onFocus = fn
}

func (e *Element) SetOnBlur(fn func(*Element)) {
    e.focusable = true
    e.onBlur = fn
}
```

### Dispatch changes (`HandleEvent`)

The dispatch code passes `e` (self) when calling handlers:

```go
func (e *Element) HandleEvent(event tui.Event) bool {
    if e.onEvent != nil {
        if e.onEvent(e, event) {  // pass self
            return true
        }
    }

    if keyEvent, ok := event.(tui.KeyEvent); ok {
        if e.onKeyPress != nil {
            e.onKeyPress(e, keyEvent)  // pass self
            return true
        }
        if e.onClick != nil && (keyEvent.Key == tui.KeyEnter || keyEvent.Rune == ' ') {
            e.onClick(e)  // pass self
            return true
        }
    }

    if mouseEvent, ok := event.(tui.MouseEvent); ok {
        if mouseEvent.Button == tui.MouseLeft {
            if e.onClick != nil {
                e.onClick(e)  // pass self
                return true
            }
        }
    }
    return false
}

// Focus/blur:
func (e *Element) Focus() {
    // ...
    if e.onFocus != nil {
        e.onFocus(e)  // pass self
    }
}

func (e *Element) Blur() {
    // ...
    if e.onBlur != nil {
        e.onBlur(e)  // pass self
    }
}
```

### Watcher handlers are NOT changed

`onChannel` and `onTimer` are external event sources attached to an element, not callbacks on the element itself. Their handlers do NOT receive the element:

- `tui.Watch(ch, func(val T))` — unchanged
- `tui.OnTimer(d, func())` — unchanged

Cross-element access in watchers uses explicit refs, which is the right pattern since watchers typically modify *other* elements.

---

## Part 3: GSX Syntax Changes

### Remove `#Name`

The `#Name` syntax is removed entirely from the language.

### Add `ref={}` Attribute

`ref` becomes a special attribute (like `key`). It is:
- Parsed as a regular attribute
- Extracted by the analyzer/generator for special handling
- Not passed as an element option

### Full Streaming Example (Before/After)

**Before (current):**
```gsx
templ StreamApp(dataCh <-chan string) {
    lineCount := tui.NewState(0)
    elapsed := tui.NewState(0)
    <div
        class="flex-col"
        onTimer={tui.OnTimer(time.Second, tickElapsed(elapsed))}
        onChannel={tui.Watch(dataCh, addLine(lineCount, Content))}>
        <div
            #Content
            class="flex-col border-cyan"
            border={tui.BorderSingle}
            flexGrow={1}
            scrollable={element.ScrollVertical}
            focusable={true}
            onKeyPress={handleScrollKeys(Content)}
            onEvent={handleEvent(Content)}></div>
        // ... footer ...
    </div>
}

func handleScrollKeys(content *element.Element) func(tui.KeyEvent) {
    return func(e tui.KeyEvent) {
        switch e.Rune {
        case 'j': content.ScrollBy(0, 1)
        case 'k': content.ScrollBy(0, -1)
        }
    }
}

func handleEvent(content *element.Element) func(tui.Event) bool {
    return func(e tui.Event) bool {
        if mouse, ok := e.(tui.MouseEvent); ok {
            switch mouse.Button {
            case tui.MouseWheelUp:  content.ScrollBy(0, -1); return true
            case tui.MouseWheelDown: content.ScrollBy(0, 1); return true
            }
        }
        return false
    }
}

func addLine(lineCount *tui.State[int], content *element.Element) func(string) {
    return func(line string) {
        lineCount.Set(lineCount.Get() + 1)
        stayAtBottom := content.IsAtBottom()
        lineElem := element.New(
            element.WithText(line),
            element.WithTextStyle(tui.NewStyle().Foreground(tui.Green)),
        )
        content.AddChild(lineElem)
        if stayAtBottom { content.ScrollToBottom() }
    }
}
```

**After (new):**
```gsx
templ StreamApp(dataCh <-chan string) {
    lineCount := tui.NewState(0)
    elapsed := tui.NewState(0)
    content := element.NewRef()
    <div
        class="flex-col"
        onTimer={tui.OnTimer(time.Second, tickElapsed(elapsed))}
        onChannel={tui.Watch(dataCh, addLine(lineCount, content))}>
        <div
            ref={content}
            class="flex-col border-cyan"
            border={tui.BorderSingle}
            flexGrow={1}
            scrollable={element.ScrollVertical}
            focusable={true}
            onKeyPress={handleScrollKeys}
            onEvent={handleEvent}></div>
        // ... footer ...
    </div>
}

// Self-inject: plain functions, no closures needed
func handleScrollKeys(el *element.Element, e tui.KeyEvent) {
    switch e.Rune {
    case 'j': el.ScrollBy(0, 1)
    case 'k': el.ScrollBy(0, -1)
    }
}

func handleEvent(el *element.Element, e tui.Event) bool {
    if mouse, ok := e.(tui.MouseEvent); ok {
        switch mouse.Button {
        case tui.MouseWheelUp:  el.ScrollBy(0, -1); return true
        case tui.MouseWheelDown: el.ScrollBy(0, 1); return true
        }
    }
    return false
}

// Cross-element: explicit ref, closure captures ref variable
func addLine(lineCount *tui.State[int], content *element.Ref) func(string) {
    return func(line string) {
        lineCount.Set(lineCount.Get() + 1)
        el := content.El()
        stayAtBottom := el.IsAtBottom()
        lineElem := element.New(
            element.WithText(line),
            element.WithTextStyle(tui.NewStyle().Foreground(tui.Green)),
        )
        el.AddChild(lineElem)
        if stayAtBottom { el.ScrollToBottom() }
    }
}
```

---

## Part 4: Implementation — Compiler Pipeline

### 4.1 Lexer (`pkg/tuigen/lexer.go`)

**Remove** `TokenHash` handling (lines 237-239):
```go
// DELETE:
case '#':
    l.readChar()
    return l.makeToken(TokenHash, "#")
```

**Remove** `TokenHash` from token types (`pkg/tuigen/token.go`, line 69).

The `#` character becomes a syntax error if used in element position, giving clear feedback.

### 4.2 Parser (`pkg/tuigen/parser.go`)

**Remove `#Name` parsing** (lines 837-850 in `parseElement`):
```go
// DELETE this entire block:
namedRefLine := pos.Line
p.skipNewlines()
if p.current.Type == TokenHash {
    p.advance()
    if p.current.Type != TokenIdent {
        p.errors.AddError(p.position(), "expected identifier after '#' for named ref")
        return nil
    }
    namedRefLine = p.current.Line
    elem.NamedRef = p.current.Literal
    p.advance()
}
```

**`ref={}` needs no parser changes** — it's already parsed as a regular attribute via `parseAttribute`. The analyzer/generator extract it, just like `key={}` is extracted today (lines 863-873).

### 4.3 AST (`pkg/tuigen/ast.go`)

**Remove** `NamedRef` field from `Element` struct (line 108):
```go
type Element struct {
    Tag        string
    // NamedRef   string   // REMOVE
    RefKey     *GoExpr    // Keep — still needed for keyed refs
    RefExpr    *GoExpr    // ADD — the expression from ref={expr}
    Attributes []*Attribute
    // ...
}
```

### 4.4 Analyzer (`pkg/tuigen/analyzer.go`)

**Replace `validateNamedRefs`** with `validateRefs`:

The analyzer:
1. Scans element attributes for `ref={expr}`
2. Extracts the expression and stores in `elem.RefExpr`
3. Removes `ref` from the attribute list (like `key` is removed today)
4. Validates the ref expression is a simple identifier
5. Detects context (in loop? has key?) and records the ref type
6. Validates no duplicate ref names

```go
type RefInfo struct {
    Name          string   // Variable name (e.g., "content")
    ExportName    string   // Capitalized for View struct (e.g., "Content")
    Element       *Element
    InLoop        bool
    InConditional bool
    KeyExpr       string
    KeyType       string
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

The ref kind is determined by context:
- Not in loop → `RefSingle`
- In loop, no key → `RefList`
- In loop, with key → `RefMap`

**Remove** `isValidRefName` uppercase-first-letter requirement. Ref names are now regular Go variable names (lowercase is fine since they're local variables). The View struct export name is generated by capitalizing.

**Update `knownAttributes`** map:
- Add `"ref": true`
- Keep `"key": true`

### 4.5 Generator (`pkg/tuigen/generator.go`)

This is the largest change area. The generator needs to:

#### A. Remove forward-declared ref variables

**Delete** the block at lines 216-233 that forward-declares `var Content *element.Element`. Refs are now user-declared Go variables (`content := element.NewRef()`), so the generator doesn't need to declare them.

#### B. Handle `ref={}` attribute

When building element options, extract `ref={expr}` and generate a `.Set()` call after element creation:

```go
// During element generation, when ref={expr} is present:
// 1. Create the element as usual
__tui_3 := element.New(
    element.WithDirection(layout.Column),
    // ... other options (ref is NOT included)
)
// 2. Bind the ref
content.Set(__tui_3)
```

For loop refs:
```go
// RefList (no key):
__tui_5 := element.New(...)
items.Append(__tui_5)

// RefMap (with key):
__tui_5 := element.New(...)
users.Put(user.ID, __tui_5)
```

#### C. Handler emission changes

**Handlers are no longer deferred.** Since refs are user-declared variables (not forward-declared by the generator), and handler signatures now receive the element via self-inject, handlers can be set inline during element creation using `With*` options instead of deferred `Set*` calls.

However, handlers that reference cross-element refs (via closure captures) still work because the ref variable is declared before element creation. The ref may not be bound yet, but the closure captures the ref pointer, not the element — the element is looked up when the handler fires.

For handlers, the generator emits as element options:

```go
// Self-inject handler (bare function identifier):
__tui_3 := element.New(
    element.WithDirection(layout.Column),
    element.WithOnKeyPress(handleScrollKeys),
)

// Call expression handler (returns func with self-inject signature):
__tui_3 := element.New(
    element.WithOnKeyPress(makeHandler(someState)),
)
```

The generator adds new `With*` option functions for handlers:

```go
// pkg/tui/element/options.go — new options
func WithOnKeyPress(fn func(*Element, tui.KeyEvent)) Option {
    return func(e *Element) {
        e.focusable = true
        e.onKeyPress = fn
    }
}

func WithOnClick(fn func(*Element)) Option {
    return func(e *Element) {
        e.focusable = true
        e.onClick = fn
    }
}

func WithOnEvent(fn func(*Element, tui.Event) bool) Option {
    return func(e *Element) {
        e.focusable = true
        e.onEvent = fn
    }
}

func WithOnFocus(fn func(*Element)) Option {
    return func(e *Element) {
        e.focusable = true
        e.onFocus = fn
    }
}

func WithOnBlur(fn func(*Element)) Option {
    return func(e *Element) {
        e.focusable = true
        e.onBlur = fn
    }
}
```

Update the `handlerAttributes` map to use `With*` options instead of `Set*` methods:

```go
var handlerAttributes = map[string]string{
    "onKeyPress": "WithOnKeyPress",
    "onClick":    "WithOnClick",
    "onEvent":    "WithOnEvent",
    "onFocus":    "WithOnFocus",
    "onBlur":     "WithOnBlur",
}
```

And generate handlers as options instead of deferred setters:

```go
// Old generated:
Content.SetOnKeyPress(handleScrollKeys(Content))

// New generated:
__tui_3 := element.New(
    // ... other options ...
    element.WithOnKeyPress(handleScrollKeys),
)
```

Note: Watchers are still deferred since they use `AddWatcher()` on the parent element.

#### D. View struct generation

The View struct changes from exposing `*element.Element` directly to dereferencing refs:

```go
// Old:
type StreamAppView struct {
    Root     *element.Element
    watchers []tui.Watcher
    Content  *element.Element
}

// New:
type StreamAppView struct {
    Root     *element.Element
    watchers []tui.Watcher
    Content  *element.Element   // resolved from ref
}

// Construction:
view = StreamAppView{
    Root:     __tui_0,
    watchers: watchers,
    Content:  content.El(),  // resolve ref to element
}
```

For loop refs:
```go
type DemoView struct {
    Root  *element.Element
    Items []*element.Element         // from RefList
    Users map[string]*element.Element // from RefMap
}

// Construction:
view = DemoView{
    Root:  __tui_0,
    Items: items.All(),
    Users: users.All(),
}
```

The View struct exposes resolved `*element.Element` values (not ref types) for external consumers, maintaining API compatibility.

#### E. Remove deferred handler infrastructure

Delete:
- `deferredHandler` struct (lines 18-23)
- Handler collection during element generation
- Deferred handler emission block (lines 288-294)

Keep:
- Deferred watcher emission (watchers are still deferred since they reference the parent element which must be created first)

### 4.6 Formatter (`pkg/formatter/`)

**Remove `#Name` formatting logic.** The formatter currently handles `#Name` in element printing. Since `ref={expr}` is a regular attribute, the formatter handles it automatically with no changes needed.

Search for `NamedRef` in the formatter and remove any special handling.

---

## Part 5: Implementation — LSP

### 5.1 Schema (`pkg/lsp/schema/schema.go`)

**Add `ref` to element attributes:**

```go
func genericAttrs() []AttributeDef {
    return []AttributeDef{
        {Name: "id", Type: "string", Description: "Unique identifier", Category: "generic"},
        {Name: "class", Type: "string", Description: "Tailwind-style classes", Category: "generic"},
        {Name: "ref", Type: "expression", Description: "Bind an element.Ref to this element for cross-element access", Category: "generic"},
        // ...
    }
}
```

### 5.2 CursorContext (`pkg/lsp/context.go`)

**Remove `NodeKindNamedRef`** from the NodeKind enum. Replace with logic that recognizes `ref` as an attribute.

**Update scope collection**: Remove named ref collection from scope. Refs are now regular Go variables, visible through normal Go code analysis.

### 5.3 Completions (`pkg/lsp/provider/completion.go`)

**Remove `#` trigger character handling** if any exists.

**Add `ref` attribute completion**: Already handled by the schema-driven attribute completion system — just adding `ref` to the schema is sufficient.

**Add ref method completions**: When typing `content.` where `content` is a ref, suggest `El()`, `IsSet()`, `Set()`. This requires detecting ref variable types, which can be done by pattern-matching `element.NewRef()` in preceding code.

### 5.4 References (`pkg/lsp/provider/references.go`)

**Update named ref reference finding** (lines 428-453):
- Remove `findNamedRefDeclInNodes()` that looks for `#Name`
- Add logic to find `ref={identifier}` in element attributes
- Find the ref variable declaration (e.g., `content := element.NewRef()`)
- Find all usages of the ref identifier in the component

### 5.5 Definition (`pkg/lsp/provider/definition.go`)

**Update go-to-definition for refs:**
- When cursor is on `ref={content}`, go to the declaration of `content`
- When cursor is on `content.El()` in a handler, go to the ref declaration

### 5.6 Hover (`pkg/lsp/provider/hover.go`)

**Update hover for ref attribute:**
- When hovering over `ref={content}`, show: "Element reference — binds this element to the `content` ref variable"
- When hovering over `element.NewRef()`, show: "Creates an unbound element reference. Bind with `ref={...}` in a template."

### 5.7 Semantic Tokens (`pkg/lsp/provider/semantic.go`)

**Remove `#Name` semantic token** (currently highlighted as `punctuation.special` + `variable.definition`).

**Add `ref` attribute value semantic token**: Highlight the value in `ref={content}` as a variable reference.

### 5.8 Diagnostics (`pkg/lsp/provider/diagnostics.go`)

**Update diagnostics:**
- Error if `#Name` syntax is used: "The `#Name` syntax has been removed. Use `ref={varName}` with `element.NewRef()` instead."
- Warning if ref is declared but never bound (no `ref={...}` uses it)
- Error if ref name is used in `ref={}` but not declared

---

## Part 6: Implementation — Editor Support

### 6.1 Tree-sitter Grammar (`editor/tree-sitter-gsx/grammar.js`)

**Remove `named_ref` rule** (lines 116-117):
```js
// DELETE:
named_ref: $ => seq('#', $.identifier),
```

**Remove `named_ref` from element rules** (lines 98-114):
```js
// Remove optional(field('named_ref', $.named_ref)) from both:
// self_closing_element (line 100)
// element_with_children (line 107)
```

The `ref={expr}` attribute is automatically handled by the existing `attribute` rule.

**Update test corpus** (`editor/tree-sitter-gsx/test/corpus/basic.txt`):
- Lines 367-386: Update "Named ref on element" test
- Lines 389-407: Update "Named ref on self-closing element" test
- Lines 466-488: Update "Element with named ref and attributes" test

### 6.2 Tree-sitter Highlights (`editor/tree-sitter-gsx/queries/highlights.scm`)

**Remove named ref highlighting** (lines 60-66):
```scheme
; DELETE:
(named_ref
  "#" @punctuation.special
  (identifier) @variable.definition)
```

**Add ref attribute highlighting:**
```scheme
; Ref attribute name gets special highlight
(attribute
  name: (identifier) @attribute.ref
  (#eq? @attribute.ref "ref"))
```

### 6.3 VSCode Syntax Highlighting (`editor/vscode/syntaxes/gsx.tmLanguage.json`)

**Remove `named-ref` pattern** (lines 216-227):
```json
// DELETE entire "named-ref" pattern
```

**Remove `#named-ref` include** from `element-open-tag` patterns.

**Add `ref` attribute pattern** to `attributes` section:
```json
{
  "name": "meta.attribute.ref.gsx",
  "match": "(ref)\\s*(=)\\s*(\\{)([^}]*)(\\})",
  "captures": {
    "1": { "name": "entity.other.attribute-name.ref.gsx support.type.property-name.gsx" },
    "2": { "name": "punctuation.separator.key-value.gsx" },
    "3": { "name": "punctuation.section.embedded.begin.gsx" },
    "4": { "name": "variable.other.ref.gsx" },
    "5": { "name": "punctuation.section.embedded.end.gsx" }
  }
}
```

### 6.4 VSCode Test Files

Update:
- `editor/vscode/test/simple.gsx` — change `#Main`, `#Title` to `ref={main}`, `ref={title}`
- `editor/vscode/test/complex.gsx` — change all `#Name` to `ref={name}`

---

## Part 7: Implementation — Examples

All examples using `#Name` must be updated. Each needs:
1. Add ref variable declaration(s) at top of templ function
2. Replace `#Name` with `ref={name}` attribute
3. Update handler signatures to use self-inject where applicable
4. Update cross-element handlers to use `*element.Ref`
5. Regenerate `*_gsx.go` files

### Files to update:

| File | Refs | Changes |
|------|------|---------|
| `examples/08-focus/focus.gsx` | `#BoxA`, `#BoxB`, `#BoxC` | Add `element.NewRef()` declarations, use `ref={}` |
| `examples/09-scrollable/scrollable.gsx` | `#Content` | Add ref declaration, self-inject scroll handlers |
| `examples/10-refs/refs.gsx` | `#Counter`, `#IncrementBtn`, `#DecrementBtn`, `#Status` | Add ref declarations, update handlers |
| `examples/11-streaming/streaming.gsx` | `#Content` | Add ref declaration, self-inject handlers |
| `examples/refs-demo/refs.gsx` | `#Header`, `#Content`, `#Items`, `#Warning`, `#StatusBar`, `#Users` | All ref types: single, list, map, conditional |
| `examples/streaming-dsl/streaming.gsx` | `#Content` | Primary showcase — ref + self-inject |

---

## Part 8: Implementation — Tests

### 8.1 Core Type Tests

**New file: `pkg/tui/ref_test.go`**
- `TestRef_SetAndGet` — basic set/get
- `TestRef_IsSet` — nil before set, true after
- `TestRef_NilBeforeSet` — El() returns nil before Set()

**New file: `pkg/tui/element/ref_test.go`**
- `TestNewRef` — creates unbound ref
- `TestRefList_Append` — append and retrieval
- `TestRefList_At` — bounds checking
- `TestRefMap_PutGet` — key-based access
- `TestRefMap_All` — full map retrieval

### 8.2 Parser Tests (`pkg/tuigen/parser_test.go`)

**Remove/Update:**
- `TestParser_NamedRef` — remove (no longer valid syntax)
- `TestParser_NamedRefWithKey` — remove
- `TestParser_MultipleNamedRefs` — remove

**Add:**
- `TestParser_RefAttribute` — `ref={content}` parsed as attribute
- `TestParser_RefAttributeWithKey` — `ref={items} key={id}` parsed
- `TestParser_HashSymbolError` — `#Name` produces parse error

### 8.3 Analyzer Tests (`pkg/tuigen/analyzer_test.go`)

**Remove/Update:**
- `TestAnalyzer_NamedRefValidation` — replace with ref attribute validation
- `TestAnalyzer_NamedRefInLoop` — replace with loop ref detection
- `TestAnalyzer_NamedRefInConditional` — replace with conditional ref detection
- `TestAnalyzer_CollectNamedRefs` — replace with ref collection from attributes

**Add:**
- `TestAnalyzer_RefAttribute` — validates ref extraction from attributes
- `TestAnalyzer_RefInLoop` — detects RefList/RefMap context
- `TestAnalyzer_RefDuplicateError` — duplicate ref names
- `TestAnalyzer_RefKeyOutsideLoop` — key without loop

### 8.4 Generator Tests (`pkg/tuigen/generator_test.go`)

**Update:**
- `TestGenerator_OnKeyPressAttribute` — handler uses new self-inject signature
- `TestGenerator_OnClickAttribute` — handler uses new signature
- `TestGenerator_WatcherGeneration` — watchers unchanged but handler format changes

**Add:**
- `TestGenerator_RefAttribute` — generates `content.Set(el)` binding
- `TestGenerator_RefListInLoop` — generates `items.Append(el)`
- `TestGenerator_RefMapInLoop` — generates `users.Put(key, el)`
- `TestGenerator_HandlerAsOption` — handler generated as `WithOnKeyPress()` option
- `TestGenerator_ViewStructWithRef` — view struct includes resolved refs

### 8.5 Element Tests (`pkg/tui/element/element_test.go`)

**Update existing handler tests** to pass `*Element` as first arg:
- All tests using `SetOnKeyPress`, `SetOnClick`, `SetOnEvent`, `SetOnFocus`, `SetOnBlur`

### 8.6 LSP Tests

**Update any tests** in `pkg/lsp/` that reference `#Name` syntax.

---

## Part 9: Generated Code Example

For the streaming-dsl example, the new generated code would be:

```go
// Code generated by tui generate. DO NOT EDIT.
// Source: streaming.gsx

package main

import (
    "fmt"
    "time"

    "github.com/grindlemire/go-tui/pkg/layout"
    "github.com/grindlemire/go-tui/pkg/tui"
    "github.com/grindlemire/go-tui/pkg/tui/element"
)

// Self-inject handlers — plain functions
func handleScrollKeys(el *element.Element, e tui.KeyEvent) {
    switch e.Rune {
    case 'j': el.ScrollBy(0, 1)
    case 'k': el.ScrollBy(0, -1)
    }
}

func handleEvent(el *element.Element, e tui.Event) bool {
    if mouse, ok := e.(tui.MouseEvent); ok {
        switch mouse.Button {
        case tui.MouseWheelUp:  el.ScrollBy(0, -1); return true
        case tui.MouseWheelDown: el.ScrollBy(0, 1); return true
        }
    }
    return false
}

// Closure handlers for cross-element/state access
func tickElapsed(elapsed *tui.State[int]) func() {
    return func() {
        elapsed.Set(elapsed.Get() + 1)
    }
}

func addLine(lineCount *tui.State[int], content *element.Ref) func(string) {
    return func(line string) {
        lineCount.Set(lineCount.Get() + 1)
        el := content.El()
        stayAtBottom := el.IsAtBottom()
        lineElem := element.New(
            element.WithText(line),
            element.WithTextStyle(tui.NewStyle().Foreground(tui.Green)),
        )
        el.AddChild(lineElem)
        if stayAtBottom {
            el.ScrollToBottom()
        }
    }
}

type StreamAppView struct {
    Root     *element.Element
    watchers []tui.Watcher
    Content  *element.Element
}

func (v StreamAppView) GetRoot() tui.Renderable { return v.Root }
func (v StreamAppView) GetWatchers() []tui.Watcher { return v.watchers }

func StreamApp(dataCh <-chan string) StreamAppView {
    var view StreamAppView
    var watchers []tui.Watcher

    lineCount := tui.NewState(0)
    elapsed := tui.NewState(0)
    content := element.NewRef()

    __tui_0 := element.New(
        element.WithDirection(layout.Column),
    )
    __tui_1 := element.New(
        element.WithBorderStyle(tui.NewStyle().Foreground(tui.Blue)),
        element.WithBorder(tui.BorderSingle),
        element.WithHeight(3),
        element.WithDirection(layout.Row),
        element.WithJustify(layout.JustifyCenter),
        element.WithAlign(layout.AlignCenter),
    )
    __tui_2 := element.New(
        element.WithText("Streaming DSL Demo - Use j/k to scroll, q to quit"),
        element.WithTextStyle(tui.NewStyle().Bold().Foreground(tui.White)),
    )
    __tui_1.AddChild(__tui_2)
    __tui_0.AddChild(__tui_1)

    __tui_3 := element.New(
        element.WithDirection(layout.Column),
        element.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan)),
        element.WithBorder(tui.BorderSingle),
        element.WithFlexGrow(1),
        element.WithScrollable(element.ScrollVertical),
        element.WithFocusable(true),
        element.WithOnKeyPress(handleScrollKeys),
        element.WithOnEvent(handleEvent),
    )
    content.Set(__tui_3)
    __tui_0.AddChild(__tui_3)

    __tui_4 := element.New(
        element.WithBorderStyle(tui.NewStyle().Foreground(tui.Blue)),
        element.WithBorder(tui.BorderSingle),
        element.WithHeight(3),
        element.WithDirection(layout.Row),
        element.WithJustify(layout.JustifyCenter),
        element.WithAlign(layout.AlignCenter),
    )
    __tui_5 := element.New(
        element.WithText(fmt.Sprintf("Lines: %d | Elapsed: %ds | Press q to exit", lineCount.Get(), elapsed.Get())),
        element.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
    )
    __tui_4.AddChild(__tui_5)
    __tui_0.AddChild(__tui_4)

    // Attach watchers
    __tui_0.AddWatcher(tui.OnTimer(time.Second, tickElapsed(elapsed)))
    __tui_0.AddWatcher(tui.Watch(dataCh, addLine(lineCount, content)))

    // State bindings
    __update___tui_5 := func() {
        __tui_5.SetText(fmt.Sprintf("Lines: %d | Elapsed: %ds | Press q to exit", lineCount.Get(), elapsed.Get()))
    }
    lineCount.Bind(func(_ int) { __update___tui_5() })
    elapsed.Bind(func(_ int) { __update___tui_5() })

    view = StreamAppView{
        Root:     __tui_0,
        watchers: watchers,
        Content:  content.El(),
    }
    return view
}
```

Key differences from current generated code:
1. No `var Content *element.Element` forward declaration
2. `content := element.NewRef()` — user-declared variable
3. `content.Set(__tui_3)` — explicit binding after element creation
4. Handlers as `WithOnKeyPress(handleScrollKeys)` options — not deferred `SetOnKeyPress`
5. `Content: content.El()` in view struct — resolved ref
6. Self-inject handler functions are plain (not closures returning closures)

---

## Part 10: Implementation Order

### Phase 1: Core Types
1. Add `tui.Ref[T]` generic type (`pkg/tui/ref.go`)
2. Add `element.Ref`, `element.RefList`, `element.RefMap` (`pkg/tui/element/ref.go`)
3. Add tests for all ref types
4. Change handler signatures on `Element` struct
5. Add `WithOn*` option functions
6. Update handler dispatch to pass self
7. Update element tests

### Phase 2: Compiler Pipeline
8. Remove `TokenHash` from lexer
9. Remove `#Name` parsing from parser
10. Update AST: remove `NamedRef`, add `RefExpr`
11. Update analyzer: replace ref validation
12. Update generator: ref binding, handler options, view struct
13. Update formatter: remove `#Name` handling
14. Update all compiler tests

### Phase 3: Editor Support
15. Update tree-sitter grammar: remove `named_ref`
16. Update tree-sitter highlights
17. Update tree-sitter tests
18. Update VSCode tmLanguage: remove `#` pattern, add `ref={}` pattern
19. Update VSCode test files

### Phase 4: LSP
20. Update schema: add `ref` attribute
21. Update CursorContext: remove `NodeKindNamedRef`
22. Update completions
23. Update references
24. Update definition
25. Update hover
26. Update semantic tokens
27. Update diagnostics

### Phase 5: Examples & Polish
28. Update all 6 example `.gsx` files
29. Regenerate all `*_gsx.go` files
30. Update VSCode test `.gsx` files
31. Run full test suite
32. Verify all examples build and run
