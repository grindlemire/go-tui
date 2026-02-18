# Buffer and Rendering Reference

## Overview

go-tui uses a double-buffered character grid to manage terminal output. Every frame, your component tree renders into the back buffer. The framework then diffs the back buffer against the front buffer, sends only the changed cells to the terminal, and swaps the buffers. This keeps screen updates fast and flicker-free.

Most applications never interact with the buffer directly. The `App` handles layout, rendering, and flushing automatically. These types matter when you're writing tests (reading buffer contents), building custom rendering logic, or working with the lower-level drawing functions.

## Buffer

```go
type Buffer struct {
    // unexported fields
}
```

A double-buffered 2D grid of `Cell` values. Writes go to the back buffer. `Diff()` computes what changed, and `Swap()` promotes the back buffer to front.

### NewBuffer

```go
func NewBuffer(width, height int) *Buffer
```

Creates a new buffer with the given dimensions. Both buffers are initialized with spaces and default styling. Negative dimensions are clamped to 0.

```go
buf := tui.NewBuffer(80, 24)
```

### Query methods

#### Width

```go
func (b *Buffer) Width() int
```

Returns the buffer width in columns.

#### Height

```go
func (b *Buffer) Height() int
```

Returns the buffer height in rows.

#### Size

```go
func (b *Buffer) Size() (width, height int)
```

Returns both dimensions at once.

#### Rect

```go
func (b *Buffer) Rect() Rect
```

Returns the buffer bounds as a `Rect` starting at `(0, 0)`.

```go
r := buf.Rect() // Rect{X: 0, Y: 0, Width: 80, Height: 24}
```

### Reading cells

#### Cell

```go
func (b *Buffer) Cell(x, y int) Cell
```

Returns the cell at `(x, y)` from the back buffer. Returns an empty `Cell{}` if the position is out of bounds.

```go
c := buf.Cell(5, 3)
if c.Rune == '>' {
    // ...
}
```

### Writing cells

#### SetCell

```go
func (b *Buffer) SetCell(x, y int, c Cell)
```

Sets the cell at `(x, y)` in the back buffer. Does nothing if the position is out of bounds.

```go
buf.SetCell(0, 0, tui.NewCell('X', tui.NewStyle().Bold()))
```

#### SetRune

```go
func (b *Buffer) SetRune(x, y int, r rune, style Style)
```

Sets a rune at `(x, y)` with the given style. Handles wide characters (CJK, emoji) by automatically placing continuation cells. Also cleans up any overlapped wide characters at the target position.

If a wide character (width 2) would land in the last column, a space is placed instead since the character can't fit.

```go
buf.SetRune(10, 5, 'A', tui.NewStyle().Foreground(tui.ANSIColor(tui.Cyan)))
```

#### SetString

```go
func (b *Buffer) SetString(x, y int, s string, style Style) int
```

Writes a string starting at `(x, y)`. Returns the total display width consumed. Stops at the buffer edge without wrapping. Handles wide characters correctly.

```go
w := buf.SetString(2, 0, "Hello, world!", tui.NewStyle())
// w == 13
```

#### SetStringClipped

```go
func (b *Buffer) SetStringClipped(x, y int, s string, style Style, clipRect Rect) int
```

Writes a string clipped to a rectangle. Characters outside `clipRect` are not rendered. Returns the display width of rendered characters.

```go
clip := tui.NewRect(5, 0, 10, 1)
buf.SetStringClipped(0, 0, "This is a long string", tui.NewStyle(), clip)
// Only characters within columns 5-14 are drawn
```

### Gradient writing

#### SetStringGradient

```go
func (b *Buffer) SetStringGradient(x, y int, s string, g Gradient, baseStyle Style) int
```

Writes a string with a gradient applied per-character along the horizontal axis. The gradient color is applied to the foreground. Returns the total display width consumed.

```go
g := tui.NewGradient(tui.ANSIColor(tui.Cyan), tui.ANSIColor(tui.Magenta))
buf.SetStringGradient(0, 0, "Rainbow Text", g, tui.NewStyle())
```

#### FillGradient

```go
func (b *Buffer) FillGradient(rect Rect, r rune, g Gradient, baseStyle Style)
```

Fills a rectangle with a gradient background. The gradient direction determines how the color interpolates:

| Direction | Interpolation |
|-----------|--------------|
| `GradientHorizontal` | Left to right |
| `GradientVertical` | Top to bottom |
| `GradientDiagonalDown` | Top-left to bottom-right |
| `GradientDiagonalUp` | Bottom-left to top-right |

The gradient color is applied to the background of each cell.

```go
g := tui.NewGradient(tui.ANSIColor(tui.Blue), tui.ANSIColor(tui.Green))
g = g.WithDirection(tui.GradientVertical)
buf.FillGradient(buf.Rect(), ' ', g, tui.NewStyle())
```

### Fill and clear

#### Fill

```go
func (b *Buffer) Fill(rect Rect, r rune, style Style)
```

Fills a rectangle with the given rune and style. The rect is intersected with the buffer bounds, so out-of-bounds regions are ignored.

```go
buf.Fill(tui.NewRect(0, 0, 20, 5), '.', tui.NewStyle().Foreground(tui.ANSIColor(tui.BrightBlack)))
```

#### Clear

```go
func (b *Buffer) Clear()
```

Clears the entire back buffer to spaces with default styling. Equivalent to `ClearRect(buf.Rect())`.

#### ClearRect

```go
func (b *Buffer) ClearRect(rect Rect)
```

Clears a rectangular region to spaces with default styling. Handles wide characters at the edges of the cleared region (clears originating or continuation cells as needed).

### Diffing and swapping

#### Diff

```go
func (b *Buffer) Diff() []CellChange
```

Returns all cells that changed between the front and back buffers. Changes are returned in row-major order (top-to-bottom, left-to-right), which minimizes cursor movement when writing to the terminal.

#### Swap

```go
func (b *Buffer) Swap()
```

Copies the back buffer to the front buffer. Call this after flushing changes to the terminal so the next diff starts from the current state.

The typical rendering cycle is:

```go
// 1. Write to the back buffer
buf.SetString(0, 0, "Hello", style)

// 2. Compute what changed
changes := buf.Diff()

// 3. Send changes to the terminal
term.Flush(changes)

// 4. Promote back to front
buf.Swap()
```

### Output

#### String

```go
func (b *Buffer) String() string
```

Renders the back buffer to a string for debugging or testing. Each row is separated by a newline. Continuation cells (from wide characters) are skipped.

```go
fmt.Println(buf.String())
```

#### StringTrimmed

```go
func (b *Buffer) StringTrimmed() string
```

Like `String()`, but trailing spaces are removed from each line. Useful in tests where you want to assert on content without worrying about trailing whitespace.

```go
got := buf.StringTrimmed()
if got != "Hello\nWorld" {
    t.Errorf("unexpected: %q", got)
}
```

### Resizing

#### Resize

```go
func (b *Buffer) Resize(width, height int)
```

Changes the buffer dimensions. Content in the overlapping region is preserved; new areas are filled with spaces. Does nothing if the dimensions haven't changed. Negative values are clamped to 0.

## Cell

```go
type Cell struct {
    Rune  rune   // The character (0 for continuation cells)
    Style Style  // Visual styling
    Width uint8  // Display width: 1 for normal, 2 for wide, 0 for continuation
}
```

A single character cell in the terminal buffer. Wide characters (CJK, emoji) occupy two cells: the first cell holds the rune with `Width == 2`, and the next cell is a continuation with `Rune == 0` and `Width == 0`.

### NewCell

```go
func NewCell(r rune, style Style) Cell
```

Creates a Cell with automatic width detection via `RuneWidth()`.

```go
c := tui.NewCell('A', tui.NewStyle().Bold())
// c.Width == 1

c = tui.NewCell('漢', tui.NewStyle())
// c.Width == 2
```

### NewCellWithWidth

```go
func NewCellWithWidth(r rune, style Style, width uint8) Cell
```

Creates a Cell with an explicit width. Use this for continuation cells (`width == 0`) or when the width is already known.

```go
// Continuation cell for the second half of a wide character
cont := tui.NewCellWithWidth(0, tui.NewStyle(), 0)
```

### IsContinuation

```go
func (c Cell) IsContinuation() bool
```

Returns `true` if this cell is the second half of a wide character (`Width == 0`). Continuation cells should not be drawn directly.

### Equal

```go
func (c Cell) Equal(other Cell) bool
```

Returns `true` if both cells have the same rune, style, and width. Used internally by `Diff()` to detect changes.

### IsEmpty

```go
func (c Cell) IsEmpty() bool
```

Returns `true` if the cell is blank. A cell is empty when:
- Its rune is `0` (regardless of style), or
- Its rune is `' '` (space) with default styling

### RuneWidth

```go
func RuneWidth(r rune) int
```

Package-level function that returns the display width of a rune in terminal cells. Returns 1 for most characters and 2 for wide characters (CJK ideographs, fullwidth forms, emoji).

Zero-width Unicode categories (combining marks, variation selectors, format controls) are treated as width 1 in this model because `Width == 0` is reserved for continuation cells.

```go
tui.RuneWidth('A')  // 1
tui.RuneWidth('漢') // 2
tui.RuneWidth('🎉') // 2
```

## CellChange

```go
type CellChange struct {
    X, Y int
    Cell Cell
}
```

Represents a single cell that differs between the front and back buffers. Returned by `Buffer.Diff()` and consumed by `Terminal.Flush()`.

## Render functions

These package-level functions handle the full render cycle.

### Render

```go
func Render(term Terminal, buf *Buffer)
```

The primary rendering function for normal frame updates. Computes the diff between front and back buffers, flushes only the changed cells to the terminal via `term.Flush()`, then swaps the buffers. If nothing changed, no terminal I/O occurs.

### RenderFull

```go
func RenderFull(term Terminal, buf *Buffer)
```

Forces a complete redraw. Sends every cell to the terminal regardless of whether it changed. Calls `term.Clear()` first, then flushes all cells, then swaps. Use this after:

- Initial application startup
- Terminal resize
- Recovering from external terminal corruption
- Switching back from alternate screen

### RenderTree

```go
func RenderTree(buf *Buffer, root *Element)
```

Traverses an element tree and draws each element to the buffer. Handles background fills, borders, text content, gradients, scroll clipping, and overflow clipping. This is the bridge between the layout system and the buffer: after layout calculates positions, `RenderTree` writes the visual output.

```go
root := tui.New(
    tui.WithText("Hello"),
    tui.WithBorder(tui.BorderRounded),
)
root.Calculate(80, 24)

buf := tui.NewBuffer(80, 24)
tui.RenderTree(buf, root)
fmt.Println(buf.StringTrimmed())
```

## Drawing functions

These functions draw borders and filled boxes directly into a buffer. They live in the `tui` package alongside the buffer types. For full details, see [Styling Reference](styling.md).

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
| `DrawBox` | Draws a border around a rectangle. |
| `DrawBoxGradient` | Draws a border with a gradient applied to the border characters. |
| `DrawBoxClipped` | Draws a border clipped to a visible region. |
| `DrawBoxGradientClipped` | Draws a gradient border clipped to a visible region. |
| `DrawBoxWithTitle` | Draws a border with a title string inset in the top edge. |
| `FillBox` | Fills the interior of a bordered rectangle (inside the border, not including it). |

## See also

- [Element Reference](element.md) -- element tree that produces the content rendered into buffers
- [Styling Reference](styling.md) -- Style, Color, Gradient, and BorderStyle types
- [Terminal Reference](terminal.md) -- the Terminal interface that receives flushed changes
- [Testing Reference](testing.md) -- MockTerminal for reading buffer output in tests
