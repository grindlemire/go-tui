# Phase 6: Focus & Rendering Improvements Specification

**Status:** Draft
**Version:** 3.0
**Last Updated:** 2025-01-24

---

## 1. Overview

### Purpose

Phase 5 delivered a working event system with focus management, but the API has ergonomic issues. This phase simplifies the design by:

1. Making Element itself focusable (no separate Focusable interface needed)
2. Adding text content directly to Element (no separate Text type needed)
3. Auto-registering focusables when added to tree
4. Cascading focus state down to children via Focus()/Blur() calls

### Goals

- **Element is focusable**: Add `focusable`, `focused`, `onFocus`, `onBlur`, `onEvent` to Element
- **Text on Element**: Add `text`, `textStyle`, `textAlign` to Element
- **Auto-registration**: When a focusable element is added via `AddChild`, notify App
- **Focus cascades down**: Focus()/Blur() recursively call children
- **Single render traversal**: Render just reads current `focused` state

### Non-Goals

- Automatic Tab/Shift+Tab handling (user still controls navigation)
- Mouse support
- Multi-line text (single line for now)

---

## 2. Architecture

### Element Changes

```go
type Element struct {
    // Layout (existing)
    children []*Element
    parent   *Element
    style    layout.Style
    layout   layout.Layout
    dirty    bool

    // Visual (existing)
    border      tui.BorderStyle
    borderStyle tui.Style
    background  *tui.Style

    // Text (new)
    text      string
    textStyle tui.Style
    textAlign TextAlign

    // Focus (new - focusable is set implicitly by WithOnFocus/WithOnBlur/WithOnEvent)
    focusable bool
    focused   bool
    onFocus   func()
    onBlur    func()
    onEvent   func(tui.Event) bool

    // Tree notification (new)
    onChildAdded func(*Element)
}
```

### Traversal Model

```
Focus Change (infrequent - only when user presses Tab):
┌──────────────────────────────────────────────────────────────────┐
│  FocusManager.Next()                                             │
│  1. Call Blur() on old focused element (cascades to children)    │
│  2. Call Focus() on new focused element (cascades to children)   │
└──────────────────────────────────────────────────────────────────┘

Render (every frame):
┌──────────────────────────────────────────────────────────────────┐
│  RenderTree(root)                                                │
│  Single traversal:                                               │
│  1. Render background                                            │
│  2. Render border                                                │
│  3. Render text if non-empty                                     │
│  4. Recurse to children                                          │
│  (reads elem.focused for focus-aware styling if needed)          │
└──────────────────────────────────────────────────────────────────┘
```

---

## 3. Core Entity Changes

### 3.1 Element Text Fields

```go
// pkg/tui/element/element.go

// TextAlign specifies how text is aligned within its content area.
type TextAlign int

const (
    TextAlignLeft TextAlign = iota
    TextAlignCenter
    TextAlignRight
)

// WithText sets the text content and calculates intrinsic size.
func WithText(content string) Option {
    return func(e *Element) {
        e.text = content
        // Set intrinsic size: width = text width, height = 1
        e.style.Width = layout.Fixed(stringWidth(content))
        e.style.Height = layout.Fixed(1)
    }
}

// WithTextStyle sets the style for text content.
func WithTextStyle(style tui.Style) Option

// WithTextAlign sets text alignment within the content area.
func WithTextAlign(align TextAlign) Option

// SetText updates text content and recalculates intrinsic width.
func (e *Element) SetText(content string) {
    e.text = content
    e.style.Width = layout.Fixed(stringWidth(content))
    e.MarkDirty()
}

// Text returns the text content.
func (e *Element) Text() string

// SetTextStyle sets the text style.
func (e *Element) SetTextStyle(style tui.Style)

// TextStyle returns the text style.
func (e *Element) TextStyle() tui.Style
```

### 3.2 Element Focus Fields

```go
// pkg/tui/element/element.go

// WithOnFocus sets the callback for when this element gains focus.
// Implicitly sets focusable = true.
func WithOnFocus(fn func()) Option {
    return func(e *Element) {
        e.focusable = true
        e.onFocus = fn
    }
}

// WithOnBlur sets the callback for when this element loses focus.
// Implicitly sets focusable = true.
func WithOnBlur(fn func()) Option {
    return func(e *Element) {
        e.focusable = true
        e.onBlur = fn
    }
}

// WithOnEvent sets the event handler for this element.
// Implicitly sets focusable = true.
func WithOnEvent(fn func(tui.Event) bool) Option {
    return func(e *Element) {
        e.focusable = true
        e.onEvent = fn
    }
}

// IsFocusable returns whether this element can receive focus.
func (e *Element) IsFocusable() bool {
    return e.focusable
}

// IsFocused returns whether this element currently has focus.
func (e *Element) IsFocused() bool {
    return e.focused
}

// Focus marks this element and all children as focused.
// Calls onFocus callback if set, then cascades to children.
func (e *Element) Focus() {
    e.focused = true
    if e.onFocus != nil {
        e.onFocus()
    }
    for _, child := range e.children {
        child.Focus()
    }
}

// Blur marks this element and all children as not focused.
// Calls onBlur callback if set, then cascades to children.
func (e *Element) Blur() {
    e.focused = false
    if e.onBlur != nil {
        e.onBlur()
    }
    for _, child := range e.children {
        child.Blur()
    }
}

// HandleEvent dispatches an event to this element's handler.
// Returns true if the event was consumed.
func (e *Element) HandleEvent(event tui.Event) bool {
    if e.onEvent != nil {
        return e.onEvent(event)
    }
    return false
}
```

### 3.3 Child Notification

```go
// pkg/tui/element/element.go

// AddChild appends children to this Element.
// Notifies root's onChildAdded callback for each child.
func (e *Element) AddChild(children ...*Element) {
    for _, child := range children {
        child.parent = e
        e.children = append(e.children, child)
        e.notifyChildAdded(child)
    }
    e.MarkDirty()
}

// notifyChildAdded walks up to root and calls onChildAdded if set.
func (e *Element) notifyChildAdded(child *Element) {
    root := e
    for root.parent != nil {
        root = root.parent
    }
    if root.onChildAdded != nil {
        root.onChildAdded(child)
    }
}

// SetOnChildAdded sets the callback for when any descendant is added.
func (e *Element) SetOnChildAdded(fn func(*Element))
```

### 3.4 RenderTree Changes

```go
// pkg/tui/element/render.go

// renderElement renders a single element and recurses to its children.
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

    // 3. Draw text content if present
    if e.text != "" {
        renderTextContent(buf, e)
    }

    // 4. Recurse to children
    for _, child := range e.children {
        renderElement(buf, child)
    }
}

// renderTextContent draws text within the element's content rect.
func renderTextContent(buf *tui.Buffer, e *Element) {
    contentRect := e.ContentRect()
    if contentRect.IsEmpty() {
        return
    }

    textWidth := stringWidth(e.text)
    x := contentRect.X

    // Apply alignment if element is wider than text
    if contentRect.Width > textWidth {
        switch e.textAlign {
        case TextAlignCenter:
            x += (contentRect.Width - textWidth) / 2
        case TextAlignRight:
            x += contentRect.Width - textWidth
        }
    }

    buf.SetString(x, contentRect.Y, e.text, e.textStyle)
}
```

### 3.5 App Changes

```go
// pkg/tui/app.go

// SetRoot sets the root element and sets up focusable discovery.
func (a *App) SetRoot(root *element.Element) {
    a.root = root

    // Set up notification for new focusable elements
    root.SetOnChildAdded(func(child *element.Element) {
        if child.IsFocusable() {
            a.focus.Register(child)
        }
    })

    // Discover existing focusables in tree
    a.discoverFocusables(root)
}

func (a *App) discoverFocusables(elem *element.Element) {
    if elem.IsFocusable() {
        a.focus.Register(elem)
    }
    for _, child := range elem.Children() {
        a.discoverFocusables(child)
    }
}

// FocusNext moves focus to the next focusable element.
func (a *App) FocusNext() {
    a.focus.Next()
}

// FocusPrev moves focus to the previous focusable element.
func (a *App) FocusPrev() {
    a.focus.Prev()
}

// Focused returns the currently focused element.
func (a *App) Focused() *element.Element {
    return a.focus.Focused()
}
```

### 3.6 FocusManager Changes

```go
// pkg/tui/focus.go

type FocusManager struct {
    elements []*element.Element
    current  int
}

func NewFocusManager() *FocusManager {
    return &FocusManager{current: -1}
}

func (f *FocusManager) Register(elem *element.Element) {
    f.elements = append(f.elements, elem)
    // Focus first element if this is the first one
    if len(f.elements) == 1 {
        f.current = 0
        elem.Focus()
    }
}

func (f *FocusManager) Next() {
    if len(f.elements) == 0 {
        return
    }
    // Blur current
    if f.current >= 0 && f.current < len(f.elements) {
        f.elements[f.current].Blur()
    }
    // Move to next
    f.current = (f.current + 1) % len(f.elements)
    f.elements[f.current].Focus()
}

func (f *FocusManager) Prev() {
    if len(f.elements) == 0 {
        return
    }
    // Blur current
    if f.current >= 0 && f.current < len(f.elements) {
        f.elements[f.current].Blur()
    }
    // Move to previous (wrap around)
    f.current--
    if f.current < 0 {
        f.current = len(f.elements) - 1
    }
    f.elements[f.current].Focus()
}

func (f *FocusManager) Focused() *element.Element {
    if f.current >= 0 && f.current < len(f.elements) {
        return f.elements[f.current]
    }
    return nil
}

func (f *FocusManager) Dispatch(event tui.Event) bool {
    if f.current >= 0 && f.current < len(f.elements) {
        return f.elements[f.current].HandleEvent(event)
    }
    return false
}
```

---

## 4. User Experience

### 4.1 Creating Elements with Text

```go
// Before: Separate Text type, manual rendering
title := element.NewText("Hello World",
    element.WithTextStyle(tui.NewStyle().Bold()),
)
root.AddChild(title.Element)
// Must call element.RenderText(buf, title) separately!

// After: Text on Element, automatic rendering
title := element.New(
    element.WithText("Hello World"),
    element.WithTextStyle(tui.NewStyle().Bold()),
)
root.AddChild(title)
// Rendered automatically by RenderTree
```

### 4.2 Creating Focusable Elements

```go
// Before: Separate Focusable interface, manual registration
type FocusableBox struct {
    *element.Element
    focused bool
}
func (b *FocusableBox) IsFocusable() bool { return true }
func (b *FocusableBox) Focus() { ... }
func (b *FocusableBox) Blur() { ... }
func (b *FocusableBox) HandleEvent(e Event) bool { ... }

app.Focus().Register(box)

// After: Focus callbacks imply focusable, auto-registration
box := element.New(
    element.WithSize(20, 5),
    element.WithBorder(tui.BorderSingle),
    element.WithBorderStyle(normalStyle),
    element.WithOnFocus(func() {  // Implies focusable = true
        box.SetBorder(tui.BorderDouble)
        box.SetBorderStyle(focusedStyle)
    }),
    element.WithOnBlur(func() {
        box.SetBorder(tui.BorderSingle)
        box.SetBorderStyle(normalStyle)
    }),
)
root.AddChild(box)  // Auto-registered!
```

### 4.3 Focus-Aware Children

```go
// Child receives Focus()/Blur() calls when parent is focused
var label *element.Element
label = element.New(
    element.WithText("Label"),
    element.WithTextStyle(normalStyle),
    element.WithOnFocus(func() {
        label.SetTextStyle(boldStyle)
    }),
    element.WithOnBlur(func() {
        label.SetTextStyle(normalStyle)
    }),
)

var box *element.Element
box = element.New(
    element.WithOnFocus(func() {  // Implies focusable = true
        box.SetBorder(tui.BorderDouble)
    }),
)
box.AddChild(label)

// When box.Focus() is called, label.Focus() is called automatically
```

### 4.4 Complete Example

```go
func main() {
    app, _ := tui.NewApp()
    defer app.Close()

    width, height := app.Size()

    root := element.New(
        element.WithSize(width, height),
        element.WithDirection(layout.Column),
        element.WithPadding(2),
        element.WithGap(2),
    )

    title := element.New(
        element.WithText("Focus Demo"),
        element.WithTextStyle(tui.NewStyle().Bold()),
    )

    box1 := createBox("Box 1", tui.Red)
    box2 := createBox("Box 2", tui.Blue)

    root.AddChild(title, box1, box2)
    app.SetRoot(root)

    for {
        event, ok := app.PollEvent(50 * time.Millisecond)
        if ok {
            switch e := event.(type) {
            case tui.KeyEvent:
                if e.Key == tui.KeyEscape {
                    return
                }
                if e.Key == tui.KeyTab {
                    if e.Mod.Has(tui.ModShift) {
                        app.FocusPrev()
                    } else {
                        app.FocusNext()
                    }
                } else {
                    app.Dispatch(event)
                }
            }
        }
        app.Render()
    }
}

func createBox(label string, color tui.Color) *element.Element {
    normalStyle := tui.NewStyle().Foreground(color)
    focusedStyle := tui.NewStyle().Foreground(tui.White).Background(color)

    // Declare box first so closures can capture it
    var box *element.Element
    box = element.New(
        element.WithSize(20, 3),
        element.WithBorder(tui.BorderSingle),
        element.WithBorderStyle(normalStyle),
        element.WithJustify(layout.JustifyCenter),
        element.WithAlign(layout.AlignCenter),
        element.WithOnFocus(func() {  // Implies focusable = true
            box.SetBorderStyle(focusedStyle)
            box.SetBorder(tui.BorderDouble)
        }),
        element.WithOnBlur(func() {
            box.SetBorderStyle(normalStyle)
            box.SetBorder(tui.BorderSingle)
        }),
    )

    labelElem := element.New(
        element.WithText(label),
        element.WithTextStyle(tui.NewStyle().Foreground(tui.White)),
    )
    box.AddChild(labelElem)

    return box
}
```

---

## 5. Backward Compatibility

### Text Type

Remove the `Text` struct entirely. Use `WithText` on Element instead:

```go
// Before
title := element.NewText("Hello", element.WithTextStyle(style))
root.AddChild(title.Element)

// After
title := element.New(
    element.WithText("Hello"),
    element.WithTextStyle(style),
)
root.AddChild(title)
```

### Focusable Interface

Element now satisfies the Focusable interface:

```go
type Focusable interface {
    IsFocusable() bool
    HandleEvent(event Event) bool
    Focus()
    Blur()
}

var _ Focusable = (*element.Element)(nil)  // Element satisfies Focusable
```

---

## 6. Implementation Phases

### Phase 1: Text on Element
- Add `text`, `textStyle`, `textAlign` fields to Element
- Add `WithText`, `WithTextStyle`, `WithTextAlign` options
- Add `SetText`, `Text`, `SetTextStyle`, `TextStyle` methods
- Update `renderElement` to render text content
- Remove `Text` struct and related code
- Update tests

### Phase 2: Focus on Element
- Add `focusable`, `focused`, `onFocus`, `onBlur`, `onEvent` fields
- Add `WithOnFocus`, `WithOnBlur`, `WithOnEvent` options (each implies focusable=true)
- Implement `Focus()` and `Blur()` with child cascade
- Implement `HandleEvent()` delegation
- Update tests

### Phase 3: Auto-Registration
- Add `onChildAdded` field to Element
- Update `AddChild` to call `notifyChildAdded`
- Update `App.SetRoot` to set up callback and discover focusables
- Add `App.FocusNext`, `App.FocusPrev`, `App.Focused`
- Update `FocusManager` to work with `*Element`
- Update examples
- Update tests

---

## 7. Testing Strategy

### Text Tests

```go
func TestElementText(t *testing.T) {
    type tc struct {
        content  string
        wantW    int
        wantH    int
    }

    tests := map[string]tc{
        "simple text": {
            content: "Hello",
            wantW:   5,
            wantH:   1,
        },
        "empty text": {
            content: "",
            wantW:   0,
            wantH:   1,
        },
    }

    for name, tt := range tests {
        t.Run(name, func(t *testing.T) {
            e := New(WithText(tt.content))
            if e.style.Width != layout.Fixed(tt.wantW) {
                t.Errorf("width = %v, want Fixed(%d)", e.style.Width, tt.wantW)
            }
        })
    }
}
```

### Focus Tests

```go
func TestFocusCascade(t *testing.T) {
    parent := New(WithOnFocus(func() {}))  // Implies focusable
    child := New()
    grandchild := New()

    parent.AddChild(child)
    child.AddChild(grandchild)

    parent.Focus()

    if !parent.IsFocused() {
        t.Error("parent should be focused")
    }
    if !child.IsFocused() {
        t.Error("child should be focused when parent is focused")
    }
    if !grandchild.IsFocused() {
        t.Error("grandchild should be focused when parent is focused")
    }
}

func TestFocusCallbacks(t *testing.T) {
    focusCalled := false
    blurCalled := false

    e := New(
        WithOnFocus(func() { focusCalled = true }),
        WithOnBlur(func() { blurCalled = true }),
    )

    if !e.IsFocusable() {
        t.Error("element with onFocus should be focusable")
    }

    e.Focus()
    if !focusCalled {
        t.Error("onFocus should be called")
    }

    e.Blur()
    if !blurCalled {
        t.Error("onBlur should be called")
    }
}
```

### Auto-Registration Tests

```go
func TestAutoRegistration(t *testing.T) {
    root := New()

    registered := []*Element{}
    root.SetOnChildAdded(func(child *Element) {
        if child.IsFocusable() {
            registered = append(registered, child)
        }
    })

    box := New(WithOnFocus(func() {}))  // Implies focusable
    root.AddChild(box)

    if len(registered) != 1 || registered[0] != box {
        t.Error("focusable child should be registered")
    }
}
```

---

## 8. Success Criteria

1. Element has `text`, `textStyle`, `textAlign` fields
2. `RenderTree` renders text automatically
3. `Text` struct is removed
4. Element has `focusable`, `focused`, `onFocus`, `onBlur`, `onEvent` fields
5. `WithOnFocus`/`WithOnBlur`/`WithOnEvent` imply `focusable = true`
6. `Focus()` and `Blur()` cascade to children
7. `AddChild` notifies App when focusable elements are added
8. Examples simplified to use new API
9. Single render traversal (Focus/Blur only traverse on focus change)

---

## 9. Summary

| Change | Description |
|--------|-------------|
| Text on Element | `text`, `textStyle`, `textAlign` fields; rendered in RenderTree |
| Remove Text type | `Text` struct removed; use `WithText` on Element |
| Focus on Element | `focusable`, `focused`, `onFocus`, `onBlur`, `onEvent` fields |
| Implicit focusable | `WithOnFocus`/`WithOnBlur`/`WithOnEvent` imply `focusable = true` |
| Focus cascade | `Focus()`/`Blur()` call children recursively |
| Auto-registration | `AddChild` notifies root via `onChildAdded` callback |
| Single render | Render just reads `focused` state; no extra traversal per frame |
