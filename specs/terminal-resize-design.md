# Terminal Resize Resilience Specification

**Status:** Planned
**Version:** 1.0
**Last Updated:** 2025-01-25

---

## 1. Overview

### Purpose

Make go-tui automatically handle terminal resize events without screen corruption, buffer scrambling, or visual artifacts. The framework should fully manage resize internally so application developers don't need to write any resize-handling code.

### Goals

- Automatically detect terminal size changes via SIGWINCH
- Debounce rapid resize events (16ms window) to avoid excessive redraws during window drag
- Clear screen and perform full redraw after resize to eliminate visual corruption
- Update buffer dimensions to match new terminal size
- Mark element tree dirty to trigger layout recalculation
- Maintain backward compatibility with existing applications

### Non-Goals

- Custom per-app resize handling callbacks (apps can still listen to ResizeEvent if they want extra behavior)
- Animated transitions during resize
- Minimum/maximum size constraints (future feature)
- Multi-monitor handling

---

## 2. Architecture

### Current State (Problem)

```
┌─────────────┐    SIGWINCH    ┌─────────────┐    ResizeEvent    ┌─────────────┐
│  Terminal   │───────────────►│ EventReader │─────────────────►│     App     │
└─────────────┘                └─────────────┘                   └──────┬──────┘
                                                                        │
                                                          Dispatch() resizes buffer
                                                          but Render() uses diff-only
                                                                        │
                                                                        ▼
                                                              Screen has stale content
                                                              from old terminal size
```

**Current issues:**
1. After buffer resize, front buffer is cleared but terminal still has old content
2. `Render()` only flushes changed cells, leaving garbage on screen
3. No debouncing - rapid resizes cause excessive work
4. Applications must manually call `RenderFull()` after resize

### Proposed Solution

```
┌─────────────┐    SIGWINCH    ┌─────────────┐    Debounced     ┌─────────────┐
│  Terminal   │───────────────►│ EventReader │───ResizeEvent───►│     App     │
└─────────────┘                └──────┬──────┘                   └──────┬──────┘
                                      │                                 │
                              Coalesce events                   Dispatch() sets
                              within 16ms window                needsFullRedraw flag
                                                                        │
                                                                        ▼
                                                               Render() checks flag
                                                               and uses RenderFull()
                                                                        │
                                                                        ▼
                                                               Clean screen, no artifacts
```

### Component Changes

| Component | Change |
|-----------|--------|
| `pkg/tui/reader.go` | Add debouncing for SIGWINCH signals |
| `pkg/tui/app.go` | Add `needsFullRedraw` flag; auto-upgrade Render to RenderFull when set |
| `pkg/tui/buffer.go` | Clear front buffer to force full diff (already works, no change needed) |

---

## 3. Core Entities

### App Changes

```go
// App manages the application lifecycle: terminal setup, event loop, and rendering.
type App struct {
    terminal        *ANSITerminal
    buffer          *Buffer
    reader          EventReader
    focus           *FocusManager
    root            Renderable
    needsFullRedraw bool  // NEW: Set after resize, cleared after RenderFull
}
```

### EventReader Debouncing

The `stdinReader` will coalesce multiple SIGWINCH signals received within a 16ms window into a single `ResizeEvent`.

```go
type stdinReader struct {
    fd              int
    buf             []byte
    partialBuf      []byte
    pending         []Event
    sigCh           chan os.Signal
    lastResizeTime  time.Time     // NEW: Track last resize for debouncing
    pendingResize   *ResizeEvent  // NEW: Buffered resize event
}
```

---

## 4. User Experience

### Before (Current Behavior)

```
User resizes terminal → Screen scrambles → User must restart app
```

### After (Proposed Behavior)

```
User resizes terminal → Brief flash → Clean redraw at new size
```

### API Changes

**None** - This is fully transparent to application developers. Existing code continues to work:

```go
// Existing app code - no changes needed
for {
    event, ok := app.PollEvent(50 * time.Millisecond)
    if ok {
        app.Dispatch(event)  // Automatically handles resize
    }
    app.Render()  // Automatically uses full redraw when needed
}
```

Applications can still listen for `ResizeEvent` if they want to do additional work (e.g., update root element dimensions):

```go
case tui.ResizeEvent:
    // Optional: Update root size (for apps that want explicit control)
    width, height = e.Width, e.Height
    style := root.Style()
    style.Width = layout.Fixed(width)
    style.Height = layout.Fixed(height)
    root.SetStyle(style)
    // No need to call Dispatch() - framework handles the rest
```

---

## 5. Complexity Assessment

| Size | Phases | When to Use |
|------|--------|-------------|
| Small | 1-2 | Single component, bug fix, minor enhancement |
| **Medium** | **3-4** | **New feature touching multiple files/components** |
| Large | 5-6 | Cross-cutting feature, new subsystem |

**Assessed Size:** Small
**Recommended Phases:** 2
**Rationale:**
- Changes are localized to two files (app.go, reader.go)
- Well-defined scope with clear boundaries
- No new public APIs - internal implementation change
- Low risk of breaking existing functionality
- Tests can be embedded in implementation phases

---

## 6. Success Criteria

1. Resizing terminal window does not cause screen corruption or visual artifacts
2. Buffer dimensions match terminal size after resize
3. Element tree layout is recalculated at new dimensions
4. Rapid resize events (window drag) are debounced - only one redraw per 16ms
5. Existing applications work without code changes
6. Tests verify resize behavior in isolation

---

## 7. Open Questions

1. ~~Should resize be automatic or manual?~~ → **Automatic** (user preference)
2. ~~Should resize events be debounced?~~ → **Yes, 16ms window** (user preference)
3. Should we emit the debounced ResizeEvent to apps or suppress it entirely? → **Emit it** - apps may want to react (e.g., update root size)
