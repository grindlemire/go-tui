package tuigen

import (
	"bytes"
	"strings"
	"testing"
)

// parseFileForTest parses source and fails the test on error.
func parseFileForTest(t *testing.T, source string) *File {
	t.Helper()
	lexer := NewLexer("test.gsx", source)
	parser := NewParser(lexer)
	file, err := parser.ParseFile()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	return file
}

// parseAnalyzeGenerate runs the full production pipeline: parse, analyze
// (which transforms let-binding references into RawGoExpr), then generate
// with SkipImports for speed.
func parseAnalyzeGenerate(t *testing.T, source string) string {
	t.Helper()
	file := parseFileForTest(t, source)
	analyzer := NewAnalyzer()
	if err := analyzer.Analyze(file); err != nil {
		t.Fatalf("analyze failed: %v", err)
	}
	gen := NewGenerator()
	gen.SkipImports = true
	out, err := gen.Generate(file, "test.gsx")
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	return string(out)
}

func TestGenerator_GenerateString(t *testing.T) {
	file := parseFileForTest(t, `package x
templ Empty() {
	<div></div>
}`)

	gen := NewGenerator()
	gen.SkipImports = true
	code, err := gen.GenerateString(file, "test.gsx")
	if err != nil {
		t.Fatalf("GenerateString failed: %v", err)
	}
	if !strings.Contains(code, "func Empty() *EmptyView") {
		t.Errorf("GenerateString output missing component function, got:\n%s", code)
	}
	if !strings.Contains(code, "// Source: test.gsx") {
		t.Errorf("GenerateString output missing source header, got:\n%s", code)
	}
}

func TestGenerator_GenerateToBuffer(t *testing.T) {
	file := parseFileForTest(t, `package x
templ Box() {
	<div></div>
}`)

	gen := NewGenerator()
	gen.SkipImports = true
	var buf bytes.Buffer
	if err := gen.GenerateToBuffer(&buf, file, "test.gsx"); err != nil {
		t.Fatalf("GenerateToBuffer failed: %v", err)
	}
	if !strings.Contains(buf.String(), "func Box() *BoxView") {
		t.Errorf("buffer missing generated function, got:\n%s", buf.String())
	}
}

func TestGenerator_ParseAndGenerate_FullImports(t *testing.T) {
	// ParseAndGenerate uses the goimports pipeline (imports.Process), which
	// also exercises adjustSourceMapForGoimports.
	source := `package x

import (
	"fmt"
)

templ Greeting(name string) {
	<div>
		<span>{fmt.Sprintf("hello %s", name)}</span>
	</div>
}`
	out, err := ParseAndGenerate("test.gsx", source)
	if err != nil {
		t.Fatalf("ParseAndGenerate failed: %v", err)
	}
	code := string(out)
	for _, want := range []string{
		"func Greeting(name string) *GreetingView",
		`fmt.Sprintf("hello %s", name)`,
		`tui "github.com/grindlemire/go-tui"`,
	} {
		if !strings.Contains(code, want) {
			t.Errorf("output missing %q, got:\n%s", want, code)
		}
	}
}

func TestGenerator_GetSourceMap(t *testing.T) {
	file := parseFileForTest(t, `package x

func helper() int {
	return 1
}

templ Box() {
	<div></div>
}`)

	gen := NewGenerator()
	gen.SkipImports = true
	if _, err := gen.Generate(file, "test.gsx"); err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	sm := gen.GetSourceMap()
	if sm == nil {
		t.Fatal("GetSourceMap returned nil")
	}
	if len(sm.Mappings) == 0 {
		t.Fatal("source map has no mappings for passthrough Go code")
	}
}

func TestGenerator_FindFirstContentLineAfterImports(t *testing.T) {
	type tc struct {
		code string
		want int
	}

	tests := map[string]tc{
		"block import": {
			code: "package x\n\nimport (\n\t\"fmt\"\n)\n\nfunc main() {}\n",
			want: 6,
		},
		"single line import": {
			code: "package x\n\nimport \"fmt\"\n\nfunc main() {}\n",
			want: 4,
		},
		"no content after imports": {
			code: "package x\n\nimport \"fmt\"\n\n",
			want: 5, // falls back to line count
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := findFirstContentLineAfterImports([]byte(tt.code))
			if got != tt.want {
				t.Errorf("findFirstContentLineAfterImports = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGenerator_ViewTypeName(t *testing.T) {
	type tc struct {
		component string
		want      string
	}

	tests := map[string]tc{
		"local component":   {component: "Badge", want: "*BadgeView"},
		"package qualified": {component: "pkg.Badge", want: "*pkg.BadgeView"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := viewTypeName(tt.component); got != tt.want {
				t.Errorf("viewTypeName(%q) = %q, want %q", tt.component, got, tt.want)
			}
		})
	}
}

func TestGenerator_TextExpr(t *testing.T) {
	type tc struct {
		code string
		want string
	}

	tests := map[string]tc{
		"integer literal":      {code: "42", want: "fmt.Sprint(42)"},
		"float literal":        {code: "3.14", want: "fmt.Sprint(3.14)"},
		"negative literal":     {code: "-1", want: "fmt.Sprint(-1)"},
		"leading dot float":    {code: ".5", want: "fmt.Sprint(.5)"},
		"hex literal":          {code: "0xFF", want: "fmt.Sprint(0xFF)"},
		"identifier untouched": {code: "name", want: "name"},
		"empty string":         {code: "", want: ""},
		"bare minus":           {code: "-", want: "-"},
		"whitespace only":      {code: "   ", want: "   "},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := textExpr(tt.code); got != tt.want {
				t.Errorf("textExpr(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestGenerator_StateTextBindings(t *testing.T) {
	type tc struct {
		input           string
		wantContains    []string
		wantNotContains []string
	}

	tests := map[string]tc{
		"single state var binds SetText": {
			input: `package x
import "fmt"
templ Counter() {
	count := tui.NewState(0)
	<div>{fmt.Sprintf("count: %d", count.Get())}</div>
}`,
			wantContains: []string{
				"// State bindings",
				"count.Bind(func(_ int) {",
				`__tui_1.SetText(fmt.Sprintf("count: %d", count.Get()))`,
			},
		},
		"multiple state vars share update func": {
			input: `package x
import "fmt"
templ Counter() {
	count := tui.NewState(0)
	name := tui.NewState("hi")
	<div>{fmt.Sprintf("%s: %d", name.Get(), count.Get())}</div>
}`,
			wantContains: []string{
				"__update___tui_1 := func() { __tui_1.SetText(",
				"count.Bind(func(_ int) { __update___tui_1() })",
				"name.Bind(func(_ string) { __update___tui_1() })",
			},
		},
		"dynamic class binding has no setter": {
			input: `package x
templ Styled() {
	count := tui.NewState(0)
	<div class={count.Get()}></div>
}`,
			wantContains: []string{
				"// State bindings",
			},
			wantNotContains: []string{
				".Bind(",
				"SetClass",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := parseAndGenerateSkipImports("test.gsx", tt.input)
			if err != nil {
				t.Fatalf("generation failed: %v", err)
			}
			code := string(output)
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing %q\nGot:\n%s", want, code)
				}
			}
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(code, notWant) {
					t.Errorf("output contains unexpected %q\nGot:\n%s", notWant, code)
				}
			}
		})
	}
}
