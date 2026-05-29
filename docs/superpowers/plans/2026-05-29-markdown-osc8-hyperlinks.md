# OSC 8 Hyperlinks Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `TextSpan.Link` render as a real terminal hyperlink (OSC 8) so markdown links are clickable, with graceful fallback to plain styled text on terminals that don't support it.

**Architecture:** A link travels with the cell through the whole render pipeline: `TextSpan.Link` → `wrapSpans` (carried on `styledRune`) → `drawSpanLines` (written via a new `SetRuneLink`) → `Cell.Link` → buffer `Diff` (already diffs via `Cell.Equal`, which now includes `Link`) → `ANSITerminal.Flush`, which runs a small open/close state machine emitting `OSC 8` around contiguous same-link runs, gated on a new `Capabilities.Hyperlinks` flag. The standalone `bufferRowToANSI` path gets the same treatment.

**Tech Stack:** Go 1.25, no external dependencies. Byte-level tests assert emitted escape sequences via `NewANSITerminalWithCaps(out, in, caps)` with a `bytes.Buffer`; cell-level tests use `RenderTree` + `buf.Cell(x,y)`.

**Plan sequence:** This is **Plan 2 of 4** for issue #62 (spec: `docs/superpowers/specs/2026-05-29-markdown-component-design.md`).
1. Rich-text primitive — **DONE** (merged to main).
2. **OSC 8 hyperlinks (this plan)** — `tui` package.
3. Markdown parser — `internal/markdown`, zero-dep.
4. Markdown component + gsx — composes 1, 2, 3.

This plan depends on Plan 1 (it threads `Link` through `wrapSpans`/`drawSpanLines`, both added in Plan 1).

**Deliberate deviation from the spec:** the spec proposed interning URLs into a per-buffer table referenced by id. This plan stores `Link string` directly on `Cell` instead. Rationale: an extra 16-byte string header per cell is negligible (≈64KB for an 80×50 grid, almost all empty), and interning adds a table lifecycle (allocation, reset on resize, id remapping) with no measured benefit — premature optimization. If profiling later shows cell copies dominate, interning can be added behind the same `Link` accessor without changing callers.

**Conventions for every commit:** use `gcommit -m "..."` (NOT `git commit`), conventional-commit format. Run `go test ./` before each commit and `go test -race ./` at the end. Work on a branch, not `main`.

---

## OSC 8 reference

- Open: `ESC ] 8 ; ; <URL> ST` where `ST` (String Terminator) is the two bytes `ESC \` (`0x1b 0x5c`).
- Close: `ESC ] 8 ; ; ST` (empty URL).
- Bytes below `0x20` and `0x7f` must be stripped from the URL so they can't terminate or corrupt the sequence.

---

## File Structure

- Modify: `cell.go` — add `Cell.Link` field; include it in `Cell.Equal` (`cell.go:40`).
- Modify: `buffer.go` — add `SetRuneLink` (next to `SetRune` at `buffer.go:101`).
- Modify: `escape.go` — add `OpenHyperlink`/`CloseHyperlink` (after `WriteBytes`, `escape.go:355`).
- Modify: `terminal.go` — add `Capabilities.Hyperlinks` (`terminal.go:18`).
- Modify: `caps.go` — set `Hyperlinks` in `DetectCapabilities`.
- Modify: `terminal_ansi.go` — link-run state machine in `Flush` (`terminal_ansi.go:78`).
- Modify: `render_element.go` — link handling in `bufferRowToANSI` (`render_element.go:6`).
- Modify: `text_wrap.go` — carry `link` on `styledRune`; split segments on link change in `wrapSpans`.
- Modify: `element_render.go` — `drawSpanLines` writes links via `SetRuneLink`.
- Tests: `cell_test.go`, `buffer_test.go`, `escape_test.go`, `terminal_ansi_test.go` (create if absent), `caps_test.go`, `text_wrap_test.go`, `element_render_test.go`.

---

## Task 1: Cell carries a link and Diff reacts to it

**Files:**
- Modify: `cell.go` (struct at `:8`, `Equal` at `:40`)
- Modify: `cell_test.go`

- [ ] **Step 1: Write the failing test**

Append to `cell_test.go`:

```go
func TestCell_EqualConsidersLink(t *testing.T) {
	a := NewCell('x', NewStyle())
	b := NewCell('x', NewStyle())
	b.Link = "https://example.com"
	if a.Equal(b) {
		t.Error("cells differing only in Link should not be Equal")
	}
	b.Link = ""
	if !a.Equal(b) {
		t.Error("cells with identical fields (empty Link) should be Equal")
	}
}

func TestBuffer_DiffDetectsLinkChange(t *testing.T) {
	buf := NewBuffer(3, 1)
	buf.Swap() // front == back, no diff
	c := NewCell('a', NewStyle())
	c.Link = "https://example.com"
	buf.SetCell(0, 0, c)
	changes := buf.Diff()
	if len(changes) != 1 || changes[0].Cell.Link != "https://example.com" {
		t.Fatalf("expected one change carrying the link, got %+v", changes)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ -run 'TestCell_EqualConsidersLink|TestBuffer_DiffDetectsLinkChange' -v`
Expected: FAIL — `c.Link` undefined.

- [ ] **Step 3: Add the field and update Equal**

In `cell.go`, add the field to the struct:

```go
type Cell struct {
	Rune  rune  // The character (0 for continuation cells)
	Style Style // Visual styling
	Width uint8 // Display width (1 or 2; 0 for continuation)
	Link  string // Optional OSC 8 hyperlink target ("" = none)
}
```

Update `Equal`:

```go
func (c Cell) Equal(other Cell) bool {
	return c.Rune == other.Rune && c.Style.Equal(other.Style) && c.Width == other.Width && c.Link == other.Link
}
```

(`NewCell`/`NewCellWithWidth` leave `Link` as its zero value `""`, which is correct for plain text.)

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./ -run 'TestCell_EqualConsidersLink|TestBuffer_DiffDetectsLinkChange' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cell.go cell_test.go
gcommit -m "feat: add Link field to Cell and include it in Equal/Diff"
```

---

## Task 2: escBuilder emits OSC 8 open/close

**Files:**
- Modify: `escape.go` (after `WriteBytes`, `:355`)
- Modify: `escape_test.go`

- [ ] **Step 1: Write the failing test**

Append to `escape_test.go`:

```go
func TestEscBuilder_Hyperlink(t *testing.T) {
	e := newEscBuilder(64)
	e.OpenHyperlink("https://example.com")
	e.WriteString("link")
	e.CloseHyperlink()
	got := string(e.Bytes())
	want := "\x1b]8;;https://example.com\x1b\\link\x1b]8;;\x1b\\"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEscBuilder_HyperlinkStripsControlBytes(t *testing.T) {
	e := newEscBuilder(64)
	e.OpenHyperlink("https://x\x1b\n\x07/y") // control bytes must be dropped
	got := string(e.Bytes())
	want := "\x1b]8;;https://x/y\x1b\\"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ -run TestEscBuilder_Hyperlink -v`
Expected: FAIL — `e.OpenHyperlink` / `e.CloseHyperlink` undefined.

- [ ] **Step 3: Implement the methods**

Append to `escape.go`:

```go
// OpenHyperlink writes an OSC 8 hyperlink-open sequence for the given URL.
// Control bytes (< 0x20 and 0x7f) are stripped so they cannot terminate or
// corrupt the sequence.
func (e *escBuilder) OpenHyperlink(url string) {
	e.buf = append(e.buf, 0x1b, ']', '8', ';', ';')
	for i := 0; i < len(url); i++ {
		if c := url[i]; c >= 0x20 && c != 0x7f {
			e.buf = append(e.buf, c)
		}
	}
	e.buf = append(e.buf, 0x1b, '\\')
}

// CloseHyperlink writes an OSC 8 hyperlink-close sequence (empty URL).
func (e *escBuilder) CloseHyperlink() {
	e.buf = append(e.buf, 0x1b, ']', '8', ';', ';', 0x1b, '\\')
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./ -run TestEscBuilder_Hyperlink -v`
Expected: PASS (both).

- [ ] **Step 5: Commit**

```bash
git add escape.go escape_test.go
gcommit -m "feat: add OSC 8 hyperlink escape sequences to escBuilder"
```

---

## Task 3: Capability flag for hyperlinks

**Files:**
- Modify: `terminal.go` (`Capabilities` at `:18`)
- Modify: `caps.go` (`DetectCapabilities`)
- Modify: `caps_test.go`

OSC 8 support is hard to detect reliably and most modern emulators support it, so we enable it by default and only disable it for clearly dumb terminals. This mirrors how the codebase already optimistically assumes `Unicode: true`.

- [ ] **Step 1: Write the failing test**

Append to `caps_test.go`:

```go
func TestDetectCapabilities_HyperlinksDefaultOn(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	if !DetectCapabilities().Hyperlinks {
		t.Error("Hyperlinks should default on for a normal terminal")
	}
}

func TestDetectCapabilities_HyperlinksOffForDumb(t *testing.T) {
	t.Setenv("TERM", "dumb")
	if DetectCapabilities().Hyperlinks {
		t.Error("Hyperlinks should be off for TERM=dumb")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ -run TestDetectCapabilities_Hyperlinks -v`
Expected: FAIL — `Hyperlinks` undefined on `Capabilities`.

- [ ] **Step 3: Add the field**

In `terminal.go`, add to `Capabilities`:

```go
	// KittyKeyboard indicates the Kitty keyboard protocol was successfully negotiated.
	KittyKeyboard bool
	// Hyperlinks indicates the terminal supports OSC 8 hyperlinks.
	Hyperlinks bool
```

- [ ] **Step 4: Set it in DetectCapabilities**

In `caps.go`, in the initial struct literal inside `DetectCapabilities`, add `Hyperlinks: true`:

```go
	caps := Capabilities{
		Colors:     Color16,
		Unicode:    true,
		TrueColor:  false,
		AltScreen:  true,
		Hyperlinks: true,
	}
```

Then, find the place where `DetectCapabilities` handles `TERM=dumb` (search for `"dumb"`); if there is such a guard, also set `caps.Hyperlinks = false` there. If there is no `dumb` handling yet, add this near the top of `DetectCapabilities` after the struct literal:

```go
	if strings.ToLower(os.Getenv("TERM")) == "dumb" {
		caps.Hyperlinks = false
	}
```

(`caps.go` already imports `os` and `strings`.)

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./ -run TestDetectCapabilities -v`
Expected: PASS (the new tests and any existing capability tests).

- [ ] **Step 6: Commit**

```bash
git add terminal.go caps.go caps_test.go
gcommit -m "feat: add Hyperlinks capability, default on except TERM=dumb"
```

---

## Task 4: Flush emits OSC 8 around contiguous same-link runs

**Files:**
- Modify: `terminal_ansi.go` (`Flush` at `:78`)
- Create/Modify: `terminal_ansi_test.go`

- [ ] **Step 1: Write the failing test**

Create `terminal_ansi_test.go` (or append if it exists):

```go
package tui

import (
	"bytes"
	"strings"
	"testing"
)

func linkChanges(url string) []CellChange {
	mk := func(x int, r rune) CellChange {
		c := NewCell(r, NewStyle())
		c.Link = url
		return CellChange{X: x, Y: 0, Cell: c}
	}
	return []CellChange{mk(0, 'a'), mk(1, 'b')}
}

func TestFlush_EmitsHyperlinkOnce(t *testing.T) {
	var out bytes.Buffer
	caps := Capabilities{Colors: Color16, Hyperlinks: true}
	term := NewANSITerminalWithCaps(&out, nil, caps)
	term.Flush(linkChanges("https://example.com"))
	s := out.String()
	if strings.Count(s, "\x1b]8;;https://example.com\x1b\\") != 1 {
		t.Errorf("want exactly one open seq, got: %q", s)
	}
	if strings.Count(s, "\x1b]8;;\x1b\\") != 1 {
		t.Errorf("want exactly one close seq, got: %q", s)
	}
	// Open before the text, close after it.
	if !strings.Contains(s, "\x1b\\ab") || !strings.HasSuffix(s, "\x1b]8;;\x1b\\") {
		t.Errorf("link run not wrapped correctly: %q", s)
	}
}

func TestFlush_NoHyperlinkWhenUnsupported(t *testing.T) {
	var out bytes.Buffer
	caps := Capabilities{Colors: Color16, Hyperlinks: false}
	term := NewANSITerminalWithCaps(&out, nil, caps)
	term.Flush(linkChanges("https://example.com"))
	if strings.Contains(out.String(), "]8;;") {
		t.Errorf("must not emit OSC 8 when unsupported: %q", out.String())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ -run TestFlush_ -v`
Expected: FAIL — no OSC 8 sequences emitted (Flush ignores `Cell.Link`).

- [ ] **Step 3: Add the link-run state machine to Flush**

In `terminal_ansi.go`, modify `Flush`. Add `openLink := ""` next to the `lastX, lastY` initialization, force-close on cursor moves, open/close when the link changes, and close at the end. The full updated function:

```go
func (t *ANSITerminal) Flush(changes []CellChange) {
	if len(changes) == 0 {
		return
	}

	t.esc.Reset()
	lastX, lastY := -1, -1
	openLink := "" // currently-open OSC 8 hyperlink ("" = none)

	for _, ch := range changes {
		if ch.Cell.IsContinuation() {
			continue
		}
		needsMove := false
		if ch.Y != lastY {
			needsMove = true
		} else if ch.X != lastX+1 {
			needsMove = true
		}

		if needsMove {
			// A non-contiguous jump ends any open hyperlink run.
			if t.caps.Hyperlinks && openLink != "" {
				t.esc.CloseHyperlink()
				openLink = ""
			}
			t.esc.MoveTo(ch.X, ch.Y)
		}

		if t.caps.Hyperlinks {
			link := ch.Cell.Link
			if link != openLink {
				if openLink != "" {
					t.esc.CloseHyperlink()
				}
				if link != "" {
					t.esc.OpenHyperlink(link)
				}
				openLink = link
			}
		}

		if !ch.Cell.Style.Equal(t.lastStyle) {
			t.esc.SetStyle(ch.Cell.Style, t.caps)
			t.lastStyle = ch.Cell.Style
		}

		if ch.Cell.Rune != 0 {
			t.esc.WriteRune(ch.Cell.Rune)
		} else {
			t.esc.WriteRune(' ')
		}

		lastX = ch.X
		if ch.Cell.Width > 1 {
			lastX = ch.X + int(ch.Cell.Width) - 1
		}
		lastY = ch.Y
	}

	if t.caps.Hyperlinks && openLink != "" {
		t.esc.CloseHyperlink()
	}

	t.out.Write(t.esc.Bytes())
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./ -run TestFlush_ -v`
Expected: PASS (both).

- [ ] **Step 5: Run the full package**

Run: `go test ./`
Expected: PASS (Flush behavior unchanged when no cell carries a link).

- [ ] **Step 6: Commit**

```bash
git add terminal_ansi.go terminal_ansi_test.go
gcommit -m "feat: emit OSC 8 hyperlinks for linked cells in Flush"
```

---

## Task 5: Buffer.SetRuneLink

**Files:**
- Modify: `buffer.go` (after `SetRune`, `:101`)
- Modify: `buffer_test.go`

- [ ] **Step 1: Write the failing test**

Append to `buffer_test.go`:

```go
func TestSetRuneLink(t *testing.T) {
	buf := NewBuffer(4, 1)
	buf.SetRuneLink(0, 0, 'a', NewStyle(), "https://example.com")
	if got := buf.Cell(0, 0); got.Rune != 'a' || got.Link != "https://example.com" {
		t.Errorf("got rune=%q link=%q, want 'a' / the URL", got.Rune, got.Link)
	}
	// Empty link leaves Link clear (and still writes the rune).
	buf.SetRuneLink(1, 0, 'b', NewStyle(), "")
	if got := buf.Cell(1, 0); got.Rune != 'b' || got.Link != "" {
		t.Errorf("got rune=%q link=%q, want 'b' / empty", got.Rune, got.Link)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./ -run TestSetRuneLink -v`
Expected: FAIL — `buf.SetRuneLink` undefined.

- [ ] **Step 3: Implement SetRuneLink**

In `buffer.go`, add after `SetRune` (which ends around `:149`):

```go
// SetRuneLink sets a rune like SetRune and, when link is non-empty, attaches it
// as the cell's OSC 8 hyperlink target. Wide-character handling matches SetRune.
func (b *Buffer) SetRuneLink(x, y int, r rune, style Style, link string) {
	b.SetRune(x, y, r, style)
	if link == "" {
		return
	}
	if idx := b.idx(x, y); idx >= 0 {
		b.back[idx].Link = link
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./ -run TestSetRuneLink -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add buffer.go buffer_test.go
gcommit -m "feat: add Buffer.SetRuneLink to attach hyperlink targets"
```

---

## Task 6: Thread Link through wrapSpans and drawSpanLines

**Files:**
- Modify: `text_wrap.go` (`styledRune`, `emit`, the rune loop in `wrapSpans`)
- Modify: `element_render.go` (`drawSpanLines`)
- Modify: `text_wrap_test.go`, `element_render_test.go`

- [ ] **Step 1: Write the failing tests**

Append to `text_wrap_test.go`:

```go
func TestWrapSpans_PreservesLinkAndSplitsOnLinkChange(t *testing.T) {
	// "ab"(link X) + "cd"(link Y), no whitespace = one logical word "abcd" that
	// must split into two segments at the link boundary, each keeping its link.
	lines := wrapSpans([]TextSpan{
		{Text: "ab", Link: "X"},
		{Text: "cd", Link: "Y"},
	}, 40)
	if len(lines) != 1 || len(lines[0]) != 2 {
		t.Fatalf("want one line of two segments, got %+v", lines)
	}
	if lines[0][0].Text != "ab" || lines[0][0].Link != "X" {
		t.Errorf("seg 0 = %+v, want {ab, X}", lines[0][0])
	}
	if lines[0][1].Text != "cd" || lines[0][1].Link != "Y" {
		t.Errorf("seg 1 = %+v, want {cd, Y}", lines[0][1])
	}
}
```

Append to `element_render_test.go`:

```go
func TestRichText_LinkReachesCell(t *testing.T) {
	buf := NewBuffer(10, 1)
	e := New(
		WithSize(6, 1),
		WithRichText(TextSpan{Text: "ab", Link: "https://example.com"}),
	)
	e.Calculate(10, 1)
	RenderTree(buf, e)
	if got := buf.Cell(0, 0).Link; got != "https://example.com" {
		t.Errorf("cell(0,0).Link = %q, want the URL", got)
	}
	if got := buf.Cell(1, 0).Link; got != "https://example.com" {
		t.Errorf("cell(1,0).Link = %q, want the URL", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./ -run 'TestWrapSpans_PreservesLinkAndSplitsOnLinkChange|TestRichText_LinkReachesCell' -v`
Expected: FAIL — segments lose `Link` (it's dropped by `styledRune`), and `cell.Link` is empty.

- [ ] **Step 3: Carry link on styledRune**

In `text_wrap.go`, update the struct (and its comment) to carry the link:

```go
// styledRune is one rune carrying the style and link of its source span.
type styledRune struct {
	r    rune
	st   Style
	link string
}
```

- [ ] **Step 4: Make emit merge only on identical style AND link**

In `wrapSpans`, update `emit` so a link change starts a new segment, and rebuilt segments carry the link:

```go
	emit := func(rs []styledRune) {
		for _, sr := range rs {
			if n := len(cur); n > 0 && cur[n-1].Style == sr.st && cur[n-1].Link == sr.link {
				cur[n-1].Text += string(sr.r)
			} else {
				cur = append(cur, TextSpan{Text: string(sr.r), Style: sr.st, Link: sr.link})
			}
		}
	}
```

- [ ] **Step 5: Populate link when building words**

In `wrapSpans`, in the `default:` branch of the rune switch, carry the span's link:

```go
			default:
				word = append(word, styledRune{r: r, st: sp.Style, link: sp.Link})
				wordWidth += RuneWidth(r)
```

(Leave `emitSpace` as-is: separator spaces remain neutral — no style, no link — matching the existing deliberate choice. A multi-word link therefore renders as per-word clickable runs with neutral spaces between; acceptable for v1.)

- [ ] **Step 6: Write the link into cells in drawSpanLines**

In `element_render.go`, in `drawSpanLines`, replace the `buf.SetRune(...)` call with `SetRuneLink`, passing the segment's link:

```go
				if x >= clip.X {
					style := st
					if style.Bg.IsDefault() {
						if cellBg := buf.Cell(x, y).Style.Bg; !cellBg.IsDefault() {
							style.Bg = cellBg
						}
					}
					buf.SetRuneLink(x, y, r, style, span.Link)
				}
```

- [ ] **Step 7: Run tests to verify they pass**

Run: `go test ./ -run 'TestWrapSpans|TestRichText' -v`
Expected: PASS (the new tests plus all existing rich-text/wrap tests).

- [ ] **Step 8: Commit**

```bash
git add text_wrap.go element_render.go text_wrap_test.go element_render_test.go
gcommit -m "feat: thread TextSpan.Link through wrapping and rich-text rendering"
```

---

## Task 7: bufferRowToANSI honors links

`bufferRowToANSI` (`render_element.go:6`) is the standalone row renderer used by `PrintAbove`/inline output. It renders one contiguous row, so link handling is simpler than `Flush`: open when entering a linked run, close when it ends or the row ends.

**Files:**
- Modify: `render_element.go` (`bufferRowToANSI`)
- Modify: `render_element_test.go` (create if absent)

- [ ] **Step 1: Read the current function**

Run: `sed -n '1,70p' render_element.go` and read `bufferRowToANSI` so you match its existing per-cell loop and style handling.

- [ ] **Step 2: Write the failing test**

Append to `render_element_test.go`:

```go
func TestBufferRowToANSI_EmitsHyperlink(t *testing.T) {
	buf := NewBuffer(3, 1)
	buf.SetRuneLink(0, 0, 'a', NewStyle(), "https://example.com")
	buf.SetRuneLink(1, 0, 'b', NewStyle(), "https://example.com")
	esc := newEscBuilder(64)
	caps := Capabilities{Colors: Color16, Hyperlinks: true}
	row := bufferRowToANSI(buf, 0, esc, caps)
	if !strings.Contains(row, "\x1b]8;;https://example.com\x1b\\") {
		t.Errorf("missing OSC 8 open: %q", row)
	}
	if !strings.Contains(row, "\x1b]8;;\x1b\\") {
		t.Errorf("missing OSC 8 close: %q", row)
	}
}
```

Ensure the test file imports `strings`.

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./ -run TestBufferRowToANSI_EmitsHyperlink -v`
Expected: FAIL — no OSC 8 in the row output.

- [ ] **Step 4: Add link handling**

In `bufferRowToANSI`, track an `openLink := ""` across the row's cell loop (gated on `caps.Hyperlinks`): before writing each cell's rune, if `cell.Link != openLink`, close the open link (if any) then open the new one (if non-empty) and update `openLink`. After the loop, if `openLink != ""`, close it. Mirror the exact pattern from `Flush` Task 4, Step 3 (minus the cursor-move handling, since a row is contiguous). Use `esc.OpenHyperlink`/`esc.CloseHyperlink`.

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./ -run TestBufferRowToANSI_EmitsHyperlink -v`
Expected: PASS.

- [ ] **Step 6: Run the full package**

Run: `go test ./`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add render_element.go render_element_test.go
gcommit -m "feat: emit OSC 8 hyperlinks in bufferRowToANSI"
```

---

## Task 8: End-to-end integration + race

Prove a linked rich-text span renders to cells AND produces exactly one OSC 8 run through the real diff→Flush pipeline.

**Files:**
- Modify: `terminal_ansi_test.go`

- [ ] **Step 1: Write the test**

Append to `terminal_ansi_test.go`:

```go
func TestRichTextLink_EndToEnd(t *testing.T) {
	buf := NewBuffer(12, 1)
	e := New(
		WithSize(12, 1),
		WithRichText(
			TextSpan{Text: "see "},
			TextSpan{Text: "site", Link: "https://example.com"},
		),
	)
	e.Calculate(12, 1)
	RenderTree(buf, e)

	var out bytes.Buffer
	term := NewANSITerminalWithCaps(&out, nil, Capabilities{Colors: Color16, Hyperlinks: true})
	term.Flush(buf.Diff())
	s := out.String()
	if strings.Count(s, "\x1b]8;;https://example.com\x1b\\") != 1 {
		t.Errorf("want one hyperlink open around \"site\", got: %q", s)
	}
	if strings.Count(s, "\x1b]8;;\x1b\\") != 1 {
		t.Errorf("want one hyperlink close, got: %q", s)
	}
}
```

- [ ] **Step 2: Run the test**

Run: `go test ./ -run TestRichTextLink_EndToEnd -v`
Expected: PASS (no new production code needed — this exercises Tasks 1–6 together).

- [ ] **Step 3: Run the full suite with the race detector**

Run: `go test -race ./...`
Expected: PASS across all packages.

- [ ] **Step 4: Commit**

```bash
git add terminal_ansi_test.go
gcommit -m "test: end-to-end OSC 8 hyperlink from rich text through Flush"
```

---

## Self-Review (completed by plan author)

**Spec coverage (Layer 2 of the spec):**
- `Cell` carries a link; `Equal`/`Diff` react to it — Task 1.
- OSC 8 escape sequences — Task 2.
- Capability gating (`Hyperlinks`, default on, off for dumb) — Task 3.
- `Flush` link-run state machine (open/close, contiguous, hot-path: zero work when no cell has a link because the `link != openLink` check is a cheap string compare and is skipped entirely when `!caps.Hyperlinks`) — Task 4.
- `SetRuneLink` to attach links to cells — Task 5.
- Threading `Link` through `wrapSpans`/`drawSpanLines` (the gap Plan 1 documented) — Task 6.
- `bufferRowToANSI` parity — Task 7.
- End-to-end through diff→Flush — Task 8.

**Deviation from spec:** plain `Cell.Link string` instead of an interned id table (justified above; YAGNI). Documented in the header.

**Known limitations (carried forward, not bugs):** separator spaces inside a multi-word link are not themselves linked (matches the existing neutral-separator decision); a non-contiguous cursor jump closes and reopens the link rather than relying on OSC 8 `id=` continuation (simpler and avoids accidentally linking skipped cells).

**Placeholder scan:** none — every code/test step has complete content. Task 7 Step 4 references "mirror the Flush pattern from Task 4 Step 3," which is fully spelled out there; Task 7 Step 1 asks the engineer to read the existing function first because its exact loop shape (variable names, style-emit call) must be matched in place.

**Type consistency:** `Cell.Link string`; `SetRuneLink(x, y int, r rune, style Style, link string)`; `escBuilder.OpenHyperlink(url string)` / `CloseHyperlink()`; `Capabilities.Hyperlinks bool`; `styledRune{r, st, link}`. Names match across all tasks.
