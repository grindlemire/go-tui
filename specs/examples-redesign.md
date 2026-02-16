# Examples Redesign Spec

## Overview

Redesign the go-tui examples into a clean 14-example progressive tutorial. Phases 1-14 each create one example; Phase 0 is cleanup and Phase 15 is final verification.

Run **Phase 0: Cleanup Old Examples** first. After Phase 0 is complete, Phases 1-14 are independent and can be executed in any order by different agents.

## Framework Context (for all phases)

### .gsx File Structure

Every `.gsx` file follows this pattern:

```gsx
package examplename

import (
    "fmt"
    "time"
    tui "github.com/grindlemire/go-tui"
)

// Struct component (stateful)
type myApp struct {
    someState *tui.State[int]
    someRef   *tui.Ref
}

// Constructor
func MyApp() *myApp {
    return &myApp{
        someState: tui.NewState(0),
        someRef:   tui.NewRef(),
    }
}

// Render method — generates a *tui.Element
templ (a *myApp) Render() {
    <div class="flex-col gap-1 p-1">
        <span class="font-bold text-cyan">{fmt.Sprintf("Count: %d", a.someState.Get())}</span>
    </div>
}

// KeyMap — returns key bindings
func (a *myApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnRune('+', func(ke tui.KeyEvent) { a.someState.Update(func(v int) int { return v + 1 }) }),
    }
}

// HandleMouse — handles mouse events using refs
func (a *myApp) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(a.someRef, a.doSomething),
    )
}

// Watchers — returns background watchers
func (a *myApp) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(time.Second, a.tick),
    }
}

// Pure templ component (no state, just params)
templ Card(title string) {
    <div class="border-rounded p-1">
        <span class="font-bold">{title}</span>
        {children...}
    </div>
}
```

### main.go Structure

Every example's `main.go` follows this pattern:

```go
package main

import (
    "fmt"
    "os"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    app, err := tui.NewApp(
        tui.WithRootComponent(MyApp()),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Available Tailwind Classes

**Layout**: `flex`, `flex-col`, `flex-row`, `gap-N`, `grow`, `shrink`, `flex-1`, `flex-none`, `flex-grow-N`, `flex-shrink-N`
**Justify**: `justify-start`, `justify-center`, `justify-end`, `justify-between`, `justify-around`, `justify-evenly`
**Align**: `items-start`, `items-center`, `items-end`, `items-stretch`, `self-start`, `self-center`, `self-end`, `self-stretch`
**Text align**: `text-left`, `text-center`, `text-right`
**Spacing**: `p-N`, `px-N`, `py-N`, `pt-N`, `pr-N`, `pb-N`, `pl-N`, `m-N`, `mx-N`, `my-N`, `mt-N`, `mr-N`, `mb-N`, `ml-N`
**Sizing**: `w-N`, `h-N`, `w-full`, `h-full`, `w-auto`, `h-auto`, `w-1/2`, `w-1/3`, `w-2/3`, `min-w-N`, `max-w-N`, `min-h-N`, `max-h-N`
**Borders**: `border`, `border-single`, `border-double`, `border-rounded`, `border-thick`, `border-{color}` (e.g. `border-cyan`)
**Text styles**: `font-bold`, `font-dim`, `text-dim`, `italic`, `underline`, `strikethrough`, `reverse`
**Text colors**: `text-red`, `text-green`, `text-blue`, `text-cyan`, `text-magenta`, `text-yellow`, `text-white`, `text-black`, `text-bright-red`, `text-bright-green`, `text-bright-blue`, `text-bright-cyan`, `text-bright-magenta`, `text-bright-yellow`, `text-bright-white`, `text-bright-black`
**Background colors**: `bg-red`, `bg-green`, `bg-blue`, `bg-cyan`, `bg-magenta`, `bg-yellow`, `bg-white`, `bg-black`, `bg-bright-red`, `bg-bright-green`, `bg-bright-blue`, `bg-bright-cyan`, `bg-bright-magenta`, `bg-bright-yellow`, `bg-bright-white`, `bg-bright-black`
**Gradients**: `text-gradient-{c1}-{c2}[-direction]`, `bg-gradient-{c1}-{c2}[-direction]`, `border-gradient-{c1}-{c2}[-direction]` where direction is `-h` (horizontal), `-v` (vertical), `-dd` (diagonal down), `-du` (diagonal up)
**Scroll**: `overflow-scroll`, `overflow-y-scroll`, `overflow-x-scroll`

### Available Elements

`<div>` (block container), `<span>` (inline text), `<p>` (paragraph), `<ul>` (unordered list), `<li>` (list item), `<button>` (clickable), `<input>` (text input, self-closing), `<table>` (table), `<progress>` (progress bar, self-closing), `<hr>` (horizontal rule, self-closing), `<br>` (line break, self-closing)

### Element Attributes

**Common**: `id`, `class`, `disabled`, `ref`, `deps`
**Layout**: `width`, `widthPercent`, `height`, `heightPercent`, `minWidth`, `minHeight`, `maxWidth`, `maxHeight`, `direction`, `justify`, `align`, `gap`, `flexGrow`, `flexShrink`, `alignSelf`, `padding`, `margin`
**Visual**: `border`, `borderStyle`, `background`, `text`, `textStyle`, `textAlign`
**Scroll**: `scrollable`, `scrollOffset`, `scrollbarStyle`, `scrollbarThumbStyle`
**Input-specific**: `value`, `placeholder`
**Progress-specific**: `value` (current), `max` (maximum)
**Focus**: `focusable`, `onFocus`, `onBlur`

### Key Types

```go
// Dimensions
tui.Fixed(10), tui.Percent(50), tui.Auto()

// Borders
tui.BorderNone, tui.BorderSingle, tui.BorderDouble, tui.BorderRounded, tui.BorderThick

// Direction
tui.Row, tui.Column

// Justify
tui.JustifyStart, tui.JustifyCenter, tui.JustifyEnd, tui.JustifySpaceBetween, tui.JustifySpaceAround, tui.JustifySpaceEvenly

// Align
tui.AlignStart, tui.AlignCenter, tui.AlignEnd, tui.AlignStretch

// Scroll
tui.ScrollNone, tui.ScrollVertical, tui.ScrollHorizontal, tui.ScrollBoth

// Style construction
tui.NewStyle().Foreground(tui.ANSIColor(tui.Cyan)).Bold()

// Colors
tui.Black, tui.Red, tui.Green, tui.Yellow, tui.Blue, tui.Magenta, tui.Cyan, tui.White
tui.BrightBlack, tui.BrightRed, tui.BrightGreen, tui.BrightYellow, tui.BrightBlue, tui.BrightMagenta, tui.BrightCyan, tui.BrightWhite
tui.ANSIColor(index), tui.RGBColor(r,g,b), tui.HexColor("#RRGGBB")

// State
state := tui.NewState(initialValue)
state.Get(), state.Set(v), state.Update(func(v T) T)

// Watchers
tui.OnTimer(duration, handler)
tui.Watch(channel, handler)

// Events
events := tui.NewEvents[string]()

// KeyMap entries
tui.OnKey(tui.KeyEscape, handler)
tui.OnRune('q', handler)
tui.OnRunes(handler)           // catch-all for any rune
tui.OnRuneStop('x', handler)   // stops propagation
tui.OnKeyStop(tui.KeyEnter, handler)

// Mouse
tui.HandleClicks(mouseEvent, tui.Click(ref, handler), ...)

// Refs
ref := tui.NewRef()      // single element ref
ref.El()                 // get the *Element (may be nil)
refs := tui.NewRefList() // slice of elements from @for loops

// App options
tui.WithRootComponent(component)
tui.WithRootView(viewable)
tui.WithInlineHeight(rows)
tui.WithFrameRate(fps)
tui.WithMouse()
tui.WithoutMouse()
```

### Control Flow in .gsx

```gsx
// Conditionals
@if condition {
    <span>Shown when true</span>
} @else {
    <span>Shown when false</span>
}

// Loops
@for i, item := range items {
    <span>{item}</span>
}

// Let bindings
@let label = fmt.Sprintf("Count: %d", count)
<span>{label}</span>
```

### Code Generation

After writing `.gsx` files, run `go run ./cmd/tui generate ./examples/...` from the project root to produce `_gsx.go` files. The generated files should NOT be hand-edited.

### Design Principles (apply to ALL examples)

- **Rounded borders** (`border-rounded`) as default style
- **Consistent palette**: cyan=primary, magenta=accent, yellow=warning, green=success, red=error
- **Gradient titles** on section headers using `text-gradient-cyan-magenta` or similar
- **Key hint bar** at bottom of every interactive example in `font-dim`
- **Generous spacing**: `p-1` minimum on cards, `gap-1` between sections
- **No cramped layouts** — whitespace over density
- Root container should use `h-full` (or explicit `heightPercent={100}`) for full-screen examples

### Commit Convention

Use `gcommit -m "message"` for all commits.

---

## Phase 0: Cleanup Old Examples

**Goal**: Remove redundant and low-level examples that will be replaced.

**Delete these directories**:
- `examples/counter-state/` — redundant counter
- `examples/dsl-counter/` — redundant counter
- `examples/hello_layout/` — low-level API demo, not .gsx
- `examples/hello_rect/` — low-level API demo, not .gsx
- `examples/focus/` — covered by 12-multi-component
- `examples/scrollable/` — non-DSL scrollable, covered by 09-scrolling
- `examples/streaming/` — non-DSL streaming, covered by 11-streaming
- `examples/streaming-dsl/` — covered by 11-streaming
- `examples/inline-test/` — covered by 11-streaming
- `examples/refs-demo/` — covered by 07-refs-and-clicks
- `examples/claude-chat/` — already deleted

**Delete these numbered dirs** (will be recreated with new content):
- `examples/00-hello/`
- `examples/01-styling/`
- `examples/02-layout/`
- `examples/03-conditionals/`
- `examples/04-loops/`
- `examples/05-composition/`
- `examples/06-interactive/`
- `examples/07-keyboard/`
- `examples/09-scrollable/`
- `examples/10-refs/`
- `examples/11-streaming/`

**Delete these that get renamed/merged**:
- `examples/component-model/`
- `examples/dashboard/`
- `examples/ai-chat/` — will be recreated as `examples/14-ai-chat/`

**Verification**: `ls examples/` should show an empty directory (all old examples removed).

---

## Phase 1: 01-hello

**Goal**: Absolute minimum go-tui app. A beautiful centered welcome screen.

**Directory**: `examples/01-hello/`
**Files to create**: `hello.gsx`, `main.go`
**Package name**: `main`

### Visual Design

```
╭──────────────────────────────────────╮
│                                      │
│        Welcome to go-tui             │  ← gradient title, font-bold
│                                      │
│   Build beautiful terminal UIs       │  ← text-cyan
│   with Go and flexbox                │  ← text-cyan
│                                      │
│          Press q to quit             │  ← font-dim
│                                      │
╰──────────────────────────────────────╯
```

### Requirements

- `helloApp` struct with no fields
- Constructor `HelloApp() *helloApp` that returns `&helloApp{}`
- `Render` templ: root div `flex-col items-center justify-center h-full`, inner card with `border-rounded border-cyan p-2 gap-1 items-center w-44`
- Title span: `text-gradient-cyan-magenta font-bold` with text "Welcome to go-tui"
- Two subtitle spans: `text-cyan` with "Build beautiful terminal UIs" and "with Go and flexbox"
- Hint span: `font-dim` with "Press q to quit"
- `KeyMap()`: `tui.OnRune('q', ...)` and `tui.OnKey(tui.KeyEscape, ...)` both call `ke.App().Stop()`
- `main.go`: `tui.NewApp(tui.WithRootComponent(HelloApp()))`, run, handle errors

### Verification

```bash
cd /Users/joelholsteen/go/src/github.com/grindlemire/go-tui
go run ./cmd/tui generate ./examples/01-hello/...
go build ./examples/01-hello/
go run ./examples/01-hello/   # visual check: centered card, q quits
```

---

## Phase 2: 02-styling

**Goal**: Showcase all styling capabilities — colors, text styles, borders, gradients. Uses scrolling as infrastructure to display content, but the taught concept is styling only.

**Directory**: `examples/02-styling/`
**Files to create**: `styling.gsx`, `main.go`
**Package name**: `main`

### Visual Design

A vertically scrollable gallery with sections. Each section has a bordered header and content:

1. **Text Colors** — Two rows: standard 8 colors, then bright 8 colors. Each color name rendered in that color.
2. **Background Colors** — Two rows of colored blocks with labels.
3. **Text Styles** — Row showing: Bold, Dim, Italic, Underline, Strikethrough, Reverse — each styled accordingly.
4. **Border Styles** — Four small boxes side by side: Single, Double, Rounded, Thick. Each with colored borders.
5. **Text Gradients** — Four gradient text samples: horizontal cyan→magenta, vertical green→yellow, diagonal-down red→blue, diagonal-up magenta→cyan.
6. **Background Gradients** — Four gradient background blocks in different directions.
7. **Border Gradients** — Two boxes with border gradients.
8. **Combined Styles** — A few examples combining bg + fg + text style (e.g. bold cyan on magenta background).

### Requirements

- `stylingApp` struct with `scrollY *tui.State[int]` and `content *tui.Ref`
- Constructor initializes state and ref
- `Render` templ: root div with `flex-col h-full`, scrollable content area with `scrollable={tui.ScrollVertical}` and `scrollOffset={0, s.scrollY.Get()}` and `ref={s.content}`
- Use reusable `templ Section(title string)` component: `border-rounded border-cyan p-1 gap-1` with title span in `font-bold text-gradient-cyan-magenta`, then `{children...}`
- Color rows: use `@for` over slices of color names, render each as a span with appropriate class
- Border boxes: four inline divs with different border classes
- Gradient texts: spans with `text-gradient-X-Y-dir` classes
- `KeyMap()`: j/k and arrow keys for scrolling, q/Escape to quit
- `HandleMouse()`: mouse wheel for scrolling
- Helper method `scrollBy(delta int)` that clamps scroll position using `s.content.El().MaxScroll()`
- Key hint bar at the bottom (outside scroll area): `font-dim` showing "j/k scroll  q quit"

### Verification

```bash
go run ./cmd/tui generate ./examples/02-styling/...
go build ./examples/02-styling/
go run ./examples/02-styling/   # visual: scroll through all style sections
```

---

## Phase 3: 03-layout

**Goal**: Demonstrate all flexbox layout concepts with clear, labeled visual examples.

**Directory**: `examples/03-layout/`
**Files to create**: `layout.gsx`, `main.go`
**Package name**: `main`

### Visual Design

Scrollable gallery with sections, each demonstrating a layout concept:

1. **Direction** — Row vs Column: colored blocks [A] [B] [C] arranged horizontally vs vertically, side by side.
2. **Justify Content** — Six rows (start, center, end, between, around, evenly), each with 3 colored blocks showing the spacing.
3. **Align Items** — Four columns (start, center, end, stretch) with blocks of different heights.
4. **Gap** — Three examples with gap-0, gap-1, gap-2 showing the spacing difference.
5. **Flex Grow** — Three blocks: first fixed-width, second with `grow`, third fixed, showing flexible fill.
6. **Sizing** — Fixed width (`w-20`), percentage (`w-1/2`), auto width side by side.
7. **Padding & Margin** — Boxes with visible padding (bg color shows padding area) and margin (gap between border and neighbors).

### Requirements

- Same scroll infrastructure as 02-styling (`scrollY` state, `content` ref, j/k/arrows/mouse-wheel)
- Reuse the `Section(title string)` component pattern with `{children...}`
- Helper `templ Block(label string, bgClass string)` — small colored block with a letter label, fixed height 3, padding 1, specific bg color class
- For justify section: each row is a div with the justify class applied, containing 3 Block children. Label each row with the justify name.
- For align section: use items of varying height (h-3, h-5, h-7) to make alignment visible.
- Key hint bar at bottom: `font-dim` showing "j/k scroll  q quit"

### Verification

```bash
go run ./cmd/tui generate ./examples/03-layout/...
go build ./examples/03-layout/
go run ./examples/03-layout/   # visual: scroll through layout demos, blocks positioned correctly
```

---

## Phase 4: 04-components

**Goal**: Show reusable templ components with params and `{children...}`, component nesting, and composition patterns.

**Directory**: `examples/04-components/`
**Files to create**: `components.gsx`, `main.go`
**Package name**: `main`

### Visual Design

```
╭─────────────────────────────────────────────────────╮
│            Component Showcase                        │  ← gradient title
╰─────────────────────────────────────────────────────╯

╭─ Alice ─────────────╮  ╭─ Bob ───────────────────╮
│  Role: Engineer      │  │  Role: Designer          │
│  Level: Senior       │  │  Level: Junior           │
│  ● Online            │  │  ○ Offline               │
╰──────────────────────╯  ╰──────────────────────────╯

╭─ Charlie ───────────╮  ╭─ Diana ─────────────────╮
│  Role: Manager       │  │  Role: DevOps            │
│  Level: Lead         │  │  Level: Mid              │
│  ● Online            │  │  ● Online                │
╰──────────────────────╯  ╰──────────────────────────╯

╭─ Status ────────────────────────────────────────────╮
│  ✓ Build   │  3 Warnings   │  v1.2.0               │
╰─────────────────────────────────────────────────────╯

                    Press q to quit
```

### Requirements

- `componentsApp` struct with no fields, just quit key
- Reusable templ components (each a pure `templ` function, NOT struct):
  - `templ Card(title string)` — bordered card with title, `{children...}` for body. Uses `border-rounded border-cyan`, title in `font-bold text-cyan`
  - `templ Badge(label string, colorClass string)` — inline span with the given color class and text
  - `templ UserCard(name, role, level string, online bool)` — calls `Card(name)` with role, level, and online status badge inside. Online=`text-green ●`, Offline=`text-bright-black ○`
  - `templ StatusBar()` — horizontal bar with badges for build status, warnings, version. Uses `border-rounded`, `justify-between`
  - `templ Header(title string)` — full-width `text-gradient-cyan-magenta font-bold text-center p-1 border-rounded border-cyan`
- `Render` templ: column layout with Header, row of UserCards (2x2 grid using two rows), StatusBar, quit hint
- Static display — no state needed (only quit KeyMap)

### Verification

```bash
go run ./cmd/tui generate ./examples/04-components/...
go build ./examples/04-components/
go run ./examples/04-components/   # visual: cards render correctly, composition works
```

---

## Phase 5: 05-state

**Goal**: Introduce reactive state, `@if/@else`, `@for`, `@let`, and KeyMap interaction. Merges concepts from old 03-conditionals and 04-loops.

**Directory**: `examples/05-state/`
**Files to create**: `state.gsx`, `main.go`
**Package name**: `main`

### Visual Design

```
╭─ Counter ─────────╮  ╭─ Status ──────────────────────────╮
│                    │  │                                    │
│       42           │  │  ✓ Positive                       │
│                    │  │  Value is even                     │
│   + increment      │  │  Range: high (> 20)               │
│   - decrement      │  │                                    │
│   r reset          │  │                                    │
╰────────────────────╯  ╰──────────────────────────────────╯
╭─ Favorite Languages ──────────────────────────────────────╮
│   Rust                                                     │
│ ▶ Go              ← selected                              │
│   TypeScript                                               │
│   Python                                                   │
│   Zig                                                      │
╰───────────────────────────────────────────────────────────╯

      +/- count   j/k navigate   r reset   q quit
```

### Requirements

- `stateApp` struct:
  - `count *tui.State[int]`
  - `selected *tui.State[int]`
  - `items []string` (fixed list: "Rust", "Go", "TypeScript", "Python", "Zig")
- Constructor: initializes count=0, selected=0
- `Render` templ:
  - Top row: two panels side-by-side
    - Counter panel: `border-rounded border-cyan p-1`, shows count with `@let formatted = fmt.Sprintf(...)`, large text. Also shows key hints inside the panel.
    - Status panel: `border-rounded border-cyan p-1`, uses `@if`/`@else` chains:
      - `@if count > 0` → "✓ Positive" in `text-green`, `@else` → "✗ Negative" in `text-red`, or "Zero" in `text-yellow` (use `@if count > 0 { } @else { @if count < 0 { } @else { } }`)
      - `@if count % 2 == 0` → "Value is even"
      - `@if count > 20` → "Range: high", `@else if count > 0` → "Range: low", etc.
  - Bottom: Items panel with `@for i, item := range s.items`, using `@if i == s.selected.Get()` for highlight (▶ prefix, `text-cyan font-bold` vs normal)
  - Key hint bar at bottom
- `KeyMap()`:
  - `+` → increment count
  - `-` → decrement count
  - `r` → reset count to 0
  - `j` / `KeyDown` → next item (wrap around)
  - `k` / `KeyUp` → prev item (wrap around)
  - `q` / `Escape` → quit

### Verification

```bash
go run ./cmd/tui generate ./examples/05-state/...
go build ./examples/05-state/
go run ./examples/05-state/   # visual: +/- changes count, status reacts, j/k moves selection
```

---

## Phase 6: 06-keyboard

**Goal**: Comprehensive keyboard event handling — show every key type, catch-all handlers, and modifier detection.

**Directory**: `examples/06-keyboard/`
**Files to create**: `keyboard.gsx`, `main.go`
**Package name**: `main`

### Visual Design

```
╭─ Keyboard Explorer ──────────────────────────────────────╮
│                                                           │
│  Last Key:   Ctrl+A                                       │  ← large, gradient text
│  Key Count:  17                                           │
│                                                           │
│  ╭─ Special Keys ──────────────────────────────────────╮ │
│  │ Enter ✓  Tab ✓  Backspace ·  Delete ·               │ │
│  │ ↑ ✓  ↓ ✓  ← ·  → ·  Home ·  End ·                 │ │
│  │ PgUp ·  PgDn ·  Insert ·                            │ │
│  ╰──────────────────────────────────────────────────────╯ │
│                                                           │
│  Press any key to see it displayed · q to quit            │
╰──────────────────────────────────────────────────────────╯
```

### Requirements

- `keyboardApp` struct:
  - `lastKey *tui.State[string]`
  - `lastMod *tui.State[string]` (display `ke.Mod.String()`, default "None")
  - `keyCount *tui.State[int]`
  - Boolean states for tracking which special keys have been pressed: `pressedEnter`, `pressedTab`, `pressedBackspace`, `pressedDelete`, `pressedUp`, `pressedDown`, `pressedLeft`, `pressedRight`, `pressedHome`, `pressedEnd`, `pressedPgUp`, `pressedPgDn`, `pressedInsert` — all `*tui.State[bool]`
- `Render` templ:
  - Outer container: `border-rounded border-cyan p-2 flex-col gap-2 h-full`
  - Last key display: large text with `text-gradient-cyan-magenta font-bold`
  - Last modifier display (e.g. `Modifier: Ctrl`, `Modifier: None`) near key display
  - Key count: `text-cyan`
  - Special keys grid: `border-rounded p-1`, rows of key indicators. Each shows name + ✓ (text-green) if pressed or · (text-bright-black) if not. Use `@if k.pressedEnter.Get()` for each.
  - Hint at bottom: `font-dim`
- `KeyMap()`:
  - `OnRuneStop('q', ...)` → quit
  - `OnRunes(func(ke) { ... })` → catch-all, update lastKey with rune display, update `lastMod`, increment count
  - `OnKey(tui.KeyEnter, ...)` → set pressedEnter=true, update lastKey="Enter", update `lastMod`, increment count
  - Same pattern for Tab, Backspace, Delete, Up, Down, Left, Right, Home, End, PageUp, PageDown, Insert (always record `ke.Mod`)
  - `OnKey(tui.KeyEscape, ...)` → quit
- Helper: `recordKey(name string, mod tui.Modifier)` method that updates lastKey, `lastMod`, and increments keyCount

### Verification

```bash
go run ./cmd/tui generate ./examples/06-keyboard/...
go build ./examples/06-keyboard/
go run ./examples/06-keyboard/   # visual: press keys, see them light up in the grid
```

---

## Phase 7: 07-refs-and-clicks

**Goal**: Demonstrate refs for element access and mouse click handling. Interactive color mixer with clickable +/- buttons.

**Directory**: `examples/07-refs-and-clicks/`
**Files to create**: `clicks.gsx`, `main.go`
**Package name**: `main`

### Visual Design

```
╭─ Color Mixer ────────────────────────────────────────────╮
│                                                           │
│    ████████████████████████████████                       │  ← preview swatch
│    R: 128   G: 64    B: 200                              │
│                                                           │
│  ╭─ Red ───────╮  ╭─ Green ─────╮  ╭─ Blue ──────╮     │
│  │    [  +  ]   │  │    [  +  ]   │  │    [  +  ]   │     │
│  │     128      │  │      64     │  │     200      │     │
│  │    [  -  ]   │  │    [  -  ]   │  │    [  -  ]   │     │
│  ╰──────────────╯  ╰──────────────╯  ╰──────────────╯     │
│                                                           │
│  Click +/- or press r/g/b increase · R/G/B decrease · q quit │
╰──────────────────────────────────────────────────────────╯
```

### Requirements

- `colorMixer` struct:
  - `red`, `green`, `blue` — `*tui.State[int]` (0-255, start at 128, 64, 200)
  - `redUp`, `redDown`, `greenUp`, `greenDown`, `blueUp`, `blueDown` — `*tui.Ref` (6 button refs)
- Methods:
  - `adjustRed(delta int)`, `adjustGreen(delta int)`, `adjustBlue(delta int)` — clamp 0-255
  - `previewStyle() tui.Style` — returns style with `tui.RGBColor(r, g, b)` as background
  - `channelStyle(value int, baseColor int) tui.Style` — returns text style colored by channel intensity
- `Render` templ:
  - Outer: `border-rounded border-cyan p-2 flex-col gap-2 h-full items-center justify-center`
  - Preview swatch: div with `background={c.previewStyle()}`, `w-36 h-3`
  - RGB display: `@let` for formatted string
  - Three channel panels side-by-side (row), each: `border-rounded p-1 items-center gap-1 flex-col`
    - Channel border colored: Red panel gets `border-red`, Green gets `border-green`, Blue gets `border-blue`
    - Up button: `<button ref={c.redUp} class="border-rounded px-2">[  +  ]</button>` (as a div with ref, styled like a button)
    - Value display: `textStyle={c.channelStyle(...)}`
    - Down button: similar with ref
  - Hint bar at bottom
- `HandleMouse()`:
  ```
  tui.HandleClicks(me,
      tui.Click(c.redUp, func() { c.adjustRed(8) }),
      tui.Click(c.redDown, func() { c.adjustRed(-8) }),
      tui.Click(c.greenUp, func() { c.adjustGreen(8) }),
      tui.Click(c.greenDown, func() { c.adjustGreen(-8) }),
      tui.Click(c.blueUp, func() { c.adjustBlue(8) }),
      tui.Click(c.blueDown, func() { c.adjustBlue(-8) }),
  )
  ```
- `KeyMap()`: `r/g/b` → +8 for red/green/blue, `R/G/B` → -8, `q/Escape` → quit

### Verification

```bash
go run ./cmd/tui generate ./examples/07-refs-and-clicks/...
go build ./examples/07-refs-and-clicks/
go run ./examples/07-refs-and-clicks/   # visual: click buttons, press keys, see color change
```

---

## Phase 8: 08-elements

**Goal**: Showcase ALL built-in HTML-like elements — `<p>`, `<hr>`, `<br>`, `<ul>/<li>`, `<table>`, `<progress>`, `<input>`, `<button>`.

**Directory**: `examples/08-elements/`
**Files to create**: `elements.gsx`, `main.go`
**Package name**: `main`

### Visual Design

```
╭─ Built-in Elements ──────────────────────────────────────╮
│                                                           │
│  ╭─ Text ─────────────────────────────────────────────╮  │
│  │  Paragraph text wraps automatically when the        │  │
│  │  content exceeds the available width of the         │  │
│  │  container element.                                 │  │
│  │  ──────────────────────────────────────────────     │  │
│  │  Inline span with styling                           │  │
│  ╰─────────────────────────────────────────────────────╯  │
│                                                           │
│  ╭─ Lists ────────────╮  ╭─ Table ───────────────────╮   │
│  │  • Rust             │  │  Name    Role     Level   │   │
│  │  • Go               │  │  Alice   Eng      Senior  │   │
│  │  • TypeScript       │  │  Bob     Design   Junior  │   │
│  │  • Python           │  │  Carol   DevOps   Mid     │   │
│  ╰─────────────────────╯  ╰──────────────────────────╯   │
│                                                           │
│  ╭─ Progress ─────────────────────────────────────────╮  │
│  │  Build:    ████████████████░░░░  78%               │  │
│  │  Tests:    ████████████████████  100%              │  │
│  │  Deploy:   ████████░░░░░░░░░░░  38%               │  │
│  ╰─────────────────────────────────────────────────────╯  │
│                                                           │
│  +/- adjust build progress · q quit                       │
╰──────────────────────────────────────────────────────────╯
```

### Requirements

- `elementsApp` struct:
  - `buildProgress *tui.State[int]` (start at 78)
- Render templ uses scrollable container (content may be tall):
  - **Text section**: `<p>` element with multi-sentence text, `<br />` line break, `<hr />` separator, `<span>` with styling
  - **Lists section**: `<ul>` with `<li>` children listing programming languages
  - **Table section**: `<table>` with header row and data rows (use text content for cells)
  - **Controls section**: include one `<input />` (with `placeholder` or `value`) and one `<button>` element to demonstrate both tags
  - **Progress section**: Three `<progress>` bars:
    - Build: `value={e.buildProgress.Get()}` `max={100}`
    - Tests: `value={100}` `max={100}` (always full)
    - Deploy: `value={38}` `max={100}` (static)
  - Each section wrapped in a Card-like container with `border-rounded border-cyan p-1`
- `KeyMap()`: `+`/`-` to adjust buildProgress (clamp 0-100), `q`/`Escape` to quit

### Verification

```bash
go run ./cmd/tui generate ./examples/08-elements/...
go build ./examples/08-elements/
go run ./examples/08-elements/   # visual: all element types render, +/- changes progress
```

---

## Phase 9: 09-scrolling

**Goal**: Dedicated deep-dive into scroll mechanics — scrollable containers, keyboard/mouse scroll, bounds checking, position tracking, scrollbar styling.

**Directory**: `examples/09-scrolling/`
**Files to create**: `scrolling.gsx`, `main.go`
**Package name**: `main`

### Visual Design

```
╭─ System Log ──────────────────────────────── 1/50 ───────╮
│  12:00:01  INFO   Application started              ▲     │
│  12:00:02  DEBUG  Loading configuration...         █     │
│  12:00:03  INFO   Database connected               █     │
│  12:00:04  WARN   Cache miss on key "user:42"      ░     │
│  12:00:05  ERROR  Connection timeout: redis        ░     │
│  12:00:06  INFO   Retry succeeded                  ░     │
│  12:00:07  DEBUG  Processing batch 1/10            ░     │
│  12:00:08  INFO   Batch complete: 150 records      ░     │
│                                                    ▼     │
╰──────────────────────────────────────────────────────────╯
  j/k ↑/↓ scroll   PgUp/PgDn page   Home/End jump   q quit
```

### Requirements

- `scrollingApp` struct:
  - `scrollY *tui.State[int]`
  - `content *tui.Ref`
  - `entries []logEntry` — pre-generated slice of 50 log entries
- `logEntry` struct: `time string`, `level string`, `message string`
- Constructor: generate 50 log entries with varying levels (INFO, DEBUG, WARN, ERROR) and realistic messages
- Helper: `levelClass(level string) string` — returns tailwind class for level color: INFO→`text-cyan`, DEBUG→`text-bright-black`, WARN→`text-yellow`, ERROR→`text-red`
- `Render` templ:
  - Root: `flex-col h-full`
  - Log container: `border-rounded border-cyan flex-col flex-1`, `scrollable={tui.ScrollVertical}`, `scrollOffset={0, s.scrollY.Get()}`, `ref={s.content}`
  - Scroll position in title area: compute position indicator text
  - Each entry: `@for i, entry := range s.entries`, row with time (font-dim), level (colored), message
  - Custom scrollbar styling: `scrollbarStyle` and `scrollbarThumbStyle` attributes for visual scrollbar
  - Hint bar at bottom (outside scroll)
- `KeyMap()`:
  - `j` / `KeyDown` → scroll by 1
  - `k` / `KeyUp` → scroll by -1
  - `KeyPageDown` → scroll by 10
  - `KeyPageUp` → scroll by -10
  - `KeyHome` → scroll to 0
  - `KeyEnd` → scroll to maxY
  - `q` / `Escape` → quit
- `HandleMouse()`: `MouseWheelDown` → scroll by 3, `MouseWheelUp` → scroll by -3
- `scrollBy(delta int)` helper: clamps to `[0, maxY]` using `s.content.El().MaxScroll()`

### Verification

```bash
go run ./cmd/tui generate ./examples/09-scrolling/...
go build ./examples/09-scrolling/
go run ./examples/09-scrolling/   # visual: 50 colored log entries, all scroll methods work
```

---

## Phase 10: 10-timers-and-watchers

**Goal**: Demonstrate `OnTimer`, channel `Watch`, `Events[T]` broadcast, and the `Watchers()` interface.

**Directory**: `examples/10-timers-and-watchers/`
**Files to create**: `watchers.gsx`, `main.go`
**Package name**: `main`

### Visual Design

```
╭─ Timers & Watchers ─────────────────────────────────────╮
│                                                          │
│  ╭─ Stopwatch ──────────╮  ╭─ Countdown ──────────╮    │
│  │                       │  │                       │    │
│  │     00:01:23          │  │      00:04:37         │    │
│  │                       │  │                       │    │
│  │   s start/stop        │  │   c start/stop        │    │
│  │   r reset             │  │   r reset             │    │
│  ╰───────────────────────╯  ╰───────────────────────╯    │
│                                                          │
│  ╭─ Live Feed ──────────────────────────────────────╮   │
│  │  #1  Hello from producer                          │   │
│  │  #2  Data packet received                         │   │
│  │  #3  Processing complete                          │   │
│  │  Waiting for next message...                      │   │
│  ╰──────────────────────────────────────────────────╯   │
│                                                          │
│  Events received: 3          s/c toggle  r reset  q quit │
╰──────────────────────────────────────────────────────────╯
```

### Requirements

- `watchersApp` struct:
  - `elapsed *tui.State[int]` (stopwatch seconds)
  - `swRunning *tui.State[bool]`
  - `countdown *tui.State[int]` (start at 300 = 5 minutes)
  - `cdRunning *tui.State[bool]`
  - `messages *tui.State[[]string]`
  - `msgCount *tui.State[int]`
  - `dataCh <-chan string`
  - `events *tui.Events[string]`
- Constructor: takes `dataCh <-chan string`, initializes all states/events, subscribes to `events` to increment `msgCount`
- `main.go`: create `dataCh := make(chan string)`, start producer goroutine (every 2-3 seconds), pass it via `tui.WithRootComponent(WatchersApp(dataCh))`
- `Watchers()`:
  - `tui.OnTimer(time.Second, w.tick)` — increments stopwatch (if running) and decrements countdown (if running and > 0)
  - `tui.Watch(w.dataCh, w.addMessage)` — adds message to list
- Methods:
  - `tick()` — conditionally update stopwatch and countdown
  - `addMessage(msg string)` — append to messages and emit event
  - `toggleStopwatch()`, `toggleCountdown()`, `resetAll()`
  - `formatTime(seconds int) string` — returns "HH:MM:SS" format
- `Render` templ:
  - Outer: `border-rounded border-cyan p-2 flex-col gap-2 h-full`
  - Top row: two panels side-by-side (Stopwatch + Countdown), each `border-rounded p-1 flex-col items-center gap-1 flex-1`
    - Time display: large, `text-gradient-cyan-magenta font-bold`
    - Running indicator: `@if running` → `text-green "Running"`, else `text-yellow "Stopped"`
    - Key hints inside panel
  - Feed panel: `border-rounded p-1 flex-col gap-0 flex-1`
    - `@for i, msg := range w.messages.Get()` — show each message with index
    - `@if len(messages) == 0` → "Waiting for messages..." in dim
  - Bottom row: event count + key hints
- `KeyMap()`: `s` toggle stopwatch, `c` toggle countdown, `r` reset all, `q/Escape` quit

### Verification

```bash
go run ./cmd/tui generate ./examples/10-timers-and-watchers/...
go build ./examples/10-timers-and-watchers/
go run ./examples/10-timers-and-watchers/   # visual: timers tick, messages arrive, events count
```

---

## Phase 11: 11-streaming

**Goal**: Real-time streaming data with auto-scroll, sticky-scroll toggle, channel watchers, and timer. Simulates a live metrics stream.

**Directory**: `examples/11-streaming/`
**Files to create**: `streaming.gsx`, `main.go`
**Package name**: `main`

### Visual Design

```
╭─ Live Stream ──────────────────────────── 147 lines ─────╮
│  [12:00:01] cpu  ████████░░░░░░░░░░░░  42%   mem 1.2G   │
│  [12:00:02] cpu  ██████████░░░░░░░░░░  51%   mem 1.3G   │
│  [12:00:03] cpu  ███████░░░░░░░░░░░░░  37%   mem 1.1G   │
│  [12:00:04] cpu  ████████████░░░░░░░░  62%   mem 1.4G   │
│  [12:00:05] cpu  █████████████████░░░  85%   mem 1.8G   │
│  ...                                                      │
│  [12:00:15] cpu  ██████████████░░░░░░  71%   mem 1.5G ▼ │
╰──────────────────────────────────────────────────────────╯
  j/k scroll  Space auto-scroll: ON  Elapsed: 15s  q quit
```

### Requirements

- `streamingApp` struct:
  - `dataCh <-chan string`
  - `lines *tui.State[[]string]`
  - `scrollY *tui.State[int]`
  - `stickToBottom *tui.State[bool]` (start true)
  - `content *tui.Ref`
  - `elapsed *tui.State[int]`
- Constructor: takes `dataCh <-chan string`, initializes states, and stores the channel
- `main.go`: create `dataCh := make(chan string)`, start a producer goroutine that sends formatted metric lines every 200ms with random cpu% (30-95), memory (1.0-2.5G), and visual bar using block characters (█ and ░), then call `tui.WithRootComponent(Streaming(dataCh))`
- `Watchers()`:
  - `tui.OnTimer(time.Second, s.tick)` — increment elapsed
  - `tui.Watch(s.dataCh, s.addLine)` — append line, auto-scroll if sticky
- `addLine(line string)`: append to lines state. If `stickToBottom`, set scrollY to `math.MaxInt` (gets clamped to maxY).
- `Render` templ:
  - Root: `flex-col h-full`
  - Content area: `border-rounded border-cyan flex-1`, `scrollable={tui.ScrollVertical}`, `scrollOffset={0, s.scrollY.Get()}`, `ref={s.content}`
  - Title shows line count
  - `@for _, line := range s.lines.Get()` — each line as a span
  - Status bar: shows auto-scroll state (ON=`text-green`, OFF=`text-yellow`), elapsed time, key hints
- `KeyMap()`:
  - `j/KeyDown` → scroll +1, set stickToBottom=false
  - `k/KeyUp` → scroll -1, set stickToBottom=false
  - `Space` → toggle stickToBottom (if turning on, scroll to bottom)
  - `KeyEnd` → scroll to bottom, stickToBottom=true
  - `KeyHome` → scroll to 0, stickToBottom=false
  - `q/Escape` → quit
- `HandleMouse()`: wheel scroll, disable stickToBottom on manual scroll

### Verification

```bash
go run ./cmd/tui generate ./examples/11-streaming/...
go build ./examples/11-streaming/
go run ./examples/11-streaming/   # visual: metrics stream in, auto-scroll follows, Space toggles
```

---

## Phase 12: 12-multi-component

**Goal**: Multiple struct components sharing state, conditional KeyMaps, component constructors with state params. A file-explorer-like UI.

**Directory**: `examples/12-multi-component/`
**Files to create**: `app.gsx`, `search.gsx`, `sidebar.gsx`, `main.go`
**Package name**: `main`

### Visual Design

```
╭─ Sidebar ──────╮ ╭─ Content ────────────────────────────╮
│                 │ │                                       │
│  ▶ Documents   │ │   Documents/                          │
│    Images      │ │   ├── report.pdf                      │
│    Music       │ │   ├── notes.md                        │
│    Projects    │ │   ├── budget.xlsx                     │
│                 │ │   └── readme.txt                      │
│                 │ │                                       │
│  Ctrl+B toggle │ │                                       │
╰─────────────────╯ ╰──────────────────────────────────────╯
╭─ Search: report ─────────────────────────────────────────╮
│                                                           │
╰──────────────────────────────────────────────────────────╯
   / search  Ctrl+B sidebar  j/k navigate  q quit
```

### Requirements

**app.gsx** — `myApp` struct (orchestrator):
  - `showSidebar *tui.State[bool]` (start true)
  - `searchActive *tui.State[bool]` (start false)
  - `query *tui.State[string]` (start "")
  - `selectedCategory *tui.State[int]` (start 0)
  - `categories []string` = ["Documents", "Images", "Music", "Projects"]
  - `categoryFiles map[string][]string` — mock file data per category
  - Constructor: init all, populate file data
  - `Render`: row layout with conditional sidebar + content area, search bar below
  - `KeyMap()`:
    - `OnRuneStop('/', ...)` → activate search (only when search not active)
    - `OnKey(tui.KeyCtrlB, ...)` → toggle sidebar
    - `j/KeyDown` → next category (only when search is not active)
    - `k/KeyUp` → prev category (only when search is not active)
    - `q/Escape` → quit (only when search not active)

**sidebar.gsx** — `sidebar` struct:
  - Constructor takes: `categories []string`, `selected *tui.State[int]`, `visible *tui.State[bool]`
  - `Render`: `@if visible`, list categories with `@for`, highlight selected with ▶ prefix and `text-cyan font-bold`
  - `border-rounded border-cyan p-1 flex-col`

**search.gsx** — `searchInput` struct:
  - Constructor takes: `active *tui.State[bool]`, `query *tui.State[string]`
  - `Render`: `border-rounded p-1`, shows "Search: " + query text when active, "/ to search" when inactive
  - Active: `border-magenta`, Inactive: `border-bright-black`
  - `KeyMap()` — only returns bindings when active:
    - `OnRunesStop(...)` → append char to query and prevent parent bindings from handling typed runes
    - `OnKey(tui.KeyBackspace, ...)` → delete last char
    - `OnKey(tui.KeyEscape, ...)` → deactivate, clear query
    - `OnKey(tui.KeyEnter, ...)` → deactivate (keep query for filtering)

**main.go**: `tui.NewApp(tui.WithRootComponent(MyApp()))`

### Verification

```bash
go run ./cmd/tui generate ./examples/12-multi-component/...
go build ./examples/12-multi-component/
go run ./examples/12-multi-component/   # visual: sidebar toggles, search captures input, categories navigate
```

---

## Phase 13: 13-dashboard

**Goal**: Capstone example combining animation, timers, channels, gradients, and complex layout. A live metrics dashboard.

**Directory**: `examples/13-dashboard/`
**Files to create**: `dashboard.gsx`, `main.go`
**Package name**: `main`

### Visual Design

```
╭─ Dashboard ──────────────────────────────────────────────╮
│  ╭─ CPU ────────────╮ ╭─ Memory ────────╮ ╭─ Disk ─────╮│
│  │  ████████░░  78% │ │ ██████░░░░  62% │ │ ████░░  45%││
│  ╰──────────────────╯ ╰─────────────────╯ ╰────────────╯│
│                                                           │
│  ╭─ Network ─────────────────────────────────────────╮   │
│  │  ▁▂▃▅▇█▇▅▃▂▁▂▃▅▇█▇▅▃▂▁▂▃▅▇                     │   │
│  │  In: 142 MB/s        Out: 89 MB/s                  │   │
│  ╰────────────────────────────────────────────────────╯   │
│                                                           │
│  ╭─ Recent Events ──────────────────────────────────╮    │
│  │  12:00:15  Deploy completed successfully          │    │
│  │  12:00:12  Health check passed                    │    │
│  │  12:00:09  New connection from 10.0.0.5           │    │
│  │  12:00:06  Cache warmed: 1.2k entries             │    │
│  ╰───────────────────────────────────────────────────╯    │
│                                                           │
│                                                q quit     │
╰──────────────────────────────────────────────────────────╯
```

### Requirements

- `dashboardApp` struct:
  - `cpu`, `memory`, `disk` — `*tui.State[int]` (percentages, start at random values)
  - `netIn`, `netOut` — `*tui.State[int]` (MB/s values)
  - `sparkline *tui.State[[]int]` — last 30 network readings for sparkline chart
  - `events *tui.State[[]string]` — last 5 event messages
  - `eventCh <-chan string` — channel for event feed
- Constructor: takes `eventCh <-chan string` and initializes all states
- `Watchers()`:
  - `tui.OnTimer(500*time.Millisecond, d.animate)` — fluctuate cpu/memory/disk/network randomly within bounds
  - `tui.Watch(d.eventCh, d.addEvent)` — add event to list (keep last 5)
- `animate()`:
  - Each metric += random(-5, +5), clamped to [10, 95] for cpu/mem/disk
  - Network in/out: random walk within [20, 200]
  - Append current netIn to sparkline (keep last 30 values)
- Helper `barString(value, max, width int) string` — returns "████░░░░" string of given width
- Helper `sparklineString(values []int) string` — maps values to block chars: ▁▂▃▄▅▆▇█ (8 levels)
- Helper `metricColor(value int) string` — returns class: <50→`text-green`, 50-80→`text-yellow`, >80→`text-red`
- `Render` templ:
  - Outer: `border-rounded border-cyan p-2 flex-col gap-2 h-full`
  - **Metrics row**: three panels (CPU, Memory, Disk) in a row, each `border-rounded p-1 flex-1 items-center`:
    - Bar visualization: `@let bar = barString(cpu, 100, 12)` rendered with metric-appropriate color
    - Percentage display
  - **Network panel**: `border-rounded border-cyan p-1 flex-col gap-1`
    - Sparkline: `@let spark = sparklineString(d.sparkline.Get())` displayed as `text-gradient-green-cyan`
    - In/Out values
  - **Events panel**: `border-rounded border-cyan p-1 flex-col`
    - `@for _, event := range d.events.Get()` — each event with timestamp, latest at top
    - `@if len(events) == 0` → "Waiting for events..." in dim
  - Quit hint at bottom right
- `KeyMap()`: `q/Escape` → quit
- `main.go`: create `eventCh := make(chan string)`, start producer goroutine that sends random event messages every 3-5 seconds (deploy, health check, connection, cache events), then call `tui.WithRootComponent(Dashboard(eventCh))`

### Verification

```bash
go run ./cmd/tui generate ./examples/13-dashboard/...
go build ./examples/13-dashboard/
go run ./examples/13-dashboard/   # visual: metrics fluctuate, sparkline updates, events appear
```

---

## Phase 14: 14-ai-chat

**Goal**: Real-world production app pattern — TextArea widget, inline mode, alternate screen switching, AppBinder, multi-screen navigation with settings modal.

**Directory**: `examples/14-ai-chat/` (move from `examples/ai-chat/`)
**Files to create**: `chat.gsx`, `settings/settings.gsx`, `main.go`
**Package name**: `main` (chat), `settings` (settings subpackage)

### Visual Design

**Chat screen (inline mode)** — occupies bottom 3+ rows of terminal, grows with input:
```
╭───────────────────────────────────────────────────────────╮
│  Type a message...                                        │
│                                                  Ctrl+S ⚙ │
╰───────────────────────────────────────────────────────────╯
```

Messages print above the widget via `PrintAboveln()`:
```
You: Hello, how are you?
AI: I'm doing great! How can I help you today?
╭───────────────────────────────────────────────────────────╮
│  Tell me about Go                                         │
│                                                  Ctrl+S ⚙ │
╰───────────────────────────────────────────────────────────╯
```

**Settings screen (alternate screen)** — full-screen settings panel:
```
╭─ Settings ───────────────────────────────────────────────╮
│                                                           │
│  ╭─ Provider ────────────────╮                           │
│  │  ◀  OpenAI  ▶             │  ← focused: border-double │
│  ╰───────────────────────────╯                           │
│  ╭─ Model ───────────────────╮                           │
│  │  ◀  gpt-4.1-mini  ▶      │                           │
│  ╰───────────────────────────╯                           │
│  ╭─ Temperature ─────────────╮                           │
│  │  ━━━━━━━●━━━━━━  0.7     │                           │
│  ╰───────────────────────────╯                           │
│  ╭─ System Prompt ───────────╮                           │
│  │  You are a helpful        │                           │
│  │  assistant that...        │                           │
│  ╰───────────────────────────╯                           │
│                                                           │
│  Tab: next section  ←/→ adjust  Ctrl+S: back to chat    │
╰──────────────────────────────────────────────────────────╯
```

### Requirements

**chat.gsx** — `chat` struct:
  - `app *tui.App` (bound via AppBinder)
  - `textarea *tui.TextArea`
  - `showSettings *tui.State[bool]`
  - `settingsView *settings.SettingsApp`
- Implements `tui.AppBinder`:
  ```go
  func (c *chat) BindApp(app *tui.App) {
      c.app = app
      c.showSettings.BindApp(app)
      c.textarea.BindApp(app)
      c.settingsView.BindApp(app)
  }
  ```
- Constructor `Chat()`: creates textarea with `tui.NewTextArea(tui.WithTextAreaPlaceholder("Type a message..."), tui.WithTextAreaOnSubmit(c.submit))`, creates settings view, init showSettings=false
- `submit(text string)`: prints "You: {text}" above widget via `app.PrintAboveln()`, then prints a simulated AI response
- `toggleSettings()`: toggles `showSettings`, calls `app.EnterAlternateScreen()` or `app.ExitAlternateScreen()`
- `Render` templ: `@if c.showSettings.Get()` → render settings view, `@else` → render textarea with border
- `KeyMap()`: context-aware — when settings showing, delegate to settings KeyMap + Ctrl+S to close; when chat, Ctrl+S to open settings, Escape to quit
- `Watchers()`: delegates to settings if it has watchers

**settings/settings.gsx** — `SettingsApp` struct:
  - `provider *tui.State[string]` (cycles: "openai", "anthropic", "google")
  - `model *tui.State[string]` (maps to available models per provider)
  - `temperature *tui.State[float64]` (0.0-1.0, step 0.1)
  - `systemPrompt *tui.State[string]` (cycles through presets)
  - `focusedSection *tui.State[int]` (0-3 for the 4 sections)
  - `AvailableProviders []string`
  - `ProviderModels map[string][]string`
  - `SystemPrompts []string` (preset prompts)
- Constructor: init all states with defaults
- `KeyMap()`:
  - `Tab` → next section (wraps)
  - `Shift+Tab` / `KeyUp` → prev section
  - `KeyDown` → next section
  - `KeyLeft` / `h` → previous value for current section (cycle provider, model, decrease temp, prev prompt)
  - `KeyRight` / `l` → next value for current section
- Methods:
  - `cycleProvider(delta int)`, `cycleModel(delta int)`, `adjustTemp(delta float64)`, `cycleSystemPrompt(delta int)`
  - `tempBar() string` — visual temperature bar using ━ and ● characters
  - `promptPreview() string` — word-wrapped preview of current system prompt, truncated to 3 lines
  - `sectionBorder(idx int) tui.BorderStyle` — returns `BorderDouble` if focused, `BorderRounded` otherwise
  - `sectionBorderClass(idx int) string` — returns border color class: section 0=cyan, 1=blue, 2=yellow, 3=green
- `Render` templ:
  - Outer: `border-rounded border-cyan p-2 flex-col gap-2 h-full`
  - Title: `text-gradient-cyan-magenta font-bold text-center` "Settings"
  - Four section panels, each with dynamic border style and color based on focus
  - Provider: shows `◀ name ▶` with arrows
  - Model: shows `◀ name ▶` with arrows
  - Temperature: visual bar `━━━━━━━●━━━━━━ 0.7`
  - System Prompt: wrapped preview text
  - Decorative gradient borders on outer container: `border-gradient-cyan-magenta`
  - Hint bar at bottom

**main.go**:
  - Detect terminal height with `golang.org/x/term`
  - `tui.NewApp(tui.WithRootComponent(Chat()), tui.WithInlineHeight(3))`
  - Run, handle errors

### Verification

```bash
go run ./cmd/tui generate ./examples/14-ai-chat/...
go build ./examples/14-ai-chat/
go run ./examples/14-ai-chat/   # visual: inline chat, type and submit, Ctrl+S opens settings, settings navigate
```

---

## Phase 15: Final Verification

**Goal**: Generate all code, build everything, run tests.

```bash
cd /Users/joelholsteen/go/src/github.com/grindlemire/go-tui

# Generate all _gsx.go files
go run ./cmd/tui generate ./examples/...

# Build all examples
go build ./examples/01-hello/
go build ./examples/02-styling/
go build ./examples/03-layout/
go build ./examples/04-components/
go build ./examples/05-state/
go build ./examples/06-keyboard/
go build ./examples/07-refs-and-clicks/
go build ./examples/08-elements/
go build ./examples/09-scrolling/
go build ./examples/10-timers-and-watchers/
go build ./examples/11-streaming/
go build ./examples/12-multi-component/
go build ./examples/13-dashboard/
go build ./examples/14-ai-chat/

# Run project tests
go test ./...

# Visually run each example (manual check)
```

Fix any compilation errors. Ensure all examples build cleanly and render beautifully.
