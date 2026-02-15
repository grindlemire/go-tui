---
title: "App Options"
order: 2
---

Configuration options for NewApp.

## WithFrameRate

```go
func WithFrameRate(fps int) AppOption
```

Sets the target frame rate (default 60).

## WithMouseEnabled

```go
func WithMouseEnabled() AppOption
```

Enables mouse event reporting.

## WithInlineHeight

```go
func WithInlineHeight(h int) AppOption
```

Enables inline mode with the given height instead of fullscreen.
