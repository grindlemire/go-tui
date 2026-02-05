package tui

import "testing"

// --- Mock components for dispatch testing ---

// mockKeyComponent implements Component and KeyListener.
type mockKeyComponent struct {
	keyMap KeyMap
}

func (m *mockKeyComponent) Render() *Element {
	return New()
}

func (m *mockKeyComponent) KeyMap() KeyMap {
	return m.keyMap
}

// mockNoKeyComponent implements Component but not KeyListener.
type mockNoKeyComponent struct{}

func (m *mockNoKeyComponent) Render() *Element {
	return New()
}

// --- Helpers ---

// buildTestTree creates an element tree with components attached for testing.
// Returns the root element. Components are attached in DFS order.
func buildTestTree(components ...Component) *Element {
	root := New()
	for _, comp := range components {
		child := New()
		child.component = comp
		root.AddChild(child)
	}
	return root
}

// buildNestedTestTree creates a nested tree: root -> child1 -> child2.
// Each component is at a different depth for tree-order testing.
func buildNestedTestTree(parent, child Component) *Element {
	childEl := New()
	childEl.component = child

	parentEl := New()
	parentEl.component = parent
	parentEl.AddChild(childEl)

	root := New()
	root.AddChild(parentEl)
	return root
}

// --- dispatchEntry.matches tests ---

func TestDispatchEntry_Matches(t *testing.T) {
	type tc struct {
		pattern KeyPattern
		event   KeyEvent
		want    bool
	}

	tests := map[string]tc{
		"AnyRune matches printable character": {
			pattern: KeyPattern{AnyRune: true},
			event:   KeyEvent{Key: KeyRune, Rune: 'a'},
			want:    true,
		},
		"AnyRune does not match special key": {
			pattern: KeyPattern{AnyRune: true},
			event:   KeyEvent{Key: KeyEscape},
			want:    false,
		},
		"AnyRune does not match ctrl key": {
			pattern: KeyPattern{AnyRune: true},
			event:   KeyEvent{Key: KeyCtrlC},
			want:    false,
		},
		"exact rune matches same rune": {
			pattern: KeyPattern{Rune: 'q'},
			event:   KeyEvent{Key: KeyRune, Rune: 'q'},
			want:    true,
		},
		"exact rune does not match different rune": {
			pattern: KeyPattern{Rune: 'q'},
			event:   KeyEvent{Key: KeyRune, Rune: 'w'},
			want:    false,
		},
		"exact rune does not match special key": {
			pattern: KeyPattern{Rune: 'q'},
			event:   KeyEvent{Key: KeyEscape},
			want:    false,
		},
		"exact key matches same key": {
			pattern: KeyPattern{Key: KeyEscape},
			event:   KeyEvent{Key: KeyEscape},
			want:    true,
		},
		"exact key does not match different key": {
			pattern: KeyPattern{Key: KeyEscape},
			event:   KeyEvent{Key: KeyEnter},
			want:    false,
		},
		"exact key matches ctrl key": {
			pattern: KeyPattern{Key: KeyCtrlC},
			event:   KeyEvent{Key: KeyCtrlC},
			want:    true,
		},
		"modifier required and present": {
			pattern: KeyPattern{Key: KeyRune, Mod: ModAlt},
			event:   KeyEvent{Key: KeyRune, Rune: 'x', Mod: ModAlt},
			want:    true,
		},
		"modifier required but absent": {
			pattern: KeyPattern{Key: KeyRune, Mod: ModAlt},
			event:   KeyEvent{Key: KeyRune, Rune: 'x', Mod: ModNone},
			want:    false,
		},
		"modifier required but wrong modifier": {
			pattern: KeyPattern{Key: KeyRune, Mod: ModAlt},
			event:   KeyEvent{Key: KeyRune, Rune: 'x', Mod: ModCtrl},
			want:    false,
		},
		"no modifier required ignores event mods": {
			pattern: KeyPattern{Key: KeyEscape},
			event:   KeyEvent{Key: KeyEscape, Mod: ModShift},
			want:    true,
		},
		"RequireNoMods matches event with no modifiers": {
			pattern: KeyPattern{Key: KeyTab, RequireNoMods: true},
			event:   KeyEvent{Key: KeyTab, Mod: ModNone},
			want:    true,
		},
		"RequireNoMods rejects event with shift": {
			pattern: KeyPattern{Key: KeyTab, RequireNoMods: true},
			event:   KeyEvent{Key: KeyTab, Mod: ModShift},
			want:    false,
		},
		"RequireNoMods rejects event with alt": {
			pattern: KeyPattern{Key: KeyTab, RequireNoMods: true},
			event:   KeyEvent{Key: KeyTab, Mod: ModAlt},
			want:    false,
		},
		"empty pattern matches nothing": {
			pattern: KeyPattern{},
			event:   KeyEvent{Key: KeyRune, Rune: 'a'},
			want:    false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			entry := &dispatchEntry{pattern: tt.pattern}
			got := entry.matches(tt.event)
			if got != tt.want {
				t.Errorf("matches(%+v) = %v, want %v", tt.event, got, tt.want)
			}
		})
	}
}

// --- Broadcast dispatch tests ---

func TestDispatch_BroadcastMultipleHandlers(t *testing.T) {
	var calls []int

	comp1 := &mockKeyComponent{
		keyMap: KeyMap{
			OnKey(KeyCtrlC, func(ke KeyEvent) { calls = append(calls, 1) }),
		},
	}
	comp2 := &mockKeyComponent{
		keyMap: KeyMap{
			OnKey(KeyCtrlC, func(ke KeyEvent) { calls = append(calls, 2) }),
		},
	}
	comp3 := &mockKeyComponent{
		keyMap: KeyMap{
			OnKey(KeyCtrlC, func(ke KeyEvent) { calls = append(calls, 3) }),
		},
	}

	root := buildTestTree(comp1, comp2, comp3)
	table, err := buildDispatchTable(root)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}

	table.dispatch(KeyEvent{Key: KeyCtrlC})

	if len(calls) != 3 {
		t.Fatalf("got %d handler calls, want 3", len(calls))
	}
	if calls[0] != 1 || calls[1] != 2 || calls[2] != 3 {
		t.Errorf("calls = %v, want [1 2 3]", calls)
	}
}

func TestDispatch_StopPreventsLaterHandlers(t *testing.T) {
	var calls []int

	comp1 := &mockKeyComponent{
		keyMap: KeyMap{
			OnKey(KeyEscape, func(ke KeyEvent) { calls = append(calls, 1) }),
		},
	}
	comp2 := &mockKeyComponent{
		keyMap: KeyMap{
			OnKeyStop(KeyEscape, func(ke KeyEvent) { calls = append(calls, 2) }),
		},
	}
	comp3 := &mockKeyComponent{
		keyMap: KeyMap{
			OnKey(KeyEscape, func(ke KeyEvent) { calls = append(calls, 3) }),
		},
	}

	root := buildTestTree(comp1, comp2, comp3)
	table, err := buildDispatchTable(root)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}

	table.dispatch(KeyEvent{Key: KeyEscape})

	// Handler 1 (broadcast) fires, handler 2 (stop) fires, handler 3 is blocked
	if len(calls) != 2 {
		t.Fatalf("got %d handler calls, want 2", len(calls))
	}
	if calls[0] != 1 || calls[1] != 2 {
		t.Errorf("calls = %v, want [1 2]", calls)
	}
}

func TestDispatch_TreeOrder(t *testing.T) {
	var calls []string

	parent := &mockKeyComponent{
		keyMap: KeyMap{
			OnKey(KeyEnter, func(ke KeyEvent) { calls = append(calls, "parent") }),
		},
	}
	child := &mockKeyComponent{
		keyMap: KeyMap{
			OnKey(KeyEnter, func(ke KeyEvent) { calls = append(calls, "child") }),
		},
	}

	root := buildNestedTestTree(parent, child)
	table, err := buildDispatchTable(root)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}

	table.dispatch(KeyEvent{Key: KeyEnter})

	if len(calls) != 2 {
		t.Fatalf("got %d handler calls, want 2", len(calls))
	}
	// Parent comes first in DFS order
	if calls[0] != "parent" || calls[1] != "child" {
		t.Errorf("calls = %v, want [parent child]", calls)
	}
}

func TestDispatch_UnifiedOrdering_ExactAndAnyRune(t *testing.T) {
	var calls []string

	// comp1 has an exact rune handler for 'a'
	comp1 := &mockKeyComponent{
		keyMap: KeyMap{
			OnRune('a', func(ke KeyEvent) { calls = append(calls, "exact-a") }),
		},
	}
	// comp2 has an AnyRune handler
	comp2 := &mockKeyComponent{
		keyMap: KeyMap{
			OnRunes(func(ke KeyEvent) { calls = append(calls, "any-rune") }),
		},
	}

	root := buildTestTree(comp1, comp2)
	table, err := buildDispatchTable(root)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}

	// Pressing 'a' should fire both in tree order
	table.dispatch(KeyEvent{Key: KeyRune, Rune: 'a'})

	if len(calls) != 2 {
		t.Fatalf("got %d handler calls, want 2", len(calls))
	}
	if calls[0] != "exact-a" || calls[1] != "any-rune" {
		t.Errorf("calls = %v, want [exact-a any-rune]", calls)
	}
}

func TestDispatch_AnyRuneMatchesPrintableOnly(t *testing.T) {
	called := false

	comp := &mockKeyComponent{
		keyMap: KeyMap{
			OnRunes(func(ke KeyEvent) { called = true }),
		},
	}

	root := buildTestTree(comp)
	table, err := buildDispatchTable(root)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}

	// Printable character should match
	table.dispatch(KeyEvent{Key: KeyRune, Rune: 'z'})
	if !called {
		t.Error("AnyRune should match printable character")
	}

	// Special key should not match
	called = false
	table.dispatch(KeyEvent{Key: KeyEscape})
	if called {
		t.Error("AnyRune should not match special key")
	}

	// Ctrl key should not match
	called = false
	table.dispatch(KeyEvent{Key: KeyCtrlC})
	if called {
		t.Error("AnyRune should not match ctrl key")
	}
}

func TestDispatch_ExactRuneMatch(t *testing.T) {
	type tc struct {
		patternRune rune
		eventRune   rune
		eventKey    Key
		wantCalled  bool
	}

	tests := map[string]tc{
		"matching rune fires": {
			patternRune: '/',
			eventRune:   '/',
			eventKey:    KeyRune,
			wantCalled:  true,
		},
		"different rune does not fire": {
			patternRune: '/',
			eventRune:   'q',
			eventKey:    KeyRune,
			wantCalled:  false,
		},
		"special key does not fire": {
			patternRune: '/',
			eventRune:   0,
			eventKey:    KeyEscape,
			wantCalled:  false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			called := false
			comp := &mockKeyComponent{
				keyMap: KeyMap{
					OnRune(tt.patternRune, func(ke KeyEvent) { called = true }),
				},
			}

			root := buildTestTree(comp)
			table, err := buildDispatchTable(root)
			if err != nil {
				t.Fatalf("buildDispatchTable: %v", err)
			}

			table.dispatch(KeyEvent{Key: tt.eventKey, Rune: tt.eventRune})
			if called != tt.wantCalled {
				t.Errorf("called = %v, want %v", called, tt.wantCalled)
			}
		})
	}
}

func TestDispatch_ExactKeyMatch(t *testing.T) {
	type tc struct {
		patternKey Key
		eventKey   Key
		wantCalled bool
	}

	tests := map[string]tc{
		"matching key fires": {
			patternKey: KeyEscape,
			eventKey:   KeyEscape,
			wantCalled: true,
		},
		"different key does not fire": {
			patternKey: KeyEscape,
			eventKey:   KeyEnter,
			wantCalled: false,
		},
		"ctrl key fires": {
			patternKey: KeyCtrlC,
			eventKey:   KeyCtrlC,
			wantCalled: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			called := false
			comp := &mockKeyComponent{
				keyMap: KeyMap{
					OnKey(tt.patternKey, func(ke KeyEvent) { called = true }),
				},
			}

			root := buildTestTree(comp)
			table, err := buildDispatchTable(root)
			if err != nil {
				t.Fatalf("buildDispatchTable: %v", err)
			}

			table.dispatch(KeyEvent{Key: tt.eventKey})
			if called != tt.wantCalled {
				t.Errorf("called = %v, want %v", called, tt.wantCalled)
			}
		})
	}
}

// --- Conflict validation tests ---

func TestDispatch_ConflictValidation_TwoStopHandlersSamePattern(t *testing.T) {
	comp1 := &mockKeyComponent{
		keyMap: KeyMap{
			OnKeyStop(KeyEscape, func(ke KeyEvent) {}),
		},
	}
	comp2 := &mockKeyComponent{
		keyMap: KeyMap{
			OnKeyStop(KeyEscape, func(ke KeyEvent) {}),
		},
	}

	root := buildTestTree(comp1, comp2)
	_, err := buildDispatchTable(root)
	if err == nil {
		t.Fatal("expected error for conflicting stop handlers, got nil")
	}
}

func TestDispatch_NoConflict_StopPlusBroadcast(t *testing.T) {
	comp1 := &mockKeyComponent{
		keyMap: KeyMap{
			OnKeyStop(KeyEscape, func(ke KeyEvent) {}),
		},
	}
	comp2 := &mockKeyComponent{
		keyMap: KeyMap{
			OnKey(KeyEscape, func(ke KeyEvent) {}),
		},
	}

	root := buildTestTree(comp1, comp2)
	_, err := buildDispatchTable(root)
	if err != nil {
		t.Fatalf("should not error for stop + broadcast: %v", err)
	}
}

func TestDispatch_NoConflict_TwoBroadcastHandlers(t *testing.T) {
	comp1 := &mockKeyComponent{
		keyMap: KeyMap{
			OnKey(KeyCtrlC, func(ke KeyEvent) {}),
		},
	}
	comp2 := &mockKeyComponent{
		keyMap: KeyMap{
			OnKey(KeyCtrlC, func(ke KeyEvent) {}),
		},
	}

	root := buildTestTree(comp1, comp2)
	_, err := buildDispatchTable(root)
	if err != nil {
		t.Fatalf("should not error for two broadcast handlers: %v", err)
	}
}

func TestDispatch_ConflictValidation_TwoStopAnyRune(t *testing.T) {
	comp1 := &mockKeyComponent{
		keyMap: KeyMap{
			OnRunesStop(func(ke KeyEvent) {}),
		},
	}
	comp2 := &mockKeyComponent{
		keyMap: KeyMap{
			OnRunesStop(func(ke KeyEvent) {}),
		},
	}

	root := buildTestTree(comp1, comp2)
	_, err := buildDispatchTable(root)
	if err == nil {
		t.Fatal("expected error for conflicting AnyRune stop handlers, got nil")
	}
}

func TestDispatch_ConflictValidation_DifferentPatterns(t *testing.T) {
	comp1 := &mockKeyComponent{
		keyMap: KeyMap{
			OnKeyStop(KeyEscape, func(ke KeyEvent) {}),
		},
	}
	comp2 := &mockKeyComponent{
		keyMap: KeyMap{
			OnKeyStop(KeyEnter, func(ke KeyEvent) {}),
		},
	}

	root := buildTestTree(comp1, comp2)
	_, err := buildDispatchTable(root)
	if err != nil {
		t.Fatalf("different patterns should not conflict: %v", err)
	}
}

// --- Edge case tests ---

func TestDispatch_NilKeyMap(t *testing.T) {
	comp := &mockKeyComponent{keyMap: nil}

	root := buildTestTree(comp)
	table, err := buildDispatchTable(root)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}
	if len(table.entries) != 0 {
		t.Errorf("expected 0 entries for nil KeyMap, got %d", len(table.entries))
	}
}

func TestDispatch_EmptyKeyMap(t *testing.T) {
	comp := &mockKeyComponent{keyMap: KeyMap{}}

	root := buildTestTree(comp)
	table, err := buildDispatchTable(root)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}
	if len(table.entries) != 0 {
		t.Errorf("expected 0 entries for empty KeyMap, got %d", len(table.entries))
	}
}

func TestDispatch_NonKeyListenerComponent(t *testing.T) {
	comp := &mockNoKeyComponent{}

	root := buildTestTree(comp)
	table, err := buildDispatchTable(root)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}
	if len(table.entries) != 0 {
		t.Errorf("expected 0 entries for non-KeyListener, got %d", len(table.entries))
	}
}

func TestDispatch_MixedComponents(t *testing.T) {
	var calls []int

	keyComp := &mockKeyComponent{
		keyMap: KeyMap{
			OnKey(KeyEnter, func(ke KeyEvent) { calls = append(calls, 1) }),
		},
	}
	noKeyComp := &mockNoKeyComponent{}
	keyComp2 := &mockKeyComponent{
		keyMap: KeyMap{
			OnKey(KeyEnter, func(ke KeyEvent) { calls = append(calls, 2) }),
		},
	}

	root := buildTestTree(keyComp, noKeyComp, keyComp2)
	table, err := buildDispatchTable(root)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}

	table.dispatch(KeyEvent{Key: KeyEnter})

	if len(calls) != 2 {
		t.Fatalf("got %d handler calls, want 2", len(calls))
	}
	if calls[0] != 1 || calls[1] != 2 {
		t.Errorf("calls = %v, want [1 2]", calls)
	}
}

func TestDispatch_NilDispatchTable(t *testing.T) {
	// dispatch on nil table should not panic
	var dt *dispatchTable
	dt.dispatch(KeyEvent{Key: KeyRune, Rune: 'a'})
}

func TestDispatch_EmptyTree(t *testing.T) {
	root := New()
	table, err := buildDispatchTable(root)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}
	if len(table.entries) != 0 {
		t.Errorf("expected 0 entries for empty tree, got %d", len(table.entries))
	}

	// Should not panic
	table.dispatch(KeyEvent{Key: KeyEscape})
}

func TestDispatch_NonMatchingKeyPassesThrough(t *testing.T) {
	called := false
	comp := &mockKeyComponent{
		keyMap: KeyMap{
			OnKey(KeyEscape, func(ke KeyEvent) { called = true }),
		},
	}

	root := buildTestTree(comp)
	table, err := buildDispatchTable(root)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}

	// Press a different key — handler should not fire
	table.dispatch(KeyEvent{Key: KeyEnter})
	if called {
		t.Error("handler should not fire for non-matching key")
	}
}

func TestDispatch_StopOnlyAffectsMatchingPattern(t *testing.T) {
	var calls []string

	comp := &mockKeyComponent{
		keyMap: KeyMap{
			OnKeyStop(KeyEscape, func(ke KeyEvent) { calls = append(calls, "escape") }),
			OnKey(KeyEnter, func(ke KeyEvent) { calls = append(calls, "enter") }),
		},
	}

	root := buildTestTree(comp)
	table, err := buildDispatchTable(root)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}

	// Press Enter — escape stop doesn't affect enter
	table.dispatch(KeyEvent{Key: KeyEnter})
	if len(calls) != 1 || calls[0] != "enter" {
		t.Errorf("calls = %v, want [enter]", calls)
	}
}

func TestDispatch_MultipleBindingsPerComponent(t *testing.T) {
	var calls []string

	comp := &mockKeyComponent{
		keyMap: KeyMap{
			OnKey(KeyCtrlC, func(ke KeyEvent) { calls = append(calls, "ctrl-c") }),
			OnRune('/', func(ke KeyEvent) { calls = append(calls, "slash") }),
			OnRunesStop(func(ke KeyEvent) { calls = append(calls, "any-rune") }),
		},
	}

	root := buildTestTree(comp)
	table, err := buildDispatchTable(root)
	if err != nil {
		t.Fatalf("buildDispatchTable: %v", err)
	}

	// Press ctrl+c
	table.dispatch(KeyEvent{Key: KeyCtrlC})
	if len(calls) != 1 || calls[0] != "ctrl-c" {
		t.Errorf("ctrl+c: calls = %v, want [ctrl-c]", calls)
	}

	// Press '/' — matches both exact rune and AnyRune, but AnyRune has Stop
	calls = nil
	table.dispatch(KeyEvent{Key: KeyRune, Rune: '/'})
	if len(calls) != 2 {
		t.Fatalf("'/': got %d calls, want 2", len(calls))
	}
	if calls[0] != "slash" || calls[1] != "any-rune" {
		t.Errorf("'/': calls = %v, want [slash any-rune]", calls)
	}
}

// --- buildDispatchTable entry count tests ---

func TestBuildDispatchTable_EntryCount(t *testing.T) {
	type tc struct {
		name       string
		components []Component
		wantCount  int
	}

	tests := []tc{
		{
			name:       "single component with one binding",
			components: []Component{&mockKeyComponent{keyMap: KeyMap{OnKey(KeyCtrlC, func(ke KeyEvent) {})}}},
			wantCount:  1,
		},
		{
			name: "single component with three bindings",
			components: []Component{&mockKeyComponent{keyMap: KeyMap{
				OnKey(KeyCtrlC, func(ke KeyEvent) {}),
				OnRune('q', func(ke KeyEvent) {}),
				OnRunes(func(ke KeyEvent) {}),
			}}},
			wantCount: 3,
		},
		{
			name: "two components with bindings",
			components: []Component{
				&mockKeyComponent{keyMap: KeyMap{OnKey(KeyCtrlC, func(ke KeyEvent) {})}},
				&mockKeyComponent{keyMap: KeyMap{OnKey(KeyEnter, func(ke KeyEvent) {})}},
			},
			wantCount: 2,
		},
		{
			name: "nil keymap component skipped",
			components: []Component{
				&mockKeyComponent{keyMap: nil},
				&mockKeyComponent{keyMap: KeyMap{OnKey(KeyEnter, func(ke KeyEvent) {})}},
			},
			wantCount: 1,
		},
		{
			name: "non-key-listener component skipped",
			components: []Component{
				&mockNoKeyComponent{},
				&mockKeyComponent{keyMap: KeyMap{OnKey(KeyEnter, func(ke KeyEvent) {})}},
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := buildTestTree(tt.components...)
			table, err := buildDispatchTable(root)
			if err != nil {
				t.Fatalf("buildDispatchTable: %v", err)
			}
			if len(table.entries) != tt.wantCount {
				t.Errorf("entries count = %d, want %d", len(table.entries), tt.wantCount)
			}
		})
	}
}

// --- Validate tests ---

func TestValidate_NoEntries(t *testing.T) {
	table := &dispatchTable{}
	if err := table.validate(); err != nil {
		t.Errorf("empty table should validate: %v", err)
	}
}

func TestValidate_SingleStop(t *testing.T) {
	table := &dispatchTable{
		entries: []dispatchEntry{
			{pattern: KeyPattern{Key: KeyEscape}, stop: true, position: 0},
		},
	}
	if err := table.validate(); err != nil {
		t.Errorf("single stop should validate: %v", err)
	}
}

func TestValidate_MultipleBroadcast(t *testing.T) {
	table := &dispatchTable{
		entries: []dispatchEntry{
			{pattern: KeyPattern{Key: KeyEscape}, stop: false, position: 0},
			{pattern: KeyPattern{Key: KeyEscape}, stop: false, position: 1},
		},
	}
	if err := table.validate(); err != nil {
		t.Errorf("multiple broadcast should validate: %v", err)
	}
}

func TestValidate_ConflictingStops(t *testing.T) {
	table := &dispatchTable{
		entries: []dispatchEntry{
			{pattern: KeyPattern{Key: KeyEscape}, stop: true, position: 0},
			{pattern: KeyPattern{Key: KeyEscape}, stop: true, position: 1},
		},
	}
	err := table.validate()
	if err == nil {
		t.Fatal("expected error for conflicting stops")
	}
}
