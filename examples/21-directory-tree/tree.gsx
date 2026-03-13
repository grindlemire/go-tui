package main

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	tui "github.com/grindlemire/go-tui"
)

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
	onPath    bool
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

// DirectoryTree creates a new directory tree component rooted at the given path.
func DirectoryTree(root string) *directoryTree {
	rootNode := buildRootNode(root)
	return &directoryTree{
		rootPath:        root,
		tree:            []Node{rootNode},
		cursor:          tui.NewState(0),
		expanded:        tui.NewState(map[string]bool{rootNode.Name: true}),
		scrollY:         tui.NewState(0),
		scrollContainer: tui.NewRef(),
	}
}

// navigateUp re-roots the tree at the parent of the current root.
func (d *directoryTree) navigateUp() {
	parent := filepath.Dir(d.rootPath)
	if parent == d.rootPath {
		return
	}
	d.rootPath = parent
	rootNode := buildRootNode(parent)
	d.tree = []Node{rootNode}
	d.cursor.Set(0)
	d.expanded.Set(map[string]bool{rootNode.Name: true})
	d.scrollY.Set(0)
}

// visibleSelectedPath returns the path of the currently selected node for display.
func (d *directoryTree) visibleSelectedPath() string {
	visible := d.visibleNodes()
	cur := d.cursor.Get()
	if cur >= len(visible) {
		return ""
	}
	return visible[cur].path
}

// loadChildren reads one level of children for the directory at the given logical path.
func (d *directoryTree) loadChildren(nodePath string) {
	parts := strings.Split(nodePath, "/")
	node := &d.tree[0]
	fsPath := d.rootPath
	for _, part := range parts[1:] {
		found := false
		for i := range node.Children {
			if node.Children[i].Name == part {
				node = &node.Children[i]
				fsPath = filepath.Join(fsPath, part)
				found = true
				break
			}
		}
		if !found {
			return
		}
	}
	if len(node.Children) == 0 {
		node.Children = readDir(fsPath)
		if node.Children == nil {
			node.Children = []Node{}
		}
	}
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

func (d *directoryTree) visibleNodes() []visibleNode {
	var result []visibleNode
	expanded := d.expanded.Get()
	for i, node := range d.tree {
		flattenNode(node, 0, node.Name, i == len(d.tree)-1, nil, expanded, &result)
	}
	cur := d.cursor.Get()
	if cur < len(result) {
		sel := result[cur].path
		for i := range result {
			result[i].onPath = result[i].path == sel || strings.HasPrefix(sel, result[i].path+"/")
		}
	}
	return result
}

func (d *directoryTree) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.OnChange(d.cursor, func(int) { d.scrollToCursor() }),
	}
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
	visible := d.visibleNodes()
	d.cursor.Update(func(v int) int {
		if v < len(visible)-1 {
			return v + 1
		}
		return v
	})
}

func (d *directoryTree) toggle() {
	visible := d.visibleNodes()
	cur := d.cursor.Get()
	if cur >= len(visible) {
		return
	}
	vn := visible[cur]
	if !vn.isDir {
		return
	}
	expanding := !d.expanded.Get()[vn.path]
	if expanding {
		d.loadChildren(vn.path)
	}
	d.expanded.Update(func(m map[string]bool) map[string]bool {
		return cloneExpandedWith(m, vn.path, expanding)
	})
}

func (d *directoryTree) collapseOrParent() {
	visible := d.visibleNodes()
	cur := d.cursor.Get()
	if cur >= len(visible) {
		return
	}
	vn := visible[cur]

	// If on an expanded directory, collapse it
	expanded := d.expanded.Get()
	if vn.isDir && expanded[vn.path] {
		d.expanded.Update(func(m map[string]bool) map[string]bool {
			return cloneExpandedWith(m, vn.path, false)
		})
		return
	}

	// At root, navigate to parent directory
	if vn.depth == 0 {
		d.navigateUp()
		return
	}

	// Jump to parent node
	parentPath := path.Dir(vn.path)
	for i := cur - 1; i >= 0; i-- {
		if visible[i].path == parentPath {
			d.cursor.Set(i)
			return
		}
	}
}

templ (d *directoryTree) Render() {
	<div class="flex-col w-full h-full border-rounded border-cyan">
		<div class="flex-col p-1">
			<span class="text-gradient-cyan-magenta font-bold">Directory Tree</span>
			<span class="text-cyan font-dim">{d.visibleSelectedPath()}</span>
		</div>
		<hr class="border-single" />
		<div
			ref={d.scrollContainer}
			class="flex-col grow overflow-y-scroll scrollbar-cyan scrollbar-thumb-bright-cyan"
			scrollOffset={0, d.scrollY.Get()}>
			@for i, vn := range d.visibleNodes() {
				@if i == d.cursor.Get() {
					<span class="bg-bright-black text-cyan font-bold">{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
				} @else @if vn.onPath {
					<span class="text-cyan font-bold">{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
				} @else @if vn.isDir {
					<span class="font-bold">{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
				} @else {
					<span>{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
				}
			}
		</div>
		<hr class="border-single" />
		<div class="flex justify-center p-1">
			<span class="font-dim">j/k: navigate | enter/l: expand | h: collapse | q: quit</span>
		</div>
	</div>
}

// readDir reads one level of a directory and returns its children as Nodes, sorted dirs-first then alphabetically.
// Subdirectory children are not read until the user expands them (lazy loading).
func readDir(dirPath string) []Node {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil
	}

	var children []Node
	for _, entry := range entries {
		if entry.Name()[0] == '.' {
			continue
		}
		node := Node{Name: entry.Name()}
		if entry.IsDir() {
			node.Children = []Node{}
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
			return iDir
		}
		return children[i].Name < children[j].Name
	})
}

// buildRootNode reads the filesystem at absPath and returns a root Node.
func buildRootNode(absPath string) Node {
	root := Node{
		Name:     filepath.Base(absPath),
		Children: readDir(absPath),
	}
	if root.Children == nil {
		root.Children = []Node{}
	}
	return root
}

// cloneExpandedWith returns a copy of m with the given key set to val.
func cloneExpandedWith(m map[string]bool, key string, val bool) map[string]bool {
	out := make(map[string]bool, len(m))
	for k, v := range m {
		out[k] = v
	}
	if val {
		out[key] = true
	} else {
		delete(out, key)
	}
	return out
}

func flattenNode(n Node, depth int, nodePath string, isLast bool, ancestors []bool, expanded map[string]bool, result *[]visibleNode) {
	isDir := n.Children != nil
	*result = append(*result, visibleNode{
		node:      n,
		depth:     depth,
		path:      nodePath,
		isDir:     isDir,
		isLast:    isLast,
		ancestors: ancestors,
	})
	if isDir && expanded[nodePath] {
		newAncestors := make([]bool, len(ancestors)+1)
		copy(newAncestors, ancestors)
		newAncestors[len(ancestors)] = isLast
		for i, child := range n.Children {
			childPath := nodePath + "/" + child.Name
			flattenNode(child, depth+1, childPath, i == len(n.Children)-1, newAncestors, expanded, result)
		}
	}
}

func buildPrefix(vn visibleNode) string {
	if vn.depth == 0 {
		return ""
	}
	var b strings.Builder
	for i := 0; i < vn.depth-1; i++ {
		if vn.ancestors[i+1] {
			b.WriteString("    ")
		} else {
			b.WriteString("│   ")
		}
	}
	if vn.isLast {
		b.WriteString("└── ")
	} else {
		b.WriteString("├── ")
	}
	return b.String()
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
