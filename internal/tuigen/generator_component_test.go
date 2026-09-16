package tuigen

import (
	"slices"
	"strings"
	"testing"
)

func TestGenerator_TrackComponentExprField(t *testing.T) {
	type tc struct {
		expr        string
		loops       []loopVarEntry
		wantPlain   []string
		wantIndexed []string
	}

	tests := map[string]tc{
		"plain field":         {expr: "c.footer", wantPlain: []string{"footer"}},
		"nested field":        {expr: "c.a.b"},
		"method call":         {expr: "c.view()"},
		"other receiver":      {expr: "x.footer"},
		"slice index":         {expr: "c.items[i]", wantIndexed: []string{"items"}},
		"literal index":       {expr: "c.content[0]", wantIndexed: []string{"content"}},
		"map key":             {expr: `c.m["k"]`, wantIndexed: []string{"m"}},
		"nested index":        {expr: "c.rows[c.order[0]]", wantIndexed: []string{"rows"}},
		"bracket in key":      {expr: `c.m["]"]`, wantIndexed: []string{"m"}},
		"field of indexed":    {expr: "c.items[i].view"},
		"call on indexed":     {expr: "c.items[i]()"},
		"unterminated index":  {expr: "c.items[i"},
		"loop value of field": {expr: "it", loops: []loopVarEntry{{value: "it", iterable: "c.items"}}, wantIndexed: []string{"items"}},
		"loop value of local": {expr: "it", loops: []loopVarEntry{{value: "it", iterable: "items"}}},
		"loop value of call":  {expr: "it", loops: []loopVarEntry{{value: "it", iterable: "c.items()"}}},
		"other loop variable": {expr: "x", loops: []loopVarEntry{{value: "it", iterable: "c.items"}}},
		"outer loop value":    {expr: "row", loops: []loopVarEntry{{value: "row", iterable: "c.rows"}, {value: "cell", iterable: "row"}}, wantIndexed: []string{"rows"}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			g := &Generator{currentReceiver: "c", loopVarStack: tt.loops}
			g.trackComponentExprField(tt.expr)
			g.trackComponentExprField(tt.expr) // second call must not duplicate
			if !slices.Equal(g.componentExprFields, tt.wantPlain) {
				t.Errorf("componentExprFields = %v, want %v", g.componentExprFields, tt.wantPlain)
			}
			if !slices.Equal(g.componentExprIndexedFields, tt.wantIndexed) {
				t.Errorf("componentExprIndexedFields = %v, want %v", g.componentExprIndexedFields, tt.wantIndexed)
			}
		})
	}
}

func TestGenerator_IndexedComponentExprBindApp(t *testing.T) {
	type tc struct {
		input           string
		wantContains    []string
		wantNotContains []string
		wantCount       map[string]int
	}

	const bindItemsLoop = "for _, item := range c.items {\n\t\tif binder, ok := any(item).(tui.AppBinder); ok {"
	const unbindItemsLoop = "for _, item := range c.items {\n\t\tif unbinder, ok := any(item).(tui.AppUnbinder); ok {"

	tests := map[string]tc{
		"slice field used through an index": {
			input: `package x

import "github.com/grindlemire/go-tui"

type shell struct {
	active int
	items  []tui.Component
}

templ (c *shell) Render() {
	<div>@c.items[c.active]</div>
}`,
			wantContains: []string{bindItemsLoop, unbindItemsLoop},
		},
		"map field used through a key": {
			input: `package x

import "github.com/grindlemire/go-tui"

type shell struct {
	m map[string]tui.Component
}

templ (c *shell) Render() {
	<div>@c.m["k"]</div>
}`,
			wantContains: []string{
				"for _, item := range c.m {\n\t\tif binder, ok := any(item).(tui.AppBinder); ok {",
				"for _, item := range c.m {\n\t\tif unbinder, ok := any(item).(tui.AppUnbinder); ok {",
			},
		},
		"slice field used only through a loop": {
			input: `package x

import "github.com/grindlemire/go-tui"

type shell struct {
	items []tui.Component
}

templ (c *shell) Render() {
	<div>
		for _, it := range c.items {
			@it
		}
	</div>
}`,
			wantContains: []string{bindItemsLoop, unbindItemsLoop},
		},
		// Unbind assertions appear twice: updatePropsFields unbinds the
		// replaced values and unbindAppFields unbinds on eviction.
		"plain and indexed fields are each emitted once": {
			input: `package x

import "github.com/grindlemire/go-tui"

type shell struct {
	view  tui.Component
	views []tui.Component
}

templ (c *shell) Render() {
	<div>
		@c.view
		@c.views[0]
		@c.view
		@c.views[1]
	</div>
}`,
			wantCount: map[string]int{
				"if binder, ok := any(c.view).(tui.AppBinder); ok {":     1,
				"if unbinder, ok := any(c.view).(tui.AppUnbinder); ok {": 2,
				"for _, item := range c.views {":                         3,
				"if binder, ok := any(item).(tui.AppBinder); ok {":       1,
				"if unbinder, ok := any(item).(tui.AppUnbinder); ok {":   2,
			},
		},
		"named collection type is not forwarded": {
			input: `package x

import "github.com/grindlemire/go-tui"

type shell struct {
	content tui.Panes
}

templ (c *shell) Render() {
	<div>@c.content[0]</div>
}`,
			wantNotContains: []string{"range c.content", "bindAppFields", "unbindAppFields"},
		},
		"state field keeps the direct BindApp call": {
			input: `package x

import "github.com/grindlemire/go-tui"

type shell struct {
	items *tui.State[[]tui.Component]
}

templ (c *shell) Render() {
	<div>@c.items.Get()[0]</div>
}`,
			wantContains:    []string{"c.items.BindApp(app)"},
			wantNotContains: []string{"range c.items"},
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
					t.Errorf("output should not contain %q\nGot:\n%s", notWant, code)
				}
			}
			for snippet, want := range tt.wantCount {
				if got := strings.Count(code, snippet); got != want {
					t.Errorf("output contains %q %d times, want %d\nGot:\n%s", snippet, got, want, code)
				}
			}
		})
	}
}

// updatePropsFieldsBody returns the body of the generated updatePropsFields
// helper so assertions cannot accidentally match bindAppFields or unbindAppFields.
func updatePropsFieldsBody(t *testing.T, code string) string {
	t.Helper()
	const sig = "updatePropsFields(fresh tui.Component) {\n"
	start := strings.Index(code, sig)
	if start < 0 {
		t.Fatalf("output has no updatePropsFields helper\nGot:\n%s", code)
	}
	rest := code[start+len(sig):]
	end := strings.Index(rest, "\n}\n")
	if end < 0 {
		t.Fatalf("updatePropsFields helper is unterminated\nGot:\n%s", code)
	}
	return rest[:end]
}

func TestGenerator_UpdatePropsUnbindsComponentExprFields(t *testing.T) {
	type tc struct {
		input           string
		wantBefore      [2]string // wantBefore[0] must appear before wantBefore[1] in the helper body
		wantNotContains []string
	}

	tests := map[string]tc{
		"plain field is unbound before it is copied": {
			input: `package x

import "github.com/grindlemire/go-tui"

type shell struct {
	view tui.Component
}

templ (c *shell) Render() {
	<div>@c.view</div>
}`,
			wantBefore: [2]string{
				"if unbinder, ok := any(c.view).(tui.AppUnbinder); ok {\n\t\tunbinder.UnbindApp()\n\t}",
				"c.view = f.view",
			},
		},
		"indexed slice field is ranged before it is copied": {
			input: `package x

import "github.com/grindlemire/go-tui"

type shell struct {
	active int
	items  []tui.Component
}

templ (c *shell) Render() {
	<div>@c.items[c.active]</div>
}`,
			wantBefore: [2]string{
				"for _, item := range c.items {\n\t\tif unbinder, ok := any(item).(tui.AppUnbinder); ok {\n\t\t\tunbinder.UnbindApp()\n\t\t}\n\t}",
				"c.items = f.items",
			},
		},
		"props not rendered through @expr are only copied": {
			input: `package x

type shell struct {
	title string
	count int
}

templ (c *shell) Render() {
	<div>{c.title}</div>
}`,
			wantBefore:      [2]string{"c.title = f.title", "c.count = f.count"},
			wantNotContains: []string{"AppUnbinder"},
		},
		"state field is neither copied nor unbound": {
			input: `package x

import "github.com/grindlemire/go-tui"

type shell struct {
	count *tui.State[int]
	view  tui.Component
}

templ (c *shell) Render() {
	<div>@c.view</div>
}`,
			wantBefore:      [2]string{"if unbinder, ok := any(c.view).(tui.AppUnbinder); ok {", "c.view = f.view"},
			wantNotContains: []string{"c.count"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := parseAndGenerateSkipImports("test.gsx", tt.input)
			if err != nil {
				t.Fatalf("generation failed: %v", err)
			}
			body := updatePropsFieldsBody(t, string(output))
			first := strings.Index(body, tt.wantBefore[0])
			second := strings.Index(body, tt.wantBefore[1])
			if first < 0 || second < 0 {
				t.Fatalf("updatePropsFields missing %q or %q\nGot:\n%s", tt.wantBefore[0], tt.wantBefore[1], body)
			}
			if first > second {
				t.Errorf("updatePropsFields has %q after %q\nGot:\n%s", tt.wantBefore[0], tt.wantBefore[1], body)
			}
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(body, notWant) {
					t.Errorf("updatePropsFields should not contain %q\nGot:\n%s", notWant, body)
				}
			}
		})
	}
}
