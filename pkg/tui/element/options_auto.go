package element

import (
	"github.com/grindlemire/go-tui/pkg/layout"
)

// WithWidthAuto sets width to auto (size to content).
func WithWidthAuto() Option {
	return func(e *Element) {
		e.style.Width = layout.Auto()
	}
}

// WithHeightAuto sets height to auto (size to content).
func WithHeightAuto() Option {
	return func(e *Element) {
		e.style.Height = layout.Auto()
	}
}
