package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

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
	rootPath        string
	tree            []Node
	cursor          *tui.State[int]
	expanded        *tui.State[map[string]bool]
	scrollY         *tui.State[int]
	scrollContainer *tui.Ref
}

// readDir reads a directory and returns its children as Nodes, sorted dirs-first then alphabetically.
func readDir(path string) []Node {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}

	var children []Node
	for _, entry := range entries {
		// Skip hidden files/directories
		if entry.Name()[0] == '.' {
			continue
		}
		node := Node{Name: entry.Name()}
		if entry.IsDir() {
			node.Children = readDir(filepath.Join(path, entry.Name()))
			if node.Children == nil {
				// Empty directory: use a sentinel so it's still recognized as a dir
				node.Children = []Node{}
			}
		}
		children = append(children, node)
	}
	sortChildren(children)
	return children
}

func sortChildren(children []Node) {
	sort.Slice(children, func(i, j int) bool {
		iDir := children[i].Children != nil
		jDir := children[j].Children != nil
		if iDir != jDir {
			return iDir // directories first
		}
		return children[i].Name < children[j].Name
	})
}

// DirectoryTree creates a new directory tree component rooted at the given path.
func DirectoryTree(root string) *directoryTree {
	name := filepath.Base(root)
	rootNode := Node{
		Name:     name,
		Children: readDir(root),
	}
	if rootNode.Children == nil {
		rootNode.Children = []Node{}
	}
	tree := []Node{rootNode}
	return &directoryTree{
		rootPath:        root,
		cursor:          tui.NewState(0),
		expanded:        tui.NewState(map[string]bool{name: true}),
		scrollY:         tui.NewState(0),
		tree:            tree,
		scrollContainer: tui.NewRef(),
	}
}

// navigateUp re-roots the tree at the parent of the current root.
func (d *directoryTree) navigateUp() {
	parent := filepath.Dir(d.rootPath)
	if parent == d.rootPath {
		return // already at filesystem root
	}
	d.rootPath = parent
	name := filepath.Base(parent)
	rootNode := Node{
		Name:     name,
		Children: readDir(parent),
	}
	if rootNode.Children == nil {
		rootNode.Children = []Node{}
	}
	d.tree = []Node{rootNode}
	d.cursor.Set(0)
	d.expanded.Set(map[string]bool{name: true})
	d.scrollY.Set(0)
}

// selectedPath returns the path of the currently selected node.
func (d *directoryTree) selectedPath() string {
	visible := d.flatten()
	cur := d.cursor.Get()
	if cur >= len(visible) {
		return ""
	}
	return visible[cur].path
}

// scrollToCursor adjusts scrollY state so the cursor row is visible.
func (d *directoryTree) scrollToCursor() {
	el := d.scrollContainer.El()
	if el == nil {
		return
	}
	cur := d.cursor.Get()
	_, vpH := el.ViewportSize()
	y := d.scrollY.Get()
	if cur < y {
		d.scrollY.Set(cur)
	} else if cur >= y+vpH {
		d.scrollY.Set(cur - vpH + 1)
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
	isDir := n.Children != nil
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

// isOnPath returns true if the given node's path is an ancestor of (or equal to) the selected node's path.
func isOnPath(vn visibleNode, selectedPath string) bool {
	if vn.path == selectedPath {
		return true
	}
	// Check if selectedPath starts with this node's path followed by "/"
	return len(selectedPath) > len(vn.path) && selectedPath[:len(vn.path)+1] == vn.path+"/"
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
	d.scrollToCursor()
}

func (d *directoryTree) moveDown() {
	visible := d.flatten()
	d.cursor.Update(func(v int) int {
		if v < len(visible)-1 {
			return v + 1
		}
		return v
	})
	d.scrollToCursor()
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
		d.navigateUp()
		return
	}
	// Find parent: walk backwards for a node at depth-1 that is a directory
	parentPath := vn.path[:len(vn.path)-len("/"+vn.node.Name)]
	for i := cur - 1; i >= 0; i-- {
		if visible[i].path == parentPath {
			d.cursor.Set(i)
			d.scrollToCursor()
			return
		}
	}
}

templ (d *directoryTree) Render() {
	<div class="flex-col w-full h-full border-rounded border-cyan">
		<div class="flex-col p-1">
			<span class="text-gradient-cyan-magenta font-bold">Directory Tree</span>
			<span class="text-cyan font-dim">{d.selectedPath()}</span>
		</div>
		<hr class="border-single" />
		<div class="flex-col grow overflow-y-scroll scrollbar-cyan scrollbar-thumb-bright-cyan"
			ref={d.scrollContainer} scrollOffset={0, d.scrollY.Get()}>
			@for i, vn := range d.flatten() {
				@if i == d.cursor.Get() {
					<span class="bg-bright-black text-white">{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
				} @else {
					@if isOnPath(vn, d.selectedPath()) {
						<span class="text-cyan font-bold">{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
					} @else {
						@if vn.isDir {
							<span class="font-bold">{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
						} @else {
							<span>{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
						}
					}
				}
			}
		</div>
		<hr class="border-single" />
		<div class="flex justify-center p-1">
			<span class="font-dim">j/k: navigate | enter/l: expand | h: collapse | q: quit</span>
		</div>
	</div>
}
