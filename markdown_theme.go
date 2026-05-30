package tui

// MarkdownTheme controls how a Markdown component styles each construct. It is a
// flat struct of Style fields plus a few non-style extras. Construct a sensible
// default with DefaultMarkdownTheme and override fields as needed.
type MarkdownTheme struct {
	// Heading holds per-level heading styles, indexed by (level-1) for levels 1..6.
	Heading   [6]Style
	Paragraph Style
	// Bold, Italic, CodeSpan, and Link are layered over the surrounding text via
	// the inline scanner's flags. Their Attrs OR in and non-default colors replace.
	Bold     Style
	Italic   Style
	CodeSpan Style // inline `code`
	Link     Style

	CodeBlockText   Style
	CodeBlockBg     Color       // default => no fill
	CodeBlockBorder BorderStyle // a real full-box border around the code element

	// CodeHighlighter colorizes fenced code blocks. nil disables highlighting
	// (code renders in CodeBlockText). DefaultMarkdownTheme sets the built-in.
	CodeHighlighter CodeHighlighter

	// Tables render as a full grid (outer box, column separators, and a rule under
	// the header) in the TableBorder style. BorderNone falls back to a plain grid
	// using rounded characters. TableHeader styles the header cells.
	TableHeader Style
	TableBorder BorderStyle

	// Blockquotes render a 1-wide glyph column (borders draw full boxes, so a
	// BorderStyle cannot be used for a left bar).
	BlockquoteBar      rune
	BlockquoteBarStyle Style
	BlockquoteText     Style

	BulletMarker string // unordered-list marker, e.g. "• "
}

// DefaultMarkdownTheme returns a glow-inspired theme that reads well on dark and
// light terminals using only attributes and a couple of muted colors.
func DefaultMarkdownTheme() MarkdownTheme {
	return MarkdownTheme{
		Heading: [6]Style{
			NewStyle().Bold().Underline().Italic(), // h1: bold + underline + italic
			NewStyle().Bold().Italic(),             // h2: bold + italic
			NewStyle().Italic(),                    // h3: italic
			NewStyle().Bold(),                      // h4
			NewStyle().Bold(),                      // h5
			NewStyle().Bold(),                      // h6
		},
		Paragraph: NewStyle(),
		Bold:      NewStyle().Bold(),
		Italic:    NewStyle().Italic(),
		CodeSpan:  NewStyle().Foreground(BrightMagenta),
		Link:      NewStyle().Underline().Foreground(BrightBlue),

		CodeBlockText:   NewStyle().Foreground(BrightWhite),
		CodeBlockBg:     DefaultColor(),
		CodeBlockBorder: BorderRounded,
		CodeHighlighter: NewHighlighter(DefaultPalette()),

		TableHeader: NewStyle().Bold(),
		TableBorder: BorderRounded,

		BlockquoteBar:      '│',
		BlockquoteBarStyle: NewStyle().Foreground(BrightBlack),
		BlockquoteText:     NewStyle().Italic().Foreground(BrightBlack),

		BulletMarker: "• ",
	}
}
