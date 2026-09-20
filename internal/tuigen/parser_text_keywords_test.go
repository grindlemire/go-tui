package tuigen

import (
	"fmt"
	"testing"
)

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

// Control flow that starts on the same line as prose must still end the text
// and parse as a loop or conditional, as it did before keywords became text.
func TestParser_KeywordsInsideText_SameLineControlFlow(t *testing.T) {
	type tc struct {
		body     string
		wantText string
		wantNode string
	}

	tests := map[string]tc{
		"if after prose": {
			body:     `Hello if show { <span>y</span> }`,
			wantText: "Hello",
			wantNode: "*tuigen.IfStmt",
		},
		"range for after prose": {
			body:     `Items: for _, x := range xs { <span>{x}</span> }`,
			wantText: "Items:",
			wantNode: "*tuigen.ForLoop",
		},
		"loop-like prose without := stays text": {
			body:     `for each item in range`,
			wantText: "for each item in range",
			wantNode: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			input := "package x\ntempl Test(show bool, xs []string) {\n\t<div>" + tt.body + "</div>\n}"
			file, err := NewParser(NewLexer("test.gsx", input)).ParseFile()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			div := file.Components[0].Body[0].(*Element)
			if len(div.Children) == 0 {
				t.Fatal("expected children")
			}
			text, ok := div.Children[0].(*TextContent)
			if !ok {
				t.Fatalf("child 0 = %T, want *TextContent", div.Children[0])
			}
			if text.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", text.Text, tt.wantText)
			}
			if tt.wantNode == "" {
				if len(div.Children) != 1 {
					t.Errorf("expected only text, got %d children", len(div.Children))
				}
				return
			}
			if len(div.Children) != 2 {
				t.Fatalf("expected text + control flow, got %d children: %#v", len(div.Children), div.Children)
			}
			if got := fmt.Sprintf("%T", div.Children[1]); got != tt.wantNode {
				t.Errorf("child 1 = %s, want %s", got, tt.wantNode)
			}
		})
	}
}

// A Go expression splits prose into separate text nodes on either side.
func TestParser_KeywordsInsideText_AroundExpr(t *testing.T) {
	input := "package x\ntempl Test(key string) {\n\t<span>Press {key} to return</span>\n}"
	file, err := NewParser(NewLexer("test.gsx", input)).ParseFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	span := file.Components[0].Body[0].(*Element)
	if len(span.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(span.Children))
	}
	if got := span.Children[2].(*TextContent).Text; got != "to return" {
		t.Errorf("trailing text = %q, want %q", got, "to return")
	}
}
