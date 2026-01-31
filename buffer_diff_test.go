package tui

import (
	"testing"
)

func TestBuffer_Diff_Empty(t *testing.T) {
	b := NewBuffer(5, 3)

	// No changes - diff should be empty
	changes := b.Diff()
	if len(changes) != 0 {
		t.Errorf("Diff() returned %d changes, want 0", len(changes))
	}
}

func TestBuffer_Diff_SingleChange(t *testing.T) {
	b := NewBuffer(5, 3)
	style := NewStyle()

	b.SetRune(2, 1, 'A', style)

	changes := b.Diff()
	if len(changes) != 1 {
		t.Fatalf("Diff() returned %d changes, want 1", len(changes))
	}

	if changes[0].X != 2 || changes[0].Y != 1 {
		t.Errorf("Change at (%d, %d), want (2, 1)", changes[0].X, changes[0].Y)
	}
	if changes[0].Cell.Rune != 'A' {
		t.Errorf("Change cell rune = %q, want 'A'", changes[0].Cell.Rune)
	}
}

func TestBuffer_Diff_MultipleChanges(t *testing.T) {
	b := NewBuffer(5, 3)
	style := NewStyle()

	b.SetRune(0, 0, 'A', style)
	b.SetRune(4, 0, 'B', style)
	b.SetRune(2, 2, 'C', style)

	changes := b.Diff()
	if len(changes) != 3 {
		t.Fatalf("Diff() returned %d changes, want 3", len(changes))
	}

	// Changes should be in row-major order
	expected := []struct{ x, y int; r rune }{
		{0, 0, 'A'},
		{4, 0, 'B'},
		{2, 2, 'C'},
	}

	for i, e := range expected {
		if changes[i].X != e.x || changes[i].Y != e.y {
			t.Errorf("Change %d at (%d, %d), want (%d, %d)", i, changes[i].X, changes[i].Y, e.x, e.y)
		}
		if changes[i].Cell.Rune != e.r {
			t.Errorf("Change %d rune = %q, want %q", i, changes[i].Cell.Rune, e.r)
		}
	}
}

func TestBuffer_Diff_RowMajorOrder(t *testing.T) {
	b := NewBuffer(3, 3)
	style := NewStyle()

	// Fill in non-row-major order
	b.SetRune(2, 2, 'I', style)
	b.SetRune(0, 0, 'A', style)
	b.SetRune(1, 1, 'E', style)

	changes := b.Diff()
	if len(changes) != 3 {
		t.Fatalf("Diff() returned %d changes, want 3", len(changes))
	}

	// Should come out in row-major order regardless of insertion order
	if changes[0].X != 0 || changes[0].Y != 0 {
		t.Errorf("First change at (%d, %d), want (0, 0)", changes[0].X, changes[0].Y)
	}
	if changes[1].X != 1 || changes[1].Y != 1 {
		t.Errorf("Second change at (%d, %d), want (1, 1)", changes[1].X, changes[1].Y)
	}
	if changes[2].X != 2 || changes[2].Y != 2 {
		t.Errorf("Third change at (%d, %d), want (2, 2)", changes[2].X, changes[2].Y)
	}
}

func TestBuffer_Swap(t *testing.T) {
	b := NewBuffer(5, 3)
	style := NewStyle()

	// Make a change
	b.SetRune(2, 1, 'X', style)

	// Diff should show the change
	changes1 := b.Diff()
	if len(changes1) != 1 {
		t.Fatal("Expected 1 change before swap")
	}

	// Swap
	b.Swap()

	// Diff should now be empty
	changes2 := b.Diff()
	if len(changes2) != 0 {
		t.Errorf("Diff() after Swap() returned %d changes, want 0", len(changes2))
	}
}
