package tui

import (
	"fmt"
	"slices"
)

// Re-renders rebuild every element, so hover cannot remember element pointers.
// Instead it remembers, for each element it entered, the element's path in the
// tree and the onMouseLeave it had at that moment. Paths use an element's key
// when it has one and its sibling index otherwise. On each mouse event the
// chain under the pointer is compared by path: a re-render that keeps the
// structure keeps the paths, so nothing fires and a handler that sets state
// cannot loop. Anything that drops out (pointer moved, modal opened, list
// changed) gets its saved onMouseLeave, so every enter gets exactly one leave.
//
// Content can also change under a still pointer (wheel scroll, a key press, a
// background update). After a render that followed any dispatched event, hover
// is re-checked once at the last pointer position. Renders caused only by
// hover callbacks do not re-check, which bounds a handler that moves its own
// element to one extra round instead of a loop.

// hoverEntry is an element the pointer entered.
type hoverEntry struct {
	path  []pathSeg
	leave func()
}

// pathSeg is one step from a parent to a child: the child's key, or its
// index when it has none.
type pathSeg struct {
	key   string
	keyed bool
	index int
}

// dispatchMouseHover updates hover for a mouse event at (x, y) in buffer
// coordinates. inside is false when the pointer left the app's region
// (inline mode), which leaves everything currently hovered.
func (a *App) dispatchMouseHover(x, y int, inside bool) {
	if !a.mouseMotion {
		return
	}
	a.pointerX, a.pointerY, a.pointerInside = x, y, inside
	a.updateHover()
}

// recheckHover re-runs hover at the last pointer position after a render, if
// an event was dispatched since the last check.
func (a *App) recheckHover() {
	if !a.hoverStale {
		return
	}
	a.hoverStale = false
	if a.mouseMotion && a.pointerInside {
		a.updateHover()
	}
}

// updateHover fires onMouseLeave/onMouseEnter for the pointer at its last
// known position against the current tree, then syncs hover: styles.
func (a *App) updateHover() {
	var chain []*Element
	if a.pointerInside {
		chain = a.hoverChain(a.pointerX, a.pointerY)
	}
	paths := make([][]pathSeg, len(chain))
	for i, el := range chain {
		paths[i] = elementPath(el)
	}
	hovered := func(path []pathSeg) bool {
		return slices.ContainsFunc(paths, func(p []pathSeg) bool { return slices.Equal(p, path) })
	}

	// Leave deepest first, enter outermost first (matches DOM mouseleave/mouseenter).
	var kept []hoverEntry
	var left []func()
	for _, h := range a.hovered {
		if hovered(h.path) {
			kept = append(kept, h)
		} else if h.leave != nil {
			left = append(left, h.leave)
		}
	}
	var entered []func()
	var next []hoverEntry
	for i, el := range chain {
		if !hasHover(el) || slices.ContainsFunc(kept, func(h hoverEntry) bool { return slices.Equal(h.path, paths[i]) }) {
			continue
		}
		if el.onMouseEnter != nil {
			entered = append(entered, el.onMouseEnter)
		}
		next = append(next, hoverEntry{path: paths[i], leave: el.onMouseLeave})
	}
	a.hovered = append(kept, next...)
	slices.SortStableFunc(a.hovered, func(x, y hoverEntry) int { return len(y.path) - len(x.path) })

	a.syncHoverStyles(true)
	for _, fn := range left {
		fn()
	}
	for _, fn := range slices.Backward(entered) {
		fn()
	}
}

// syncHoverStyles makes the elements at hovered paths in the current tree,
// and only those, show their hover: classes. Called after hover changes
// (markDirty) and after a re-render builds fresh elements, before layout
// (no markDirty, or every frame would schedule another).
func (a *App) syncHoverStyles(markDirty bool) {
	var styled []*Element
	for _, h := range a.hovered {
		if el := a.elementAtPath(h.path); el != nil && el.hoverClass != "" {
			styled = append(styled, el)
		}
	}
	for _, el := range a.hoverStyled {
		if el.hoverActive && !slices.Contains(styled, el) {
			el.setHoverStyle(false)
			if markDirty {
				el.MarkDirty()
			}
		}
	}
	for _, el := range styled {
		if !el.hoverActive {
			el.setHoverStyle(true)
			if markDirty {
				el.MarkDirty()
			}
		}
	}
	a.hoverStyled = styled
}

func hasHover(el *Element) bool {
	return el.onMouseEnter != nil || el.onMouseLeave != nil || el.hoverClass != ""
}

// hoverChain returns the element under (x, y) and its ancestors, deepest
// first. While an overlay is open only the topmost overlay is hit-tested, and
// the chain stops at the overlay so content behind the backdrop never hovers.
func (a *App) hoverChain(x, y int) []*Element {
	scope := a.hitScope()
	if scope == nil {
		return nil
	}
	var chain []*Element
	for el := scope.ElementAt(x, y); el != nil; el = el.parent {
		chain = append(chain, el)
		if el == scope {
			break
		}
	}
	return chain
}

// hitScope is the subtree mouse hit testing searches: the topmost open
// overlay while one is open, so content behind its backdrop never reacts,
// and the root otherwise.
func (a *App) hitScope() *Element {
	if n := len(a.overlays); n > 0 {
		return a.overlays[n-1].element
	}
	return a.root
}

// elementPath is the element's path from the tree root, which survives a
// re-render that keeps the same structure (and, for keyed elements, a reorder).
func elementPath(el *Element) []pathSeg {
	var path []pathSeg
	for ; el.parent != nil; el = el.parent {
		path = append(path, segmentOf(el))
	}
	slices.Reverse(path)
	return path
}

func segmentOf(el *Element) pathSeg {
	if el.hasKey {
		return pathSeg{key: fmt.Sprint(el.key), keyed: true}
	}
	return pathSeg{index: slices.Index(el.parent.children, el)}
}

// elementAtPath resolves a path in the current tree, or nil if nothing is
// there any more.
func (a *App) elementAtPath(path []pathSeg) *Element {
	el := a.root
	for _, seg := range path {
		if el == nil {
			return nil
		}
		var next *Element
		if seg.keyed {
			for _, c := range el.children {
				if c.hasKey && fmt.Sprint(c.key) == seg.key {
					next = c
					break
				}
			}
		} else if seg.index < len(el.children) && !el.children[seg.index].hasKey {
			next = el.children[seg.index]
		}
		el = next
	}
	return el
}

// moveCoalescer holds back MouseMove events while a batch of events drains,
// so a burst of motion dispatches only its last position.
type moveCoalescer struct {
	ev   MouseEvent
	held bool
}

// dispatchCoalescing dispatches ev, holding back MouseMove events until the
// next other event (dispatched after the held move, preserving order) or flush.
func (a *App) dispatchCoalescing(ev Event, c *moveCoalescer) {
	if me, ok := ev.(MouseEvent); ok && me.Action == MouseMove {
		c.ev, c.held = me, true
		return
	}
	a.flushMove(c)
	a.Dispatch(ev)
}

func (a *App) flushMove(c *moveCoalescer) {
	if c.held {
		c.held = false
		a.Dispatch(c.ev)
	}
}
