# Markdown Component Design

- Date: 2026-05-29
- Status: Approved (design); pending implementation plan
- Issue: [#62 markdown support](https://github.com/grindlemire/go-tui/issues/62)

## Summary

Add a `Markdown` widget that renders a markdown string into the go-tui widget
tree, in the spirit of `glow`. The widget is a pure content renderer: it returns
a block tree and leaves scrolling to the caller. It is usable from Go directly
(`tui.NewMarkdown(...)`) and from `.gsx` via a self-closing `<markdown>` tag.

Rendering markdown correctly requires inline mixed styling (a bold word inside an
otherwise-plain sentence that still wraps at word boundaries). The current
`Element` holds a single `text` string with a single `textStyle`, so this design
introduces a reusable rich-text primitive (`TextSpan` + `WithRichText`) as a
first-class public framework feature. Markdown is its first consumer.

## Goals

- Render headings, bold, italic, inline code, fenced code blocks, pipe tables,
  ordered and unordered lists (including nesting), blockquotes, and links.
- Provide a public, documented, tested rich-text primitive that other widgets
  can reuse (inline code, styled log views, future syntax highlighting).
- Keep the project's zero-external-dependency policy intact.
- Support both static `string` content and reactive `*State[string]` content.

## Non-goals (v1)

- Code-block syntax highlighting. The rich-text primitive makes it possible
  later, but per-language tokenizers are a separate effort.
- Images, HTML passthrough, footnotes, task lists, definition lists. Table column
  alignment markers (`:--`, `:-:`, `--:`) are parsed but not applied in v1; all
  columns render left-aligned.
- Writing literal markdown as `.gsx` children (would require a raw-text lexer
  mode). Content arrives through an expression attribute only.
- The widget owning scroll state or key bindings. Callers wrap it in a
  `scrollable` container, consistent with existing examples.

## Architecture

The work divides into five layers, each independently testable:

```
Layer 1  Rich-text primitive            (tui package: TextSpan, wrapping, render, measure)
Layer 2  OSC 8 hyperlink support        (tui package: Cell, escape, ANSI emit)
Layer 3  Markdown parser                (internal/markdown: zero-dep, recursive block tree)
Layer 4  Markdown component             (tui package: markdown.go, markdown_options.go, theme)
Layer 5  gsx integration                (tuigen generator, analyzer, LSP schema, testdata)
```

Layers 1 and 2 are general framework features with no markdown knowledge. Layer 3
has no dependency on the `tui` package. Layer 4 composes 1, 2, and 3. Layer 5 is
codegen plumbing.

## Layer 1: Rich-text primitive

### Types and API

```go
// TextSpan is a run of text sharing one style. A zero-value Style means the
// span inherits the element's textStyle; set fields override it.
type TextSpan struct {
    Text  string
    Style Style
    Link  string // optional OSC 8 hyperlink target (see Layer 2)
}

// WithRichText sets styled, multi-segment text on an element. When set it takes
// precedence over WithText. Wrapping and alignment behave as for plain text.
func WithRichText(spans ...TextSpan) Option
```

`Element` gains a `richText []TextSpan` field alongside `text`. When `richText`
is non-nil it is the source of truth for text rendering and measurement.

### Style-merge semantics

For each span, the effective per-cell style is the element's inherited
`textStyle` with the span's set fields layered on top:

- Attribute bits (bold, italic, underline, etc.) are OR-ed in.
- A non-default `Fg`/`Bg` on the span overrides the inherited color; a default
  color leaves the inherited value untouched.
- Background from the element/inherited context still applies, matching the
  existing plain-text background handling in `renderTextContent`.

### Wrapping

Add `wrapSpans(spans []TextSpan, maxWidth int) [][]TextSpan` in `text_wrap.go`,
mirroring `wrapParagraph`'s word-packing algorithm:

- Break at word boundaries; fall back to mid-word breaks when a single word
  exceeds `maxWidth` (same rule as today).
- Each emitted word carries the style of the span it came from, so a multi-word
  bold run stays bold across a line break.
- Adjacent same-style segments on a line are merged to keep lines compact.
- Existing newlines split paragraphs, as in `wrapText`.

### Rendering

Add a rich-text branch to `renderTextContent` (`element_render.go`). The existing
per-cell loop used by the text-gradient path already walks runes and calls
`buf.SetRune(x, y, r, style)` with a computed style. The rich-text branch reuses
that loop, supplying each segment's merged style instead of a gradient color.
Per-line alignment, truncation, and the scroll/offset logic operate on line
counts and carry over unchanged.

### Measurement (flagged risk)

Auto-width rich text must report an intrinsic width equal to its concatenated
segment text so table cells and inline contexts size correctly. The exact
integration point in `element_layout.go` / `internal/layout` is resolved during
planning and is the primary technical risk. A short spike confirms plain-text
sizing does not regress.

## Layer 2: OSC 8 hyperlinks

Links render as styled text and, on capable terminals, as real clickable
hyperlinks via the OSC 8 escape sequence. On other terminals the text is still
shown styled and the URL is simply inert.

- `Cell` gains an optional link target (string or interned id) so the ANSI
  emitter knows where hyperlink runs start and end.
- `bufferRowToANSI` / the escape builder (`escape.go`, `render_element.go`)
  wraps contiguous cells sharing a link target in `OSC 8 ; ; URL ST ... OSC 8 ; ; ST`.
- Capability gating reuses the `caps.go` detection approach; when unsupported the
  emitter omits the escape and only the style is applied.

This layer is general framework functionality; markdown links are its first
consumer via `TextSpan.Link`.

## Layer 3: Markdown parser (`internal/markdown`)

Zero-dependency, line-oriented, and independent of the `tui` package so it can be
unit-tested on its own data types.

### Block tree

Because blockquotes and nested lists contain other blocks, the parser produces a
recursive block tree rather than a flat list:

```go
type Block struct {
    Kind     BlockKind   // Heading, Paragraph, CodeFence, Table, List, ListItem, Blockquote
    Level    int         // heading level, or list-nesting depth
    Ordered  bool        // list ordering
    Lang     string      // code fence language tag (stored, unused in v1)
    Inline   []Inline      // inline content for leaf blocks (heading, paragraph)
    Rows     [][]TableCell // table rows; first row is the header
    Children []Block       // nested blocks (blockquote contents, list items)
    Lines    []string      // raw lines for code fences
}

// TableCell holds one cell's inline content (named to avoid colliding with the
// buffer Cell type in the tui package).
type TableCell struct {
    Inline []Inline
}
```

(Concrete field set is finalized in the plan; the shape above is the contract.)

### Inline scanner

Turns a string into `[]Inline` (which Layer 4 converts to `[]TextSpan`) by
toggling state on markers:

- `**text**` / `__text__` -> bold
- `*text*` / `_text_` -> italic
- `` `code` `` -> inline code; suppresses other markers inside
- `[text](url)` -> link with text + target

### Block constructs and disambiguation

- ATX headings: `#`..`######`.
- Setext headings: a paragraph line followed by `===` (h1) or `---` (h2). The
  `---` form is a setext underline only when the preceding line is non-blank
  paragraph text; otherwise `---` is a horizontal rule or, inside a table, the
  separator row.
- Fenced code blocks: ```` ``` ```` delimited; contents are literal, not scanned
  for inline markers.
- Pipe tables: header row, separator row (`---`/`:--:` etc.), body rows.
- Lists: ordered (`1.`) and unordered (`-`, `*`, `+`); nesting tracked by
  indentation, producing nested `List`/`ListItem` blocks.
- Blockquotes: `>` prefix; contents are parsed recursively as blocks.
- Anything unrecognized degrades to a Paragraph of plain text.

## Layer 4: Markdown component

### Files

- `markdown.go` -- `Markdown` struct and `Render`.
- `markdown_options.go` -- option funcs.
- `markdown_theme.go` -- `MarkdownTheme` and `DefaultMarkdownTheme()`.

### Struct and interfaces

```go
type Markdown struct {
    source string
    state  *State[string]   // optional reactive source
    width  int              // 0 = fill available width
    theme  MarkdownTheme
    cache  parseCache       // last source -> parsed block tree
}

var (
    _ Component = (*Markdown)(nil)
    _ AppBinder = (*Markdown)(nil) // only when state-backed
)
```

`Render` resolves the current source (state takes precedence when set), parses it
(or returns the cached tree when the source is unchanged), and walks the block
tree into a `flex-col` `*Element` root.

### Block-to-element mapping

| Block        | Element                                                                 |
|--------------|-------------------------------------------------------------------------|
| Heading      | `WithRichText(spans)` styled by `theme.Heading[level]`                   |
| Paragraph    | `WithRichText(spans)` styled by `theme.Paragraph`                        |
| Inline code  | a `TextSpan` styled by `theme.CodeSpan`                                  |
| Link         | a `TextSpan` styled by `theme.Link` with `Link` set                     |
| Code fence   | bordered/`bg` element, `WithWrap(false)`, plain text                     |
| Table        | existing `<table>/<tr>/<th>/<td>` element tree; each cell is rich text   |
| List         | `flex-col`; each item prefixed `theme.BulletMarker` or `"N. "`, indented per depth |
| Blockquote   | left bar (border) + indent + `theme.Blockquote` text style, recursive    |

### Options

```go
func WithMarkdownSource(s string) MarkdownOption
func WithMarkdownState(s *State[string]) MarkdownOption
func WithMarkdownWidth(w int) MarkdownOption
func WithMarkdownTheme(t MarkdownTheme) MarkdownOption
```

### Caching

A single-entry cache keyed on the source string avoids re-parsing identical
content every frame. State-backed content invalidates the cache when the string
changes. The precise invalidation hook is finalized in the plan.

## MarkdownTheme

A flat struct of `Style` fields plus a few non-style extras, with a
`DefaultMarkdownTheme()` constructor:

```go
type MarkdownTheme struct {
    Heading    [6]Style    // per-level heading styles
    Paragraph  Style
    Bold       Style
    Italic     Style
    CodeSpan   Style       // inline `code`
    Link       Style

    CodeBlockText   Style
    CodeBlockBg     Color
    CodeBlockBorder BorderStyle

    TableBorder BorderStyle
    TableHeader Style

    BlockquoteBar  BorderStyle // left bar
    BlockquoteText Style

    BulletMarker string      // e.g. "• "
}
```

## Layer 5: gsx integration

Expose a self-closing `<markdown>` tag fed by an expression attribute.

- `internal/tuigen/generator_element.go`: add `markdown` to `isComponentElement`,
  map it to `tui.NewMarkdown` in `componentConstructor`, and add an attribute map
  (`source`, `value`, `width`, `theme`). `value`/`source` accept a string literal
  or a `*State[string]` expression; the generator distinguishes them and emits
  `WithMarkdownSource` vs `WithMarkdownState`. Exact discrimination rule is
  finalized in the plan.
- `internal/tuigen/analyzer.go`: add `markdown` to `knownTags`, mark it
  self-closing/void (children rejected).
- `internal/lsp/schema/schema.go`: add the element definition with attribute
  documentation.
- `cmd/tui/testdata/`: add a `markdown.gsx` fixture and its expected
  `markdown_gsx.go`.

Usage:

```gsx
<markdown source={readme} width={80} />
<markdown value={mdState} />
```

## Testing strategy

- Parser (`internal/markdown`): table-driven tests asserting the block tree and
  inline output for each construct and for the disambiguation rules (setext vs
  rule vs table separator, nested lists, blockquote recursion).
- `wrapSpans`: table-driven wrapping cases, including style runs that straddle a
  line break and mid-word breaks.
- Rendering: `MockTerminal` golden tests asserting cell styles (bold/italic runs,
  code-span background, link styling) and layout for tables, code blocks,
  blockquotes, and nested lists.
- OSC 8: assert the emitted ANSI wraps link runs when capable and omits the
  escape when not.
- Codegen: `testdata` golden comparison for the `<markdown>` tag.

## Open questions resolved during planning

These are implementation-level and are answered by reading the relevant code
while writing the plan. They do not change the public design above.

1. Measurement integration point for rich-text intrinsic width
   (`element_layout.go` / `internal/layout`). Flagged as the primary risk; a
   spike confirms no plain-text regression.
2. Parse-cache invalidation hook for state-backed content.
3. Confirming `<td>`/`<th>` intrinsic sizing works when fed `richText`.
4. Code-block overflow: fixed to content width, wrap disabled, horizontal scroll
   left to the caller's container.
5. The `Cell` link representation (inline string vs interned id) for OSC 8.
6. Generator discrimination between string-literal and `*State[string]`
   attribute expressions.

## Future work

- Code-block syntax highlighting built on the rich-text primitive.
- Additional inline/block constructs (images, task lists, footnotes).
- Optional self-scrolling convenience wrapper if demand appears.
