package layout

import "testing"

// testNode is a minimal Layoutable implementation for testing the layout algorithm.
// It provides the same tree structure and dirty tracking that Element uses.
type testNode struct {
	style    Style
	children []*testNode
	layout   Layout
	dirty    bool
	parent   *testNode
}

// newTestNode creates a new testNode with the given style.
func newTestNode(style Style) *testNode {
	return &testNode{
		style: style,
		dirty: true,
	}
}

// Implement Layoutable interface

func (n *testNode) LayoutStyle() Style { return n.style }

func (n *testNode) LayoutChildren() []Layoutable {
	result := make([]Layoutable, len(n.children))
	for i, child := range n.children {
		result[i] = child
	}
	return result
}

func (n *testNode) SetLayout(l Layout) { n.layout = l }
func (n *testNode) GetLayout() Layout  { return n.layout }
func (n *testNode) IsDirty() bool      { return n.dirty }
func (n *testNode) SetDirty(d bool)    { n.dirty = d }

// Additional methods for testing

// AddChild appends children and marks this node dirty.
func (n *testNode) AddChild(children ...*testNode) {
	for _, child := range children {
		child.parent = n
		n.children = append(n.children, child)
	}
	n.markDirty()
}

// RemoveChild removes a child by pointer and marks dirty.
// Returns true if the child was found and removed.
func (n *testNode) RemoveChild(child *testNode) bool {
	for i, c := range n.children {
		if c == child {
			// Remove by swapping with last element and truncating
			n.children[i] = n.children[len(n.children)-1]
			n.children = n.children[:len(n.children)-1]
			child.parent = nil
			n.markDirty()
			return true
		}
	}
	return false
}

// SetStyle updates the style and marks the node dirty.
func (n *testNode) SetStyle(style Style) {
	n.style = style
	n.markDirty()
}

// markDirty marks this node and all ancestors as needing recalculation.
func (n *testNode) markDirty() {
	for node := n; node != nil && !node.dirty; node = node.parent {
		node.dirty = true
	}
}

// Tests for the testNode implementation (ensures it implements Layoutable correctly)

func TestTestNode_ImplementsLayoutable(t *testing.T) {
	// Compile-time check that testNode implements Layoutable
	var _ Layoutable = (*testNode)(nil)
}

func TestTestNode_NewNode(t *testing.T) {
	style := DefaultStyle()
	style.Width = Fixed(100)
	style.Height = Fixed(50)

	node := newTestNode(style)

	if node.style.Width != Fixed(100) {
		t.Errorf("newTestNode style.Width = %+v, want Fixed(100)", node.style.Width)
	}
	if node.style.Height != Fixed(50) {
		t.Errorf("newTestNode style.Height = %+v, want Fixed(50)", node.style.Height)
	}
	if !node.IsDirty() {
		t.Error("newTestNode should be dirty")
	}
	if len(node.children) != 0 {
		t.Errorf("newTestNode should have no children, got %d", len(node.children))
	}
}

func TestTestNode_AddChild(t *testing.T) {
	parent := newTestNode(DefaultStyle())
	child1 := newTestNode(DefaultStyle())
	child2 := newTestNode(DefaultStyle())

	// Clear dirty flag to test that AddChild marks dirty
	parent.dirty = false

	parent.AddChild(child1, child2)

	if len(parent.children) != 2 {
		t.Errorf("AddChild: len(children) = %d, want 2", len(parent.children))
	}
	if parent.children[0] != child1 {
		t.Error("AddChild: first child mismatch")
	}
	if parent.children[1] != child2 {
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

func TestTestNode_RemoveChild(t *testing.T) {
	type tc struct {
		setup       func() (*testNode, *testNode, *testNode)
		removeChild func(*testNode, *testNode, *testNode) *testNode // returns child to remove
		expectFound bool
		expectLen   int
	}

	tests := map[string]tc{
		"remove existing child": {
			setup: func() (*testNode, *testNode, *testNode) {
				parent := newTestNode(DefaultStyle())
				child1 := newTestNode(DefaultStyle())
				child2 := newTestNode(DefaultStyle())
				parent.AddChild(child1, child2)
				parent.dirty = false
				return parent, child1, child2
			},
			removeChild: func(parent, child1, child2 *testNode) *testNode { return child1 },
			expectFound: true,
			expectLen:   1,
		},
		"remove non-existent child": {
			setup: func() (*testNode, *testNode, *testNode) {
				parent := newTestNode(DefaultStyle())
				child1 := newTestNode(DefaultStyle())
				parent.AddChild(child1)
				parent.dirty = false
				otherChild := newTestNode(DefaultStyle())
				return parent, child1, otherChild
			},
			removeChild: func(parent, child1, otherChild *testNode) *testNode { return otherChild },
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
			if len(parent.children) != tt.expectLen {
				t.Errorf("len(children) = %d, want %d", len(parent.children), tt.expectLen)
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

func TestTestNode_SetStyle(t *testing.T) {
	node := newTestNode(DefaultStyle())
	node.dirty = false

	newStyle := DefaultStyle()
	newStyle.Width = Fixed(200)
	node.SetStyle(newStyle)

	if node.style.Width != Fixed(200) {
		t.Errorf("SetStyle did not update style, Width = %+v", node.style.Width)
	}
	if !node.IsDirty() {
		t.Error("SetStyle should mark node dirty")
	}
}

func TestTestNode_MarkDirty_PropagatesUp(t *testing.T) {
	// Build tree: root -> middle -> leaf
	root := newTestNode(DefaultStyle())
	middle := newTestNode(DefaultStyle())
	leaf := newTestNode(DefaultStyle())
	root.AddChild(middle)
	middle.AddChild(leaf)

	// Clear all dirty flags
	root.dirty = false
	middle.dirty = false
	leaf.dirty = false

	// Mark leaf dirty
	leaf.markDirty()

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

func TestTestNode_MarkDirty_StopsAtAlreadyDirty(t *testing.T) {
	// Build tree: root -> middle -> leaf
	root := newTestNode(DefaultStyle())
	middle := newTestNode(DefaultStyle())
	leaf := newTestNode(DefaultStyle())
	root.AddChild(middle)
	middle.AddChild(leaf)

	// Clear all dirty flags
	root.dirty = false
	middle.dirty = false
	leaf.dirty = false

	// Mark middle dirty first
	middle.dirty = true

	// Mark leaf dirty - should stop at middle since it's already dirty
	leaf.markDirty()

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

func TestTestNode_DirtyPropagationAfterAddChild(t *testing.T) {
	// Build tree: root -> parent
	root := newTestNode(DefaultStyle())
	parent := newTestNode(DefaultStyle())
	root.AddChild(parent)

	// Clear all dirty flags
	root.dirty = false
	parent.dirty = false

	// Add new child to parent
	child := newTestNode(DefaultStyle())
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
