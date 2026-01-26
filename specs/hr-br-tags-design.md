# HR and BR Tags Specification

**Status:** Planned\
**Version:** 1.1\
**Last Updated:** 2025-01-25

---

## 1. Overview

### Purpose

Add support for `<hr>` (horizontal rule) and `<br>` (line break) elements to the .tui grammar. These are common HTML void elements used for visual separation and line breaking in layouts.

### Goals

- Support both self-closing (`<hr/>`, `<br/>`) and non-self-closing (`<hr>`, `<br>`) syntax forms
- `<hr>` renders as a horizontal line using the Unicode box drawing character `─` (U+2500)
- `<hr>` fills available width by default (block-level behavior)
- `<hr>` supports styling via the `class` attribute and other standard attributes
- `<br>` creates a single empty line (height: 1, width: 0)
- Both elements are void elements (children rejected by analyzer)
- Proper LSP completion support for both tags

### Non-Goals

- Custom line characters via attributes (can be achieved via CSS-like classes in future)
- Multi-line breaks (`<br>` always produces exactly one line)
- Vertical rules (`<vr>` - could be added separately)

---

## 2. Architecture

### Directory Structure

```
pkg/
├── tuigen/
│   ├── analyzer.go     # Add hr/br to knownTags, void element validation
│   └── generator.go    # Special handling for hr/br tag generation
├── tui/element/
│   ├── options.go      # Add WithHR() option
│   ├── element.go      # Add hr field to Element struct
│   └── render.go       # HR line rendering logic
└── lsp/
    └── completion.go   # Add hr/br tag completions
```

### Component Overview

| Component | Purpose |
|-----------|---------|
| `pkg/tuigen/analyzer.go` | Add `hr`, `br` to `knownTags`; validate void elements have no children |
| `pkg/tuigen/generator.go` | Handle `hr` and `br` tags specially in `buildElementOptions()` |
| `pkg/tui/element/options.go` | Add `WithHR()` option that configures HR rendering |
| `pkg/tui/element/element.go` | Add `hr bool` field to Element struct; update `IntrinsicSize()` |
| `pkg/tui/element/render.go` | Add `renderHR()` function for drawing horizontal lines |
| `pkg/lsp/completion.go` | Add `hr` and `br` to `getElementCompletions()` |

### Flow

```
.tui file with <hr/> or <br/>
        │ parser (existing - already handles self-closing)
        ▼
Element AST node with Tag="hr" or Tag="br"
        │ analyzer
        ▼
Validate tag known + void element has no children
        │ generator
        ▼
element.New(element.WithHR()) or element.New(element.WithHeight(1))
        │ layout
        ▼
Proper sizing (hr: width=fill, height=1; br: width=0, height=1)
        │ render
        ▼
HR draws ─ characters across width; BR is empty
```

---

## 3. Core Entities

### 3.1 Void Element Set

Add a set of void elements that cannot have children:

```go
// pkg/tuigen/analyzer.go

// voidElements lists elements that cannot have children.
var voidElements = map[string]bool{
    "hr":    true,
    "br":    true,
    "input": true, // already in knownTags, adding void validation
}
```

> **Note:** `input` is already in `knownTags`. Adding it to `voidElements` adds child validation. Verified no existing `.tui` files use `<input>...</input>` pattern.

### 3.2 Element Struct Changes

```go
// pkg/tui/element/element.go

type Element struct {
    // ... existing fields ...

    // HR properties
    hr bool // true if this element is a horizontal rule
}

// Method to check if element is HR
func (e *Element) IsHR() bool {
    return e.hr
}
```

### 3.3 IntrinsicSize for HR/BR

```go
// pkg/tui/element/element.go - update IntrinsicSize()

func (e *Element) IntrinsicSize() (width, height int) {
    // HR has intrinsic height of 1, but 0 intrinsic width.
    // The 0 width is intentional - HR relies on AlignSelf=Stretch (set by WithHR)
    // to fill the container width, similar to how block elements work in CSS.
    if e.hr {
        return 0, 1
    }

    // ... existing text/children logic ...
}
```

### 3.4 WithHR Option

```go
// pkg/tui/element/options.go

// WithHR configures an element as a horizontal rule.
// The element renders a horizontal line character across its width.
// Uses ─ (U+2500) by default, or other characters based on border style:
//   - BorderDouble → ═ (U+2550)
//   - BorderThick  → ━ (U+2501)
//
// Sets AlignSelf to Stretch so HR fills container width regardless
// of parent's AlignItems setting.
func WithHR() Option {
    return func(e *Element) {
        e.hr = true
        e.style.Height = layout.Fixed(1)
        stretch := layout.AlignStretch
        e.style.AlignSelf = &stretch // Always stretch to fill width
    }
}
```

### 3.5 HR Character Mapping

```go
// pkg/tui/element/render.go

// hrCharacter returns the horizontal rule character based on border style.
func hrCharacter(border tui.BorderStyle) rune {
    switch border {
    case tui.BorderDouble:
        return '═' // U+2550
    case tui.BorderThick:
        return '━' // U+2501
    default:
        return '─' // U+2500
    }
}
```

### 3.6 HR Rendering

```go
// pkg/tui/element/render.go

// renderHR draws a horizontal rule across the element's width.
func renderHR(buf *tui.Buffer, e *Element) {
    rect := e.ContentRect()
    char := hrCharacter(e.border)

    for x := rect.X; x < rect.Right(); x++ {
        buf.SetRune(x, rect.Y, char, e.textStyle)
    }
}

// Update renderElement() to check for HR:
func renderElement(buf *tui.Buffer, e *Element) {
    // ... existing checks ...

    // Handle HR specially
    if e.hr {
        renderHR(buf, e)
        return // HR has no children
    }

    // ... rest of rendering ...
}
```

### 3.7 Generator Changes

```go
// pkg/tuigen/generator.go - update buildElementOptions()

func (g *Generator) buildElementOptions(elem *Element) []string {
    var options []string

    // Handle tag-specific options
    switch elem.Tag {
    case "hr":
        options = append(options, "element.WithHR()")
    case "br":
        // BR is just a zero-width, single-height element
        options = append(options, "element.WithWidth(0)")
        options = append(options, "element.WithHeight(1)")
    case "span", "p":
        // ... existing span/p handling ...
    }

    // ... rest of attribute processing ...
    return options
}
```

### 3.8 Analyzer Validation

```go
// pkg/tuigen/analyzer.go - update analyzeElement()

func (a *Analyzer) analyzeElement(elem *Element) {
    // Check if tag is known
    if !knownTags[elem.Tag] {
        a.errors.AddErrorf(elem.Position, "unknown element tag <%s>", elem.Tag)
    }

    // Check for children on void elements
    if voidElements[elem.Tag] && len(elem.Children) > 0 {
        a.errors.AddErrorf(elem.Position,
            "<%s> is a void element and cannot have children", elem.Tag)
    }

    // ... rest of validation ...
}
```

### 3.9 LSP Completion

```go
// pkg/lsp/completion.go - add to getElementCompletions()

{
    Label:      "hr",
    Kind:       CompletionItemKindClass,
    Detail:     "Horizontal rule",
    InsertText: "hr/>",
    Documentation: &MarkupContent{
        Kind:  "markdown",
        Value: "A horizontal dividing line.\n\n```tui\n<hr/>\n<hr class=\"border-double text-cyan\"/>\n```",
    },
},
{
    Label:      "br",
    Kind:       CompletionItemKindClass,
    Detail:     "Line break",
    InsertText: "br/>",
    Documentation: &MarkupContent{
        Kind:  "markdown",
        Value: "An empty line break.\n\n```tui\n<span>Line 1</span>\n<br/>\n<span>Line 2</span>\n```",
    },
},
```

---

## 4. User Experience

### TUI Syntax Examples

```tui
package example

@component DividerExample() {
    <div class="flex-col gap-1">
        <span>Above the line</span>
        <hr/>
        <span>Below the line</span>
    </div>
}

@component StyledHR() {
    <div class="flex-col">
        <hr class="border-double text-cyan"/>
        <hr class="border-thick"/>
    </div>
}

@component LineBreakExample() {
    <div class="flex-col">
        <span>First line</span>
        <br/>
        <span>Second line</span>
    </div>
}

// Non-self-closing form also works
@component AlternativeSyntax() {
    <div class="flex-col">
        <hr>
        <br>
    </div>
}

// Fixed width HR
@component FixedWidthHR() {
    <div class="flex-col items-center">
        <hr width="20"/>
    </div>
}
```

### Rendered Output

```
Above the line
────────────────────────────────
Below the line
```

### Line Character Mapping

| Class | Character |
|-------|-----------|
| (default) | `─` (U+2500) |
| `border-double` | `═` (U+2550) |
| `border-thick` | `━` (U+2501) |

### Attribute Support

HR and BR support all standard attributes:
- `width`, `height` (override defaults)
- `class` (Tailwind-style classes for styling)
- `id` (identifier)
- Border-related classes change HR line character

### Layout Behavior

**HR in column container (flex-col):**
- Uses `AlignSelf(Stretch)` to fill available width regardless of parent's `AlignItems`
- Works correctly even with `items-center` or `items-start`

**HR in row container (flex):**
- HR stretches on cross axis (height), not main axis
- For horizontal layouts, document that HR should typically be used in column containers
- User can set explicit `width` if needed

### Parser Behavior for Non-Self-Closing

The existing parser treats `<hr>` without a closing tag as attempting to have children:
```tui
<hr><span>Next</span>  // Parser tries to find </hr>
```

This is **parse-then-validate** behavior:
1. Parser parses `<span>` as a child of `<hr>` (or fails looking for `</hr>`)
2. Analyzer rejects children on void elements

The self-closing form (`<hr/>`) is preferred for void elements. Non-self-closing (`<hr>`) followed by content may produce confusing parse errors. This matches the existing behavior for other void-ish elements like `<input>`.

---

## 5. Complexity Assessment

| Size | Phases | When to Use |
|------|--------|-------------|
| Small | 1-2 | Single component, bug fix, minor enhancement |
| Medium | 3-4 | New feature touching multiple files/components |
| Large | 5-6 | Cross-cutting feature, new subsystem |

**Assessed Size:** Medium\
**Recommended Phases:** 3\
**Rationale:** The feature touches 6 files across multiple packages (tuigen, element, lsp). While each change is straightforward, proper void element validation, HR rendering, and LSP support require careful coordination:
1. Phase 1: Analyzer and generator changes
2. Phase 2: Element struct, options, and rendering
3. Phase 3: LSP completions and tests

---

## 6. Test Cases

### Parser Tests (`pkg/tuigen/parser_test.go`)

| Test | Description |
|------|-------------|
| `TestParseHRSelfClosing` | `<hr/>` parses correctly |
| `TestParseHRNonSelfClosing` | `<hr>` parses correctly |
| `TestParseBRSelfClosing` | `<br/>` parses correctly |
| `TestParseBRNonSelfClosing` | `<br>` parses correctly |
| `TestParseHRWithClass` | `<hr class="border-double"/>` parses attributes |

### Analyzer Tests (`pkg/tuigen/analyzer_test.go`)

| Test | Description |
|------|-------------|
| `TestAnalyzeHRValid` | `<hr/>` passes validation |
| `TestAnalyzeBRValid` | `<br/>` passes validation |
| `TestAnalyzeVoidWithChildren` | `<hr>text</hr>` produces error |

### Generator Tests (`pkg/tuigen/generator_test.go`)

| Test | Description |
|------|-------------|
| `TestGenerateHR` | `<hr/>` generates `element.New(element.WithHR())` |
| `TestGenerateBR` | `<br/>` generates `element.New(element.WithWidth(0), element.WithHeight(1))` |
| `TestGenerateHRWithBorder` | `<hr class="border-double"/>` includes border option |
| `TestGenerateHRWithTextClass` | `<hr class="text-cyan"/>` generates proper TextStyle option via Tailwind parser |

### Render Tests (`pkg/tui/element/render_test.go`)

| Test | Description |
|------|-------------|
| `TestRenderHRDefault` | HR draws `─` characters |
| `TestRenderHRDouble` | HR with BorderDouble draws `═` characters |
| `TestRenderHRThick` | HR with BorderThick draws `━` characters |
| `TestRenderHRWithColor` | HR respects textStyle for color |

---

## 7. Success Criteria

1. `<hr/>` and `<hr>` parse without errors
2. `<br/>` and `<br>` parse without errors
3. `<hr>children</hr>` produces analyzer error
4. `<hr>` renders as `─` characters filling available width
5. `<hr class="border-double">` renders as `═` characters
6. `<hr class="border-thick">` renders as `━` characters
7. `<hr>` supports `text-*` classes for color styling
8. `<br>` creates an empty single-line element with width: 0
9. Generated Go code compiles and runs correctly
10. LSP provides proper completions for hr and br tags
11. All tests pass

---

## 8. Open Questions

1. ~~Should syntax be self-closing only or both forms?~~ → Both forms
2. ~~Should HR support styling?~~ → Yes, via class attribute
3. ~~What default character for HR?~~ → `─` (U+2500)
4. ~~Should HR fill width by default?~~ → Yes

### Deferred Items

- Tree-sitter grammar (`editor/tree-sitter-tui/`) - existing grammar handles void elements via the self-closing syntax; no changes expected but verify during implementation
- Vertical rule (`<vr>`) - same pattern, can be added later
