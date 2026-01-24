package layout

// Layout holds the computed position and size after layout calculation.
type Layout struct {
	// Rect is the border box—the space allocated by the parent after
	// applying this node's margin. Use for hit testing and bounds.
	Rect Rect

	// ContentRect is Rect minus padding—the area where children are placed.
	// Use for rendering content and positioning children.
	ContentRect Rect
}

// Node represents an element in the layout tree.
type Node struct {
	// Configuration (user-set)
	Style    Style
	Children []*Node

	// Computed (set by layout engine)
	Layout Layout

	// Internal state
	dirty  bool  // Needs recalculation
	parent *Node // Back-pointer for dirty propagation
}

// NewNode creates a new node with the given style.
func NewNode(style Style) *Node {
	return &Node{
		Style: style,
		dirty: true, // New nodes need layout
	}
}

// AddChild appends children and marks this node dirty.
func (n *Node) AddChild(children ...*Node) {
	for _, child := range children {
		child.parent = n
		n.Children = append(n.Children, child)
	}
	n.MarkDirty()
}

// RemoveChild removes a child by pointer and marks dirty.
// Returns true if the child was found and removed.
func (n *Node) RemoveChild(child *Node) bool {
	for i, c := range n.Children {
		if c == child {
			// Remove by swapping with last element and truncating
			n.Children[i] = n.Children[len(n.Children)-1]
			n.Children = n.Children[:len(n.Children)-1]
			child.parent = nil
			n.MarkDirty()
			return true
		}
	}
	return false
}

// SetStyle updates the style and marks the node dirty.
func (n *Node) SetStyle(style Style) {
	n.Style = style
	n.MarkDirty()
}

// MarkDirty marks this node and all ancestors as needing recalculation.
func (n *Node) MarkDirty() {
	for node := n; node != nil && !node.dirty; node = node.parent {
		node.dirty = true
	}
}

// IsDirty returns whether this node needs recalculation.
func (n *Node) IsDirty() bool {
	return n.dirty
}
