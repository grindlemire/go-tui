# Scrolling

## Overview

When content is taller or wider than its container, go-tui supports scrollable regions. You mark an element as scrollable, and the framework handles clipping content to the visible area and drawing a scrollbar. Scroll position is controlled through keyboard bindings, mouse wheel events, or programmatic calls on the element.

## Making an Element Scrollable

Add scrolling to any container with a Tailwind class or the `scrollable` attribute.

**Via class:**

```gsx
<div class="overflow-y-scroll h-20">
    // content taller than 20 rows will scroll vertically
</div>

<div class="overflow-x-scroll w-40">
    // content wider than 40 columns will scroll horizontally
</div>

<div class="overflow-scroll h-20 w-40">
    // scrolls in both directions
</div>
```

**Via attribute:**

```gsx
<div scrollable={tui.ScrollVertical} height={20}>
    // equivalent to overflow-y-scroll h-20
</div>
```

The three scroll modes are `tui.ScrollVertical`, `tui.ScrollHorizontal`, and `tui.ScrollBoth`. Use `tui.ScrollNone` (the default) to disable scrolling.

When you enable scrolling on an element, the framework also makes it focusable so it can receive keyboard and mouse events. Content inside the scrollable container is laid out at its natural size, then clipped to the visible area. A vertical scrollbar appears automatically along the right edge when content exceeds the viewport height.

## Controlling Scroll Position

To move the scroll position in response to user input, you need a ref to the scrollable element. Attach a `*tui.Ref` in your render method and call scroll methods from your event handlers:

```gsx
package main

import (
    "fmt"

    tui "github.com/grindlemire/go-tui"
)

type scrollApp struct {
    content   *tui.Ref
    scrollY   *tui.State[int]
    items     []string
}

func ScrollApp() *scrollApp {
    items := make([]string, 100)
    for i := range items {
        items[i] = fmt.Sprintf("Item %d", i+1)
    }
    return &scrollApp{
        content: tui.NewRef(),
        scrollY: tui.NewState(0),
        items:   items,
    }
}

templ (s *scrollApp) Render() {
    <div class="flex-col p-1 border-rounded border-cyan">
        <span class="font-bold text-cyan">Scrollable List</span>
        <div class="overflow-y-scroll" height={15} ref={s.content}
             scrollOffset={0, s.scrollY.Get()}>
            for i, item := range s.items {
                <span>{fmt.Sprintf("  %s", item)}</span>
            }
        </div>
        <span class="font-dim">j/k scroll | esc quit</span>
    </div>
}
```

The `scrollOffset` attribute accepts two values: the horizontal and vertical offset. By binding the vertical offset to `s.scrollY.Get()`, the scroll position persists across renders. Without this, the position would reset to zero every time the component re-renders, since go-tui recreates elements on each render pass.

## Keyboard Scrolling

Wire up key bindings that adjust the scroll position through the ref. A common pattern is a `scrollBy` helper that clamps the offset to valid bounds:

```go
func (s *scrollApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.On(tui.Rune('j'), func(ke tui.KeyEvent) { s.scrollBy(1) }),
        tui.On(tui.Rune('k'), func(ke tui.KeyEvent) { s.scrollBy(-1) }),
        tui.On(tui.KeyDown, func(ke tui.KeyEvent) { s.scrollBy(1) }),
        tui.On(tui.KeyUp, func(ke tui.KeyEvent) { s.scrollBy(-1) }),
        tui.On(tui.KeyPageDown, func(ke tui.KeyEvent) { s.scrollBy(10) }),
        tui.On(tui.KeyPageUp, func(ke tui.KeyEvent) { s.scrollBy(-10) }),
        tui.On(tui.KeyHome, func(ke tui.KeyEvent) { s.scrollY.Set(0) }),
        tui.On(tui.KeyEnd, func(ke tui.KeyEvent) { s.scrollToEnd() }),
    }
}

func (s *scrollApp) scrollBy(delta int) {
    el := s.content.El()
    if el == nil {
        return
    }
    _, maxY := el.MaxScroll()
    newY := s.scrollY.Get() + delta
    if newY < 0 {
        newY = 0
    }
    if newY > maxY {
        newY = maxY
    }
    s.scrollY.Set(newY)
}

func (s *scrollApp) scrollToEnd() {
    el := s.content.El()
    if el == nil {
        return
    }
    _, maxY := el.MaxScroll()
    s.scrollY.Set(maxY)
}
```

`MaxScroll()` returns the maximum valid offset in each direction. It's calculated as `contentSize - viewportSize`, so scrolling to `maxY` puts the last line of content at the bottom of the viewport.

You can also call scroll methods directly on the element instead of managing state yourself:

```go
func (s *scrollApp) scrollBy(delta int) {
    if el := s.content.El(); el != nil {
        el.ScrollBy(0, delta)
    }
}
```

`ScrollBy` clamps automatically, so you don't need the bounds check. The trade-off: the element manages its own position internally rather than through your `State[int]`. The state-driven approach gives you more control (you can read the position in your render method to display "line X of Y"). The element method is more concise.

## Mouse Scrolling

Mouse wheel events arrive as `MouseEvent` values with `MouseWheelUp` and `MouseWheelDown` buttons. Enable mouse support on the app and implement `HandleMouse`:

```go
app, err := tui.NewApp(
    tui.WithRootComponent(ScrollApp()),
    tui.WithMouse(),
)
```

```go
func (s *scrollApp) HandleMouse(me tui.MouseEvent) bool {
    switch me.Button {
    case tui.MouseWheelUp:
        s.scrollBy(-3)
        return true
    case tui.MouseWheelDown:
        s.scrollBy(3)
        return true
    }
    return false
}
```

Three lines per wheel tick is a reasonable default. Adjust to taste.

## Bounds Checking

The element provides several methods for querying scroll geometry:

```go
el := s.content.El()

// Maximum valid scroll offset in each direction
maxX, maxY := el.MaxScroll()

// Current scroll position
x, y := el.ScrollOffset()

// Total content dimensions (may exceed viewport)
contentW, contentH := el.ContentSize()

// Visible area dimensions
viewportW, viewportH := el.ViewportSize()

// True when scrolled to the bottom
atBottom := el.IsAtBottom()
```

`MaxScroll` returns `(0, 0)` when content fits within the viewport. All scroll operations clamp to the valid range automatically, so calling `ScrollTo(0, -100)` is the same as calling `ScrollToTop()`.

## Auto-Scroll (Sticky Bottom)

A common pattern for log viewers and chat apps: new content appears at the bottom, and the view follows along automatically unless the user has scrolled up to read older entries.

```gsx
package main

import (
    "fmt"
    "time"

    tui "github.com/grindlemire/go-tui"
)

type logViewer struct {
    lines   *tui.State[[]string]
    scrollY *tui.State[int]
    sticky  *tui.State[bool]
    content *tui.Ref
    counter int
}

func LogViewer() *logViewer {
    return &logViewer{
        lines:   tui.NewState([]string{}),
        scrollY: tui.NewState(0),
        sticky:  tui.NewState(true),
        content: tui.NewRef(),
    }
}

func (l *logViewer) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnTimer(500*time.Millisecond, l.addLine),
    }
}

func (l *logViewer) addLine() {
    l.counter++
    l.lines.Update(func(v []string) []string {
        return append(v, fmt.Sprintf("[%s] Event #%d", time.Now().Format("15:04:05"), l.counter))
    })
    if l.sticky.Get() {
        if el := l.content.El(); el != nil {
            _, maxY := el.MaxScroll()
            l.scrollY.Set(maxY + 1)
        }
    }
}

func (l *logViewer) scrollBy(delta int) {
    el := l.content.El()
    if el == nil {
        return
    }
    _, maxY := el.MaxScroll()
    newY := l.scrollY.Get() + delta
    if newY < 0 {
        newY = 0
    }
    if newY > maxY {
        newY = maxY
    }
    l.scrollY.Set(newY)

    // Disable sticky when user scrolls up, re-enable at bottom
    l.sticky.Set(newY >= maxY)
}

func (l *logViewer) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.On(tui.Rune('j'), func(ke tui.KeyEvent) { l.scrollBy(1) }),
        tui.On(tui.Rune('k'), func(ke tui.KeyEvent) { l.scrollBy(-1) }),
        tui.On(tui.KeyDown, func(ke tui.KeyEvent) { l.scrollBy(1) }),
        tui.On(tui.KeyUp, func(ke tui.KeyEvent) { l.scrollBy(-1) }),
        tui.On(tui.KeyEnd, func(ke tui.KeyEvent) {
            if el := l.content.El(); el != nil {
                _, maxY := el.MaxScroll()
                l.scrollY.Set(maxY)
                l.sticky.Set(true)
            }
        }),
    }
}

templ (l *logViewer) Render() {
    <div class="flex-col p-1 border-rounded border-cyan">
        <div class="flex justify-between">
            <span class="font-bold text-cyan">Log Viewer</span>
            if l.sticky.Get() {
                <span class="text-green font-bold">FOLLOWING</span>
            } else {
                <span class="text-yellow font-dim">PAUSED</span>
            }
        </div>
        <div class="overflow-y-scroll" height={12} ref={l.content}
             scrollOffset={0, l.scrollY.Get()}>
            for _, line := range l.lines.Get() {
                <span class="font-dim">{line}</span>
            }
        </div>
        <span class="font-dim">j/k scroll | End to follow | esc quit</span>
    </div>
}
```

The pattern works like this: `sticky` starts as `true`. Each time a new line appears, if sticky is on, the scroll position jumps to the bottom. When the user scrolls up manually, `sticky` turns off and new content no longer auto-scrolls. Pressing End re-enables it.

You can also use `ScrollToBottom()` on the element directly, which has built-in deferred behavior: it scrolls to the bottom immediately and also schedules a second scroll after the next layout pass, catching any content that was added between the scroll call and the render.

## Scrollbar Styling

The scrollbar track and thumb have default styles (BrightBlack track, White thumb). Customize them with Tailwind classes:

```gsx
<div class="overflow-y-scroll scrollbar-cyan scrollbar-thumb-bright-white" height={15}>
    // cyan track, bright white thumb
</div>
```

Scrollbar classes follow the same color names as text and border colors: `scrollbar-red`, `scrollbar-green`, `scrollbar-cyan`, etc., plus bright variants like `scrollbar-bright-cyan`. Thumb colors use the `scrollbar-thumb-` prefix.

For hex colors, use bracket notation: `scrollbar-[#FF6600]`, `scrollbar-thumb-[#AACCFF]`.

For programmatic control in Go code, use the option functions directly:

```go
tui.New(
    tui.WithScrollable(tui.ScrollVertical),
    tui.WithHeight(15),
    tui.WithScrollbarStyle(tui.NewStyle().Foreground(tui.ANSIColor(tui.Cyan))),
    tui.WithScrollbarThumbStyle(tui.NewStyle().Foreground(tui.ANSIColor(tui.BrightWhite))),
)
```

The scrollbar is a single column on the right edge of the scrollable area. The track character is `│` (box drawing vertical) and the thumb is `█` (full block). Thumb size is proportional to how much of the content is visible, with a minimum of one row.

## Scroll Methods

The `*Element` type provides these methods for programmatic scroll control:

| Method | Description |
|---|---|
| `ScrollTo(x, y)` | Set absolute scroll position (clamped to valid range) |
| `ScrollBy(dx, dy)` | Adjust scroll position by delta (clamped) |
| `ScrollToTop()` | Scroll to the top of content |
| `ScrollToBottom()` | Scroll to the bottom (with deferred re-scroll after layout) |
| `ScrollIntoView(child)` | Scroll minimally to make a child element visible |

Query methods:

| Method | Returns |
|---|---|
| `ScrollOffset()` | `(x, y int)` — current scroll position |
| `ContentSize()` | `(width, height int)` — total content dimensions |
| `ViewportSize()` | `(width, height int)` — visible area dimensions |
| `MaxScroll()` | `(maxX, maxY int)` — maximum scroll offset |
| `IsScrollable()` | `bool` — whether scrolling is enabled |
| `ScrollModeValue()` | `ScrollMode` — current scroll mode |
| `IsAtBottom()` | `bool` — whether scrolled to the bottom |

`ScrollIntoView` is useful when you have a selected item in a long list and want to keep it visible as the selection moves:

```go
func (s *scrollApp) selectNext() {
    s.selected.Update(func(v int) int {
        if v >= len(s.items)-1 {
            return 0
        }
        return v + 1
    })
    // Ensure the selected item is visible
    if el := s.itemRefs.At(s.selected.Get()); el != nil {
        if container := s.content.El(); container != nil {
            container.ScrollIntoView(el)
        }
    }
}
```

## Complete Example

A scrollable list with keyboard navigation, mouse wheel support, and a styled scrollbar:

```gsx
package main

import (
    "fmt"

    tui "github.com/grindlemire/go-tui"
)

type fileList struct {
    files    []string
    selected *tui.State[int]
    scrollY  *tui.State[int]
    content  *tui.Ref
}

func FileList() *fileList {
    files := []string{
        "main.go", "app.go", "app_loop.go", "app_render.go",
        "buffer.go", "cell.go", "color.go", "component.go",
        "dirty.go", "element.go", "element_options.go",
        "element_render.go", "element_scroll.go", "escape.go",
        "event.go", "focus.go", "key.go", "keymap.go",
        "layout.go", "mount.go", "ref.go", "render.go",
        "state.go", "style.go", "terminal.go", "watcher.go",
    }
    return &fileList{
        files:    files,
        selected: tui.NewState(0),
        scrollY:  tui.NewState(0),
        content:  tui.NewRef(),
    }
}

func (f *fileList) scrollBy(delta int) {
    el := f.content.El()
    if el == nil {
        return
    }
    _, maxY := el.MaxScroll()
    newY := f.scrollY.Get() + delta
    if newY < 0 {
        newY = 0
    }
    if newY > maxY {
        newY = maxY
    }
    f.scrollY.Set(newY)
}

func (f *fileList) moveTo(idx int) {
    if idx < 0 {
        idx = len(f.files) - 1
    }
    if idx >= len(f.files) {
        idx = 0
    }
    f.selected.Set(idx)

    // Keep selected item visible by adjusting scroll
    el := f.content.El()
    if el == nil {
        return
    }
    _, vpH := el.ViewportSize()
    y := f.scrollY.Get()
    if idx < y {
        f.scrollY.Set(idx)
    } else if idx >= y+vpH {
        f.scrollY.Set(idx - vpH + 1)
    }
}

func (f *fileList) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.On(tui.Rune('j'), func(ke tui.KeyEvent) { f.moveTo(f.selected.Get() + 1) }),
        tui.On(tui.Rune('k'), func(ke tui.KeyEvent) { f.moveTo(f.selected.Get() - 1) }),
        tui.On(tui.KeyDown, func(ke tui.KeyEvent) { f.moveTo(f.selected.Get() + 1) }),
        tui.On(tui.KeyUp, func(ke tui.KeyEvent) { f.moveTo(f.selected.Get() - 1) }),
        tui.On(tui.KeyPageDown, func(ke tui.KeyEvent) { f.moveTo(f.selected.Get() + 10) }),
        tui.On(tui.KeyPageUp, func(ke tui.KeyEvent) { f.moveTo(f.selected.Get() - 10) }),
        tui.On(tui.KeyHome, func(ke tui.KeyEvent) { f.moveTo(0) }),
        tui.On(tui.KeyEnd, func(ke tui.KeyEvent) { f.moveTo(len(f.files) - 1) }),
    }
}

func (f *fileList) HandleMouse(me tui.MouseEvent) bool {
    switch me.Button {
    case tui.MouseWheelUp:
        f.scrollBy(-3)
        return true
    case tui.MouseWheelDown:
        f.scrollBy(3)
        return true
    }
    return false
}

templ (f *fileList) Render() {
    <div class="flex-col p-1 border-rounded border-cyan">
        <span class="text-gradient-cyan-magenta font-bold">Files</span>
        <div class="overflow-y-scroll scrollbar-cyan scrollbar-thumb-bright-cyan"
             height={12} ref={f.content} scrollOffset={0, f.scrollY.Get()}>
            for i, name := range f.files {
                if i == f.selected.Get() {
                    <span class="text-cyan font-bold bg-bright-black">{fmt.Sprintf(" > %s ", name)}</span>
                } else {
                    <span>{fmt.Sprintf("   %s", name)}</span>
                }
            }
        </div>
        <div class="flex justify-between">
            <span class="font-dim">j/k navigate | esc quit</span>
            <span class="font-dim">{fmt.Sprintf("%d/%d", f.selected.Get()+1, len(f.files))}</span>
        </div>
    </div>
}
```

With `main.go`:

```go
package main

import (
    "fmt"
    "os"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    app, err := tui.NewApp(
        tui.WithRootComponent(FileList()),
        tui.WithMouse(),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

Generate and run:

```bash
tui generate ./...
go run .
```

The file browser with scroll should look like this:

![Scrolling screenshot](/guides/10.png)

## Next Steps

- [Timers, Watchers, and Channels](watchers) -- Background operations that update scrollable content over time
- [Event Handling](events) -- Keyboard and mouse input in depth
