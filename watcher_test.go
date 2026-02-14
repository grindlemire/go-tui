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
	var mu sync.Mutex

	handler := func(s string) {
		mu.Lock()
		received = append(received, s)
		mu.Unlock()
	}

	watcher := Watch(ch, handler)
	watcher.Start(eventQueue, stopCh)

	// Send values to channel
	ch <- "hello"
	ch <- "world"

	// Wait for values to be enqueued
	time.Sleep(50 * time.Millisecond)

	// Process events from queue
	for len(eventQueue) > 0 {
		fn := <-eventQueue
		fn()
	}

	mu.Lock()
	defer mu.Unlock()

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

	handler := func(s string) {
		// Handler intentionally empty - we're testing the exit behavior
	}

	watcher := Watch(ch, handler)
	watcher.Start(eventQueue, stopCh)

	// Close the channel
	close(ch)

	// Give goroutine time to exit
	time.Sleep(50 * time.Millisecond)

	// The watcher goroutine should have exited - no way to test directly,
	// but we can verify no events were enqueued
	if len(eventQueue) > 0 {
		t.Error("eventQueue should be empty after channel close")
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

	// Send one value
	ch <- "first"
	time.Sleep(20 * time.Millisecond)

	// Close stop channel
	close(stopCh)

	// Try to send more values - they should not be processed
	select {
	case ch <- "second":
	default:
	}

	time.Sleep(50 * time.Millisecond)

	// Process any queued events
	for len(eventQueue) > 0 {
		fn := <-eventQueue
		fn()
	}

	mu.Lock()
	defer mu.Unlock()

	// Should have received at most the first value
	if len(received) > 1 {
		t.Errorf("received %d values, want at most 1 (watcher should have stopped)", len(received))
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
	var mu sync.Mutex
	handler := func() {
		mu.Lock()
		count++
		mu.Unlock()
	}

	watcher := OnTimer(20*time.Millisecond, handler)
	watcher.Start(eventQueue, stopCh)

	// Wait for a few ticks
	time.Sleep(70 * time.Millisecond)

	// Process events from queue
	for len(eventQueue) > 0 {
		fn := <-eventQueue
		fn()
	}

	mu.Lock()
	defer mu.Unlock()

	// Should have fired 2-3 times in 70ms with 20ms interval
	if count < 2 || count > 4 {
		t.Errorf("timer fired %d times, want 2-4 times", count)
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

	// Let it tick once or twice and process events
	time.Sleep(25 * time.Millisecond)
	for len(eventQueue) > 0 {
		fn := <-eventQueue
		fn()
	}

	// Record count before stop
	mu.Lock()
	countBeforeStop := count
	mu.Unlock()

	// Close stop channel
	close(stopCh)

	// Wait long enough for multiple additional ticks if it hadn't stopped
	time.Sleep(50 * time.Millisecond)

	// Process any events that were in flight when we stopped
	for len(eventQueue) > 0 {
		fn := <-eventQueue
		fn()
	}

	mu.Lock()
	finalCount := count
	mu.Unlock()

	// Count should not have increased significantly after stop
	// Allow for events that were already queued or in flight
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
	case <-time.After(100 * time.Millisecond):
		t.Fatal("no event in queue")
	}

	// Now handler should have been called
	select {
	case <-handlerCalled:
		// Good
	case <-time.After(100 * time.Millisecond):
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

	// Give goroutine time to process
	time.Sleep(10 * time.Millisecond)

	// Drain event queue to execute handler
	for len(eventQueue) > 0 {
		fn := <-eventQueue
		fn()
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

	// Send one value
	ch <- "first"
	time.Sleep(20 * time.Millisecond)

	// Close stop channel
	close(stopCh)

	// Try to send more values - they should not be processed
	select {
	case ch <- "second":
	default:
	}

	time.Sleep(50 * time.Millisecond)

	// Process any queued events
	for len(eventQueue) > 0 {
		fn := <-eventQueue
		fn()
	}

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

	w := NewChannelWatcher(ch, func(s string) {
		// Handler intentionally empty
	})

	w.Start(eventQueue, stopCh)

	// Close the channel
	close(ch)

	// Give goroutine time to exit
	time.Sleep(50 * time.Millisecond)

	// The watcher goroutine should have exited - no way to test directly,
	// but we can verify no events were enqueued
	if len(eventQueue) > 0 {
		t.Error("eventQueue should be empty after channel close")
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
				time.Sleep(20 * time.Millisecond)
				(<-eventQueue)()

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
				time.Sleep(20 * time.Millisecond)
				(<-eventQueue)()

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
				time.Sleep(20 * time.Millisecond)
				(<-eventQueue)()

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
