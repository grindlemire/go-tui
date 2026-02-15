---
title: "State"
order: 5
---

Reactive state management.

## NewState

```go
func NewState[T any](initial T) *State[T]
```

Creates a new reactive state with an initial value.

## State.Get

```go
func (s *State[T]) Get() T
```

Returns the current value.

## State.Set

```go
func (s *State[T]) Set(v T)
```

Sets the value and marks dirty if changed.

## State.Bind

```go
func (s *State[T]) Bind(fn func(T))
```

Registers a callback invoked when the value changes.

## Batch

```go
func Batch(fn func())
```

Executes fn, coalescing all state changes into a single re-render.
