# Styling Reference

## Overview

go-tui's visual styling system covers text decoration, colors, gradients, and borders. You can apply styles through Tailwind-like CSS classes in `.gsx` files or programmatically with Go types. Both approaches produce the same result. Classes compile to the same `Style`, `Color`, and `Gradient` types described here.

```go
import tui "github.com/grindlemire/go-tui"
```

## Style

`Style` controls the visual appearance of text and element decorations. It holds a foreground color, background color, and a set of attribute flags. Style is a value type, so all methods return a new `Style` rather than modifying the receiver.

### Creating a Style

```go
func NewStyle() Style
```

Returns a zero-value `Style` with no colors and no attributes set. Build up the style using chainable methods:

```go
s := tui.NewStyle().Bold().Foreground(tui.ANSIColor(tui.Cyan))
```

### Style Fields

| Field | Type | Description |
|-------|------|-------------|
| `Fg` | `Color` | Foreground (text) color |
| `Bg` | `Color` | Background color |
| `Attrs` | `Attr` | Bitfield of text attributes |

### Chainable Methods

Each method returns a new `Style` with the specified property applied:

```go
func (s Style) Foreground(c Color) Style
func (s Style) Background(c Color) Style
func (s Style) Bold() Style
func (s Style) Dim() Style
func (s Style) Italic() Style
func (s Style) Underline() Style
func (s Style) Blink() Style
func (s Style) Reverse() Style
func (s Style) Strikethrough() Style
```

Chain multiple methods to combine properties:

```go
highlight := tui.NewStyle().
    Foreground(tui.ANSIColor(tui.Yellow)).
    Background(tui.ANSIColor(tui.Blue)).
    Bold().
    Underline()
```

### Query Methods

```go
func (s Style) Equal(other Style) bool
func (s Style) HasAttr(a Attr) bool
```

`Equal` compares two styles for identical foreground, background, and attributes. `HasAttr` checks whether a specific attribute flag is set:

```go
if style.HasAttr(tui.AttrBold) {
    // style includes bold
}
```

### Applying Styles to Elements

In `.gsx` files, use the `textStyle`, `borderStyle`, and `background` attributes:

```gsx
<span textStyle={tui.NewStyle().Bold().Foreground(tui.ANSIColor(tui.Red))}>
    Error message
</span>
```

Or use the equivalent `Option` functions when building elements in Go:

| Function | Signature | Description |
|----------|-----------|-------------|
| `WithTextStyle` | `func WithTextStyle(style Style) Option` | Sets text style |
| `WithBorderStyle` | `func WithBorderStyle(style Style) Option` | Sets border style (color, attributes) |
| `WithBackground` | `func WithBackground(style Style) Option` | Sets background fill style |
| `WithTextAlign` | `func WithTextAlign(align TextAlign) Option` | Sets text alignment |

```go
el := tui.New(
    tui.WithText("Warning"),
    tui.WithTextStyle(tui.NewStyle().Bold().Foreground(tui.ANSIColor(tui.Yellow))),
    tui.WithBackground(tui.NewStyle().Background(tui.ANSIColor(tui.Black))),
)
```

## Attr

`Attr` is a bitfield that represents text decorations. Combine multiple attributes with the `|` operator.

```go
type Attr uint8
```

### Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `AttrNone` | `0` | No attributes |
| `AttrBold` | `1 << 0` | Bold / increased intensity |
| `AttrDim` | `1 << 1` | Dim / decreased intensity |
| `AttrItalic` | `1 << 2` | Italic text |
| `AttrUnderline` | `1 << 3` | Underlined text |
| `AttrBlink` | `1 << 4` | Blinking text |
| `AttrReverse` | `1 << 5` | Swapped foreground and background |
| `AttrStrikethrough` | `1 << 6` | Struck-through text |

```go
attrs := tui.AttrBold | tui.AttrUnderline
```

### Tailwind Class Equivalents

| Attr | Tailwind Class |
|------|---------------|
| `AttrBold` | `font-bold` |
| `AttrDim` | `font-dim` or `text-dim` |
| `AttrItalic` | `italic` |
| `AttrUnderline` | `underline` |
| `AttrReverse` | `reverse` |
| `AttrStrikethrough` | `strikethrough` |

```gsx
<span class="font-bold underline text-cyan">Styled text</span>
```

## Color

`Color` represents a terminal color. Three kinds are supported: the terminal's default color, an ANSI palette index (0-255), and 24-bit RGB.

```go
type Color struct {
    // unexported fields
}
```

### Constructors

```go
func DefaultColor() Color
func ANSIColor(index uint8) Color
func RGBColor(r, g, b uint8) Color
func HexColor(hex string) (Color, error)
```

`DefaultColor` returns a color that tells the terminal to use its own default. `ANSIColor` takes a palette index from 0-255. `RGBColor` takes individual red, green, and blue channel values. `HexColor` parses a hex string in `#RGB` or `#RRGGBB` format:

```go
defaultFg := tui.DefaultColor()
red := tui.ANSIColor(tui.Red)       // ANSI palette index
teal := tui.RGBColor(0, 128, 128)   // 24-bit RGB
coral, err := tui.HexColor("#FF7F50")  // Hex notation
short, err := tui.HexColor("#F80")     // Short hex (expands to #FF8800)
```

### Standard Color Constants

These variables hold pre-defined ANSI palette indices for the 16 standard terminal colors:

**Basic colors (indices 0-7):**

| Variable | Index | Typical Appearance |
|----------|-------|--------------------|
| `Black` | 0 | Black |
| `Red` | 1 | Red |
| `Green` | 2 | Green |
| `Yellow` | 3 | Yellow / Brown |
| `Blue` | 4 | Blue |
| `Magenta` | 5 | Magenta / Purple |
| `Cyan` | 6 | Cyan |
| `White` | 7 | White / Light Gray |

**Bright colors (indices 8-15):**

| Variable | Index | Typical Appearance |
|----------|-------|--------------------|
| `BrightBlack` | 8 | Dark Gray |
| `BrightRed` | 9 | Light Red |
| `BrightGreen` | 10 | Light Green |
| `BrightYellow` | 11 | Light Yellow |
| `BrightBlue` | 12 | Light Blue |
| `BrightMagenta` | 13 | Light Magenta |
| `BrightCyan` | 14 | Light Cyan |
| `BrightWhite` | 15 | Bright White |

Use these constants with `ANSIColor`:

```go
style := tui.NewStyle().Foreground(tui.ANSIColor(tui.Cyan))
```

### Tailwind Color Classes

In `.gsx` files, apply colors with class names. Text and background colors use the same set of color names:

**Text colors:** `text-red`, `text-green`, `text-blue`, `text-cyan`, `text-magenta`, `text-yellow`, `text-white`, `text-black`, `text-bright-red`, `text-bright-green`, `text-bright-blue`, `text-bright-cyan`, `text-bright-magenta`, `text-bright-yellow`, `text-bright-white`, `text-bright-black`

**Background colors:** `bg-red`, `bg-green`, `bg-blue`, `bg-cyan`, `bg-magenta`, `bg-yellow`, `bg-white`, `bg-black`, `bg-bright-red`, `bg-bright-green`, `bg-bright-blue`, `bg-bright-cyan`, `bg-bright-magenta`, `bg-bright-yellow`, `bg-bright-white`, `bg-bright-black`

**Hex colors in classes:** `text-#FF7F50`, `bg-#2A2A2A`, `border-#00FF88`

```gsx
<div class="bg-black">
    <span class="text-cyan font-bold">Cyan on black</span>
    <span class="text-#FF7F50">Coral via hex</span>
</div>
```

### Query Methods

```go
func (c Color) Type() ColorType
func (c Color) IsDefault() bool
func (c Color) ANSI() uint8
func (c Color) RGB() (r, g, b uint8)
func (c Color) Equal(other Color) bool
```

`Type` returns the color kind. `IsDefault` returns `true` for the default color. `ANSI` returns the palette index for ANSI colors (0 for others). `RGB` returns the red, green, and blue components for RGB colors (all zeros for others).

### Conversion Methods

```go
func (c Color) ToANSI() Color
func (c Color) ToRGBValues() (r, g, b uint8)
func (c Color) Luminance() float64
func (c Color) IsLight() bool
```

`ToANSI` approximates an RGB color to the nearest entry in the ANSI 256 palette. `ToRGBValues` converts any color kind to approximate RGB values. `Luminance` returns the W3C relative luminance (0.0 for black, 1.0 for white). `IsLight` returns `true` when luminance exceeds 0.5.

```go
bg := tui.RGBColor(30, 30, 30)
if bg.IsLight() {
    // use dark text
} else {
    // use light text
}
```

### ColorType

```go
type ColorType uint8

const (
    ColorDefault ColorType = iota
    ColorANSI
    ColorRGB
)
```

| Constant | Description |
|----------|-------------|
| `ColorDefault` | Terminal's own default color |
| `ColorANSI` | ANSI 256 palette color (index 0-255) |
| `ColorRGB` | 24-bit RGB color |

## Gradient

`Gradient` interpolates between two colors over a direction. Gradients can be applied to text, backgrounds, and borders.

```go
type Gradient struct {
    Start     Color
    End       Color
    Direction GradientDirection
}
```

### Creating a Gradient

```go
func NewGradient(start, end Color) Gradient
```

Creates a gradient that defaults to horizontal direction. Use `WithDirection` to change it:

```go
g := tui.NewGradient(
    tui.ANSIColor(tui.Cyan),
    tui.ANSIColor(tui.Magenta),
).WithDirection(tui.GradientVertical)
```

### Methods

```go
func (g Gradient) WithDirection(d GradientDirection) Gradient
func (g Gradient) At(t float64) Color
```

`WithDirection` returns a new gradient with the given direction. `At` interpolates between the start and end colors at position `t` (0.0 returns the start color, 1.0 returns the end color):

```go
g := tui.NewGradient(tui.ANSIColor(tui.Red), tui.ANSIColor(tui.Blue))
mid := g.At(0.5) // color halfway between red and blue
```

### GradientDirection

```go
type GradientDirection int

const (
    GradientHorizontal  GradientDirection = iota
    GradientVertical
    GradientDiagonalDown
    GradientDiagonalUp
)
```

| Constant | Direction | Description |
|----------|-----------|-------------|
| `GradientHorizontal` | Left to right | Default direction |
| `GradientVertical` | Top to bottom | Vertical sweep |
| `GradientDiagonalDown` | Top-left to bottom-right | Diagonal sweep downward |
| `GradientDiagonalUp` | Bottom-left to top-right | Diagonal sweep upward |

### Applying Gradients to Elements

Use the gradient `Option` functions or `.gsx` attributes:

| Function | Signature | Description |
|----------|-----------|-------------|
| `WithTextGradient` | `func WithTextGradient(g Gradient) Option` | Gradient applied to text color |
| `WithBackgroundGradient` | `func WithBackgroundGradient(g Gradient) Option` | Gradient applied to background |
| `WithBorderGradient` | `func WithBorderGradient(g Gradient) Option` | Gradient applied to border |

```go
el := tui.New(
    tui.WithText("Rainbow text"),
    tui.WithTextGradient(tui.NewGradient(
        tui.ANSIColor(tui.Red),
        tui.ANSIColor(tui.Blue),
    )),
)
```

### Tailwind Gradient Classes

In `.gsx` files, gradients follow the pattern `{target}-gradient-{start}-{end}[-{direction}]`:

**Targets:** `text`, `bg`, `border`

**Colors:** `red`, `green`, `blue`, `cyan`, `magenta`, `yellow`, `white`, `black`, and `bright-*` variants

**Directions:** `-h` (horizontal, default), `-v` (vertical), `-dd` (diagonal down), `-du` (diagonal up)

```gsx
<span class="text-gradient-cyan-magenta">Horizontal gradient text</span>
<span class="text-gradient-red-yellow-v">Vertical gradient text</span>
<div class="bg-gradient-blue-cyan-dd border-rounded p-1">
    <span>Diagonal gradient background</span>
</div>
<div class="border-rounded border-gradient-magenta-cyan">
    <span>Gradient border</span>
</div>
```

### Buffer Gradient Methods

For low-level rendering, `Buffer` provides gradient-aware write methods:

```go
func (b *Buffer) SetStringGradient(x, y int, s string, g Gradient, baseStyle Style) int
func (b *Buffer) FillGradient(rect Rect, r rune, g Gradient, baseStyle Style)
```

`SetStringGradient` writes a string with the gradient applied per-character as the foreground color. It returns the total display width consumed. `FillGradient` fills a rectangle with the gradient applied as the background color, respecting the gradient's direction setting.

## BorderStyle

`BorderStyle` selects the character set used to draw element borders.

```go
type BorderStyle int
```

### Constants

| Constant | Characters | Example |
|----------|-----------|---------|
| `BorderNone` | (no border) | |
| `BorderSingle` | `┌─┐│└─┘` | `┌───┐` / `│   │` / `└───┘` |
| `BorderDouble` | `╔═╗║╚═╝` | `╔═══╗` / `║   ║` / `╚═══╝` |
| `BorderRounded` | `╭─╮│╰─╯` | `╭───╮` / `│   │` / `╰───╯` |
| `BorderThick` | `┏━┓┃┗━┛` | `┏━━━┓` / `┃   ┃` / `┗━━━┛` |

### Tailwind Border Classes

| Class | BorderStyle |
|-------|-------------|
| `border-single` | `BorderSingle` |
| `border-double` | `BorderDouble` |
| `border-rounded` | `BorderRounded` |
| `border-thick` | `BorderThick` |

Border color classes: `border-red`, `border-cyan`, etc. apply a `Style` to the border characters.

```gsx
<div class="border-rounded border-cyan p-1">
    Rounded cyan border
</div>
```

### BorderChars

`Chars()` returns the individual rune characters for a border style:

```go
func (b BorderStyle) Chars() BorderChars

type BorderChars struct {
    TopLeft     rune
    Top         rune
    TopRight    rune
    Left        rune
    Right       rune
    BottomLeft  rune
    Bottom      rune
    BottomRight rune
}
```

```go
chars := tui.BorderRounded.Chars()
// chars.TopLeft == '╭', chars.Top == '─', chars.TopRight == '╮'
```

### Applying Borders to Elements

```go
func WithBorder(style BorderStyle) Option
```

Sets the border style on an element:

```go
el := tui.New(
    tui.WithBorder(tui.BorderRounded),
    tui.WithBorderStyle(tui.NewStyle().Foreground(tui.ANSIColor(tui.Cyan))),
    tui.WithText("Bordered content"),
)
```

## Drawing Functions

These functions render borders and fills directly to a `Buffer`. They're used internally by the rendering system, but are available for custom rendering needs.

```go
func DrawBox(buf *Buffer, rect Rect, border BorderStyle, style Style)
func DrawBoxGradient(buf *Buffer, rect Rect, border BorderStyle, g Gradient, baseStyle Style)
func DrawBoxClipped(buf *Buffer, rect Rect, border BorderStyle, style Style, clipRect Rect)
func DrawBoxGradientClipped(buf *Buffer, rect Rect, border BorderStyle, g Gradient, baseStyle Style, clipRect Rect)
func DrawBoxWithTitle(buf *Buffer, rect Rect, border BorderStyle, title string, style Style)
func FillBox(buf *Buffer, rect Rect, r rune, style Style)
```

| Function | Description |
|----------|-------------|
| `DrawBox` | Draws a border around a rectangle |
| `DrawBoxGradient` | Draws a border with gradient color applied around the perimeter |
| `DrawBoxClipped` | Draws a border clipped to a visible region (for scrolling) |
| `DrawBoxGradientClipped` | Draws a gradient border clipped to a visible region |
| `DrawBoxWithTitle` | Draws a border with a centered title in the top edge |
| `FillBox` | Fills the interior of a rectangle with a rune and style |

```go
buf := tui.NewBuffer(40, 10)
rect := tui.NewRect(0, 0, 40, 10)
style := tui.NewStyle().Foreground(tui.ANSIColor(tui.Cyan))

tui.DrawBox(buf, rect, tui.BorderRounded, style)
tui.DrawBoxWithTitle(buf, rect, tui.BorderSingle, "My Panel", style)
```

## Capabilities

`Capabilities` describes what a terminal supports. The framework uses this to downgrade colors automatically. An RGB color on a 16-color terminal maps to the nearest ANSI equivalent.

```go
type Capabilities struct {
    Colors    ColorCapability
    Unicode   bool
    TrueColor bool
    AltScreen bool
}
```

### DetectCapabilities

```go
func DetectCapabilities() Capabilities
```

Reads environment variables to determine terminal support. Detection checks, in order:

1. `COLORTERM` — `"truecolor"` or `"24bit"` sets `ColorTrue`
2. Terminal-specific variables — `WT_SESSION`, `ITERM_SESSION_ID`, `KITTY_WINDOW_ID`, `KONSOLE_VERSION`, `VTE_VERSION` each set `ColorTrue`
3. `TERM` — contains `"256color"` sets `Color256`, contains `"truecolor"` sets `ColorTrue`, equals `"dumb"` sets `ColorNone` with Unicode and AltScreen disabled
4. Default — `Color16`

### ColorCapability

```go
type ColorCapability int

const (
    ColorNone  ColorCapability = iota
    Color16
    Color256
    ColorTrue
)
```

| Constant | Description |
|----------|-------------|
| `ColorNone` | Monochrome, no color support |
| `Color16` | Standard 16 ANSI colors |
| `Color256` | Extended 256-color palette |
| `ColorTrue` | Full 24-bit RGB (16 million colors) |

### Methods

```go
func (c Capabilities) SupportsColor(color Color) bool
func (c Capabilities) EffectiveColor(color Color) Color
func (c Capabilities) String() string
```

`SupportsColor` returns `true` if the terminal can render the given color without conversion. `EffectiveColor` returns the color as-is if supported, or its nearest fallback if not. `String` returns a human-readable description.

```go
caps := tui.DetectCapabilities()
coral := tui.RGBColor(255, 127, 80)

if caps.SupportsColor(coral) {
    // terminal handles 24-bit color natively
} else {
    fallback := caps.EffectiveColor(coral)
    // fallback is the nearest ANSI approximation
}
```

## See Also

- [Element Reference](element.md) — Element option functions for styling
- [GSX Syntax Reference](gsx-syntax.md) — Complete Tailwind class listing
- [Layout Reference](layout.md) — Layout-related element options
