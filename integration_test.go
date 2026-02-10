package tui

import "testing"

// --- Integration test components ---

// intRoot is a root component that mounts two children.
// Its KeyMap changes based on the searchActive state.
type intRoot struct {
	searchActive *State[bool]
	query        *State[string]
}

func newIntRoot() *intRoot {
	return &intRoot{
		searchActive: NewState(false),
		query:        NewState(""),
	}
}

func (r *intRoot) KeyMap() KeyMap {
	km := KeyMap{
		OnKey(KeyCtrlC, func(ke KeyEvent) {}),
	}
	if !r.searchActive.Get() {
		km = append(km, OnRune('/', func(ke KeyEvent) {
			r.searchActive.Set(true)
		}))
	}
	return km
}

func (r *intRoot) Render(app *App) *Element {
	root := New(WithDirection(Row))

	// Mount sidebar at index 0
	el0 := app.Mount(r, 0, func() Component {
		return newIntSidebar(r.query)
	})
	root.AddChild(el0)

	// Conditionally mount search at index 1
	el1 := app.Mount(r, 1, func() Component {
		return newIntSearch(r.searchActive, r.query)
	})
	root.AddChild(el1)

	return root
}

// intSidebar is a child component with a KeyCtrlB binding.
type intSidebar struct {
	query    *State[string]
	expanded *State[bool]
}

func newIntSidebar(query *State[string]) *intSidebar {
	return &intSidebar{
		query:    query,
		expanded: NewState(true),
	}
}

func (s *intSidebar) KeyMap() KeyMap {
	return KeyMap{
		OnKey(KeyCtrlB, func(ke KeyEvent) {
			s.expanded.Set(!s.expanded.Get())
		}),
	}
}

func (s *intSidebar) Render(app *App) *Element {
	return New(WithText("sidebar"))
}

// intSearch is a child component with conditional stop-propagation bindings.
type intSearch struct {
	active *State[bool]
	query  *State[string]
}

func newIntSearch(active *State[bool], query *State[string]) *intSearch {
	return &intSearch{active: active, query: query}
}

func (s *intSearch) KeyMap() KeyMap {
	if !s.active.Get() {
		return nil
	}
	return KeyMap{
		OnRunesStop(func(ke KeyEvent) {
			s.query.Set(s.query.Get() + string(ke.Rune))
		}),
		OnKeyStop(KeyEscape, func(ke KeyEvent) {
			s.active.Set(false)
			s.query.Set("")
		}),
	}
}

func (s *intSearch) Render(app *App) *Element {
	return New(WithText("search"))
}

// intInitComponent tracks Init/cleanup lifecycle.
type intInitComponent struct {
	initCalled   bool
	cleanupCalls int
}

func (c *intInitComponent) Render(app *App) *Element { return New() }
func (c *intInitComponent) Init() func() {
	c.initCalled = true
	return func() { c.cleanupCalls++ }
}

// --- Integration tests ---

func TestIntegration_MountCachesAndDiscoverKeyMaps(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	root := newIntRoot()
	el := root.Render(DefaultApp())
	el.component = root

	// Verify mount cached two child instances
	ms := DefaultApp().mounts
	if len(ms.cache) != 2 {
		t.Fatalf("mount cache has %d entries, want 2", len(ms.cache))
	}

	// Build dispatch table from the rendered tree
	table, err := buildDispatchTable(el)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}

	// Root has: ctrl+c, '/' (since searchActive=false)
	// Sidebar has: ctrl+b
	// Search has: nil (since active=false)
	// Total: 3 entries
	if len(table.entries) != 3 {
		t.Fatalf("dispatch table has %d entries, want 3", len(table.entries))
	}
}

func TestIntegration_DispatchBroadcastAndStopPropagation(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	root := newIntRoot()
	el := root.Render(DefaultApp())
	el.component = root

	table, err := buildDispatchTable(el)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}

	// Press '/': root handler fires, searchActive becomes true
	table.dispatch(KeyEvent{Key: KeyRune, Rune: '/'})
	if !root.searchActive.Get() {
		t.Error("expected searchActive=true after '/' dispatch")
	}
}

func TestIntegration_ConditionalKeyMapActivation(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()
	resetDirty()

	root := newIntRoot()

	// Initial render: search is inactive
	el := root.Render(DefaultApp())
	el.component = root

	table, err := buildDispatchTable(el)
	if err != nil {
		t.Fatalf("buildDispatchTable (initial): %v", err)
	}

	// Press '/' to activate search
	table.dispatch(KeyEvent{Key: KeyRune, Rune: '/'})
	if !root.searchActive.Get() {
		t.Fatal("expected searchActive=true")
	}

	// Re-render (simulating dirty frame)
	el = root.Render(DefaultApp())
	el.component = root

	// Rebuild dispatch table with new KeyMaps
	table, err = buildDispatchTable(el)
	if err != nil {
		t.Fatalf("buildDispatchTable (after activation): %v", err)
	}

	// Now search should have stop-propagation bindings:
	// Root: ctrl+c (no '/' since searchActive=true)
	// Sidebar: ctrl+b
	// Search: AnyRune(stop), Escape(stop)
	// Total: 4 entries
	if len(table.entries) != 4 {
		t.Fatalf("dispatch table has %d entries after activation, want 4", len(table.entries))
	}

	// Type a character — should go to search exclusively (stop propagation)
	table.dispatch(KeyEvent{Key: KeyRune, Rune: 'h'})
	if root.query.Get() != "h" {
		t.Errorf("query = %q, want %q", root.query.Get(), "h")
	}

	// Type more
	table.dispatch(KeyEvent{Key: KeyRune, Rune: 'i'})
	if root.query.Get() != "hi" {
		t.Errorf("query = %q, want %q", root.query.Get(), "hi")
	}
}

func TestIntegration_EscapeDeactivatesSearch(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()
	resetDirty()

	root := newIntRoot()

	// Activate search
	root.searchActive.Set(true)
	root.query.Set("test")

	el := root.Render(DefaultApp())
	el.component = root

	table, err := buildDispatchTable(el)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}

	// Press Escape — search deactivates
	table.dispatch(KeyEvent{Key: KeyEscape})

	if root.searchActive.Get() {
		t.Error("expected searchActive=false after Escape")
	}
	if root.query.Get() != "" {
		t.Errorf("expected query=%q after Escape, got %q", "", root.query.Get())
	}

	// Re-render: search should return nil KeyMap
	el = root.Render(DefaultApp())
	el.component = root

	table, err = buildDispatchTable(el)
	if err != nil {
		t.Fatalf("buildDispatchTable (after deactivation): %v", err)
	}

	// Root: ctrl+c, '/' (back since searchActive=false)
	// Sidebar: ctrl+b
	// Search: nil
	// Total: 3
	if len(table.entries) != 3 {
		t.Fatalf("dispatch table has %d entries after deactivation, want 3", len(table.entries))
	}
}

func TestIntegration_SweepCleansUnmountedComponents(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	parent := &mockParent{}
	initComp := &intInitComponent{}

	// Mount the component
	DefaultApp().Mount(parent, 0, func() Component { return initComp })

	if !initComp.initCalled {
		t.Fatal("Init should have been called on first mount")
	}

	ms := DefaultApp().mounts

	// Simulate a render where the component is not active (e.g., removed by @if)
	ms.activeKeys = make(map[mountKey]bool)
	ms.sweep()

	if initComp.cleanupCalls != 1 {
		t.Errorf("cleanup called %d times, want 1", initComp.cleanupCalls)
	}
	if len(ms.cache) != 0 {
		t.Errorf("cache has %d entries after sweep, want 0", len(ms.cache))
	}
}

func TestIntegration_SharedStatePropagation(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()
	resetDirty()

	root := newIntRoot()

	// Initial render
	el := root.Render(DefaultApp())
	el.component = root

	// The query state is shared between root, sidebar, and search.
	// Setting it from search should be visible to sidebar.
	root.query.Set("hello")

	if root.query.Get() != "hello" {
		t.Errorf("query = %q, want %q", root.query.Get(), "hello")
	}

	// Re-render and verify the shared state is accessible
	el = root.Render(DefaultApp())
	el.component = root

	// walkComponents should find all 3 components
	var found []Component
	walkComponents(el, func(c Component) {
		found = append(found, c)
	})

	// root + sidebar + search = 3
	if len(found) != 3 {
		t.Fatalf("walkComponents found %d components, want 3", len(found))
	}
}

func TestIntegration_DispatchTableRebuiltOnStateChange(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()
	resetDirty()

	root := newIntRoot()

	// Phase 1: searchActive=false
	el := root.Render(DefaultApp())
	el.component = root

	table1, err := buildDispatchTable(el)
	if err != nil {
		t.Fatalf("buildDispatchTable phase 1: %v", err)
	}
	count1 := len(table1.entries)

	// Phase 2: activate search
	root.searchActive.Set(true)
	el = root.Render(DefaultApp())
	el.component = root

	table2, err := buildDispatchTable(el)
	if err != nil {
		t.Fatalf("buildDispatchTable phase 2: %v", err)
	}
	count2 := len(table2.entries)

	// Phase 1: 3 entries (ctrl+c, '/', ctrl+b)
	// Phase 2: 4 entries (ctrl+c, ctrl+b, AnyRune(stop), Escape(stop))
	if count1 != 3 {
		t.Errorf("phase 1 entry count = %d, want 3", count1)
	}
	if count2 != 4 {
		t.Errorf("phase 2 entry count = %d, want 4", count2)
	}

	// Phase 2 should have different entries than phase 1
	if count1 == count2 {
		t.Error("dispatch table should change after state change and re-render")
	}
}

func TestIntegration_CtrlBTogglesSidebar(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	root := newIntRoot()
	el := root.Render(DefaultApp())
	el.component = root

	table, err := buildDispatchTable(el)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}

	// Find the sidebar component in the cache
	var sidebar *intSidebar
	walkComponents(el, func(c Component) {
		if s, ok := c.(*intSidebar); ok {
			sidebar = s
		}
	})
	if sidebar == nil {
		t.Fatal("sidebar component not found in tree")
	}

	if !sidebar.expanded.Get() {
		t.Error("sidebar should start expanded")
	}

	// Press Ctrl+B to toggle sidebar
	table.dispatch(KeyEvent{Key: KeyCtrlB})

	if sidebar.expanded.Get() {
		t.Error("sidebar should be collapsed after Ctrl+B")
	}

	// Press again to toggle back
	table.dispatch(KeyEvent{Key: KeyCtrlB})

	if !sidebar.expanded.Get() {
		t.Error("sidebar should be expanded after second Ctrl+B")
	}
}
