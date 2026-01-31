package tuigen

import (
	"testing"
)

func TestParser_CommentAttachment_InForLoop(t *testing.T) {
	input := `package x

templ List(items []string) {
	@for _, item := range items { // loop comment
		// comment before span
		<span>{item}</span>
	}
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]
	forLoop, ok := comp.Body[0].(*ForLoop)
	if !ok {
		t.Fatalf("expected *ForLoop, got %T", comp.Body[0])
	}

	// Check trailing comment on opening brace
	if forLoop.TrailingComments == nil {
		t.Fatal("expected TrailingComments on ForLoop, got nil")
	}

	if forLoop.TrailingComments.List[0].Text != "// loop comment" {
		t.Errorf("trailing comment text = %q", forLoop.TrailingComments.List[0].Text)
	}

	// Check leading comment on span
	if len(forLoop.Body) != 1 {
		t.Fatalf("expected 1 body node, got %d", len(forLoop.Body))
	}

	span, ok := forLoop.Body[0].(*Element)
	if !ok {
		t.Fatalf("expected *Element, got %T", forLoop.Body[0])
	}

	if span.LeadingComments == nil {
		t.Fatal("expected LeadingComments on span, got nil")
	}

	if span.LeadingComments.List[0].Text != "// comment before span" {
		t.Errorf("leading comment text = %q", span.LeadingComments.List[0].Text)
	}
}

func TestParser_CommentAttachment_InIfStmt(t *testing.T) {
	input := `package x

templ Cond(show bool) {
	@if show { // if comment
		// comment before visible
		<span>Visible</span>
	}
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]
	ifStmt, ok := comp.Body[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected *IfStmt, got %T", comp.Body[0])
	}

	// Check trailing comment on opening brace
	if ifStmt.TrailingComments == nil {
		t.Fatal("expected TrailingComments on IfStmt, got nil")
	}

	if ifStmt.TrailingComments.List[0].Text != "// if comment" {
		t.Errorf("trailing comment text = %q", ifStmt.TrailingComments.List[0].Text)
	}

	// Check leading comment on span
	if len(ifStmt.Then) != 1 {
		t.Fatalf("expected 1 then node, got %d", len(ifStmt.Then))
	}

	span, ok := ifStmt.Then[0].(*Element)
	if !ok {
		t.Fatalf("expected *Element, got %T", ifStmt.Then[0])
	}

	if span.LeadingComments == nil {
		t.Fatal("expected LeadingComments on span, got nil")
	}
}

func TestParser_CommentAttachment_EmptyForLoopBody(t *testing.T) {
	input := `package x

templ Empty() {
	@for _, item := range items {
		// only a comment, no elements
	}
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]
	forLoop, ok := comp.Body[0].(*ForLoop)
	if !ok {
		t.Fatalf("expected *ForLoop, got %T", comp.Body[0])
	}

	// Comment in empty body should be an orphan
	if len(forLoop.OrphanComments) != 1 {
		t.Fatalf("expected 1 orphan comment group, got %d", len(forLoop.OrphanComments))
	}

	if forLoop.OrphanComments[0].List[0].Text != "// only a comment, no elements" {
		t.Errorf("orphan comment text = %q", forLoop.OrphanComments[0].List[0].Text)
	}
}

func TestParser_CommentAttachment_TrailingOnElement(t *testing.T) {
	input := `package x

templ Test() {
	<span>Hello</span>  // trailing on span
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]
	elem, ok := comp.Body[0].(*Element)
	if !ok {
		t.Fatalf("expected *Element, got %T", comp.Body[0])
	}

	if elem.TrailingComments == nil {
		t.Fatal("expected TrailingComments on element, got nil")
	}

	if elem.TrailingComments.List[0].Text != "// trailing on span" {
		t.Errorf("trailing comment text = %q", elem.TrailingComments.List[0].Text)
	}
}

func TestParser_CommentAttachment_SelfClosingElement(t *testing.T) {
	input := `package x

templ Test() {
	<input />  // trailing on self-closing
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]
	elem, ok := comp.Body[0].(*Element)
	if !ok {
		t.Fatalf("expected *Element, got %T", comp.Body[0])
	}

	if !elem.SelfClose {
		t.Error("expected SelfClose to be true")
	}

	if elem.TrailingComments == nil {
		t.Fatal("expected TrailingComments on self-closing element, got nil")
	}

	if elem.TrailingComments.List[0].Text != "// trailing on self-closing" {
		t.Errorf("trailing comment text = %q", elem.TrailingComments.List[0].Text)
	}
}

func TestParser_CommentAttachment_LeadingOnGoFunc(t *testing.T) {
	input := `package x

// Helper function comment
func helper() string {
	return "hello"
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Funcs) != 1 {
		t.Fatalf("expected 1 func, got %d", len(file.Funcs))
	}

	fn := file.Funcs[0]
	if fn.LeadingComments == nil {
		t.Fatal("expected LeadingComments on func, got nil")
	}

	if fn.LeadingComments.List[0].Text != "// Helper function comment" {
		t.Errorf("leading comment text = %q", fn.LeadingComments.List[0].Text)
	}
}

func TestParser_CommentAttachment_BlockComment(t *testing.T) {
	input := `package x

/* Block comment
   spanning multiple lines */
templ Header() {
	<span>Hello</span>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]
	if comp.LeadingComments == nil {
		t.Fatal("expected LeadingComments, got nil")
	}

	if len(comp.LeadingComments.List) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comp.LeadingComments.List))
	}

	comment := comp.LeadingComments.List[0]
	if !comment.IsBlock {
		t.Error("expected block comment")
	}
}

func TestParser_CommentAttachment_NestedElements(t *testing.T) {
	input := `package x

templ Nested() {
	<div>
		// comment before inner span
		<span>Hello</span>
	</div>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp := file.Components[0]
	div, ok := comp.Body[0].(*Element)
	if !ok {
		t.Fatalf("expected *Element, got %T", comp.Body[0])
	}

	if len(div.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(div.Children))
	}

	span, ok := div.Children[0].(*Element)
	if !ok {
		t.Fatalf("expected *Element child, got %T", div.Children[0])
	}

	if span.LeadingComments == nil {
		t.Fatal("expected LeadingComments on nested span, got nil")
	}

	if span.LeadingComments.List[0].Text != "// comment before inner span" {
		t.Errorf("leading comment text = %q", span.LeadingComments.List[0].Text)
	}
}
