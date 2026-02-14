package tui

import "github.com/grindlemire/go-tui/internal/debug"

// mountKey identifies a component instance by its parent and position.
// Components at the same (parent, index) are considered the same instance
// across renders and are reused from cache.
type mountKey struct {
	parent Component
	index  int
}

// mountState is per-App state for component instance caching.
// Stored on the App struct, accessed via the explicit *App reference during render.
// Uses mark-and-sweep: each render marks active keys, then sweep
// cleans up unmounted components.
type mountState struct {
	cache      map[mountKey]Component
	cleanups   map[mountKey]func()
	activeKeys map[mountKey]bool // Marked during render, swept after
}

// newMountState creates a new mountState with initialized maps.
func newMountState() *mountState {
	return &mountState{
		cache:      make(map[mountKey]Component),
		cleanups:   make(map[mountKey]func()),
		activeKeys: make(map[mountKey]bool),
	}
}

// PropsUpdater is an optional interface that components can implement
// to receive updated props when re-rendered from cache. Mount will call
// UpdateProps with a fresh instance containing the new props, allowing
// the cached instance to copy the relevant fields.
type PropsUpdater interface {
	UpdateProps(fresh Component)
}

// Mount creates or retrieves a cached component instance and returns
// its rendered element tree. Called by generated code from @Component() syntax.
//
// On first call: executes factory, caches instance, calls Init() if Initializer.
// On subsequent calls: returns cached instance's Render() result.
// If the cached instance implements PropsUpdater, UpdateProps is called
// with a fresh instance to allow prop updates.
// Mark-and-sweep: marks the key as active. Sweep after render cleans stale entries.
func (a *App) Mount(parent Component, index int, factory func() Component) *Element {
	app := a
	ms := app.mounts
	key := mountKey{parent: parent, index: index}
	ms.activeKeys[key] = true // Mark as active this render

	instance, cached := ms.cache[key]
	if !cached {
		instance = factory()
		ms.cache[key] = instance
		debug.Log("Mount: NEW component at index %d, type %T", index, instance)

		// Bind app before Init so state/events are wired up
		if binder, ok := instance.(AppBinder); ok {
			binder.BindApp(app)
		}

		// Call Init() if component implements Initializer
		if init, ok := instance.(Initializer); ok {
			cleanup := init.Init()
			if cleanup != nil {
				ms.cleanups[key] = cleanup
			}
		}
	} else {
		// Component is cached - check if it can receive updated props
		if updater, ok := instance.(PropsUpdater); ok {
			fresh := factory()
			debug.Log("Mount: CACHED component at index %d, calling UpdateProps, type %T", index, instance)
			updater.UpdateProps(fresh)
		} else {
			debug.Log("Mount: CACHED component at index %d, NO UpdateProps, type %T", index, instance)
		}
		// Rebind after props update — fresh Events fields may be unbound
		if binder, ok := instance.(AppBinder); ok {
			binder.BindApp(app)
		}
	}

	// Render the component and tag the element for framework discovery
	el := instance.Render(a)
	el.component = instance
	return el
}

// sweep removes cached instances that were not marked active during the last
// render pass. Calls cleanup functions for removed components.
func (ms *mountState) sweep() {
	for key := range ms.cache {
		if !ms.activeKeys[key] {
			if cleanup, ok := ms.cleanups[key]; ok {
				cleanup()
				delete(ms.cleanups, key)
			}
			delete(ms.cache, key)
		}
	}
	// Reset active keys for next render
	ms.activeKeys = make(map[mountKey]bool)
}
