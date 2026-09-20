package tuigen

import "testing"

// Text coalescing stopped at any Go keyword token, and a keyword that did not
// start valid control flow was skipped silently, so "Press Escape to return"
// generated as "Press Escape to" (examples/15-inline-mode).
func TestParser_KeywordsInsideText(t *testing.T) {
	type tc struct {
		text string
	}

	tests := map[string]tc{
		"return at end":      {text: "Press Escape to return"},
		"return mid":         {text: "Press Escape to return now"},
		"return alone":       {text: "return"},
		"for mid":            {text: "wait for it"},
		"for at start":       {text: "for the win"},
		"if mid":             {text: "what if"},
		"if at start":        {text: "if only"},
		"else mid":           {text: "or else"},
		"func mid":           {text: "a func here"},
		"var at start":       {text: "var names"},
		"range mid":          {text: "a range of"},
		"type and const":     {text: "type const import package"},
		"templ mid":          {text: "each templ renders"},
		"keyword with punct": {text: "return, or else!"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			input := "package x\ntempl Test() {\n\t<span>" + tt.text + "</span>\n}"
			file, err := NewParser(NewLexer("test.gsx", input)).ParseFile()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			span := file.Components[0].Body[0].(*Element)
			if len(span.Children) != 1 {
				t.Fatalf("expected 1 coalesced text child, got %d: %#v", len(span.Children), span.Children)
			}
			text, ok := span.Children[0].(*TextContent)
			if !ok {
				t.Fatalf("expected *TextContent, got %T", span.Children[0])
			}
			if text.Text != tt.text {
				t.Errorf("Text = %q, want %q", text.Text, tt.text)
			}
		})
	}
}

// Real control flow at the start of a child must still parse as control flow.
func TestParser_KeywordsInsideText_ControlFlowStillParses(t *testing.T) {
	input := `package x
templ Test(items []string, show bool) {
	<div>
		for _, item := range items {
			<span>{item}</span>
		}
		if show {
			<span>on</span>
		} else {
			<span>off</span>
		}
		<span>done for now</span>
	</div>
}`
	file, err := NewParser(NewLexer("test.gsx", input)).ParseFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	div := file.Components[0].Body[0].(*Element)
	if len(div.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(div.Children))
	}
	if _, ok := div.Children[0].(*ForLoop); !ok {
		t.Errorf("child 0 = %T, want *ForLoop", div.Children[0])
	}
	if _, ok := div.Children[1].(*IfStmt); !ok {
		t.Errorf("child 1 = %T, want *IfStmt", div.Children[1])
	}
	span, ok := div.Children[2].(*Element)
	if !ok || len(span.Children) != 1 {
		t.Fatalf("child 2 = %#v, want span with one text child", div.Children[2])
	}
	if got := span.Children[0].(*TextContent).Text; got != "done for now" {
		t.Errorf("Text = %q, want %q", got, "done for now")
	}
}
