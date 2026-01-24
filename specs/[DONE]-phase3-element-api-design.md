# Phase 3: Element API Specification

**Status:** Planned
**Version:** 1.1
**Last Updated:** 2025-01-24

---

## 1. Overview

### Purpose

The Element API provides a unified, intuitive abstraction for building TUI layouts with visual properties. It replaces the separate `layout.Node` type with an interface-based design where `Element` implements `layout.Layoutable` directly, eliminating the need for wrapper types or dual tree structures.

Currently, users must:
1. Build a `layout.Node` tree with layout styles
2. Call `layout.Calculate()`
3. Extract `node.Layout.Rect` for each node
4. Manually call `tui.DrawBox()` with visual properties

With the Element API, users will:
1. Build an `element.Element` tree with both layout and visual properties
2. Call `root.Render(buf, width, height)` — done

### Goals

- Provide an intuitive, composable API for building TUI layouts
- Eliminate manual bridging between layout calculation and rendering
- Use functional options pattern (`WithX`) for clean, flexible configuration
- Single source of truth for children (Element owns its children directly)
- Refactor `pkg/layout` to use interface instead of concrete `Node` type
- Support separate `Text` element for text content
- Enable advanced users to inspect layout before rendering

### Non-Goals

- Widget system with state/events (Phase 4+)
- Input handling (Phase 4)
- Animation or transitions
- Replacing the low-level `tui` drawing functions (those remain available)
- Keeping `layout.Node` as a separate type (it will be removed)

---

## 2. Architecture

### Directory Structure

```
pkg/layout/
├── layoutable.go       # Layoutable interface definition
├── rect.go             # Rect type (unchanged)
├── point.go            # Point type (unchanged)
├── edges.go            # Edges type (unchanged)
├── value.go            # Value type (unchanged)
├── style.go            # Style type (unchanged)
├── layout.go           # Layout result type
├── calculate.go        # Layout algorithm (refactored for interface)
├── flex.go             # Flexbox calculation (refactored for interface)
└── *_test.go           # Tests

pkg/tui/element/
├── element.go          # Element type implementing Layoutable
├── element_test.go
├── options.go          # Functional option definitions
├── options_test.go
├── text.go             # Text element type
├── text_test.go
├── render.go           # Tree rendering logic
└── render_test.go
```

### Component Overview

| Component | Purpose |
|-----------|---------|
| `layout.Layoutable` | Interface for anything that can participate in layout |
| `layout.Calculate()` | Layout algorithm that works with Layoutable interface |
| `element.Element` | Primary container type, implements Layoutable |
| `element.Text` | Text content element, embeds Element |
| `element.Render()` | Tree traversal and rendering |

### Import Graph

```
             ┌─────────────┐
             │ user code   │
             └──────┬──────┘
                    │ imports
                    ▼
             ┌─────────────┐
             │ tui/element │
             └──────┬──────┘
                    │ imports
          ┌─────────┴─────────┐
          ▼                   ▼
   ┌─────────────┐     ┌─────────────┐
   │    tui      │     │   layout    │
   └─────────────┘     └─────────────┘
                       (Element implements
                        Layoutable)
```

**Key design**: `pkg/layout` defines the `Layoutable` interface. `pkg/tui/element` provides `Element` which implements it. The layout package has no knowledge of Element—it only knows about the interface.

---

## 3. Core Entities

### 3.1 Layoutable Interface (layout package)

```go
// pkg/layout/layoutable.go

// Layoutable is the interface for anything that can participate in layout calculation.
// The layout engine works entirely with this interface, enabling custom implementations.
type Layoutable interface {
    // LayoutStyle returns the layout style properties for this element.
    LayoutStyle() Style

    // LayoutChildren returns the children to be laid out.
    LayoutChildren() []Layoutable

    // SetLayout is called by the layout engine to store computed layout.
    SetLayout(Layout)

    // GetLayout returns the last computed layout.
    GetLayout() Layout

    // IsDirty returns whether this element needs layout recalculation.
    IsDirty() bool

    // ClearDirty marks this element as no longer needing recalculation.
    ClearDirty()
}
```

### 3.2 Layout Result Type

```go
// pkg/layout/layout.go

// Layout holds the computed position and size after layout calculation.
type Layout struct {
    // Rect is the border box—the space allocated by the parent.
    Rect Rect

    // ContentRect is Rect minus padding—where children are placed.
    ContentRect Rect
}
```

### 3.3 Element (implements Layoutable)

```go
// pkg/tui/element/element.go

// Element is a layout container with visual properties.
// It implements layout.Layoutable and owns its children directly.
type Element struct {
    // Tree structure (single source of truth)
    children []*Element
    parent   *Element

    // Layout properties
    style  layout.Style
    layout layout.Layout
    dirty  bool

    // Visual properties
    border      tui.BorderStyle
    borderStyle tui.Style
    background  *tui.Style  // nil = transparent
}

// Implement layout.Layoutable interface

func (e *Element) LayoutStyle() layout.Style { return e.style }

func (e *Element) LayoutChildren() []layout.Layoutable {
    result := make([]layout.Layoutable, len(e.children))
    for i, child := range e.children {
        result[i] = child
    }
    return result
}

func (e *Element) SetLayout(l layout.Layout) { e.layout = l }
func (e *Element) GetLayout() layout.Layout  { return e.layout }
func (e *Element) IsDirty() bool             { return e.dirty }
func (e *Element) ClearDirty()               { e.dirty = false }

// Element's own API

// New creates a new Element with the given options.
// By default, an Element has Auto width/height (flexes to fill available space).
func New(opts ...Option) *Element

// AddChild appends children to this Element.
func (e *Element) AddChild(children ...*Element)

// RemoveChild removes a child from this Element.
func (e *Element) RemoveChild(child *Element) bool

// Children returns the child elements.
func (e *Element) Children() []*Element

// Calculate computes layout for this Element and all descendants.
func (e *Element) Calculate(availableWidth, availableHeight int)

// Rect returns the computed border box.
func (e *Element) Rect() layout.Rect

// ContentRect returns the computed content area.
func (e *Element) ContentRect() layout.Rect

// Render calculates layout (if needed) and renders the entire tree to the buffer.
func (e *Element) Render(buf *tui.Buffer, width, height int)

// MarkDirty marks this Element and ancestors as needing recalculation.
func (e *Element) MarkDirty()
```

### 3.4 Option (Functional Options)

```go
// pkg/tui/element/options.go

// Option configures an Element.
type Option func(*Element)

// --- Dimension Options ---
func WithWidth(cells int) Option
func WithWidthPercent(percent float64) Option
func WithHeight(cells int) Option
func WithHeightPercent(percent float64) Option
func WithSize(width, height int) Option
func WithMinWidth(cells int) Option
func WithMinHeight(cells int) Option
func WithMaxWidth(cells int) Option
func WithMaxHeight(cells int) Option

// --- Flex Container Options ---
func WithDirection(d layout.Direction) Option
func WithJustify(j layout.Justify) Option
func WithAlign(a layout.Align) Option
func WithGap(cells int) Option

// --- Flex Item Options ---
func WithFlexGrow(factor float64) Option
func WithFlexShrink(factor float64) Option
func WithAlignSelf(a layout.Align) Option

// --- Spacing Options ---
func WithPadding(cells int) Option
func WithPaddingTRBL(top, right, bottom, left int) Option
func WithMargin(cells int) Option
func WithMarginTRBL(top, right, bottom, left int) Option

// --- Visual Options ---
func WithBorder(style tui.BorderStyle) Option
func WithBorderStyle(style tui.Style) Option
func WithBackground(style tui.Style) Option
```

### 3.5 Text Element

```go
// pkg/tui/element/text.go

// Text is an Element variant that displays text content.
type Text struct {
    *Element
    content      string
    contentStyle tui.Style
    align        TextAlign
}

type TextAlign int

const (
    TextAlignLeft TextAlign = iota
    TextAlignCenter
    TextAlignRight
)

func NewText(content string, opts ...TextOption) *Text
func (t *Text) SetContent(content string)
func (t *Text) Content() string

// TextOption configures a Text element.
type TextOption func(*Text)

func WithTextStyle(style tui.Style) TextOption
func WithTextAlign(align TextAlign) TextOption
func WithElementOption(opt Option) TextOption
```

---

## 4. Layout Package Refactoring

### 4.1 Calculate Function (Refactored)

```go
// pkg/layout/calculate.go

// Calculate computes layout for the tree rooted at root.
func Calculate(root Layoutable, availableWidth, availableHeight int) {
    calculateNode(root, Rect{0, 0, availableWidth, availableHeight})
}

func calculateNode(node Layoutable, available Rect) {
    if !node.IsDirty() {
        return
    }

    style := node.LayoutStyle()

    // Compute border box
    borderBox := computeBorderBox(style, available)
    contentRect := borderBox.Inset(style.Padding)

    // Layout children using flexbox algorithm
    layoutChildren(node, style, contentRect)

    // Store computed layout via interface
    node.SetLayout(Layout{
        Rect:        borderBox,
        ContentRect: contentRect,
    })

    node.ClearDirty()
}
```

### 4.2 Flex Algorithm (Refactored)

```go
// pkg/layout/flex.go

func layoutChildren(parent Layoutable, style Style, contentRect Rect) {
    children := parent.LayoutChildren()
    if len(children) == 0 {
        return
    }

    isRow := style.Direction == Row

    // Phase 1: Compute base sizes and flex factors
    items := make([]flexItem, len(children))
    for i, child := range children {
        childStyle := child.LayoutStyle()
        items[i] = flexItem{
            node:     child,
            grow:     childStyle.FlexGrow,
            shrink:   childStyle.FlexShrink,
            baseSize: resolveBaseSize(childStyle, isRow, contentRect),
        }
    }

    // Phase 2-6: Flex distribution, constraints, positioning
    // (Same algorithm as before, but uses Layoutable interface)

    // ... distribute free space ...
    // ... apply min/max constraints ...
    // ... calculate positions ...

    // Phase 6: Recurse to children
    for i, child := range children {
        childRect := items[i].computedRect(contentRect, isRow)

        // Apply child's margin
        childStyle := child.LayoutStyle()
        childBorderBox := childRect.Inset(childStyle.Margin)

        calculateNode(child, childBorderBox)
    }
}

type flexItem struct {
    node      Layoutable  // Changed from *Node
    baseSize  int
    mainSize  int
    crossSize int
    mainPos   int
    crossPos  int
    grow      float64
    shrink    float64
}
```

### 4.3 Files to Remove

The following files in `pkg/layout` will be removed:
- `node.go` — replaced by Layoutable interface
- `node_test.go` — tests moved to element package or interface tests

---

## 5. Rendering

### 5.1 Render Implementation

```go
// pkg/tui/element/render.go

// RenderTree traverses the Element tree and renders to the buffer.
func RenderTree(buf *tui.Buffer, root *Element) {
    renderElement(buf, root)
}

func renderElement(buf *tui.Buffer, e *Element) {
    rect := e.Rect()

    // Skip if outside buffer bounds
    if !rect.Intersects(buf.Rect()) {
        return
    }

    // 1. Fill background
    if e.background != nil {
        buf.Fill(rect, ' ', *e.background)
    }

    // 2. Draw border
    if e.border != tui.BorderNone {
        tui.DrawBox(buf, rect, e.border, e.borderStyle)
    }

    // 3. Recurse to children (direct access - no type assertions!)
    for _, child := range e.children {
        renderElement(buf, child)
    }
}

func renderText(buf *tui.Buffer, t *Text) {
    renderElement(buf, t.Element)

    contentRect := t.ContentRect()
    x := contentRect.X

    switch t.align {
    case TextAlignCenter:
        x += (contentRect.Width - runeWidth(t.content)) / 2
    case TextAlignRight:
        x += contentRect.Width - runeWidth(t.content)
    }

    buf.SetString(x, contentRect.Y, t.content, t.contentStyle)
}
```

### 5.2 Element.Render Method

```go
func (e *Element) Render(buf *tui.Buffer, width, height int) {
    if e.dirty {
        layout.Calculate(e, width, height)
    }
    RenderTree(buf, e)
}
```

---

## 6. User Experience

### 6.1 Complete Example

```go
package main

import (
    "os"

    "github.com/grindlemire/go-tui/pkg/layout"
    "github.com/grindlemire/go-tui/pkg/tui"
    "github.com/grindlemire/go-tui/pkg/tui/element"
)

func main() {
    term, _ := tui.NewANSITerminal(os.Stdout, os.Stdin)
    term.EnterRawMode()
    defer term.ExitRawMode()
    term.EnterAltScreen()
    defer term.ExitAltScreen()

    width, height := term.Size()
    buf := tui.NewBuffer(width, height)

    // Build layout tree with visual properties
    root := element.New(
        element.WithSize(width, height),
        element.WithJustify(layout.JustifyCenter),
        element.WithAlign(layout.AlignCenter),
    )

    panel := element.New(
        element.WithSize(40, 10),
        element.WithBorder(tui.BorderRounded),
        element.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan)),
        element.WithPadding(1),
        element.WithDirection(layout.Column),
        element.WithJustify(layout.JustifyCenter),
        element.WithAlign(layout.AlignCenter),
    )

    title := element.NewText("Layout Engine Demo",
        element.WithTextStyle(tui.NewStyle().Bold().Foreground(tui.Green)),
        element.WithTextAlign(element.TextAlignCenter),
    )

    hint := element.NewText("Press any key to exit",
        element.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
    )

    panel.AddChild(title, hint)
    root.AddChild(panel)

    // Single call does everything
    root.Render(buf, width, height)
    tui.Render(term, buf)

    // Wait for keypress
    b := make([]byte, 1)
    os.Stdin.Read(b)
}
```

### 6.2 Before/After Comparison

**Before (75 lines)**:
```go
root := layout.NewNode(layout.DefaultStyle())
root.Style.Width = layout.Fixed(width)
root.Style.Height = layout.Fixed(height)
root.Style.Direction = layout.Column
root.Style.JustifyContent = layout.JustifyCenter
root.Style.AlignItems = layout.AlignCenter

box := layout.NewNode(layout.DefaultStyle())
box.Style.Width = layout.Fixed(40)
box.Style.Height = layout.Fixed(10)

root.AddChild(box)
layout.Calculate(root, width, height)

buf := tui.NewBuffer(width, height)
rect := box.Layout.Rect
boxStyle := tui.NewStyle().Foreground(tui.Cyan)
tui.DrawBox(buf, tui.NewRect(rect.X, rect.Y, rect.Width, rect.Height),
    tui.BorderRounded, boxStyle)

msg := "Layout Engine Demo"
msgX := rect.X + (rect.Width-len(msg))/2
msgY := rect.Y + rect.Height/2
buf.SetString(msgX, msgY, msg, tui.NewStyle().Foreground(tui.Green).Bold())
```

**After (25 lines)**:
```go
root := element.New(
    element.WithSize(width, height),
    element.WithJustify(layout.JustifyCenter),
    element.WithAlign(layout.AlignCenter),
)

panel := element.New(
    element.WithSize(40, 10),
    element.WithBorder(tui.BorderRounded),
    element.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan)),
)

title := element.NewText("Layout Engine Demo",
    element.WithTextStyle(tui.NewStyle().Bold().Foreground(tui.Green)),
)

panel.AddChild(title)
root.AddChild(panel)
root.Render(buf, width, height)
```

---

## 7. Complexity Assessment

| Factor | Assessment |
|--------|------------|
| Layout package refactor | Moderate — remove Node, add interface, update flex.go |
| New element package | 6-8 files |
| Core types | 2 types (Element, Text) + options |
| Algorithm complexity | Low — same flexbox logic, just interface-based |
| Testing scope | Significant — new tests for Element, update layout tests |

**Assessed Size:** Medium
**Recommended Phases:** 3

**Rationale:**
1. **Phase 1**: Refactor layout package to interface-based, create Element type with options
2. **Phase 2**: Add Text type, implement rendering
3. **Phase 3**: Integration tests, update examples, documentation

---

## 8. Success Criteria

1. `layout.Node` is removed; `layout.Layoutable` interface exists
2. `layout.Calculate()` works with any `Layoutable` implementation
3. `element.New()` creates an Element with default Auto dimensions
4. All `WithX` options correctly configure Element's style
5. `AddChild()` maintains Element's children slice (single source of truth)
6. `Calculate()` correctly computes layout via `Layoutable` interface
7. `Render()` draws all elements with their visual properties
8. `Text` correctly renders with alignment
9. No type assertions needed anywhere in the codebase
10. Dashboard example reduced from ~75 lines to ~25 lines

---

## 9. Migration Notes

### Breaking Changes

1. `layout.Node` is removed
2. `layout.Calculate(node, w, h)` signature changes to `layout.Calculate(layoutable, w, h)`
3. Existing code using `layout.Node` must migrate to `element.Element`

### Migration Path

Since the library is not yet released, no migration guide is needed. The examples will be updated to use the new API.

---

## 10. Open Questions

1. **Should Text support multi-line content?**
   → Deferred. Initial version supports single-line.

2. **Should Element support border titles?**
   → Deferred. Users can position a Text child.

3. **Should we provide convenience constructors?**
   → Consider after Phase 1. Start with explicit options.

4. **How to handle invisible layout containers?**
   → Element with no border/background is invisible. Visual properties are optional.
