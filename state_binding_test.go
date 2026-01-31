package tui

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestState_Set_CallsBindings(t *testing.T) {
	s := NewState(0)

	var called bool
	var receivedValue int
	s.Bind(func(v int) {
		called = true
		receivedValue = v
	})

	s.Set(42)

	if !called {
		t.Error("binding was not called")
	}
	if receivedValue != 42 {
		t.Errorf("binding received %d, want 42", receivedValue)
	}
}

func TestState_Update(t *testing.T) {
	s := NewState(10)

	s.Update(func(v int) int { return v + 5 })

	if got := s.Get(); got != 15 {
		t.Errorf("after Update(+5), Get() = %d, want 15", got)
	}
}

func TestState_Update_CallsBindings(t *testing.T) {
	s := NewState(0)

	var receivedValue int
	s.Bind(func(v int) {
		receivedValue = v
	})

	s.Update(func(v int) int { return v + 100 })

	if receivedValue != 100 {
		t.Errorf("binding received %d, want 100", receivedValue)
	}
}

func TestState_Bind(t *testing.T) {
	s := NewState(0)

	var callCount int
	s.Bind(func(v int) {
		callCount++
	})

	// Should not be called on registration
	if callCount != 0 {
		t.Errorf("binding called %d times on registration, want 0", callCount)
	}

	// Should be called on Set
	s.Set(1)
	if callCount != 1 {
		t.Errorf("binding called %d times after Set, want 1", callCount)
	}

	// Should be called on each Set
	s.Set(2)
	if callCount != 2 {
		t.Errorf("binding called %d times after second Set, want 2", callCount)
	}
}

func TestState_Bind_ReceivesNewValue(t *testing.T) {
	s := NewState("initial")

	var values []string
	s.Bind(func(v string) {
		values = append(values, v)
	})

	s.Set("first")
	s.Set("second")
	s.Set("third")

	expected := []string{"first", "second", "third"}
	if len(values) != len(expected) {
		t.Errorf("binding called %d times, want %d", len(values), len(expected))
		return
	}
	for i, v := range values {
		if v != expected[i] {
			t.Errorf("values[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestState_Unbind(t *testing.T) {
	s := NewState(0)

	var callCount int
	unbind := s.Bind(func(v int) {
		callCount++
	})

	// Should be called before unbind
	s.Set(1)
	if callCount != 1 {
		t.Errorf("before unbind: callCount = %d, want 1", callCount)
	}

	// Unbind
	unbind()

	// Should NOT be called after unbind
	s.Set(2)
	if callCount != 1 {
		t.Errorf("after unbind: callCount = %d, want 1 (should not increase)", callCount)
	}

	// Further sets should not call the binding
	s.Set(3)
	s.Set(4)
	if callCount != 1 {
		t.Errorf("after multiple sets post-unbind: callCount = %d, want 1", callCount)
	}
}

func TestState_MultipleBindings(t *testing.T) {
	s := NewState(0)

	var callOrder []int
	s.Bind(func(v int) { callOrder = append(callOrder, 1) })
	s.Bind(func(v int) { callOrder = append(callOrder, 2) })
	s.Bind(func(v int) { callOrder = append(callOrder, 3) })

	s.Set(42)

	if len(callOrder) != 3 {
		t.Errorf("got %d calls, want 3", len(callOrder))
	}
	// Verify order is preserved
	for i, v := range callOrder {
		if v != i+1 {
			t.Errorf("callOrder[%d] = %d, want %d", i, v, i+1)
		}
	}
}

func TestState_UnbindSpecificBinding(t *testing.T) {
	s := NewState(0)

	var calls []int
	s.Bind(func(v int) { calls = append(calls, 1) })
	unbind2 := s.Bind(func(v int) { calls = append(calls, 2) })
	s.Bind(func(v int) { calls = append(calls, 3) })

	// All three should fire
	s.Set(1)
	if len(calls) != 3 {
		t.Errorf("before unbind: got %d calls, want 3", len(calls))
	}

	// Unbind only the second binding
	calls = nil
	unbind2()

	// Only first and third should fire
	s.Set(2)
	if len(calls) != 2 {
		t.Errorf("after unbind: got %d calls, want 2", len(calls))
	}
	if calls[0] != 1 || calls[1] != 3 {
		t.Errorf("after unbind: calls = %v, want [1 3]", calls)
	}
}

func TestState_ConcurrentGet(t *testing.T) {
	s := NewState(42)

	// Spawn multiple goroutines that call Get concurrently
	var wg sync.WaitGroup
	const numGoroutines = 100

	results := make([]int, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = s.Get()
		}(i)
	}

	wg.Wait()

	// All results should be the same value
	for i, v := range results {
		if v != 42 {
			t.Errorf("results[%d] = %d, want 42", i, v)
		}
	}
}

func TestState_ConcurrentGetDuringSet(t *testing.T) {
	// Test that concurrent Get() calls are safe while Set() is being called.
	// This verifies the RWMutex properly handles read/write contention.
	s := NewState(0)

	var wg sync.WaitGroup
	const numReaders = 50
	const numWrites = 100

	// Track invalid values using atomic counter (safe for goroutines)
	var invalidCount atomic.Int64

	// Start readers that continuously call Get()
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numWrites; j++ {
				// Get should never panic or return invalid data
				v := s.Get()
				// Value should be non-negative and not exceed final value.
				// Any value from 0 to numWrites is valid since the reader
				// may observe any intermediate state during concurrent writes.
				if v < 0 || v > numWrites {
					invalidCount.Add(1)
				}
			}
		}()
	}

	// Writer goroutine that calls Set()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= numWrites; i++ {
			s.Set(i)
		}
	}()

	wg.Wait()

	// Check for any invalid values observed by readers
	if count := invalidCount.Load(); count > 0 {
		t.Errorf("Get() returned %d invalid values (expected 0)", count)
	}

	// Final value should be numWrites
	if got := s.Get(); got != numWrites {
		t.Errorf("final Get() = %d, want %d", got, numWrites)
	}
}

func TestState_BindingCanCallGet(t *testing.T) {
	// This tests that binding execution happens outside the lock,
	// so calling Get() inside a binding doesn't deadlock.
	s := NewState(0)

	var gotValue int
	s.Bind(func(v int) {
		// This should not deadlock
		gotValue = s.Get()
	})

	// Call Set directly - if there's a deadlock, the test will hang/timeout
	s.Set(42)

	if gotValue != 42 {
		t.Errorf("gotValue = %d, want 42", gotValue)
	}
}

func TestState_SetWithZeroBindings(t *testing.T) {
	// Ensure Set works even with no bindings
	TestResetDirty()

	s := NewState(0)
	s.Set(42) // Should not panic

	if got := s.Get(); got != 42 {
		t.Errorf("Get() = %d, want 42", got)
	}

	// Should still mark dirty
	if !TestCheckAndClearDirty() {
		t.Error("should be dirty after Set() even with no bindings")
	}
}

func TestState_UnbindIdempotent(t *testing.T) {
	s := NewState(0)

	var callCount int
	unbind := s.Bind(func(v int) {
		callCount++
	})

	// Unbind multiple times should not panic
	unbind()
	unbind()
	unbind()

	// Binding should still not fire
	s.Set(1)
	if callCount != 0 {
		t.Errorf("callCount = %d, want 0 after unbind", callCount)
	}
}

func TestState_InactiveBindingsCleanup(t *testing.T) {
	// Test that inactive bindings are cleaned up during Set()
	// to prevent memory leaks
	s := NewState(0)

	// Add several bindings
	unbind1 := s.Bind(func(v int) {})
	unbind2 := s.Bind(func(v int) {})
	unbind3 := s.Bind(func(v int) {})

	// Unbind first and third
	unbind1()
	unbind3()

	// Initial bindings slice has 3 entries (all still present but 2 inactive)
	s.mu.RLock()
	beforeSet := len(s.bindings)
	s.mu.RUnlock()
	if beforeSet != 3 {
		t.Errorf("before Set: bindings length = %d, want 3", beforeSet)
	}

	// Set triggers cleanup
	s.Set(42)

	// After Set, only active bindings should remain
	s.mu.RLock()
	afterSet := len(s.bindings)
	s.mu.RUnlock()
	if afterSet != 1 {
		t.Errorf("after Set: bindings length = %d, want 1 (only active binding)", afterSet)
	}

	// Unbind the last one
	unbind2()
	s.Set(43)

	// Now all bindings should be cleaned up
	s.mu.RLock()
	afterAllUnbind := len(s.bindings)
	s.mu.RUnlock()
	if afterAllUnbind != 0 {
		t.Errorf("after all unbind: bindings length = %d, want 0", afterAllUnbind)
	}
}
