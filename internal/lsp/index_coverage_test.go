package lsp

import (
	"strings"
	"testing"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

const indexCoverageSrc = `package main

templ Card(title string, count int) {
	<div>{title}</div>
}

func format(prefix string, value int) string {
	return prefix
}
`

func newIndexedFixture(t *testing.T) (*ComponentIndex, string) {
	t.Helper()
	uri := "file:///index.gsx"
	lexer := tuigen.NewLexer("index.gsx", indexCoverageSrc)
	parser := tuigen.NewParser(lexer)
	ast, err := parser.ParseFile()
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	idx := NewComponentIndex()
	idx.IndexDocument(uri, ast)
	return idx, uri
}

func TestComponentIndex_Lookups(t *testing.T) {
	idx, uri := newIndexedFixture(t)

	t.Run("ComponentsInFile", func(t *testing.T) {
		names := idx.ComponentsInFile(uri)
		if len(names) != 1 || names[0] != "Card" {
			t.Errorf("ComponentsInFile = %v, want [Card]", names)
		}
		if got := idx.ComponentsInFile("file:///other.gsx"); len(got) != 0 {
			t.Errorf("ComponentsInFile(other) = %v, want empty", got)
		}
	})

	t.Run("GetInfo", func(t *testing.T) {
		info := idx.GetInfo("Card")
		if info == nil || info.Name != "Card" {
			t.Fatalf("GetInfo = %+v, want Card", info)
		}
		if len(info.Params) != 2 {
			t.Errorf("Card params = %d, want 2", len(info.Params))
		}
		if idx.GetInfo("Nope") != nil {
			t.Error("GetInfo(Nope) should be nil")
		}
	})

	t.Run("LookupFunc", func(t *testing.T) {
		info, ok := idx.LookupFunc("format")
		if !ok || info == nil {
			t.Fatal("format not found")
		}
		if info.Signature != "func format(prefix string, value int) string" {
			t.Errorf("Signature = %q", info.Signature)
		}
		if info.Returns != "string" {
			t.Errorf("Returns = %q, want string", info.Returns)
		}
		if _, ok := idx.LookupFunc("missing"); ok {
			t.Error("LookupFunc(missing) should not be found")
		}
	})

	t.Run("LookupParam", func(t *testing.T) {
		info, ok := idx.LookupParam("Card", "title")
		if !ok || info == nil {
			t.Fatal("Card.title not found")
		}
		if info.Type != "string" || info.ComponentName != "Card" {
			t.Errorf("param info = %+v", info)
		}
		if _, ok := idx.LookupParam("Card", "missing"); ok {
			t.Error("LookupParam(Card, missing) should not be found")
		}
	})

	t.Run("LookupParamInAnyComponent", func(t *testing.T) {
		info, ok := idx.LookupParamInAnyComponent("count")
		if !ok || info == nil {
			t.Fatal("count not found in any component")
		}
		if info.ComponentName != "Card" {
			t.Errorf("ComponentName = %q, want Card", info.ComponentName)
		}
		if _, ok := idx.LookupParamInAnyComponent("ghost"); ok {
			t.Error("LookupParamInAnyComponent(ghost) should not be found")
		}
	})

	t.Run("LookupFuncParam", func(t *testing.T) {
		param, gotURI, ok := idx.LookupFuncParam("format", "value")
		if !ok || param == nil {
			t.Fatal("format.value not found")
		}
		if param.Type != "int" {
			t.Errorf("Type = %q, want int", param.Type)
		}
		if gotURI != uri {
			t.Errorf("URI = %q, want %q", gotURI, uri)
		}
		if _, _, ok := idx.LookupFuncParam("nofunc", "value"); ok {
			t.Error("LookupFuncParam(nofunc) should not be found")
		}
		if _, _, ok := idx.LookupFuncParam("format", "noparam"); ok {
			t.Error("LookupFuncParam(format, noparam) should not be found")
		}
	})
}

func TestComponentIndex_Remove(t *testing.T) {
	idx, uri := newIndexedFixture(t)

	idx.Remove(uri)

	if _, ok := idx.Lookup("Card"); ok {
		t.Error("Card still present after Remove")
	}
	if _, ok := idx.LookupFunc("format"); ok {
		t.Error("format still present after Remove")
	}
	if _, ok := idx.LookupParam("Card", "title"); ok {
		t.Error("Card.title still present after Remove")
	}
}

func TestComponentIndex_AddFuncInvalidCode(t *testing.T) {
	idx := NewComponentIndex()
	idx.AddFunc("file:///x.gsx", &tuigen.GoFunc{
		Code:     "var notAFunc = 1",
		Position: tuigen.Position{Line: 1, Column: 1},
	})
	if got := idx.AllFunctions(); len(got) != 0 {
		t.Errorf("AllFunctions = %v, want empty for invalid code", got)
	}
}

func TestParseFuncSignature(t *testing.T) {
	type tc struct {
		code       string
		wantName   string
		wantSig    string
		wantRet    string
		wantParams []FuncParam
	}

	tests := map[string]tc{
		"two params with returns": {
			code:     "func add(a int, b int) int {",
			wantName: "add",
			wantSig:  "func add(a int, b int) int",
			wantRet:  "int",
			wantParams: []FuncParam{
				{Name: "a", Type: "int", Position: Position{Character: 9}},
				{Name: "b", Type: "int", Position: Position{Character: 16}},
			},
		},
		"no params": {
			code:       "func noop() {",
			wantName:   "noop",
			wantSig:    "func noop()",
			wantRet:    "",
			wantParams: nil,
		},
		"variadic": {
			code:     "func join(items ...string) string {",
			wantName: "join",
			wantSig:  "func join(items ...string) string",
			wantRet:  "string",
			wantParams: []FuncParam{
				{Name: "items", Type: "...string", Position: Position{Character: 10}},
			},
		},
		"type only param is not a named param": {
			code:       "func cb(int) {",
			wantName:   "cb",
			wantSig:    "func cb(int)",
			wantRet:    "",
			wantParams: nil,
		},
		"grouped names share a type": {
			code:     "func sum(a, b int) int { return a + b }",
			wantName: "sum",
			wantSig:  "func sum(a, b int) int",
			wantRet:  "int",
			wantParams: []FuncParam{
				{Name: "a", Type: "int", Position: Position{Character: 9}},
				{Name: "b", Type: "int", Position: Position{Character: 12}},
			},
		},
		"mixed grouped and variadic": {
			code:     "func f(a string, b, c int, opts ...tui.Option) {",
			wantName: "f",
			wantSig:  "func f(a string, b, c int, opts ...tui.Option)",
			wantRet:  "",
			wantParams: []FuncParam{
				{Name: "a", Type: "string", Position: Position{Character: 7}},
				{Name: "b", Type: "int", Position: Position{Character: 17}},
				{Name: "c", Type: "int", Position: Position{Character: 20}},
				{Name: "opts", Type: "...tui.Option", Position: Position{Character: 27}},
			},
		},
		"commas nested in types": {
			code:     "func g(m map[string]int, fn func(int, int) bool) bool {",
			wantName: "g",
			wantSig:  "func g(m map[string]int, fn func(int, int) bool) bool",
			wantRet:  "bool",
			wantParams: []FuncParam{
				{Name: "m", Type: "map[string]int", Position: Position{Character: 7}},
				{Name: "fn", Type: "func(int, int) bool", Position: Position{Character: 25}},
			},
		},
		"not a function": {
			code:     "var x = 1",
			wantName: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotName, gotSig, gotParams, gotRet := parseFuncSignature(tt.code)
			if gotName != tt.wantName {
				t.Fatalf("name = %q, want %q", gotName, tt.wantName)
			}
			if tt.wantName == "" {
				return
			}
			if gotSig != tt.wantSig {
				t.Errorf("signature = %q, want %q", gotSig, tt.wantSig)
			}
			if gotRet != tt.wantRet {
				t.Errorf("returns = %q, want %q", gotRet, tt.wantRet)
			}
			if len(gotParams) != len(tt.wantParams) {
				t.Fatalf("params = %+v, want %+v", gotParams, tt.wantParams)
			}
			for i, want := range tt.wantParams {
				got := gotParams[i]
				if got.Name != want.Name || got.Type != want.Type {
					t.Errorf("param[%d] = %+v, want %+v", i, got, want)
				}
				if want.Position.Character != 0 && got.Position.Character != want.Position.Character {
					t.Errorf("param[%d] position = %d, want %d", i, got.Position.Character, want.Position.Character)
				}
			}
		})
	}
}

// The indexed location of a component spans its name so go-to-definition
// highlights "Header", not "templ ".
func TestComponentIndex_LocationSpansName(t *testing.T) {
	idx := NewComponentIndex()
	idx.Add("file:///w/a.gsx", &tuigen.Component{
		Name:     "Header",
		Position: tuigen.Position{Line: 3, Column: 1},
		NamePos:  tuigen.Position{Line: 3, Column: 7},
	})
	info, ok := idx.Lookup("Header")
	if !ok {
		t.Fatal("component not indexed")
	}
	r := info.Location.Range
	if r.Start.Line != 2 || r.Start.Character != 6 || r.End.Line != 2 || r.End.Character != 12 {
		t.Errorf("range = %d:%d..%d:%d, want 2:6..2:12", r.Start.Line, r.Start.Character, r.End.Line, r.End.Character)
	}
}

// Columns are rune-based, so a non-ASCII name must not overshoot by its
// extra UTF-8 bytes.
func TestComponentIndex_LocationSpansUnicodeName(t *testing.T) {
	idx := NewComponentIndex()
	idx.Add("file:///w/a.gsx", &tuigen.Component{
		Name:     "Café",
		Position: tuigen.Position{Line: 1, Column: 1},
		NamePos:  tuigen.Position{Line: 1, Column: 7},
	})
	info, _ := idx.Lookup("Café")
	if got := info.Location.Range.End.Character; got != 10 {
		t.Errorf("end = %d, want 10 (4 runes after column 6)", got)
	}
}

// A func's indexed location spans its name, so @Sidebar(...) highlights
// "Sidebar" rather than "func Sidebar".
func TestComponentIndex_FuncLocationSpansName(t *testing.T) {
	type tc struct {
		code      string
		wantStart int
		wantEnd   int
	}

	tests := map[string]tc{
		"plain func":       {code: "func Sidebar(cat *tui.State[string]) *sidebar {\n\treturn nil\n}", wantStart: 5, wantEnd: 12},
		"extra whitespace": {code: "func   Sidebar(x int) *sidebar {}", wantStart: 7, wantEnd: 14},
		"unicode name":     {code: "func Café() *sidebar {}", wantStart: 5, wantEnd: 9},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			idx := NewComponentIndex()
			idx.AddFunc("file:///w/a.gsx", &tuigen.GoFunc{Code: tt.code, Position: tuigen.Position{Line: 10, Column: 1}})
			fnName := strings.Fields(strings.TrimPrefix(tt.code, "func"))[0]
			fnName = fnName[:strings.IndexAny(fnName, "([")]
			info, ok := idx.LookupFunc(fnName)
			if !ok {
				t.Fatalf("func %q not indexed", fnName)
			}
			r := info.Location.Range
			if r.Start.Line != 9 || r.Start.Character != tt.wantStart || r.End.Character != tt.wantEnd {
				t.Errorf("range = %d:%d..%d, want 9:%d..%d", r.Start.Line, r.Start.Character, r.End.Character, tt.wantStart, tt.wantEnd)
			}
		})
	}
}
