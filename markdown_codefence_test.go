package tui

import (
	"strings"
	"testing"

	"github.com/grindlemire/go-tui/internal/markdown"
)

func TestRenderCodeFenceHighlights(t *testing.T) {
	m := NewMarkdown(WithMarkdownSource("")) // theme defaults, includes highlighter

	block := markdown.Block{Kind: markdown.KindCodeFence, Lang: "go", Lines: []string{"var x = 1"}}
	box := m.renderCodeFence(block)

	// box -> inner -> one child per line.
	inner := box.children[0]
	line := inner.children[0]
	spans := line.RichText()
	if len(spans) == 0 {
		t.Fatal("expected rich-text spans on the highlighted code line")
	}
	// Contract: concatenated span text equals the source line.
	var b strings.Builder
	for _, s := range spans {
		b.WriteString(s.Text)
	}
	if b.String() != "var x = 1" {
		t.Errorf("reconstructed %q, want %q", b.String(), "var x = 1")
	}
	// "var" span is keyword-colored.
	kw := DefaultPalette()[TokenKeyword]
	for _, s := range spans {
		if s.Text == "var" && !s.Style.Fg.Equal(kw) {
			t.Errorf("var foreground = %v, want %v", s.Style.Fg, kw)
		}
	}
}

func TestRenderCodeFencePlainWhenNoHighlighter(t *testing.T) {
	th := DefaultMarkdownTheme()
	th.CodeHighlighter = nil
	m := NewMarkdown(WithMarkdownSource(""), WithMarkdownTheme(th))

	block := markdown.Block{Kind: markdown.KindCodeFence, Lang: "go", Lines: []string{"var x = 1"}}
	box := m.renderCodeFence(block)
	line := box.children[0].children[0]
	if len(line.RichText()) != 0 {
		t.Error("expected plain text (no rich-text spans) when highlighter is nil")
	}
}
