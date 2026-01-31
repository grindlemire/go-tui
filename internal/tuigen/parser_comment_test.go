package tuigen

import (
	"testing"
)

func TestParser_CommentAttachment_LeadingCommentOnComponent(t *testing.T) {
	input := `package x

// This is a doc comment for Header
// It spans multiple lines
templ Header() {
	<span>Hello</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(file.Components))
	}

	comp := file.Components[0]
	if comp.LeadingComments == nil {
		t.Fatal("expected LeadingComments, got nil")
	}

	if len(comp.LeadingComments.List) != 2 {
		t.Fatalf("expected 2 comments in group, got %d", len(comp.LeadingComments.List))
	}

	if comp.LeadingComments.List[0].Text != "// This is a doc comment for Header" {
		t.Errorf("comment 0 text = %q", comp.LeadingComments.List[0].Text)
	}
	if comp.LeadingComments.List[1].Text != "// It spans multiple lines" {
		t.Errorf("comment 1 text = %q", comp.LeadingComments.List[1].Text)
	}
}

func TestParser_CommentAttachment_TrailingCommentOnComponent(t *testing.T) {
	input := `package x

templ Header() { // trailing comment on brace
	<span>Hello</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]
	if comp.TrailingComments == nil {
		t.Fatal("expected TrailingComments, got nil")
	}

	if len(comp.TrailingComments.List) != 1 {
		t.Fatalf("expected 1 trailing comment, got %d", len(comp.TrailingComments.List))
	}

	if comp.TrailingComments.List[0].Text != "// trailing comment on brace" {
		t.Errorf("trailing comment text = %q", comp.TrailingComments.List[0].Text)
	}
}

func TestParser_CommentAttachment_OrphanCommentInComponentBody(t *testing.T) {
	input := `package x

templ Header() {
	// orphan comment in body
	<span>Hello</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]

	// The orphan comment should be attached as leading comment to the span element
	if len(comp.Body) != 1 {
		t.Fatalf("expected 1 body node, got %d", len(comp.Body))
	}

	elem, ok := comp.Body[0].(*Element)
	if !ok {
		t.Fatalf("expected *Element, got %T", comp.Body[0])
	}

	if elem.LeadingComments == nil {
		t.Fatal("expected LeadingComments on element, got nil")
	}

	if len(elem.LeadingComments.List) != 1 {
		t.Fatalf("expected 1 leading comment, got %d", len(elem.LeadingComments.List))
	}

	if elem.LeadingComments.List[0].Text != "// orphan comment in body" {
		t.Errorf("leading comment text = %q", elem.LeadingComments.List[0].Text)
	}
}

func TestParser_CommentAttachment_OrphanCommentWithNoFollowingNode(t *testing.T) {
	input := `package x

templ Header() {
	<span>Hello</span>
	// trailing orphan comment
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]

	// The trailing orphan comment should be in OrphanComments
	if len(comp.OrphanComments) != 1 {
		t.Fatalf("expected 1 orphan comment group, got %d", len(comp.OrphanComments))
	}

	if len(comp.OrphanComments[0].List) != 1 {
		t.Fatalf("expected 1 comment in orphan group, got %d", len(comp.OrphanComments[0].List))
	}

	if comp.OrphanComments[0].List[0].Text != "// trailing orphan comment" {
		t.Errorf("orphan comment text = %q", comp.OrphanComments[0].List[0].Text)
	}
}

func TestParser_CommentAttachment_OrphanCommentInFile(t *testing.T) {
	input := `package x

templ Header() {
	<span>Hello</span>
}

// orphan comment at end of file`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.OrphanComments) != 1 {
		t.Fatalf("expected 1 orphan comment group in file, got %d", len(file.OrphanComments))
	}

	if file.OrphanComments[0].List[0].Text != "// orphan comment at end of file" {
		t.Errorf("orphan comment text = %q", file.OrphanComments[0].List[0].Text)
	}
}

func TestParser_CommentAttachment_LeadingCommentBeforePackage(t *testing.T) {
	input := `// File header comment
// License info
package x

templ Header() {
	<span>Hello</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if file.LeadingComments == nil {
		t.Fatal("expected LeadingComments on file, got nil")
	}

	if len(file.LeadingComments.List) != 2 {
		t.Fatalf("expected 2 comments in file leading group, got %d", len(file.LeadingComments.List))
	}
}

func TestParser_CommentGrouping_BlankLineSeparation(t *testing.T) {
	type tc struct {
		comments   []*Comment
		wantGroups int
	}

	tests := map[string]tc{
		"single comment": {
			comments: []*Comment{
				{Text: "// a", Position: Position{Line: 1}, EndLine: 1},
			},
			wantGroups: 1,
		},
		"adjacent comments": {
			comments: []*Comment{
				{Text: "// a", Position: Position{Line: 1}, EndLine: 1},
				{Text: "// b", Position: Position{Line: 2}, EndLine: 2},
			},
			wantGroups: 1,
		},
		"blank line separation": {
			comments: []*Comment{
				{Text: "// a", Position: Position{Line: 1}, EndLine: 1},
				{Text: "// b", Position: Position{Line: 3}, EndLine: 3}, // line 2 is blank
			},
			wantGroups: 2,
		},
		"multiple groups": {
			comments: []*Comment{
				{Text: "// a", Position: Position{Line: 1}, EndLine: 1},
				{Text: "// b", Position: Position{Line: 2}, EndLine: 2},
				{Text: "// c", Position: Position{Line: 5}, EndLine: 5}, // blank lines
				{Text: "// d", Position: Position{Line: 6}, EndLine: 6},
			},
			wantGroups: 2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			groups := groupComments(tt.comments)
			if len(groups) != tt.wantGroups {
				t.Errorf("got %d groups, want %d", len(groups), tt.wantGroups)
			}
		})
	}
}

