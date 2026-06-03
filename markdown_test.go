package tui

import (
	"strings"
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
	m.ensureParsed() // unchanged source => same slice reused
	if &first[0] != &m.cached[0] {
		t.Errorf("cache should be reused when source is unchanged")
	}
	m.source = "# Bye"
	m.ensureParsed() // changed source => re-parse
	if len(m.cached) == 0 || m.cached[0].Kind != markdown.KindHeading {
		t.Errorf("re-parse failed: %+v", m.cached)
	}
}

func TestMarkdown_RenderHeadingAndParagraph(t *testing.T) {
	m := NewMarkdown(WithMarkdownSource("# Title\n\nHello **world**"), WithMarkdownWidth(20))
	root := m.Render(nil)

	buf := NewBuffer(20, 6)
	root.Render(buf, 20, 6)

	out := buf.StringTrimmed()
	if !strings.Contains(out, "Title") || !strings.Contains(out, "world") {
		t.Fatalf("rendered output missing text:\n%s", out)
	}
	// "Title" on row 0 should be bold (heading style).
	if buf.Cell(0, 0).Style.Attrs&AttrBold == 0 {
		t.Errorf("heading first cell should be bold")
	}
}

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
}

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
	if !strings.Contains(out, "•") {
		t.Errorf("expected bullet marker • in output:\n%s", out)
	}
	_, colAlpha := findCell(buf, 'a')  // first 'a' is in "alpha"
	_, colNested := findCell(buf, 'n') // first 'n' is in "nested"
	if colNested <= colAlpha {
		t.Errorf("nested item col %d should exceed top-level col %d", colNested, colAlpha)
	}
}

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
	if buf.Cell(0, 0).Rune != '│' {
		t.Errorf("expected │ bar at (0,0), got %q", buf.Cell(0, 0).Rune)
	}
}

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
	rName, cName := findCell(buf, 'N')
	if rName < 0 {
		t.Fatal("could not locate header cell")
	}
	if buf.Cell(cName, rName).Style.Attrs&AttrBold == 0 {
		t.Errorf("header cell should be bold")
	}
}

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

	r, c := findCell(buf, 'd') // first 'd' is in "docs"
	if r < 0 {
		t.Fatal("could not find link label")
	}
	if buf.Cell(c, r).Link != "https://example.com" {
		t.Errorf("link cell target = %q, want https://example.com", buf.Cell(c, r).Link)
	}
}

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

func TestMarkdown_HeadingHasTrailingSpace(t *testing.T) {
	m := NewMarkdown(WithMarkdownSource("# Title\nbody text\n"), WithMarkdownWidth(20))
	buf := NewBuffer(20, 5)
	m.Render(nil).Render(buf, 20, 5)
	// Heading on row 0, a blank line on row 1, body on row 2.
	if buf.Cell(0, 0).Rune != 'T' {
		t.Fatalf("heading should be on row 0, got %q", buf.Cell(0, 0).Rune)
	}
	if r := buf.Cell(0, 1).Rune; r != 0 && r != ' ' {
		t.Errorf("expected a blank line after the heading, got %q at (0,1)", r)
	}
	if buf.Cell(0, 2).Rune != 'b' {
		t.Errorf("body should be on row 2 after heading + blank line, got %q", buf.Cell(0, 2).Rune)
	}
}

func TestMarkdown_TableRuleBetweenEveryRow(t *testing.T) {
	src := "| A | B |\n| - | - |\n| 1 | 2 |\n| 3 | 4 |\n"
	m := NewMarkdown(WithMarkdownSource(src), WithMarkdownWidth(20))
	buf := NewBuffer(20, 10)
	m.Render(nil).Render(buf, 20, 10)
	// header + 2 body rows => 2 interior rules (after header, between body rows),
	// each starting with the left-tee junction.
	tees := 0
	for y := range 10 {
		if buf.Cell(0, y).Rune == '├' {
			tees++
		}
	}
	if tees != 2 {
		t.Errorf("expected 2 interior row rules, got %d", tees)
	}
}

func TestMarkdown_HeadingSpacingCountsInHeight(t *testing.T) {
	// The blank line after a heading must be a real row counted in the rendered
	// height (a bottom margin is not counted in a container's auto height, which
	// left the last line of scrollable content unreachable). heading(1) +
	// spacer(1) + body(1) = 3.
	m := NewMarkdown(WithMarkdownSource("# H\n\nbody"), WithMarkdownWidth(20))
	_, h := m.Render(nil).IntrinsicSize()
	if h != 3 {
		t.Errorf("expected total height 3 (heading + blank + body), got %d", h)
	}
}

func TestMarkdown_HeadingSpacingBeforeAndAfterDeduped(t *testing.T) {
	// Adjacent headings get exactly one blank line between them (no duplication);
	// a heading followed by a paragraph gets one blank line too.
	m := NewMarkdown(WithMarkdownSource("# A\n\n## B\n\nbody"), WithMarkdownWidth(20))
	buf := NewBuffer(20, 8)
	m.Render(nil).Render(buf, 20, 8)
	rune0 := func(y int) rune { return buf.Cell(0, y).Rune }
	if rune0(0) != 'A' {
		t.Fatalf("row 0 should be 'A', got %q", rune0(0))
	}
	if r := rune0(1); r != 0 && r != ' ' {
		t.Errorf("row 1 should be blank between the two headings, got %q", r)
	}
	if rune0(2) != 'B' {
		t.Errorf("row 2 should be 'B' (one blank between headings), got %q", rune0(2))
	}
	if r := rune0(3); r != 0 && r != ' ' {
		t.Errorf("row 3 should be blank after the heading, got %q", r)
	}
	if rune0(4) != 'b' {
		t.Errorf("row 4 should be 'body', got %q", rune0(4))
	}
}

func TestMarkdown_CodeFenceDoesNotWrapLongLines(t *testing.T) {
	// A long code line must not wrap: the block height stays lines + border
	// regardless of line length (it scrolls/clips horizontally instead).
	long := "x := veryLongIdentifierThatFarExceedsTheNarrowRenderWidthHere()"
	m := NewMarkdown(WithMarkdownSource("```go\n"+long+"\nshort\n```"), WithMarkdownWidth(20))
	root := m.Render(nil)
	_, h := root.IntrinsicSize()
	// 2 code lines + 2 border rows = 4, no extra rows from wrapping.
	if h != 4 {
		t.Errorf("code fence with a long line should be height 4 (no wrap), got %d", h)
	}
}

func TestMarkdown_TableFullGrid(t *testing.T) {
	src := "| A | B |\n| - | - |\n| 1 | 2 |\n"
	m := NewMarkdown(WithMarkdownSource(src), WithMarkdownWidth(20))
	buf := NewBuffer(20, 6)
	m.Render(nil).Render(buf, 20, 6)
	out := buf.StringTrimmed()

	// DefaultMarkdownTheme draws a full rounded grid: outer corners, a top
	// column junction, a header-rule cross, and column separators.
	if buf.Cell(0, 0).Rune != '╭' {
		t.Errorf("top-left corner should be ╭, got %q", buf.Cell(0, 0).Rune)
	}
	for _, want := range []rune{'┬', '┼', '┴', '│'} {
		if !strings.ContainsRune(out, want) {
			t.Errorf("grid should contain %q:\n%s", want, out)
		}
	}
}

func TestMarkdown_BlockquoteTextIsItalic(t *testing.T) {
	m := NewMarkdown(WithMarkdownSource("> hello world\n"), WithMarkdownWidth(20))
	buf := NewBuffer(20, 4)
	m.Render(nil).Render(buf, 20, 4)
	r, c := findCell(buf, 'h') // first 'h' is "hello"
	if r < 0 {
		t.Fatal("blockquote text not found")
	}
	if buf.Cell(c, r).Style.Attrs&AttrItalic == 0 {
		t.Errorf("blockquote text should be italic")
	}
}

func TestMarkdown_BlockquoteWrapsLongContent(t *testing.T) {
	long := "the quick brown fox jumps over the lazy dog repeatedly"
	m := NewMarkdown(WithMarkdownSource("> "+long+"\n"), WithMarkdownWidth(24))
	root := m.Render(nil)

	buf := NewBuffer(24, 10)
	root.Render(buf, 24, 10)
	out := buf.StringTrimmed()

	// Tail of the line must survive (it was clipped before the wrap fix).
	if !strings.Contains(out, "repeatedly") {
		t.Fatalf("blockquote content should wrap, not clip; got:\n%s", out)
	}
	// Content occupies more than one row, and the bar spans each content row.
	if buf.Cell(0, 0).Rune != '│' || buf.Cell(0, 1).Rune != '│' {
		t.Errorf("bar should span wrapped content rows; row0=%q row1=%q",
			buf.Cell(0, 0).Rune, buf.Cell(0, 1).Rune)
	}
}

func TestMarkdown_ListWrapsLongContent(t *testing.T) {
	long := "the quick brown fox jumps over the lazy dog repeatedly"
	m := NewMarkdown(WithMarkdownSource("- "+long+"\n"), WithMarkdownWidth(24))
	root := m.Render(nil)

	buf := NewBuffer(24, 10)
	root.Render(buf, 24, 10)
	out := buf.StringTrimmed()

	if !strings.Contains(out, "repeatedly") {
		t.Fatalf("list content should wrap, not clip; got:\n%s", out)
	}
}

// findCell returns the row,col of the first occurrence of r in the buffer.
func findCell(buf *Buffer, r rune) (int, int) {
	w, h := buf.Size()
	for y := range h {
		for x := range w {
			if buf.Cell(x, y).Rune == r {
				return y, x
			}
		}
	}
	return -1, -1
}
