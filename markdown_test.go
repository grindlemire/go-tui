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
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if buf.Cell(x, y).Rune == r {
				return y, x
			}
		}
	}
	return -1, -1
}
