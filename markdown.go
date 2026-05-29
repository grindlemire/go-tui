package tui

import (
	"fmt"
	"strings"

	"github.com/grindlemire/go-tui/internal/markdown"
)

// Markdown renders a markdown string into the widget tree. It is a pure content
// renderer: it owns no scroll state or key bindings. Wrap it in a scrollable
// container to scroll long documents. Construct with NewMarkdown.
type Markdown struct {
	source string
	state  *State[string] // optional reactive source; takes precedence over source
	width  int            // 0 = fill available width
	theme  MarkdownTheme

	// single-entry parse cache keyed on the resolved source string
	lastSource string
	cached     []markdown.Block
	parsed     bool
}

var (
	_ Component    = (*Markdown)(nil)
	_ AppBinder    = (*Markdown)(nil)
	_ PropsUpdater = (*Markdown)(nil)
)

// NewMarkdown creates a Markdown component.
func NewMarkdown(opts ...MarkdownOption) *Markdown {
	m := &Markdown{
		theme: DefaultMarkdownTheme(),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// BindApp binds the reactive source (if any) to the app. It is a no-op when the
// component has only a static source.
func (m *Markdown) BindApp(app *App) {
	if m.state != nil {
		m.state.BindApp(app)
	}
}

// UpdateProps copies new props from a freshly-constructed instance when this
// cached instance is re-rendered. The parse cache is intentionally preserved;
// Render re-parses when the resolved source string changes.
func (m *Markdown) UpdateProps(fresh Component) {
	f, ok := fresh.(*Markdown)
	if !ok {
		return
	}
	m.source = f.source
	m.state = f.state
	m.width = f.width
	m.theme = f.theme
}

// resolveSource returns the current markdown text (state wins when present).
func (m *Markdown) resolveSource() string {
	if m.state != nil {
		return m.state.Get()
	}
	return m.source
}

// ensureParsed (re)parses when the resolved source changed since last parse.
func (m *Markdown) ensureParsed() {
	src := m.resolveSource()
	if m.parsed && src == m.lastSource {
		return
	}
	m.cached = markdown.Parse(src)
	m.lastSource = src
	m.parsed = true
}

// Render parses the current source and walks the block tree into a flex-col root.
func (m *Markdown) Render(app *App) *Element {
	m.ensureParsed()

	opts := []Option{WithDirection(Column)}
	if m.width > 0 {
		opts = append(opts, WithWidth(m.width))
	}
	root := New(opts...)

	for _, b := range m.cached {
		if el := m.renderBlock(b, m.width); el != nil {
			root.AddChild(el)
		}
	}
	return root
}

// renderBlock dispatches one block to its renderer. contentWidth is the width
// available to this block (0 = auto/unknown).
func (m *Markdown) renderBlock(b markdown.Block, contentWidth int) *Element {
	switch b.Kind {
	case markdown.KindHeading:
		return m.renderHeading(b)
	case markdown.KindParagraph:
		return m.renderParagraph(b)
	case markdown.KindCodeFence:
		return m.renderCodeFence(b)
	case markdown.KindList:
		return m.renderList(b, 0)
	case markdown.KindBlockquote:
		return m.renderBlockquote(b, contentWidth)
	case markdown.KindTable:
		return m.renderTable(b)
	default:
		// Unknown leaf: render its inline text as a paragraph so nothing is
		// silently dropped.
		return m.renderParagraph(b)
	}
}

func (m *Markdown) renderHeading(b markdown.Block) *Element {
	level := b.Level
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}
	return New(
		WithTextStyle(m.theme.Heading[level-1]),
		WithRichText(m.inlineToSpans(b.Inline)...),
	)
}

func (m *Markdown) renderParagraph(b markdown.Block) *Element {
	return New(
		WithTextStyle(m.theme.Paragraph),
		WithRichText(m.inlineToSpans(b.Inline)...),
	)
}

// renderCodeFence renders a fenced code block as a bordered/background column
// with one child element per source line. A single multiline WithText would
// collapse to one line under no-wrap and measure as height 1, so each line is
// its own element; blank lines render a space to keep height 1.
func (m *Markdown) renderCodeFence(b markdown.Block) *Element {
	opts := []Option{WithDirection(Column)}
	if m.theme.CodeBlockBorder != BorderNone {
		opts = append(opts, WithBorder(m.theme.CodeBlockBorder))
	}
	if !m.theme.CodeBlockBg.IsDefault() {
		opts = append(opts, WithBackground(NewStyle().Background(m.theme.CodeBlockBg)))
	}
	box := New(opts...)

	for _, line := range b.Lines {
		text := line
		if text == "" {
			text = " " // keep blank lines from collapsing to height 0
		}
		box.AddChild(New(
			WithText(text),
			WithWrap(false),
			WithTextStyle(m.theme.CodeBlockText),
		))
	}
	return box
}

// renderList renders a list and its items at the given nesting depth.
func (m *Markdown) renderList(list markdown.Block, depth int) *Element {
	col := New(WithDirection(Column))
	for i, item := range list.Children {
		marker := m.theme.BulletMarker
		if list.Ordered {
			marker = fmt.Sprintf("%d. ", i+1)
		}
		col.AddChild(m.renderListItem(item, marker, depth))
	}
	return col
}

// renderListItem renders one item: an indented "marker + inline text" row,
// followed by any nested list rendered at depth+1.
func (m *Markdown) renderListItem(item markdown.Block, marker string, depth int) *Element {
	itemCol := New(WithDirection(Column))

	indent := strings.Repeat("  ", depth)
	row := New(WithDirection(Row))
	row.AddChild(New(WithText(indent+marker), WithWrap(false)))
	row.AddChild(New(WithRichText(m.inlineToSpans(item.Inline)...)))
	itemCol.AddChild(row)

	for _, child := range item.Children {
		if child.Kind == markdown.KindList {
			itemCol.AddChild(m.renderList(child, depth+1))
		}
	}
	return itemCol
}

// renderBlockquote renders a recursive blockquote: a 1-wide glyph bar column
// beside the indented, recursively-rendered content. The bar's height matches
// the content height (measured at the available width; at auto width it assumes
// no wrapping).
func (m *Markdown) renderBlockquote(b markdown.Block, contentWidth int) *Element {
	childWidth := 0
	if contentWidth > 0 {
		childWidth = contentWidth - 2 // bar (1) + gap (1)
		if childWidth < 1 {
			childWidth = 1
		}
	}

	content := New(WithDirection(Column))
	for _, child := range b.Children {
		if el := m.renderBlock(child, childWidth); el != nil {
			content.AddChild(el)
		}
	}

	// Measure content height to size the bar.
	height := 0
	if childWidth > 0 {
		height = content.HeightForWidth(childWidth)
	} else {
		_, height = content.IntrinsicSize()
	}
	if height < 1 {
		height = 1
	}

	bar := New(WithDirection(Column), WithWidth(1))
	for i := 0; i < height; i++ {
		bar.AddChild(New(
			WithText(string(m.theme.BlockquoteBar)),
			WithTextStyle(m.theme.BlockquoteBarStyle),
		))
	}

	row := New(WithDirection(Row), WithGap(1))
	row.AddChild(bar)
	row.AddChild(content)
	return row
}

// renderTable renders a pipe table into the existing <table>/<tr>/<th>/<td>
// element tree. Row 0 is the header. An optional separator row is drawn when the
// theme requests it.
func (m *Markdown) renderTable(b markdown.Block) *Element {
	table := New(WithTag("table"), WithDisplay(DisplayFlex), WithDirection(Column))
	if len(b.Rows) == 0 {
		return table
	}

	// Header row.
	header := b.Rows[0]
	table.AddChild(m.renderTableRow(header, true))

	// Optional separator row sized to each header cell's text width.
	if m.theme.TableSeparator {
		sep := New(WithTag("tr"), WithDisplay(DisplayFlex), WithDirection(Row))
		for _, cell := range header {
			w := 0
			for _, in := range cell.Inline {
				w += stringWidth(in.Text)
			}
			if w < 1 {
				w = 1
			}
			sep.AddChild(New(
				WithTag("td"),
				WithText(strings.Repeat(string(m.theme.TableSeparatorChar), w)),
			))
		}
		table.AddChild(sep)
	}

	// Body rows.
	for _, row := range b.Rows[1:] {
		table.AddChild(m.renderTableRow(row, false))
	}
	return table
}

func (m *Markdown) renderTableRow(cells []markdown.TableCell, header bool) *Element {
	tr := New(WithTag("tr"), WithDisplay(DisplayFlex), WithDirection(Row))
	tag := "td"
	if header {
		tag = "th"
	}
	for _, cell := range cells {
		opts := []Option{WithTag(tag), WithRichText(m.inlineToSpans(cell.Inline)...)}
		if header {
			opts = append(opts, WithTextStyle(m.theme.TableHeader))
		}
		tr.AddChild(New(opts...))
	}
	return tr
}

// inlineToSpans converts parser inline runs into themed TextSpans. The element's
// textStyle supplies the base; each span layers only its inline-specific style.
func (m *Markdown) inlineToSpans(inls []markdown.Inline) []TextSpan {
	spans := make([]TextSpan, 0, len(inls))
	for _, in := range inls {
		st := Style{}
		if in.Bold {
			st = mergeSpanStyle(st, m.theme.Bold)
		}
		if in.Italic {
			st = mergeSpanStyle(st, m.theme.Italic)
		}
		if in.Code {
			st = mergeSpanStyle(st, m.theme.CodeSpan)
		}
		if in.Link != "" {
			st = mergeSpanStyle(st, m.theme.Link)
		}
		spans = append(spans, TextSpan{Text: in.Text, Style: st, Link: in.Link})
	}
	return spans
}
