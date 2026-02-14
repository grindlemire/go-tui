package tui

import (
	"sync"
	"testing"
)

func TestDirty_MarkDirty(t *testing.T) {
	// Reset state before test
	testApp.resetDirty()

	// Initially not dirty
	if testApp.checkAndClearDirty() {
		t.Error("checkAndClearDirty() should return false when not marked dirty")
	}

	// Mark dirty
	testApp.MarkDirty()

	// Now should be dirty
	if !testApp.checkAndClearDirty() {
		t.Error("checkAndClearDirty() should return true after MarkDirty()")
	}
}

func TestDirty_CheckAndClearDirty(t *testing.T) {
	type tc struct {
		markDirty    bool
		expectFirst  bool
		expectSecond bool
	}

	tests := map[string]tc{
		"returns true and clears flag when dirty": {
			markDirty:    true,
			expectFirst:  true,
			expectSecond: false,
		},
		"returns false when not dirty": {
			markDirty:    false,
			expectFirst:  false,
			expectSecond: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Reset state before test
			testApp.resetDirty()

			if tt.markDirty {
				testApp.MarkDirty()
			}

			// First check
			first := testApp.checkAndClearDirty()
			if first != tt.expectFirst {
				t.Errorf("first checkAndClearDirty() = %v, want %v", first, tt.expectFirst)
			}

			// Second check should always be false (flag was cleared)
			second := testApp.checkAndClearDirty()
			if second != tt.expectSecond {
				t.Errorf("second checkAndClearDirty() = %v, want %v", second, tt.expectSecond)
			}
		})
	}
}

func TestDirty_ConcurrentMarkDirty(t *testing.T) {
	// Reset state before test
	testApp.resetDirty()

	// Spawn multiple goroutines that call MarkDirty concurrently
	var wg sync.WaitGroup
	const numGoroutines = 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			testApp.MarkDirty()
		}()
	}

	wg.Wait()

	// After all goroutines complete, dirty flag should be set
	if !testApp.checkAndClearDirty() {
		t.Error("checkAndClearDirty() should return true after concurrent MarkDirty() calls")
	}

	// And now it should be cleared
	if testApp.checkAndClearDirty() {
		t.Error("checkAndClearDirty() should return false after first check cleared the flag")
	}
}

func TestDirty_ResetDirty(t *testing.T) {
	// Mark dirty
	testApp.MarkDirty()

	// Reset should clear
	testApp.resetDirty()

	// Should not be dirty anymore
	if testApp.checkAndClearDirty() {
		t.Error("checkAndClearDirty() should return false after resetDirty()")
	}
}
