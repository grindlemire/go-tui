package tui

import (
	"strings"

	"github.com/grindlemire/go-tui/internal/tailwind"
)

// SetOnMouseEnter sets the callback for when the pointer moves onto this
// element or one of its descendants.
func (e *Element) SetOnMouseEnter(fn func()) {
	e.onMouseEnter = fn
	e.noteHover()
}

// SetOnMouseLeave sets the callback for when the pointer moves off this
// element and all of its descendants.
func (e *Element) SetOnMouseLeave(fn func()) {
	e.onMouseLeave = fn
	e.noteHover()
}

// noteHover tells an attached app that hover is in use, so it turns on motion
// reporting after the next render. Detached elements are noted when attached.
func (e *Element) noteHover() {
	if e.app != nil && hasHover(e) {
		e.app.hoverUsed = true
	}
}

// WithHoverClass adds classes applied only while the pointer is over this
// element or a descendant, written without the hover: prefix ("bg-blue
// font-bold"). Only color and text style classes take effect, since a class
// that changed layout could move the element out from under the pointer.
// Generated code emits it for hover: classes in a class attribute.
func WithHoverClass(classes string) Option {
	return func(e *Element) {
		e.addHoverClass(classes)
	}
}

// WithKey gives the element an identity among its siblings, so hover tracking
// follows it when siblings are inserted, removed, or reordered. Generated code
// emits it for key={...} attributes on plain elements. Keys must be comparable.
func WithKey(key any) Option {
	return func(e *Element) {
		e.key, e.hasKey = key, true
	}
}

func (e *Element) addHoverClass(classes string) {
	classes = strings.TrimSpace(classes)
	if classes == "" {
		return
	}
	if e.hoverClass != "" {
		classes = e.hoverClass + " " + classes
	}
	active := e.hoverActive
	e.setHoverStyle(false)
	e.hoverClass = classes
	e.setHoverStyle(active)
	e.noteHover()
}

// setHoverStyle applies or removes the hover: classes. Removal restores only
// properties that still hold what the hover classes wrote, mirroring SetClass.
// Text modifiers are not written to the element; effectiveStyles layers them
// onto the effective (possibly inherited) text style at render time.
func (e *Element) setHoverStyle(on bool) {
	if on == e.hoverActive || (on && e.hoverClass == "") {
		return
	}
	if !on {
		b, a := e.hoverBase, e.hoverApplied
		for _, op := range e.hoverOps {
			classRestore(e, b, a, op)
		}
		e.hoverActive, e.hoverOps, e.hoverText, e.hoverBase, e.hoverApplied = false, nil, nil, nil, nil
		return
	}
	result := tailwind.ParseHover(e.hoverClass)
	e.hoverBase = e.captureClassBase()
	e.hoverOps = result.Ops()
	for _, op := range e.hoverOps {
		e.Apply(classOption(op))
	}
	e.hoverText = result.Text()
	e.hoverApplied = e.captureClassBase()
	e.hoverActive = true
}
