# Directory Tree

## What We're Building

A keyboard-driven file browser that reads a real directory from disk and displays it as a foldable tree with Unicode box-drawing characters. You can navigate with vim-style keys, expand and collapse directories, and walk upward past the root to re-root the tree at a parent directory.

Concepts used:

- **State** ([Guide 05](state)): reactive `State[T]` for cursor position, expanded directories, and scroll offset
- **Components** ([Guide 06](components)): a struct component with constructor and render method
- **Events** ([Guide 07](events)): `KeyMap` for navigation bindings
- **Refs** ([Guide 08](refs-and-clicks)): `*tui.Ref` for the scrollable container
- **Scrolling** ([Guide 10](scrolling)): state-driven scroll with `scrollOffset` and `ViewportSize`
- **Watchers** ([Guide 09](watchers)): `OnChange` to auto-scroll when the cursor moves
- **Styling** ([Guide 03](styling)): conditional classes for cursor highlight, ancestor path, and directory names

Most of this project is plain Go code that manages state variables on the component struct. go-tui handles rendering and UI updates, so the code you write is just the tree logic.

## Project Setup

Create a new directory and initialize the module:

```bash
mkdir directory-tree && cd directory-tree
go mod init directory-tree
go get github.com/grindlemire/go-tui
```

You'll create two files:

- `tree.gsx` - the component and all its logic
- `main.go` - the entry point

## Data Model

The tree is built from two types. `Node` is the raw filesystem data: a name plus optional children. `visibleNode` is a flattened row ready for rendering carrying the depth, logical path, and enough information to draw the correct tree prefix characters.

```go
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
```

The `Children` field doubles as a directory indicator: `nil` means it's a file, while a non-nil slice (even an empty one) means it's a directory. This distinction matters because an empty directory should still show the expand arrow.

The `ancestors` slice on `visibleNode` tracks whether each ancestor was the last child at its depth. `buildPrefix` uses this to decide between `│` (more siblings below) and blank space (last child, no continuation line). The `onPath` field is set during flattening to mark nodes that are ancestors of the currently selected node, used for teal path highlighting in the render template.

## Reading the Filesystem

`readDir` reads a single directory level, skipping hidden files and sorting results with directories first. Subdirectories get an empty `Children` slice to mark them as directories, but their contents aren't read yet. Since only the top level is read at startup, opening a huge directory like `/usr` is fast:

```go
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
```

When `os.ReadDir` fails (permissions, broken symlinks), the function returns `nil` and the parent directory silently omits that subtree.

Children are loaded on demand when the user first expands a directory. `loadChildren` walks the tree to find the node by its logical path, then calls `readDir` to populate its children:

```go
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
```

The `len(node.Children) == 0` check means children are only read once per directory. After collapsing and re-expanding, the previously loaded children are still in the tree.

The sort puts directories before files, then alphabetical within each group:

```go
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
```

## Component State

The component holds five pieces of state: the filesystem root path, the node tree, a cursor position, a set of expanded directory paths, and a scroll offset. A ref tracks the scrollable container so the component can query its viewport height.

```go
type directoryTree struct {
    rootPath        string
    tree            []Node
    cursor          *tui.State[int]
    expanded        *tui.State[map[string]bool]
    scrollY         *tui.State[int]
    scrollContainer *tui.Ref
}
```

The `expanded` state is a `map[string]bool` keyed by logical path (like `myproject/internal/config`). When the user expands a directory, its path gets added to the map. When they collapse it, the path gets removed.

The constructor reads the filesystem and starts with the root directory expanded:

```go
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
```

## Flattening the Tree

The render method needs a flat list of all rows from expanded directories. `visibleNodes` walks the tree and only recurses into directories that appear in the `expanded` map. After flattening, it marks which nodes are on the path from root to the cursor, so the render template can apply teal highlighting without recomputing it per row:

```go
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
```

`flattenNode` is recursive. It appends the current node and, if the node is an expanded directory, recurses into its children. Each level builds up the `ancestors` slice so the prefix renderer knows which columns need continuation lines:

```go
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
```

## Drawing Tree Lines

The `buildPrefix` function turns the depth and ancestor information into Unicode box-drawing characters. For each ancestor level, it draws either `│` (more siblings follow) or a blank space (last child). The final segment is `├──` or `└──` depending on whether the node itself is the last sibling:

```go
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
```

Directories get an expand/collapse indicator prepended to their name:

```go
func nodeLabel(vn visibleNode, expanded map[string]bool) string {
    if vn.isDir {
        if expanded[vn.path] {
            return "▼ " + vn.node.Name
        }
        return "▶ " + vn.node.Name
    }
    return vn.node.Name
}
```

## Keyboard Navigation

The component implements `KeyListener` with vim-style bindings. `j`/`k` and arrow keys move the cursor. `Enter`/`l`/Right expand or collapse directories. `h`/Left collapses the current directory or jumps to its parent:

```go
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
```

`moveUp` and `moveDown` use `State.Update` to atomically adjust the cursor within bounds:

```go
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
```

Note that neither `moveUp` nor `moveDown` handles scrolling. Instead they update the cursor and there is a watcher on the cursor to ensure the scroll state shows that cursor on the screen (more on that below).

### Toggling Directories

`toggle` flips a directory's expanded state. When expanding, it first calls `loadChildren` to read the directory contents from disk if they haven't been loaded yet. Because `State[T]` holds a map, you can't mutate it in place. The `cloneExpandedWith` helper copies the map and applies the change:

```go
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
```

### Collapsing and Walking Up

`collapseOrParent` has three behaviors depending on context. On an expanded directory, it collapses it. On a file or collapsed directory, it jumps the cursor to the parent. On the root node, it re-roots the entire tree at the parent directory:

```go
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
```

`navigateUp` re-reads the filesystem at the parent path and resets all state. The root stops moving when `filepath.Dir` returns the same path (the filesystem root):

```go
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
```

## Scrolling with OnChange

The tree can be taller than the terminal. The scrollable container clips content and draws a scrollbar automatically. The remaining problem is keeping the cursor visible as it moves.

Rather than calling a scroll function after every cursor movement, the component uses an `OnChange` watcher. Whenever the cursor state changes, the watcher fires and adjusts `scrollY` to keep the cursor row inside the viewport:

```go
func (d *directoryTree) Watchers() []tui.Watcher {
    return []tui.Watcher{
        tui.OnChange(d.cursor, func(int) { d.scrollToCursor() }),
    }
}

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
```

This approach keeps scroll logic out of the key handlers entirely. Any code that moves the cursor (keyboard navigation, collapsing, re-rooting) gets automatic scroll adjustment for free. This is equivalent to React's effects API.

## Conditional Row Styling

Each row in the tree gets one of four visual treatments: teal text on a dark background for the cursor row, teal bold for ancestor nodes on the path to the selection, bold for other directories, and plain for files.

The `onPath` field was already computed during flattening in `visibleNodes()`, so the render template just checks the boolean. This avoids recomputing path ancestry per row, which would otherwise turn the render loop into O(N²) work.

The render template uses chained `if`/`else if` to apply the right class to each row:

```gsx
for i, vn := range d.visibleNodes() {
    if i == d.cursor.Get() {
        <span class="bg-bright-black text-cyan font-bold">{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
    } else if vn.onPath {
        <span class="text-cyan font-bold">{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
    } else if vn.isDir {
        <span class="font-bold">{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
    } else {
        <span>{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
    }
}
```

The selected row gets the same teal color as the ancestor path, plus a dark background to mark the cursor. This keeps the path visually continuous from root to selection. The ancestor highlighting helps you orient in deep trees. As you navigate, the teal path updates from the root down to your current position.

## Complete Example

Here's the full `tree.gsx`:

```gsx
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
            for i, vn := range d.visibleNodes() {
                if i == d.cursor.Get() {
                    <span class="bg-bright-black text-cyan font-bold">{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
                } else if vn.onPath {
                    <span class="text-cyan font-bold">{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
                } else if vn.isDir {
                    <span class="font-bold">{buildPrefix(vn) + nodeLabel(vn, d.expanded.Get())}</span>
                } else {
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
```

With `main.go`:

```go
package main

import (
    "fmt"
    "os"
    "path/filepath"

    tui "github.com/grindlemire/go-tui"
)

func main() {
    root := "."
    if len(os.Args) > 1 {
        root = os.Args[1]
    }

    absRoot, err := filepath.Abs(root)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    app, err := tui.NewApp(
        tui.WithRootComponent(DirectoryTree(absRoot)),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    defer app.Close()

    if err := app.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

Generate and run:

```bash
tui generate ./...
go run .
```

Pass a path to browse a specific directory:

```bash
go run . /usr/local
```

The directory tree browser should look like this:

![Directory Tree screenshot](/guides/21.png)

## Next Steps

- [Scrolling](scrolling) -- Scrollable containers, keyboard and mouse scroll control
- [Timers, Watchers, and Channels](watchers) -- Background operations and state change watchers
- [Building a Dashboard](dashboard) -- A larger example combining multiple go-tui features
