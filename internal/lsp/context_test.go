package lsp

import (
	"testing"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

// parseTestDoc parses a .gsx source string into a Document for testing.
func parseTestDoc(src string) *Document {
	doc := &Document{
		URI:     "file:///test.gsx",
		Content: src,
		Version: 1,
	}
	lexer := tuigen.NewLexer("test.gsx", src)
	parser := tuigen.NewParser(lexer)
	ast, _ := parser.ParseFile()
	doc.AST = ast
	return doc
}

func TestResolveCursorContext_ComponentName(t *testing.T) {
	src := `package test

templ Header(title string) {
	<div>{title}</div>
}
`
	doc := parseTestDoc(src)

	// Cursor on "Header" (line 2, within the name)
	// Component position is at "templ" (tuigen col 1). nameEnd = 1 + len("Header") = 7.
	// The resolver checks tuigen col >= nameStart(1) && col <= nameEnd(7).
	// LSP Character 0 maps to tuigen col 1. So Character 0-6 should match.
	ctx := ResolveCursorContext(doc, Position{Line: 2, Character: 3})

	if ctx.NodeKind != NodeKindComponent {
		t.Errorf("expected NodeKindComponent, got %s", ctx.NodeKind)
	}
	if ctx.Scope.Component == nil {
		t.Fatal("expected Scope.Component to be set")
	}
	if ctx.Scope.Component.Name != "Header" {
		t.Errorf("expected component name 'Header', got %q", ctx.Scope.Component.Name)
	}
}

func TestResolveCursorContext_Parameter(t *testing.T) {
	src := `package test

templ Header(title string) {
	<div>{title}</div>
}
`
	doc := parseTestDoc(src)

	// Cursor on "title" parameter — "templ Header(" is 14 chars, "title" starts at col 14
	ctx := ResolveCursorContext(doc, Position{Line: 2, Character: 14})

	if ctx.NodeKind != NodeKindParameter {
		t.Errorf("expected NodeKindParameter, got %s", ctx.NodeKind)
	}
}

func TestResolveCursorContext_ElementTag(t *testing.T) {
	src := `package test

templ Header(title string) {
	<div>{title}</div>
}
`
	doc := parseTestDoc(src)

	// Cursor on "div" tag (line 3, the <div>)
	// "\t<div>" — tab is col 0, '<' is col 1, 'div' starts at col 2
	ctx := ResolveCursorContext(doc, Position{Line: 3, Character: 2})

	if ctx.NodeKind != NodeKindElement {
		t.Errorf("expected NodeKindElement, got %s", ctx.NodeKind)
	}
}

func TestResolveCursorContext_Attribute(t *testing.T) {
	src := `package test

templ Header(title string) {
	<div class="p-1">{title}</div>
}
`
	doc := parseTestDoc(src)

	// Cursor on "class" attribute — "\t<div class=..." — '<' is col 1, 'div' at col 2, space at col 5, 'class' starts at col 6
	ctx := ResolveCursorContext(doc, Position{Line: 3, Character: 7})

	if ctx.NodeKind != NodeKindAttribute {
		t.Errorf("expected NodeKindAttribute, got %s", ctx.NodeKind)
	}
	if ctx.AttrTag != "div" {
		t.Errorf("expected AttrTag 'div', got %q", ctx.AttrTag)
	}
	if ctx.AttrName != "class" {
		t.Errorf("expected AttrName 'class', got %q", ctx.AttrName)
	}
}

func TestResolveCursorContext_EventHandler(t *testing.T) {
	src := `package test

templ Button(label string) {
	<button onClick={handleClick}>{label}</button>
}
`
	doc := parseTestDoc(src)

	// Cursor on "onClick" — "\t<button onClick=..." — 'onClick' starts after '<button '
	// '<' is col 1, 'button' is col 2-7, space at col 8, 'onClick' starts at col 9
	ctx := ResolveCursorContext(doc, Position{Line: 3, Character: 10})

	if ctx.NodeKind != NodeKindEventHandler {
		t.Errorf("expected NodeKindEventHandler, got %s", ctx.NodeKind)
	}
}

func TestResolveCursorContext_NamedRef(t *testing.T) {
	src := `package test

templ Layout() {
	<div #Header class="p-1">content</div>
}
`
	doc := parseTestDoc(src)

	// Cursor on "#Header" — "\t<div #Header ..." — '#' appears after '<div '
	// We need to find the exact column. '#' char in the line.
	line := getLineText(doc.Content, 3)
	hashIdx := -1
	for i := range line {
		if line[i] == '#' {
			hashIdx = i
			break
		}
	}
	if hashIdx < 0 {
		t.Fatal("could not find # in line")
	}

	ctx := ResolveCursorContext(doc, Position{Line: 3, Character: hashIdx + 1})

	if ctx.NodeKind != NodeKindNamedRef {
		t.Errorf("expected NodeKindNamedRef, got %s", ctx.NodeKind)
	}
}

func TestResolveCursorContext_NamedRef_Multiline(t *testing.T) {
	src := `package test

templ Layout() {
	<div
		#Header
		class="p-1">content</div>
}
`
	doc := parseTestDoc(src)

	// Cursor on "#Header" on its own line (line 4)
	line := getLineText(doc.Content, 4)
	hashIdx := -1
	for i := range line {
		if line[i] == '#' {
			hashIdx = i
			break
		}
	}
	if hashIdx < 0 {
		t.Fatal("could not find # in line")
	}

	ctx := ResolveCursorContext(doc, Position{Line: 4, Character: hashIdx + 1})

	if ctx.NodeKind != NodeKindNamedRef {
		t.Errorf("expected NodeKindNamedRef, got %s", ctx.NodeKind)
	}
}
