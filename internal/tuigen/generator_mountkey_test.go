package tuigen

import (
	"strings"
	"testing"
)

// TestGenerator_MountKeyExpressions pins the mount cache key expressions
// emitted for component call sites. Standalone sites use a plain int; sites
// inside loops combine the site id with each enclosing loop's key value via
// tui.MountKey (which is what makes map-keyed loops compile, issue #92);
// a key={...} attribute replaces the loop values with user identity.
func TestGenerator_MountKeyExpressions(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"slice loop with explicit index": {
			input: `package x

type app struct{}

templ (c *app) Render() {
	<div>
		for i, item := range c.items {
			<markdown source={item} />
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, tui.MountKey(0, i), func() tui.Component {",
			},
		},
		"slice loop with discarded index uses synthetic variable": {
			input: `package x

type app struct{}

templ (c *app) Render() {
	<div>
		for _, item := range c.items {
			<markdown source={item} />
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, tui.MountKey(0, __idx_0), func() tui.Component {",
			},
		},
		"map loop keys by the map key (issue 92)": {
			input: `package x

type app struct{}

templ (c *app) Render() {
	<div>
		for name, doc := range c.docs {
			<markdown source={doc} />
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, tui.MountKey(0, name), func() tui.Component {",
			},
		},
		"nested loops pass both loop keys": {
			input: `package x

type app struct{}

templ (c *app) Render() {
	<div>
		for i, row := range c.rows {
			for j, item := range row {
				<markdown source={item} />
			}
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, tui.MountKey(0, i, j), func() tui.Component {",
			},
		},
		"standalone component after loop keeps plain site index": {
			input: `package x

type app struct{}

templ (c *app) Render() {
	<div>
		for i, item := range c.items {
			<markdown source={item} />
		}
		<textarea onSubmit={c.submit} />
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, tui.MountKey(0, i), func() tui.Component {",
				"app.MountPersistent(c, 1, func() tui.Component {",
			},
		},
		"key attribute replaces the innermost loop key with user identity": {
			input: `package x

type app struct{}

templ (c *app) Render() {
	<div>
		for _, item := range c.items {
			<markdown key={item.ID} source={item.Text} />
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, tui.MountKey(0, item.ID), func() tui.Component {",
			},
		},
		"key attribute in nested loop is scoped to the innermost loop": {
			input: `package x

type app struct{}

templ (c *app) Render() {
	<div>
		for i, group := range c.groups {
			for _, item := range group {
				<markdown key={item.ID} source={item.Text} />
			}
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, tui.MountKey(0, i, item.ID), func() tui.Component {",
			},
		},
		"key attribute still drives RefMap.Put alongside mount identity": {
			input: `package x

type app struct{}

templ (c *app) Render() {
	<div>
		for _, item := range c.items {
			<textarea ref={c.areas} key={item.ID} placeholder={item.Name} />
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, tui.MountKey(0, item.ID), func() tui.Component {",
				"c.areas.Put(item.ID, __tui_",
			},
		},
		"standalone component with key uses sweepable Mount": {
			input: `package x

type app struct{}

templ (c *app) Render() {
	<div>
		<markdown key={c.selectedDoc} source={c.body} />
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, tui.MountKey(0, c.selectedDoc), func() tui.Component {",
			},
		},
		"struct component call in loop": {
			input: `package x

type app struct{}

templ (c *app) Render() {
	<div>
		for i, item := range c.items {
			@Widget(item)
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, tui.MountKey(0, i), func() tui.Component {",
			},
		},
		"keyed wrapper div keys struct component call": {
			input: `package x

type app struct{}

templ (c *app) Render() {
	<div>
		for _, item := range c.items {
			<div key={item.ID}>
				@ChatMessage(item)
			</div>
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, tui.MountKey(0, item.ID), func() tui.Component {",
			},
		},
		"keyed wrapper outside inner loop prefixes inner loop keys": {
			input: `package x

type app struct{}

templ (c *app) Render() {
	<div>
		for _, group := range c.groups {
			<div key={group.ID}>
				for _, item := range group.Items {
					@Widget(item)
				}
			</div>
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, tui.MountKey(0, group.ID, __idx_1), func() tui.Component {",
			},
		},
		"component element inherits wrapper key": {
			input: `package x

type app struct{}

templ (c *app) Render() {
	<div>
		for _, item := range c.items {
			<div key={item.ID}>
				<textarea placeholder={item.Name} />
			</div>
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, tui.MountKey(0, item.ID), func() tui.Component {",
			},
		},
		"own key wins over inherited wrapper key": {
			input: `package x

type app struct{}

templ (c *app) Render() {
	<div>
		for _, item := range c.items {
			<div key={item.ID}>
				<markdown key={item.Sub} source={item.Text} />
			</div>
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, tui.MountKey(0, item.Sub), func() tui.Component {",
			},
		},
		"sibling after keyed wrapper falls back to loop key": {
			input: `package x

type app struct{}

templ (c *app) Render() {
	<div>
		for i, item := range c.items {
			<div key={item.ID}>
				<textarea placeholder={item.Name} />
			</div>
			<markdown source={item.Text} />
		}
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, tui.MountKey(0, item.ID), func() tui.Component {",
				"app.Mount(c, tui.MountKey(1, i), func() tui.Component {",
			},
		},
		"standalone keyed wrapper uses sweepable Mount": {
			input: `package x

type app struct{}

templ (c *app) Render() {
	<div key={c.activeID}>
		<textarea placeholder="notes" />
	</div>
}`,
			wantContains: []string{
				"app.Mount(c, tui.MountKey(0, c.activeID), func() tui.Component {",
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
