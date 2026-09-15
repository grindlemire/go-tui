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
				"if unbinder, ok := any(c.view).(tui.AppUnbinder); ok {": 1,
				"for _, item := range c.views {":                         2,
				"if binder, ok := any(item).(tui.AppBinder); ok {":       1,
				"if unbinder, ok := any(item).(tui.AppUnbinder); ok {":   1,
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
