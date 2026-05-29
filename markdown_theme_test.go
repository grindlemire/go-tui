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
