---
title: "Interfaces"
order: 8
---

Key interfaces for the component model.

## Renderable

```go
interface { Render() *Element }
```

Components implement this to produce their Element tree.

## WatcherProvider

```go
interface { Watchers() []Watcher }
```

Components that provide timers/channel watchers.

## KeyMapProvider

```go
interface { KeyMap() KeyMap }
```

Components that provide keyboard bindings.

## Focusable

```go
interface { IsFocusable() bool; Focus(); Blur(); HandleEvent(Event) bool }
```

Elements that can receive focus and handle events.
