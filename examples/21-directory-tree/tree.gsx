package main

import (
	"fmt"

	tui "github.com/grindlemire/go-tui"
)

// Ensure fmt is used.
var _ = fmt.Sprintf

// Node represents a file or directory in the tree.
type Node struct {
	Name     string
	Children []Node
}

// visibleNode is a flattened node for rendering.
type visibleNode struct {
	node      Node
	depth     int
	path      string
	isDir     bool
	isLast    bool
	ancestors []bool
}

// directoryTree is a foldable directory tree component.
type directoryTree struct {
	tree     []Node
	cursor   *tui.State[int]
	expanded *tui.State[map[string]bool]
}

// DirectoryTree creates a new directory tree component with sample data.
func DirectoryTree() *directoryTree {
	return &directoryTree{
		cursor:   tui.NewState(0),
		expanded: tui.NewState(map[string]bool{"myproject": true}),
		tree: []Node{
			{Name: "myproject", Children: []Node{
				{Name: "README.md"},
				{Name: "go.mod"},
				{Name: "go.sum"},
				{Name: "main.go"},
				{Name: "cmd", Children: []Node{
					{Name: "server", Children: []Node{
						{Name: "main.go"},
						{Name: "config.go"},
					}},
				}},
				{Name: "internal", Children: []Node{
					{Name: "api", Children: []Node{
						{Name: "handler.go"},
						{Name: "middleware.go"},
						{Name: "routes.go"},
					}},
					{Name: "db", Children: []Node{
						{Name: "migrations", Children: []Node{
							{Name: "001_init.sql"},
							{Name: "002_users.sql"},
						}},
						{Name: "connection.go"},
						{Name: "queries.go"},
					}},
					{Name: "models", Children: []Node{
						{Name: "user.go"},
						{Name: "post.go"},
					}},
				}},
				{Name: "pkg", Children: []Node{
					{Name: "logger", Children: []Node{
						{Name: "logger.go"},
					}},
				}},
				{Name: ".gitignore"},
			}},
		},
	}
}

func (d *directoryTree) flatten() []visibleNode {
	var result []visibleNode
	expanded := d.expanded.Get()
	for i, node := range d.tree {
		d.flattenNode(node, 0, node.Name, i == len(d.tree)-1, nil, expanded, &result)
	}
	return result
}

func (d *directoryTree) flattenNode(n Node, depth int, path string, isLast bool, ancestors []bool, expanded map[string]bool, result *[]visibleNode) {
	isDir := len(n.Children) > 0
	*result = append(*result, visibleNode{
		node:      n,
		depth:     depth,
		path:      path,
		isDir:     isDir,
		isLast:    isLast,
		ancestors: ancestors,
	})
	if isDir && expanded[path] {
		newAncestors := make([]bool, len(ancestors)+1)
		copy(newAncestors, ancestors)
		newAncestors[len(ancestors)] = isLast
		for i, child := range n.Children {
			childPath := path + "/" + child.Name
			d.flattenNode(child, depth+1, childPath, i == len(n.Children)-1, newAncestors, expanded, result)
		}
	}
}

func buildPrefix(vn visibleNode) string {
	if vn.depth == 0 {
		return ""
	}
	prefix := ""
	for i := 0; i < vn.depth-1; i++ {
		if vn.ancestors[i+1] {
			prefix += "    "
		} else {
			prefix += "│   "
		}
	}
	if vn.isLast {
		prefix += "└── "
	} else {
		prefix += "├── "
	}
	return prefix
}

func nodeLabel(vn visibleNode, expanded map[string]bool) string {
	if vn.isDir {
		if expanded[vn.path] {
			return "▼ " + vn.node.Name
		}
		return "▶ " + vn.node.Name
	}
	return vn.node.Name
}

templ (d *directoryTree) Render() {
	<div class="flex-col">
		<span>placeholder</span>
	</div>
}
