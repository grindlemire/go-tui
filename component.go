package tui

// Component is the base interface for struct components.
// Any struct with a Render(app *App) method can be used as a component in the element tree.
// Components are instantiated by constructor functions and cached by the Mount system.
type Component interface {
	Render(app *App) *Element
}

// KeyListener is implemented by components that handle keyboard input.
// KeyMap() returns the current set of key bindings. It is called on every
// tree walk (when dirty), so it can return different bindings based on state.
type KeyListener interface {
	KeyMap() KeyMap
}

// MouseListener is implemented by components that handle mouse input.
// HandleMouse receives mouse events (clicks, wheel, etc.) and returns
// true if the event was consumed. Like KeyListener, it is discovered by
// walking the component tree.
type MouseListener interface {
	HandleMouse(MouseEvent) bool
}

// Initializer is implemented by components that need setup when first mounted.
// Init() is called once when the component first enters the tree.
// The returned function (if non-nil) is called when the component leaves
// the tree. This pairs setup and cleanup at the same call site.
type Initializer interface {
	Init() func()
}

// WatcherProvider is an optional interface for components that provide
// timers, tickers, or channel watchers. Watchers() is called after the
// component is mounted and the returned watchers are started.
type WatcherProvider interface {
	Watchers() []Watcher
}

// AppBinder is implemented by components that contain State or Events fields
// needing an App reference. Generated code emits BindApp methods that
// delegate to each field's BindApp. The mount system calls BindApp
// automatically — users never call it directly.
type AppBinder interface {
	BindApp(app *App)
}
