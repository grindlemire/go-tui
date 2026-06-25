package tui

import (
	"testing"
)

// testInputComponent mimics Input: implements Component, KeyListener, Focusable
type testInputComponent struct {
	focused *State[bool]
	text    *State[string]
}

func newTestInputComponent() *testInputComponent {
	return &testInputComponent{
		focused: NewState(false),
		text:    NewState(""),
	}
}

func (c *testInputComponent) BindApp(app *App) {
	c.focused.BindApp(app)
	c.text.BindApp(app)
}

func (c *testInputComponent) Render(app *App) *Element {
	root := New(WithFocusable(true))
	root.SetOnFocus(func(e *Element) { c.Focus() })
	root.SetOnBlur(func(e *Element) { c.Blur() })
	return root
}

func (c *testInputComponent) IsFocusable() bool { return true }
func (c *testInputComponent) IsTabStop() bool   { return true }
func (c *testInputComponent) IsFocused() bool   { return c.focused.Get() }

func (c *testInputComponent) Focus() {
	if c.focused.Get() {
		return
	}
	c.focused.Set(true)
}

func (c *testInputComponent) Blur() {
	if !c.focused.Get() {
		return
	}
	c.focused.Set(false)
}

func (c *testInputComponent) HandleEvent(e Event) bool { return false }

func (c *testInputComponent) KeyMap() KeyMap {
	return KeyMap{
		OnFocused(AnyRune, func(ke KeyEvent) {
			c.text.Set(c.text.Get() + string(ke.Rune))
		}),
	}
}

// testAppComponent is the root component containing an input.
// Like the real generated code, the factory creates a NEW component each time.
type testAppComponent struct{}

func (c *testAppComponent) Render(app *App) *Element {
	root := New()
	el := app.MountPersistent(c, 0, func() Component {
		return newTestInputComponent() // NEW instance each call (like real NewInput)
	})
	root.AddChild(el)
	return root
}

func (c *testAppComponent) KeyMap() KeyMap {
	return KeyMap{
		On(Rune('q'), func(ke KeyEvent) {
			ke.App().Stop()
		}),
		On(KeyTab, func(ke KeyEvent) {
			ke.App().FocusNext()
		}),
	}
}

// getInputFromApp extracts the cached testInputComponent from the mount cache.
func getInputFromApp(app *App, parent Component) *testInputComponent {
	key := mountKey{parent: parent, key: 0}
	if comp, ok := app.mounts.cache[key]; ok {
		return comp.(*testInputComponent)
	}
	return nil
}

func TestFocusIntegration_TabFocusesInput(t *testing.T) {
	appComp := &testAppComponent{}

	term := NewMockTerminal(80, 24)
	app := &App{
		terminal:     term,
		focus:        newFocusManager(),
		buffer:       NewBuffer(80, 24),
		merged:       make(chan Event, 256),
		watcherQueue: make(chan func(), 256),
		stopCh:       make(chan struct{}),
		mounts:       newMountState(),
		batch:        newBatchContext(),
	}

	// Phase 1: SetRootComponent (like the real app)
	app.SetRootComponent(appComp)

	// Phase 2: Initial render + dispatch table build (like App.Run)
	app.Render()
	app.rebuildDispatchTable()

	input := getInputFromApp(app, appComp)
	if input == nil {
		t.Fatal("input component not found in mount cache")
	}

	// Verify: focus manager has exactly 1 focusable element
	if len(app.focus.elements) != 1 {
		t.Fatalf("expected 1 focusable element, got %d", len(app.focus.elements))
	}
	if app.focus.current != -1 {
		t.Fatalf("expected no focus initially (current=-1), got %d", app.focus.current)
	}

	// Phase 3: Press Tab
	tabEvent := KeyEvent{Key: KeyTab, app: app}
	app.dispatchTable.dispatch(tabEvent)

	// After Tab: input should be focused
	if !input.focused.Get() {
		t.Error("after first Tab: input.focused should be true")
	}
	if app.focus.current != 0 {
		t.Errorf("after first Tab: focus.current should be 0, got %d", app.focus.current)
	}

	// Phase 4: Re-render (dirty from focus change)
	app.Render()
	app.rebuildDispatchTable()

	// Get input again (may have been replaced by mount cache)
	input = getInputFromApp(app, appComp)
	if input == nil {
		t.Fatal("input component not found after re-render")
	}

	// After re-render: input should still be focused
	if !input.focused.Get() {
		t.Error("after re-render: input.focused should still be true")
	}
	if app.focus.current != 0 {
		t.Errorf("after re-render: focus.current should be 0, got %d", app.focus.current)
	}

	// Phase 5: Press 'a' - should be captured by focus-gated handler
	aEvent := KeyEvent{Key: KeyRune, Rune: 'a', app: app}
	stopped := app.dispatchTable.dispatch(aEvent)

	if !stopped {
		t.Error("after pressing 'a' while focused: dispatch should return stopped=true")
	}
	if input.text.Get() != "a" {
		t.Errorf("after pressing 'a' while focused: text should be 'a', got %q", input.text.Get())
	}

	// Phase 6: Press 'q' - should be captured by input (not quit)
	qEvent := KeyEvent{Key: KeyRune, Rune: 'q', app: app}
	stopped = app.dispatchTable.dispatch(qEvent)

	if !stopped {
		t.Error("after pressing 'q' while focused: dispatch should return stopped=true (captured by input)")
	}
	if input.text.Get() != "aq" {
		t.Errorf("after pressing 'q' while focused: text should be 'aq', got %q", input.text.Get())
	}

	// Verify app is NOT stopped (q should go to input, not quit)
	if app.stopped {
		t.Error("app should NOT be stopped - 'q' should be captured by focused input")
	}
}

// testTwoInputApp mounts two focusable inputs so dispatch must route a key to the
// focused one and skip the other. Each is its own mounted component, so its
// OnFocused bindings carry its own IsFocused gate.
type testTwoInputApp struct{}

func (c *testTwoInputApp) Render(app *App) *Element {
	root := New(WithDirection(Column))
	root.AddChild(app.MountPersistent(c, 0, func() Component { return newTestInputComponent() }))
	root.AddChild(app.MountPersistent(c, 1, func() Component { return newTestInputComponent() }))
	return root
}

func (c *testTwoInputApp) KeyMap() KeyMap {
	return KeyMap{
		On(KeyTab, func(ke KeyEvent) { ke.App().FocusNext() }),
	}
}

func mountedInput(app *App, parent Component, key int) *testInputComponent {
	if comp, ok := app.mounts.cache[mountKey{parent: parent, key: key}]; ok {
		return comp.(*testInputComponent)
	}
	return nil
}

// TestFocusIntegration_RoutesKeysToFocusedWidget pins the behavior that broke when
// widgets were hand-aggregated instead of mounted: with two focusable widgets, a
// typed key reaches only the focused one. The single-input integration test above
// cannot catch a regression here because it has nothing for the wrong widget to be.
func TestFocusIntegration_RoutesKeysToFocusedWidget(t *testing.T) {
	host := &testTwoInputApp{}
	app := newTestApp(80, 24)
	app.SetRootComponent(host)
	app.MarkDirty()
	app.Render()

	in0 := mountedInput(app, host, 0)
	in1 := mountedInput(app, host, 1)
	if in0 == nil || in1 == nil {
		t.Fatal("both inputs should be mounted")
	}

	// Tab focuses the first input: 'a' lands there only.
	app.Dispatch(KeyEvent{Key: KeyTab})
	app.Render()
	app.Dispatch(KeyEvent{Key: KeyRune, Rune: 'a'})
	if in0.text.Get() != "a" || in1.text.Get() != "" {
		t.Fatalf("after Tab to input 0: in0=%q in1=%q, want in0=\"a\" in1=\"\"", in0.text.Get(), in1.text.Get())
	}

	// Tab moves focus to the second input: 'b' lands there, input 0 is untouched.
	app.Dispatch(KeyEvent{Key: KeyTab})
	app.Render()
	app.Dispatch(KeyEvent{Key: KeyRune, Rune: 'b'})
	if in1.text.Get() != "b" || in0.text.Get() != "a" {
		t.Fatalf("after Tab to input 1: in0=%q in1=%q, want in0=\"a\" in1=\"b\"", in0.text.Get(), in1.text.Get())
	}
}
