package tui

import (
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
	testApp.resetDirty()

	s := NewState(0)
	s.BindApp(testApp)
	s.Set(42)

	if got := s.Get(); got != 42 {
		t.Errorf("after Set(42), Get() = %d, want 42", got)
	}
}

func TestState_Set_MarksDirty(t *testing.T) {
	// Reset dirty flag before test
	testApp.resetDirty()

	s := NewState(0)
	s.BindApp(testApp)

	// Should not be dirty initially
	if testApp.checkAndClearDirty() {
		t.Error("should not be dirty before Set()")
	}

	s.Set(1)

	// Should be dirty after Set
	if !testApp.checkAndClearDirty() {
		t.Error("should be dirty after Set()")
	}
}

func TestState_Set_ReentrantCycleBreaks(t *testing.T) {
	a := NewState(0)
	b := NewState(0)
	a.BindApp(testApp)
	b.BindApp(testApp)
	testApp.resetDirty()

	// Create a cycle: a's binding sets b, b's binding sets a
	var aBindCount, bBindCount int
	a.Bind(func(v int) {
		aBindCount++
		b.Set(v * 2)
	})
	b.Bind(func(v int) {
		bBindCount++
		a.Set(v + 1)
	})

	a.Set(5)

	// a should have been set twice: once explicitly (5), once by b's binding (11)
	if got := a.Get(); got != 11 {
		t.Errorf("a.Get() = %d, want 11", got)
	}
	// b should have been set once by a's binding (10)
	if got := b.Get(); got != 10 {
		t.Errorf("b.Get() = %d, want 10", got)
	}
	// a's binding should fire once (the re-entrant Set from b skips a's bindings)
	if aBindCount != 1 {
		t.Errorf("a binding fired %d times, want 1", aBindCount)
	}
	// b's binding should fire once
	if bBindCount != 1 {
		t.Errorf("b binding fired %d times, want 1", bBindCount)
	}
}

func TestState_Set_ReentrantSelfReference(t *testing.T) {
	s := NewState(0)
	s.BindApp(testApp)
	testApp.resetDirty()

	var bindCount int
	s.Bind(func(v int) {
		bindCount++
		if v < 10 {
			s.Set(v + 1) // re-entrant: updates value but skips bindings
		}
	})

	s.Set(1)

	// Value should be 2 (binding ran once, set to 1+1=2, re-entrant Set skipped bindings)
	if got := s.Get(); got != 2 {
		t.Errorf("s.Get() = %d, want 2", got)
	}
	if bindCount != 1 {
		t.Errorf("binding fired %d times, want 1", bindCount)
	}
}

func TestState_Set_TreeDependencyNoReentrant(t *testing.T) {
	// A → B and C → B (no cycle, B should get set twice)
	a := NewState(0)
	b := NewState(0)
	c := NewState(0)
	a.BindApp(testApp)
	b.BindApp(testApp)
	c.BindApp(testApp)
	testApp.resetDirty()

	var bBindCount int
	a.Bind(func(v int) { b.Set(v * 10) })
	c.Bind(func(v int) { b.Set(v * 100) })
	b.Bind(func(_ int) { bBindCount++ })

	a.Set(1)
	c.Set(2)

	// b should hold the last value set (from c's binding)
	if got := b.Get(); got != 200 {
		t.Errorf("b.Get() = %d, want 200", got)
	}
	// b's binding should fire twice (once per Set from a and c)
	if bBindCount != 2 {
		t.Errorf("b binding fired %d times, want 2", bBindCount)
	}
}

func TestState_Set_ReentrantWithNoApp(t *testing.T) {
	// Same cycle detection should work without an app bound
	a := NewState(0)
	b := NewState(0)

	var aBindCount int
	a.Bind(func(v int) {
		aBindCount++
		b.Set(v * 2)
	})
	b.Bind(func(v int) {
		a.Set(v + 1)
	})

	a.Set(5)

	if got := a.Get(); got != 11 {
		t.Errorf("a.Get() = %d, want 11", got)
	}
	if got := b.Get(); got != 10 {
		t.Errorf("b.Get() = %d, want 10", got)
	}
	if aBindCount != 1 {
		t.Errorf("a binding fired %d times, want 1", aBindCount)
	}
}

