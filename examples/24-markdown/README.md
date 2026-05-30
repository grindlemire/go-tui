# 24 - Markdown

Renders a markdown document with the `<markdown>` element inside a scrollable
container.

```bash
go run ../../cmd/tui generate markdown.gsx
go run .
```

## What it shows

- The `<markdown>` tag fed by a `source` expression with a responsive `width`:
  `mdWidth(app)` derives the width from `app.Size()`, so the text fills a wide
  terminal and wraps on a narrow one (re-evaluated on resize).
- A `MarkdownTheme` default styling headings, emphasis, inline code, code blocks,
  tables, lists, and blockquotes.
- The component wrapped in a bordered `overflow-y-scroll` container that grows to
  fill the height, with the title and help text outside the frame. `Markdown`
  owns no scroll state of its own.

The sample document in `main.go` exercises every supported construct:

- ATX headings (`#`..`######`) and single-line setext headings (`===`, `---`)
- Bold, italic, combined bold-italic, inline code, and links (OSC 8 on capable
  terminals), with both `*`/`_` and `**`/`__` markers
- Edge cases: a delimiter with no closer stays literal (`see **docs`, `3 * 4`)
- Fenced code blocks, including a preserved blank line
- Pipe tables rendered as a full grid (outer box, column separators, header
  rule) with inline formatting preserved in cells
- Unordered lists (`-`, `*`, `+`), ordered lists, and nesting
- Blockquotes, including a nested quote and a quote containing a list

## Controls

- `j` / `k` or arrow keys: scroll
- `PageUp` / `PageDown`: scroll a page
- mouse wheel: scroll
- `q` / `Esc`: quit

## Selecting text and clicking links

This example calls `tui.WithoutMouse()`, so it does not capture the mouse. That
leaves the terminal's native behavior intact: you can select and copy text, and
click OSC 8 hyperlinks (rendered on capable terminals such as Ghostty, iTerm2,
kitty, and WezTerm) to open them, the same as in any other terminal program.

The mouse wheel still scrolls. In full-screen mode with mouse reporting off,
go-tui enables alternate-scroll (DEC mode 1007), so the terminal translates wheel
notches into arrow keys, which the keymap scrolls on. The result: native
selection and clickable links without giving up wheel scrolling.

To keep copied text clean, the viewer draws no border, padding, or visible
scrollbar, so lines sit flush against the edges. go-tui also clears empty space
to the right of each line with an erase-to-end-of-line rather than writing
spaces, so the terminal trims trailing whitespace from a copied selection.

If you instead call `tui.WithMouse()`, the app captures the mouse for click and
wheel events, and the terminal no longer does native selection or link opening
(hold the terminal's bypass modifier, usually Shift, to do so while captured).
