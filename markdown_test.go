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
