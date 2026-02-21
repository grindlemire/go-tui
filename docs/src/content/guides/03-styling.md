# Styling and Colors

## Overview

go-tui uses a Tailwind-inspired class system for styling. You set colors, text decorations, borders, and gradients through the `class` attribute on elements, and the code generator converts them into `tui.Element` options at compile time. For cases where classes aren't flexible enough, you can construct `Style` values in Go and pass them through attributes.

## Text Styles

Text decorations are applied through classes on any element that renders text:

```gsx
<div class="flex-col gap-1">
    <span class="font-bold">Bold text</span>
    <span class="font-dim">Dimmed text</span>
    <span class="italic">Italic text</span>
    <span class="underline">Underlined text</span>
    <span class="strikethrough">Struck-through text</span>
    <span class="reverse">Reversed foreground/background</span>
</div>
```

The full set of text style classes:

| Class | Effect |
|-------|--------|
| `font-bold` | Bold / bright |
| `font-dim` or `text-dim` | Dimmed / faint |
| `italic` | Italic |
| `underline` | Underlined |
| `strikethrough` | Struck through |
| `reverse` | Swaps foreground and background colors |

`font-dim` and `text-dim` are aliases — they produce the same result.

## Text Colors

go-tui supports the standard 16 ANSI terminal colors. Each has a `text-` prefixed class:

```gsx
<div class="flex-col">
    <span class="text-red">Red</span>
    <span class="text-green">Green</span>
    <span class="text-blue">Blue</span>
    <span class="text-cyan">Cyan</span>
    <span class="text-magenta">Magenta</span>
    <span class="text-yellow">Yellow</span>
    <span class="text-white">White</span>
    <span class="text-black">Black</span>
</div>
```

Each color has a bright variant:

```gsx
<div class="flex-col">
    <span class="text-bright-red">Bright red</span>
    <span class="text-bright-green">Bright green</span>
    <span class="text-bright-blue">Bright blue</span>
    <span class="text-bright-cyan">Bright cyan</span>
    <span class="text-bright-magenta">Bright magenta</span>
    <span class="text-bright-yellow">Bright yellow</span>
    <span class="text-bright-white">Bright white</span>
    <span class="text-bright-black">Bright black (gray)</span>
</div>
```

### Arbitrary Hex Colors

For colors beyond the 16 ANSI palette, use the bracket syntax with a hex value:

```gsx
<span class="text-[#FF6B35]">Custom orange</span>
<span class="text-[#A3F]">Short-form hex (expands to #AA33FF)</span>
```

Both `#RRGGBB` (6-digit) and `#RGB` (3-digit shorthand) formats work. These produce true-color output if the terminal supports it, with automatic fallback to the nearest ANSI color on limited terminals.

## Background Colors

The same color set is available with the `bg-` prefix:

```gsx
<div class="flex-col gap-1">
    <span class="bg-red text-white p-1">Red background</span>
    <span class="bg-cyan text-black p-1">Cyan background</span>
    <span class="bg-bright-yellow text-black p-1">Bright yellow background</span>
    <span class="bg-[#2D1B69] text-white p-1">Custom purple background</span>
</div>
```

Every named color and bright variant that works with `text-` also works with `bg-`. Arbitrary hex values work the same way: `bg-[#RRGGBB]` or `bg-[#RGB]`.

## Combining Styles

Multiple classes compose. List them space-separated in the `class` attribute:

```gsx
<span class="font-bold text-cyan bg-black underline">
    Bold, cyan, on black, underlined
</span>
```

Styles on parent elements do not cascade to children. Each element's `class` applies only to that element:

```gsx
<div class="text-cyan">
    <span>This text is NOT cyan — it uses the default color</span>
    <span class="text-cyan">This text IS cyan</span>
</div>
```

If you want consistent styling across a subtree, apply the class to each element that needs it, or use a pure component to encapsulate the pattern.

## Borders

Borders wrap an element in box-drawing characters. Four styles are available:

| Class | Style | Characters |
|-------|-------|------------|
| `border-single` | Single line | `┌─┐│└─┘` |
| `border-double` | Double line | `╔═╗║╚═╝` |
| `border-rounded` | Rounded corners | `╭─╮│╰─╯` |
| `border-thick` | Heavy line | `┏━┓┃┗━┛` |

```gsx
<div class="flex gap-2">
    <div class="border-single p-1">
        <span>Single</span>
    </div>
    <div class="border-double p-1">
        <span>Double</span>
    </div>
    <div class="border-rounded p-1">
        <span>Rounded</span>
    </div>
    <div class="border-thick p-1">
        <span>Thick</span>
    </div>
</div>
```

### Border Colors

Color the border with `border-{color}`:

```gsx
<div class="border-rounded border-cyan p-1">
    <span>Cyan bordered box</span>
</div>
```

All 16 named colors and their bright variants work: `border-red`, `border-bright-green`, etc. Arbitrary hex values are supported too: `border-[#FF6B35]`.

## Gradients

Gradients interpolate between two colors across text, backgrounds, or borders.

### Syntax

The general form is:

```
{target}-gradient-{start}-{end}[-{direction}]
```

Where:
- **target**: `text`, `bg`, or `border`
- **start** and **end**: any named ANSI color (`red`, `cyan`, `bright-blue`, etc.)
- **direction** (optional): `-h` (horizontal, the default), `-v` (vertical), `-dd` (diagonal down), `-du` (diagonal up)

### Text Gradients

```gsx
<div class="flex-col gap-1">
    <span class="text-gradient-cyan-magenta">Horizontal gradient (default)</span>
    <span class="text-gradient-red-yellow-h">Explicit horizontal</span>
    <span class="text-gradient-green-blue-v">Vertical gradient</span>
    <span class="text-gradient-cyan-magenta-dd">Diagonal down</span>
    <span class="text-gradient-yellow-red-du">Diagonal up</span>
</div>
```

### Background Gradients

```gsx
<div class="bg-gradient-blue-cyan-h p-1">
    <span class="text-white">Blue-to-cyan background</span>
</div>
```

### Border Gradients

```gsx
<div class="border-rounded border-gradient-cyan-magenta p-1">
    <span>Gradient border</span>
</div>
```

The gradient travels around the perimeter of the border, shifting from the start color to the end color over the first half, then back again over the second half.

### Direction Reference

| Suffix | Direction | Description |
|--------|-----------|-------------|
| `-h` (or omitted) | Horizontal | Left to right |
| `-v` | Vertical | Top to bottom |
| `-dd` | Diagonal down | Top-left to bottom-right |
| `-du` | Diagonal up | Bottom-left to top-right |

## Programmatic Styling

When you need dynamic styles that depend on state or computed values, construct `Style` objects in Go and pass them through element attributes.

### Building Styles

Use `tui.NewStyle()` with chainable methods:

```go
highlight := tui.NewStyle().
    Bold().
    Foreground(tui.Cyan).
    Background(tui.Black)
```

Available methods on `Style`:

| Method | Effect |
|--------|--------|
| `Foreground(Color)` | Set text color |
| `Background(Color)` | Set background color |
| `Bold()` | Bold text |
| `Dim()` | Dimmed text |
| `Italic()` | Italic text |
| `Underline()` | Underlined text |
| `Strikethrough()` | Struck-through text |
| `Reverse()` | Swap foreground and background |
| `Blink()` | Blinking text (rarely supported by terminals) |

All methods return a new `Style`, so they chain:

```go
style := tui.NewStyle().Bold().Underline().Foreground(tui.Red)
```

### Color Constructors

Several ways to create a `Color`:

```go
// Named color variables (already Color values)
tui.Cyan
tui.BrightRed

// ANSI 256 palette index (0-255)
tui.ANSIColor(33)

// 24-bit RGB
tui.RGBColor(255, 107, 53)

// Hex string (returns Color and error)
color, err := tui.HexColor("#FF6B35")
```

The named color variables (`tui.Red`, `tui.Cyan`, `tui.BrightGreen`, etc.) are already `Color` values, so you can use them directly with `Foreground()` and `Background()`:

```go
style := tui.NewStyle().Foreground(tui.Cyan).Bold()
```

### Applying Styles to Elements

Pass styles through element attributes in `.gsx`:

```gsx
templ (s *myApp) Render() {
    <div borderStyle={tui.NewStyle().Foreground(tui.Magenta)} class="border-rounded p-1">
        <span textStyle={tui.NewStyle().Bold().Foreground(tui.Cyan)}>Dynamic style</span>
    </div>
}
```

Style attributes you can set:

| Attribute | Controls |
|-----------|----------|
| `textStyle` | Text foreground color, background, and decorations |
| `borderStyle` | Border line color and decorations |
| `background` | Element background fill |

### When to Use Programmatic Styles

Class-based styling handles most cases. Programmatic styles are useful when the style depends on runtime values, like coloring a number red when negative and green when positive, or when you need 256-palette indices or RGB colors computed at runtime.

```gsx
package main

import (
    "fmt"

    tui "github.com/grindlemire/go-tui"
)

type statusApp struct {
    value *tui.State[int]
}

func StatusApp() *statusApp {
    return &statusApp{
        value: tui.NewState(0),
    }
}

func (s *statusApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('+', func(ke tui.KeyEvent) {
            s.value.Update(func(v int) int { return v + 1 })
        }),
        tui.OnRune('-', func(ke tui.KeyEvent) {
            s.value.Update(func(v int) int { return v - 1 })
        }),
    }
}

func valueStyle(v int) tui.Style {
    if v > 0 {
        return tui.NewStyle().Bold().Foreground(tui.Green)
    }
    if v < 0 {
        return tui.NewStyle().Bold().Foreground(tui.Red)
    }
    return tui.NewStyle().Dim()
}

templ (s *statusApp) Render() {
    <div class="flex-col items-center justify-center h-full gap-1">
        <span textStyle={valueStyle(s.value.Get())}>{fmt.Sprintf("Value: %d", s.value.Get())}</span>
        <span class="font-dim">Press + / - to change, Esc to quit</span>
    </div>
}
```

## Color Capabilities

Terminals vary in color support. go-tui detects the current terminal's capabilities at startup and falls back to supported colors automatically.

### Detection

`tui.DetectCapabilities()` checks environment variables (`COLORTERM`, `TERM`, and terminal-specific variables like `ITERM_SESSION_ID` or `KITTY_WINDOW_ID`) to determine what the terminal supports:

```go
caps := tui.DetectCapabilities()
fmt.Println(caps.Colors)    // Color16, Color256, or ColorTrue
fmt.Println(caps.TrueColor) // true if 24-bit RGB is supported
fmt.Println(caps.Unicode)   // true if Unicode rendering is supported
```

### Color Levels

| Level | Constant | Colors |
|-------|----------|--------|
| None | `tui.ColorNone` | Monochrome |
| 16 | `tui.Color16` | Standard + bright ANSI |
| 256 | `tui.Color256` | ANSI 256 palette |
| True color | `tui.ColorTrue` | 24-bit RGB |

### Automatic Fallback

You don't need to handle color fallback yourself. When you use an RGB color (via `tui.RGBColor()` or hex classes like `text-[#FF6B35]`) on a terminal that only supports 256 or 16 colors, the framework approximates it to the nearest supported color using the `Color.ToANSI()` method. This maps RGB values to the closest entry in the ANSI 256 palette's 6x6x6 color cube and grayscale ramp.

If even ANSI colors aren't supported, colors are dropped entirely and the terminal's default foreground and background are used.

## Next Steps

- [Layout](layout) -- Flexbox layout: direction, alignment, spacing, and sizing
- [State and Reactivity](state) -- Reactive state with `State[T]`
