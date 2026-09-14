package tuigen

import (
	"strings"
	"testing"
)

// assertOrdered checks that each want string appears in code after the previous one.
func assertOrdered(t *testing.T, code string, wants []string) {
	t.Helper()
	last := 0
	for _, want := range wants {
		idx := strings.Index(code[last:], want)
		if idx < 0 {
			t.Errorf("output missing %q after previous string\nGot:\n%s", want, code)
			return
		}
		last += idx + len(want)
	}
}

func TestGenerator_ChildrenSlotInsideComponentCall(t *testing.T) {
	type tc struct {
		input           string
		wantContains    []string
		wantNotContains []string
		wantOrdered     []string
	}

	tests := map[string]tc{
		"function templ forwards its children to a nested call": {
			input: `package x

templ Box() {
	<div>{children...}</div>
}

templ Card(title string) {
	<div>
		<span>{title}</span>
		@Box() {
			{children...}
		}
	</div>
}`,
			wantContains: []string{
				"func Card(title string, children []*tui.Element) *CardView",
				"_children, children...)",
			},
		},
		"slot keeps its position among sibling children": {
			input: `package x

templ Box() {
	<div>{children...}</div>
}

templ Card() {
	<div>
		@Box() {
			<span>before</span>
			{children...}
			<span>after</span>
		}
	</div>
}`,
			wantOrdered: []string{
				`tui.WithText("before")`,
				"_children, children...)",
				`tui.WithText("after")`,
			},
		},
		"method templ forwards receiver children to a nested call": {
			input: `package x

templ Box() {
	<div>{children...}</div>
}

type card struct {
	children []*tui.Element
}

templ (c *card) Render() {
	<div>
		@Box() {
			{children...}
		}
	</div>
}`,
			wantContains: []string{
				"_children, c.children...)",
			},
			wantNotContains: []string{
				"_children, children...)",
			},
		},
		"method templ forwards receiver children to a struct mount": {
			input: `package x

type card struct {
	children []*tui.Element
}

templ (c *card) Render() {
	<div>
		@NewPanel("title") {
			{children...}
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(",
				"_children, c.children...)",
			},
		},
		"slot inside if within a call block": {
			input: `package x

templ Box() {
	<div>{children...}</div>
}

templ Card(show bool) {
	<div>
		@Box() {
			if show {
				{children...}
			} else {
				<span>hidden</span>
			}
		}
	</div>
}`,
			wantOrdered: []string{
				"if show {",
				"_children, children...)",
				"} else {",
				`tui.WithText("hidden")`,
			},
		},
		"method templ slot inside for in the element body uses receiver children": {
			input: `package x

type card struct {
	rows     []int
	children []*tui.Element
}

templ (c *card) Render() {
	<div>
		for _, i := range c.rows {
			{children...}
		}
	</div>
}`,
			wantContains:    []string{"range c.children {"},
			wantNotContains: []string{"range children {"},
		},
		"let-bound call block forwards children": {
			input: `package x

templ Box() {
	<div>{children...}</div>
}

templ Card() {
	inner := @Box() {
		{children...}
	}
	<div>{inner}</div>
}`,
			wantContains: []string{"_children, children...)"},
		},
		"slot inside for within a call block": {
			input: `package x

templ Box() {
	<div>{children...}</div>
}

templ Card(n []int) {
	<div>
		@Box() {
			for _, i := range n {
				{children...}
			}
		}
	</div>
}`,
			wantOrdered: []string{
				"range n {",
				"_children, children...)",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			code := parseAnalyzeGenerate(t, tt.input)
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing expected string: %q\nGot:\n%s", want, code)
				}
			}
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(code, notWant) {
					t.Errorf("output contains unexpected string: %q\nGot:\n%s", notWant, code)
				}
			}
			assertOrdered(t, code, tt.wantOrdered)
		})
	}
}

// Nodes that the element-body paths already handle must also survive inside
// a component call block, including nested for/if bodies there.
func TestGenerator_ComponentCallChildrenNodeKinds(t *testing.T) {
	type tc struct {
		input           string
		wantContains    []string
		wantNotContains []string
	}

	tests := map[string]tc{
		"let reference inside for within a call block": {
			input: `package x

templ Box() {
	<div>{children...}</div>
}

templ App(items []string) {
	label := <span>hi</span>
	<div>
		@Box() {
			for _, it := range items {
				{label}
			}
		}
	</div>
}`,
			wantContains:    []string{"_children, label)"},
			wantNotContains: []string{"tui.WithText(label)"},
		},
		"let reference inside if within a call block": {
			input: `package x

templ Box() {
	<div>{children...}</div>
}

templ App(show bool) {
	label := <span>hi</span>
	<div>
		@Box() {
			if show {
				{label}
			}
		}
	</div>
}`,
			wantContains:    []string{"_children, label)"},
			wantNotContains: []string{"tui.WithText(label)"},
		},
		"component expression directly in a call block": {
			input: `package x

templ Box() {
	<div>{children...}</div>
}

type host struct {
	widget tui.Component
}

templ (c *host) Render() {
	<div>
		@Box() {
			@c.widget
		}
	</div>
}`,
			wantContains: []string{
				"c.widget.Render(app)",
				"any(c.widget).(tui.AppBinder)",
			},
		},
		"go statement directly in a call block": {
			input: `package x

templ Box() {
	<div>{children...}</div>
}

templ App() {
	<div>
		@Box() {
			n := 3
			<span>{fmt.Sprint(n)}</span>
		}
	</div>
}`,
			wantContains: []string{"n := 3"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			code := parseAnalyzeGenerate(t, tt.input)
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing expected string: %q\nGot:\n%s", want, code)
				}
			}
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(code, notWant) {
					t.Errorf("output contains unexpected string: %q\nGot:\n%s", notWant, code)
				}
			}
		})
	}
}

// Function templs hoist conditional component vars and collect loop views for
// watcher aggregation; calls nested inside a call block must follow the same rules.
func TestGenerator_NestedCallInsideCallBlockHoisting(t *testing.T) {
	type tc struct {
		input           string
		wantContains    []string
		wantNotContains []string
	}

	const prelude = `package x

templ Box() {
	<div>{children...}</div>
}

templ Inner(x string) {
	<span>{x}</span>
}
`

	tests := map[string]tc{
		"call block inside if in the element body hoists the nested call": {
			input: prelude + `
templ App(show bool) {
	<div>
		if show {
			@Box() {
				@Inner("a")
			}
		}
	</div>
}`,
			wantContains:    []string{` = Inner("a")`},
			wantNotContains: []string{`:= Inner("a")`},
		},
		"call block inside for in the element body collects nested views": {
			input: prelude + `
templ App(xs []string) {
	<div>
		for _, x := range xs {
			@Box() {
				@Inner(x)
			}
		}
	</div>
}`,
			wantContains: []string{"[]*InnerView", ":= Inner(x)"},
		},
		"for inside a call block collects nested views": {
			input: prelude + `
templ App(xs []string) {
	<div>
		@Box() {
			for _, x := range xs {
				@Inner(x)
			}
		}
	</div>
}`,
			wantContains: []string{"[]*InnerView", ":= Inner(x)"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			code := parseAnalyzeGenerate(t, tt.input)
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing expected string: %q\nGot:\n%s", want, code)
				}
			}
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(code, notWant) {
					t.Errorf("output contains unexpected string: %q\nGot:\n%s", notWant, code)
				}
			}
		})
	}
}
