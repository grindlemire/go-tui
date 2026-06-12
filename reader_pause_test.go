//go:build !windows

package tui

import (
	"os"
	"testing"
	"time"
)

// pollResult carries the return values of a PollEvent call made from a
// goroutine back to the test.
type pollResult struct {
	ev Event
	ok bool
}

// newPipeStdinReader creates a stdinReader backed by an os.Pipe. It returns
// the reader and the write end of the pipe. Cleanup of the reader and both
// pipe ends is registered on the test.
func newPipeStdinReader(t *testing.T) (*stdinReader, *os.File) {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}

	reader, err := NewEventReader(r)
	if err != nil {
		t.Fatalf("NewEventReader() error = %v", err)
	}
	sr, ok := reader.(*stdinReader)
	if !ok {
		t.Fatalf("NewEventReader() returned %T, want *stdinReader", reader)
	}

	t.Cleanup(func() {
		sr.Close()
		r.Close()
		w.Close()
	})
	return sr, w
}

// pollUntilEvent polls the reader a bounded number of times. A leftover
// interrupt byte produces one spurious (nil, false) wakeup that this helper
// absorbs before the real data is read.
func pollUntilEvent(t *testing.T, r *stdinReader, attempts int) (Event, bool) {
	t.Helper()
	for range attempts {
		if ev, ok := r.PollEvent(50 * time.Millisecond); ok {
			return ev, true
		}
	}
	return nil, false
}

func TestStdinReader_PauseAndResume(t *testing.T) {
	type tc struct {
		enableInterrupt bool
	}

	tests := map[string]tc{
		"without interrupt enabled": {
			enableInterrupt: false,
		},
		"with interrupt enabled": {
			enableInterrupt: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			reader, w := newPipeStdinReader(t)
			if tt.enableInterrupt {
				if err := reader.EnableInterrupt(); err != nil {
					t.Fatalf("EnableInterrupt() error = %v", err)
				}
			}

			// Data is already waiting on the pipe before Pause is called.
			if _, err := w.Write([]byte{'a'}); err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			reader.Pause()
			if !reader.paused.Load() {
				t.Fatal("Pause() should set the paused flag")
			}

			// While paused, PollEvent must return (nil, false) immediately
			// without reading the pending byte, even with a long timeout.
			start := time.Now()
			ev, ok := reader.PollEvent(200 * time.Millisecond)
			elapsed := time.Since(start)
			if ok || ev != nil {
				t.Errorf("PollEvent() while paused = (%v, %v), want (nil, false)", ev, ok)
			}
			if elapsed > 150*time.Millisecond {
				t.Errorf("PollEvent() while paused took %v, want immediate return", elapsed)
			}

			reader.Resume()
			if reader.paused.Load() {
				t.Fatal("Resume() should clear the paused flag")
			}

			// After Resume the pending byte is readable again. With interrupt
			// enabled, Pause wrote an interrupt byte that costs one spurious
			// wakeup before the data arrives.
			ev, ok = pollUntilEvent(t, reader, 3)
			if !ok {
				t.Fatal("PollEvent() after Resume() returned no event, want the buffered key")
			}
			ke, isKey := ev.(KeyEvent)
			if !isKey {
				t.Fatalf("PollEvent() after Resume() returned %T, want KeyEvent", ev)
			}
			if ke.Key != KeyRune || ke.Rune != 'a' {
				t.Errorf("PollEvent() after Resume() = %+v, want KeyEvent{Key: KeyRune, Rune: 'a'}", ke)
			}
		})
	}
}

func TestStdinReader_PauseInterruptsBlockingPoll(t *testing.T) {
	reader, _ := newPipeStdinReader(t)
	if err := reader.EnableInterrupt(); err != nil {
		t.Fatalf("EnableInterrupt() error = %v", err)
	}

	done := make(chan pollResult, 1)
	started := make(chan struct{})
	go func() {
		close(started)
		ev, ok := reader.PollEvent(-1) // Block until input or interrupt
		done <- pollResult{ev: ev, ok: ok}
	}()

	<-started
	reader.Pause()

	// Pause must wake the blocking poll via the interrupt pipe. If the
	// goroutine had not yet entered the blocking select, the paused flag
	// makes PollEvent return immediately; both paths yield (nil, false).
	select {
	case res := <-done:
		if res.ok || res.ev != nil {
			t.Errorf("blocking PollEvent() after Pause() = (%v, %v), want (nil, false)", res.ev, res.ok)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Pause() did not unblock the in-progress PollEvent")
	}

	reader.Resume()
	if reader.paused.Load() {
		t.Error("Resume() should clear the paused flag")
	}
}

func TestStdinReader_InterruptWakesBlockingPoll(t *testing.T) {
	reader, w := newPipeStdinReader(t)
	if err := reader.EnableInterrupt(); err != nil {
		t.Fatalf("EnableInterrupt() error = %v", err)
	}

	done := make(chan pollResult, 1)
	started := make(chan struct{})
	go func() {
		close(started)
		ev, ok := reader.PollEvent(-1) // Block until input or interrupt
		done <- pollResult{ev: ev, ok: ok}
	}()

	<-started
	if err := reader.Interrupt(); err != nil {
		t.Fatalf("Interrupt() error = %v", err)
	}

	select {
	case res := <-done:
		if res.ok || res.ev != nil {
			t.Errorf("interrupted PollEvent() = (%v, %v), want (nil, false)", res.ev, res.ok)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Interrupt() did not unblock the in-progress PollEvent")
	}

	// The interrupt byte is drained by the woken poll, so a normal read
	// still works afterward.
	if _, err := w.Write([]byte{'b'}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	ev, ok := pollUntilEvent(t, reader, 3)
	if !ok {
		t.Fatal("PollEvent() after Interrupt() returned no event, want the written key")
	}
	ke, isKey := ev.(KeyEvent)
	if !isKey {
		t.Fatalf("PollEvent() after Interrupt() returned %T, want KeyEvent", ev)
	}
	if ke.Key != KeyRune || ke.Rune != 'b' {
		t.Errorf("PollEvent() after Interrupt() = %+v, want KeyEvent{Key: KeyRune, Rune: 'b'}", ke)
	}
}

func TestStdinReader_EnableInterruptIdempotent(t *testing.T) {
	reader, _ := newPipeStdinReader(t)

	if reader.hasInterrupt {
		t.Fatal("hasInterrupt should be false before EnableInterrupt()")
	}
	if err := reader.EnableInterrupt(); err != nil {
		t.Fatalf("EnableInterrupt() error = %v", err)
	}
	if !reader.hasInterrupt {
		t.Fatal("hasInterrupt should be true after EnableInterrupt()")
	}
	firstPipe := reader.interruptPipe

	// A second call must be a no-op that keeps the existing pipe.
	if err := reader.EnableInterrupt(); err != nil {
		t.Fatalf("second EnableInterrupt() error = %v", err)
	}
	if reader.interruptPipe != firstPipe {
		t.Errorf("second EnableInterrupt() replaced the pipe: got %v, want %v", reader.interruptPipe, firstPipe)
	}

	// The original pipe still delivers interrupts: a poll with a pending
	// interrupt byte returns immediately instead of waiting out the timeout.
	if err := reader.Interrupt(); err != nil {
		t.Fatalf("Interrupt() error = %v", err)
	}
	start := time.Now()
	ev, ok := reader.PollEvent(200 * time.Millisecond)
	elapsed := time.Since(start)
	if ok || ev != nil {
		t.Errorf("PollEvent() with pending interrupt = (%v, %v), want (nil, false)", ev, ok)
	}
	if elapsed > 150*time.Millisecond {
		t.Errorf("PollEvent() with pending interrupt took %v, want immediate return", elapsed)
	}
}

func TestStdinReader_InterruptWithoutEnableInterrupt(t *testing.T) {
	reader, w := newPipeStdinReader(t)

	if reader.hasInterrupt {
		t.Fatal("hasInterrupt should be false by default")
	}
	if err := reader.Interrupt(); err != nil {
		t.Errorf("Interrupt() without EnableInterrupt() error = %v, want nil no-op", err)
	}

	// The reader still works normally after the no-op interrupt.
	if _, err := w.Write([]byte{'x'}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	ev, ok := reader.PollEvent(100 * time.Millisecond)
	if !ok {
		t.Fatal("PollEvent() returned no event after no-op Interrupt()")
	}
	ke, isKey := ev.(KeyEvent)
	if !isKey {
		t.Fatalf("PollEvent() returned %T, want KeyEvent", ev)
	}
	if ke.Key != KeyRune || ke.Rune != 'x' {
		t.Errorf("PollEvent() = %+v, want KeyEvent{Key: KeyRune, Rune: 'x'}", ke)
	}
}
