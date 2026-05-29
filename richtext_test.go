package tui

import "testing"

func TestRichText_AccessorRoundTrip(t *testing.T) {
	spans := []TextSpan{
		{Text: "hello "},
		{Text: "world", Style: NewStyle().Bold()},
	}
	e := New(WithRichText(spans...))
	got := e.RichText()
	if len(got) != 2 || got[0].Text != "hello " || got[1].Text != "world" {
		t.Fatalf("RichText() = %+v, want 2 spans hello/world", got)
	}
	if got[1].Style.Attrs&AttrBold == 0 {
		t.Errorf("second span should be bold, got attrs %v", got[1].Style.Attrs)
	}
}

func TestRichText_SettingPlainTextClearsRichText(t *testing.T) {
	e := New(WithRichText(TextSpan{Text: "rich"}))
	e.SetText("plain")
	if len(e.RichText()) != 0 {
		t.Errorf("SetText should clear richText, got %+v", e.RichText())
	}
	if e.Text() != "plain" {
		t.Errorf("Text() = %q, want \"plain\"", e.Text())
	}
}

func TestRichText_SettingRichTextClearsPlainText(t *testing.T) {
	e := New(WithText("plain"))
	e.SetRichText(TextSpan{Text: "rich"})
	if e.Text() != "" {
		t.Errorf("SetRichText should clear text, got %q", e.Text())
	}
	if len(e.RichText()) != 1 {
		t.Errorf("RichText() len = %d, want 1", len(e.RichText()))
	}
}
