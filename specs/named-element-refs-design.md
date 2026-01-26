# Named Element References & onUpdate Hook Specification

**Status:** Draft\
**Version:** 1.0\
**Last Updated:** 2026-01-26

---

## 1. Overview

### Purpose

Enable users to declaratively name elements in `.tui` files and access them from Go code. This bridges the gap between declarative UI structure and imperative behavior (scrolling, dynamic children, animations).

Currently, DSL components return only the root element. To interact with inner elements (call `ScrollToBottom()`, `AddChild()`, etc.), users must create those elements imperatively in Go and pass them into the DSL—defeating the DSL's purpose.

### Goals

- **Named element syntax**: `#Name` marker on any element makes it accessible
- **Struct return type**: Components with named elements return a struct with typed fields
- **Backwards compatible**: Components without `#Name` still return `*element.Element`
- **onUpdate in DSL**: Expose the existing `onUpdate` hook as a DSL attribute
- **Seamless composition**: Named elements work naturally with component composition

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
├── analyzer.go       # MODIFY: Add onUpdate to knownAttributes, validate #Name uniqueness
├── generator.go      # MODIFY: Add onUpdate to attributeToOption, generate struct returns
└── generator_test.go # MODIFY: Add tests for named refs

examples/
└── streaming-dsl/    # UPDATE: Use named refs pattern
    ├── streaming.tui
    └── main.go
```

### Component Overview

| Component | Change |
|-----------|--------|
| `lexer.go` | Add `TokenHash` for `#` character |
| `ast.go` | Add `NamedRef string` field to `Element` struct |
| `parser.go` | Parse `#Name` after tag name: `<div #Content ...>` |
| `analyzer.go` | Validate unique names per component, add `onUpdate` attribute |
| `generator.go` | Generate view struct when component has named refs, add `onUpdate` mapping |

### Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│  .tui Source                                                    │
│  @component StreamBox() {                                       │
│      <div #Content scrollable={...} onUpdate={poll}></div>      │
│  }                                                              │
└─────────────────────────────┬───────────────────────────────────┘
                              │ Parser detects #Content
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  AST                                                            │
│  Element{                                                       │
│      Tag: "div",                                                │
│      NamedRef: "Content",   ← NEW FIELD                         │
│      Attributes: [..., {Name: "onUpdate", Value: poll}]         │
│  }                                                              │
└─────────────────────────────┬───────────────────────────────────┘
                              │ Analyzer validates names
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Generator detects named refs → generates struct                │
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
│      Content := element.New(                                    │
│          element.WithScrollable(...),                           │
│          element.WithOnUpdate(poll),                            │
│      )                                                          │
│      Root := element.New()                                      │
│      Root.AddChild(Content)                                     │
│      return StreamBoxView{Root: Root, Content: Content}         │
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

Add `NamedRef` field to Element:

```go
// In ast.go
type Element struct {
    Tag        string
    NamedRef   string       // NEW: Name for this element (e.g., "Content" from #Content)
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

Add `onUpdate` to known attributes and validate named refs:

```go
// In analyzer.go

var knownAttributes = map[string]bool{
    // ... existing attributes ...

    // Focus/Callbacks
    "onFocus":  true,
    "onBlur":   true,
    "onEvent":  true,
    "onUpdate": true,  // NEW

    // ... rest ...
}

// NEW: Validate named refs in component
func (a *Analyzer) validateNamedRefs(comp *Component) error {
    names := make(map[string]Position)

    var check func(nodes []Node) error
    check = func(nodes []Node) error {
        for _, node := range nodes {
            if elem, ok := node.(*Element); ok {
                if elem.NamedRef != "" {
                    // Must be valid Go identifier (PascalCase recommended)
                    if !isValidIdentifier(elem.NamedRef) {
                        return fmt.Errorf("%s: invalid ref name %q", elem.Pos, elem.NamedRef)
                    }
                    // Must be unique
                    if prev, exists := names[elem.NamedRef]; exists {
                        return fmt.Errorf("%s: duplicate ref name %q (first defined at %s)",
                            elem.Pos, elem.NamedRef, prev)
                    }
                    names[elem.NamedRef] = elem.Pos
                }
                if err := check(elem.Children); err != nil {
                    return err
                }
            }
            // ... handle other node types with children ...
        }
        return nil
    }

    return check(comp.Body)
}
```

### 3.5 Generator Changes

Add `onUpdate` attribute mapping and struct generation:

```go
// In generator.go

var attributeToOption = map[string]string{
    // ... existing mappings ...

    // Focus/Callbacks
    "onFocus":  "element.WithOnFocus(%s)",
    "onBlur":   "element.WithOnBlur(%s)",
    "onEvent":  "element.WithOnEvent(%s)",
    "onUpdate": "element.WithOnUpdate(%s)",  // NEW

    // ... rest ...
}

// NEW: Collect named refs from component
func (g *Generator) collectNamedRefs(comp *Component) []NamedRef {
    var refs []NamedRef

    var collect func(nodes []Node)
    collect = func(nodes []Node) {
        for _, node := range nodes {
            if elem, ok := node.(*Element); ok {
                if elem.NamedRef != "" {
                    refs = append(refs, NamedRef{
                        Name:    elem.NamedRef,
                        Element: elem,
                    })
                }
                collect(elem.Children)
            }
            // ... handle other node types ...
        }
    }

    collect(comp.Body)
    return refs
}

// NEW: Generate struct type for component with named refs
func (g *Generator) generateViewStruct(comp *Component, refs []NamedRef) {
    structName := comp.Name + "View"

    g.writef("type %s struct {\n", structName)
    g.writef("\tRoot *element.Element\n")
    for _, ref := range refs {
        g.writef("\t%s *element.Element\n", ref.Name)
    }
    g.writef("}\n\n")
}

// MODIFY: generateComponent to return struct when named refs exist
func (g *Generator) generateComponent(comp *Component) {
    refs := g.collectNamedRefs(comp)
    hasNamedRefs := len(refs) > 0

    if hasNamedRefs {
        g.generateViewStruct(comp, refs)
    }

    // Generate function signature
    returnType := "*element.Element"
    if hasNamedRefs {
        returnType = comp.Name + "View"
    }

    g.writef("func %s(%s) %s {\n", comp.Name, g.formatParams(comp.Params), returnType)

    // ... generate body ...

    // Generate return
    if hasNamedRefs {
        g.writef("\treturn %sView{\n", comp.Name)
        g.writef("\t\tRoot: %s,\n", rootVarName)
        for _, ref := range refs {
            g.writef("\t\t%s: %s,\n", ref.Name, g.varNameFor(ref.Element))
        }
        g.writef("\t}\n")
    } else {
        g.writef("\treturn %s\n", rootVarName)
    }

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

### 4.2 Component with Named Refs

```tui
@component StreamBox(pollFn func()) {
    <div class="flex-col">
        <div #Header class="border-single" height={3}>
            <span class="font-bold">{"Stream"}</span>
        </div>
        <div #Content
             class="border-cyan p-1"
             scrollable={element.ScrollVertical}
             onUpdate={pollFn}
             flexGrow={1}>
        </div>
        <div #Footer class="border-single" height={3}>
            <span #Status>{"Ready"}</span>
        </div>
    </div>
}
```

### 4.3 Generated Output

```go
type StreamBoxView struct {
    Root    *element.Element
    Header  *element.Element
    Content *element.Element
    Footer  *element.Element
    Status  *element.Element
}

func StreamBox(pollFn func()) StreamBoxView {
    Header := element.New(
        element.WithBorder(tui.BorderSingle),
        element.WithHeight(3),
    )
    // ... header children ...

    Content := element.New(
        element.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan)),
        element.WithPadding(1),
        element.WithScrollable(element.ScrollVertical),
        element.WithOnUpdate(pollFn),
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

@component StreamApp(pollFn func()) {
    <div class="flex-col">
        @Header()
        <div #Content
             class="border-cyan p-1"
             border={tui.BorderSingle}
             scrollable={element.ScrollVertical}
             onUpdate={pollFn}
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
    "os"
    "strings"
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
    textCh := make(chan string, 100)

    // State
    autoScroll := true
    textStyle := tui.NewStyle().Foreground(tui.Green)

    // Poll function - will be connected to Content via onUpdate
    var view StreamAppView
    poll := func() {
        for {
            select {
            case text, ok := <-textCh:
                if !ok {
                    return
                }
                for _, line := range strings.Split(text, "\n") {
                    if line == "" {
                        continue
                    }
                    view.Content.AddChild(element.New(
                        element.WithText(line),
                        element.WithTextStyle(textStyle),
                    ))
                }
            default:
                if autoScroll {
                    view.Content.ScrollToBottom()
                }
                return
            }
        }
    }

    // Build UI - get back struct with named element refs
    view = StreamApp(poll)

    // Set root size
    style := view.Root.Style()
    style.Width = layout.Fixed(width)
    style.Height = layout.Fixed(height)
    view.Root.SetStyle(style)

    app.SetRoot(view.Root)
    app.Focus().Register(view.Content)

    go simulateProcess(textCh)

    for {
        event, ok := app.PollEvent(50 * time.Millisecond)
        if ok {
            switch e := event.(type) {
            case tui.KeyEvent:
                if e.Key == tui.KeyEscape {
                    return
                }
                // Scroll handling - direct access to Content!
                switch e.Rune {
                case 'j':
                    view.Content.ScrollBy(0, 1)
                    autoScroll = false
                case 'k':
                    view.Content.ScrollBy(0, -1)
                    autoScroll = false
                case 'G':
                    view.Content.ScrollToBottom()
                    autoScroll = true
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

func simulateProcess(ch chan<- string) {
    defer close(ch)
    for i := 0; i < 100; i++ {
        ch <- fmt.Sprintf("[%s] Log line %d", time.Now().Format("15:04:05"), i)
        time.Sleep(200 * time.Millisecond)
    }
}
```

### 5.2 Key Benefits

1. **Direct element access**: `view.Content.ScrollToBottom()` instead of wrapping or passing refs
2. **Type-safe**: Compiler catches typos in ref names
3. **Discoverable**: Autocomplete shows available refs
4. **Declarative structure**: UI layout stays in `.tui` file
5. **Imperative behavior**: Scroll, add children, etc. in Go

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
dash.Main.AddChild(StreamBox(pollFn).Root)
```

---

## 6. Rules and Constraints

1. **`#Name` must be valid Go identifier** (start with letter, alphanumeric)
2. **Names must be unique within a component** (analyzer error if duplicate)
3. **PascalCase recommended** for consistency with Go exported fields
4. **`Root` is always present** in generated struct (the outermost element)
5. **Backwards compatible**: Components without `#Name` return `*element.Element`
6. **Deeply nested refs work**: `#Name` can be on any element at any depth

---

## 7. Complexity Assessment

| Size | Phases | When to Use |
|------|--------|-------------|
| Small | 1-2 | Single component, bug fix, minor enhancement |
| Medium | 3-4 | New feature touching multiple files/components |
| Large | 5-6 | Cross-cutting feature, new subsystem |

**Assessed Size:** Medium\
**Recommended Phases:** 3

**Rationale:**
- Lexer change: trivial (add `#` token)
- AST change: trivial (add field)
- Parser change: small (detect `#Name` after tag)
- Analyzer change: small (validate uniqueness, add `onUpdate`)
- Generator change: moderate (struct generation, return type logic)
- Testing: moderate (new syntax needs comprehensive tests)

### Phase Breakdown

1. **Phase 1: onUpdate attribute** (Small)
   - Add `onUpdate` to `knownAttributes` in analyzer.go
   - Add `onUpdate` to `attributeToOption` in generator.go
   - Update streaming-dsl example to use `onUpdate` in DSL

2. **Phase 2: Lexer, AST, Parser for #Name** (Small)
   - Add `TokenHash` to lexer
   - Add `NamedRef` field to Element AST
   - Parse `#Name` syntax in parser

3. **Phase 3: Generator struct returns** (Medium)
   - Collect named refs from component
   - Generate view struct type
   - Modify return type and return statement
   - Update streaming-dsl example to use named refs

---

## 8. Success Criteria

1. `onUpdate={fn}` attribute works in DSL and generates `element.WithOnUpdate(fn)`
2. `#Name` syntax parses without error on any element
3. Duplicate `#Name` within component produces analyzer error
4. Invalid `#Name` (e.g., `#123invalid`) produces analyzer error
5. Component with `#Name` elements returns struct type `ComponentNameView`
6. Struct contains `Root` plus all named elements as fields
7. Component without `#Name` still returns `*element.Element` (backwards compatible)
8. Nested `#Name` elements at any depth are captured in struct
9. Generated code compiles and runs correctly
10. streaming-dsl example works with new pattern

---

## 9. Open Questions

1. **Should `Root` be renamed if there's a `#Root` element?**
   → Recommend: disallow `#Root` as a name (reserved)

2. **What about naming elements inside `@for` loops?**
   → Each iteration would overwrite. Recommend: analyzer warning or error for `#Name` inside loops

3. **Should there be a shorthand for common patterns like scrollable content?**
   → Deferred. Users can create their own component library.

4. **What if user wants both root element AND named refs separately?**
   → The struct always has `Root`. User can access `view.Root` for the root.
