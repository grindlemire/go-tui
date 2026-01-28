package tui

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestState_NewState(t *testing.T) {
	type tc struct {
		initial int
	}

	tests := map[string]tc{
		"creates state with zero value": {
			initial: 0,
		},
		"creates state with positive value": {
			initial: 42,
		},
		"creates state with negative value": {
			initial: -10,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := NewState(tt.initial)
			if s.Get() != tt.initial {
				t.Errorf("NewState(%d).Get() = %d, want %d", tt.initial, s.Get(), tt.initial)
			}
		})
	}
}

func TestState_NewState_TypeInference(t *testing.T) {
	// Test that NewState correctly infers various types
	t.Run("int", func(t *testing.T) {
		s := NewState(42)
		if got := s.Get(); got != 42 {
			t.Errorf("Get() = %d, want 42", got)
		}
	})

	t.Run("string", func(t *testing.T) {
		s := NewState("hello")
		if got := s.Get(); got != "hello" {
			t.Errorf("Get() = %q, want %q", got, "hello")
		}
	})

	t.Run("bool", func(t *testing.T) {
		s := NewState(true)
		if got := s.Get(); got != true {
			t.Errorf("Get() = %v, want true", got)
		}
	})

	t.Run("slice", func(t *testing.T) {
		s := NewState([]string{"a", "b"})
		got := s.Get()
		if len(got) != 2 || got[0] != "a" || got[1] != "b" {
			t.Errorf("Get() = %v, want [a b]", got)
		}
	})

	t.Run("struct pointer", func(t *testing.T) {
		type User struct{ Name string }
		s := NewState(&User{Name: "Alice"})
		got := s.Get()
		if got == nil || got.Name != "Alice" {
			t.Errorf("Get() = %v, want &User{Name:Alice}", got)
		}
	})
}

func TestState_Get(t *testing.T) {
	s := NewState(100)

	// Get should return current value
	if got := s.Get(); got != 100 {
		t.Errorf("Get() = %d, want 100", got)
	}

	// Get should be idempotent
	if got := s.Get(); got != 100 {
		t.Errorf("Get() second call = %d, want 100", got)
	}
}

func TestState_Set(t *testing.T) {
	// Reset dirty flag before test
	TestResetDirty()

	s := NewState(0)
	s.Set(42)

	if got := s.Get(); got != 42 {
		t.Errorf("after Set(42), Get() = %d, want 42", got)
	}
}

func TestState_Set_MarksDirty(t *testing.T) {
	// Reset dirty flag before test
	TestResetDirty()

	s := NewState(0)

	// Should not be dirty initially
	if TestCheckAndClearDirty() {
		t.Error("should not be dirty before Set()")
	}

	s.Set(1)

	// Should be dirty after Set
	if !TestCheckAndClearDirty() {
		t.Error("should be dirty after Set()")
	}
}

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

// === Batch Tests ===

func TestBatch_DefersBindingExecution(t *testing.T) {
	TestResetBatch()

	s := NewState(0)

	var callCount int
	var lastValue int
	s.Bind(func(v int) {
		callCount++
		lastValue = v
	})

	Batch(func() {
		// Binding should not be called during batch
		s.Set(42)
		if callCount != 0 {
			t.Errorf("binding called during batch: callCount = %d, want 0", callCount)
		}
	})

	// Binding should be called after batch completes
	if callCount != 1 {
		t.Errorf("after batch: callCount = %d, want 1", callCount)
	}
	if lastValue != 42 {
		t.Errorf("after batch: lastValue = %d, want 42", lastValue)
	}
}

func TestBatch_MultipleSetsToSameState(t *testing.T) {
	TestResetBatch()

	s := NewState(0)

	var callCount int
	var receivedValues []int
	s.Bind(func(v int) {
		callCount++
		receivedValues = append(receivedValues, v)
	})

	Batch(func() {
		s.Set(1)
		s.Set(2)
		s.Set(3) // final value
	})

	// Binding should only be called once with the final value
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1", callCount)
	}
	if len(receivedValues) != 1 || receivedValues[0] != 3 {
		t.Errorf("receivedValues = %v, want [3]", receivedValues)
	}
}

func TestBatch_FinalValueReceived(t *testing.T) {
	TestResetBatch()

	s := NewState("initial")

	var receivedValue string
	s.Bind(func(v string) {
		receivedValue = v
	})

	Batch(func() {
		s.Set("first")
		s.Set("second")
		s.Set("final")
	})

	if receivedValue != "final" {
		t.Errorf("receivedValue = %q, want %q", receivedValue, "final")
	}
}

func TestBatch_MultipleDifferentStates(t *testing.T) {
	TestResetBatch()

	s1 := NewState(0)
	s2 := NewState("")

	var s1CallCount, s2CallCount int
	var s1Value int
	var s2Value string

	s1.Bind(func(v int) {
		s1CallCount++
		s1Value = v
	})
	s2.Bind(func(v string) {
		s2CallCount++
		s2Value = v
	})

	Batch(func() {
		s1.Set(42)
		s2.Set("hello")
	})

	// Each binding should be called exactly once
	if s1CallCount != 1 {
		t.Errorf("s1CallCount = %d, want 1", s1CallCount)
	}
	if s2CallCount != 1 {
		t.Errorf("s2CallCount = %d, want 1", s2CallCount)
	}
	if s1Value != 42 {
		t.Errorf("s1Value = %d, want 42", s1Value)
	}
	if s2Value != "hello" {
		t.Errorf("s2Value = %q, want %q", s2Value, "hello")
	}
}

func TestBatch_NestedBatches(t *testing.T) {
	TestResetBatch()

	s := NewState(0)

	var callCount int
	var lastValue int
	s.Bind(func(v int) {
		callCount++
		lastValue = v
	})

	Batch(func() {
		s.Set(1)

		// Nested batch
		Batch(func() {
			s.Set(2)

			// Further nested
			Batch(func() {
				s.Set(3)
			})

			// Bindings should still not have fired
			if callCount != 0 {
				t.Errorf("binding called during nested batch: callCount = %d, want 0", callCount)
			}
		})

		// Still in outer batch
		if callCount != 0 {
			t.Errorf("binding called before outer batch complete: callCount = %d, want 0", callCount)
		}
	})

	// Now all bindings should fire (only once with final value)
	if callCount != 1 {
		t.Errorf("after all batches: callCount = %d, want 1", callCount)
	}
	if lastValue != 3 {
		t.Errorf("after all batches: lastValue = %d, want 3", lastValue)
	}
}

func TestBatch_DeduplicationByBindingID(t *testing.T) {
	TestResetBatch()

	// This test verifies that deduplication uses binding IDs, not function
	// pointer comparison. Multiple bindings on the same state should each
	// fire once, even if they have the same function signature.
	s := NewState(0)

	var binding1Count, binding2Count int
	s.Bind(func(v int) { binding1Count++ })
	s.Bind(func(v int) { binding2Count++ })

	Batch(func() {
		s.Set(1)
		s.Set(2)
		s.Set(3)
	})

	// Each binding should fire exactly once
	if binding1Count != 1 {
		t.Errorf("binding1Count = %d, want 1", binding1Count)
	}
	if binding2Count != 1 {
		t.Errorf("binding2Count = %d, want 1", binding2Count)
	}
}

func TestBatch_NoSetsDoesntError(t *testing.T) {
	TestResetBatch()

	// Batch with no Set calls should not error
	Batch(func() {
		// do nothing
	})
	// Test passes if no panic occurs
}

func TestBatch_EmptyPendingAfterExecution(t *testing.T) {
	TestResetBatch()

	s := NewState(0)

	var callCount int
	s.Bind(func(v int) {
		callCount++
	})

	// First batch
	Batch(func() {
		s.Set(1)
	})

	if callCount != 1 {
		t.Errorf("after first batch: callCount = %d, want 1", callCount)
	}

	// Second batch - should work independently
	Batch(func() {
		s.Set(2)
	})

	if callCount != 2 {
		t.Errorf("after second batch: callCount = %d, want 2", callCount)
	}
}

func TestBatch_SetOutsideBatchStillWorks(t *testing.T) {
	TestResetBatch()

	s := NewState(0)

	var callCount int
	s.Bind(func(v int) {
		callCount++
	})

	// Set outside batch should work immediately
	s.Set(1)
	if callCount != 1 {
		t.Errorf("after Set outside batch: callCount = %d, want 1", callCount)
	}

	// Batch should also work
	Batch(func() {
		s.Set(2)
	})
	if callCount != 2 {
		t.Errorf("after batch: callCount = %d, want 2", callCount)
	}

	// Set after batch should work immediately again
	s.Set(3)
	if callCount != 3 {
		t.Errorf("after Set after batch: callCount = %d, want 3", callCount)
	}
}

func TestBatch_MarksDirty(t *testing.T) {
	TestResetDirty()
	TestResetBatch()

	s := NewState(0)

	// Should not be dirty initially
	if TestCheckAndClearDirty() {
		t.Error("should not be dirty before batch")
	}

	Batch(func() {
		s.Set(1)
		// Dirty should be marked immediately within batch
		if !TestCheckAndClearDirty() {
			t.Error("should be dirty after Set within batch")
		}
	})
}

func TestBatch_MultipleBindingsPerState(t *testing.T) {
	TestResetBatch()

	s := NewState(0)

	var values1, values2, values3 []int
	s.Bind(func(v int) { values1 = append(values1, v) })
	s.Bind(func(v int) { values2 = append(values2, v) })
	s.Bind(func(v int) { values3 = append(values3, v) })

	Batch(func() {
		s.Set(10)
		s.Set(20)
		s.Set(30)
	})

	// Each binding should receive only the final value
	expected := []int{30}
	if len(values1) != 1 || values1[0] != 30 {
		t.Errorf("values1 = %v, want %v", values1, expected)
	}
	if len(values2) != 1 || values2[0] != 30 {
		t.Errorf("values2 = %v, want %v", values2, expected)
	}
	if len(values3) != 1 || values3[0] != 30 {
		t.Errorf("values3 = %v, want %v", values3, expected)
	}
}

func TestBatch_PanicRecovery(t *testing.T) {
	// Test that if fn panics, the batch state is properly cleaned up
	// and subsequent batches work correctly.
	TestResetBatch()

	s := NewState(0)

	var callCount int
	s.Bind(func(v int) {
		callCount++
	})

	// First batch that panics
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic did not occur")
			}
		}()
		Batch(func() {
			s.Set(1)
			panic("test panic")
		})
	}()

	// After panic, batch depth should be reset to 0
	// A subsequent batch should work correctly
	Batch(func() {
		s.Set(2)
	})

	// The binding should have fired for the second batch
	// (and possibly for the first if bindings ran before panic cleanup)
	// Most importantly, subsequent batches should work
	if callCount < 1 {
		t.Errorf("callCount = %d, want at least 1 (batch should work after panic)", callCount)
	}

	// Final value should be 2
	if got := s.Get(); got != 2 {
		t.Errorf("Get() = %d, want 2", got)
	}
}
