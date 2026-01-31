package provider

import (
	"testing"
)

func TestFormat_FixesIndentation(t *testing.T) {
	fp := NewFormattingProvider()

	src := `package test

templ Hello() {
<div class="p-1">
<span>Hello</span>
</div>
}
`
	doc := &Document{
		URI:     "file:///test.gsx",
		Content: src,
		Version: 1,
	}

	edits, err := fp.Format(doc, FormattingOptions{TabSize: 4, InsertSpaces: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(edits) == 0 {
		t.Fatal("expected at least one edit for badly indented document")
	}

	// The edit should replace the full document
	if edits[0].Range.Start.Line != 0 || edits[0].Range.Start.Character != 0 {
		t.Error("expected edit to start at 0:0")
	}

	// The formatted content should not equal the original
	if edits[0].NewText == src {
		t.Error("expected formatted content to differ from original")
	}
}

func TestFormat_ReturnsFullDocumentEdit(t *testing.T) {
	fp := NewFormattingProvider()

	src := `package test

templ Hello() {
<div class="p-1">
<span>Hello</span>
</div>
}
`
	doc := &Document{
		URI:     "file:///test.gsx",
		Content: src,
		Version: 1,
	}

	edits, err := fp.Format(doc, FormattingOptions{TabSize: 4, InsertSpaces: false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(edits) != 1 {
		t.Fatalf("expected 1 edit (full document replace), got %d", len(edits))
	}

	// The edit should start at 0:0
	if edits[0].Range.Start.Line != 0 || edits[0].Range.Start.Character != 0 {
		t.Error("expected edit to start at 0:0")
	}

	// The edit end should cover the last line
	if edits[0].Range.End.Line == 0 {
		t.Error("expected edit to cover multiple lines")
	}
}
