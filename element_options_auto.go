package tui


// WithWidthAuto sets width to auto (size to content).
func WithWidthAuto() Option {
	return func(e *Element) {
		e.style.Width = Auto()
	}
}

// WithHeightAuto sets height to auto (size to content).
func WithHeightAuto() Option {
	return func(e *Element) {
		e.style.Height = Auto()
	}
}
