# Phase 1: Terminal Rendering Foundation — Design Specification

**Status:** Planned
**Version:** 1.0
**Last Updated:** 2025-01-24

---

## 1. Overview

### Purpose

Phase 1 establishes the foundational rendering layer for go-tui: the ability to efficiently render styled text and shapes at arbitrary terminal positions. This layer provides the building blocks that all subsequent phases depend on—without a solid, performant rendering foundation, the layout engine, widget system, and everything above will suffer.

This design prioritizes:
1. **Performance** — Double buffering with diff-based updates minimizes terminal I/O
2. **Flexibility** — True color support with ANSI 256 fallback; wide character handling
3. **Maintainability** — Clean interfaces, value semantics where appropriate, idiomatic Go
4. **Testability** — All components are testable without a real terminal

### Goals

- Render styled text (bold, italic, colors) at arbitrary x,y positions
- Support both ANSI 256 and true color (24-bit RGB)
- Handle wide characters (CJK, emoji) that span multiple columns
- Implement double buffering to compute minimal screen updates
- Provide terminal abstraction with capability auto-detection
- Draw box-drawing characters for borders
- Achieve sub-millisecond render times for typical 80x24 to 200x60 screens

### Non-Goals

- Layout computation (Phase 2)
- Widget abstraction (Phase 3)
- Event/input handling (Phase 4)
- Mouse support
- Animation or frame timing utilities

---

## 2. Architecture

### Directory Structure

```
pkg/tui/
├── color.go          # Color type with true color and ANSI 256 support
├── color_test.go
├── style.go          # Text styling (bold, italic, fg/bg colors)
├── style_test.go
├── cell.go           # Single terminal cell (rune + style + width)
├── cell_test.go
├── rect.go           # Rectangle geometry utilities
├── rect_test.go
├── buffer.go         # Double-buffered character grid
├── buffer_test.go
├── terminal.go       # Terminal interface definition
├── terminal_ansi.go  # ANSI terminal implementation
├── terminal_test.go
├── caps.go           # Terminal capability detection
├── caps_test.go
└── escape.go         # ANSI escape sequence builder (internal)
```

### Component Overview

| Component | Purpose |
|-----------|---------|
| `color.go` | Unified color representation supporting default, ANSI 256, and true color |
| `style.go` | Text attributes (bold, italic, etc.) combined with foreground/background colors |
| `cell.go` | Atomic unit of the terminal: one grapheme with styling and display width |
| `rect.go` | Rectangle geometry for positioning and clipping |
| `buffer.go` | Double-buffered 2D grid of cells with diff computation |
| `terminal.go` | Interface abstracting terminal operations |
| `terminal_ansi.go` | Concrete implementation using ANSI escape sequences |
| `caps.go` | Runtime detection of terminal capabilities |
| `escape.go` | Internal helper for building escape sequences efficiently |

### Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  Application Code                                               │
│  buf.SetCell(x, y, 'A', style)                                  │
│  buf.SetString(x, y, "Hello", style)                            │
└─────────────────────┬───────────────────────────────────────────┘
                      │ Write to back buffer
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│  Buffer (Double Buffered)                                       │
│  ┌─────────────┐    ┌─────────────┐                             │
│  │ Front Buffer│    │ Back Buffer │  ← writes go here           │
│  │ (displayed) │    │ (building)  │                             │
│  └─────────────┘    └─────────────┘                             │
└─────────────────────┬───────────────────────────────────────────┘
                      │ Flush() computes diff
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│  Terminal                                                       │
│  - Receives only changed cells                                  │
│  - Batches escape sequences                                     │
│  - Writes to underlying io.Writer                               │
└─────────────────────┬───────────────────────────────────────────┘
                      │ ANSI escape sequences
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│  TTY / stdout                                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## 3. Core Entities

### 3.1 Color

The `Color` type must handle three distinct color spaces while maintaining a clean API and efficient representation.

```go
// ColorType distinguishes between color representations
type ColorType uint8

const (
    ColorDefault ColorType = iota  // Terminal default (no color set)
    ColorANSI                      // ANSI 256 palette (0-255)
    ColorRGB                       // True color (24-bit)
)

// Color represents a terminal color with support for default, ANSI 256, and true color.
// Zero value represents the terminal default color.
type Color struct {
    typ ColorType
    // For ANSI: r holds the palette index (0-255)
    // For RGB: r, g, b hold the color components
    r, g, b uint8
}

// Constructors
func DefaultColor() Color                    // Terminal default
func ANSIColor(index uint8) Color            // ANSI 256 palette
func RGBColor(r, g, b uint8) Color           // True color
func HexColor(hex string) (Color, error)     // Parse "#RRGGBB" or "#RGB"

// Methods
func (c Color) Type() ColorType
func (c Color) IsDefault() bool
func (c Color) ANSI() uint8                  // Returns palette index (panics if not ANSI)
func (c Color) RGB() (r, g, b uint8)         // Returns RGB values (panics if not RGB)
func (c Color) ToANSI() Color                // Approximate RGB as nearest ANSI color
func (c Color) Equal(other Color) bool

// Common ANSI colors as package-level vars for convenience
var (
    Black   = ANSIColor(0)
    Red     = ANSIColor(1)
    Green   = ANSIColor(2)
    Yellow  = ANSIColor(3)
    Blue    = ANSIColor(4)
    Magenta = ANSIColor(5)
    Cyan    = ANSIColor(6)
    White   = ANSIColor(7)
    // ... bright variants (8-15)
)
```

**Design Rationale:**
- **Struct instead of interface**: Avoids allocation on every color, enables value semantics
- **Compact representation**: 4 bytes total (1 type + 3 color bytes), fits in a register
- **Explicit type discrimination**: Prevents accidental misuse of RGB values as ANSI indices
- **ToANSI() method**: Enables graceful degradation for terminals without true color

### 3.2 Style

```go
// Attr represents text attributes as a bitfield for efficient comparison and storage.
type Attr uint8

const (
    AttrNone      Attr = 0
    AttrBold      Attr = 1 << iota
    AttrDim
    AttrItalic
    AttrUnderline
    AttrBlink
    AttrReverse
    AttrStrikethrough
)

// Style combines text attributes with foreground and background colors.
// Zero value represents default styling (no attributes, default colors).
type Style struct {
    Fg    Color
    Bg    Color
    Attrs Attr
}

// Constructors and modifiers (fluent API)
func NewStyle() Style
func (s Style) Foreground(c Color) Style
func (s Style) Background(c Color) Style
func (s Style) Bold() Style
func (s Style) Dim() Style
func (s Style) Italic() Style
func (s Style) Underline() Style
func (s Style) Reverse() Style
func (s Style) Strikethrough() Style

// Queries
func (s Style) Equal(other Style) bool
func (s Style) HasAttr(a Attr) bool
```

**Design Rationale:**
- **Bitfield for attributes**: Single comparison checks all attributes; efficient storage
- **Fluent API**: Enables readable style construction: `NewStyle().Bold().Foreground(Red)`
- **Value semantics**: Styles are immutable; methods return new Style values
- **Zero value is useful**: Default style requires no initialization

### 3.3 Cell

```go
// Cell represents a single character cell in the terminal buffer.
// Wide characters (CJK, emoji) occupy multiple cells; the first cell holds
// the rune, subsequent cells are marked as continuations.
type Cell struct {
    Rune  rune   // The character (0 for continuation cells)
    Style Style  // Visual styling
    Width uint8  // Display width (1 or 2; 0 for continuation)
}

// Constructors
func NewCell(r rune, style Style) Cell           // Auto-detects width
func NewCellWithWidth(r rune, style Style, width uint8) Cell

// Methods
func (c Cell) IsContinuation() bool              // Width == 0
func (c Cell) Equal(other Cell) bool
func (c Cell) IsEmpty() bool                     // Rune == 0 or space with default style

// Width detection
func RuneWidth(r rune) int                       // Returns 1 or 2
```

**Design Rationale:**
- **Width field**: Essential for correct cursor positioning with wide characters
- **Continuation cell concept**: When a wide character spans columns 5-6, column 5 holds the rune with Width=2, column 6 holds a continuation cell (Rune=0, Width=0)
- **Compact struct**: 8 bytes typical (4 rune + 3 style + 1 width), cache-friendly

**Wide Character Handling:**
```
"你好" renders as:
  Col: 0   1   2   3
     ┌───┬───┬───┬───┐
     │你 │(c)│好 │(c)│   (c) = continuation cell
     └───┴───┴───┴───┘

  cells[0] = Cell{Rune: '你', Width: 2}
  cells[1] = Cell{Rune: 0, Width: 0}     // continuation
  cells[2] = Cell{Rune: '好', Width: 2}
  cells[3] = Cell{Rune: 0, Width: 0}     // continuation
```

### 3.4 Rect

```go
// Rect represents a rectangle in terminal coordinates.
// X and Y are 0-indexed from the top-left of the terminal.
type Rect struct {
    X, Y          int
    Width, Height int
}

// Constructors
func NewRect(x, y, width, height int) Rect

// Geometry methods
func (r Rect) Right() int                        // X + Width
func (r Rect) Bottom() int                       // Y + Height
func (r Rect) Area() int
func (r Rect) IsEmpty() bool
func (r Rect) Contains(x, y int) bool
func (r Rect) ContainsRect(other Rect) bool

// Transformations (return new Rect)
func (r Rect) Inset(top, right, bottom, left int) Rect
func (r Rect) InsetUniform(n int) Rect
func (r Rect) Intersect(other Rect) Rect
func (r Rect) Union(other Rect) Rect
func (r Rect) Translate(dx, dy int) Rect
func (r Rect) Clamp(x, y int) (int, int)         // Clamp point to rect bounds
```

**Design Rationale:**
- **Value semantics**: All methods return new Rect values
- **Inset with named edges**: `Inset(top, right, bottom, left)` follows CSS convention
- **Intersect for clipping**: Critical for implementing scrolling regions later

---

## 4. Buffer Design

### 4.1 Buffer Structure

```go
// Buffer is a double-buffered 2D grid of cells.
// Writes go to the back buffer; Flush() computes the diff and swaps buffers.
type Buffer struct {
    front  []Cell  // Currently displayed state
    back   []Cell  // State being built
    width  int
    height int
}

// Constructors
func NewBuffer(width, height int) *Buffer

// Dimensions
func (b *Buffer) Width() int
func (b *Buffer) Height() int
func (b *Buffer) Size() (width, height int)
func (b *Buffer) Rect() Rect                     // Returns Rect{0, 0, width, height}
func (b *Buffer) Resize(width, height int)       // Preserves content where possible

// Cell access (all operations on back buffer)
func (b *Buffer) Cell(x, y int) Cell             // Get cell at position
func (b *Buffer) SetCell(x, y int, c Cell)       // Set cell at position
func (b *Buffer) SetRune(x, y int, r rune, style Style)  // Convenience method

// Drawing primitives
func (b *Buffer) SetString(x, y int, s string, style Style) int  // Returns width consumed
func (b *Buffer) Fill(rect Rect, r rune, style Style)
func (b *Buffer) Clear()                         // Fill with spaces, default style
func (b *Buffer) ClearRect(rect Rect)

// Diff computation
func (b *Buffer) Diff() []CellChange             // Returns changed cells
func (b *Buffer) Swap()                          // Swap front/back (called after flush)
```

### 4.2 Diff Computation

```go
// CellChange represents a single cell that differs between front and back buffers.
type CellChange struct {
    X, Y int
    Cell Cell
}

// Diff returns all cells that changed between front and back buffers.
// Cells are returned in row-major order (top-to-bottom, left-to-right)
// which optimizes terminal output by minimizing cursor moves.
func (b *Buffer) Diff() []CellChange {
    changes := make([]CellChange, 0, b.width) // Pre-allocate one row
    for y := 0; y < b.height; y++ {
        for x := 0; x < b.width; x++ {
            idx := y*b.width + x
            if !b.back[idx].Equal(b.front[idx]) {
                changes = append(changes, CellChange{X: x, Y: y, Cell: b.back[idx]})
            }
        }
    }
    return changes
}
```

### 4.3 Wide Character Handling in Buffer

```go
// SetRune handles wide characters by:
// 1. Determining the rune's display width
// 2. Setting the primary cell with the rune
// 3. Setting continuation cells for wide characters
// 4. Clearing any cells that would be "overwritten" by the wide char
func (b *Buffer) SetRune(x, y int, r rune, style Style) {
    if x < 0 || x >= b.width || y < 0 || y >= b.height {
        return // Out of bounds
    }

    width := RuneWidth(r)

    // If this position was a continuation, clear the previous wide char
    if cell := b.Cell(x, y); cell.IsContinuation() {
        b.clearWideCharAt(x, y)
    }

    // If placing a wide char would overlap an existing wide char, clear it
    if width == 2 && x+1 < b.width {
        if next := b.Cell(x+1, y); !next.IsContinuation() && next.Width == 2 {
            b.clearWideCharAt(x+1, y)
        }
    }

    // Set the primary cell
    b.SetCell(x, y, NewCellWithWidth(r, style, uint8(width)))

    // Set continuation cell for wide characters
    if width == 2 && x+1 < b.width {
        b.SetCell(x+1, y, Cell{Rune: 0, Style: style, Width: 0})
    }
}
```

### 4.4 Memory Layout

```
Single contiguous slice for cache efficiency:

back = []Cell with length width * height

Index calculation: idx = y * width + x

For a 4x3 buffer:
  ┌───┬───┬───┬───┐
  │ 0 │ 1 │ 2 │ 3 │  y=0
  ├───┼───┼───┼───┤
  │ 4 │ 5 │ 6 │ 7 │  y=1
  ├───┼───┼───┼───┤
  │ 8 │ 9 │10 │11 │  y=2
  └───┴───┴───┴───┘
```

**Design Rationale:**
- **Flat slice vs [][]Cell**: Single allocation, better cache locality, simpler index math
- **Row-major order**: Matches terminal's natural scan order, optimizes sequential writes
- **Pre-allocated diff slice**: Reduces allocations during render

---

## 5. Terminal Abstraction

### 5.1 Terminal Interface

```go
// Terminal abstracts terminal operations for rendering and input.
// Implementations handle ANSI, Windows Console, or mock terminals for testing.
type Terminal interface {
    // Dimensions
    Size() (width, height int)

    // Rendering
    Flush(changes []CellChange)    // Write changed cells to terminal
    Clear()                        // Clear entire terminal

    // Cursor control
    SetCursor(x, y int)
    HideCursor()
    ShowCursor()

    // Mode control
    EnterRawMode() error
    ExitRawMode() error
    EnterAltScreen()
    ExitAltScreen()

    // Capabilities
    Caps() Capabilities
}

// Capabilities describes what features the terminal supports.
type Capabilities struct {
    Colors    ColorCapability  // Color support level
    Unicode   bool             // Can render Unicode characters
    TrueColor bool             // Can use 24-bit RGB colors
    AltScreen bool             // Supports alternate screen buffer
}

type ColorCapability int

const (
    ColorNone  ColorCapability = iota  // Monochrome
    Color16                            // Basic 16 colors
    Color256                           // ANSI 256 palette
    ColorTrue                          // 24-bit true color
)
```

### 5.2 ANSI Terminal Implementation

```go
// ANSITerminal implements Terminal using ANSI escape sequences.
type ANSITerminal struct {
    out       io.Writer        // Usually os.Stdout
    in        io.Reader        // Usually os.Stdin (for raw mode)
    caps      Capabilities     // Detected or configured capabilities
    lastStyle Style            // Optimization: track last emitted style
    buf       *bytes.Buffer    // Batch escape sequences before write
    rawState  *rawModeState    // Platform-specific raw mode state
}

// Constructor
func NewANSITerminal(out io.Writer, in io.Reader) (*ANSITerminal, error)
func NewANSITerminalWithCaps(out io.Writer, in io.Reader, caps Capabilities) *ANSITerminal

// Platform-specific raw mode (build-tagged implementations)
// terminal_unix.go, terminal_windows.go
type rawModeState struct { /* platform-specific */ }
func enableRawMode(fd int) (*rawModeState, error)
func disableRawMode(state *rawModeState) error
```

### 5.3 Escape Sequence Generation

```go
// escape.go - internal helpers for building ANSI sequences

// escBuilder efficiently builds escape sequences
type escBuilder struct {
    buf []byte
}

func (e *escBuilder) Reset()
func (e *escBuilder) MoveTo(x, y int)            // \x1b[{y};{x}H (1-indexed)
func (e *escBuilder) MoveUp(n int)               // \x1b[{n}A
func (e *escBuilder) MoveDown(n int)             // \x1b[{n}B
func (e *escBuilder) MoveRight(n int)            // \x1b[{n}C
func (e *escBuilder) MoveLeft(n int)             // \x1b[{n}D
func (e *escBuilder) ClearScreen()               // \x1b[2J
func (e *escBuilder) ClearLine()                 // \x1b[2K
func (e *escBuilder) HideCursor()                // \x1b[?25l
func (e *escBuilder) ShowCursor()                // \x1b[?25h
func (e *escBuilder) EnterAltScreen()            // \x1b[?1049h
func (e *escBuilder) ExitAltScreen()             // \x1b[?1049l
func (e *escBuilder) ResetStyle()                // \x1b[0m
func (e *escBuilder) SetStyle(s Style, caps Capabilities)
func (e *escBuilder) WriteRune(r rune)
func (e *escBuilder) Bytes() []byte
```

### 5.4 Optimized Flush Algorithm

```go
// Flush writes changed cells to the terminal efficiently.
// It batches escape sequences and optimizes cursor movement.
func (t *ANSITerminal) Flush(changes []CellChange) {
    if len(changes) == 0 {
        return
    }

    t.buf.Reset()
    esc := &escBuilder{buf: t.buf.Bytes()}

    lastX, lastY := -1, -1

    for _, ch := range changes {
        // Optimize cursor movement
        if ch.Y != lastY || ch.X != lastX+1 {
            // Must move cursor (not sequential)
            esc.MoveTo(ch.X, ch.Y)
        }
        // If X == lastX+1 and Y == lastY, cursor is already in position

        // Only emit style changes when style differs
        if !ch.Cell.Style.Equal(t.lastStyle) {
            esc.SetStyle(ch.Cell.Style, t.caps)
            t.lastStyle = ch.Cell.Style
        }

        // Write the character
        if !ch.Cell.IsContinuation() {
            esc.WriteRune(ch.Cell.Rune)
        }

        lastX, lastY = ch.X, ch.Y
    }

    t.out.Write(esc.Bytes())
}
```

---

## 6. Capability Detection

### 6.1 Detection Strategy

```go
// DetectCapabilities determines terminal capabilities from environment
// and optional terminal queries.
func DetectCapabilities() Capabilities {
    caps := Capabilities{
        Colors:    Color16,      // Safe default
        Unicode:   true,         // Assume modern terminal
        TrueColor: false,
        AltScreen: true,
    }

    // Check TERM environment variable
    term := os.Getenv("TERM")
    switch {
    case strings.Contains(term, "256color"):
        caps.Colors = Color256
    case strings.Contains(term, "truecolor"):
        caps.Colors = ColorTrue
        caps.TrueColor = true
    case term == "dumb":
        caps.Colors = ColorNone
        caps.Unicode = false
        caps.AltScreen = false
    }

    // Check COLORTERM for explicit true color support
    if ct := os.Getenv("COLORTERM"); ct == "truecolor" || ct == "24bit" {
        caps.Colors = ColorTrue
        caps.TrueColor = true
    }

    // Check terminal emulator-specific env vars
    if os.Getenv("WT_SESSION") != "" {  // Windows Terminal
        caps.Colors = ColorTrue
        caps.TrueColor = true
    }
    if os.Getenv("ITERM_SESSION_ID") != "" {  // iTerm2
        caps.Colors = ColorTrue
        caps.TrueColor = true
    }

    return caps
}
```

### 6.2 Color Fallback

```go
// ToANSI approximates an RGB color to the nearest ANSI 256 palette entry.
// Uses the 6x6x6 color cube (indices 16-231) plus grayscale (232-255).
func (c Color) ToANSI() Color {
    if c.typ != ColorRGB {
        return c
    }

    r, g, b := c.r, c.g, c.b

    // Check if grayscale
    if r == g && g == b {
        // Grayscale ramp: 232-255 (24 shades)
        // 0 maps to 232, 255 maps to 255
        gray := 232 + int(r)*23/255
        return ANSIColor(uint8(gray))
    }

    // 6x6x6 color cube: 16-231
    // Each component maps to 0-5
    ri := int(r) * 5 / 255
    gi := int(g) * 5 / 255
    bi := int(b) * 5 / 255

    index := 16 + 36*ri + 6*gi + bi
    return ANSIColor(uint8(index))
}
```

---

## 7. Testing Strategy

### 7.1 Unit Tests

| Component | Test Focus |
|-----------|------------|
| `color_test.go` | Color construction, type discrimination, ANSI approximation, equality |
| `style_test.go` | Style construction, attribute bitfield, fluent API chaining |
| `cell_test.go` | Width detection, continuation cells, equality |
| `rect_test.go` | Geometry operations, intersection, clamping edge cases |
| `buffer_test.go` | SetRune with wide chars, diff computation, resize preservation |
| `caps_test.go` | Capability detection for various TERM values |

### 7.2 Mock Terminal

```go
// MockTerminal captures output for testing without a real terminal.
type MockTerminal struct {
    width, height int
    cells         []Cell      // What was "rendered"
    cursor        struct{ x, y int }
    cursorHidden  bool
    inRawMode     bool
    inAltScreen   bool
    caps          Capabilities
}

func NewMockTerminal(width, height int) *MockTerminal
func (m *MockTerminal) Flush(changes []CellChange)
func (m *MockTerminal) CellAt(x, y int) Cell     // For test assertions
func (m *MockTerminal) String() string           // Render to string for snapshots
```

### 7.3 Integration Test

```go
// Example integration test (manual verification)
func TestRenderBorderedBox(t *testing.T) {
    buf := NewBuffer(40, 10)

    // Draw a box with title
    rect := Rect{X: 2, Y: 1, Width: 20, Height: 5}
    DrawBox(buf, rect, BorderSingle, "Title")

    // Draw some content inside
    buf.SetString(4, 3, "Hello, World!", NewStyle().Bold())

    // Verify with mock terminal
    mock := NewMockTerminal(40, 10)
    mock.Flush(buf.Diff())

    // Snapshot test
    expected := `
  ┌─Title─────────────┐
  │                   │
  │  Hello, World!    │
  │                   │
  └───────────────────┘
`
    if mock.String() != expected {
        t.Errorf("unexpected output:\n%s", mock.String())
    }
}
```

---

## 8. Performance Considerations

### 8.1 Buffer Operations

| Operation | Target | Notes |
|-----------|--------|-------|
| SetCell | O(1) | Direct index access |
| SetString | O(n) | n = string length, includes width detection |
| Diff | O(w×h) | Full scan; typically < 1ms for 200×60 |
| Clear | O(w×h) | Consider memset optimization if needed |

### 8.2 Memory Allocation

- **Buffer**: Two allocations per buffer (front + back), each `w×h×sizeof(Cell)`
- **For 200×60 buffer**: ~200KB total (assuming 16-byte Cell)
- **Diff slice**: Pre-allocated to typical change size, may grow

### 8.3 Terminal I/O

- **Batched writes**: All escape sequences buffered before single write syscall
- **Style tracking**: Only emit SGR codes when style changes
- **Cursor optimization**: Skip cursor moves for sequential cells

---

## 9. Complexity Assessment

| Size | Phases | When to Use |
|------|--------|-------------|
| Small | 1-2 | Single component, bug fix, minor enhancement |
| Medium | 3-4 | New feature touching multiple files/components |
| **Large** | **5-6** | **Cross-cutting feature, new subsystem** |

**Assessed Size:** Large
**Recommended Phases:** 5
**Rationale:** Phase 1 establishes a complete rendering subsystem with multiple interconnected components (color, style, cell, rect, buffer, terminal, capabilities). While each component is well-defined, the careful handling of wide characters, double buffering, terminal capabilities, and platform-specific raw mode requires methodical implementation and thorough testing. Five phases allow for incremental validation without being excessive.

**Proposed Phase Breakdown:**
1. **Core Types** (Color, Style, Attr) — Foundational types with tests
2. **Geometry & Cell** (Rect, Cell, RuneWidth) — Position and character handling
3. **Buffer** — Double-buffered grid with wide char support
4. **Terminal Core** — Interface, ANSI implementation, escape sequences
5. **Capabilities & Integration** — Detection, fallback, integration tests

> **IMPORTANT:** Please approve the overall design approach and complexity assessment before proceeding to the implementation plan.

---

## 10. Success Criteria

1. All core types (Color, Style, Cell, Rect) have comprehensive unit tests passing
2. Buffer correctly handles wide characters across all edge cases
3. Diff computation produces minimal change set (verified by tests)
4. Terminal abstraction works on macOS, Linux (Windows deferred to later phase)
5. Capability detection correctly identifies color support for common terminals
6. Integration test renders a bordered box with styled text
7. Mock terminal enables testing without real TTY

---

## 11. Open Questions

1. **Should Buffer track dirty rectangles for partial diff?**
   → Deferred. Full diff is fast enough for typical terminal sizes. Can optimize later if profiling shows need.

2. **Should we support bracketed paste mode?**
   → Out of scope for Phase 1 (input handling is Phase 4).

3. **Windows Console API support?**
   → Deferred. Phase 1 targets Unix terminals. Windows support can be added as a separate terminal implementation later.

4. **Should Color support CSS named colors (e.g., "coral")?**
   → Deferred to Phase 6 (polish). The current API supports it via `HexColor` or future `NamedColor` function.
