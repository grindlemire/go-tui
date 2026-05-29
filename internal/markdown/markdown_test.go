package markdown

import "testing"

func TestParse_Headings(t *testing.T) {
	blocks := Parse("# H1\n## H2\n### H3")
	if len(blocks) != 3 {
		t.Fatalf("want 3 blocks, got %d", len(blocks))
	}
	for i, want := range []int{1, 2, 3} {
		if blocks[i].Kind != KindHeading || blocks[i].Level != want {
			t.Errorf("block %d = kind %d level %d, want heading level %d", i, blocks[i].Kind, blocks[i].Level, want)
		}
	}
	if blocks[0].Inline[0].Text != "H1" {
		t.Errorf("h1 text = %q, want H1", blocks[0].Inline[0].Text)
	}
}

func TestParse_Setext(t *testing.T) {
	blocks := Parse("Title\n=====\n\nSub\n---")
	if len(blocks) != 2 {
		t.Fatalf("want 2 blocks, got %d: %+v", len(blocks), blocks)
	}
	if blocks[0].Kind != KindHeading || blocks[0].Level != 1 || blocks[0].Inline[0].Text != "Title" {
		t.Errorf("block 0 = %+v, want setext h1 Title", blocks[0])
	}
	if blocks[1].Kind != KindHeading || blocks[1].Level != 2 || blocks[1].Inline[0].Text != "Sub" {
		t.Errorf("block 1 = %+v, want setext h2 Sub", blocks[1])
	}
}

func TestParse_Paragraph(t *testing.T) {
	blocks := Parse("one line\ntwo line")
	if len(blocks) != 1 || blocks[0].Kind != KindParagraph {
		t.Fatalf("want 1 paragraph, got %+v", blocks)
	}
	if blocks[0].Inline[0].Text != "one line two line" {
		t.Errorf("joined text = %q", blocks[0].Inline[0].Text)
	}
}

func TestParse_CodeFence(t *testing.T) {
	blocks := Parse("```go\nfunc main() {}\nx := 1\n```")
	if len(blocks) != 1 || blocks[0].Kind != KindCodeFence {
		t.Fatalf("want 1 code fence, got %+v", blocks)
	}
	if blocks[0].Lang != "go" {
		t.Errorf("lang = %q, want go", blocks[0].Lang)
	}
	if len(blocks[0].Lines) != 2 || blocks[0].Lines[0] != "func main() {}" || blocks[0].Lines[1] != "x := 1" {
		t.Errorf("lines = %+v", blocks[0].Lines)
	}
}

func TestParse_Table(t *testing.T) {
	blocks := Parse("| A | B |\n| - | - |\n| 1 | 2 |\n| 3 | 4 |")
	if len(blocks) != 1 || blocks[0].Kind != KindTable {
		t.Fatalf("want 1 table, got %+v", blocks)
	}
	rows := blocks[0].Rows
	if len(rows) != 3 {
		t.Fatalf("want 3 rows (header + 2), got %d", len(rows))
	}
	if len(rows[0]) != 2 || rows[0][0].Inline[0].Text != "A" || rows[0][1].Inline[0].Text != "B" {
		t.Errorf("header = %+v", rows[0])
	}
	if rows[2][1].Inline[0].Text != "4" {
		t.Errorf("cell (2,1) = %+v, want 4", rows[2][1])
	}
}

func TestParse_FullDocument(t *testing.T) {
	src := "# Title\n" +
		"\n" +
		"A para with **bold** and a [link](https://go.dev).\n" +
		"\n" +
		"```go\nx := 1\n```\n" +
		"\n" +
		"| A | B |\n| - | - |\n| 1 | 2 |\n" +
		"\n" +
		"- one\n- two\n  - nested\n" +
		"\n" +
		"> quoted para\n"
	blocks := Parse(src)

	wantKinds := []BlockKind{KindHeading, KindParagraph, KindCodeFence, KindTable, KindList, KindBlockquote}
	if len(blocks) != len(wantKinds) {
		t.Fatalf("want %d blocks, got %d: %+v", len(wantKinds), len(blocks), blocks)
	}
	for i, want := range wantKinds {
		if blocks[i].Kind != want {
			t.Errorf("block %d kind = %d, want %d", i, blocks[i].Kind, want)
		}
	}

	// Link URL inside the paragraph.
	para := blocks[1]
	var foundLink bool
	for _, in := range para.Inline {
		if in.Link == "https://go.dev" && in.Text == "link" {
			foundLink = true
		}
	}
	if !foundLink {
		t.Errorf("paragraph missing link: %+v", para.Inline)
	}

	// Nested list item under the second top-level item.
	list := blocks[4]
	if len(list.Children) != 2 {
		t.Fatalf("want 2 list items, got %d", len(list.Children))
	}
	second := list.Children[1]
	if len(second.Children) != 1 || second.Children[0].Kind != KindList {
		t.Fatalf("second item should have a nested list, got %+v", second.Children)
	}
	nestedItem := second.Children[0].Children[0]
	if nestedItem.Inline[0].Text != "nested" {
		t.Errorf("nested item text = %q, want nested", nestedItem.Inline[0].Text)
	}

	// Blockquote contains a paragraph.
	bq := blocks[5]
	if len(bq.Children) != 1 || bq.Children[0].Kind != KindParagraph {
		t.Errorf("blockquote children = %+v", bq.Children)
	}
}
