---
title: "Events"
order: 6
---

Input event types and handling.

## KeyEvent

```go
type KeyEvent struct { Key Key; Rune rune; Mod Modifier }
```

Represents a keyboard event with key, rune, and modifiers.

## MouseEvent

```go
type MouseEvent struct { X, Y int; Button MouseButton; Action MouseAction }
```

Represents a mouse event with position, button, and action.

## ResizeEvent

```go
type ResizeEvent struct { Width, Height int }
```

Fired when the terminal is resized.

## KeyMap

```go
type KeyMap []KeyBinding
```

Declarative key binding list for components.

## OnKey

```go
func OnKey(key Key, fn func(KeyEvent)) KeyBinding
```

Creates a key binding for a special key.

## OnRune

```go
func OnRune(r rune, fn func(KeyEvent)) KeyBinding
```

Creates a key binding for a character.

## HandleClicks

```go
func HandleClicks(me MouseEvent, clicks ...ClickHandler) bool
```

Tests mouse event against ref-based click handlers. Returns true if handled.
