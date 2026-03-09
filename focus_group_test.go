package tui

import "testing"

func TestFocusGroup_NextCyclesForward(t *testing.T) {
	type tc struct {
		numMembers int
		nextCalls  int
		wantIndex  int
	}

	tests := map[string]tc{
		"next from first to second": {
			numMembers: 3,
			nextCalls:  1,
			wantIndex:  1,
		},
		"next from second to third": {
			numMembers: 3,
			nextCalls:  2,
			wantIndex:  2,
		},
		"next wraps to beginning": {
			numMembers: 3,
			nextCalls:  3,
			wantIndex:  0,
		},
		"next wraps with two members": {
			numMembers: 2,
			nextCalls:  2,
			wantIndex:  0,
		},
		"multiple full cycles": {
			numMembers: 3,
			nextCalls:  7,
			wantIndex:  1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			members := make([]*State[bool], tt.numMembers)
			for i := range members {
				s := NewState(false)
				s.BindApp(testApp)
				members[i] = s
			}
			fg, err := NewFocusGroup(members...)
			if err != nil {
				t.Fatalf("NewFocusGroup: %v", err)
			}

			for i := 0; i < tt.nextCalls; i++ {
				fg.Next()
			}

			if fg.Current() != tt.wantIndex {
				t.Errorf("Current() = %d, want %d", fg.Current(), tt.wantIndex)
			}
		})
	}
}

func TestFocusGroup_PrevCyclesBackward(t *testing.T) {
	type tc struct {
		numMembers int
		prevCalls  int
		wantIndex  int
	}

	tests := map[string]tc{
		"prev from first wraps to last": {
			numMembers: 3,
			prevCalls:  1,
			wantIndex:  2,
		},
		"prev from last to second": {
			numMembers: 3,
			prevCalls:  2,
			wantIndex:  1,
		},
		"prev full cycle": {
			numMembers: 3,
			prevCalls:  3,
			wantIndex:  0,
		},
		"prev with two members": {
			numMembers: 2,
			prevCalls:  1,
			wantIndex:  1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			members := make([]*State[bool], tt.numMembers)
			for i := range members {
				s := NewState(false)
				s.BindApp(testApp)
				members[i] = s
			}
			fg, err := NewFocusGroup(members...)
			if err != nil {
				t.Fatalf("NewFocusGroup: %v", err)
			}

			for i := 0; i < tt.prevCalls; i++ {
				fg.Prev()
			}

			if fg.Current() != tt.wantIndex {
				t.Errorf("Current() = %d, want %d", fg.Current(), tt.wantIndex)
			}
		})
	}
}

func TestFocusGroup_MutualExclusion(t *testing.T) {
	m0 := NewState(false)
	m0.BindApp(testApp)
	m1 := NewState(false)
	m1.BindApp(testApp)
	m2 := NewState(false)
	m2.BindApp(testApp)
	fg, err := NewFocusGroup(m0, m1, m2)
	if err != nil {
		t.Fatalf("NewFocusGroup: %v", err)
	}

	// After construction, only first member should be active
	assertMemberStates(t, "initial", []*State[bool]{m0, m1, m2}, 0)

	// After Next(), only second member should be active
	fg.Next()
	assertMemberStates(t, "after Next()", []*State[bool]{m0, m1, m2}, 1)

	// After another Next(), only third member should be active
	fg.Next()
	assertMemberStates(t, "after 2x Next()", []*State[bool]{m0, m1, m2}, 2)

	// After Prev(), only second member should be active
	fg.Prev()
	assertMemberStates(t, "after Prev()", []*State[bool]{m0, m1, m2}, 1)
}

func TestFocusGroup_InitialState(t *testing.T) {
	m0 := NewState(false)
	m1 := NewState(false)
	m2 := NewState(false)
	_, err := NewFocusGroup(m0, m1, m2)
	if err != nil {
		t.Fatalf("NewFocusGroup: %v", err)
	}

	// First member should be active, rest should be inactive
	if !m0.Get() {
		t.Error("first member should be active after construction")
	}
	if m1.Get() {
		t.Error("second member should be inactive after construction")
	}
	if m2.Get() {
		t.Error("third member should be inactive after construction")
	}
}

func TestFocusGroup_KeyMap(t *testing.T) {
	m0 := NewState(false)
	m1 := NewState(false)
	fg, err := NewFocusGroup(m0, m1)
	if err != nil {
		t.Fatalf("NewFocusGroup: %v", err)
	}

	km := fg.KeyMap()
	if km == nil {
		t.Fatal("KeyMap() returned nil")
	}
	if len(km) != 2 {
		t.Fatalf("KeyMap() returned %d bindings, want 2", len(km))
	}

	// First binding should be Tab with RequireNoMods
	if km[0].Pattern.Key != KeyTab {
		t.Errorf("first binding key = %v, want KeyTab", km[0].Pattern.Key)
	}
	if !km[0].Pattern.RequireNoMods {
		t.Error("first binding should have RequireNoMods=true")
	}
	if km[0].Stop {
		t.Error("Tab binding should be broadcast (Stop=false)")
	}

	// Second binding should be Shift+Tab
	if km[1].Pattern.Key != KeyTab {
		t.Errorf("second binding key = %v, want KeyTab", km[1].Pattern.Key)
	}
	if km[1].Pattern.Mod != ModShift {
		t.Errorf("second binding mod = %v, want ModShift", km[1].Pattern.Mod)
	}
	if km[1].Stop {
		t.Error("Shift+Tab binding should be broadcast (Stop=false)")
	}
}

func TestFocusGroup_KeyMapHandlersWork(t *testing.T) {
	m0 := NewState(false)
	m0.BindApp(testApp)
	m1 := NewState(false)
	m1.BindApp(testApp)
	m2 := NewState(false)
	m2.BindApp(testApp)
	fg, err := NewFocusGroup(m0, m1, m2)
	if err != nil {
		t.Fatalf("NewFocusGroup: %v", err)
	}

	km := fg.KeyMap()

	// Tab handler should advance focus
	km[0].Handler(KeyEvent{Key: KeyTab})
	assertMemberStates(t, "after Tab", []*State[bool]{m0, m1, m2}, 1)

	// Shift+Tab handler should go backward
	km[1].Handler(KeyEvent{Key: KeyTab, Mod: ModShift})
	assertMemberStates(t, "after Shift+Tab", []*State[bool]{m0, m1, m2}, 0)
}

func TestFocusGroup_ErrorWithFewerThanTwoMembers(t *testing.T) {
	t.Run("zero members", func(t *testing.T) {
		if _, err := NewFocusGroup(); err == nil {
			t.Error("expected error with 0 members")
		}
	})

	t.Run("one member", func(t *testing.T) {
		if _, err := NewFocusGroup(NewState(false)); err == nil {
			t.Error("expected error with 1 member")
		}
	})
}

func TestFocusGroup_ShiftTabDoesNotTriggerNext(t *testing.T) {
	m0 := NewState(false)
	m0.BindApp(testApp)
	m1 := NewState(false)
	m1.BindApp(testApp)
	m2 := NewState(false)
	m2.BindApp(testApp)
	fg, err := NewFocusGroup(m0, m1, m2)
	if err != nil {
		t.Fatalf("NewFocusGroup: %v", err)
	}

	km := fg.KeyMap()

	// Simulate Shift+Tab — only the Shift+Tab binding should match
	event := KeyEvent{Key: KeyTab, Mod: ModShift}
	matched := 0
	for _, binding := range km {
		entry := dispatchEntry{pattern: binding.Pattern, handler: binding.Handler, stop: binding.Stop}
		if entry.matches(event) {
			entry.handler(event)
			matched++
		}
	}

	if matched != 1 {
		t.Errorf("Shift+Tab matched %d bindings, want 1", matched)
	}
	// Should have gone backward to index 2
	assertMemberStates(t, "after Shift+Tab", []*State[bool]{m0, m1, m2}, 2)
}

// assertMemberStates verifies that only the member at expectedActive index
// is true and all others are false.
func assertMemberStates(t *testing.T, context string, members []*State[bool], expectedActive int) {
	t.Helper()
	for i, m := range members {
		if i == expectedActive {
			if !m.Get() {
				t.Errorf("%s: member[%d] should be active (true), got false", context, i)
			}
		} else {
			if m.Get() {
				t.Errorf("%s: member[%d] should be inactive (false), got true", context, i)
			}
		}
	}
}
