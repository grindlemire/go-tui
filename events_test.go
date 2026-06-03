package tui

import "testing"

func TestEvents_TopicDelivery(t *testing.T) {
	app := &App{}
	source := NewEvents[string]("build.events")
	sink := NewEvents[string]("build.events")
	source.BindApp(app)
	sink.BindApp(app)

	var got []string
	sink.Subscribe(func(v string) {
		got = append(got, v)
	})

	source.Emit("started")
	source.Emit("finished")

	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}
	if got[0] != "started" || got[1] != "finished" {
		t.Fatalf("unexpected events: %v", got)
	}
}

func TestEvents_TopicIsolation(t *testing.T) {
	app := &App{}
	source := NewEvents[string]("topic.a")
	sink := NewEvents[string]("topic.b")
	source.BindApp(app)
	sink.BindApp(app)

	calls := 0
	sink.Subscribe(func(v string) {
		calls++
	})

	source.Emit("ignored")
	if calls != 0 {
		t.Fatalf("expected 0 calls, got %d", calls)
	}
}

func TestEvents_CrossAppIsolation(t *testing.T) {
	appA := &App{}
	appB := &App{}

	source := NewEvents[string]("shared.topic")
	sink := NewEvents[string]("shared.topic")
	source.BindApp(appA)
	sink.BindApp(appB)

	calls := 0
	sink.Subscribe(func(v string) {
		calls++
	})

	source.Emit("ignored")
	if calls != 0 {
		t.Fatalf("expected 0 calls, got %d", calls)
	}
}

func TestEvents_TypeMismatchDropsEmit(t *testing.T) {
	app := &App{}
	strBus := NewEvents[string]("shared.topic")
	intBus := NewEvents[int]("shared.topic")
	strBus.BindApp(app)
	intBus.BindApp(app)

	var got []string
	strBus.Subscribe(func(v string) {
		got = append(got, v)
	})

	// Emit with mismatched type should be silently dropped, not panic
	intBus.Emit(42)

	strBus.Emit("hello")
	if len(got) != 1 || got[0] != "hello" {
		t.Fatalf("expected [hello], got %v", got)
	}
}

func TestEvents_UnbindAppStopsDelivery(t *testing.T) {
	app := &App{}
	source := NewEvents[string]("notifications")
	sink := NewEvents[string]("notifications")
	source.BindApp(app)
	sink.BindApp(app)

	calls := 0
	sink.Subscribe(func(v string) { calls++ })
	sink.UnbindApp()

	source.Emit("hello")
	if calls != 0 {
		t.Fatalf("expected 0 calls after UnbindApp, got %d", calls)
	}
}

func TestEvents_SubscribeUnsubscribe(t *testing.T) {
	app := &App{}
	bus := NewEvents[string]("notifications")
	bus.BindApp(app)

	calls := 0
	unsub := bus.Subscribe(func(v string) { calls++ })
	bus.Emit("first")
	unsub()
	bus.Emit("second")

	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestEvents_SubscribeBeforeBind(t *testing.T) {
	app := &App{}
	bus := NewEvents[string]("deferred.topic")

	var got []string
	bus.Subscribe(func(v string) { got = append(got, v) })

	bus.BindApp(app)

	bus.Emit("hello")
	bus.Emit("world")

	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d: %v", len(got), got)
	}
}

func TestEvents_SubscribeBeforeBind_CrossBusSameTopic(t *testing.T) {
	app := &App{}
	source := NewEvents[string]("shared")
	sink := NewEvents[string]("shared")

	var got int
	sink.Subscribe(func(string) { got++ })
	sink.BindApp(app)
	source.BindApp(app)

	source.Emit("x")
	if got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
}

func TestEvents_SubscribeBeforeBind_DoubleBind(t *testing.T) {
	app := &App{}
	bus := NewEvents[string]("dbl")

	var got int
	bus.Subscribe(func(string) { got++ })
	bus.BindApp(app)
	bus.BindApp(app)

	bus.Emit("x")
	if got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
}

// TestEvents_SubscribeBeforeBind_SurvivesTopicReset reproduces the gtop report:
// Subscribe in the component constructor, then SetRootComponent binds → renders
// → resetRootSession wipes a.topics → second BindApp is supposed to re-register
// the subscribers. Prior to the fix, the second BindApp took an early-return
// path that skipped re-registration because sub.unsubscribe was still non-nil
// (a handle into the wiped topics map), so Emit silently found zero listeners.
func TestEvents_SubscribeBeforeBind_SurvivesTopicReset(t *testing.T) {
	app := &App{}
	bus := NewEvents[string]("tick")

	var got int
	bus.Subscribe(func(string) { got++ })

	bus.BindApp(app)       // first bind registers into app.topics
	app.resetRootSession() // SetRootComponent's applyRoot path wipes app.topics
	bus.BindApp(app)       // second bind must re-register despite same app

	bus.Emit("x")
	if got != 1 {
		t.Fatalf("expected 1 after topic reset + re-bind, got %d", got)
	}
}

// TestEvents_SetRootComponent_SubscribeBeforeBind exercises the full user-facing
// path: build a root component whose constructor subscribes to a bus, then hand
// it to the App via SetRootComponent. The bus must deliver events after setup.
func TestEvents_SetRootComponent_SubscribeBeforeBind(t *testing.T) {
	app := &App{
		batch:  newBatchContext(),
		mounts: newMountState(),
	}

	var got int
	root := newSubscribeBeforeBindRoot(&got)

	app.SetRootComponent(root)

	root.tickBus.Emit("x")
	if got != 1 {
		t.Fatalf("expected 1 event after SetRootComponent, got %d", got)
	}
}

type subscribeBeforeBindRoot struct {
	tickBus *Events[string]
}

func newSubscribeBeforeBindRoot(counter *int) *subscribeBeforeBindRoot {
	bus := NewEvents[string]("subscribe-before-bind")
	r := &subscribeBeforeBindRoot{tickBus: bus}
	bus.Subscribe(func(string) { *counter++ })
	return r
}

func (r *subscribeBeforeBindRoot) Render(app *App) *Element { return New() }

func (r *subscribeBeforeBindRoot) BindApp(app *App) {
	if r.tickBus != nil {
		r.tickBus.BindApp(app)
	}
}

// TestResetRootSession_CallsUnbindOnCachedMounts verifies the teardown
// contract: cached components must receive UnbindApp before the cache is
// tossed, so their Events subscriptions deregister from a.topics rather than
// leaking.
func TestResetRootSession_CallsUnbindOnCachedMounts(t *testing.T) {
	app := &App{
		mounts: newMountState(),
		batch:  newBatchContext(),
	}

	unbound := 0
	tracker := &componentTracker{onUnbind: func() { unbound++ }}
	key := mountKey{parent: nil, index: 0}
	app.mounts.cache[key] = tracker

	app.resetRootSession()

	if unbound != 1 {
		t.Fatalf("expected UnbindApp to be called once, got %d", unbound)
	}
	if len(app.mounts.cache) != 0 {
		t.Fatalf("expected mount cache to be replaced, still has %d entries", len(app.mounts.cache))
	}
}

// TestSetRootComponent_UnbindsOldRoot verifies that swapping roots calls
// UnbindApp on the outgoing root before binding the new one. Without this,
// the previous root's Events subscriptions would leak into the new session.
func TestSetRootComponent_UnbindsOldRoot(t *testing.T) {
	app := &App{
		mounts: newMountState(),
		batch:  newBatchContext(),
	}

	unbound := 0
	first := &componentTracker{onUnbind: func() { unbound++ }}
	app.SetRootComponent(first)
	if unbound != 0 {
		t.Fatalf("unexpected UnbindApp on first SetRootComponent: %d", unbound)
	}

	second := &componentTracker{}
	app.SetRootComponent(second)
	if unbound != 1 {
		t.Fatalf("expected old root to be unbound on swap, got %d", unbound)
	}
}

// componentTracker is a Component + AppBinder + AppUnbinder used by the
// teardown and root-swap tests. onUnbind (if set) records UnbindApp calls.
type componentTracker struct {
	onUnbind func()
}

func (c *componentTracker) Render(app *App) *Element { return New() }
func (c *componentTracker) BindApp(app *App)         {}
func (c *componentTracker) UnbindApp() {
	if c.onUnbind != nil {
		c.onUnbind()
	}
}

// TestSetRootView_UnbindsPreviousView verifies that swapping views via
// SetRootView drains the outgoing view's subscriptions. Previously,
// a.rootComponent was nil'd on each SetRootView call, so the guard that
// looked at a.rootComponent never fired and two consecutive SetRootView
// calls would leak the first view's listeners into a.topics.
func TestSetRootView_UnbindsPreviousView(t *testing.T) {
	app := &App{
		mounts: newMountState(),
		batch:  newBatchContext(),
	}

	unboundA := 0
	viewA := &viewWithUnbind{onUnbind: func() { unboundA++ }}
	app.SetRootView(viewA)
	if unboundA != 0 {
		t.Fatalf("unexpected UnbindApp on first SetRootView: %d", unboundA)
	}

	viewB := &viewWithUnbind{}
	app.SetRootView(viewB)
	if unboundA != 1 {
		t.Fatalf("expected first view to be unbound on swap, got %d", unboundA)
	}
}

// TestSetRootComponent_AfterSetRootView_UnbindsView verifies that crossing
// between setters also drains the previous root. Without the unified
// rootUnbinder field, the view set via SetRootView would be unreachable
// from SetRootComponent's guard (rootComponent was nil'd).
func TestSetRootComponent_AfterSetRootView_UnbindsView(t *testing.T) {
	app := &App{
		mounts: newMountState(),
		batch:  newBatchContext(),
	}

	unboundView := 0
	view := &viewWithUnbind{onUnbind: func() { unboundView++ }}
	app.SetRootView(view)

	component := &componentTracker{}
	app.SetRootComponent(component)

	if unboundView != 1 {
		t.Fatalf("expected view to be unbound when switching to SetRootComponent, got %d", unboundView)
	}
}

type viewWithUnbind struct {
	onUnbind func()
}

func (v *viewWithUnbind) GetRoot() *Element      { return New() }
func (v *viewWithUnbind) GetWatchers() []Watcher { return nil }
func (v *viewWithUnbind) BindApp(app *App)       {}
func (v *viewWithUnbind) UnbindApp() {
	if v.onUnbind != nil {
		v.onUnbind()
	}
}

func TestNewEvents_EmptyTopicPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for empty topic")
		}
	}()
	_ = NewEvents[string]("")
}
