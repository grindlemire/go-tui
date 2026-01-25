// Package element provides a high-level API for building TUI layouts.
// Elements combine layout properties (from the layout package) with visual
// properties (borders, backgrounds) and can be composed into trees that
// are rendered to a buffer.
package element

import (
	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
)

var _ tui.Renderable = (*Element)(nil)
var _ tui.Focusable = (*Element)(nil)

// TextAlign specifies how text is aligned within its content area.
type TextAlign int

const (
	// TextAlignLeft aligns text to the left edge (default).
	TextAlignLeft TextAlign = iota
	// TextAlignCenter centers text horizontally.
	TextAlignCenter
	// TextAlignRight aligns text to the right edge.
	TextAlignRight
)

// ScrollMode specifies how an element scrolls its content.
type ScrollMode int

const (
	// ScrollNone disables scrolling (default).
	ScrollNone ScrollMode = iota
	// ScrollVertical enables vertical scrolling.
	ScrollVertical
	// ScrollHorizontal enables horizontal scrolling.
	ScrollHorizontal
	// ScrollBoth enables both vertical and horizontal scrolling.
	ScrollBoth
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

	// Text properties
	text      string
	textStyle tui.Style
	textAlign TextAlign

	// Focus properties
	focusable bool
	focused   bool
	onFocus   func()
	onBlur    func()
	onEvent   func(tui.Event) bool

	// Tree notification
	onChildAdded     func(*Element)
	onFocusableAdded func(tui.Focusable)

	// Custom render hook (used by wrappers that need custom rendering)
	onRender func(*Element, *tui.Buffer)

	// Scroll properties
	scrollMode    ScrollMode
	scrollX       int // Current horizontal scroll offset
	scrollY       int // Current vertical scroll offset
	contentWidth  int // Computed content width (may exceed viewport)
	contentHeight int // Computed content height (may exceed viewport)

	// Scrollbar styles
	scrollbarStyle      tui.Style
	scrollbarThumbStyle tui.Style
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

// IntrinsicSize returns the natural content-based dimensions of this element.
// For text elements, returns the text width and height (1 line).
// For containers, returns the computed intrinsic size based on children.
func (e *Element) IntrinsicSize() (width, height int) {
	// Text content has explicit intrinsic size
	if e.text != "" {
		textWidth := stringWidth(e.text)
		textHeight := 1
		// Add padding to get the element's intrinsic size
		width = textWidth + e.style.Padding.Horizontal()
		height = textHeight + e.style.Padding.Vertical()
		// Add border if present (borders take 1 cell on each side)
		if e.border != tui.BorderNone {
			width += 2
			height += 2
		}
		return width, height
	}

	// For containers without text, compute from children
	if len(e.children) == 0 {
		// Empty container has no intrinsic size
		return 0, 0
	}

	// Compute intrinsic size from children
	isRow := e.style.Direction == layout.Row
	var intrinsicW, intrinsicH int

	for i, child := range e.children {
		childW, childH := child.IntrinsicSize()
		childStyle := child.LayoutStyle()
		marginH := childStyle.Margin.Horizontal()
		marginV := childStyle.Margin.Vertical()

		if isRow {
			intrinsicW += childW + marginH
			if childH+marginV > intrinsicH {
				intrinsicH = childH + marginV
			}
		} else {
			if childW+marginH > intrinsicW {
				intrinsicW = childW + marginH
			}
			intrinsicH += childH + marginV
		}

		// Add gap between children (not before first)
		if i > 0 {
			if isRow {
				intrinsicW += e.style.Gap
			} else {
				intrinsicH += e.style.Gap
			}
		}
	}

	// Add padding
	intrinsicW += e.style.Padding.Horizontal()
	intrinsicH += e.style.Padding.Vertical()

	// Add border if present
	if e.border != tui.BorderNone {
		intrinsicW += 2
		intrinsicH += 2
	}

	return intrinsicW, intrinsicH
}

// --- Element's own API ---

// AddChild appends children to this Element.
// Notifies root's onChildAdded callback for each child.
func (e *Element) AddChild(children ...*Element) {
	for _, child := range children {
		child.parent = e
		e.children = append(e.children, child)
		e.notifyChildAdded(child)
	}
	e.MarkDirty()
}

// notifyChildAdded walks up to root and calls appropriate callbacks.
func (e *Element) notifyChildAdded(child *Element) {
	root := e
	for root.parent != nil {
		root = root.parent
	}
	if root.onChildAdded != nil {
		root.onChildAdded(child)
	}
	// Notify App about focusable elements for auto-registration
	if root.onFocusableAdded != nil && child.IsFocusable() {
		root.onFocusableAdded(child)
	}
}

// SetOnChildAdded sets the callback for when any descendant is added.
func (e *Element) SetOnChildAdded(fn func(*Element)) {
	e.onChildAdded = fn
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

// --- Text API ---

// Text returns the text content.
func (e *Element) Text() string {
	return e.text
}

// SetText updates the text content and recalculates intrinsic width.
func (e *Element) SetText(content string) {
	e.text = content
	e.style.Width = layout.Fixed(stringWidth(content))
	e.MarkDirty()
}

// TextStyle returns the style used to render the text.
func (e *Element) TextStyle() tui.Style {
	return e.textStyle
}

// SetTextStyle sets the style used to render the text.
func (e *Element) SetTextStyle(style tui.Style) {
	e.textStyle = style
}

// TextAlign returns the text alignment.
func (e *Element) TextAlign() TextAlign {
	return e.textAlign
}

// SetTextAlign sets the text alignment.
func (e *Element) SetTextAlign(align TextAlign) {
	e.textAlign = align
}

// stringWidth returns the display width of a string in terminal cells.
func stringWidth(s string) int {
	width := 0
	for _, r := range s {
		width += tui.RuneWidth(r)
	}
	return width
}

// --- Focus API ---

// IsFocusable returns whether this element can receive focus.
func (e *Element) IsFocusable() bool {
	return e.focusable
}

// IsFocused returns whether this element currently has focus.
func (e *Element) IsFocused() bool {
	return e.focused
}

// Focus marks this element and all children as focused.
// Calls onFocus callback if set, then cascades to children.
func (e *Element) Focus() {
	e.focused = true
	if e.onFocus != nil {
		e.onFocus()
	}
	for _, child := range e.children {
		child.Focus()
	}
}

// Blur marks this element and all children as not focused.
// Calls onBlur callback if set, then cascades to children.
func (e *Element) Blur() {
	e.focused = false
	if e.onBlur != nil {
		e.onBlur()
	}
	for _, child := range e.children {
		child.Blur()
	}
}

// HandleEvent dispatches an event to this element's handler.
// Returns true if the event was consumed.
func (e *Element) HandleEvent(event tui.Event) bool {
	// First, let user handler try to consume the event
	if e.onEvent != nil {
		if e.onEvent(event) {
			return true
		}
	}

	// Handle scroll events for scrollable elements
	if e.scrollMode != ScrollNone {
		if e.handleScrollEvent(event) {
			return true
		}
	}

	return false
}

// handleScrollEvent handles keyboard events for scrolling.
func (e *Element) handleScrollEvent(event tui.Event) bool {
	key, ok := event.(tui.KeyEvent)
	if !ok {
		return false
	}

	_, viewportHeight := e.ViewportSize()

	switch key.Key {
	case tui.KeyUp:
		if e.scrollMode == ScrollVertical || e.scrollMode == ScrollBoth {
			e.ScrollBy(0, -1)
			return true
		}
	case tui.KeyDown:
		if e.scrollMode == ScrollVertical || e.scrollMode == ScrollBoth {
			e.ScrollBy(0, 1)
			return true
		}
	case tui.KeyLeft:
		if e.scrollMode == ScrollHorizontal || e.scrollMode == ScrollBoth {
			e.ScrollBy(-1, 0)
			return true
		}
	case tui.KeyRight:
		if e.scrollMode == ScrollHorizontal || e.scrollMode == ScrollBoth {
			e.ScrollBy(1, 0)
			return true
		}
	case tui.KeyPageUp:
		if e.scrollMode == ScrollVertical || e.scrollMode == ScrollBoth {
			e.ScrollBy(0, -viewportHeight)
			return true
		}
	case tui.KeyPageDown:
		if e.scrollMode == ScrollVertical || e.scrollMode == ScrollBoth {
			e.ScrollBy(0, viewportHeight)
			return true
		}
	case tui.KeyHome:
		e.ScrollTo(0, 0)
		return true
	case tui.KeyEnd:
		e.ScrollToBottom()
		return true
	}

	return false
}

// --- Focus Tree Discovery API ---

// SetOnFocusableAdded sets a callback called when a focusable descendant is added.
// This is used by App to auto-register focusable elements.
func (e *Element) SetOnFocusableAdded(fn func(tui.Focusable)) {
	e.onFocusableAdded = fn
}

// WalkFocusables calls fn for each focusable element in the tree.
// This is used by App to discover existing focusable elements.
func (e *Element) WalkFocusables(fn func(tui.Focusable)) {
	if e.IsFocusable() {
		fn(e)
	}
	for _, child := range e.children {
		child.WalkFocusables(fn)
	}
}
