---
title: "App"
order: 1
---

Application lifecycle and configuration.

## NewApp

```go
func NewApp(opts ...AppOption) (*App, error)
```

Creates a new terminal application with the given options.

## App.Run

```go
func (a *App) Run() error
```

Starts the main event loop. Blocks until Stop() is called or SIGINT received.

## App.Stop

```go
func (a *App) Stop()
```

Signals the Run loop to exit gracefully. Idempotent.

## App.Close

```go
func (a *App) Close()
```

Restores the terminal to its original state. Always defer this.

## App.SetRoot

```go
func (a *App) SetRoot(root *Element)
```

Sets the root element of the UI tree.

## App.SetRootComponent

```go
func (a *App) SetRootComponent(c Renderable)
```

Sets a Renderable component as root. Handles watchers and focus registration.

## App.Render

```go
func (a *App) Render()
```

Triggers a layout + render pass if dirty.

## App.QueueUpdate

```go
func (a *App) QueueUpdate(fn func())
```

Enqueues a function to run on the main loop. Safe from any goroutine.

## App.Dispatch

```go
func (a *App) Dispatch(event Event) bool
```

Routes an event through the focus manager.

## App.Size

```go
func (a *App) Size() (int, int)
```

Returns the current terminal width and height.

## App.PrintAboveln

```go
func (a *App) PrintAboveln(format string, args ...any)
```

Prints a line above the inline widget (inline mode only).
