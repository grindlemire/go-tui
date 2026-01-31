package tui

import (
	"testing"
)

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
