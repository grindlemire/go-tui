package tui

// mountKey identifies a component instance by its parent and position.
// Components at the same (parent, index) are considered the same instance
// across renders and are reused from cache.
type mountKey struct {
	parent Component
	index  int
}

// mountState is per-App state for component instance caching.
// Stored on the App struct, accessed via currentApp during render.
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

// Mount creates or retrieves a cached component instance and returns
// its rendered element tree. Called by generated code from @Component() syntax.
//
// On first call: executes factory, caches instance, calls Init() if Initializer.
// On subsequent calls: returns cached instance's Render() result.
// Mark-and-sweep: marks the key as active. Sweep after render cleans stale entries.
func Mount(parent Component, index int, factory func() Component) *Element {
	ms := currentApp.mounts
	key := mountKey{parent: parent, index: index}
	ms.activeKeys[key] = true // Mark as active this render

	instance, cached := ms.cache[key]
	if !cached {
		instance = factory()
		ms.cache[key] = instance

		// Call Init() if component implements Initializer
		if init, ok := instance.(Initializer); ok {
			cleanup := init.Init()
			if cleanup != nil {
				ms.cleanups[key] = cleanup
			}
		}
	}

	// Render the component and tag the element for framework discovery
	el := instance.Render()
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
