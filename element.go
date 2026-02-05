package tui

// Compile-time check that Element implements the required interfaces.
var (
	_ Renderable = (*Element)(nil)
	_ Focusable  = (*Element)(nil)
	_ Layoutable = (*Element)(nil)
)

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
// It implements Layoutable and owns its children directly.
type Element struct {
	// Tree structure (single source of truth)
	children []*Element
	parent   *Element

	// Layout properties
	style  LayoutStyle
	layout LayoutResult
	dirty  bool

	// Visual properties
	border      BorderStyle
	borderStyle Style
	background  *Style // nil = transparent

	// Text properties
	text         string
	textStyle    Style
	textStyleSet bool // true if textStyle was explicitly configured (false = inherit from parent)
	textAlign    TextAlign

	// Focus properties
	focusable bool
	focused   bool
	onFocus   func(*Element)
	onBlur    func(*Element)
	onEvent   func(*Element, Event) bool

	// Event handlers
	onKeyPress func(*Element, KeyEvent) bool // returns true to consume (stops bubbling)
	onClick    func(*Element)

	// Tree notification
	onChildAdded     func(*Element)
	onFocusableAdded func(Focusable)

	// Custom render hook (used by wrappers that need custom rendering)
	onRender func(*Element, *Buffer)

	// Scroll properties
	scrollMode            ScrollMode
	scrollX               int  // Current horizontal scroll offset
	scrollY               int  // Current vertical scroll offset
	contentWidth          int  // Computed content width (may exceed viewport)
	contentHeight         int  // Computed content height (may exceed viewport)
	scrollToBottomPending bool // Scroll to bottom after next layout

	// Scrollbar styles
	scrollbarStyle      Style
	scrollbarThumbStyle Style

	// HR properties
	hr bool // true if this element is a horizontal rule

	// Gradient properties (nil = no gradient, use solid color)
	textGradient   *Gradient
	bgGradient     *Gradient
	borderGradient *Gradient

	// Pre-render hook for custom update logic (polling, animations, etc.)
	onUpdate func()

	// Watchers attached to this element (timers, channel watchers, etc.)
	watchers []Watcher
}

// New creates a new Element with the given options.
// By default, an Element has Auto width/height (flexes to fill available space).
func New(opts ...Option) *Element {
	e := &Element{
		style: DefaultLayoutStyle(),
		dirty: true,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}
