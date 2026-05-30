# Markdown

## Overview

The `<markdown>` element renders a markdown string into the widget tree. You hand it a document as text and it produces styled elements for headings, emphasis, inline code, fenced code blocks, tables, lists, blockquotes, and links. Fenced code blocks are syntax-highlighted out of the box for Go, JSON, Bash, and JS/TS.

It is a pure content renderer that owns no scroll position or key handling, so you wrap it in a scrollable container when the document is taller than the screen. This keeps the component small and lets you reuse the same scrolling pattern you already use for any other overflowing content.

## The Markdown Element

The tag is self-closing. Content comes from an expression attribute rather than from children, because the generator cannot tell a literal markdown string apart from a Go expression that returns one.

```gsx
templ Doc() {
    <markdown source={"# Hello\n\nSome **bold** text and `inline code`."} />
}
```

The `source` attribute takes any string expression, so the document usually lives in a variable or a function:

```gsx
templ Doc(readme string) {
    <markdown source={readme} />
}
```

A markdown string with backticks and newlines is awkward to write inline in `.gsx`. The example keeps its sample document in `main.go` as a plain Go constant, where a double-quoted string concatenation can hold the backticks that code fences and inline code need, then passes it through the component constructor.

## Static and Reactive Sources

`source` is for content that does not change. When the document updates at runtime, bind a `*State[string]` to the `state` attribute instead and the component re-renders on every change:

```gsx
templ Preview(text *tui.State[string]) {
    <markdown state={text} />
}
```

If you set both, `state` wins. This matches the underlying options: `tui.WithMarkdownState` takes precedence over `tui.WithMarkdownSource`. A live markdown preview pane writes to the state from an input handler and the rendered view follows along.

## Width and Wrapping

The `width` attribute fixes the render width in characters. Leave it at the default of `0` to fill whatever width the parent assigns. At width `0`, paragraphs and headings wrap to the available space, but list and blockquote content renders on a single line and clips on overflow. Set an explicit width to wrap list and blockquote content too, which matters for documents with long bullet or quote lines.

The example derives its width from the terminal size so the document reflows on resize:

```gsx
func (v *viewer) mdWidth(app *tui.App) int {
    w, _ := app.Size()
    if w < 10 {
        w = 10
    }
    return w
}

templ (v *viewer) Render() {
    <markdown source={v.doc} width={v.mdWidth(app)} />
}
```

`app.Size()` is re-read on each render, so resizing the terminal rewraps the text. The floor of 10 keeps the layout sane on a very narrow window.

## Scrolling a Long Document

Since `Markdown` holds no scroll state, you scroll it the same way you scroll any overflowing container: put it inside a `scrollable` element, attach a ref, and drive the offset from a `*State[int]`. This is the pattern from the [Scrolling](scrolling) guide applied to one markdown child.

```gsx
type viewer struct {
    doc     string
    scrollY *tui.State[int]
    content *tui.Ref
}

func Viewer() *viewer {
    return &viewer{
        doc:     sampleDoc,
        scrollY: tui.NewState(0),
        content: tui.NewRef(),
    }
}

func (v *viewer) scrollBy(delta int) {
    el := v.content.El()
    if el == nil {
        return
    }
    _, maxY := el.MaxScroll()
    newY := v.scrollY.Get() + delta
    if newY < 0 {
        newY = 0
    }
    if newY > maxY {
        newY = maxY
    }
    v.scrollY.Set(newY)
}

func (v *viewer) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.On(tui.Rune('q'), func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.On(tui.Rune('j'), func(ke tui.KeyEvent) { v.scrollBy(1) }),
        tui.On(tui.Rune('k'), func(ke tui.KeyEvent) { v.scrollBy(-1) }),
        tui.On(tui.KeyDown, func(ke tui.KeyEvent) { v.scrollBy(1) }),
        tui.On(tui.KeyUp, func(ke tui.KeyEvent) { v.scrollBy(-1) }),
        tui.On(tui.KeyPageDown, func(ke tui.KeyEvent) { v.scrollBy(10) }),
        tui.On(tui.KeyPageUp, func(ke tui.KeyEvent) { v.scrollBy(-10) }),
    }
}

templ (v *viewer) Render() {
    <div class="flex-col">
        <span class="text-gradient-cyan-magenta font-bold">Markdown Viewer</span>
        <div
            ref={v.content}
            class="overflow-y-scroll scrollbar-hidden grow"
            scrollOffset={0, v.scrollY.Get()}>
            <markdown source={v.doc} width={v.mdWidth(app)} />
        </div>
        <span class="font-dim">scroll: wheel/j/k/arrows | q/esc quit</span>
    </div>
}
```

The container uses `grow` to fill the height between the title and the help line, and `scrollbar-hidden` to drop the scrollbar and reclaim its column. Binding `scrollOffset` to `v.scrollY.Get()` keeps the position across renders, since go-tui rebuilds elements on every pass.

## Theming

A `MarkdownTheme` controls how each construct is styled. It is a flat struct of `Style` fields with a few extras for borders and markers. Start from `tui.DefaultMarkdownTheme()` and override what you want, then pass the result to the `theme` attribute.

```go
func docTheme() tui.MarkdownTheme {
    t := tui.DefaultMarkdownTheme()
    t.Heading[0] = tui.NewStyle().Bold().Foreground(tui.Cyan)
    t.CodeSpan = tui.NewStyle().Foreground(tui.Yellow)
    t.BulletMarker = "› "
    return t
}
```

```gsx
<markdown source={doc} theme={docTheme()} />
```

The default theme is glow-inspired and leans on text attributes plus a couple of muted colors so it reads on both dark and light terminals. The fields you are most likely to touch:

| Field | Controls |
|---|---|
| `Heading [6]Style` | Per-level heading styles, indexed `0` for h1 through `5` for h6 |
| `Paragraph` | Body text |
| `Bold`, `Italic`, `CodeSpan`, `Link` | Inline runs, layered over the surrounding text |
| `CodeBlockText`, `CodeBlockBg`, `CodeBlockBorder` | Fenced code block text, fill, and border |
| `TableHeader`, `TableBorder` | Table header cells and the grid border |
| `BlockquoteBar`, `BlockquoteBarStyle`, `BlockquoteText` | The left bar glyph, its style, and quoted text |
| `BulletMarker` | The unordered-list marker string, e.g. `"• "` |

Tables draw as a full grid with an outer box, column separators, and a rule under the header. Blockquotes render a one-column glyph bar on the left rather than a box border, since a `BorderStyle` always draws a full box.

## Syntax Highlighting

Fenced code blocks run through the theme's `CodeHighlighter`. The default is a built-in zero-dependency lexer that colors Go, JSON, Bash, and JS/TS. A language it does not recognize renders in the plain `CodeBlockText` style.

To turn highlighting off, clear the field:

```go
t := tui.DefaultMarkdownTheme()
t.CodeHighlighter = nil // fenced blocks render uncolored
```

To recolor the built-in lexer, build a highlighter from your own palette:

```go
p := tui.DefaultPalette()
p[tui.TokenKeyword] = tui.BrightMagenta
p[tui.TokenString] = tui.BrightGreen
t.CodeHighlighter = tui.NewHighlighter(p)
```

A palette value is any `tui.Color`, so you can also pass a hex color with `tui.HexColor("#ff79c6")` (it returns an error alongside the color).

A `Palette` maps token kinds (`TokenKeyword`, `TokenString`, `TokenComment`, `TokenNumber`, `TokenType`, and so on) to foreground colors. A missing entry means no color, so the base code style shows through.

To plug in a different engine, implement `CodeHighlighter`:

```go
type CodeHighlighter interface {
    Highlight(lang, code string) [][]TextSpan
}
```

`Highlight` receives the whole block so it can track multi-line constructs like raw strings and block comments, and returns one `[]TextSpan` per input line. The concatenated text of each line's spans must equal the input line, so a highlighter colors the code and never rewrites it. A chroma adapter that maps chroma's lexer output to per-line spans fits this interface.

## Native Selection and Clickable Links

The example calls `tui.WithoutMouse()`, so the app does not capture the mouse. The terminal keeps its native behavior: you can select and copy text, and click OSC 8 hyperlinks to open them on capable terminals such as Ghostty, iTerm2, kitty, and WezTerm.

The wheel still scrolls. In full-screen mode with mouse reporting off, go-tui enables alternate-scroll (DEC mode 1007), so the terminal turns wheel notches into arrow keys, and the keymap scrolls on those.

```go
app, err := tui.NewApp(
    tui.WithRootComponent(Viewer()),
    tui.WithoutMouse(),
)
```

To keep copied text clean, the viewer draws no border, padding, or visible scrollbar, so lines sit flush against the left edge. go-tui also clears the space to the right of each line with an erase-to-end-of-line instead of writing spaces, so a copied selection carries no trailing whitespace.

Calling `tui.WithMouse()` instead captures the mouse for click and wheel events. The terminal then gives up native selection and link opening while captured, though you can hold the terminal's bypass modifier (usually Shift) to select anyway.

## Supported Markdown

The sample document in the example exercises the full set:

- ATX headings (`#` through `######`) and single-line setext headings (`===`, `---`)
- Bold, italic, combined bold-italic, and inline code, with both `*`/`_` and `**`/`__` markers
- Links, rendered as OSC 8 hyperlinks on capable terminals
- Fenced code blocks, with blank lines inside the fence preserved
- Pipe tables, rendered as a grid with inline formatting kept inside cells
- Unordered lists (`-`, `*`, `+`), ordered lists, and nested lists
- Blockquotes, including nested quotes and quotes that contain a list

A delimiter with no closer stays literal, so `see **docs` and `3 * 4` render as written rather than turning bold or italic.

## Run

```bash
tui generate markdown.gsx
go run .
```

Scroll with `j`/`k`, the arrow keys, `PageUp`/`PageDown`, or the mouse wheel. Select text with the mouse and click a link to open it. Press `q` or `Esc` to quit.

The markdown viewer should look like this:

![Markdown screenshot](/guides/24.png)

## Next Steps

- [Scrolling](scrolling) for the scrollable container and `scrollOffset` binding that wraps the markdown.
- [State](state) for the `*State[string]` source used to drive a live preview.
- [Styling](styling) for the `Style` and `Color` types that a `MarkdownTheme` is built from.
