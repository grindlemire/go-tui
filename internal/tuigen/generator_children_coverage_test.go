package tuigen

import (
	"strings"
	"testing"
)

func TestGenerator_ComponentCallChildKinds(t *testing.T) {
	type tc struct {
		input           string
		wantContains    []string
		wantNotContains []string
	}

	tests := map[string]tc{
		"element child appended to children slice": {
			input: `package x
templ App() {
	<div>
		@Card() {
			<div></div>
		}
	</div>
}`,
			wantContains: []string{
				"_children := []*tui.Element{}",
				"_children = append(",
				":= Card(",
			},
		},
		"string literal child wrapped in text element": {
			input: `package x
templ App() {
	<div>
		@Card() {
			{"hello world"}
		}
	</div>
}`,
			wantContains: []string{
				`tui.New(tui.WithText("hello world"))`,
				"_children = append(",
			},
		},
		"go expr child wrapped in text element": {
			input: `package x
templ App(msg string) {
	<div>
		@Card() {
			{msg}
		}
	</div>
}`,
			wantContains: []string{
				"tui.New(tui.WithText(msg))",
				"_children = append(",
			},
		},
		"nested component call child appends Root": {
			input: `package x
templ App() {
	<div>
		@Card() {
			@Badge("a")
		}
	</div>
}`,
			wantContains: []string{
				`:= Badge("a")`,
				".Root)",
			},
		},
		"let binding child appended by name": {
			input: `package x
templ App() {
	<div>
		@Card() {
			lbl := <span>hi</span>
		}
	</div>
}`,
			wantContains: []string{
				"lbl := tui.New(",
				", lbl)",
			},
		},
		"for loop child uses slice append": {
			input: `package x
templ App(items []string) {
	<div>
		@Card() {
			for i, item := range items {
				<span>{item}</span>
			}
		}
	</div>
}`,
			wantContains: []string{
				"for i, item := range items {",
				"_ = i",
				"tui.New(",
				"_children = append(",
			},
		},
		"for loop child with underscore index gets synthetic index": {
			input: `package x
templ App(items []string) {
	<div>
		@Card() {
			for _, item := range items {
				<span>{item}</span>
			}
		}
	</div>
}`,
			wantContains: []string{
				"for __idx_0, item := range items {",
				"_ = __idx_0",
			},
		},
		"if else child appends conditionally": {
			input: `package x
templ App(show bool) {
	<div>
		@Card() {
			if show {
				<span>yes</span>
			} else {
				<span>no</span>
			}
		}
	</div>
}`,
			wantContains: []string{
				"if show {",
				"} else {",
				`tui.WithText("yes")`,
				`tui.WithText("no")`,
				"_children = append(",
			},
		},
		"else if chain in component children": {
			input: `package x
templ App(n int) {
	<div>
		@Card() {
			if n == 0 {
				<span>zero</span>
			} else if n == 1 {
				<span>one</span>
			} else {
				<span>many</span>
			}
		}
	</div>
}`,
			wantContains: []string{
				"if n == 0 {",
				"} else if n == 1 {",
				`tui.WithText("many")`,
			},
		},
		"nested control flow inside component children": {
			input: `package x
templ App(items []string, show bool) {
	<div>
		@Card() {
			for _, item := range items {
				if show {
					<span>{item}</span>
				}
			}
			if show {
				for _, item := range items {
					<span>{item}</span>
				}
			}
		}
	</div>
}`,
			wantContains: []string{
				"for __idx_0, item := range items {",
				"if show {",
				"_children = append(",
			},
		},
		"component call and let binding inside if child": {
			input: `package x
templ App(show bool) {
	<div>
		@Card() {
			if show {
				@Badge("b")
				lbl := <span>x</span>
			} else {
				@Badge("c")
				other := <span>y</span>
			}
		}
	</div>
}`,
			wantContains: []string{
				// Conditional component calls are hoisted, so they use
				// assignment rather than short declaration.
				`= Badge("b")`,
				`= Badge("c")`,
				", lbl)",
				", other)",
			},
		},
		"component call inside for loop child appends Root": {
			input: `package x
templ App(items []string) {
	<div>
		@Card() {
			for _, item := range items {
				@Badge(item)
			}
		}
	</div>
}`,
			wantContains: []string{
				"= Badge(item)",
				".Root)",
			},
		},
		"go expr inside for loop child": {
			input: `package x
templ App(items []string) {
	<div>
		@Card() {
			for _, item := range items {
				{item}
			}
		}
	</div>
}`,
			wantContains: []string{
				"tui.New(tui.WithText(item))",
			},
		},
		"call with args and children combines both": {
			input: `package x
templ App(title string) {
	<div>
		@Card(title) {
			<span>body</span>
		}
	</div>
}`,
			wantContains: []string{
				"Card(title, __tui_",
				"_children)",
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

func TestGenerator_StructMountChildren(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"struct mount with element and text children": {
			input: `package x

type shell struct{}

templ (c *shell) Render() {
	<div>
		@Sidebar() {
			<span>nav</span>
			{"plain text"}
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, ",
				"return Sidebar(__tui_",
				"_children)",
				`tui.New(tui.WithText("plain text"))`,
			},
		},
		"struct mount with for and if children": {
			input: `package x

type shell struct{}

templ (c *shell) Render() {
	<div>
		@Sidebar() {
			for _, item := range c.items {
				<span>{item}</span>
			}
			if c.open {
				<span>open</span>
			}
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, ",
				"for __idx_0, item := range c.items {",
				"if c.open {",
				"_children = append(",
			},
		},
		"struct mount with nested call and let binding children": {
			input: `package x

type shell struct{}

templ (c *shell) Render() {
	<div>
		@Sidebar() {
			@Footer()
			lbl := <span>tag</span>
			{c.title}
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, ",
				"return Footer()",
				", lbl)",
				"tui.New(tui.WithText(c.title))",
			},
		},
		"struct mount with args appends children to args": {
			input: `package x

type shell struct{}

templ (c *shell) Render() {
	<div>
		@Sidebar(c.title) {
			<span>nav</span>
		}
	</div>
}`,
			wantContains: []string{
				"return Sidebar(c.title, __tui_",
			},
		},
		"struct mount inside for loop uses runtime index expr": {
			input: `package x

type shell struct{}

templ (c *shell) Render() {
	<div>
		for i, item := range c.items {
			@Sidebar(item)
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, (1)*1000000+i, func() tui.Component {",
				"return Sidebar(item)",
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

func TestGenerator_ComponentExpr(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"component expr as element child": {
			input: `package x

type shell struct{}

templ (c *shell) Render() {
	<div>
		@c.editor
	</div>
}`,
			wantContains: []string{
				":= c.editor.Render(app)",
				".AddChild(__tui_",
			},
		},
		"component expr inside if statement": {
			input: `package x

type shell struct{}

templ (c *shell) Render() {
	<div>
		if c.open {
			@c.editor
		}
	</div>
}`,
			wantContains: []string{
				"if c.open {",
				":= c.editor.Render(app)",
			},
		},
		"component expr inside for loop": {
			input: `package x

type shell struct{}

templ (c *shell) Render() {
	<div>
		for _, p := range c.panes {
			@c.editor
		}
	</div>
}`,
			wantContains: []string{
				"for __idx_0, p := range c.panes {",
				":= c.editor.Render(app)",
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

func TestGenerator_ChildrenSlotInMethodTempl(t *testing.T) {
	input := `package x

type shell struct {
	children []*tui.Element
}

templ (c *shell) Render() {
	<div>
		{children...}
	</div>
}`

	output, err := parseAndGenerateSkipImports("test.gsx", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}
	code := string(output)
	if !strings.Contains(code, "for _, __child := range c.children {") {
		t.Errorf("method templ children slot should range over receiver children\nGot:\n%s", code)
	}
	if !strings.Contains(code, ".AddChild(__child)") {
		t.Errorf("children slot should add each child\nGot:\n%s", code)
	}
}

func TestGenerator_RawGoExprChildrenViaAnalyzer(t *testing.T) {
	// The analyzer transforms {label} references to let bindings into RawGoExpr
	// nodes, which the generator adds directly instead of wrapping in a text
	// element.
	type tc struct {
		input           string
		wantContains    []string
		wantNotContains []string
	}

	tests := map[string]tc{
		"let binding referenced as element child": {
			input: `package x
templ App() {
	label := <span>hi</span>
	<div>{label}</div>
}`,
			wantContains: []string{
				"label := tui.New(",
				".AddChild(label)",
			},
			wantNotContains: []string{
				"tui.New(tui.WithText(label))",
			},
		},
		"let binding referenced inside component call children": {
			input: `package x
templ Card(children []*tui.Element) {
	<div>{children...}</div>
}

templ App() {
	label := <span>hi</span>
	<div>
		@Card() {
			{label}
		}
	</div>
}`,
			wantContains: []string{
				"_children = append(",
				", label)",
			},
		},
		"let binding referenced inside if body": {
			input: `package x
templ App(show bool) {
	label := <span>hi</span>
	<div>
		if show {
			{label}
		}
	</div>
}`,
			wantContains: []string{
				"if show {",
				".AddChild(label)",
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
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(code, notWant) {
					t.Errorf("output contains unexpected %q\nGot:\n%s", notWant, code)
				}
			}
		})
	}
}

func TestGenerator_BodyNodesInsideIf(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"go expr inside if": {
			input: `package x
templ App(show bool, msg string) {
	<div>
		if show {
			{msg}
		}
	</div>
}`,
			wantContains: []string{
				"tui.New(tui.WithText(msg))",
			},
		},
		"component call inside if": {
			input: `package x
templ App(show bool) {
	<div>
		if show {
			@Badge("x")
		}
	</div>
}`,
			wantContains: []string{
				`= Badge("x")`,
				".AddChild(",
			},
		},
		"let binding inside if": {
			input: `package x
templ App(show bool) {
	<div>
		if show {
			lbl := <span>a</span>
		}
	</div>
}`,
			wantContains: []string{
				"lbl := tui.New(",
			},
		},
		"nested for inside if": {
			input: `package x
templ App(show bool, items []string) {
	<div>
		if show {
			for _, item := range items {
				<span>{item}</span>
			}
		}
	</div>
}`,
			wantContains: []string{
				"if show {",
				"for __idx_0, item := range items {",
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
