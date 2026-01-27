# Named Element References Specification

**Status:** Draft  
**Version:** 2.1  
**Last Updated:** 2026-01-27

---

## 1. Overview

### Purpose

Enable users to declaratively name elements in `.tui` files and access them from Go code. This bridges the gap between declarative UI structure and imperative behavior (scrolling, dynamic children, animations).

Currently, DSL components return only the root element. To interact with inner elements (call `ScrollToBottom()`, `AddChild()`, etc.), users must create those elements imperatively in Go and pass them into the DSL—defeating the DSL's purpose.

### Goals

- **Named element syntax**: `#Name` marker on any element makes it accessible
- **Struct return type**: All components return a struct with typed fields
- **Consistent API**: All components return `ComponentNameView` struct, preventing breaking changes when refs are added
- **Loop support**: Refs inside `@for` loops generate slice fields (or map fields with `key` attribute)
- **Conditional support**: Refs inside `@if`/`@else` may be nil at runtime
- **Keyed refs**: `key={expr}` attribute generates map-based refs for stable element correlation
- **Seamless composition**: Named elements work naturally with component composition
- **Closure-captured view**: Generated code pre-declares view variable so handlers can reference it

> **Note:** The `onUpdate` attribute was originally part of this spec but has been moved to the Event Handling spec. See `specs/event-handling-design.md` for details on event handlers including `onUpdate`, `onKeyPress`, `onClick`, `onChannel`, and `onTimer`.

### Non-Goals

- Reactive/signal-based state management
- Automatic re-rendering on state change
- Multiple return value syntax (tuples)
- String-based ID lookup (already exists via `id` attribute)

---

## 2. Architecture

### Directory Structure

```
pkg/tuigen/
├── ast.go            # MODIFY: Add NamedRef field to Element
├── lexer.go          # MODIFY: Add TokenHash for #Name syntax
├── parser.go         # MODIFY: Parse #Name on elements
├── analyzer.go       # MODIFY: Validate #Name uniqueness, check reserved names, track loop/conditional context
├── generator.go      # MODIFY: Always generate struct returns, handle slice types for loop refs
└── generator_test.go # MODIFY: Add tests for named refs

editor/
├── tree-sitter-tui/grammar.js      # UPDATE: Parse # token and ref name
└── vscode/syntaxes/tui.tmLanguage.json  # UPDATE: Highlight #Name syntax

examples/
└── streaming-dsl/    # UPDATE: Use named refs pattern
    ├── streaming.tui
    └── main.go
```

### Component Overview


| Component      | Change                                                                |
| -------------- | --------------------------------------------------------------------- |
| `lexer.go`     | Add `TokenHash` for `#` character                                     |
| `ast.go`       | Add `NamedRef string` field to `Element` struct                       |
| `parser.go`    | Parse `#Name` after tag name: `<div #Content ...>`                    |
| `analyzer.go`  | Validate unique names, reserved names, track loop/conditional context |
| `generator.go` | Always generate view struct, handle slice types for loop refs         |


### Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│  .tui Source                                                    │
│  @component StreamBox() {                                       │
│      <div #Content scrollable={...}></div>                      │
│  }                                                              │
└─────────────────────────────┬───────────────────────────────────┘
                              │ Parser detects #Content
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  AST                                                            │
│  Element{                                                       │
│      Tag: "div",                                                │
│      NamedRef: "Content",   ← NEW FIELD                         │
│      Attributes: [...]                                          │
│  }                                                              │
└─────────────────────────────┬───────────────────────────────────┘
                              │ Analyzer validates names, tracks context
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Generator always generates struct (even without refs)          │
└─────────────────────────────┬───────────────────────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Generated Go                                                   │
│                                                                 │
│  type StreamBoxView struct {                                    │
│      Root    *element.Element                                   │
│      Content *element.Element                                   │
│  }                                                              │
│                                                                 │
│  func StreamBox() StreamBoxView {                               │
│      var view StreamBoxView  // pre-declared for closure capture│
│                                                                 │
│      Content := element.New(                                    │
│          element.WithScrollable(...),                           │
│      )                                                          │
│      Root := element.New()                                      │
│      Root.AddChild(Content)                                     │
│                                                                 │
│      view = StreamBoxView{Root: Root, Content: Content}         │
│      return view                                                │
│  }                                                              │
└─────────────────────────────────────────────────────────────────┘
```

---

## 3. Core Entities

### 3.1 Lexer Changes

Add a new token type for `#`:

```go
// In token.go
const (
    // ... existing tokens ...
    TokenHash        // #
)
```

```go
// In lexer.go - add case in Next()
case '#':
    l.advance()
    return Token{Type: TokenHash, Literal: "#", Line: l.line, Column: col}
```

### 3.2 AST Changes

Add `NamedRef` and `RefKey` fields to Element:

```go
// In ast.go
type Element struct {
    Tag        string
    NamedRef   string       // NEW: Name for this element (e.g., "Content" from #Content)
    RefKey     *Expression  // NEW: Key expression for map-based refs (e.g., key={item.ID})
    Attributes []Attribute
    Children   []Node
    SelfClose  bool
    Pos        Position
}
```

### 3.3 Parser Changes

Parse `#Name` after the tag name:

```go
// In parser.go - parseElement()

// After parsing tag name, check for #Name
func (p *Parser) parseElement() (*Element, error) {
    // ... parse < and tag name ...

    elem := &Element{Tag: tagName}

    // Check for #Name
    if p.peek().Type == TokenHash {
        p.next() // consume #
        nameTok := p.expect(TokenIdent)
        elem.NamedRef = nameTok.Literal
    }

    // ... continue parsing attributes and children ...
}
```

### 3.4 Analyzer Changes

Validate named refs and track loop/conditional context:

```go
// In analyzer.go

// NamedRef tracks information about a named element reference
type NamedRef struct {
    Name          string
    Element       *Element
    InLoop        bool       // true = generate slice or map type
    InConditional bool       // true = may be nil at runtime
    KeyExpr       string     // if set, generate map[KeyType]*element.Element
    KeyType       string     // inferred type of key expression (e.g., "string", "int")
    Position      Position
}

// NEW: Validate named refs in component
func (a *Analyzer) validateNamedRefs(comp *Component) ([]NamedRef, error) {
    names := make(map[string]Position)
    var refs []NamedRef

    var check func(nodes []Node, inLoop, inConditional bool) error
    check = func(nodes []Node, inLoop, inConditional bool) error {
        for _, node := range nodes {
            switch n := node.(type) {
            case *Element:
                if n.NamedRef != "" {
                    // Must be valid Go identifier (PascalCase recommended)
                    if !isValidIdentifier(n.NamedRef) {
                        return fmt.Errorf("%s: invalid ref name %q - must be valid Go identifier starting with uppercase letter",
                            n.Pos, n.NamedRef)
                    }
                    // Reserved name check
                    if n.NamedRef == "Root" {
                        return fmt.Errorf("%s: ref name 'Root' is reserved", n.Pos)
                    }
                    // Must be unique
                    if prev, exists := names[n.NamedRef]; exists {
                        return fmt.Errorf("%s: duplicate ref name %q (first defined at %s)",
                            n.Pos, n.NamedRef, prev)
                    }
                    names[n.NamedRef] = n.Pos

                    ref := NamedRef{
                        Name:          n.NamedRef,
                        Element:       n,
                        InLoop:        inLoop,
                        InConditional: inConditional,
                        Position:      n.Pos,
                    }

                    // Check for key attribute (for map-based refs)
                    if n.RefKey != nil {
                        if !inLoop {
                            return fmt.Errorf("%s: key attribute on ref %q only valid inside @for loop",
                                n.Pos, n.NamedRef)
                        }
                        ref.KeyExpr = n.RefKey.String()
                        ref.KeyType = a.inferType(n.RefKey) // infer key type
                    }

                    refs = append(refs, ref)
                }
                if err := check(n.Children, inLoop, inConditional); err != nil {
                    return err
                }

            case *ForLoop:
                // Refs inside loops get slice type
                if err := check(n.Body, true, inConditional); err != nil {
                    return err
                }

            case *IfStmt:
                // Refs inside conditionals may be nil
                if err := check(n.Then, inLoop, true); err != nil {
                    return err
                }
                if err := check(n.Else, inLoop, true); err != nil {
                    return err
                }
            }
        }
        return nil
    }

    err := check(comp.Body, false, false)
    return refs, err
}
```

**Error message examples:**

```
streaming.tui:15:9: duplicate ref name 'Content' (first defined at streaming.tui:8:13)
streaming.tui:22:5: ref name 'Root' is reserved
streaming.tui:12:14: invalid ref name '123invalid' - must be valid Go identifier starting with uppercase letter
```

### 3.5 Generator Changes

Always generate struct return type, handle loop/conditional refs:

```go
// In generator.go

// Generate struct type for component (always generated)
func (g *Generator) generateViewStruct(comp *Component, refs []NamedRef) {
    structName := comp.Name + "View"

    g.writef("type %s struct {\n", structName)
    g.writef("\tRoot *element.Element\n")
    for _, ref := range refs {
        if ref.InLoop {
            if ref.KeyExpr != "" {
                // Map type for keyed refs
                g.writef("\t%s map[%s]*element.Element\n", ref.Name, ref.KeyType)
            } else {
                // Slice type for unkeyed loop refs
                g.writef("\t%s []*element.Element\n", ref.Name)
            }
        } else {
            if ref.InConditional {
                g.writef("\t%s *element.Element // may be nil\n", ref.Name)
            } else {
                g.writef("\t%s *element.Element\n", ref.Name)
            }
        }
    }
    g.writef("}\n\n")
}

// generateComponent always returns struct
func (g *Generator) generateComponent(comp *Component) {
    refs := g.collectNamedRefs(comp)
    structName := comp.Name + "View"

    // Always generate struct
    g.generateViewStruct(comp, refs)

    // Generate function signature - always returns struct
    g.writef("func %s(%s) %s {\n", comp.Name, g.formatParams(comp.Params), structName)

    // IMPORTANT: Pre-declare view variable so closures can capture it.
    // This solves the circular reference problem where handlers need
    // to reference the view before it's fully constructed.
    g.writef("\tvar view %s\n\n", structName)

    // Declare slice/map variables for loop refs at function scope
    for _, ref := range refs {
        if ref.InLoop {
            if ref.KeyExpr != "" {
                g.writef("\t%s := make(map[%s]*element.Element)\n", ref.Name, ref.KeyType)
            } else {
                g.writef("\tvar %s []*element.Element\n", ref.Name)
            }
        }
    }

    // Declare pointer variables for conditional refs at function scope
    for _, ref := range refs {
        if ref.InConditional && !ref.InLoop {
            g.writef("\tvar %s *element.Element\n", ref.Name)
        }
    }

    // ... generate body ...
    // Event handlers can now reference 'view' variable safely:
    //   element.WithOnKeyPress(func(e tui.KeyEvent) {
    //       view.Content.ScrollBy(0, 1)  // closure captures 'view' variable
    //   })

    // Populate view struct before returning
    g.writef("\tview = %s{\n", structName)
    g.writef("\t\tRoot: %s,\n", rootVarName)
    for _, ref := range refs {
        g.writef("\t\t%s: %s,\n", ref.Name, ref.Name)
    }
    g.writef("\t}\n")
    g.writef("\treturn view\n")

    g.writef("}\n")
}
```

---

## 4. DSL Syntax

### 4.1 Named Element Syntax

```tui
// Name an element with #Name after the tag
<div #Content scrollable={element.ScrollVertical}>
</div>

// Works on any element
<span #Title class="font-bold">{"Hello"}</span>

// Self-closing elements too
<div #Spacer height={2} />
```

### 4.2 Component Without Refs (Still Returns Struct)

```tui
@component Header() {
    <div class="border-single" height={3}>
        <span class="font-bold">{"Title"}</span>
    </div>
}
```

**Generated:**

```go
type HeaderView struct {
    Root *element.Element
}

func Header() HeaderView {
    root := element.New(
        element.WithBorder(tui.BorderSingle),
        element.WithHeight(3),
    )
    // ... children ...
    return HeaderView{Root: root}
}
```

### 4.3 Component with Named Refs

```tui
@component StreamBox() {
    <div class="flex-col">
        <div #Header class="border-single" height={3}>
            <span class="font-bold">{"Stream"}</span>
        </div>
        <div #Content
             class="border-cyan p-1"
             scrollable={element.ScrollVertical}
             flexGrow={1}>
        </div>
        <div #Footer class="border-single" height={3}>
            <span #Status>{"Ready"}</span>
        </div>
    </div>
}
```

**Generated:**

```go
type StreamBoxView struct {
    Root    *element.Element
    Header  *element.Element
    Content *element.Element
    Footer  *element.Element
    Status  *element.Element
}

func StreamBox() StreamBoxView {
    Header := element.New(
        element.WithBorder(tui.BorderSingle),
        element.WithHeight(3),
    )
    // ... header children ...

    Content := element.New(
        element.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan)),
        element.WithPadding(1),
        element.WithScrollable(element.ScrollVertical),
        element.WithFlexGrow(1),
    )

    Status := element.New(
        element.WithText("Ready"),
    )

    Footer := element.New(
        element.WithBorder(tui.BorderSingle),
        element.WithHeight(3),
    )
    Footer.AddChild(Status)

    Root := element.New(
        element.WithDirection(layout.Column),
    )
    Root.AddChild(Header, Content, Footer)

    return StreamBoxView{
        Root:    Root,
        Header:  Header,
        Content: Content,
        Footer:  Footer,
        Status:  Status,
    }
}
```

### 4.4 Refs Inside Loops (Slice Type)

```tui
@component ItemList(items []string) {
    <ul>
        @for _, item := range items {
            <li #Items>{item}</li>
        }
    </ul>
}
```

**Generated:**

```go
type ItemListView struct {
    Root  *element.Element
    Items []*element.Element  // slice because ref is inside loop
}

func ItemList(items []string) ItemListView {
    var Items []*element.Element  // declared at function scope

    Root := element.New()
    for _, item := range items {
        elem := element.New(element.WithText(item))
        Items = append(Items, elem)
        Root.AddChild(elem)
    }

    return ItemListView{Root: Root, Items: Items}
}
```

**Usage:**

```go
view := ItemList([]string{"a", "b", "c"})
app.SetRoot(view.Root)

// Access individual items
view.Items[0].SetTextStyle(highlightStyle)

// Iterate all items
for _, item := range view.Items {
    item.SetBorder(tui.BorderSingle)
}
```

**Notes:**

- If a conditional inside a loop filters elements, slice indices won't align with input indices
- Multiple refs in one loop body create parallel slices of the same length
- Empty input produces empty slice (not nil)

### 4.5 Keyed Refs Inside Loops (Map Type)

When refs are inside loops with conditionals, slice indices don't correlate with input indices. Use the `key` attribute to generate a map instead, enabling stable element correlation:

```tui
@component UserList(users []User) {
    <ul>
        @for _, user := range users {
            @if user.Active {
                <li #Users key={user.ID}>{user.Name}</li>
            }
        }
    </ul>
}
```

**Generated:**

```go
type UserListView struct {
    Root  *element.Element
    Users map[string]*element.Element  // map because key={user.ID}
}

func UserList(users []User) UserListView {
    var view UserListView

    Users := make(map[string]*element.Element)

    Root := element.New()
    for _, user := range users {
        if user.Active {
            elem := element.New(element.WithText(user.Name))
            Users[user.ID] = elem  // keyed by user.ID
            Root.AddChild(elem)
        }
    }

    view = UserListView{Root: Root, Users: Users}
    return view
}
```

**Usage:**

```go
view := UserList(users)
app.SetRoot(view.Root)

// Access by key - stable correlation with input data
view.Users["user-123"].SetTextStyle(highlightStyle)

// Check if user element exists (may be filtered by conditional)
if elem, ok := view.Users["user-456"]; ok {
    elem.SetBorder(tui.BorderSingle)
}
```

**Key Attribute Rules:**

- `key` attribute is only valid on refs inside `@for` loops
- Key expression must evaluate to a comparable type (string, int, etc.)
- Key type is inferred from the expression
- Duplicate keys overwrite previous elements (last wins)
- Without `key`, refs in loops generate slice type (indices may not correlate)

### 4.6 Refs Inside Conditionals (May Be Nil)

```tui
@component Foo(showLabel bool) {
    <div>
        @if showLabel {
            <span #Label>{"Hi"}</span>
        }
    </div>
}
```

**Generated:**

```go
type FooView struct {
    Root  *element.Element
    Label *element.Element  // may be nil
}

func Foo(showLabel bool) FooView {
    var Label *element.Element  // declared outside conditional

    Root := element.New()
    if showLabel {
        Label = element.New(element.WithText("Hi"))
        Root.AddChild(Label)
    }

    return FooView{Root: Root, Label: Label}
}
```

**Usage:**

```go
view := Foo(false)
app.SetRoot(view.Root)

// Must nil-check before use!
if view.Label != nil {
    view.Label.SetText("Updated")
}
```

---

## 5. User Experience

### 5.1 Complete Streaming Example

```tui
// streaming.tui
package main

import "fmt"

@component Header() {
    <div class="border-blue" border={tui.BorderSingle} height={3}
         justify={layout.JustifyCenter} align={layout.AlignCenter}>
        <span class="font-bold text-white">{"Streaming Demo"}</span>
    </div>
}

@component Footer(scrollY int, maxScroll int, status string) {
    <div class="border-blue" border={tui.BorderSingle} height={3}
         justify={layout.JustifyCenter} align={layout.AlignCenter}>
        <span class="text-white">{fmt.Sprintf("Scroll: %d/%d | Auto: %s | ESC exit", scrollY, maxScroll, status)}</span>
    </div>
}

@component StreamApp() {
    <div class="flex-col">
        @Header()
        <div #Content
             class="border-cyan p-1"
             border={tui.BorderSingle}
             scrollable={element.ScrollVertical}
             flexGrow={1}
             direction={layout.Column}>
        </div>
        @Footer(0, 0, "ON")
    </div>
}
```

```go
// main.go
package main

import (
    "fmt"
    "time"

    "github.com/grindlemire/go-tui/pkg/layout"
    "github.com/grindlemire/go-tui/pkg/tui"
    "github.com/grindlemire/go-tui/pkg/tui/element"
)

//go:generate go run ../../cmd/tui generate streaming.tui

func main() {
    app, _ := tui.NewApp()
    defer app.Close()

    width, height := app.Size()

    // Build UI - get back struct with named element refs
    view := StreamApp()

    // Set root size
    style := view.Root.Style()
    style.Width = layout.Fixed(width)
    style.Height = layout.Fixed(height)
    view.Root.SetStyle(style)

    app.SetRoot(view.Root)
    app.Focus().Register(view.Content)

    // Simulate adding content - direct access to Content ref!
    go func() {
        for i := 0; i < 100; i++ {
            time.Sleep(200 * time.Millisecond)
            view.Content.AddChild(element.New(
                element.WithText(fmt.Sprintf("[%s] Log line %d", time.Now().Format("15:04:05"), i)),
                element.WithTextStyle(tui.NewStyle().Foreground(tui.Green)),
            ))
        }
    }()

    for {
        event, ok := app.PollEvent(50 * time.Millisecond)
        if ok {
            switch e := event.(type) {
            case tui.KeyEvent:
                if e.Key == tui.KeyEscape {
                    return
                }
                // Scroll handling - direct access to Content ref!
                switch e.Rune {
                case 'j':
                    view.Content.ScrollBy(0, 1)
                case 'k':
                    view.Content.ScrollBy(0, -1)
                case 'G':
                    view.Content.ScrollToBottom()
                }
            case tui.ResizeEvent:
                width, height = e.Width, e.Height
                style := view.Root.Style()
                style.Width = layout.Fixed(width)
                style.Height = layout.Fixed(height)
                view.Root.SetStyle(style)
            }
        }
        app.Render()
    }
}
```

> **Note:** This example uses a manual event loop. Once the Event Handling spec is implemented, this simplifies to using `app.Run()` with `onChannel` and `onKeyPress` handlers. See `specs/event-handling-design.md`.

### 5.2 Key Benefits

1. **Direct element access**: `view.Content.ScrollToBottom()` instead of wrapping or passing refs
2. **Type-safe**: Compiler catches typos in ref names
3. **Discoverable**: Autocomplete shows available refs
4. **Declarative structure**: UI layout stays in `.tui` file
5. **Imperative behavior**: Scroll, add children, etc. in Go
6. **No breaking changes**: Adding refs to a component doesn't change return type

### 5.3 Composition Pattern

Named refs work naturally with composition:

```tui
@component Dashboard() {
    <div class="flex-col">
        <div #Sidebar width={20}>
            // sidebar content
        </div>
        <div #Main flexGrow={1}>
            // main content
        </div>
    </div>
}
```

```go
dash := Dashboard()
app.SetRoot(dash.Root)

// Later, update sidebar
dash.Sidebar.AddChild(newMenuItem)

// Update main content
stream := StreamBox(pollFn)
dash.Main.AddChild(stream.Root)
```

---

## 6. Rules and Constraints

1. **`#Name` must be valid Go identifier** - Must start with uppercase letter, alphanumeric only
2. **`#Root` is reserved** - Conflicts with the always-present `Root` field; analyzer error if used
3. **Names must be unique within a component** - Including across different branches of conditionals
4. **PascalCase required** - For consistency with Go exported struct fields
5. **All components return struct** - `ComponentNameView` struct containing at minimum a `Root` field; adding refs adds fields without changing return type
6. **Refs inside `@for` loops generate slice fields** - Type is `[]*element.Element` (without `key`)
7. **Refs with `key` attribute generate map fields** - Type is `map[KeyType]*element.Element`; key only valid inside loops
8. **Refs inside `@if`/`@else` may be nil at runtime** - Users must nil-check before use
9. **Deeply nested refs work** - `#Name` can be on any element at any depth
10. **View variable pre-declared** - Generated code declares `var view ComponentView` first so handlers can capture it
11. **Nested loops produce flat slices** - Refs inside nested `@for` loops generate a single flat `[]*element.Element`; use `key={expr}` for correlation
12. **Refs on root element allowed** - Both `Root` and the named ref point to the same element

---

## 7. Complexity Assessment


| Size   | Phases | When to Use                                    |
| ------ | ------ | ---------------------------------------------- |
| Small  | 1-2    | Single component, bug fix, minor enhancement   |
| Medium | 3-4    | New feature touching multiple files/components |
| Large  | 5-6    | Cross-cutting feature, new subsystem           |


**Assessed Size:** Medium  
**Recommended Phases:** 2

**Rationale:**

- Lexer change: trivial (add `#` token)
- AST change: trivial (add field)
- Parser change: small (detect `#Name` after tag)
- Analyzer change: moderate (validate uniqueness, reserved names, track loop/conditional context)
- Generator change: moderate (always generate struct, slice types for loops, conditional variable hoisting)
- Testing: moderate (new syntax needs comprehensive tests)

### Phase Breakdown

1. **Phase 1: Lexer, AST, Parser, Analyzer for #Name** (Medium)
  - Add `TokenHash` to lexer
  - Add `NamedRef` field to Element AST
  - Parse `#Name` syntax in parser
  - Add `InLoop bool` and `InConditional bool` tracking to ref collection
  - Add analyzer validation for reserved name `Root`
  - Add analyzer validation for unique names
  - Update tree-sitter grammar and VSCode syntax highlighting
2. **Phase 2: Generator struct returns** (Medium)
  - Always generate `ComponentNameView` struct for all components
  - Generate slice types (`[]*element.Element`) for loop refs
  - Generate variable declarations outside conditionals for conditional refs
  - Add comments indicating which refs may be nil
  - Update examples to use named refs

---

## 8. Success Criteria

1. `#Name` syntax parses without error on any element
2. Duplicate `#Name` within component produces analyzer error
3. Invalid `#Name` (e.g., `#123invalid`) produces analyzer error
4. `#Root` produces analyzer error (reserved name)
5. All components return `ComponentNameView` struct with at minimum `Root` field
6. Struct contains `Root` plus all named elements as fields
7. `#Name` inside `@for` without `key` generates slice field `[]*element.Element`
8. `#Name` inside `@for` with `key={expr}` generates map field `map[Type]*element.Element`
9. `key` attribute outside `@for` loop produces analyzer error
10. `#Name` inside `@if` generates pointer field that may be nil at runtime
11. Parallel refs inside same loop produce parallel slices of equal length
12. Nested `#Name` elements at any depth are captured in struct
13. Generated code pre-declares `var view ComponentView` for closure capture
14. Event handlers can safely reference `view.RefName` in closures
15. Generated code compiles and runs correctly
16. Examples work with new pattern

---

## 9. Editor Tooling

### LSP Support

The language server should support `#Name` for:

- **Go-to-definition**: Jump from `#Name` in .tui file to generated struct field
- **Find references**: Find usages of named element in .go files
- **Rename refactoring**: Rename ref name across .tui and .go files

### Tree-sitter

Update `editor/tree-sitter-tui/grammar.js` to parse:

- `#` token as ref marker
- Identifier following `#` as ref name
- Full pattern: `<tag #Name attr=value>`

### VSCode Extension

Update `editor/vscode/syntaxes/tui.tmLanguage.json`:

- Highlight `#Name` distinctly (e.g., as a variable or type)
- Match pattern: `#[A-Z][a-zA-Z0-9]`*

---

## 10. Resolved Design Decisions

### Q1: Should `Root` be renamed if there's a `#Root` element?

**Decision:** Disallow `#Root` as a ref name. It conflicts with the always-present `Root` field. Analyzer produces error: `ref name 'Root' is reserved`

### Q2: What about naming elements inside `@for` loops?

**Decision:** Allowed. Generate slice type `[]*element.Element` instead of single pointer. See section 4.4.

### Q3: What about naming elements inside conditionals?

**Decision:** Allowed. Field will be nil if branch not taken. Users must nil-check before use. See section 4.5.

### Q4: Relationship to `id` attribute?

**Decision:** `#Name` and `id` are orthogonal:

- `id` is for runtime string-based lookup (`FindByID`)
- `#Name` is for compile-time typed access via struct fields
- Both can be used on the same element if needed

### Q5: How to correlate refs inside loops with conditionals?

**Decision:** Use `key={expr}` attribute to generate map-based refs.

Without `key`: Refs generate slice type where indices may not correlate with input indices (elements filtered by conditionals are skipped).

With `key={item.ID}`: Refs generate map type `map[Type]*element.Element` allowing stable correlation by key.

### Q6: How do handlers reference the view before it's constructed?

**Decision:** Generated code pre-declares `var view ComponentView` at function start. Closures capture the variable (not the value), so when handlers execute later, `view` is populated.

```go
func StreamApp() StreamAppView {
    var view StreamAppView  // declared first

    Content := element.New(
        element.WithOnKeyPress(func(e tui.KeyEvent) {
            view.Content.ScrollBy(0, 1)  // works: captures variable
        }),
    )

    view = StreamAppView{Root: Root, Content: Content}  // populated later
    return view
}
```

### Q7: What happens with refs inside nested loops?

**Decision:** Nested loops produce a flat slice. All elements are appended to a single `[]*element.Element` regardless of nesting depth. For correlation with input data, use `key={expr}` to generate a map instead.

```tui
@component Grid(rows [][]Item) {
    @for _, row := range rows {
        @for _, item := range row {
            <span #Cells key={item.ID}>{item.Name}</span>
        }
    }
}
```

Generates `Cells map[string]*element.Element` for stable lookup by ID.

### Q8: Can refs be placed on the root element?

**Decision:** Allowed. Both `Root` and the named ref point to the same element. This enables semantic naming:

```tui
@component Sidebar() {
    <nav #Navigation class="flex-col">
        ...
    </nav>
}
```

Generates:
```go
type SidebarView struct {
    Root       *element.Element
    Navigation *element.Element  // same as Root
}
```

---

## 11. Summary

### DSL Syntax

```tui
@component StreamBox() {
    <div class="flex-col">
        <div #Header height={3}>
            <span>{"Title"}</span>
        </div>
        <div #Content scrollable={element.ScrollVertical} flexGrow={1}>
        </div>
    </div>
}
```

### Generated

```go
type StreamBoxView struct {
    Root    *element.Element
    Header  *element.Element
    Content *element.Element
}
```

### Usage

```go
view := StreamBox()
app.SetRoot(view.Root)
view.Content.ScrollToBottom()
```

### Rules Table


| Scenario                                | Generated Type                 | Notes                                         |
| --------------------------------------- | ------------------------------ | --------------------------------------------- |
| No `#Name` in component                 | `ComponentNameView{Root: ...}` | Struct with only Root                         |
| `#Name` outside loops/conditionals      | `*element.Element`             | Always non-nil                                |
| `#Name` inside `@for` (no key)          | `[]*element.Element`           | May be empty slice; indices may not correlate |
| `#Name` inside `@for` with `key={expr}` | `map[Type]*element.Element`    | Stable key-based correlation                  |
| `#Name` inside nested `@for` loops      | `[]*element.Element`           | Flat slice; use key for correlation           |
| `#Name` inside `@if`                    | `*element.Element`             | May be nil                                    |
| `#Name` on root element                 | `*element.Element`             | Same as Root; semantic alias                  |
| `#Root`                                 | Error                          | Reserved name                                 |
| `key` outside `@for`                    | Error                          | Key only valid in loops                       |


