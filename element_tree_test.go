package tui

import (
	"testing"
)

func TestElement_NotFocusableByDefault(t *testing.T) {
	e := New()

	if e.IsFocusable() {
		t.Error("element should not be focusable by default")
	}
	if e.IsFocused() {
		t.Error("element should not be focused by default")
	}
}

// --- Child Notification Tests ---

func TestElement_SetOnChildAdded_Callback(t *testing.T) {
	root := New()
	var addedChildren []*Element

	root.SetOnChildAdded(func(child *Element) {
		addedChildren = append(addedChildren, child)
	})

	child := New()
	root.AddChild(child)

	if len(addedChildren) != 1 || addedChildren[0] != child {
		t.Error("onChildAdded should be called when child is added")
	}
}

func TestElement_AddChild_TriggersRootCallback(t *testing.T) {
	root := New()
	middle := New()
	root.AddChild(middle)

	var addedChildren []*Element
	root.SetOnChildAdded(func(child *Element) {
		addedChildren = append(addedChildren, child)
	})

	leaf := New()
	middle.AddChild(leaf)

	if len(addedChildren) != 1 || addedChildren[0] != leaf {
		t.Error("onChildAdded should be called on root when leaf is added to middle")
	}
}

func TestElement_SetOnFocusableAdded_Callback(t *testing.T) {
	root := New()
	var addedFocusables []Focusable

	root.SetOnFocusableAdded(func(f Focusable) {
		addedFocusables = append(addedFocusables, f)
	})

	focusable := New(WithOnFocus(func(*Element) {}))
	root.AddChild(focusable)

	if len(addedFocusables) != 1 {
		t.Errorf("onFocusableAdded should be called, got %d calls", len(addedFocusables))
	}
}

func TestElement_SetOnFocusableAdded_NotCalledForNonFocusable(t *testing.T) {
	root := New()
	var addedFocusables []Focusable

	root.SetOnFocusableAdded(func(f Focusable) {
		addedFocusables = append(addedFocusables, f)
	})

	nonFocusable := New()
	root.AddChild(nonFocusable)

	if len(addedFocusables) != 0 {
		t.Error("onFocusableAdded should not be called for non-focusable elements")
	}
}

func TestElement_WalkFocusables(t *testing.T) {
	type tc struct {
		setupTree     func() *Element
		expectedCount int
	}

	tests := map[string]tc{
		"empty tree": {
			setupTree: func() *Element {
				return New()
			},
			expectedCount: 0,
		},
		"root is focusable": {
			setupTree: func() *Element {
				return New(WithOnFocus(func(*Element) {}))
			},
			expectedCount: 1,
		},
		"child is focusable": {
			setupTree: func() *Element {
				root := New()
				root.AddChild(New(WithOnFocus(func(*Element) {})))
				return root
			},
			expectedCount: 1,
		},
		"multiple focusables in tree": {
			setupTree: func() *Element {
				root := New()
				root.AddChild(New(WithOnFocus(func(*Element) {})))
				root.AddChild(New(WithOnBlur(func(*Element) {})))
				middle := New()
				middle.AddChild(New(WithFocusable(true)))
				root.AddChild(middle)
				return root
			},
			expectedCount: 3,
		},
		"mixed focusable and non-focusable": {
			setupTree: func() *Element {
				root := New()
				root.AddChild(New(WithOnFocus(func(*Element) {})))
				root.AddChild(New()) // non-focusable
				root.AddChild(New(WithOnBlur(func(*Element) {})))
				return root
			},
			expectedCount: 2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			root := tt.setupTree()
			var found []Focusable

			root.WalkFocusables(func(f Focusable) {
				found = append(found, f)
			})

			if len(found) != tt.expectedCount {
				t.Errorf("WalkFocusables found %d, want %d", len(found), tt.expectedCount)
			}
		})
	}
}
