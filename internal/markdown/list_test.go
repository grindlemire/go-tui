package markdown

import "testing"

func TestParse_UnorderedList(t *testing.T) {
	blocks := Parse("- a\n- b\n- c")
	if len(blocks) != 1 || blocks[0].Kind != KindList || blocks[0].Ordered {
		t.Fatalf("want one unordered list, got %+v", blocks)
	}
	if len(blocks[0].Children) != 3 {
		t.Fatalf("want 3 items, got %d", len(blocks[0].Children))
	}
	for i, want := range []string{"a", "b", "c"} {
		item := blocks[0].Children[i]
		if item.Kind != KindListItem || item.Inline[0].Text != want {
			t.Errorf("item %d = %+v, want %q", i, item, want)
		}
	}
}

func TestParse_OrderedList(t *testing.T) {
	blocks := Parse("1. first\n2. second")
	if len(blocks) != 1 || !blocks[0].Ordered {
		t.Fatalf("want ordered list, got %+v", blocks)
	}
	if len(blocks[0].Children) != 2 || blocks[0].Children[1].Inline[0].Text != "second" {
		t.Errorf("items = %+v", blocks[0].Children)
	}
}

func TestParse_NestedList(t *testing.T) {
	blocks := Parse("- top\n  - child\n  - child2")
	list := blocks[0]
	if len(list.Children) != 1 {
		t.Fatalf("want 1 top item, got %d: %+v", len(list.Children), list.Children)
	}
	top := list.Children[0]
	if len(top.Children) != 1 || top.Children[0].Kind != KindList {
		t.Fatalf("top item should have a nested list, got %+v", top.Children)
	}
	nested := top.Children[0]
	if len(nested.Children) != 2 || nested.Children[0].Inline[0].Text != "child" {
		t.Errorf("nested list = %+v", nested.Children)
	}
}

func TestParse_BlockquoteParagraph(t *testing.T) {
	blocks := Parse("> hello there")
	if len(blocks) != 1 || blocks[0].Kind != KindBlockquote {
		t.Fatalf("want blockquote, got %+v", blocks)
	}
	if len(blocks[0].Children) != 1 || blocks[0].Children[0].Kind != KindParagraph {
		t.Fatalf("blockquote should contain a paragraph, got %+v", blocks[0].Children)
	}
	if blocks[0].Children[0].Inline[0].Text != "hello there" {
		t.Errorf("text = %q", blocks[0].Children[0].Inline[0].Text)
	}
}

func TestParse_BlockquoteWithNestedList(t *testing.T) {
	blocks := Parse("> - a\n> - b")
	bq := blocks[0]
	if bq.Kind != KindBlockquote || len(bq.Children) != 1 || bq.Children[0].Kind != KindList {
		t.Fatalf("blockquote should contain a list, got %+v", bq.Children)
	}
	if len(bq.Children[0].Children) != 2 {
		t.Errorf("nested list items = %+v", bq.Children[0].Children)
	}
}
