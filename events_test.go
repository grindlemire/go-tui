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

func TestNewEvents_EmptyTopicPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for empty topic")
		}
	}()
	_ = NewEvents[string]("")
}
