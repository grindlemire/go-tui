package tui

import "testing"

// --- Mock components for testing ---

// mockComponent implements Component only.
type mockComponent struct {
	renderCount int
}

func (m *mockComponent) Render(app *App) *Element {
	m.renderCount++
	return New()
}

// mockInitComponent implements Component and Initializer.
type mockInitComponent struct {
	renderCount  int
	initCalled   bool
	cleanupCalls int
}

func (m *mockInitComponent) Render(app *App) *Element {
	m.renderCount++
	return New()
}

func (m *mockInitComponent) Init() func() {
	m.initCalled = true
	return func() {
		m.cleanupCalls++
	}
}

// mockInitNoCleanup implements Initializer but returns nil cleanup.
type mockInitNoCleanup struct {
	initCalled bool
}

func (m *mockInitNoCleanup) Render(app *App) *Element {
	return New()
}

func (m *mockInitNoCleanup) Init() func() {
	m.initCalled = true
	return nil
}

// mockParent implements Component (used as mount parent).
// Has a field to ensure distinct instances have different addresses.
type mockParent struct {
	id int
}

func (m *mockParent) Render(app *App) *Element {
	return New()
}

// --- Helpers ---

// setupTestMountState sets up DefaultApp with a fresh mountState for testing.
// Returns a cleanup function to restore the previous state.
func setupTestMountState() func() {
	prev := DefaultApp()
	SetDefaultApp(&App{
		mounts: newMountState(),
	})
	return func() {
		SetDefaultApp(prev)
	}
}

// --- Tests ---

func TestMount_FirstCallCreatesInstance(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	parent := &mockParent{}
	factoryCalls := 0
	var created *mockComponent

	el := DefaultApp().Mount(parent, 0, func() Component {
		factoryCalls++
		created = &mockComponent{}
		return created
	})

	if factoryCalls != 1 {
		t.Errorf("factory called %d times, want 1", factoryCalls)
	}
	if created == nil {
		t.Fatal("factory did not create component")
	}
	if created.renderCount != 1 {
		t.Errorf("Render called %d times, want 1", created.renderCount)
	}
	if el == nil {
		t.Fatal("Mount returned nil element")
	}
	if el.component != created {
		t.Error("element.component not set to created instance")
	}
}

func TestMount_SubsequentCallReturnsCached(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	parent := &mockParent{}
	factoryCalls := 0
	instance := &mockComponent{}

	// First mount
	DefaultApp().Mount(parent, 0, func() Component {
		factoryCalls++
		return instance
	})

	// Second mount — same parent and index
	el := DefaultApp().Mount(parent, 0, func() Component {
		factoryCalls++
		return &mockComponent{} // Would create a new one, but shouldn't be called
	})

	if factoryCalls != 1 {
		t.Errorf("factory called %d times, want 1 (should use cache)", factoryCalls)
	}
	if instance.renderCount != 2 {
		t.Errorf("Render called %d times, want 2 (once per Mount call)", instance.renderCount)
	}
	if el.component != instance {
		t.Error("element.component should be the cached instance")
	}
}

func TestMount_InitCalledOnFirstMount(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	parent := &mockParent{}
	instance := &mockInitComponent{}

	DefaultApp().Mount(parent, 0, func() Component {
		return instance
	})

	if !instance.initCalled {
		t.Error("Init() should be called on first mount")
	}
}

func TestMount_InitNotCalledOnSubsequentMount(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	parent := &mockParent{}
	instance := &mockInitComponent{}

	// First mount — triggers Init
	DefaultApp().Mount(parent, 0, func() Component {
		return instance
	})
	instance.initCalled = false // Reset flag

	// Second mount — should NOT call Init again
	DefaultApp().Mount(parent, 0, func() Component {
		return instance
	})

	if instance.initCalled {
		t.Error("Init() should not be called on subsequent mount")
	}
}

func TestMount_NilCleanupHandled(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	parent := &mockParent{}
	instance := &mockInitNoCleanup{}

	DefaultApp().Mount(parent, 0, func() Component {
		return instance
	})

	if !instance.initCalled {
		t.Error("Init() should be called")
	}

	// Sweep should not panic when cleanup is nil
	ms := DefaultApp().mounts
	// Don't mark key as active so sweep removes it
	ms.activeKeys = make(map[mountKey]bool)
	ms.sweep() // Should not panic
}

func TestMount_DifferentKeysIndependent(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	parent := &mockParent{}
	instanceA := &mockComponent{}
	instanceB := &mockComponent{}

	// Mount at index 0
	DefaultApp().Mount(parent, 0, func() Component { return instanceA })
	// Mount at index 1
	DefaultApp().Mount(parent, 1, func() Component { return instanceB })

	if instanceA.renderCount != 1 {
		t.Errorf("instanceA.renderCount = %d, want 1", instanceA.renderCount)
	}
	if instanceB.renderCount != 1 {
		t.Errorf("instanceB.renderCount = %d, want 1", instanceB.renderCount)
	}
}

func TestMount_DifferentParentsIndependent(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	parentA := &mockParent{id: 1}
	parentB := &mockParent{id: 2}
	factoryCalls := 0

	// Mount with parentA at index 0
	DefaultApp().Mount(parentA, 0, func() Component {
		factoryCalls++
		return &mockComponent{}
	})
	// Mount with parentB at same index — different parent, so new instance
	DefaultApp().Mount(parentB, 0, func() Component {
		factoryCalls++
		return &mockComponent{}
	})

	if factoryCalls != 2 {
		t.Errorf("factory called %d times, want 2 (different parents)", factoryCalls)
	}
}

func TestMountState_SweepCleansInactive(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	parent := &mockParent{}
	instance := &mockInitComponent{}

	// Mount to populate cache
	DefaultApp().Mount(parent, 0, func() Component { return instance })

	ms := DefaultApp().mounts

	// Simulate a render where this component is NOT active
	ms.activeKeys = make(map[mountKey]bool) // Nothing active
	ms.sweep()

	// Cleanup should have been called
	if instance.cleanupCalls != 1 {
		t.Errorf("cleanup called %d times, want 1", instance.cleanupCalls)
	}

	// Cache should be empty
	if len(ms.cache) != 0 {
		t.Errorf("cache has %d entries, want 0 after sweep", len(ms.cache))
	}
}

func TestMountState_SweepKeepsActive(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	parent := &mockParent{}
	instance := &mockInitComponent{}

	// Mount to populate cache and mark as active
	DefaultApp().Mount(parent, 0, func() Component { return instance })

	ms := DefaultApp().mounts

	// Sweep — key was marked active by DefaultApp().Mount(), so it should survive
	ms.sweep()

	if instance.cleanupCalls != 0 {
		t.Errorf("cleanup called %d times, want 0 (component still active)", instance.cleanupCalls)
	}
	if len(ms.cache) != 1 {
		t.Errorf("cache has %d entries, want 1 after sweep", len(ms.cache))
	}
}

func TestMountState_SweepResetsActiveKeys(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	parent := &mockParent{}

	DefaultApp().Mount(parent, 0, func() Component { return &mockComponent{} })

	ms := DefaultApp().mounts

	if len(ms.activeKeys) != 1 {
		t.Fatalf("activeKeys has %d entries before sweep, want 1", len(ms.activeKeys))
	}

	ms.sweep()

	if len(ms.activeKeys) != 0 {
		t.Errorf("activeKeys has %d entries after sweep, want 0 (reset for next render)", len(ms.activeKeys))
	}
}

func TestMountState_SweepMultipleComponents(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	parent := &mockParent{}
	active := &mockInitComponent{}
	inactive := &mockInitComponent{}

	// Mount both
	DefaultApp().Mount(parent, 0, func() Component { return active })
	DefaultApp().Mount(parent, 1, func() Component { return inactive })

	ms := DefaultApp().mounts

	// Simulate next render where only index 0 is active
	ms.activeKeys = make(map[mountKey]bool)
	key0 := mountKey{parent: parent, index: 0}
	ms.activeKeys[key0] = true

	ms.sweep()

	// active should survive
	if active.cleanupCalls != 0 {
		t.Errorf("active.cleanupCalls = %d, want 0", active.cleanupCalls)
	}
	// inactive should be cleaned up
	if inactive.cleanupCalls != 1 {
		t.Errorf("inactive.cleanupCalls = %d, want 1", inactive.cleanupCalls)
	}
	if len(ms.cache) != 1 {
		t.Errorf("cache has %d entries, want 1", len(ms.cache))
	}
}

func TestWalkComponents_FindsComponents(t *testing.T) {
	comp1 := &mockComponent{}
	comp2 := &mockComponent{}

	// Build a tree: root -> child1 (has component) -> child2 (has component)
	child2 := New()
	child2.component = comp2

	child1 := New()
	child1.component = comp1
	child1.AddChild(child2)

	root := New()
	root.AddChild(child1)

	var found []Component
	walkComponents(root, func(c Component) {
		found = append(found, c)
	})

	if len(found) != 2 {
		t.Fatalf("walkComponents found %d components, want 2", len(found))
	}
	if found[0] != comp1 {
		t.Error("first component should be comp1 (DFS order)")
	}
	if found[1] != comp2 {
		t.Error("second component should be comp2 (DFS order)")
	}
}

func TestWalkComponents_SkipsNonComponentElements(t *testing.T) {
	comp := &mockComponent{}

	// Only one element has a component
	child := New()
	child.component = comp

	root := New()
	root.AddChild(New(), child, New())

	var found []Component
	walkComponents(root, func(c Component) {
		found = append(found, c)
	})

	if len(found) != 1 {
		t.Errorf("walkComponents found %d components, want 1", len(found))
	}
}

func TestWalkComponents_NilRoot(t *testing.T) {
	var found []Component
	walkComponents(nil, func(c Component) {
		found = append(found, c)
	})

	if len(found) != 0 {
		t.Errorf("walkComponents on nil root found %d components, want 0", len(found))
	}
}

func TestNewMountState(t *testing.T) {
	ms := newMountState()

	if ms.cache == nil {
		t.Error("cache should be initialized")
	}
	if ms.cleanups == nil {
		t.Error("cleanups should be initialized")
	}
	if ms.activeKeys == nil {
		t.Error("activeKeys should be initialized")
	}
	if len(ms.cache) != 0 {
		t.Errorf("cache should be empty, has %d entries", len(ms.cache))
	}
}
