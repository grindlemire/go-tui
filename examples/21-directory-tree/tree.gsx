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

func (d *directoryTree) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { d.moveUp() }),
		tui.OnRune('k', func(ke tui.KeyEvent) { d.moveUp() }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { d.moveDown() }),
		tui.OnRune('j', func(ke tui.KeyEvent) { d.moveDown() }),
		tui.OnKey(tui.KeyEnter, func(ke tui.KeyEvent) { d.toggle() }),
		tui.OnKey(tui.KeyRight, func(ke tui.KeyEvent) { d.toggle() }),
		tui.OnRune('l', func(ke tui.KeyEvent) { d.toggle() }),
		tui.OnKey(tui.KeyLeft, func(ke tui.KeyEvent) { d.collapseOrParent() }),
		tui.OnRune('h', func(ke tui.KeyEvent) { d.collapseOrParent() }),
	}
}

func (d *directoryTree) moveUp() {
	d.cursor.Update(func(v int) int {
		if v > 0 {
			return v - 1
		}
		return v
	})
}

func (d *directoryTree) moveDown() {
	visible := d.flatten()
	d.cursor.Update(func(v int) int {
		if v < len(visible)-1 {
			return v + 1
		}
		return v
	})
}

func (d *directoryTree) toggle() {
	visible := d.flatten()
	cur := d.cursor.Get()
	if cur >= len(visible) {
		return
	}
	vn := visible[cur]
	if !vn.isDir {
		return
	}
	d.expanded.Update(func(m map[string]bool) map[string]bool {
		newMap := make(map[string]bool, len(m))
		for k, v := range m {
			newMap[k] = v
		}
		if newMap[vn.path] {
			delete(newMap, vn.path)
		} else {
			newMap[vn.path] = true
		}
		return newMap
	})
}

func (d *directoryTree) collapseOrParent() {
	visible := d.flatten()
	cur := d.cursor.Get()
	if cur >= len(visible) {
		return
	}
	vn := visible[cur]

	// If on an expanded directory, collapse it
	expanded := d.expanded.Get()
	if vn.isDir && expanded[vn.path] {
		d.expanded.Update(func(m map[string]bool) map[string]bool {
			newMap := make(map[string]bool, len(m))
			for k, v := range m {
				newMap[k] = v
			}
			delete(newMap, vn.path)
			return newMap
		})
		return
	}

	// Otherwise, jump to parent directory
	if vn.depth == 0 {
		return
	}
	// Find parent: walk backwards for a node at depth-1 that is a directory
	parentPath := vn.path[:len(vn.path)-len("/"+vn.node.Name)]
	for i := cur - 1; i >= 0; i-- {
		if visible[i].path == parentPath {
			d.cursor.Set(i)
			return
		}
	}
}

templ (d *directoryTree) Render() {
	<div class="flex-col p-1 border-rounded border-cyan">
		<span class="text-gradient-cyan-magenta font-bold">Directory Tree</span>
		<hr class="border-single" />
		<div class="flex-col">
			@for i, vn := range d.flatten() {
				@if i == d.cursor.Get() {
					<span class="bg-bright-black text-white">{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
				} @else {
					@if vn.isDir {
						<span class="text-cyan font-bold">{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
					} @else {
						<span>{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
					}
				}
			}
		</div>
		<hr class="border-single" />
		<div class="flex justify-center">
			<span class="font-dim">j/k: navigate | enter/l: expand | h: collapse | q: quit</span>
		</div>
	</div>
}
