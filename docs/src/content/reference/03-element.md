---
title: "Element"
order: 3
---

Core UI building block.

## New

```go
func New(opts ...Option) *Element
```

Creates a new Element with the given options.

## AddChild

```go
func (e *Element) AddChild(children ...*Element)
```

Appends child elements.

## RemoveChild

```go
func (e *Element) RemoveChild(child *Element) bool
```

Removes a child element. Returns true if found.

## SetText

```go
func (e *Element) SetText(content string)
```

Updates the text content and marks dirty.

## SetHidden

```go
func (e *Element) SetHidden(hidden bool)
```

Hides/shows the element from layout and rendering.

## MarkDirty

```go
func (e *Element) MarkDirty()
```

Marks the element for re-render on the next frame.
