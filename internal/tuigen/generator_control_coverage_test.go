package tuigen

import (
	"strings"
	"testing"
)

func TestGenerator_LetBindings(t *testing.T) {
	// Note: cases with a component call or component expr on the RHS use
	// method templs only. Function templs currently panic on such bindings
	// (nil Element dereference in validateRefs / containsChildrenSlot /
	// transformNode), so those paths are only reachable via method templs,
	// which do not run ref collection.
	type tc struct {
		input        string
		useAnalyzer  bool
		wantContains []string
	}

	tests := map[string]tc{
		"element binding with options and children": {
			input: `package x
templ App() {
	card := <div border={tui.BorderSingle}>
		<span>inner</span>
	</div>
	<div>{card}</div>
}`,
			useAnalyzer: true,
			wantContains: []string{
				"card := tui.New(",
				"tui.WithBorder(tui.BorderSingle)",
				`tui.WithText("inner")`,
				"card.AddChild(",
			},
		},
		"element binding with no options": {
			input: `package x
templ App() {
	box := <div></div>
	<div>{box}</div>
}`,
			useAnalyzer: true,
			wantContains: []string{
				"box := tui.New()",
			},
		},
		"function templ call binding extracts Root": {
			input: `package x

type shell struct{}

templ Badge(s string) {
	<span>{s}</span>
}

templ (c *shell) Render() {
	badge := @Badge("hi")
	<div>{badge}</div>
}`,
			wantContains: []string{
				`:= Badge("hi")`,
				"badge := __tui_0.Root",
			},
		},
		"struct mount binding keeps element": {
			input: `package x

type shell struct{}

templ (c *shell) Render() {
	side := @Sidebar()
	<div>{side}</div>
}`,
			wantContains: []string{
				"app.Mount(c, 0, func() tui.Component {",
				"return Sidebar()",
				"side := __tui_0",
			},
		},
		"component expr binding renders with app": {
			input: `package x

type shell struct{}

templ (c *shell) Render() {
	ed := @c.editor
	<div>{ed}</div>
}`,
			wantContains: []string{
				"ed := c.editor.Render(app)",
			},
		},
		"let binding inside element adds to parent": {
			input: `package x
templ App() {
	<div>
		lbl := <span>tag</span>
	</div>
}`,
			useAnalyzer: true,
			wantContains: []string{
				"lbl := tui.New(",
				".AddChild(lbl)",
			},
		},
		"call binding inside element adds to parent": {
			input: `package x

type shell struct{}

templ (c *shell) Render() {
	<div>
		b := @Sidebar("x")
	</div>
}`,
			wantContains: []string{
				"b := __tui_1",
				".AddChild(b)",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var code string
			if tt.useAnalyzer {
				code = parseAnalyzeGenerate(t, tt.input)
			} else {
				output, err := parseAndGenerateSkipImports("test.gsx", tt.input)
				if err != nil {
					t.Fatalf("generation failed: %v", err)
				}
				code = string(output)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing %q\nGot:\n%s", want, code)
				}
			}
		})
	}
}

func TestGenerator_TopLevelForLoop(t *testing.T) {
	input := `package x
templ List(items []string) {
	for _, item := range items {
		<span>{item}</span>
	}
}`

	output, err := parseAndGenerateSkipImports("test.gsx", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}
	code := string(output)
	for _, want := range []string{
		// A synthetic container root wraps the loop output.
		"var __tui_0 *tui.Element",
		"__tui_1 := tui.New()",
		"for __idx_0, item := range items {",
		"__tui_1.AddChild(",
		"if __tui_0 == nil {",
		"__tui_0 = __tui_1",
		"Root:      __tui_0",
	} {
		if !strings.Contains(code, want) {
			t.Errorf("output missing %q\nGot:\n%s", want, code)
		}
	}
}

func TestGenerator_TopLevelControlFlowToRoot(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"top level if else assigns first root": {
			input: `package x
templ App(show bool) {
	if show {
		<div></div>
	} else {
		<span>fallback</span>
	}
}`,
			wantContains: []string{
				"var __tui_0 *tui.Element",
				"if show {",
				"if __tui_0 == nil {",
				"} else {",
				`tui.WithText("fallback")`,
			},
		},
		"top level else if chain": {
			input: `package x
templ App(n int) {
	if n == 0 {
		<div></div>
	} else if n == 1 {
		<span>one</span>
	}
}`,
			wantContains: []string{
				"if n == 0 {",
				"} else if n == 1 {",
				"if __tui_0 == nil {",
			},
		},
		"component call in top level if uses Root": {
			input: `package x
templ App(show bool) {
	if show {
		@Badge("x")
	}
}`,
			wantContains: []string{
				`= Badge("x")`,
				".Root",
				"if __tui_0 == nil {",
			},
		},
		"component expr in top level if renders with app": {
			input: `package x

type shell struct{}

templ (c *shell) Render() {
	if c.open {
		@c.editor
	}
}`,
			wantContains: []string{
				":= c.editor.Render(app)",
				"if __tui_0 == nil {",
			},
		},
		"let binding and go code in top level if": {
			input: `package x
templ App(show bool) {
	if show {
		x := 1
		lbl := <span>{x}</span>
		<div>{lbl}</div>
	}
}`,
			wantContains: []string{
				"x := 1",
				"lbl := tui.New(",
			},
		},
		"nested for inside top level if": {
			input: `package x
templ App(show bool, items []string) {
	if show {
		for _, item := range items {
			<span>{item}</span>
		}
	}
}`,
			wantContains: []string{
				"if show {",
				"for __idx_0, item := range items {",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			code := parseAnalyzeGenerate(t, tt.input)
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing %q\nGot:\n%s", want, code)
				}
			}
		})
	}
}

func TestGenerator_ForLoopBodyNodes(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"named index is preserved": {
			input: `package x
templ App(items []string) {
	<div>
		for i, item := range items {
			<span>{item}</span>
		}
	</div>
}`,
			wantContains: []string{
				"for i, item := range items {",
				"_ = i",
			},
		},
		"let binding in loop body": {
			input: `package x
templ App(items []string) {
	<div>
		for _, item := range items {
			row := <span>{item}</span>
			<div>{row}</div>
		}
	</div>
}`,
			wantContains: []string{
				"row := tui.New(",
				".AddChild(row)",
			},
		},
		"nested loop in loop body": {
			input: `package x
templ App(rows [][]string) {
	<div>
		for _, row := range rows {
			for _, cell := range row {
				<span>{cell}</span>
			}
		}
	</div>
}`,
			wantContains: []string{
				"for __idx_0, row := range rows {",
				"for __idx_1, cell := range row {",
			},
		},
		"go code and bare expr in loop body": {
			input: `package x
templ App(items []string) {
	<div>
		for _, item := range items {
			x := len(item)
			{item + string(rune(x))}
		}
	</div>
}`,
			wantContains: []string{
				"x := len(item)",
				"tui.New(tui.WithText(item + string(rune(x))))",
			},
		},
		"if statement in loop body stays plain": {
			input: `package x
templ App(items []string) {
	<div>
		for _, item := range items {
			if item != "" {
				<span>{item}</span>
			}
		}
	</div>
}`,
			wantContains: []string{
				`if item != "" {`,
			},
		},
		"component call in loop body collects views": {
			input: `package x
templ App(items []string) {
	<div>
		for _, item := range items {
			@Badge(item)
		}
	</div>
}`,
			wantContains: []string{
				"= Badge(item)",
				"_views = append(",
			},
		},
		"children slot in loop body of function templ": {
			input: `package x
templ App(items []string, children []*tui.Element) {
	<div>
		for _, item := range items {
			{children...}
		}
	</div>
}`,
			wantContains: []string{
				"for _, __child := range children {",
				".AddChild(__child)",
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
		})
	}
}
