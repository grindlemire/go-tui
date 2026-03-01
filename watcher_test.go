package tui

import (
	"sync"
	"testing"
	"time"
)

func TestWatch_CreatesWatcher(t *testing.T) {
	ch := make(chan string)
	handler := func(s string) {}

	watcher := Watch(ch, handler)

	if watcher == nil {
		t.Error("Watch() should return a non-nil Watcher")
	}
}

func TestWatch_ReceivesChannelValues(t *testing.T) {
	ch := make(chan string, 10)
	eventQueue := make(chan func(), 10)
	stopCh := make(chan struct{})
	defer close(stopCh)

	var received []string

	handler := func(s string) {
		received = append(received, s)
	}

	watcher := Watch(ch, handler)
	watcher.Start(eventQueue, stopCh)

	ch <- "hello"
	ch <- "world"

	// Drain exactly 2 events with timeout
	for i := 0; i < 2; i++ {
		select {
		case fn := <-eventQueue:
			fn()
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for event %d", i)
		}
	}

	if len(received) != 2 {
		t.Fatalf("received %d values, want 2", len(received))
	}
	if received[0] != "hello" {
		t.Errorf("received[0] = %q, want %q", received[0], "hello")
	}
	if received[1] != "world" {
		t.Errorf("received[1] = %q, want %q", received[1], "world")
	}
}

func TestWatch_ExitsWhenChannelCloses(t *testing.T) {
	ch := make(chan string)
	eventQueue := make(chan func(), 10)
	stopCh := make(chan struct{})

	handler := func(s string) {}

	watcher := Watch(ch, handler)
	watcher.Start(eventQueue, stopCh)

	close(ch)

	// After close, sending to eventQueue should stop.
	// Wait briefly and check nothing was enqueued.
	select {
	case fn := <-eventQueue:
		_ = fn
		t.Error("unexpected event after channel close")
	case <-time.After(200 * time.Millisecond):
		// Good — no events enqueued
	}
}

func TestWatch_ExitsWhenStopChCloses(t *testing.T) {
	ch := make(chan string, 10)
	eventQueue := make(chan func(), 10)
	stopCh := make(chan struct{})

	var received []string
	var mu sync.Mutex
	handler := func(s string) {
		mu.Lock()
		received = append(received, s)
		mu.Unlock()
	}

	watcher := Watch(ch, handler)
	watcher.Start(eventQueue, stopCh)

	// Send one value and drain it
	ch <- "first"
	select {
	case fn := <-eventQueue:
		fn()
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first event")
	}

	// Close stop channel
	close(stopCh)

	// Try to send more — may or may not go in depending on timing
	select {
	case ch <- "second":
	default:
	}

	// Drain any remaining events
	drainTimeout := time.After(200 * time.Millisecond)
	for {
		select {
		case fn := <-eventQueue:
			fn()
		case <-drainTimeout:
			goto done
		}
	}
done:

	mu.Lock()
	defer mu.Unlock()
	if len(received) > 1 {
		t.Errorf("received %d values, want at most 1", len(received))
	}
}

func TestOnTimer_CreatesWatcher(t *testing.T) {
	handler := func() {}

	watcher := OnTimer(time.Second, handler)

	if watcher == nil {
		t.Error("OnTimer() should return a non-nil Watcher")
	}
}

func TestOnTimer_FiresAtInterval(t *testing.T) {
	eventQueue := make(chan func(), 10)
	stopCh := make(chan struct{})
	defer close(stopCh)

	var count int
	handler := func() {
		count++
	}

	watcher := OnTimer(20*time.Millisecond, handler)
	watcher.Start(eventQueue, stopCh)

	// Wait for at least 2 ticks
	for i := 0; i < 2; i++ {
		select {
		case fn := <-eventQueue:
			fn()
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for tick %d", i+1)
		}
	}

	if count < 2 {
		t.Errorf("timer fired %d times, want >= 2", count)
	}
}

func TestOnTimer_ExitsWhenStopChCloses(t *testing.T) {
	eventQueue := make(chan func(), 10)
	stopCh := make(chan struct{})

	var count int
	var mu sync.Mutex
	handler := func() {
		mu.Lock()
		count++
		mu.Unlock()
	}

	watcher := OnTimer(10*time.Millisecond, handler)
	watcher.Start(eventQueue, stopCh)

	// Let it tick at least once
	select {
	case fn := <-eventQueue:
		fn()
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first tick")
	}

	mu.Lock()
	countBeforeStop := count
	mu.Unlock()

	close(stopCh)

	// Drain any in-flight events
	drainTimeout := time.After(200 * time.Millisecond)
	for {
		select {
		case fn := <-eventQueue:
			fn()
		case <-drainTimeout:
			goto done
		}
	}
done:

	mu.Lock()
	finalCount := count
	mu.Unlock()

	if finalCount > countBeforeStop+2 {
		t.Errorf("timer fired %d times (was %d before stop), should have stopped",
			finalCount, countBeforeStop)
	}
}

func TestWatcher_HandlerCalledOnMainLoop(t *testing.T) {
	ch := make(chan int, 1)
	eventQueue := make(chan func(), 10)
	stopCh := make(chan struct{})
	defer close(stopCh)

	// We verify the handler is called when we drain the eventQueue, not immediately
	handlerCalled := make(chan struct{})
	handler := func(v int) {
		close(handlerCalled)
	}

	watcher := Watch(ch, handler)
	watcher.Start(eventQueue, stopCh)

	ch <- 42

	// Handler should NOT have been called yet (it's queued, not direct)
	select {
	case <-handlerCalled:
		t.Error("handler should not be called until event is dequeued")
	case <-time.After(20 * time.Millisecond):
		// Good - handler wasn't called directly
	}

	// Now drain the queue - this should call the handler
	select {
	case fn := <-eventQueue:
		fn()
	case <-time.After(time.Second):
		t.Fatal("no event in queue")
	}

	// Now handler should have been called
	select {
	case <-handlerCalled:
		// Good
	case <-time.After(time.Second):
		t.Error("handler was not called after draining queue")
	}
}

func TestNewChannelWatcher(t *testing.T) {
	ch := make(chan string, 1)
	var received string

	w := NewChannelWatcher(ch, func(s string) {
		received = s
	})

	eventQueue := make(chan func(), 10)
	stopCh := make(chan struct{})

	w.Start(eventQueue, stopCh)
	ch <- "hello"

	select {
	case fn := <-eventQueue:
		fn()
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}

	close(stopCh)

	if received != "hello" {
		t.Fatalf("expected 'hello', got '%s'", received)
	}
}

func TestNewChannelWatcher_StopsOnStopCh(t *testing.T) {
	ch := make(chan string, 10)
	eventQueue := make(chan func(), 10)
	stopCh := make(chan struct{})

	var received []string
	var mu sync.Mutex

	w := NewChannelWatcher(ch, func(s string) {
		mu.Lock()
		received = append(received, s)
		mu.Unlock()
	})

	w.Start(eventQueue, stopCh)

	// Send one value and drain it
	ch <- "first"
	select {
	case fn := <-eventQueue:
		fn()
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first event")
	}

	// Close stop channel
	close(stopCh)

	// Try to send more values - they should not be processed
	select {
	case ch <- "second":
	default:
	}

	// Drain any remaining events
	drainTimeout := time.After(200 * time.Millisecond)
	for {
		select {
		case fn := <-eventQueue:
			fn()
		case <-drainTimeout:
			goto done
		}
	}
done:

	mu.Lock()
	defer mu.Unlock()

	// Should have received at most the first value
	if len(received) > 1 {
		t.Errorf("received %d values, want at most 1 (watcher should have stopped)", len(received))
	}
}

func TestNewChannelWatcher_ExitsWhenChannelCloses(t *testing.T) {
	ch := make(chan string)
	eventQueue := make(chan func(), 10)
	stopCh := make(chan struct{})

	w := NewChannelWatcher(ch, func(s string) {})

	w.Start(eventQueue, stopCh)

	// Close the channel
	close(ch)

	// After close, no events should be enqueued
	select {
	case fn := <-eventQueue:
		_ = fn
		t.Error("unexpected event after channel close")
	case <-time.After(200 * time.Millisecond):
		// Good — no events enqueued
	}
}

func TestWatch_WithDifferentTypes(t *testing.T) {
	type tc struct {
		testFunc func(t *testing.T)
	}

	tests := map[string]tc{
		"string channel": {
			testFunc: func(t *testing.T) {
				ch := make(chan string, 1)
				eventQueue := make(chan func(), 1)
				stopCh := make(chan struct{})
				defer close(stopCh)

				var received string
				watcher := Watch(ch, func(s string) { received = s })
				watcher.Start(eventQueue, stopCh)

				ch <- "test"
				select {
				case fn := <-eventQueue:
					fn()
				case <-time.After(time.Second):
					t.Fatal("timed out waiting for event")
				}

				if received != "test" {
					t.Errorf("received = %q, want %q", received, "test")
				}
			},
		},
		"int channel": {
			testFunc: func(t *testing.T) {
				ch := make(chan int, 1)
				eventQueue := make(chan func(), 1)
				stopCh := make(chan struct{})
				defer close(stopCh)

				var received int
				watcher := Watch(ch, func(i int) { received = i })
				watcher.Start(eventQueue, stopCh)

				ch <- 42
				select {
				case fn := <-eventQueue:
					fn()
				case <-time.After(time.Second):
					t.Fatal("timed out waiting for event")
				}

				if received != 42 {
					t.Errorf("received = %d, want %d", received, 42)
				}
			},
		},
		"struct channel": {
			testFunc: func(t *testing.T) {
				type data struct {
					Name  string
					Value int
				}

				ch := make(chan data, 1)
				eventQueue := make(chan func(), 1)
				stopCh := make(chan struct{})
				defer close(stopCh)

				var received data
				watcher := Watch(ch, func(d data) { received = d })
				watcher.Start(eventQueue, stopCh)

				ch <- data{Name: "test", Value: 123}
				select {
				case fn := <-eventQueue:
					fn()
				case <-time.After(time.Second):
					t.Fatal("timed out waiting for event")
				}

				if received.Name != "test" || received.Value != 123 {
					t.Errorf("received = %+v, want {Name:test Value:123}", received)
				}
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, tt.testFunc)
	}
}
