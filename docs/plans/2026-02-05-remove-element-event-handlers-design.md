# Remove Element-Level Event Handlers

## Summary

Move all event handling from elements to components. Elements become pure layout/rendering; components handle interaction via `KeyMap()` and `HandleMouse()`.

## What Gets Removed

### From Element struct and options
- `onEvent` field and `WithOnEvent()` / `SetOnEvent()`
- `onKeyPress` field and `WithOnKeyPress()` / `SetOnKeyPress()`
- `onClick` field and `WithOnClick()` / `SetOnClick()`
- `bubbleOnEvent()` helper
- Event bubbling logic in `HandleEvent()`

### From GSX/codegen
- `onEvent`, `onKeyPress`, `onClick` attributes
- Handler attribute generation in `generator_element.go`
- Validation in `analyzer.go`
- LSP schema entries

## New Patterns

### Event Inspection via KeyMap

```go
func (a *interactiveApp) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        // Catches all printable chars
        tui.OnRunes(func(ke tui.KeyEvent) {
            a.lastEvent.Set(fmt.Sprintf("Key '%c'", ke.Rune))
        }),
        // Special keys need individual handlers
        tui.OnKey(tui.KeyEnter, a.inspectKey("Enter")),
        tui.OnKey(tui.KeyUp, a.inspectKey("Up")),
    }
}
```

### Button Clicks via HandleMouse

```go
func (a *interactiveApp) HandleMouse(me tui.MouseEvent) bool {
    if me.Button == tui.MouseLeft && me.Action == tui.MousePress {
        if a.decrementBtn.El().ContainsPoint(me.X, me.Y) {
            a.decrement()
            return true
        }
    }
    return false
}
```

## Simplified FocusManager/HandleEvent

- FocusManager still tracks focus for visual styling and Tab navigation
- `Element.HandleEvent()` only handles scroll events for scrollable elements
- No more event bubbling

## Files to Modify

### Core
- `element.go` - Remove handler fields
- `element_focus.go` - Remove setters, simplify HandleEvent
- `element_options.go` - Remove With* options

### Codegen
- `internal/tuigen/generator_element.go` - Remove handler generation
- `internal/tuigen/analyzer.go` - Remove handler validation

### LSP
- `internal/lsp/schema/schema.go` - Remove handler attributes

### Examples
- `examples/06-interactive/` - Use KeyMap + HandleMouse
- `examples/09-scrollable/` - Remove onEvent
- `examples/11-streaming/` - Remove onEvent/onKeyPress

### Tests
- `element_focus_test.go` - Remove handler tests
