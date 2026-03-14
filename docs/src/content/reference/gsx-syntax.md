# GSX Syntax Reference

## Overview

`.gsx` files are Go source files extended with a template syntax for declaring UIs. The `tui generate` command reads `.gsx` files and produces `_gsx.go` files containing standard Go code. You never edit the generated files by hand.

A `.gsx` file can contain:

- A `package` declaration and `import` block (standard Go)
- Type declarations and regular Go functions (`type`, `func`)
- Pure template components (`templ Name(params) { ... }`)
- Struct method components (`templ (s *Struct) Render() { ... }`)
- Control flow directives (`if`, `for`, `:=`)
- HTML-like elements (`<div>`, `<span>`, etc.)

## File structure

Every `.gsx` file starts with a standard Go package and import block:

```gsx
package mypackage

import (
    "fmt"

    tui "github.com/grindlemire/go-tui"
)
```

The rest of the file can mix Go declarations (types, functions, variables) with `templ` component definitions in any order.

## Component syntax

### Pure components

A pure component is a stateless function that takes parameters and returns UI elements. Define one with the `templ` keyword:

```gsx
templ Greeting(name string) {
    <span class="text-cyan">{"Hello, " + name}</span>
}
```

Parameters use standard Go function parameter syntax. Any valid Go type works:

```gsx
templ UserList(users []string, maxVisible int) {
    <div class="flex-col">
        for i, u := range users {
            if i < maxVisible {
                <span>{u}</span>
            }
        }
    </div>
}
```

### Children slot

Pure components can accept nested content from their caller using `{children...}`:

```gsx
templ Card(title string) {
    <div class="border-rounded p-1">
        <span class="font-bold">{title}</span>
        {children...}
    </div>
}
```

Call it with children:

```gsx
<Card title="Status">
    <span class="text-green">All systems operational</span>
</Card>
```

`{children...}` is only available in pure `templ` components, not in struct method components.

### Struct method components

A struct component has its own state and lifecycle. Define the render method with `templ` using a receiver:

```gsx
type counter struct {
    count *tui.State[int]
}

func Counter() *counter {
    return &counter{
        count: tui.NewState(0),
    }
}

templ (c *counter) Render() {
    <div class="flex gap-2">
        <span class="font-bold">{fmt.Sprintf("Count: %d", c.count.Get())}</span>
    </div>
}
```

The method name must be `Render` and it takes no parameters. The receiver can be a pointer or value type, though pointer receivers are standard.

### Calling components

Call a pure component like an element, passing parameters as attributes:

```gsx
<Greeting name="World" />
<Card title="Info">
    <span>Some content</span>
</Card>
```

Component names must start with an uppercase letter to distinguish them from built-in elements.

### Regular Go functions

Standard Go functions work normally in `.gsx` files:

```gsx
func formatPercent(v, max int) string {
    if max == 0 {
        return "0%"
    }
    return fmt.Sprintf("%d%%", v*100/max)
}
```

These are passed through to the generated file without transformation.

## Elements

### Container elements

These elements can have children and support flexbox layout attributes.

**`<div>`** -- Block container, and the main layout element. Default direction is row.

```gsx
<div class="flex-col gap-1 p-1 border-rounded">
    <span>First</span>
    <span>Second</span>
</div>
```

**`<ul>`** -- Unordered list container. Use with `<li>` children.

```gsx
<ul class="flex-col">
    <li>Item one</li>
    <li>Item two</li>
</ul>
```

**`<li>`** -- List item. Renders with a bullet prefix. Should be a child of `<ul>`.

**`<table>`** -- Table container for tabular data. Supports all container attributes.

### Text elements

These hold text content and support text styling but not flex container attributes like `direction` or `justify`.

**`<span>`** -- Inline text container for styled text content.

```gsx
<span class="font-bold text-cyan">Status: OK</span>
```

**`<p>`** -- Paragraph element for text blocks.

### Interactive elements

**`<button>`** -- Clickable element. Supports text and event attributes. Attach a ref for click handling:

```gsx
<button ref={s.myBtn} class="px-2 border-rounded text-green">{" Save "}</button>
```

**`<input />`** -- Single-line text input. Self-closing. Bind `value` to a `*State[string]` for two-way binding. Also accepts `placeholder`, `width`, `border`, `focusColor`, `borderGradient`, `focusGradient`, `onSubmit`, and `onChange`.

```gsx
<input value={s.text} placeholder="Type here..." width={30} border={tui.BorderRounded} />
```

**`<textarea />`** -- Multi-line text input with word wrapping. Self-closing. Bind `value` to a `*State[string]` for two-way binding. Also accepts `placeholder`, `width`, `maxHeight`, `border`, `focusColor`, `borderGradient`, `focusGradient`, `submitKey`, and `onSubmit`.

```gsx
<textarea value={s.note} placeholder="Write here..." width={40} maxHeight={6} border={tui.BorderRounded} />
```

### Display elements

**`<progress />`** -- Progress bar. Self-closing. Accepts `value`, `max`, and `width`.

```gsx
<progress value={75} max={100} width={20} />
```

**`<hr />`** -- Horizontal rule. Self-closing. Accepts only `id` and `class`.

```gsx
<hr />
```

**`<br />`** -- Line break. Self-closing. Accepts only `id` and `class`.

```gsx
<br />
```

### Self-closing elements

`<input />`, `<textarea />`, `<progress />`, `<hr />`, and `<br />` are self-closing and cannot have children. Writing `<input>children</input>` produces a compile error.

## Attributes

### String attributes

Pass string values with double quotes:

```gsx
<div class="flex-col gap-1" id="header">
    <span text="Hello" />
</div>
```

### Integer and float attributes

Pass numbers directly without braces:

```gsx
<div width={40} height={10} gap={2}>
    <div flexGrow={1.5} />
</div>
```

### Boolean attributes

Pass `true` or `false`, or use the shorthand (attribute name alone means `true`):

```gsx
<div focusable={true} />
<div focusable />          // equivalent
<div disabled={false} />
```

### Go expression attributes

Pass any Go expression inside braces:

```gsx
<div
    width={computeWidth()}
    textStyle={tui.NewStyle().Bold().Foreground(tui.ANSIColor(tui.Cyan))}
    scrollable={tui.ScrollVertical}
    scrollOffset={0, s.scrollY.Get()}
>
```

### Ref attributes

Bind an element to a ref for later access in handlers:

```gsx
<div ref={s.myRef} class="border-single p-1" />
```

See the [Refs Reference](refs.md) for `Ref`, `RefList`, and `RefMap` details.

### Key attributes

Inside `for` loops, the `key` attribute tells the framework how to identify elements for `RefMap`:

```gsx
for _, name := range items {
    <div ref={s.itemRefs} key={name}>{name}</div>
}
```

## Attribute reference

### Generic attributes (all elements)

| Attribute | Type | Generated option | Description |
|-----------|------|------------------|-------------|
| `id` | string | -- | Unique identifier |
| `class` | string | (varies) | Tailwind-style classes |
| `disabled` | bool | -- | Disables the element |
| `ref` | expression | `ref.Set(el)` | Binds element to a ref |
| `deps` | expression | -- | Explicit state dependencies |
| `key` | expression | -- | Key for RefMap in loops |

### Layout attributes

| Attribute | Type | Generated option |
|-----------|------|------------------|
| `width` | int | `tui.WithWidth(n)` |
| `widthPercent` | int | `tui.WithWidthPercent(n)` |
| `height` | int | `tui.WithHeight(n)` |
| `heightPercent` | int | `tui.WithHeightPercent(n)` |
| `minWidth` | int | `tui.WithMinWidth(n)` |
| `minHeight` | int | `tui.WithMinHeight(n)` |
| `maxWidth` | int | `tui.WithMaxWidth(n)` |
| `maxHeight` | int | `tui.WithMaxHeight(n)` |

Available on: `div`, `ul`, `li`, `table`, `span`, `p`, `button`, `input` (width only), `progress` (width only).

### Flex attributes

| Attribute | Type | Generated option |
|-----------|------|------------------|
| `direction` | `tui.Direction` | `tui.WithDirection(d)` |
| `justify` | `tui.Justify` | `tui.WithJustify(j)` |
| `align` | `tui.Align` | `tui.WithAlign(a)` |
| `gap` | int | `tui.WithGap(n)` |
| `flexGrow` | float | `tui.WithFlexGrow(f)` |
| `flexShrink` | float | `tui.WithFlexShrink(f)` |
| `alignSelf` | `tui.Align` | `tui.WithAlignSelf(a)` |

Available on: `div`, `ul`, `li`, `table`.

### Spacing attributes

| Attribute | Type | Generated option |
|-----------|------|------------------|
| `padding` | int | `tui.WithPadding(n)` |
| `margin` | int | `tui.WithMargin(n)` |

Available on: `div`, `ul`, `li`, `table`, `span`, `p`, `button`.

### Visual attributes

| Attribute | Type | Generated option |
|-----------|------|------------------|
| `border` | `tui.BorderStyle` | `tui.WithBorder(b)` |
| `borderStyle` | `tui.Style` | `tui.WithBorderStyle(s)` |
| `background` | `tui.Style` | `tui.WithBackground(s)` |

Available on: `div`, `ul`, `li`, `table`, `span`, `p`, `button`.

### Text attributes

| Attribute | Type | Generated option |
|-----------|------|------------------|
| `text` | string | `tui.WithText(s)` |
| `textStyle` | `tui.Style` | `tui.WithTextStyle(s)` |
| `textAlign` | `tui.TextAlign` | `tui.WithTextAlign(a)` |

Available on: `span`, `p`, `button`.

### Event and focus attributes

| Attribute | Type | Generated option |
|-----------|------|------------------|
| `focusable` | bool | `tui.WithFocusable(b)` |
| `onFocus` | `func(*tui.Element)` | `tui.WithOnFocus(fn)` |
| `onBlur` | `func(*tui.Element)` | `tui.WithOnBlur(fn)` |

Available on: `div`, `ul`, `li`, `table`, `span`, `p`, `button`, `input`.

### Scroll attributes

| Attribute | Type | Generated option |
|-----------|------|------------------|
| `scrollable` | `tui.ScrollMode` | `tui.WithScrollable(m)` |
| `scrollOffset` | int, int | `tui.WithScrollOffset(x, y)` |
| `scrollbarStyle` | `tui.Style` | `tui.WithScrollbarStyle(s)` |
| `scrollbarThumbStyle` | `tui.Style` | `tui.WithScrollbarThumbStyle(s)` |

Available on: `div`, `ul`, `li`, `table`.

### Input-specific attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `value` | `*tui.State[string]` | Two-way text binding |
| `placeholder` | string | Placeholder text when empty |
| `width` | int | Input width in characters (default 20) |
| `border` | `tui.BorderStyle` | Border style |
| `textStyle` | `tui.Style` | Text styling |
| `placeholderStyle` | `tui.Style` | Placeholder text styling (default: dim) |
| `cursor` | rune | Cursor character (default '▌') |
| `focusColor` | `tui.Color` | Border color when focused (default Cyan) |
| `borderGradient` | `tui.Gradient` | Border gradient when unfocused |
| `focusGradient` | `tui.Gradient` | Border gradient when focused (overrides focusColor) |
| `onSubmit` | `func(string)` | Called when Enter is pressed |
| `onChange` | `func(string)` | Called when text changes |

### Textarea-specific attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `value` | `*tui.State[string]` | Two-way text binding |
| `placeholder` | string | Placeholder text when empty |
| `width` | int | Width in characters (default 40) |
| `maxHeight` | int | Maximum height in rows (0 = unlimited) |
| `border` | `tui.BorderStyle` | Border style |
| `textStyle` | `tui.Style` | Text styling |
| `placeholderStyle` | `tui.Style` | Placeholder text styling (default: dim) |
| `cursor` | rune | Cursor character (default '▌') |
| `focusColor` | `tui.Color` | Border color when focused (default Cyan) |
| `borderGradient` | `tui.Gradient` | Border gradient when unfocused |
| `focusGradient` | `tui.Gradient` | Border gradient when focused (overrides focusColor) |
| `submitKey` | `tui.Key` | Key that triggers submit (default KeyEnter) |
| `onSubmit` | `func(string)` | Called when submit key is pressed |

### Progress-specific attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `value` | int | Current progress (0 to max) |
| `max` | int | Maximum progress value |

## Go expressions

### Text content

Embed Go expressions inside braces to produce text:

```gsx
<span>{fmt.Sprintf("Count: %d", count)}</span>
<span>{"literal string"}</span>
<span>{s.name.Get()}</span>
```

### Attribute values

Use braces for dynamic attribute values:

```gsx
<div width={s.computeWidth()} textStyle={s.getActiveStyle()} />
```

### Component calls

Call components with the `@` prefix or as XML-like tags:

```gsx
<Card title={fmt.Sprintf("Item %d", i)}>
    <span>{content}</span>
</Card>
```

## Control flow

### if / else

Conditionally render elements:

```gsx
if s.loading.Get() {
    <span class="text-yellow">Loading...</span>
} else {
    <span class="text-green">Ready</span>
}
```

Chain conditions:

```gsx
if count > 10 {
    <span class="text-red">High</span>
} else if count > 5 {
    <span class="text-yellow">Medium</span>
} else {
    <span class="text-green">Low</span>
}
```

The condition is any valid Go boolean expression.

### for

Loop over collections:

```gsx
for i, item := range s.items.Get() {
    <span>{fmt.Sprintf("%d. %s", i+1, item)}</span>
}
```

Supports all standard Go range patterns:

```gsx
for _, v := range items {       // index ignored
for i := range items {           // value ignored
for i, v := range items {       // both used
```

### Local bindings (:=)

Bind an element to a local variable for reuse:

```gsx
badge := <span class="text-cyan font-bold">{fmt.Sprintf("%d", s.count.Get())}</span>
<div class="flex gap-2">
    {badge}
</div>
```

Note the `:=` binding assigns both element expressions (starting with `<`) to a local variable as well as normal Go expressions.

## Tailwind class reference

Classes are set via the `class` attribute. Multiple classes are space-separated. The compiler maps each class to one or more `tui.With*` option functions.

### Layout

| Class | Generated option |
|-------|------------------|
| `flex` | `tui.WithDirection(tui.Row)` |
| `flex-row` | `tui.WithDirection(tui.Row)` |
| `flex-col` | `tui.WithDirection(tui.Column)` |
| `gap-N` | `tui.WithGap(N)` |

### Flex sizing

| Class | Generated option |
|-------|------------------|
| `grow` | `tui.WithFlexGrow(1)` |
| `grow-0` | `tui.WithFlexGrow(0)` |
| `shrink` | `tui.WithFlexShrink(1)` |
| `shrink-0` | `tui.WithFlexShrink(0)` |
| `flex-1` | `tui.WithFlexGrow(1)`, `tui.WithFlexShrink(1)` |
| `flex-auto` | `tui.WithFlexGrow(1)`, `tui.WithFlexShrink(1)` |
| `flex-initial` | `tui.WithFlexGrow(0)`, `tui.WithFlexShrink(1)` |
| `flex-none` | `tui.WithFlexGrow(0)`, `tui.WithFlexShrink(0)` |
| `flex-grow-N` | `tui.WithFlexGrow(N)` |
| `flex-shrink-N` | `tui.WithFlexShrink(N)` |

### Justify content

| Class | Generated option |
|-------|------------------|
| `justify-start` | `tui.WithJustify(tui.JustifyStart)` |
| `justify-center` | `tui.WithJustify(tui.JustifyCenter)` |
| `justify-end` | `tui.WithJustify(tui.JustifyEnd)` |
| `justify-between` | `tui.WithJustify(tui.JustifySpaceBetween)` |
| `justify-around` | `tui.WithJustify(tui.JustifySpaceAround)` |
| `justify-evenly` | `tui.WithJustify(tui.JustifySpaceEvenly)` |

### Align items

| Class | Generated option |
|-------|------------------|
| `items-start` | `tui.WithAlign(tui.AlignStart)` |
| `items-center` | `tui.WithAlign(tui.AlignCenter)` |
| `items-end` | `tui.WithAlign(tui.AlignEnd)` |
| `items-stretch` | `tui.WithAlign(tui.AlignStretch)` |

### Self-alignment

| Class | Generated option |
|-------|------------------|
| `self-start` | `tui.WithAlignSelf(tui.AlignStart)` |
| `self-center` | `tui.WithAlignSelf(tui.AlignCenter)` |
| `self-end` | `tui.WithAlignSelf(tui.AlignEnd)` |
| `self-stretch` | `tui.WithAlignSelf(tui.AlignStretch)` |

### Text alignment

| Class | Generated option |
|-------|------------------|
| `text-left` | `tui.WithTextAlign(tui.TextAlignLeft)` |
| `text-center` | `tui.WithTextAlign(tui.TextAlignCenter)` |
| `text-right` | `tui.WithTextAlign(tui.TextAlignRight)` |

### Width and height

| Class | Generated option |
|-------|------------------|
| `w-N` | `tui.WithWidth(N)` |
| `w-full` | `tui.WithWidthPercent(100.00)` |
| `w-auto` | `tui.WithWidthAuto()` |
| `w-1/2` | `tui.WithWidthPercent(50.00)` |
| `w-1/3` | `tui.WithWidthPercent(33.33)` |
| `w-2/3` | `tui.WithWidthPercent(66.67)` |
| `h-N` | `tui.WithHeight(N)` |
| `h-full` | `tui.WithHeightPercent(100.00)` |
| `h-auto` | `tui.WithHeightAuto()` |
| `min-w-N` | `tui.WithMinWidth(N)` |
| `max-w-N` | `tui.WithMaxWidth(N)` |
| `min-h-N` | `tui.WithMinHeight(N)` |
| `max-h-N` | `tui.WithMaxHeight(N)` |

Fraction syntax (`w-N/D`) computes the percentage at compile time.

### Spacing

| Class | Generated option |
|-------|------------------|
| `p-N` | `tui.WithPadding(N)` |
| `px-N` | `tui.WithPaddingTRBL(0, N, 0, N)` |
| `py-N` | `tui.WithPaddingTRBL(N, 0, N, 0)` |
| `pt-N` | `tui.WithPaddingTRBL(N, 0, 0, 0)` |
| `pr-N` | `tui.WithPaddingTRBL(0, N, 0, 0)` |
| `pb-N` | `tui.WithPaddingTRBL(0, 0, N, 0)` |
| `pl-N` | `tui.WithPaddingTRBL(0, 0, 0, N)` |
| `m-N` | `tui.WithMargin(N)` |
| `mx-N` | `tui.WithMarginTRBL(0, N, 0, N)` |
| `my-N` | `tui.WithMarginTRBL(N, 0, N, 0)` |
| `mt-N` | `tui.WithMarginTRBL(N, 0, 0, 0)` |
| `mr-N` | `tui.WithMarginTRBL(0, N, 0, 0)` |
| `mb-N` | `tui.WithMarginTRBL(0, 0, N, 0)` |
| `ml-N` | `tui.WithMarginTRBL(0, 0, 0, N)` |

### Borders

| Class | Generated option |
|-------|------------------|
| `border` | `tui.WithBorder(tui.BorderSingle)` |
| `border-single` | `tui.WithBorder(tui.BorderSingle)` |
| `border-double` | `tui.WithBorder(tui.BorderDouble)` |
| `border-rounded` | `tui.WithBorder(tui.BorderRounded)` |
| `border-thick` | `tui.WithBorder(tui.BorderThick)` |

Border colors: `border-red`, `border-green`, `border-blue`, `border-cyan`, `border-magenta`, `border-yellow`, `border-white`, `border-black`. Each generates `tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Color))`.

### Text styles

These classes accumulate and combine into a single `tui.WithTextStyle(tui.NewStyle().Method1().Method2()...)` call.

| Class | Style method |
|-------|-------------|
| `font-bold` | `.Bold()` |
| `font-dim` | `.Dim()` |
| `text-dim` | `.Dim()` |
| `italic` | `.Italic()` |
| `underline` | `.Underline()` |
| `strikethrough` | `.Strikethrough()` |
| `reverse` | `.Reverse()` |
| `blink` | `.Blink()` |

### Text colors

Standard: `text-red`, `text-green`, `text-blue`, `text-cyan`, `text-magenta`, `text-yellow`, `text-white`, `text-black`.

Bright: `text-bright-red`, `text-bright-green`, `text-bright-blue`, `text-bright-cyan`, `text-bright-magenta`, `text-bright-yellow`, `text-bright-white`, `text-bright-black`.

Each adds `.Foreground(tui.Color)` to the text style.

### Background colors

Standard: `bg-red`, `bg-green`, `bg-blue`, `bg-cyan`, `bg-magenta`, `bg-yellow`, `bg-white`, `bg-black`.

Bright: `bg-bright-red`, `bg-bright-green`, `bg-bright-blue`, `bg-bright-cyan`, `bg-bright-magenta`, `bg-bright-yellow`, `bg-bright-white`, `bg-bright-black`.

Each generates `tui.WithBackground(tui.NewStyle().Background(tui.Color))`.

### Arbitrary hex colors

Use bracket syntax for hex color values:

| Class | Description |
|-------|-------------|
| `text-[#RGB]` or `text-[#RRGGBB]` | Text color from hex |
| `bg-[#RGB]` or `bg-[#RRGGBB]` | Background from hex |
| `border-[#RGB]` or `border-[#RRGGBB]` | Border color from hex |

### Gradients

Apply color gradients to text, backgrounds, or borders.

Syntax: `{target}-gradient-{start}-{end}[-direction]`

**Targets:** `text-gradient`, `bg-gradient`, `border-gradient`

**Directions:**

- `-h` -- horizontal (default)
- `-v` -- vertical
- `-dd` -- diagonal down
- `-du` -- diagonal up

Examples:

| Class | Generated option |
|-------|------------------|
| `text-gradient-red-blue` | `tui.WithTextGradient(tui.NewGradient(tui.ANSIColor(tui.Red), tui.ANSIColor(tui.Blue)))` |
| `text-gradient-cyan-magenta-v` | Same with `.WithDirection(tui.GradientVertical)` |
| `bg-gradient-green-yellow-dd` | `tui.WithBackgroundGradient(...)` with diagonal down |
| `border-gradient-red-blue` | `tui.WithBorderGradient(...)` |

### Scroll

| Class | Generated option |
|-------|------------------|
| `overflow-scroll` | `tui.WithScrollable(tui.ScrollBoth)` |
| `overflow-y-scroll` | `tui.WithScrollable(tui.ScrollVertical)` |
| `overflow-x-scroll` | `tui.WithScrollable(tui.ScrollHorizontal)` |
| `overflow-hidden` | `tui.WithOverflow(tui.OverflowHidden)` |

### Scrollbar colors

Track: `scrollbar-red`, `scrollbar-green`, etc. (all 16 standard and bright colors).

Thumb: `scrollbar-thumb-red`, `scrollbar-thumb-green`, etc. (all 16 standard and bright colors).

### Visibility and behavior

| Class | Generated option |
|-------|------------------|
| `focusable` | `tui.WithFocusable(true)` |
| `hidden` | `tui.WithHidden(true)` |
| `truncate` | `tui.WithTruncate(true)` |

## Code generation

### Commands

```bash
tui generate [path...]    # generate _gsx.go from .gsx files
tui check [path...]       # validate .gsx without generating
tui fmt [path...]         # format .gsx files
tui fmt --check [path...] # check formatting without modifying
```

### How generation works

1. The compiler lexes and parses each `.gsx` file into an AST.
2. The analyzer validates element tags, attributes, ref usage, and Tailwind classes.
3. The generator produces a `_gsx.go` file in the same directory with the same package name.
4. Each `templ` block becomes a Go function or method returning `*tui.Element`.
5. Elements become calls to `tui.New(options...)` with `AddChild` calls for children.
6. Tailwind classes become element option arguments at compile time (not at runtime).
7. Control flow (`if`, `for`, `:=`) becomes standard Go control flow.

Re-run `tui generate` after any `.gsx` change. The generated `_gsx.go` files should be committed to version control but never edited by hand.
