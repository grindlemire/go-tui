package tui

import (
	"reflect"

	"github.com/grindlemire/go-tui/internal/debug"
)

// mountKey identifies a component instance by its parent and a comparable
// key. Components at the same (parent, key) are considered the same
// instance across renders and are reused from cache. Generated code builds
// keys with MountKey (loop call sites) or a plain int (standalone sites).
type mountKey struct {
	parent Component
	key    any
}

// mountState is per-App state for component instance caching.
// Stored on the App struct, accessed via the explicit *App reference during render.
// Uses mark-and-sweep: each render marks active keys, then sweep
// cleans up unmounted components.
type mountState struct {
	cache       map[mountKey]Component
	cleanups    map[mountKey]func()
	activeKeys  map[mountKey]bool // Marked during render, swept after
	persistKeys map[mountKey]bool // keys that survive sweep even when not rendered
}

// newMountState creates a new mountState with initialized maps.
func newMountState() *mountState {
	return &mountState{
		cache:       make(map[mountKey]Component),
		cleanups:    make(map[mountKey]func()),
		activeKeys:  make(map[mountKey]bool),
		persistKeys: make(map[mountKey]bool),
	}
}

// PropsUpdater is an optional interface that components can implement
// to receive updated props when re-rendered from cache. Mount will call
// UpdateProps with a fresh instance containing the new props, allowing
// the cached instance to copy the relevant fields.
type PropsUpdater interface {
	UpdateProps(fresh Component)
}

// mount is the shared implementation for Mount and MountPersistent.
func (a *App) mount(parent Component, key any, factory func() Component) *Element {
	app := a
	ms := app.mounts
	k := mountKey{parent: parent, key: key}
	if ms.activeKeys[k] {
		// Two mounts produced the same key within one render pass: either
		// two call sites collide or a key={...} expression repeats across
		// loop items. Both tree positions will share one component instance.
		debug.Log("Mount: duplicate mount key %v (parent %T) within one render pass", key, parent)
	}
	ms.activeKeys[k] = true // Mark as active this render

	instance, cached := ms.cache[k]
	var fresh Component
	if cached {
		// The collision guard below only covers PropsUpdater types: that is
		// the one path where a fresh instance already exists to compare
		// against, so the check is free. All framework components implement
		// PropsUpdater. Checking every cached component would cost a
		// factory() allocation per cache hit per render.
		if _, ok := instance.(PropsUpdater); ok {
			fresh = factory()
			if reflect.TypeOf(fresh) != reflect.TypeOf(instance) {
				// Key collision: this call site produced a different
				// component type than the cache holds. Render the right
				// component loudly instead of the wrong one silently.
				debug.Log("Mount: key collision at %v: cached %T, factory produced %T; remounting", key, instance, fresh)
				ms.evict(k)
				cached = false
			}
		}
	}

	if !cached {
		if fresh != nil {
			instance = fresh
		} else {
			instance = factory()
		}
		ms.cache[k] = instance
		debug.Log("Mount: NEW component at key %v, type %T", key, instance)

		// Bind app before Init so state/events are wired up
		if binder, ok := instance.(AppBinder); ok {
			binder.BindApp(app)
		}

		// Call Init() if component implements Initializer
		if init, ok := instance.(Initializer); ok {
			cleanup := init.Init()
			if cleanup != nil {
				ms.cleanups[k] = cleanup
			}
		}
	} else {
		// Component is cached - check if it can receive updated props.
		// fresh was already constructed (and type-checked) above.
		if updater, ok := instance.(PropsUpdater); ok {
			debug.Log("Mount: CACHED component at key %v, calling UpdateProps, type %T", key, instance)
			updater.UpdateProps(fresh)
		} else {
			debug.Log("Mount: CACHED component at key %v, NO UpdateProps, type %T", key, instance)
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

// Mount creates or retrieves a cached component instance and returns
// its rendered element tree. Called by generated code from @Component() syntax.
//
// On first call: executes factory, caches instance, calls Init() if Initializer.
// On subsequent calls: returns cached instance's Render() result.
// If the cached instance implements PropsUpdater, UpdateProps is called
// with a fresh instance to allow prop updates.
// Mark-and-sweep: marks the key as active. Sweep after render cleans stale entries.
// The key must be comparable; generated code passes a plain int for
// standalone call sites and a MountKey composite inside loops.
func (a *App) Mount(parent Component, key any, factory func() Component) *Element {
	return a.mount(parent, key, factory)
}

// MountPersistent is like Mount but marks the component as persistent,
// preventing it from being cleaned up during sweep even when not active.
// Use this for components that must survive being hidden by conditionals.
func (a *App) MountPersistent(parent Component, key any, factory func() Component) *Element {
	k := mountKey{parent: parent, key: key}
	a.mounts.persistKeys[k] = true
	return a.mount(parent, key, factory)
}

// sweep removes cached instances that were not marked active during the last
// render pass. Calls cleanup functions for removed components.
func (ms *mountState) sweep() {
	for key := range ms.cache {
		if !ms.activeKeys[key] && !ms.persistKeys[key] {
			ms.evict(key)
		}
	}
	// Reset active keys for next render
	ms.activeKeys = make(map[mountKey]bool)
}

// evict removes one cached instance, unbinding it and running its cleanup.
func (ms *mountState) evict(key mountKey) {
	if instance, ok := ms.cache[key]; ok {
		if unbinder, ok := instance.(AppUnbinder); ok {
			unbinder.UnbindApp()
		}
	}
	if cleanup, ok := ms.cleanups[key]; ok {
		cleanup()
		delete(ms.cleanups, key)
	}
	delete(ms.cache, key)
}
