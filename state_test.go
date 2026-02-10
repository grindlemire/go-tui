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
	resetDirty()

	s := NewState(0)
	s.Set(42)

	if got := s.Get(); got != 42 {
		t.Errorf("after Set(42), Get() = %d, want 42", got)
	}
}

func TestState_Set_MarksDirty(t *testing.T) {
	// Reset dirty flag before test
	resetDirty()

	s := NewState(0)

	// Should not be dirty initially
	if checkAndClearDirty() {
		t.Error("should not be dirty before Set()")
	}

	s.Set(1)

	// Should be dirty after Set
	if !checkAndClearDirty() {
		t.Error("should be dirty after Set()")
	}
}

