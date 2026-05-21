# Event Dump

## Overview

This is a live event log that prints every keystroke, mouse click, drag, wheel tick, and resize the app receives. This is useful as a general input-handling smoke test on any new terminal or platform.

## What the status line tells you

A yellow status line at the top of the UI reads:

```
STATUS: stateY=N elY=N  following=BOOL  entries=N  contentH=N viewportH=N
```

- `stateY` is what the component's `*State[int]` thinks the scroll offset should be.
- `elY` is the actual scroll offset of the rendered element from the previous frame.
- `following` is `true` while the viewport pins to the bottom for new entries, `false` once you scroll away.
- `contentH` and `viewportH` are the laid-out heights. Their difference is the maximum scroll.

When the two scroll values disagree, the state binding to the element is broken. When they agree but the viewport still looks wrong, the bug is in the render path.

## Keyboard

Special keys (arrows, F1 through F12, Home, End, PageUp, PageDown, Tab, Enter, Escape, Backspace, Delete, Insert) each log their own `Key` constant. Modifiers combine with letters. `Ctrl+a` logs `Mod=Ctrl`. `Alt+letter` logs `Mod=Alt`. Shifted printable characters log the uppercase rune with `Mod=None`. The terminal does not emit `ModShift` for character-forming keys because the rune itself already encodes case. Modifier-only presses (Shift alone, Ctrl alone, Alt alone) do not log. The reader drops events with no useful payload.

Cmd on macOS does not reach the TUI. Terminal.app and iTerm2 both capture Cmd as their own command modifier, and the xterm wire format has no encoding for Cmd. Pressing Cmd-anything shows up as the underlying key with `Mod=None`.

## Mouse

Left, right, and middle button clicks log `Press` then `Release` at the same coordinates. Click-and-drag logs `Drag` events while moving with a button held. Wheel up and wheel down log as single `Press` events with the correct direction. Hover without buttons held does not log. Tracking pure motion would flood the log.

On the default Mac Terminal, the wheel scrolls Terminal's own scrollback buffer instead of forwarding to the app. Use iTerm2, Alacritty, or Kitty if you need to verify wheel events end-to-end.

## Resize

The framework does not route `ResizeEvent` to component listeners, so resize is not logged directly. To verify, drag the terminal corner. The outer border and the log viewport should reflow live, without flickering or losing entries.

## Code

The component captures all input through `AnyKey` and a mouse listener:

```gsx
type eventDump struct {
    log       *tui.State[[]string]
    scrollY   *tui.State[int]
    following *tui.State[bool]
    seq       *tui.State[int]
    feed      *tui.Ref
}

func (e *eventDump) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnStop(tui.KeyEscape, func(ke tui.KeyEvent) {
            e.append(formatKey(ke) + "  [QUIT]")
            ke.App().Stop()
        }),
        tui.OnStop(tui.Rune('q'), func(ke tui.KeyEvent) {
            e.append(formatKey(ke) + "  [QUIT]")
            ke.App().Stop()
        }),
        tui.On(tui.AnyKey, func(ke tui.KeyEvent) {
            e.append(formatKey(ke))
        }),
    }
}

func (e *eventDump) HandleMouse(me tui.MouseEvent) bool {
    e.append(formatMouse(me))
    switch me.Button {
    case tui.MouseWheelUp:
        e.scrollBy(-3)
        return true
    case tui.MouseWheelDown:
        e.scrollBy(3)
        return true
    }
    return false
}
```

`AnyKey` matches every key event including specials, so the broadcast handler logs all of them. The two `OnStop` handlers run before the broadcast handler and stop propagation, so Esc and q do not double-log.

Scroll position lives in state, and the follow flag decides when to auto-scroll:

```go
func (e *eventDump) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnChange(e.log, e.autoScrollToBottom),
    }
}

func (e *eventDump) autoScrollToBottom(_ []string) {
    if !e.following.Get() {
        return
    }
    el := e.feed.El()
    if el == nil {
        return
    }
    _, maxY := el.MaxScroll()
    e.scrollY.Set(maxY)
}

func (e *eventDump) scrollBy(delta int) {
    el := e.feed.El()
    if el == nil {
        return
    }
    _, maxY := el.MaxScroll()
    newY := e.scrollY.Get() + delta
    if newY < 0 {
        newY = 0
    }
    if newY > maxY {
        newY = maxY
    }
    e.scrollY.Set(newY)
    e.following.Set(newY >= maxY)
}
```

`OnChange` fires whenever the log changes. When `following` is true, the watcher pins `scrollY` to the bottom. The wheel handler calls `scrollBy` after logging the event, which clamps `scrollY` to a valid range and updates `following` based on whether the new position is at the bottom.

## Run

```bash
go run .
```

Press any key to log it. Click, drag, or scroll the mouse. Drag the terminal corner. Esc or q to quit.

## Next Steps

- [Events](events) for the keyboard and mouse handling primitives the example uses.
- [Scrolling](scrolling) for the scrollable container and scroll-offset binding pattern.
- [Watchers](watchers) for the `OnChange` watcher used to drive auto-scroll.
