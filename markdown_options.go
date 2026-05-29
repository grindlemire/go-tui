package tui

// MarkdownOption configures a Markdown component.
type MarkdownOption func(*Markdown)

// WithMarkdownSource sets static markdown content. Ignored when a state source
// is also set (state takes precedence).
func WithMarkdownSource(s string) MarkdownOption {
	return func(m *Markdown) { m.source = s }
}

// WithMarkdownState binds a reactive *State[string] source. When set it takes
// precedence over WithMarkdownSource and the component re-renders on change.
func WithMarkdownState(s *State[string]) MarkdownOption {
	return func(m *Markdown) { m.state = s }
}

// WithMarkdownWidth fixes the render width in characters. 0 (the default) fills
// the available width and wraps to it.
func WithMarkdownWidth(w int) MarkdownOption {
	return func(m *Markdown) { m.width = w }
}

// WithMarkdownTheme overrides the styling theme.
func WithMarkdownTheme(t MarkdownTheme) MarkdownOption {
	return func(m *Markdown) { m.theme = t }
}
