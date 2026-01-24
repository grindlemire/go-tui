# Phase 1: Terminal Rendering Foundation — Implementation Plan

Implementation phases for the terminal rendering foundation. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Core Types (Color, Style, Attr)

**Reference:** [phase1-rendering-design.md §3.1-3.2](./phase1-rendering-design.md#31-color)

**Review:** false

**Completed in commit:** phase1-core-types

- [x] Create `pkg/tui/color.go`
  - Define `ColorType` enum (ColorDefault, ColorANSI, ColorRGB)
  - Define `Color` struct with typ, r, g, b fields
  - Implement constructors: `DefaultColor()`, `ANSIColor(uint8)`, `RGBColor(r,g,b uint8)`
  - Implement `HexColor(string) (Color, error)` for parsing "#RRGGBB" and "#RGB"
  - Implement methods: `Type()`, `IsDefault()`, `ANSI()`, `RGB()`, `Equal()`
  - Implement `ToANSI()` color approximation using 6x6x6 cube and grayscale ramp
  - Define package-level color constants (Black, Red, Green, Yellow, Blue, Magenta, Cyan, White, and bright variants)

- [x] Create `pkg/tui/color_test.go`
  - Test all constructors produce correct ColorType
  - Test `HexColor` parsing (valid 6-digit, valid 3-digit, invalid inputs)
  - Test `Equal()` for same/different colors across all types
  - Test `ToANSI()` approximation for pure colors, grays, and mixed RGB values
  - Test panic behavior for `ANSI()` on RGB color and `RGB()` on ANSI color

- [x] Create `pkg/tui/style.go`
  - Define `Attr` bitfield type with constants (AttrBold, AttrDim, AttrItalic, AttrUnderline, AttrBlink, AttrReverse, AttrStrikethrough)
  - Define `Style` struct with Fg, Bg Color and Attrs field
  - Implement `NewStyle()` constructor returning zero-value Style
  - Implement fluent modifiers: `Foreground(Color)`, `Background(Color)`, `Bold()`, `Dim()`, `Italic()`, `Underline()`, `Reverse()`, `Strikethrough()`
  - Implement `Equal(Style)` and `HasAttr(Attr)` methods

- [x] Create `pkg/tui/style_test.go`
  - Test fluent API chaining produces correct Style
  - Test `Equal()` for identical and differing styles
  - Test `HasAttr()` for single and combined attributes
  - Test zero-value Style has no attributes and default colors

**Tests:** `go test ./pkg/tui/... -run "Color|Style"` — expect 15+ test cases passing (22 tests passing)

---

## Phase 2: Geometry & Cell (Rect, Cell, RuneWidth)

**Reference:** [phase1-rendering-design.md §3.3-3.4](./phase1-rendering-design.md#33-cell)

**Review:** false

**Completed in commit:** phase2-geometry-cell

- [x] Create `pkg/tui/rect.go`
  - Define `Rect` struct with X, Y, Width, Height int fields
  - Implement `NewRect(x, y, width, height int) Rect` constructor
  - Implement geometry methods: `Right()`, `Bottom()`, `Area()`, `IsEmpty()`, `Contains(x, y int)`, `ContainsRect(Rect)`
  - Implement transformations: `Inset(top, right, bottom, left int)`, `InsetUniform(n int)`, `Intersect(Rect)`, `Union(Rect)`, `Translate(dx, dy int)`
  - Implement `Clamp(x, y int) (int, int)` to constrain point to rect bounds

- [x] Create `pkg/tui/rect_test.go`
  - Test `Right()` and `Bottom()` calculations
  - Test `Contains()` for points inside, outside, and on boundaries
  - Test `ContainsRect()` for fully contained, partial overlap, and disjoint rects
  - Test `Inset()` with positive and negative values
  - Test `Intersect()` for overlapping, adjacent, and disjoint rects
  - Test `Union()` produces bounding rect
  - Test `Clamp()` for points inside and outside rect

- [x] Create `pkg/tui/cell.go`
  - Define `Cell` struct with Rune, Style, and Width uint8 fields
  - Implement `NewCell(r rune, style Style) Cell` with automatic width detection
  - Implement `NewCellWithWidth(r rune, style Style, width uint8) Cell`
  - Implement methods: `IsContinuation()`, `Equal(Cell)`, `IsEmpty()`
  - Implement `RuneWidth(r rune) int` function using Unicode East Asian Width property
    - Handle ASCII (width 1), CJK Unified Ideographs (width 2), common emoji (width 2)
    - Use simple range checks for common cases, not full Unicode tables

- [x] Create `pkg/tui/cell_test.go`
  - Test `NewCell` auto-detects width for ASCII, CJK, and emoji
  - Test `IsContinuation()` returns true only for Width=0 cells
  - Test `Equal()` compares all fields
  - Test `IsEmpty()` for space with default style vs other cells
  - Test `RuneWidth()` for representative characters:
    - ASCII letters/numbers → 1
    - CJK characters (你, 好, 中) → 2
    - Common emoji (😀, 🎉) → 2
    - Box drawing characters (─, │, ┌) → 1

**Tests:** `go test ./pkg/tui/... -run "Rect|Cell|RuneWidth"` — expect 20+ test cases passing (107 tests passing)

---

## Phase 3: Buffer (Double-Buffered Grid)

**Reference:** [phase1-rendering-design.md §4](./phase1-rendering-design.md#4-buffer-design)

**Review:** false

**Completed in commit:** phase3-buffer

- [x] Create `pkg/tui/buffer.go`
  - Define `Buffer` struct with front, back []Cell slices and width, height int
  - Implement `NewBuffer(width, height int) *Buffer` allocating both buffers
  - Implement dimension methods: `Width()`, `Height()`, `Size()`, `Rect()`
  - Implement index helper: `idx(x, y int) int` returning y*width+x with bounds check
  - Implement `Cell(x, y int) Cell` returning cell from back buffer (or empty Cell if out of bounds)
  - Implement `SetCell(x, y int, c Cell)` writing to back buffer with bounds check

- [x] Implement `SetRune` with wide character handling
  - Detect rune width using `RuneWidth()`
  - If target position is a continuation cell, clear the originating wide char
  - If placing wide char would overlap existing wide char, clear the existing one
  - Set primary cell with correct width
  - Set continuation cell (Width=0, Rune=0) for wide characters
  - Handle edge case: wide char at last column (truncate or skip)

- [x] Implement drawing primitives
  - `SetString(x, y int, s string, style Style) int` — iterate runes, handle widths, return total width consumed
  - `Fill(rect Rect, r rune, style Style)` — fill rectangle with single rune
  - `Clear()` — fill entire buffer with space and default style
  - `ClearRect(rect Rect)` — clear specific region

- [x] Implement diff and swap operations
  - Define `CellChange` struct with X, Y int and Cell
  - Implement `Diff() []CellChange` comparing back to front, returning changes in row-major order
  - Implement `Swap()` that copies back buffer to front buffer
  - Implement `Resize(width, height int)` preserving content where dimensions overlap

- [x] Create `pkg/tui/buffer_test.go`
  - Test `NewBuffer` creates correctly sized buffers
  - Test `SetCell` and `Cell` for in-bounds and out-of-bounds positions
  - Test `SetRune` with ASCII character (width 1)
  - Test `SetRune` with CJK character creates continuation cell
  - Test `SetRune` overwriting continuation cell clears original wide char
  - Test `SetRune` overwriting first cell of wide char clears continuation
  - Test `SetString` with mixed ASCII and CJK returns correct width
  - Test `SetString` truncation at buffer edge
  - Test `Fill` fills only specified rect
  - Test `Clear` resets all cells to space with default style
  - Test `Diff` returns empty slice when buffers match
  - Test `Diff` returns changed cells in row-major order
  - Test `Swap` makes subsequent Diff return empty
  - Test `Resize` preserves content in overlapping region

**Tests:** `go test ./pkg/tui/... -run "Buffer"` — expect 25+ test cases passing (44 tests passing)

---

## Phase 4: Terminal Core (Interface & ANSI Implementation)

**Reference:** [phase1-rendering-design.md §5](./phase1-rendering-design.md#5-terminal-abstraction)

**Completed in commit:** (pending)

- [ ] Create `pkg/tui/terminal.go`
  - Define `Terminal` interface with methods:
    - `Size() (width, height int)`
    - `Flush(changes []CellChange)`
    - `Clear()`
    - `SetCursor(x, y int)`
    - `HideCursor()`
    - `ShowCursor()`
    - `EnterRawMode() error`
    - `ExitRawMode() error`
    - `EnterAltScreen()`
    - `ExitAltScreen()`
    - `Caps() Capabilities`
  - Define `Capabilities` struct (Colors ColorCapability, Unicode bool, TrueColor bool, AltScreen bool)
  - Define `ColorCapability` enum (ColorNone, Color16, Color256, ColorTrue)

- [ ] Create `pkg/tui/escape.go` (internal escape sequence builder)
  - Define `escBuilder` struct with buf []byte
  - Implement `Reset()` clearing the buffer
  - Implement cursor methods: `MoveTo(x, y int)`, `MoveUp(n)`, `MoveDown(n)`, `MoveRight(n)`, `MoveLeft(n)`
  - Implement screen methods: `ClearScreen()`, `ClearLine()`
  - Implement cursor visibility: `HideCursor()`, `ShowCursor()`
  - Implement alt screen: `EnterAltScreen()`, `ExitAltScreen()`
  - Implement style methods: `ResetStyle()`, `SetStyle(s Style, caps Capabilities)`
    - Handle foreground color (ANSI 256 vs true color based on caps)
    - Handle background color
    - Handle attributes (bold, dim, italic, underline, reverse, strikethrough)
  - Implement `WriteRune(r rune)` appending UTF-8 encoded rune
  - Implement `Bytes() []byte` returning built sequence

- [ ] Create `pkg/tui/escape_test.go`
  - Test `MoveTo` generates correct `\x1b[{row};{col}H` (1-indexed)
  - Test `ClearScreen` generates `\x1b[2J`
  - Test `HideCursor` generates `\x1b[?25l`
  - Test `SetStyle` with bold generates `\x1b[1m`
  - Test `SetStyle` with ANSI foreground generates `\x1b[38;5;{n}m`
  - Test `SetStyle` with RGB foreground generates `\x1b[38;2;{r};{g};{b}m`
  - Test `SetStyle` respects Capabilities (falls back to ANSI when no TrueColor)
  - Test combined style generates minimal sequence

- [ ] Create `pkg/tui/terminal_ansi.go`
  - Define `ANSITerminal` struct with out io.Writer, in io.Reader, caps Capabilities, lastStyle Style, buf bytes.Buffer
  - Implement `NewANSITerminal(out io.Writer, in io.Reader) (*ANSITerminal, error)` with capability detection
  - Implement `NewANSITerminalWithCaps(out, in, caps)` for explicit capability override
  - Implement `Size()` using terminal size query (TIOCGWINSZ on Unix)
  - Implement `Flush(changes []CellChange)` with optimized cursor movement and style tracking
  - Implement `Clear()`, `SetCursor()`, `HideCursor()`, `ShowCursor()`
  - Implement `EnterAltScreen()`, `ExitAltScreen()`
  - Implement `Caps()` returning stored capabilities
  - Stub `EnterRawMode()` and `ExitRawMode()` (platform-specific, implemented next)

- [ ] Create `pkg/tui/terminal_unix.go` (build tag: `//go:build unix`)
  - Import `golang.org/x/term` or use syscall directly
  - Define `rawModeState` storing original termios
  - Implement `EnterRawMode()` saving state and setting raw mode
  - Implement `ExitRawMode()` restoring original state
  - Implement `getTerminalSize(fd int) (width, height int, err error)`

- [ ] Create `pkg/tui/mock_terminal.go`
  - Define `MockTerminal` struct capturing all operations
  - Implement full `Terminal` interface
  - Add `CellAt(x, y int) Cell` for test assertions
  - Add `String() string` rendering buffer to string for snapshot tests
  - Add `Cursor() (x, y int)` returning cursor position
  - Add `IsCursorHidden() bool`, `IsInRawMode() bool`, `IsInAltScreen() bool`

- [ ] Create `pkg/tui/terminal_test.go`
  - Test `MockTerminal` implements Terminal interface
  - Test `Flush` to MockTerminal updates correct cells
  - Test cursor tracking after `SetCursor` and `Flush`
  - Test `Clear` resets all MockTerminal cells
  - Test `EnterAltScreen`/`ExitAltScreen` state tracking

**Tests:** `go test ./pkg/tui/... -run "Terminal|Escape|Mock"` — expect 20+ test cases passing

---

## Phase 5: Capabilities & Integration

**Reference:** [phase1-rendering-design.md §6, §7](./phase1-rendering-design.md#6-capability-detection)

**Completed in commit:** (pending)

- [ ] Create `pkg/tui/caps.go`
  - Implement `DetectCapabilities() Capabilities`
    - Check `TERM` env var for "256color", "truecolor", "dumb"
    - Check `COLORTERM` for "truecolor" or "24bit"
    - Check terminal-specific vars: `WT_SESSION` (Windows Terminal), `ITERM_SESSION_ID` (iTerm2), `KITTY_WINDOW_ID` (Kitty)
    - Return conservative defaults when detection fails
  - Implement `(c Capabilities) SupportsColor(color Color) bool` helper
  - Implement `(c Capabilities) EffectiveColor(color Color) Color` returning original or fallback

- [ ] Create `pkg/tui/caps_test.go`
  - Test detection with `TERM=xterm-256color` → Color256
  - Test detection with `COLORTERM=truecolor` → ColorTrue
  - Test detection with `TERM=dumb` → ColorNone, no Unicode
  - Test detection with `WT_SESSION` set → ColorTrue
  - Test `EffectiveColor` returns original when supported
  - Test `EffectiveColor` returns ANSI approximation when RGB not supported

- [ ] Create `pkg/tui/border.go`
  - Define `BorderStyle` type (BorderNone, BorderSingle, BorderDouble, BorderRounded, BorderThick)
  - Define `BorderChars` struct with TopLeft, Top, TopRight, Left, Right, BottomLeft, Bottom, BottomRight rune fields
  - Implement `(b BorderStyle) Chars() BorderChars` returning appropriate box-drawing characters
  - Implement `DrawBox(buf *Buffer, rect Rect, border BorderStyle, style Style)`
  - Implement `DrawBoxWithTitle(buf *Buffer, rect Rect, border BorderStyle, title string, style Style)`

- [ ] Create `pkg/tui/border_test.go`
  - Test `BorderSingle.Chars()` returns correct Unicode box-drawing chars
  - Test `DrawBox` renders corners and edges correctly
  - Test `DrawBoxWithTitle` centers title in top border
  - Test `DrawBox` with rect smaller than 2x2 handles gracefully

- [ ] Create `pkg/tui/render.go`
  - Implement `Render(term Terminal, buf *Buffer)` convenience function
    - Call `buf.Diff()` to get changes
    - Call `term.Flush(changes)` to write to terminal
    - Call `buf.Swap()` to update front buffer
  - Implement `RenderFull(term Terminal, buf *Buffer)` forcing full redraw
    - Useful after resize or initial render

- [ ] Create integration test `pkg/tui/integration_test.go`
  - Test full render pipeline: Buffer → Diff → MockTerminal → verify output
  - Test bordered box with title renders correctly (snapshot test)
  - Test styled text inside box renders correctly
  - Test wide characters in buffer render with correct spacing
  - Test resize and re-render produces correct diff
  - Test capability-based color fallback in full pipeline

- [ ] Update `go.mod` with `golang.org/x/term` dependency (if used)

**Tests:** `go test ./pkg/tui/...` — all tests passing, expect 80+ total test cases

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Core Types (Color, Style, Attr) | Complete |
| 2 | Geometry & Cell (Rect, Cell, RuneWidth) | Complete |
| 3 | Buffer (double-buffered grid with wide char support) | Complete |
| 4 | Terminal Core (interface, ANSI impl, escape sequences) | Pending |
| 5 | Capabilities & Integration (detection, borders, full pipeline) | Pending |

## Files to Create

```
pkg/tui/
├── border.go
├── border_test.go
├── buffer.go
├── buffer_test.go
├── caps.go
├── caps_test.go
├── cell.go
├── cell_test.go
├── color.go
├── color_test.go
├── escape.go
├── escape_test.go
├── integration_test.go
├── mock_terminal.go
├── rect.go
├── rect_test.go
├── render.go
├── style.go
├── style_test.go
├── terminal.go
├── terminal_ansi.go
├── terminal_test.go
└── terminal_unix.go
```

## Files to Modify

| File | Changes |
|------|---------|
| `go.mod` | Add `golang.org/x/term` dependency (if needed for raw mode) |
