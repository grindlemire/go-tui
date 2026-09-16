package tui

// WithMarkdownElementOptions applies standard Element options to the rendered
// root element (a flex column), for example class-derived padding or a border.
func WithMarkdownElementOptions(opts ...Option) MarkdownOption {
	return func(m *Markdown) {
		m.elementOpts = append(m.elementOpts, opts...)
	}
}

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

// WithMarkdownWidth fixes the render width in characters. 0 (the default) lets
// the component fill the width assigned by its parent; paragraphs and headings
// wrap to that width, but list and blockquote content renders on a single line
// (it is clipped, not wrapped). Set an explicit width to wrap list and
// blockquote content for documents with long lines.
func WithMarkdownWidth(w int) MarkdownOption {
	return func(m *Markdown) { m.width = w }
}

// WithMarkdownTheme overrides the styling theme.
func WithMarkdownTheme(t MarkdownTheme) MarkdownOption {
	return func(m *Markdown) { m.theme = t }
}
