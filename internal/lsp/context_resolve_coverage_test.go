package lsp

import (
	"strings"
	"testing"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

// resolveCoverageSrc is a .gsx fixture exercising imports, top-level decls,
// loops, conditionals, refs, component calls, and helper functions.
const resolveCoverageSrc = `package test

import (
	"fmt"
	f "strings"
)

var topLevel = 1

templ List(items []string) {
	<div class="flex-col">
		for _, item := range items {
			<span ref={rows} key={item}>{item}</span>
			<li ref={cells}>cell</li>
		}
		if len(items) > 0 {
			<span ref={first}>yes</span>
		} else {
			<span>empty</span>
		}
		<p>after</p>
	</div>
}

templ App() {
	<div>
		@List(names)
	</div>
}

func helper(s string, m map[string]int) string {
	return f.ToUpper(s)
}
`

// lineCharOf returns the 0-indexed line and character of the first occurrence
// of needle on the given 0-indexed line of src.
func lineCharOf(t *testing.T, src string, line int, needle string) (int, int) {
	t.Helper()
	lines := strings.Split(src, "\n")
	if line >= len(lines) {
		t.Fatalf("line %d out of range", line)
	}
	idx := strings.Index(lines[line], needle)
	if idx < 0 {
		t.Fatalf("needle %q not found on line %d (%q)", needle, line, lines[line])
	}
	return line, idx
}

func TestResolveFromAST_Coverage(t *testing.T) {
	type tc struct {
		line       int    // 0-indexed line to search
		needle     string // substring on that line to place the cursor on
		charOffset int    // extra character offset from the needle start
		wantKind   NodeKind
		wantWord   string // skip check when empty
		wantImport string // expected ImportPath, skip when empty
		wantGoExpr bool
		inForLoop  bool
		inIfStmt   bool
	}

	tests := map[string]tc{
		"import path": {
			line: 3, needle: `"fmt"`, charOffset: 2,
			wantKind: NodeKindImportPath, wantImport: "fmt",
		},
		"aliased import alias": {
			line: 4, needle: `f "strings"`, charOffset: 0,
			wantKind: NodeKindImportPath, wantImport: "strings",
		},
		"top level decl": {
			line: 7, needle: "topLevel", charOffset: 0,
			wantKind: NodeKindGoDecl,
		},
		"component decl param type": {
			line: 9, needle: "[]string", charOffset: 2,
			wantKind: NodeKindComponent,
		},
		"element inside for loop": {
			line: 12, needle: "span", charOffset: 1,
			wantKind: NodeKindElement, inForLoop: true,
		},
		"element in if then branch": {
			line: 16, needle: "span", charOffset: 1,
			wantKind: NodeKindElement, inIfStmt: true,
		},
		"element in else branch": {
			line: 18, needle: "span", charOffset: 1,
			wantKind: NodeKindElement, inIfStmt: true,
		},
		"element after loop and if": {
			line: 20, needle: "<p>", charOffset: 1,
			wantKind: NodeKindElement,
		},
		// The cursor sits past the single-line {item} expression, so the AST
		// walk rejects it on column range and text heuristics classify the
		// position (still inside the component braces) as a Go expression.
		"go expr column out of range falls through": {
			line: 12, needle: "</span>", charOffset: 2,
			wantKind: NodeKindGoExpr,
		},
		"component call argument area": {
			line: 26, needle: "names", charOffset: 1,
			wantKind: NodeKindGoExpr, wantGoExpr: true,
		},
		"func param name": {
			line: 30, needle: "s string", charOffset: 0,
			wantKind: NodeKindParameter,
		},
		"func param with nested map type": {
			line: 30, needle: "m map[string]int", charOffset: 0,
			wantKind: NodeKindParameter,
		},
		"func body": {
			line: 31, needle: "ToUpper", charOffset: 0,
			wantKind: NodeKindFunction, wantGoExpr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			doc := parseTestDoc(resolveCoverageSrc)
			if doc.AST == nil {
				t.Fatal("fixture failed to parse")
			}
			line, char := lineCharOf(t, resolveCoverageSrc, tt.line, tt.needle)
			ctx := ResolveCursorContext(doc, Position{Line: line, Character: char + tt.charOffset})

			if ctx.NodeKind != tt.wantKind {
				t.Errorf("NodeKind = %s, want %s", ctx.NodeKind, tt.wantKind)
			}
			if tt.wantWord != "" && ctx.Word != tt.wantWord {
				t.Errorf("Word = %q, want %q", ctx.Word, tt.wantWord)
			}
			if tt.wantImport != "" && ctx.ImportPath != tt.wantImport {
				t.Errorf("ImportPath = %q, want %q", ctx.ImportPath, tt.wantImport)
			}
			if tt.wantGoExpr && !ctx.InGoExpr {
				t.Error("expected InGoExpr to be true")
			}
			if tt.inForLoop && ctx.Scope.ForLoop == nil {
				t.Error("expected Scope.ForLoop to be set")
			}
			if tt.inIfStmt && ctx.Scope.IfStmt == nil {
				t.Error("expected Scope.IfStmt to be set")
			}
		})
	}
}

func TestResolveFromAST_CursorAfterComponentEnd(t *testing.T) {
	// Line 29 is the blank line between App's closing brace and func helper.
	doc := parseTestDoc(resolveCoverageSrc)
	if doc.AST == nil {
		t.Fatal("fixture failed to parse")
	}
	ctx := ResolveCursorContext(doc, Position{Line: 29, Character: 0})
	if ctx.NodeKind != NodeKindUnknown {
		t.Errorf("NodeKind = %s, want Unknown", ctx.NodeKind)
	}
	// The blank line is past App's closing brace, so no component scope applies.
	if ctx.Scope.Component != nil {
		t.Errorf("Scope.Component = %v, want nil", ctx.Scope.Component.Name)
	}
}

func TestCollectScope_RefKinds(t *testing.T) {
	doc := parseTestDoc(resolveCoverageSrc)
	if doc.AST == nil {
		t.Fatal("fixture failed to parse")
	}
	// Resolve any position inside the List component to collect scope refs.
	line, char := lineCharOf(t, resolveCoverageSrc, 20, "<p>")
	ctx := ResolveCursorContext(doc, Position{Line: line, Character: char + 1})

	refs := map[string]tuigen.RefInfo{}
	for _, r := range ctx.Scope.Refs {
		refs[r.Name] = r
	}

	rows, ok := refs["rows"]
	if !ok {
		t.Fatal("missing ref 'rows'")
	}
	if !rows.InLoop {
		t.Error("rows: expected InLoop")
	}
	if rows.RefKind != tuigen.RefMap {
		t.Errorf("rows: RefKind = %v, want RefMap", rows.RefKind)
	}
	if rows.KeyExpr != "item" {
		t.Errorf("rows: KeyExpr = %q, want %q", rows.KeyExpr, "item")
	}

	cells, ok := refs["cells"]
	if !ok {
		t.Fatal("missing ref 'cells'")
	}
	if cells.RefKind != tuigen.RefList {
		t.Errorf("cells: RefKind = %v, want RefList", cells.RefKind)
	}

	first, ok := refs["first"]
	if !ok {
		t.Fatal("missing ref 'first'")
	}
	if !first.InConditional {
		t.Error("first: expected InConditional")
	}
	if first.RefKind != tuigen.RefSingle {
		t.Errorf("first: RefKind = %v, want RefSingle", first.RefKind)
	}
}

func TestResolveInComponentCall_Children(t *testing.T) {
	src := `package test

templ Card(title string) {
	<div>{title}</div>
}

templ App() {
	<div>
		@Card("t") {
			<span>inner</span>
		}
	</div>
}
`
	doc := parseTestDoc(src)
	if doc.AST == nil {
		t.Fatal("fixture failed to parse")
	}
	line, char := lineCharOf(t, src, 9, "span")
	ctx := ResolveCursorContext(doc, Position{Line: line, Character: char + 1})
	if ctx.NodeKind != NodeKindElement {
		t.Errorf("NodeKind = %s, want Element", ctx.NodeKind)
	}
}

func TestResolveHelpers_NilGuards(t *testing.T) {
	ctx := &CursorContext{Scope: &Scope{}, Document: &Document{}}

	if resolveInElement(ctx, nil, 1, 1) {
		t.Error("resolveInElement(nil) = true, want false")
	}
	if resolveInForLoop(ctx, nil, 1, 1) {
		t.Error("resolveInForLoop(nil) = true, want false")
	}
	if resolveInIfStmt(ctx, nil, 1, 1) {
		t.Error("resolveInIfStmt(nil) = true, want false")
	}
	if resolveInLetBinding(ctx, nil, 1, 1) {
		t.Error("resolveInLetBinding(nil) = true, want false")
	}
	if resolveInComponentCall(ctx, nil, 1, 1) {
		t.Error("resolveInComponentCall(nil) = true, want false")
	}
	if got := classifyGoExpr(nil); got != NodeKindGoExpr {
		t.Errorf("classifyGoExpr(nil) = %s, want GoExpr", got)
	}
	if got := classifyGoCode(nil); got != NodeKindGoExpr {
		t.Errorf("classifyGoCode(nil) = %s, want GoExpr", got)
	}
}

func TestClassifyGoCode(t *testing.T) {
	type tc struct {
		code string
		want NodeKind
	}

	tests := map[string]tc{
		"state declaration": {code: "count := tui.NewState(0)", want: NodeKindStateDecl},
		"state set":         {code: "count.Set(1)", want: NodeKindStateAccess},
		"state get":         {code: "x := count.Get()", want: NodeKindStateAccess},
		"plain code":        {code: "y := 1", want: NodeKindGoExpr},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := classifyGoCode(&tuigen.GoCode{Code: tt.code})
			if got != tt.want {
				t.Errorf("classifyGoCode(%q) = %s, want %s", tt.code, got, tt.want)
			}
		})
	}
}

func TestResolveInNode_GoCodeColumnOutOfRange(t *testing.T) {
	ctx := &CursorContext{Scope: &Scope{}, Document: &Document{}}
	code := &tuigen.GoCode{
		Code:     "x := 1",
		Position: tuigen.Position{Line: 1, Column: 2},
	}
	// Column 50 is past the end of the single-line code block.
	if resolveInNode(ctx, code, 1, 50) {
		t.Error("expected out-of-range column to not resolve")
	}
}

func TestResolveInLetBinding_NoElementNoNameHit(t *testing.T) {
	ctx := &CursorContext{Scope: &Scope{}, Document: &Document{}}
	let := &tuigen.LetBinding{
		Name:     "label",
		Position: tuigen.Position{Line: 1, Column: 1},
	}
	// Cursor on a different line with no element attached.
	if resolveInLetBinding(ctx, let, 2, 1) {
		t.Error("expected no match for let binding without element")
	}
}

func TestClassifyFromText_NoAST(t *testing.T) {
	type tc struct {
		content string
		needle  string
		want    NodeKind
	}

	tests := map[string]tc{
		"keyword": {
			content: "for x := range items",
			needle:  "for",
			want:    NodeKindKeyword,
		},
		"element tag": {
			content: `<div class="p-1">`,
			needle:  "div",
			want:    NodeKindElement,
		},
		"component call prefix": {
			content: "@Header",
			needle:  "@Header",
			want:    NodeKindComponentCall,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			doc := &Document{URI: "file:///t.gsx", Content: tt.content}
			idx := strings.Index(tt.content, tt.needle)
			if idx < 0 {
				t.Fatalf("needle not found")
			}
			ctx := ResolveCursorContext(doc, Position{Line: 0, Character: idx + 1})
			if ctx.NodeKind != tt.want {
				t.Errorf("NodeKind = %s, want %s", ctx.NodeKind, tt.want)
			}
		})
	}
}

func TestIsOffsetInGoExpr_ClosedBraces(t *testing.T) {
	// Offset after a balanced {x} pair should not be inside a Go expression.
	content := "{x} y"
	if isOffsetInGoExpr(content, 4) {
		t.Error("offset after balanced braces should not be in Go expr")
	}
	// Offset inside nested braces is in a Go expression.
	content2 := "{a{b}c}"
	if !isOffsetInGoExpr(content2, 5) {
		t.Error("offset inside braces should be in Go expr")
	}
}

func TestFindComponentEndLine_Unbalanced(t *testing.T) {
	content := "templ X() {\n<div>"
	comp := &tuigen.Component{Position: tuigen.Position{Line: 1, Column: 1}}
	if got := findComponentEndLine(content, comp); got != 1 {
		t.Errorf("findComponentEndLine = %d, want 1 (last line)", got)
	}
}

func TestFindFuncParamAtColumn(t *testing.T) {
	type tc struct {
		code string
		col  int
		want string
	}

	tests := map[string]tc{
		"not a func":   {code: "var x = 1", col: 1, want: ""},
		"no paren":     {code: "func f", col: 1, want: ""},
		"unbalanced":   {code: "func f(a int", col: 8, want: ""},
		"first param":  {code: "func f(a int, b string) {}", col: 8, want: "a"},
		"second param": {code: "func f(a int, b string) {}", col: 15, want: "b"},
		"nested types": {code: "func g(m map[string]int, fn func(int) bool) {}", col: 26, want: "fn"},
		"miss":         {code: "func f(a int) {}", col: 11, want: ""},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fn := &tuigen.GoFunc{
				Code:     tt.code,
				Position: tuigen.Position{Line: 1, Column: 1},
			}
			got := findFuncParamAtColumn(fn, tt.col)
			if got != tt.want {
				t.Errorf("findFuncParamAtColumn(%q, %d) = %q, want %q", tt.code, tt.col, got, tt.want)
			}
		})
	}
}
