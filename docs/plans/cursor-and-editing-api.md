# Design: Cursor reporting, text editing API, and element resize

Status: Draft / proposal (not yet implemented). Adapted to #103 on 2026-06-22.
Base: build on `main` after PR #103 (`st33l-grapheme`) merges. PRs #96 and #97 are
closed; #103 supersedes both.
Scope: public API surface for `TextArea` / `Input` / `Element` cursor and editing.

## Problem

After the grapheme-cluster work, text rendering and the framework's own cursor
are correct for emoji, flags, ZWJ sequences, and skin-tone modifiers. But
building a real text-input UI still requires reaching inside the framework. The
driving consumer app uses `reflect` and `unsafe` in four places:

| Reach inside | What it does | Fix |
|---|---|---|
| insert `\n` at the cursor | reads/writes `cursorPos`, splices text by hand | Part A: editing API (`InsertText`) |
| resize a persisted element each frame | writes `style.Height` | Part B: `Element.SetHeight` |
| place the real terminal cursor | reads cursor state, scroll, absolute position | Part C: cursor reporting protocol |
| read an element's screen position | reads `layout.AbsoluteX/AbsoluteY` | none needed; `Element.Rect().X/Y` is already public (discoverability only) |

## Relationship to #103

#103 already landed the parts of the original plan that overlap with the engine,
and chose to expose the engine rather than encapsulate it:

- Cells store `Rune` + `Combining` (the old #97 storage refactor is absorbed).
- The grapheme engine is public: `StringWidth`, `NextCluster`, `NextClusterRunes`,
  `ClusterCount`, `ClusterRuneCount`. `RuneWidth` remains public. `WithScrollOffset`
  is an `Element` option.

What #103 does NOT provide, and what this design adds:

- No programmatic text editing (`InsertText`, public cursor position).
- No element resize setters.
- No cursor-reporting protocol; the framework still does not place the real
  terminal cursor at a focused widget.

Consequence: this design no longer internalizes `RuneWidth` or hides the engine
(that fight is over, #103 exposed it). The cursor protocol becomes a clean default
layered on top of an already-exposed engine, not a replacement for it.

## Goals

- Build a real text-input UI using public API only (remove the consumer's
  `reflect`/`unsafe`).
- Coexist with #103's exposed engine; reuse its primitives.
- Keep new surface minimal and on the declarative grain where it does not fight a
  real need.

## Non-goals

- Changing `grapheme.go` segmentation.
- Reversing #103's public engine (no `RuneWidth` internalization).
- Terminal cursor shape/blink control (`DECSCUSR`); possible later follow-up.

## Priority

1. Part A (editing API) and Part B (element setters): highest value, they remove
   `reflect`/`unsafe` that #103 leaves untouched.
2. Part C (cursor protocol): clean default and removes the last reflection plus a
   timing bug, but `StringWidth` already lets a consumer compute a correct cursor
   column today, so it is polish rather than a fix.

---

## Part A: Editing API (mutation)

`TextArea` can be typed into but offers no programmatic editing entry point, so
custom edits (for example Shift+Enter inserts a newline) force reflection into
`cursorPos` plus hand-splicing the bound text.

Positions are grapheme-cluster indices (Decision 3): the count of whole glyphs
before the cursor. Translate to and from the internal rune index using #103's
primitives (`ClusterCount` / `ClusterRuneCount` and the internal
`clusterRuneStarts` / `runeIndexToDisplayCol`).

```go
func (t *TextArea) CursorPos() int        // read cursor as a cluster index
func (t *TextArea) SetCursorPos(pos int)  // move cursor to a cluster index (clamped)
func (t *TextArea) InsertText(s string)   // insert at cursor, advance cursor (cluster-correct)
```

`InsertText` routes through the same internal insert path that typing uses (#103:
`cursorPos.Set(clusterEnd(newText, pos))`), keeping the bound `State[string]` and
the cursor consistent by construction. The consumer's newline handler collapses
to:

```go
tui.OnStop(tui.KeyEnter.Shift(), func(ke tui.KeyEvent) { textArea.InsertText("\n") })
```

Document that string mutation goes through `InsertText`, not slicing the text
with a cluster index (a cluster index is not a byte or rune offset).
`DeleteBackward`/`DeleteForward` are already covered by the built-in keymap;
expose them only for symmetry. Ship all three methods on both `TextArea` and
`Input` (Decision 4).

---

## Part B: Element resize after creation (Decision 5)

A consumer holds a persisted element and changes its height each frame (a
fuzzy-match list that grows from 1 to 8 rows), and writes `style.Height` through
`unsafe` because there is no public setter.

```go
func (e *Element) SetHeight(v Value)  // sets style.Height, then e.MarkDirty()
func (e *Element) SetWidth(v Value)   // sets style.Width,  then e.MarkDirty()
```

Each setter marks the element (and ancestors) dirty so layout recomputes next
frame. Constraints:

- Target retained elements (a consumer holding the reference across renders). A
  setter on an element recreated each render has no lasting effect; document this.
- Ship `SetHeight`/`SetWidth` only for now. Hold min/max and the rest of the
  layout surface until a concrete need appears.

The consumer then replaces its `unsafe` write with
`matches.SetHeight(tui.Fixed(rows))`.

Independent of Parts A and C; can land on its own.

---

## Part C: Cursor reporting protocol (output)

Three layers, each contributing only what it owns: the widget computes the
cursor's position within its own content (cluster-aware, after internal scroll);
the element converts content-local to absolute via its laid-out `ContentRect()`;
the app drives the real terminal cursor at end of frame. The position is computed
once and consumed by both presentations (real cursor or drawn glyph) so they
cannot diverge.

### Protocol (element-level)

Focus and absolute layout both live on the element.

```go
type CursorReporter interface {
    // ReportCursor returns the absolute terminal cell and whether to show it.
    ReportCursor() (x, y int, visible bool)
}
```

`*Element` always implements it (returns `visible=false` when no source is set).

### Element side

```go
type cursorSource func() (col, row int, visible bool) // content-local
// field e.cursorSource on Element; option WithCursorSource(fn) / setter SetCursorSource(fn)

func (e *Element) ReportCursor() (int, int, bool) {
    if e.cursorSource == nil { return 0, 0, false }
    col, row, vis := e.cursorSource()
    if !vis { return 0, 0, false }

    cr := e.ContentRect()
    x, y := cr.X+col, cr.Y+row

    // The renderer applies only the OUTERMOST scrollable ancestor's transform to
    // the whole clipped subtree (it never re-bases at nested scrollables). Mirror
    // that: apply that one (ContentRect - scroll) transform and clip against its
    // viewport, shrinking the clip by one when it reserves a vertical-scrollbar
    // column. No scrollable ancestor => the content-local position is screen-absolute.
    var outer *Element
    for anc := e.parent; anc != nil; anc = anc.parent {
        if anc.IsScrollable() { outer = anc }
    }
    if outer == nil { return x, y, true }

    clip := outer.ContentRect()
    sx, sy := outer.ScrollOffset()
    x, y = x+clip.X-sx, y+clip.Y-sy
    if outer.needsVerticalScrollbar() { clip.Width = max(0, clip.Width-1) }
    if x < clip.X || x >= clip.Right() || y < clip.Y || y >= clip.Bottom() {
        return 0, 0, false
    }
    return x, y, true
}
```

### Widget side (names illustrative; map onto #103 internals)

`Input` has a clear internal scroll model on #103: `scrollPos` (a display-column
`State[int]`) with `ensureCursorVisible` and `snapColToNextClusterBoundary`. Its
source is `(col = cursorDisplayCol - scrollPos, row = 0)`.

`TextArea` on #103 has no internal scroll field; it relies on an enclosing
scrollable element. So its source reports the cursor's content-local `(col,row)`
within the textarea, and the enclosing scrollable element's offset is what shifts
it; the implementer must account for that scroll offset when composing the
absolute position (rather than the old `tempScrollOffset` the original plan
assumed). Use the existing cluster-aware helpers (`runeIndexToDisplayCol`,
`clusterEnd`, `snapRuneToClusterStart`) for the column math. If the cursor is
scrolled out of view, report `visible=false` (Decision: out-of-view hides the
cursor).

### App side (end of frame, after Flush)

```go
func (a *App) placeCursor() {
    if cr, ok := a.focusManager.Focused().(CursorReporter); ok {
        if x, y, vis := cr.ReportCursor(); vis {
            a.terminal.SetCursor(x, y); a.terminal.ShowCursor(); return
        }
    }
    a.terminal.HideCursor()
}
```

Timing: placement must be the last terminal op of the frame. `placeCursor()` runs
after `Flush` and then after `postRenderHook`, so no cell writes (including any in
the hook) can clobber the cursor position.

Inline mode (Decision 2): offset the reported coordinates by the inline start row,
mirroring the renderer, so the cursor lands in the inline block.

### Presentation: real cursor is the default (Decision 1)

The real terminal cursor is the default for focused text widgets; the framework
positions it automatically. Overrides:

1. `WithTextAreaVirtualCursor()` switches a widget to the drawn `▌` glyph
   (customizable via `WithTextAreaCursorRune(r)`), which reports `visible=false`.
2. `WithCursorSource(fn)` lets any element report its own cursor.
3. `WithManualCursor()` disables framework cursor management entirely.

Behavior-change note: making the real cursor the default means existing tests
that assert on the drawn `▌` glyph must opt into virtual mode or assert position
through the protocol, and full-screen snapshot tests lose the cursor glyph.

### Docs

Bless `StringWidth` as the width function for strings and cursor math; note
`RuneWidth` is low-level per-rune and wrong for multi-rune clusters. (No code
change; `RuneWidth` stays public per #103.)

---

## Decisions

1. Real terminal cursor is the default; drawn glyph opt-in via
   `WithTextAreaVirtualCursor()`. (Part C; behavior change.)
2. Inline mode handled: placement offsets by the inline start row.
3. `CursorPos()`/`SetCursorPos()` use grapheme-cluster indices.
4. Cursor reporting and editing API ship on both `TextArea` and `Input`.
5. Add mutable `Element.SetHeight`/`SetWidth` (narrow, retained-element scoped).
6. (Adapted 2026-06-22) `RuneWidth` is NOT internalized; the engine stays exposed
   per #103. The original "encapsulate + hide RuneWidth" stage is dropped.

### Remaining micro-questions

- Confirm the opt-into-glyph option name `WithTextAreaVirtualCursor()` (presence =
  on) and that the drawn-glyph mode is retained at all.
- Setter scope: `SetHeight`/`SetWidth` only, or include min/max now?

## Net result

| Reach inside (driving example) | Removed by |
|---|---|
| read/write `cursorPos` plus manual splice for newline | Part A `InsertText` |
| write `style.Height` | Part B `Element.SetHeight` |
| read cursor state, scroll, abs pos for the cursor | Part C protocol (or, interim, `StringWidth` in the consumer's own loop) |
| read `layout.AbsoluteX/AbsoluteY` | already public `Element.Rect()`; docs only |

Parts A and B are the high-value work and remove `reflect`/`unsafe` that #103 does
not address. Part C is a clean default that can follow. All three build on `main`
after #103 merges.
