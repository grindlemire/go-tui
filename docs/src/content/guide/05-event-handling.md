---
title: "Event Handling"
order: 5
---

## Keyboard Events

Handle keyboard input with KeyMap for declarative key bindings:

```go
func (c *counter) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) {
            ke.App().Stop()
        }),
        tui.OnRune('+', func(ke tui.KeyEvent) {
            c.count.Set(c.count.Get() + 1)
        }),
        tui.OnRune('-', func(ke tui.KeyEvent) {
            c.count.Set(c.count.Get() - 1)
        }),
    }
}
```

## Mouse Events

Handle mouse clicks with ref-based hit testing:

```go
func (c *counter) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(c.incrementBtn, c.increment),
        tui.Click(c.decrementBtn, c.decrement),
    )
}
```
