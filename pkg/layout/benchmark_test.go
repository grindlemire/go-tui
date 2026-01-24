package layout

import "testing"

// buildTree creates a tree with the specified branching factor and depth.
// Total nodes = sum of (branching^i) for i from 0 to depth = (branching^(depth+1) - 1) / (branching - 1)
func buildTree(branching, depth int) *Node {
	root := NewNode(DefaultStyle())
	root.Style.Width = Fixed(1000)
	root.Style.Height = Fixed(1000)
	root.Style.Direction = Row

	if depth > 0 {
		addChildrenRecursive(root, branching, depth-1)
	}

	return root
}

func addChildrenRecursive(parent *Node, branching, remainingDepth int) {
	for i := 0; i < branching; i++ {
		child := NewNode(DefaultStyle())
		child.Style.FlexGrow = 1

		// Alternate direction at each level
		if parent.Style.Direction == Row {
			child.Style.Direction = Column
		} else {
			child.Style.Direction = Row
		}

		parent.AddChild(child)

		if remainingDepth > 0 {
			addChildrenRecursive(child, branching, remainingDepth-1)
		}
	}
}

// buildLinearTree creates a tree with n nodes in a linear chain.
func buildLinearTree(n int) *Node {
	root := NewNode(DefaultStyle())
	root.Style.Width = Fixed(1000)
	root.Style.Height = Fixed(1000)
	root.Style.Direction = Row

	for i := 0; i < n; i++ {
		child := NewNode(DefaultStyle())
		child.Style.Width = Fixed(10)
		child.Style.Height = Fixed(100)
		root.AddChild(child)
	}

	return root
}

// countNodes counts the total number of nodes in a tree.
func countNodes(node *Node) int {
	if node == nil {
		return 0
	}
	count := 1
	for _, child := range node.Children {
		count += countNodes(child)
	}
	return count
}

// BenchmarkCalculate_10Nodes benchmarks layout calculation with ~10 nodes.
// Tree structure: branching=3, depth=2 = 1 + 3 + 9 = 13 nodes
func BenchmarkCalculate_10Nodes(b *testing.B) {
	root := buildTree(3, 2)
	nodeCount := countNodes(root)
	b.Logf("Node count: %d", nodeCount)

	// Initial calculation to warm up
	Calculate(root, 1000, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Mark root dirty to force full recalculation
		root.MarkDirty()
		Calculate(root, 1000, 1000)
	}
}

// BenchmarkCalculate_100Nodes benchmarks layout calculation with ~100 nodes.
// Tree structure: branching=3, depth=4 = 1 + 3 + 9 + 27 + 81 = 121 nodes
func BenchmarkCalculate_100Nodes(b *testing.B) {
	root := buildTree(3, 4)
	nodeCount := countNodes(root)
	b.Logf("Node count: %d", nodeCount)

	Calculate(root, 1000, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root.MarkDirty()
		Calculate(root, 1000, 1000)
	}
}

// BenchmarkCalculate_1000Nodes benchmarks layout calculation with ~1000 nodes.
// Linear tree: 1 root + 999 children = 1000 nodes
func BenchmarkCalculate_1000Nodes(b *testing.B) {
	root := buildLinearTree(999)
	nodeCount := countNodes(root)
	b.Logf("Node count: %d", nodeCount)

	Calculate(root, 10000, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root.MarkDirty()
		Calculate(root, 10000, 1000)
	}
}

// BenchmarkCalculate_Incremental benchmarks incremental layout
// when only a single leaf node is modified.
func BenchmarkCalculate_Incremental(b *testing.B) {
	root := buildTree(3, 4) // ~121 nodes
	Calculate(root, 1000, 1000)

	// Find a leaf node
	leaf := root
	for len(leaf.Children) > 0 {
		leaf = leaf.Children[0]
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Only mark the leaf dirty (should trigger path-to-root recalculation)
		leaf.MarkDirty()
		Calculate(root, 1000, 1000)
	}
}

// BenchmarkCalculate_IncrementalVsFull compares incremental vs full layout.
// Note: When a leaf is marked dirty, it propagates to the root, causing
// the entire tree to be recalculated. True incremental savings occur when
// only part of a tree (e.g., one subtree of many siblings) is dirty.
func BenchmarkCalculate_IncrementalVsFull(b *testing.B) {
	root := buildTree(3, 4) // ~121 nodes
	Calculate(root, 1000, 1000)

	// Find a leaf node
	leaf := root
	for len(leaf.Children) > 0 {
		leaf = leaf.Children[0]
	}

	b.Run("full", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			root.MarkDirty()
			Calculate(root, 1000, 1000)
		}
	})

	b.Run("incremental", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			leaf.MarkDirty()
			Calculate(root, 1000, 1000)
		}
	})
}

// BenchmarkCalculate_AllocationsPerNode verifies allocation behavior.
// Should allocate only for flexItem slices in nodes with children.
func BenchmarkCalculate_AllocationsPerNode(b *testing.B) {
	root := buildLinearTree(10) // 1 root + 10 children
	Calculate(root, 1000, 1000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		root.MarkDirty()
		Calculate(root, 1000, 1000)
	}

	// Expected: 1 allocation per Calculate call (for flexItem slice in root)
	// Leaf nodes don't allocate since they have no children
}

// BenchmarkNewNode benchmarks node creation.
func BenchmarkNewNode(b *testing.B) {
	style := DefaultStyle()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = NewNode(style)
	}
}

// BenchmarkDefaultStyle benchmarks style creation.
func BenchmarkDefaultStyle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DefaultStyle()
	}
}
