---
title: Mental Model
description: Learn the core runtime loop in go-tui.
---

# Mental Model

## What you'll learn
- The runtime flow of a `go-tui` app.
- Where state updates happen.
- What causes rerenders.

## Prerequisites
- Read [First Project](../getting-started/first-project).

## Steps
1. `App` starts and mounts your root component.
2. Your component renders an element tree.
3. Keyboard or mouse events call your handlers.
4. Handlers update state and trigger rerenders.
5. The renderer writes the updated frame to the terminal.

## Example
```text
Key press -> KeyMap handler -> state update -> Render() -> terminal refresh
```

```go
func (c *counter) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('+', func(tui.KeyEvent) { c.value.Set(c.value.Get() + 1) }),
		tui.OnRune('-', func(tui.KeyEvent) { c.value.Set(c.value.Get() - 1) }),
	}
}
```

## Recap
Think in a loop: input, state change, render, paint.

## Next
Next page: `using-tools/find-tools` (to be added).
