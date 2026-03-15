# Element Reference

## Overview

`Element` is the building block of go-tui UIs. Every visible piece of content (text, borders, containers, scrollable regions) is an Element. Elements form a tree: a root element contains children, which contain their own children, and so on. The layout engine computes positions using CSS flexbox, and the renderer draws the result to a character buffer.

You can create elements two ways: in `.gsx` templates (the common path) or programmatically with `tui.New()`.

```gsx
// In .gsx — the compiler turns this into tui.New(...) calls
<div class="flex-col gap-1 p-1 border-rounded">
    <span class="font-bold text-cyan">Hello</span>
    <span>World</span>
</div>
```

```go
// Programmatic equivalent
root := tui.New(
    tui.WithDirection(tui.Column),
    tui.WithGap(1),
    tui.WithPadding(1),
    tui.WithBorder(tui.BorderRounded),
)
title := tui.New(
    tui.WithText("Hello"),
    tui.WithTextStyle(tui.NewStyle().Bold().Foreground(tui.ANSIColor(tui.Cyan))),
)
body := tui.New(tui.WithText("World"))
root.AddChild(title, body)
```

`Element` implements the `Viewable`, `Focusable`, and `Layoutable` interfaces.

## Creating Elements

### New

```go
func New(opts ...Option) *Element
```

Creates a new Element with the given options. By default, elements use Auto width and height, Row direction, and a transparent background, with border and text unset.

```go
box := tui.New(
    tui.WithWidth(40),
    tui.WithHeight(10),
    tui.WithBorder(tui.BorderRounded),
    tui.WithText("Content"),
)
```

In `.gsx` files, elements are created with HTML-like tags:

```gsx
<div width={40} height={10} border={tui.BorderRounded}>
    <span>Content</span>
</div>
```

## Option Functions

`Option` is defined as `func(*Element)`. Pass options to `tui.New()` to configure an element at creation time. The `.gsx` compiler translates element attributes and Tailwind classes into these same option calls.

### Dimensions

| Function | Description |
|----------|-------------|
| `WithWidth(cells int)` | Fixed width in terminal cells |
| `WithWidthPercent(percent float64)` | Width as a percentage of parent's available width |
| `WithWidthAuto()` | Width sized to content (default) |
| `WithHeight(cells int)` | Fixed height in terminal rows |
| `WithHeightPercent(percent float64)` | Height as a percentage of parent's available height |
| `WithHeightAuto()` | Height sized to content (default) |
| `WithSize(width, height int)` | Sets both width and height in terminal cells |
| `WithMinWidth(cells int)` | Minimum width constraint |
| `WithMinHeight(cells int)` | Minimum height constraint |
| `WithMaxWidth(cells int)` | Maximum width constraint |
| `WithMaxHeight(cells int)` | Maximum height constraint |

```go
// Fixed 40x10 box
tui.New(tui.WithSize(40, 10))

// Half-width panel with minimum
tui.New(tui.WithWidthPercent(50), tui.WithMinWidth(20))
```

### Flex Container

These options control how an element lays out its children.

| Function | Description |
|----------|-------------|
| `WithDirection(d Direction)` | Main axis: `tui.Row` (horizontal, default) or `tui.Column` (vertical) |
| `WithJustify(j Justify)` | Main axis alignment: `JustifyStart`, `JustifyCenter`, `JustifyEnd`, `JustifySpaceBetween`, `JustifySpaceAround`, `JustifySpaceEvenly` |
| `WithAlign(a Align)` | Cross axis alignment: `AlignStart`, `AlignCenter`, `AlignEnd`, `AlignStretch` |
| `WithGap(cells int)` | Space between children on the main axis |

```go
// Vertical layout with centered children and 1-cell gap
tui.New(
    tui.WithDirection(tui.Column),
    tui.WithAlign(tui.AlignCenter),
    tui.WithGap(1),
)
```

### Flex Item

These options control how an element behaves as a child within a flex container.

| Function | Description |
|----------|-------------|
| `WithFlexGrow(factor float64)` | How much this element grows relative to siblings to fill extra space |
| `WithFlexShrink(factor float64)` | How much this element shrinks relative to siblings when space is tight |
| `WithAlignSelf(a Align)` | Overrides the parent's `AlignItems` for this element |

```go
// Sidebar (fixed) + main content (grows to fill)
sidebar := tui.New(tui.WithWidth(30))
content := tui.New(tui.WithFlexGrow(1))
```

### Spacing

| Function | Description |
|----------|-------------|
| `WithPadding(cells int)` | Uniform padding on all sides (inside the border) |
| `WithPaddingTRBL(top, right, bottom, left int)` | Per-side padding in CSS order |
| `WithMargin(cells int)` | Uniform margin on all sides (outside the border) |
| `WithMarginTRBL(top, right, bottom, left int)` | Per-side margin in CSS order |

```go
// 1-cell padding all around, 2-cell top margin
tui.New(
    tui.WithPadding(1),
    tui.WithMarginTRBL(2, 0, 0, 0),
)
```

### Visual

| Function | Description |
|----------|-------------|
| `WithBorder(style BorderStyle)` | Border shape: `BorderNone`, `BorderSingle`, `BorderDouble`, `BorderRounded`, `BorderThick` |
| `WithBorderStyle(style Style)` | Color and attributes for the border lines |
| `WithBackground(style Style)` | Background fill style |
| `WithText(content string)` | Text content for this element |
| `WithTextStyle(style Style)` | Text color and attributes. Setting this prevents style inheritance from the parent |
| `WithTextAlign(align TextAlign)` | Text alignment: `TextAlignLeft` (default), `TextAlignCenter`, `TextAlignRight` |

```go
// Cyan-bordered box with bold white text on a blue background
tui.New(
    tui.WithBorder(tui.BorderRounded),
    tui.WithBorderStyle(tui.NewStyle().Foreground(tui.ANSIColor(tui.Cyan))),
    tui.WithBackground(tui.NewStyle().Background(tui.ANSIColor(tui.Blue))),
    tui.WithText("Status: OK"),
    tui.WithTextStyle(tui.NewStyle().Bold().Foreground(tui.ANSIColor(tui.White))),
)
```

### Gradient

| Function | Description |
|----------|-------------|
| `WithTextGradient(g Gradient)` | Gradient applied per-character to text (overrides `textStyle.Fg`) |
| `WithBackgroundGradient(g Gradient)` | Gradient fill for the element background |
| `WithBorderGradient(g Gradient)` | Gradient applied around the border perimeter |

```go
tui.New(
    tui.WithText("Rainbow"),
    tui.WithTextGradient(tui.NewGradient(
        tui.ANSIColor(tui.Red),
        tui.ANSIColor(tui.Cyan),
    )),
)
```

### Focus

| Function | Description |
|----------|-------------|
| `WithFocusable(focusable bool)` | Whether this element can receive focus |
| `WithOnFocus(fn func(*Element))` | Callback when focus is gained. Implicitly sets `focusable = true` |
| `WithOnBlur(fn func(*Element))` | Callback when focus is lost. Implicitly sets `focusable = true` |
| `WithOnActivate(fn func())` | Callback when Enter is pressed while focused. Implicitly sets `focusable = true` |

```go
tui.New(
    tui.WithFocusable(true),
    tui.WithOnFocus(func(el *tui.Element) {
        el.SetBorderStyle(tui.NewStyle().Foreground(tui.ANSIColor(tui.Cyan)))
    }),
    tui.WithOnBlur(func(el *tui.Element) {
        el.SetBorderStyle(tui.NewStyle())
    }),
)
```

### Scroll

| Function | Description |
|----------|-------------|
| `WithScrollable(mode ScrollMode)` | Enables scrolling: `ScrollVertical`, `ScrollHorizontal`, or `ScrollBoth`. Implicitly sets `focusable = true` and applies default scrollbar styles |
| `WithScrollOffset(x, y int)` | Initial scroll position. Useful for preserving scroll state across re-renders via `State[int]` |
| `WithScrollbarStyle(style Style)` | Style for the scrollbar track |
| `WithScrollbarThumbStyle(style Style)` | Style for the scrollbar thumb |

```go
tui.New(
    tui.WithScrollable(tui.ScrollVertical),
    tui.WithScrollOffset(0, scrollY.Get()),
    tui.WithScrollbarThumbStyle(tui.NewStyle().Foreground(tui.ANSIColor(tui.Cyan))),
)
```

### Behavior

| Function | Description |
|----------|-------------|
| `WithHR()` | Configures the element as a horizontal rule. Sets height to 1 and `AlignSelf` to stretch. Uses `─` by default, `═` for `BorderDouble`, `━` for `BorderThick` |
| `WithTruncate(truncate bool)` | Enables text truncation with ellipsis (`…`) when text overflows the element width |
| `WithHidden(hidden bool)` | Excludes the element from layout and rendering |
| `WithOverflow(mode OverflowMode)` | How content beyond element bounds is handled: `OverflowVisible` (default) or `OverflowHidden` |
| `WithOnUpdate(fn func())` | Hook called before each render. Useful for polling or animation logic |

```go
// Truncated text in a fixed-width cell
tui.New(
    tui.WithWidth(20),
    tui.WithText("This is a very long string that will be truncated"),
    tui.WithTruncate(true),
)
```

## Accessors

Get and set methods for reading and modifying element properties after creation. Setters that affect layout or rendering call `MarkDirty()` automatically.

### Style (Layout)

```go
func (e *Element) Style() LayoutStyle
func (e *Element) SetStyle(style LayoutStyle)
```

Read or replace the full layout style. `SetStyle` marks the element dirty.

### Border

```go
func (e *Element) Border() BorderStyle
func (e *Element) SetBorder(border BorderStyle)
func (e *Element) BorderStyle() Style
func (e *Element) SetBorderStyle(style Style)
```

`Border()` returns the border shape (`BorderSingle`, `BorderRounded`, etc.). `BorderStyle()` returns the color/attribute style used to draw the border lines.

### Background

```go
func (e *Element) Background() *Style
func (e *Element) SetBackground(style *Style)
```

Returns `nil` when the background is transparent. Pass `nil` to `SetBackground` to make it transparent again.

### Text

```go
func (e *Element) Text() string
func (e *Element) SetText(content string)
func (e *Element) TextStyle() Style
func (e *Element) SetTextStyle(style Style)
func (e *Element) TextAlign() TextAlign
func (e *Element) SetTextAlign(align TextAlign)
```

`SetText` marks the element dirty. `SetTextStyle` prevents style inheritance from the parent element for this element's text.

### Truncate

```go
func (e *Element) Truncate() bool
func (e *Element) SetTruncate(truncate bool)
```

When enabled, text that overflows the element's content width is cut off with an ellipsis character (`…`).

### Hidden

```go
func (e *Element) Hidden() bool
func (e *Element) SetHidden(hidden bool)
```

Hidden elements are excluded from both layout and rendering. Their children are also skipped.

### Overflow

```go
func (e *Element) Overflow() OverflowMode
func (e *Element) SetOverflow(mode OverflowMode)
```

Controls whether content that exceeds element bounds is visible or clipped.

## Tree Methods

Elements form a tree. These methods manipulate the parent-child relationships.

### AddChild

```go
func (e *Element) AddChild(children ...*Element)
```

Appends one or more children. Each child's parent is set to this element, and the element is marked dirty. If the root element has callback hooks registered (for focus or child-added notifications), they fire for each new child.

```go
container := tui.New(tui.WithDirection(tui.Column))
container.AddChild(
    tui.New(tui.WithText("First")),
    tui.New(tui.WithText("Second")),
)
```

### RemoveChild

```go
func (e *Element) RemoveChild(child *Element) bool
```

Removes a specific child. Returns `true` if the child was found and removed. The removed child's parent is set to `nil`.

### RemoveAllChildren

```go
func (e *Element) RemoveAllChildren()
```

Removes all children from this element.

### Children

```go
func (e *Element) Children() []*Element
```

Returns the slice of child elements.

### Parent

```go
func (e *Element) Parent() *Element
```

Returns the parent element, or `nil` for the root.

### SetOnChildAdded

```go
func (e *Element) SetOnChildAdded(fn func(*Element))
```

Registers a callback that fires whenever a descendant is added anywhere in this element's subtree. Used internally by the framework for focus management.

## Layout Methods

The layout engine uses these methods during the flexbox calculation pass.

### LayoutStyle / LayoutChildren

```go
func (e *Element) LayoutStyle() LayoutStyle
func (e *Element) LayoutChildren() []Layoutable
```

Part of the `Layoutable` interface. `LayoutStyle` returns the element's style with border padding added (borders consume 1 cell on each side). `LayoutChildren` returns visible children only; hidden elements are excluded.

### SetLayout / GetLayout

```go
func (e *Element) SetLayout(l LayoutResult)
func (e *Element) GetLayout() LayoutResult
```

Called by the layout engine to store and retrieve computed layout results.

### Calculate

```go
func (e *Element) Calculate(availableWidth, availableHeight int)
```

Runs the flexbox layout algorithm on this element and all its descendants. Wraps the package-level `tui.Calculate()` function.

```go
root.Calculate(80, 24) // Layout for an 80x24 terminal
```

### Rect / ContentRect

```go
func (e *Element) Rect() Rect
func (e *Element) ContentRect() Rect
```

`Rect` returns the border box (the full area including border and padding). `ContentRect` returns the inner content area (inside border and padding). Both are populated after `Calculate` runs.

### IntrinsicSize

```go
func (e *Element) IntrinsicSize() (width, height int)
```

Returns the natural content-based dimensions. For text elements, this is the text width plus padding and border. For containers, it's computed from children. Scrollable elements return `(0, 0)` since they rely on explicit sizing or `flexGrow`. Horizontal rules return `(0, 1)`.

### Dirty Flags

```go
func (e *Element) IsDirty() bool
func (e *Element) SetDirty(dirty bool)
func (e *Element) MarkDirty()
```

`MarkDirty` walks up the parent chain setting dirty flags, and also marks the owning App as dirty (triggering a re-render on the next frame). `IsDirty` and `SetDirty` are used by the layout engine.

### IsHR

```go
func (e *Element) IsHR() bool
```

Returns `true` if this element was configured with `WithHR()`.

## Focus Methods

Focus determines which element receives keyboard input.

### Query

```go
func (e *Element) IsFocusable() bool
func (e *Element) IsFocused() bool
func (e *Element) SetFocusable(focusable bool)
```

Check or change whether an element can receive focus, and whether it currently has it.

### Focus / Blur

```go
func (e *Element) Focus()
func (e *Element) Blur()
```

`Focus` marks the element as focused and calls the `onFocus` callback if one is registered. `Blur` clears the focused state and calls `onBlur`. These don't cascade to children. Only the target element is affected.

### SetOnFocus / SetOnBlur

```go
func (e *Element) SetOnFocus(fn func(*Element))
func (e *Element) SetOnBlur(fn func(*Element))
```

Register focus/blur callbacks after creation. Both implicitly set `focusable = true`.

### HandleEvent

```go
func (e *Element) HandleEvent(event Event) bool
```

Dispatches an event to this element. For scrollable elements, handles arrow keys, Page Up/Down, Home/End, and mouse wheel events. Returns `true` if the event was consumed.

Built-in scroll key handling:
- **Up/Down arrows**: scroll vertically by 1 row
- **Left/Right arrows**: scroll horizontally by 1 column
- **Page Up/Page Down**: scroll by viewport height
- **Home**: scroll to top-left
- **End**: scroll to bottom
- **Mouse wheel**: scroll vertically by 1 row per tick

### ContainsPoint

```go
func (e *Element) ContainsPoint(x, y int) bool
```

Returns `true` if the point falls within the element's computed layout bounds. Useful for hit testing in `HandleMouse` implementations.

## Scroll Methods

Scroll methods only work on elements created with `WithScrollable()`.

### Query

```go
func (e *Element) IsScrollable() bool
func (e *Element) ScrollModeValue() ScrollMode
func (e *Element) ScrollOffset() (x, y int)
func (e *Element) ContentSize() (width, height int)
func (e *Element) ViewportSize() (width, height int)
func (e *Element) MaxScroll() (maxX, maxY int)
```

| Method | Returns |
|--------|---------|
| `IsScrollable()` | Whether scrolling is enabled |
| `ScrollModeValue()` | The scroll mode (`ScrollVertical`, `ScrollHorizontal`, `ScrollBoth`) |
| `ScrollOffset()` | Current scroll position |
| `ContentSize()` | Total content dimensions (computed during layout, may exceed viewport) |
| `ViewportSize()` | Visible area dimensions |
| `MaxScroll()` | Maximum valid scroll offset in each direction |

### Control

```go
func (e *Element) ScrollTo(x, y int)
func (e *Element) ScrollBy(dx, dy int)
func (e *Element) ScrollToTop()
func (e *Element) ScrollToBottom()
```

`ScrollTo` sets an absolute position, clamped to valid range. `ScrollBy` adjusts relative to the current position. `ScrollToBottom` scrolls immediately and also sets a pending flag so that after the next layout pass (when content size may have changed), it re-scrolls to the new bottom. This makes it reliable for following new content.

### IsAtBottom

```go
func (e *Element) IsAtBottom() bool
```

Returns `true` if the element is scrolled to the bottom. Useful for implementing sticky-scroll behavior (auto-follow new content, but stop following when the user scrolls up).

### ScrollIntoView

```go
func (e *Element) ScrollIntoView(child *Element)
```

Scrolls the minimum amount needed to make `child` fully visible within this element's viewport. Does nothing if `child` is not a descendant or if scrolling is disabled.

## Watcher and Discovery Methods

Background operations (timers, channel watchers) and tree traversal hooks.

### Watchers

```go
func (e *Element) AddWatcher(w Watcher)
func (e *Element) Watchers() []Watcher
func (e *Element) WalkWatchers(fn func(Watcher))
```

`AddWatcher` attaches a timer or channel watcher to this element. Watchers start automatically when the element tree is set as the app root. `WalkWatchers` traverses the tree and calls `fn` for every watcher (skipping hidden elements).

### Focus Discovery

```go
func (e *Element) WalkFocusables(fn func(Focusable))
func (e *Element) SetOnFocusableAdded(fn func(Focusable))
```

`WalkFocusables` does a depth-first walk calling `fn` for each focusable element. `SetOnFocusableAdded` registers a callback for when new focusable descendants are added. Both are used by the App for automatic focus management.

### Pre-Render Hook

```go
func (e *Element) SetOnUpdate(fn func())
```

Sets a function called before each render pass. Useful for polling, animations, or other per-frame logic.

### Hit Testing

```go
func (e *Element) ElementAt(x, y int) *Element
func (e *Element) ElementAtPoint(x, y int) Focusable
```

`ElementAt` finds the deepest element containing the given point. Children are checked in reverse order (last child renders on top, so it gets priority). Returns `nil` if no element contains the point.

`ElementAtPoint` does the same thing but returns a `Focusable` interface to satisfy the internal mouse hit-testing contract.

## Rendering

```go
func (e *Element) Render(buf *Buffer, width, height int)
```

The main rendering entry point. Runs layout (if dirty) and then renders the full element tree to the buffer.

```go
func RenderTree(buf *Buffer, root *Element)
```

Package-level function that traverses the element tree and draws each element to the buffer. Handles background fills, borders, text, gradients, scroll clipping, and overflow clipping. Called by `Element.Render()` after layout.

## Enums

### TextAlign

Controls horizontal text alignment within an element's content area.

| Constant | Description |
|----------|-------------|
| `TextAlignLeft` | Left-aligned (default) |
| `TextAlignCenter` | Centered horizontally |
| `TextAlignRight` | Right-aligned |

Text alignment only takes effect when the element is wider than its text content, e.g. when you set an explicit `WithWidth` larger than the text. For auto-sized elements, the text fills the element exactly, and the parent's `AlignItems` handles positioning.

### ScrollMode

Controls which directions an element can scroll.

| Constant | Description |
|----------|-------------|
| `ScrollNone` | Scrolling disabled (default) |
| `ScrollVertical` | Vertical scrolling only |
| `ScrollHorizontal` | Horizontal scrolling only |
| `ScrollBoth` | Both vertical and horizontal scrolling |

### OverflowMode

Controls how content beyond element bounds is handled.

| Constant | Description |
|----------|-------------|
| `OverflowVisible` | Content renders outside bounds (default) |
| `OverflowHidden` | Content is clipped at element bounds |
