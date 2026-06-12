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
				"app.MountPersistent(c, tui.MountKey(0, i), func() tui.Component {",
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
				"app.MountPersistent(c, tui.MountKey(0, __idx_0), func() tui.Component {",
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
				"app.MountPersistent(c, tui.MountKey(0, name), func() tui.Component {",
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
				"app.MountPersistent(c, tui.MountKey(0, i, j), func() tui.Component {",
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
				"app.MountPersistent(c, tui.MountKey(0, i), func() tui.Component {",
				"app.MountPersistent(c, 1, func() tui.Component {",
			},
		},
		"key attribute replaces loop keys with user identity": {
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
				"app.MountPersistent(c, tui.MountKey(0, item.ID), func() tui.Component {",
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
				"app.MountPersistent(c, tui.MountKey(0, item.ID), func() tui.Component {",
				"c.areas.Put(item.ID, __tui_",
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
