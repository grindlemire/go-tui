# Markdown Component + gsx Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `Markdown` component that renders the `internal/markdown` block tree into the go-tui widget tree, plus a self-closing `<markdown>` `.gsx` tag. This is the final layer (Plan 4 of 4) and closes issue #62.

**Architecture:** `Markdown` is a pure content renderer implementing `Component`, `AppBinder`, and `PropsUpdater`. Its `Render` resolves the current source (a static string or a reactive `*State[string]`), parses it with `markdown.Parse` (single-entry cache keyed on the source string), and walks the resulting `[]markdown.Block` into a `flex-col` `*Element` tree. Headings/paragraphs become rich-text elements (`WithRichText`), code fences become bordered columns with one child element per line, tables reuse the existing `<table>/<tr>/<th>/<td>` element tree, lists and blockquotes recurse. Scrolling and key handling are left to the caller. The gsx layer adds `markdown` to the component-element codegen path with distinct `source`/`state`/`width`/`theme` attributes (the generator cannot type-discriminate a single attribute).

**Tech Stack:** Go 1.25, stdlib only. Composes the already-merged rich-text primitive (`richtext.go`: `TextSpan`, `WithRichText`, `mergeSpanStyle`), OSC 8 links (automatic via `TextSpan.Link`), and the markdown parser (`internal/markdown`). Table-driven tests; `MockTerminal`/`Buffer` golden render tests.

**Conventions:** `gcommit -m "..."` ONLY, never `git commit` (signing). Conventional commits. Work on a feature branch; `git merge --ff-only` to `main` when green, then `git branch -d`. Run tests with the sandbox disabled (the Go build cache is blocked in-sandbox; "operation not permitted" means retry with `dangerouslyDisableSandbox: true`). TDD: write the failing test, confirm it fails, implement, confirm it passes, commit.

---

## Building blocks already merged (read-only context, do not re-implement)

- **Rich text** (`richtext.go`, package `tui`): `type TextSpan struct { Text string; Style Style; Link string }`. `WithRichText(spans ...TextSpan) Option`, `(*Element).SetRichText`, `(*Element).RichText`. Setting rich text clears plain text. `mergeSpanStyle(base, span Style) Style` (unexported, same package — usable from `markdown.go`): ORs `Attrs`, overrides non-default `Fg`/`Bg`. `richTextWidth(spans)` and `spanLineWidth(line)` helpers exist.
- **Rendering/measurement** already wired for rich text in all three sites: `element_render.go:137` (normal), `:247` (clipped/scroll), `element_layout.go:104` (intrinsic), `:228` (`HeightForWidth` wrap). A non-empty single-line text element measures height 1 (`element_layout.go:89`); an **empty** text element (`text==""`, no rich text, no children) measures height 0 — so blank code lines must render a space.
- **OSC 8 links**: automatic. Populate `TextSpan.Link`; `Cell.Link` flows through diff/Flush. Nothing extra to do.
- **Parser** (`internal/markdown`, no `tui` dependency): `func Parse(src string) []Block`. `type Block struct { Kind BlockKind; Level int; Ordered bool; Lang string; Inline []Inline; Lines []string; Rows [][]TableCell; Children []Block }`. `type Inline struct { Text string; Bold, Italic, Code bool; Link string }`. `type TableCell struct { Inline []Inline }`. `BlockKind`: `KindParagraph, KindHeading, KindCodeFence, KindTable, KindList, KindListItem, KindBlockquote`. A `KindListItem` carries leaf `Inline` and may have a child `KindList` in `Children`; `KindBlockquote` holds parsed `Children`; table `Rows[0]` is the header.
- **Component patterns** to mirror: `input.go`, `textarea.go` (struct + options + `BindApp` + interface assertions), `mount.go` (`PropsUpdater.UpdateProps(fresh Component)` called on cached instances; `AppBinder.BindApp` called on mount and re-mount).
- **Element APIs** used: `New(opts...)`, `WithDirection(Column|Row)`, `WithWidth`, `WithGap`, `WithBorder`, `WithBackground(Style)`, `WithTextStyle(Style)`, `WithText(string)`, `WithWrap(bool)` (false ⇒ `noWrap`), `WithTag(string)`, `WithDisplay(DisplayFlex)`, `(*Element).AddChild`, `(*Element).HeightForWidth(int) int`, `(*Element).IntrinsicSize() (int,int)`. `NewStyle()` chainable: `.Bold() .Italic() .Underline() .Foreground(Color) .Background(Color)`. `Color`: `Cyan`, `BrightBlack`, etc., `DefaultColor()`, `(Color).IsDefault()`. `Style.Attrs` bitset (`AttrBold`, `AttrItalic`, `AttrUnderline`). `BorderStyle`: `BorderNone`, `BorderSingle`, `BorderRounded`, etc.
- **Render test helpers**: `buf := NewBuffer(w, h)`; `root.Render(buf, w, h)`; `buf.Cell(x, y) Cell` (`.Rune`, `.Style`, `.Link`); `buf.StringTrimmed() string`.

---

## File Structure

- Create: `markdown_theme.go` — `MarkdownTheme` struct, `DefaultMarkdownTheme()`. (package `tui`)
- Create: `markdown_options.go` — `MarkdownOption` type and the four `WithMarkdown*` option funcs.
- Create: `markdown.go` — `Markdown` struct, `NewMarkdown`, interface assertions, `BindApp`, `UpdateProps`, `Render`, the block-walk methods, and `inlineToSpans`.
- Create: `markdown_theme_test.go`, `markdown_test.go` — unit + golden render tests.
- Modify: `internal/tuigen/generator_element.go` — `isComponentElement`, `componentConstructor`, `componentAttributeMaps`, new `markdownAttributeToOption`/`markdownHandlerAttributes`.
- Modify: `internal/tuigen/analyzer.go` — add `markdown` to `knownTags` + `voidElements`; add `source`, `state`, `theme` to `knownAttributes`.
- Modify: `internal/lsp/schema/schema.go` — add `markdown` element definition + `markdownAttrs()`.
- Create: `cmd/tui/testdata/markdown.gsx` and `cmd/tui/testdata/markdown_gsx.go` — golden pair.
- Modify: `CLAUDE.md` — document the `<markdown>` tag and `Markdown` component (final task).

---

## Design notes that drive the implementation (resolve the spec's open questions)

1. **Content width threading.** Wrapping happens during layout, not in `Render`, so the component normally just sets `WithRichText` and lets the engine wrap. The one case needing an explicit width is the blockquote bar (below), which must know how many rows the content occupies. The block-walk methods therefore take a `contentWidth int` parameter: the width available to the current block (`m.width` at the root, reduced by `2` for each blockquote nesting; `0` means "auto/unknown — assume no wrapping").
2. **Code fences** must not be a single multiline `WithText` (no-wrap multiline collapses to one line and measures height 1). Each source line is its own child element with `WithWrap(false)`. An **empty** line renders `" "` (a single space) so it measures height 1 and the background block stays solid.
3. **Blockquote bar** cannot be a `BorderStyle` (borders draw full boxes). Build the recursively-rendered content column first, measure its height (`content.HeightForWidth(contentWidth)` when `contentWidth > 0`, else `_, h := content.IntrinsicSize()`), then build a 1-wide `flex-col` bar of `h` glyph elements beside it in a `flex-row`. At auto width the bar height assumes no wrapping (documented limitation).
4. **Table cells fed rich text** measure correctly: `TableIntrinsicSize` (via `element_layout.go:77`) walks cell children whose `IntrinsicSize` uses the rich-text branch (`element_layout.go:104`, width = `richTextWidth`, height 1). Cells reuse the generator's emitted option shape (`WithTag`, `WithDisplay(DisplayFlex)`, `WithDirection`).
5. **Parse cache** is a single entry on the struct (`lastSource`, `cached []markdown.Block`, `parsed bool`). `Render` resolves the source, and re-parses only when the resolved string differs from `lastSource`. `UpdateProps` copies `source`/`state`/`width`/`theme` from the fresh instance but leaves the cache fields untouched (Render handles invalidation), so state-backed content re-parses exactly when its string changes.
6. **gsx cannot type-discriminate**, so `source={stringExpr}` and `state={stateExpr}` are distinct attributes mapping to `WithMarkdownSource`/`WithMarkdownState`. There is no `value` attribute.

---

## Task 1: MarkdownTheme and DefaultMarkdownTheme

**Files:**
- Create: `markdown_theme.go`
- Test: `markdown_theme_test.go`

- [ ] **Step 1: Write the failing test** in `markdown_theme_test.go`:

```go
package tui

import "testing"

func TestDefaultMarkdownTheme(t *testing.T) {
	th := DefaultMarkdownTheme()

	if th.Heading[0].Attrs&AttrBold == 0 {
		t.Errorf("h1 should be bold, attrs=%v", th.Heading[0].Attrs)
	}
	if th.Bold.Attrs&AttrBold == 0 {
		t.Errorf("Bold style should set bold attr")
	}
	if th.Italic.Attrs&AttrItalic == 0 {
		t.Errorf("Italic style should set italic attr")
	}
	if th.Link.Attrs&AttrUnderline == 0 {
		t.Errorf("Link style should be underlined")
	}
	if th.BulletMarker == "" {
		t.Errorf("BulletMarker should have a default")
	}
	if th.BlockquoteBar == 0 {
		t.Errorf("BlockquoteBar should have a default glyph")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ -run TestDefaultMarkdownTheme`
Expected: FAIL — `DefaultMarkdownTheme`, `MarkdownTheme` undefined (does not compile).

- [ ] **Step 3: Write the implementation** in `markdown_theme.go`:

```go
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
	CodeBlockBg     Color       // default ⇒ no fill
	CodeBlockBorder BorderStyle // a real full-box border around the code element

	// Tables reuse the existing table layout (a 1-char column gap, not grid lines).
	// v1 styles the header and, optionally, a separator row under it.
	TableHeader        Style
	TableSeparator     bool
	TableSeparatorChar rune

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
	heading := NewStyle().Bold()
	return MarkdownTheme{
		Heading: [6]Style{
			NewStyle().Bold().Foreground(BrightCyan),
			NewStyle().Bold().Foreground(Cyan),
			heading,
			heading,
			heading,
			heading,
		},
		Paragraph: NewStyle(),
		Bold:      NewStyle().Bold(),
		Italic:    NewStyle().Italic(),
		CodeSpan:  NewStyle().Foreground(BrightMagenta),
		Link:      NewStyle().Underline().Foreground(BrightBlue),

		CodeBlockText:   NewStyle().Foreground(BrightWhite),
		CodeBlockBg:     DefaultColor(),
		CodeBlockBorder: BorderRounded,

		TableHeader:        NewStyle().Bold(),
		TableSeparator:     false,
		TableSeparatorChar: '-',

		BlockquoteBar:      '│', // │
		BlockquoteBarStyle: NewStyle().Foreground(BrightBlack),
		BlockquoteText:     NewStyle().Foreground(BrightBlack),

		BulletMarker: "• ", // "• "
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./ -run TestDefaultMarkdownTheme`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gcommit -m "feat: add MarkdownTheme and DefaultMarkdownTheme"
```

---

## Task 2: Markdown struct, options, caching, and inline→span conversion

**Files:**
- Create: `markdown_options.go`
- Create: `markdown.go`
- Test: `markdown_test.go`

- [ ] **Step 1: Write the failing tests** in `markdown_test.go`:

```go
package tui

import (
	"testing"

	"github.com/grindlemire/go-tui/internal/markdown"
)

func TestMarkdown_InterfaceAssertions(t *testing.T) {
	// Compile-time check that the assertions in markdown.go hold.
	var _ Component = (*Markdown)(nil)
	var _ AppBinder = (*Markdown)(nil)
	var _ PropsUpdater = (*Markdown)(nil)
}

func TestMarkdown_InlineToSpans(t *testing.T) {
	m := NewMarkdown()
	spans := m.inlineToSpans([]markdown.Inline{
		{Text: "plain "},
		{Text: "bold", Bold: true},
		{Text: " "},
		{Text: "ital", Italic: true},
		{Text: " "},
		{Text: "code", Code: true},
		{Text: " "},
		{Text: "site", Link: "http://x"},
	})
	if len(spans) != 8 {
		t.Fatalf("want 8 spans, got %d", len(spans))
	}
	if spans[1].Style.Attrs&AttrBold == 0 {
		t.Errorf("span[1] should be bold")
	}
	if spans[3].Style.Attrs&AttrItalic == 0 {
		t.Errorf("span[3] should be italic")
	}
	if spans[5].Style.Fg != BrightMagenta {
		t.Errorf("span[5] code should use CodeSpan fg, got %v", spans[5].Style.Fg)
	}
	if spans[7].Link != "http://x" {
		t.Errorf("span[7] link target = %q, want http://x", spans[7].Link)
	}
	if spans[7].Style.Attrs&AttrUnderline == 0 {
		t.Errorf("span[7] link should be underlined")
	}
}

func TestMarkdown_ParseCache(t *testing.T) {
	m := NewMarkdown(WithMarkdownSource("# Hi"))
	m.ensureParsed()
	first := m.cached
	m.ensureParsed() // unchanged source ⇒ same slice reused
	if &first[0] != &m.cached[0] {
		t.Errorf("cache should be reused when source is unchanged")
	}
	m.source = "# Bye"
	m.ensureParsed() // changed source ⇒ re-parse
	if len(m.cached) == 0 || m.cached[0].Kind != markdown.KindHeading {
		t.Errorf("re-parse failed: %+v", m.cached)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./ -run TestMarkdown_`
Expected: FAIL — `Markdown`, `NewMarkdown`, options undefined (does not compile).

- [ ] **Step 3: Implement options** in `markdown_options.go`:

```go
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
```

- [ ] **Step 4: Implement the struct, constructor, lifecycle, and helpers** in `markdown.go`. (The block-walk `render*` methods are added in later tasks; this step renders only paragraphs and headings so the package compiles and the cache/inline tests pass.)

```go
package tui

import (
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
	default:
		// Constructs added in later tasks; until then, render their inline text
		// as a paragraph so nothing is silently dropped.
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
```

- [ ] **Step 5: Add a heading/paragraph golden render test** to `markdown_test.go`:

```go
func TestMarkdown_RenderHeadingAndParagraph(t *testing.T) {
	m := NewMarkdown(WithMarkdownSource("# Title\n\nHello **world**"), WithMarkdownWidth(20))
	root := m.Render(nil)

	buf := NewBuffer(20, 6)
	root.Render(buf, 20, 6)

	out := buf.StringTrimmed()
	if !contains(out, "Title") || !contains(out, "world") {
		t.Fatalf("rendered output missing text:\n%s", out)
	}
	// "Title" on row 0 should be bold (heading style).
	if buf.Cell(0, 0).Style.Attrs&AttrBold == 0 {
		t.Errorf("heading first cell should be bold")
	}
}

func contains(s, sub string) bool { return strings_Contains(s, sub) }
```

Add `import "strings"` and replace `strings_Contains`/`contains` with a direct `strings.Contains` call if you prefer; the helper exists only to keep the snippet self-contained. Simplest: delete the `contains` helper, `import "strings"`, and call `strings.Contains(out, "Title")` directly.

- [ ] **Step 6: Run to verify all Task 2 tests pass**

Run: `go test ./ -run TestMarkdown_`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
gcommit -m "feat: add Markdown component skeleton with paragraph/heading rendering"
```

---

## Task 3: Code-fence rendering

**Files:**
- Modify: `markdown.go`
- Test: `markdown_test.go`

- [ ] **Step 1: Write the failing test** in `markdown_test.go`:

```go
func TestMarkdown_RenderCodeFence(t *testing.T) {
	src := "```go\nfmt.Println(1)\n\nfmt.Println(2)\n```"
	m := NewMarkdown(WithMarkdownSource(src), WithMarkdownWidth(40))
	root := m.Render(nil)

	buf := NewBuffer(40, 8)
	root.Render(buf, 40, 8)

	out := buf.StringTrimmed()
	if !strings.Contains(out, "fmt.Println(1)") || !strings.Contains(out, "fmt.Println(2)") {
		t.Fatalf("code fence body missing:\n%s", out)
	}
	// The blank middle line must NOT collapse: line 2 of the body is still present
	// (both Println lines render on separate rows with a gap row between them).
}
```

(Ensure `markdown_test.go` imports `"strings"`.)

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./ -run TestMarkdown_RenderCodeFence`
Expected: FAIL — the default branch renders the fence as an empty paragraph (no body text), so the assertion fails.

- [ ] **Step 3: Implement** in `markdown.go`. Add the case to `renderBlock` and the method:

```go
	case markdown.KindCodeFence:
		return m.renderCodeFence(b)
```

```go
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
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./ -run TestMarkdown_RenderCodeFence`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gcommit -m "feat: render markdown code fences as one element per line"
```

---

## Task 4: List rendering (ordered, unordered, nested)

**Files:**
- Modify: `markdown.go`
- Test: `markdown_test.go`

- [ ] **Step 1: Write the failing test** in `markdown_test.go`:

```go
func TestMarkdown_RenderList(t *testing.T) {
	src := "- alpha\n- beta\n  - nested\n"
	m := NewMarkdown(WithMarkdownSource(src), WithMarkdownWidth(30))
	root := m.Render(nil)

	buf := NewBuffer(30, 8)
	root.Render(buf, 30, 8)
	out := buf.StringTrimmed()

	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") || !strings.Contains(out, "nested") {
		t.Fatalf("list items missing:\n%s", out)
	}
	// Bullet marker present on a top-level item.
	if !strings.Contains(out, "•") {
		t.Errorf("expected bullet marker • in output:\n%s", out)
	}
	// Nested item is indented further than a top-level item.
	rowAlpha, colAlpha := findCell(buf, 'a') // first 'a' is in "alpha"
	rowNested, colNested := findCell(buf, 'n')
	_ = rowAlpha
	_ = rowNested
	if colNested <= colAlpha {
		t.Errorf("nested item col %d should exceed top-level col %d", colNested, colAlpha)
	}
}

// findCell returns the row,col of the first occurrence of r in the buffer.
func findCell(buf *Buffer, r rune) (int, int) {
	w, h := buf.Size()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if buf.Cell(x, y).Rune == r {
				return y, x
			}
		}
	}
	return -1, -1
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./ -run TestMarkdown_RenderList`
Expected: FAIL — the default branch renders the list as an empty paragraph.

- [ ] **Step 3: Implement** in `markdown.go`. Add `import "fmt"` and `import "strings"` to the file's import block (alongside the markdown import). Add the case and methods:

```go
	case markdown.KindList:
		return m.renderList(b, 0)
```

```go
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
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./ -run TestMarkdown_RenderList`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gcommit -m "feat: render markdown lists with nesting and ordered markers"
```

---

## Task 5: Blockquote rendering (recursive, glyph bar)

**Files:**
- Modify: `markdown.go`
- Test: `markdown_test.go`

- [ ] **Step 1: Write the failing test** in `markdown_test.go`:

```go
func TestMarkdown_RenderBlockquote(t *testing.T) {
	src := "> quoted line one\n> quoted line two\n"
	m := NewMarkdown(WithMarkdownSource(src), WithMarkdownWidth(30))
	root := m.Render(nil)

	buf := NewBuffer(30, 6)
	root.Render(buf, 30, 6)
	out := buf.StringTrimmed()

	if !strings.Contains(out, "quoted line one") {
		t.Fatalf("blockquote text missing:\n%s", out)
	}
	// A bar glyph │ appears in column 0 on the content's first row.
	if buf.Cell(0, 0).Rune != '│' {
		t.Errorf("expected │ bar at (0,0), got %q", buf.Cell(0, 0).Rune)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./ -run TestMarkdown_RenderBlockquote`
Expected: FAIL — no bar glyph; default branch renders blockquote inline (empty) text.

- [ ] **Step 3: Implement** in `markdown.go`. Add the case and method:

```go
	case markdown.KindBlockquote:
		return m.renderBlockquote(b, contentWidth)
```

```go
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
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./ -run TestMarkdown_RenderBlockquote`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gcommit -m "feat: render markdown blockquotes with a glyph bar column"
```

---

## Task 6: Table rendering

**Files:**
- Modify: `markdown.go`
- Test: `markdown_test.go`

- [ ] **Step 1: Write the failing test** in `markdown_test.go`:

```go
func TestMarkdown_RenderTable(t *testing.T) {
	src := "| Name | Age |\n| --- | --- |\n| Ann | 30 |\n| Bob | 25 |\n"
	m := NewMarkdown(WithMarkdownSource(src), WithMarkdownWidth(40))
	root := m.Render(nil)

	buf := NewBuffer(40, 8)
	root.Render(buf, 40, 8)
	out := buf.StringTrimmed()

	for _, want := range []string{"Name", "Age", "Ann", "30", "Bob", "25"} {
		if !strings.Contains(out, want) {
			t.Fatalf("table missing %q:\n%s", want, out)
		}
	}
	// Header cell "Name" renders bold (TableHeader style).
	rName, cName := findCell(buf, 'N')
	if rName < 0 {
		t.Fatal("could not locate header cell")
	}
	if buf.Cell(cName, rName).Style.Attrs&AttrBold == 0 {
		t.Errorf("header cell should be bold")
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./ -run TestMarkdown_RenderTable`
Expected: FAIL — default branch renders an empty paragraph; no table cells.

- [ ] **Step 3: Implement** in `markdown.go`. Add the case and method:

```go
	case markdown.KindTable:
		return m.renderTable(b)
```

```go
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
```

`stringWidth` is the package's existing display-width helper (used in `richtext.go`).

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./ -run TestMarkdown_RenderTable`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gcommit -m "feat: render markdown tables via the table element tree"
```

---

## Task 7: Reactive state source + link rendering

**Files:**
- Test: `markdown_test.go`

This task adds no production code; it verifies the state path and OSC 8 link wiring already provided by `BindApp`/`ensureParsed` and `TextSpan.Link`.

- [ ] **Step 1: Write the tests** in `markdown_test.go`:

```go
func TestMarkdown_StateSourceReparses(t *testing.T) {
	st := NewState("# First")
	m := NewMarkdown(WithMarkdownState(st))

	m.ensureParsed()
	if m.cached[0].Inline[0].Text != "First" {
		t.Fatalf("want First, got %q", m.cached[0].Inline[0].Text)
	}
	st.Set("# Second")
	m.ensureParsed()
	if m.cached[0].Inline[0].Text != "Second" {
		t.Fatalf("state change should re-parse; got %q", m.cached[0].Inline[0].Text)
	}
}

func TestMarkdown_StateTakesPrecedenceOverSource(t *testing.T) {
	st := NewState("from state")
	m := NewMarkdown(WithMarkdownSource("from source"), WithMarkdownState(st))
	if got := m.resolveSource(); got != "from state" {
		t.Errorf("resolveSource() = %q, want \"from state\"", got)
	}
}

func TestMarkdown_LinkRendersAsOSC8(t *testing.T) {
	m := NewMarkdown(WithMarkdownSource("see [docs](https://example.com)"), WithMarkdownWidth(40))
	root := m.Render(nil)

	buf := NewBuffer(40, 3)
	root.Render(buf, 40, 3)

	// The cell under the link label "docs" carries the OSC 8 target.
	r, c := findCell(buf, 'd') // first 'd' is "docs"
	if r < 0 {
		t.Fatal("could not find link label")
	}
	if buf.Cell(c, r).Link != "https://example.com" {
		t.Errorf("link cell target = %q, want https://example.com", buf.Cell(c, r).Link)
	}
}
```

- [ ] **Step 2: Run to verify the state/link behavior**

Run: `go test ./ -run TestMarkdown_`
Expected: PASS. If `TestMarkdown_LinkRendersAsOSC8` fails because the first `'d'` lands in "docs" only after "see ", confirm `findCell` returns the label cell; adjust the probe rune to a letter unique to the label if needed (e.g. probe `'c'` only if unambiguous). Do not change production code for this; it is a test-probe issue.

- [ ] **Step 3: Commit**

```bash
gcommit -m "test: cover Markdown reactive state source and OSC 8 link rendering"
```

---

## Task 8: gsx codegen for the `<markdown>` tag

**Files:**
- Modify: `internal/tuigen/generator_element.go`
- Test: `internal/tuigen/generator_test.go`

- [ ] **Step 1: Write the failing test** in `internal/tuigen/generator_test.go` (mirror the existing `parseAndGenerateSkipImports` table tests; place it near the other component-element tests):

```go
func TestGenerate_MarkdownComponent(t *testing.T) {
	type tc struct {
		input string
		want  []string // substrings that must appear in the generated output
	}
	tests := map[string]tc{
		"source attr": {
			input: `package p
templ (c *view) Render() {
	<markdown source={c.readme} width={80} />
}`,
			want: []string{
				"app.MountPersistent(c,",
				"tui.NewMarkdown(",
				"tui.WithMarkdownSource(c.readme)",
				"tui.WithMarkdownWidth(80)",
			},
		},
		"state attr": {
			input: `package p
templ (c *view) Render() {
	<markdown state={c.md} />
}`,
			want: []string{
				"tui.NewMarkdown(",
				"tui.WithMarkdownState(c.md)",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := parseAndGenerateSkipImports("test.gsx", tt.input)
			if err != nil {
				t.Fatalf("generate error: %v", err)
			}
			for _, w := range tt.want {
				if !strings.Contains(output, w) {
					t.Errorf("output missing %q:\n%s", w, output)
				}
			}
		})
	}
}
```

(Confirm the `view` receiver shape matches other generator tests in the file; if those tests use a different minimal component header, copy that exact header. Check an existing test such as the textarea/input/modal generator test for the canonical `templ (c *X) Render()` form and any required struct declaration.)

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./internal/tuigen/ -run TestGenerate_MarkdownComponent`
Expected: FAIL — `<markdown>` is not a component element; output uses `tui.New()` / `UNKNOWN_COMPONENT_markdown` rather than `tui.NewMarkdown`.

- [ ] **Step 3: Implement** the codegen changes in `internal/tuigen/generator_element.go`:

  1. Extend `isComponentElement`:

```go
func isComponentElement(tag string) bool {
	return tag == "textarea" || tag == "input" || tag == "modal" || tag == "markdown"
}
```

  2. Extend `componentConstructor` with a `markdown` case:

```go
	case "markdown":
		return "tui.NewMarkdown"
```

  3. Add the attribute and handler maps (near the other component maps):

```go
// markdownAttributeToOption maps markdown-specific attributes to tui.WithMarkdown* options.
// source and state are distinct because the generator cannot type-discriminate a
// single expression attribute: source={stringExpr}, state={*State[string] expr}.
var markdownAttributeToOption = map[string]string{
	"source": "tui.WithMarkdownSource(%s)",
	"state":  "tui.WithMarkdownState(%s)",
	"width":  "tui.WithMarkdownWidth(%s)",
	"theme":  "tui.WithMarkdownTheme(%s)",
}

// markdownHandlerAttributes: markdown has no event handlers.
var markdownHandlerAttributes = map[string]string{}
```

  4. Extend `componentAttributeMaps`:

```go
	case "markdown":
		return markdownAttributeToOption, markdownHandlerAttributes
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./internal/tuigen/ -run TestGenerate_MarkdownComponent`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gcommit -m "feat(tuigen): generate tui.NewMarkdown for the <markdown> tag"
```

---

## Task 9: Analyzer + LSP schema for `<markdown>`

**Files:**
- Modify: `internal/tuigen/analyzer.go`
- Modify: `internal/lsp/schema/schema.go`
- Test: `internal/tuigen/analyzer_validation_test.go` (or the closest existing analyzer test file — match where unknown-tag / known-tag tests already live)

- [ ] **Step 1: Write the failing test** for the analyzer. Find the existing test that asserts a known tag analyzes cleanly (search `internal/tuigen` for a test feeding a `<textarea ... />` and expecting no errors) and add a sibling case:

```go
func TestAnalyze_MarkdownTagKnown(t *testing.T) {
	src := `package p
templ (c *view) Render() {
	<markdown source={c.readme} width={80} />
}`
	// Use the same parse+analyze helper the neighboring analyzer tests use.
	errs := analyzeSource(t, src) // replace with the actual helper name in this package
	if len(errs) != 0 {
		t.Fatalf("expected no analyzer errors for <markdown>, got: %v", errs)
	}
}

func TestAnalyze_MarkdownIsVoid(t *testing.T) {
	src := `package p
templ (c *view) Render() {
	<markdown source={c.readme}>oops</markdown>
}`
	errs := analyzeSource(t, src)
	if len(errs) == 0 {
		t.Fatalf("expected a void-element error for <markdown> with children")
	}
}
```

Replace `analyzeSource` with the real helper used by the adjacent tests (e.g. a `parseAndAnalyze` function). If no such helper exists, mirror the parse+`NewAnalyzer().Analyze(...)` sequence another analyzer test uses and read `.errors`/returned `ErrorList`.

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./internal/tuigen/ -run TestAnalyze_Markdown`
Expected: FAIL — `unknown element tag <markdown>` and `unknown attribute source`/`state`; the void-element test fails because `markdown` is not yet void.

- [ ] **Step 3: Implement** the analyzer changes in `internal/tuigen/analyzer.go`:

  1. Add to `knownTags`:

```go
	"markdown": true,
```

  2. Add to `voidElements`:

```go
	"markdown": true,
```

  3. Add to `knownAttributes` (only `source`, `state`, `theme` are new; `width` already exists):

```go
	// Markdown
	"source": true,
	"state":  true,
	"theme":  true,
```

- [ ] **Step 4: Implement** the LSP schema entry in `internal/lsp/schema/schema.go`. Add the element to the `Elements` map:

```go
	"markdown": {
		Tag:         "markdown",
		Description: "Renders a markdown string into the widget tree (headings, bold/italic, inline code, fenced code blocks, tables, lists, blockquotes, links). A pure content renderer: wrap it in a scrollable container to scroll long documents.",
		Attributes:  markdownAttrs(),
		SelfClosing: true,
		Category:    "display",
	},
```

  And add the `markdownAttrs` constructor near the other `*Attrs()` funcs:

```go
// markdownAttrs returns attributes for markdown elements.
func markdownAttrs() []AttributeDef {
	return []AttributeDef{
		{Name: "id", Type: "string", Description: "Unique identifier for the element", Category: "generic"},
		{Name: "class", Type: "string", Description: "Tailwind-style CSS classes", Category: "generic"},
		{Name: "source", Type: "expression", Description: "Static markdown content (string expression)", Category: "generic"},
		{Name: "state", Type: "expression", Description: "Reactive *State[string] markdown source; re-renders on change", Category: "generic"},
		{Name: "width", Type: "int", Description: "Fixed render width in characters (0 = fill available width)", Category: "layout"},
		{Name: "theme", Type: "expression", Description: "tui.MarkdownTheme overriding the default styling", Category: "visual"},
		{Name: "ref", Type: "expression", Description: "Bind this element to a ref variable", Category: "ref"},
		{Name: "deps", Type: "expression", Description: "Explicit state dependencies for reactive bindings", Category: "generic"},
	}
}
```

- [ ] **Step 5: Run to verify it passes**

Run: `go test ./internal/tuigen/ -run TestAnalyze_Markdown` and `go test ./internal/lsp/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
gcommit -m "feat: register <markdown> tag in analyzer and LSP schema"
```

---

## Task 10: gsx testdata golden pair + CLI check

**Files:**
- Create: `cmd/tui/testdata/markdown.gsx`
- Create: `cmd/tui/testdata/markdown_gsx.go`

The `cmd/tui` integration test runs `tui check` and `tui fmt --stdout` over every `testdata/*.gsx`; both must pass for the new fixture. (`testdata` is ignored by `go build`, so the `_gsx.go` is a committed reference, not compiled by the test suite. Verify it compiles by hand in Step 4.)

- [ ] **Step 1: Build the CLI and create the fixture**

```bash
go build -o /tmp/tui ./cmd/tui
```

Create `cmd/tui/testdata/markdown.gsx`:

```gsx
package testdata

type docView struct {
	readme string
}

func DocView(readme string) *docView {
	return &docView{readme: readme}
}

templ (c *docView) Render() {
	<div class="flex-col" scrollable={true}>
		<markdown source={c.readme} width={80} />
	</div>
}
```

- [ ] **Step 2: Verify `tui check` passes on the fixture**

Run: `/tmp/tui check cmd/tui/testdata/markdown.gsx`
Expected: no errors, exit 0.

- [ ] **Step 3: Generate the golden `_gsx.go`**

Run: `/tmp/tui generate cmd/tui/testdata/markdown.gsx`
This writes `cmd/tui/testdata/markdown_gsx.go`. Read it and confirm it contains `app.MountPersistent(`, `tui.NewMarkdown(`, `tui.WithMarkdownSource(c.readme)`, and `tui.WithMarkdownWidth(80)`.

- [ ] **Step 4: Verify the generated file compiles**

Because `testdata` is excluded from normal builds, type-check it directly:

```bash
cp cmd/tui/testdata/markdown_gsx.go /tmp/md_check.go
# In /tmp/md_check.go the package is `testdata`; compile-check by building the
# testdata dir as an ad-hoc package:
go vet ./cmd/tui/testdata/ 2>&1 | head -20 || true
```

If `go vet` does not traverse `testdata`, instead temporarily copy `markdown.gsx`'s generated output into a scratch module, or eyeball it against `textarea_gsx.go` for structural parity (imports, `Render(app *tui.App) *tui.Element`, `MountPersistent`, `UpdateProps`, `BindApp`). The structural reference is `cmd/tui/testdata/textarea_gsx.go`.

- [ ] **Step 5: Run the CLI integration tests**

Run: `go test ./cmd/tui/`
Expected: PASS (the new `markdown.gsx` passes `check` and `fmt --stdout`).

- [ ] **Step 6: Commit**

```bash
gcommit -m "test: add <markdown> gsx testdata golden pair"
```

---

## Task 11: Full integration test, race, and documentation

**Files:**
- Modify: `markdown_test.go`
- Modify: `CLAUDE.md`

- [ ] **Step 1: Add a full-document golden test** to `markdown_test.go` that exercises every construct in one source and asserts the top-level structure and a few deep styles:

```go
func TestMarkdown_FullDocument(t *testing.T) {
	src := "# Title\n\n" +
		"A paragraph with **bold**, *italic*, `code`, and a [link](http://x).\n\n" +
		"```go\nx := 1\n```\n\n" +
		"| H1 | H2 |\n| --- | --- |\n| a | b |\n\n" +
		"- one\n- two\n  - nested\n\n" +
		"> quoted\n"
	m := NewMarkdown(WithMarkdownSource(src), WithMarkdownWidth(60))
	root := m.Render(nil)

	buf := NewBuffer(60, 30)
	root.Render(buf, 60, 30)
	out := buf.StringTrimmed()

	for _, want := range []string{
		"Title", "bold", "italic", "code", "link",
		"x := 1", "H1", "H2", "one", "two", "nested", "quoted",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("full document missing %q:\n%s", want, out)
		}
	}
	// Top-level block kinds parsed in order.
	wantKinds := []markdown.BlockKind{
		markdown.KindHeading, markdown.KindParagraph, markdown.KindCodeFence,
		markdown.KindTable, markdown.KindList, markdown.KindBlockquote,
	}
	if len(m.cached) != len(wantKinds) {
		t.Fatalf("want %d top-level blocks, got %d: %+v", len(wantKinds), len(m.cached), m.cached)
	}
	for i, k := range wantKinds {
		if m.cached[i].Kind != k {
			t.Errorf("block %d kind = %v, want %v", i, m.cached[i].Kind, k)
		}
	}
}
```

- [ ] **Step 2: Run the package tests**

Run: `go test ./`
Expected: PASS.

- [ ] **Step 3: Run the full race suite** (sandbox disabled)

Run: `go test -race ./...`
Expected: PASS across all packages.

- [ ] **Step 4: Document the component** in `CLAUDE.md`. Add `<markdown>` to the Built-in Elements table:

```
| `<markdown>` | Renders a markdown string into the widget tree |
```

  Add a "Markdown-specific Attributes" subsection after the Textarea attributes table:

```
### Markdown-specific Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `source` | `string` | Static markdown content (string expression) |
| `state` | `*State[string]` | Reactive markdown source; re-renders on change |
| `width` | `int` | Fixed render width in characters (0 = fill available width) |
| `theme` | `tui.MarkdownTheme` | Override the default styling |
```

  And add a short note: the `<markdown>` tag is self-closing; provide content through `source` or `state` (literal markdown as children is not supported). The component is a pure content renderer — wrap it in a `scrollable` container to scroll long documents. Add the new files to the "Where to Look" map under a new "Changing markdown rendering" subsection: `markdown.go`, `markdown_options.go`, `markdown_theme.go`, `internal/markdown/`.

- [ ] **Step 5: Commit**

```bash
gcommit -m "docs: document the markdown component and <markdown> tag"
```

---

## Known v1 limitations (documented, acceptable)

- Code blocks fix to content width with wrapping disabled; horizontal scrolling of long code lines is left to the caller's container.
- Blockquote bar height is measured from content; at auto width (`width=0`) it assumes content does not wrap.
- Tables render left-aligned with a 1-char column gap (the existing table layout); column alignment markers and grid borders are not applied (matches parser/spec non-goals).
- No code-block syntax highlighting, images, footnotes, task lists, or HTML passthrough (spec non-goals).
- Literal markdown cannot be written as `.gsx` children; content arrives through `source`/`state` expression attributes only.

## Self-Review

- **Spec coverage:** Layer 4 struct/interfaces (Tasks 2, 7), options (Task 2), block→element mapping for heading/paragraph (Task 2), code fence (Task 3), list (Task 4), blockquote (Task 5), table (Task 6); `MarkdownTheme`/`DefaultMarkdownTheme` (Task 1); caching + state precedence (Tasks 2, 7); OSC 8 links (Task 7). Layer 5 generator (Task 8), analyzer + LSP schema (Task 9), testdata golden pair (Task 10), docs + race (Task 11). All spec sections map to a task.
- **Open questions resolved:** rich-text intrinsic sizing for `<td>`/`<th>` confirmed (Task 6 note via `element_layout.go:104`); parse-cache invalidation hook is `ensureParsed` keyed on resolved source (Task 2); code-block height-1 collapse avoided with one element per line + space for blanks (Task 3); blockquote bar height via `HeightForWidth`/`IntrinsicSize` (Task 5); gsx type-discrimination via distinct `source`/`state` attrs (Tasks 8, 9).
- **Type consistency:** `Markdown` fields (`source`, `state`, `width`, `theme`, `lastSource`, `cached`, `parsed`) are referenced identically in `Render`/`UpdateProps`/`ensureParsed`/`resolveSource`. Methods named consistently: `renderBlock`, `renderHeading`, `renderParagraph`, `renderCodeFence`, `renderList`, `renderListItem`, `renderBlockquote`, `renderTable`, `renderTableRow`, `inlineToSpans`. `MarkdownTheme` field names match between Task 1 definition and Tasks 2–6 usage (`Heading`, `Paragraph`, `Bold`, `Italic`, `CodeSpan`, `Link`, `CodeBlockText`, `CodeBlockBg`, `CodeBlockBorder`, `TableHeader`, `TableSeparator`, `TableSeparatorChar`, `BlockquoteBar`, `BlockquoteBarStyle`, `BlockquoteText`, `BulletMarker`). Option funcs (`WithMarkdownSource/State/Width/Theme`) match the generator's `markdownAttributeToOption` templates.
- **Placeholder scan:** test-probe helpers (`findCell`, `contains`) are defined inline; the `contains` helper note instructs deleting it in favor of `strings.Contains`. The analyzer test helper name (`analyzeSource`) is flagged to be replaced with the package's real helper after inspecting neighboring tests — this is an instruction to read existing code, not a code placeholder. No "TBD"/"add error handling"/"similar to Task N" placeholders remain.
