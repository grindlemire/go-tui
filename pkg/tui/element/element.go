// Package element provides a high-level API for building TUI layouts.
// Elements combine layout properties (from the layout package) with visual
// properties (borders, backgrounds) and can be composed into trees that
// are rendered to a buffer.
package element

import (
	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
)

// Element is a layout container with visual properties.
// It implements layout.Layoutable and owns its children directly.
type Element struct {
	// Tree structure (single source of truth)
	children []*Element
	parent   *Element

	// Layout properties
	style  layout.Style
	layout layout.Layout
	dirty  bool

	// Visual properties
	border      tui.BorderStyle
	borderStyle tui.Style
	background  *tui.Style // nil = transparent
}

// Compile-time check that Element implements Layoutable
var _ layout.Layoutable = (*Element)(nil)

// New creates a new Element with the given options.
// By default, an Element has Auto width/height (flexes to fill available space).
func New(opts ...Option) *Element {
	e := &Element{
		style: layout.DefaultStyle(),
		dirty: true,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// --- Implement layout.Layoutable interface ---

// LayoutStyle returns the layout style properties for this element.
func (e *Element) LayoutStyle() layout.Style {
	return e.style
}

// LayoutChildren returns the children to be laid out.
func (e *Element) LayoutChildren() []layout.Layoutable {
	result := make([]layout.Layoutable, len(e.children))
	for i, child := range e.children {
		result[i] = child
	}
	return result
}

// SetLayout is called by the layout engine to store computed layout.
func (e *Element) SetLayout(l layout.Layout) {
	e.layout = l
}

// GetLayout returns the last computed layout.
func (e *Element) GetLayout() layout.Layout {
	return e.layout
}

// IsDirty returns whether this element needs layout recalculation.
func (e *Element) IsDirty() bool {
	return e.dirty
}

// SetDirty marks this element as needing recalculation or not.
func (e *Element) SetDirty(dirty bool) {
	e.dirty = dirty
}

// --- Element's own API ---

// AddChild appends children to this Element.
func (e *Element) AddChild(children ...*Element) {
	for _, child := range children {
		child.parent = e
		e.children = append(e.children, child)
	}
	e.MarkDirty()
}

// RemoveChild removes a child from this Element.
// Returns true if the child was found and removed.
func (e *Element) RemoveChild(child *Element) bool {
	for i, c := range e.children {
		if c == child {
			// Remove by swapping with last element and truncating
			e.children[i] = e.children[len(e.children)-1]
			e.children = e.children[:len(e.children)-1]
			child.parent = nil
			e.MarkDirty()
			return true
		}
	}
	return false
}

// Children returns the child elements.
func (e *Element) Children() []*Element {
	return e.children
}

// Parent returns the parent element, or nil if this is the root.
func (e *Element) Parent() *Element {
	return e.parent
}

// Calculate computes layout for this Element and all descendants.
func (e *Element) Calculate(availableWidth, availableHeight int) {
	layout.Calculate(e, availableWidth, availableHeight)
}

// Rect returns the computed border box.
func (e *Element) Rect() layout.Rect {
	return e.layout.Rect
}

// ContentRect returns the computed content area.
func (e *Element) ContentRect() layout.Rect {
	return e.layout.ContentRect
}

// MarkDirty marks this Element and ancestors as needing recalculation.
func (e *Element) MarkDirty() {
	for elem := e; elem != nil && !elem.dirty; elem = elem.parent {
		elem.dirty = true
	}
}

// SetStyle updates the layout style and marks the element dirty.
func (e *Element) SetStyle(style layout.Style) {
	e.style = style
	e.MarkDirty()
}

// Style returns the current layout style.
func (e *Element) Style() layout.Style {
	return e.style
}

// Border returns the border style.
func (e *Element) Border() tui.BorderStyle {
	return e.border
}

// SetBorder sets the border style.
func (e *Element) SetBorder(border tui.BorderStyle) {
	e.border = border
}

// BorderStyle returns the style used to render the border.
func (e *Element) BorderStyle() tui.Style {
	return e.borderStyle
}

// SetBorderStyle sets the style used to render the border.
func (e *Element) SetBorderStyle(style tui.Style) {
	e.borderStyle = style
}

// Background returns the background style, or nil if transparent.
func (e *Element) Background() *tui.Style {
	return e.background
}

// SetBackground sets the background style. Pass nil for transparent.
func (e *Element) SetBackground(style *tui.Style) {
	e.background = style
}
