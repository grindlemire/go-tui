package layout

import "testing"

func TestNewNode(t *testing.T) {
	style := DefaultStyle()
	style.Width = Fixed(100)
	style.Height = Fixed(50)

	node := NewNode(style)

	if node.Style.Width != Fixed(100) {
		t.Errorf("NewNode style.Width = %+v, want Fixed(100)", node.Style.Width)
	}
	if node.Style.Height != Fixed(50) {
		t.Errorf("NewNode style.Height = %+v, want Fixed(50)", node.Style.Height)
	}
	if !node.IsDirty() {
		t.Error("NewNode should be dirty")
	}
	if len(node.Children) != 0 {
		t.Errorf("NewNode should have no children, got %d", len(node.Children))
	}
}

func TestNode_AddChild(t *testing.T) {
	parent := NewNode(DefaultStyle())
	child1 := NewNode(DefaultStyle())
	child2 := NewNode(DefaultStyle())

	// Clear dirty flag to test that AddChild marks dirty
	parent.dirty = false

	parent.AddChild(child1, child2)

	if len(parent.Children) != 2 {
		t.Errorf("AddChild: len(Children) = %d, want 2", len(parent.Children))
	}
	if parent.Children[0] != child1 {
		t.Error("AddChild: first child mismatch")
	}
	if parent.Children[1] != child2 {
		t.Error("AddChild: second child mismatch")
	}
	if child1.parent != parent {
		t.Error("AddChild: child1.parent not set")
	}
	if child2.parent != parent {
		t.Error("AddChild: child2.parent not set")
	}
	if !parent.IsDirty() {
		t.Error("AddChild should mark parent dirty")
	}
}

func TestNode_RemoveChild(t *testing.T) {
	type tc struct {
		setup       func() (*Node, *Node, *Node)
		removeChild func(*Node, *Node, *Node) *Node // returns child to remove
		expectFound bool
		expectLen   int
	}

	tests := map[string]tc{
		"remove existing child": {
			setup: func() (*Node, *Node, *Node) {
				parent := NewNode(DefaultStyle())
				child1 := NewNode(DefaultStyle())
				child2 := NewNode(DefaultStyle())
				parent.AddChild(child1, child2)
				parent.dirty = false
				return parent, child1, child2
			},
			removeChild: func(parent, child1, child2 *Node) *Node { return child1 },
			expectFound: true,
			expectLen:   1,
		},
		"remove non-existent child": {
			setup: func() (*Node, *Node, *Node) {
				parent := NewNode(DefaultStyle())
				child1 := NewNode(DefaultStyle())
				parent.AddChild(child1)
				parent.dirty = false
				otherChild := NewNode(DefaultStyle())
				return parent, child1, otherChild
			},
			removeChild: func(parent, child1, otherChild *Node) *Node { return otherChild },
			expectFound: false,
			expectLen:   1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			parent, child1, child2 := tt.setup()
			toRemove := tt.removeChild(parent, child1, child2)
			wasDirty := parent.IsDirty()

			found := parent.RemoveChild(toRemove)

			if found != tt.expectFound {
				t.Errorf("RemoveChild returned %v, want %v", found, tt.expectFound)
			}
			if len(parent.Children) != tt.expectLen {
				t.Errorf("len(Children) = %d, want %d", len(parent.Children), tt.expectLen)
			}
			if tt.expectFound {
				if toRemove.parent != nil {
					t.Error("removed child's parent should be nil")
				}
				if !parent.IsDirty() {
					t.Error("RemoveChild should mark parent dirty")
				}
			} else {
				// Parent dirty state should not change if child not found
				if parent.IsDirty() != wasDirty {
					t.Error("RemoveChild should not change dirty state when child not found")
				}
			}
		})
	}
}

func TestNode_SetStyle(t *testing.T) {
	node := NewNode(DefaultStyle())
	node.dirty = false

	newStyle := DefaultStyle()
	newStyle.Width = Fixed(200)
	node.SetStyle(newStyle)

	if node.Style.Width != Fixed(200) {
		t.Errorf("SetStyle did not update style, Width = %+v", node.Style.Width)
	}
	if !node.IsDirty() {
		t.Error("SetStyle should mark node dirty")
	}
}

func TestNode_MarkDirty_PropagatesUp(t *testing.T) {
	// Build tree: root -> middle -> leaf
	root := NewNode(DefaultStyle())
	middle := NewNode(DefaultStyle())
	leaf := NewNode(DefaultStyle())
	root.AddChild(middle)
	middle.AddChild(leaf)

	// Clear all dirty flags
	root.dirty = false
	middle.dirty = false
	leaf.dirty = false

	// Mark leaf dirty
	leaf.MarkDirty()

	if !leaf.IsDirty() {
		t.Error("leaf should be dirty")
	}
	if !middle.IsDirty() {
		t.Error("middle should be dirty (propagated from leaf)")
	}
	if !root.IsDirty() {
		t.Error("root should be dirty (propagated from leaf)")
	}
}

func TestNode_MarkDirty_StopsAtAlreadyDirty(t *testing.T) {
	// Build tree: root -> middle -> leaf
	root := NewNode(DefaultStyle())
	middle := NewNode(DefaultStyle())
	leaf := NewNode(DefaultStyle())
	root.AddChild(middle)
	middle.AddChild(leaf)

	// Clear all dirty flags
	root.dirty = false
	middle.dirty = false
	leaf.dirty = false

	// Mark middle dirty first
	middle.dirty = true

	// Mark leaf dirty - should stop at middle since it's already dirty
	leaf.MarkDirty()

	if !leaf.IsDirty() {
		t.Error("leaf should be dirty")
	}
	if !middle.IsDirty() {
		t.Error("middle should still be dirty")
	}
	// Root should still be clean because propagation stopped at middle
	if root.IsDirty() {
		t.Error("root should still be clean (propagation stopped at middle)")
	}
}

func TestNode_DirtyPropagationAfterAddChild(t *testing.T) {
	// Build tree: root -> parent
	root := NewNode(DefaultStyle())
	parent := NewNode(DefaultStyle())
	root.AddChild(parent)

	// Clear all dirty flags
	root.dirty = false
	parent.dirty = false

	// Add new child to parent
	child := NewNode(DefaultStyle())
	parent.AddChild(child)

	if !parent.IsDirty() {
		t.Error("parent should be dirty after AddChild")
	}
	if !root.IsDirty() {
		t.Error("root should be dirty (propagated from parent)")
	}
}

func TestDefaultStyle(t *testing.T) {
	style := DefaultStyle()

	if !style.Width.IsAuto() {
		t.Error("DefaultStyle Width should be Auto")
	}
	if !style.Height.IsAuto() {
		t.Error("DefaultStyle Height should be Auto")
	}
	if style.MinWidth != Fixed(0) {
		t.Errorf("DefaultStyle MinWidth = %+v, want Fixed(0)", style.MinWidth)
	}
	if style.MinHeight != Fixed(0) {
		t.Errorf("DefaultStyle MinHeight = %+v, want Fixed(0)", style.MinHeight)
	}
	if !style.MaxWidth.IsAuto() {
		t.Error("DefaultStyle MaxWidth should be Auto")
	}
	if !style.MaxHeight.IsAuto() {
		t.Error("DefaultStyle MaxHeight should be Auto")
	}
	if style.Direction != Row {
		t.Errorf("DefaultStyle Direction = %v, want Row", style.Direction)
	}
	if style.AlignItems != AlignStretch {
		t.Errorf("DefaultStyle AlignItems = %v, want AlignStretch", style.AlignItems)
	}
	if style.FlexShrink != 1.0 {
		t.Errorf("DefaultStyle FlexShrink = %v, want 1.0", style.FlexShrink)
	}
	if style.FlexGrow != 0 {
		t.Errorf("DefaultStyle FlexGrow = %v, want 0", style.FlexGrow)
	}
	if style.Gap != 0 {
		t.Errorf("DefaultStyle Gap = %v, want 0", style.Gap)
	}
	if style.AlignSelf != nil {
		t.Errorf("DefaultStyle AlignSelf should be nil, got %v", style.AlignSelf)
	}
}
