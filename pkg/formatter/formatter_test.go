package formatter

import (
	"strings"
	"testing"
)

// newTestFormatter creates a formatter with import fixing disabled for tests
// that don't specifically test import behavior.
func newTestFormatter() *Formatter {
	f := New()
	f.FixImports = false
	return f
}

// TestFormat tests basic formatting scenarios.
func TestFormat(t *testing.T) {
	type tc struct {
		name  string
		input string
		want  string
	}

	tests := map[string]tc{
		"simple package and component": {
			input: `package main

templ Hello() {
<span>Hello</span>
}
`,
			want: `package main

templ Hello() {
	<span>Hello</span>
}
`,
		},
		"single import": {
			input: `package main

import "fmt"

templ Hello() {
<span>{fmt.Sprintf("hi")}</span>
}
`,
			want: `package main

import "fmt"

templ Hello() {
	<span>{fmt.Sprintf("hi")}</span>
}
`,
		},
		"multiple imports": {
			input: `package main

import (
"fmt"
"strings"
)

templ Hello() {
<span>{strings.ToUpper(fmt.Sprintf("hi"))}</span>
}
`,
			want: `package main

import (
	"fmt"
	"strings"
)

templ Hello() {
	<span>{strings.ToUpper(fmt.Sprintf("hi"))}</span>
}
`,
		},
		"import with alias": {
			input: `package main

import (
tui "github.com/grindlemire/go-tui/pkg/tui"
)

templ Hello() {
<div border={tui.BorderSingle}></div>
}
`,
			want: `package main

import tui "github.com/grindlemire/go-tui/pkg/tui"

templ Hello() {
	<div border={tui.BorderSingle}></div>
}
`,
		},
		"component with parameters": {
			input: `package main

templ Card(title string, count int) {
<span>{title}</span>
}
`,
			want: `package main

templ Card(title string, count int) {
	<span>{title}</span>
}
`,
		},
		"nested elements": {
			input: `package main

templ Layout() {
<div>
<div>
<span>Hello</span>
</div>
</div>
}
`,
			want: `package main

templ Layout() {
	<div>
		<div>
			<span>Hello</span>
		</div>
	</div>
}
`,
		},
		"self-closing element": {
			input: `package main

templ Divider() {
<hr />
}
`,
			want: `package main

templ Divider() {
	<hr />
}
`,
		},
		"for loop": {
			input: `package main

templ List(items []string) {
@for i, item := range items {
<span>{item}</span>
}
}
`,
			want: `package main

templ List(items []string) {
	@for i, item := range items {
		<span>{item}</span>
	}
}
`,
		},
		"if statement": {
			input: `package main

templ Cond(show bool) {
@if show {
<span>Visible</span>
}
}
`,
			want: `package main

templ Cond(show bool) {
	@if show {
		<span>Visible</span>
	}
}
`,
		},
		"if-else statement": {
			input: `package main

templ Cond(show bool) {
@if show {
<span>Yes</span>
} @else {
<span>No</span>
}
}
`,
			want: `package main

templ Cond(show bool) {
	@if show {
		<span>Yes</span>
	} @else {
		<span>No</span>
	}
}
`,
		},
		"let binding": {
			input: `package main

templ WithLet() {
@let x = <span>Hello</span>
{x}
}
`,
			want: `package main

templ WithLet() {
	@let x = <span>Hello</span>
	{x}
}
`,
		},
		"component call": {
			input: `package main

templ Parent() {
@Child("arg1", "arg2")
}

templ Child(a string, b string) {
<span>{a}</span>
}
`,
			want: `package main

templ Parent() {
	@Child("arg1", "arg2")
}

templ Child(a string, b string) {
	<span>{a}</span>
}
`,
		},
		"component call with children": {
			input: `package main

templ Parent() {
@Card("Title") {
<span>Content</span>
}
}

templ Card(title string) {
<div>
<span>{title}</span>
{children...}
</div>
}
`,
			want: `package main

templ Parent() {
	@Card("Title") {
		<span>Content</span>
	}
}

templ Card(title string) {
	<div>
		<span>{title}</span>
		{children...}
	</div>
}
`,
		},
		"multiple attributes": {
			input: `package main

templ Box() {
<div border={1} padding={2} margin={1}>
<span>Content</span>
</div>
}
`,
			want: `package main

templ Box() {
	<div border={1} padding={2} margin={1}>
		<span>Content</span>
	</div>
}
`,
		},
		"string attribute": {
			input: `package main

templ Styled() {
<div class="flex-col gap-1">
<span>Content</span>
</div>
}
`,
			want: `package main

templ Styled() {
	<div class="flex-col gap-1">
		<span>Content</span>
	</div>
}
`,
		},
		"named ref": {
			input: `package main

templ App() {
<div #Content class="flex-col"></div>
}
`,
			want: `package main

templ App() {
	<div #Content class="flex-col"></div>
}
`,
		},
		"named ref with children": {
			input: `package main

templ App() {
<div #Wrapper>
<span #Title>Hello</span>
</div>
}
`,
			want: `package main

templ App() {
	<div #Wrapper>
		<span #Title>Hello</span>
	</div>
}
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fmtr := newTestFormatter()
			got, err := fmtr.Format("test.gsx", tt.input)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Format() mismatch:\ngot:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

// TestFormatRoundTrip tests that format(format(x)) == format(x).
func TestFormatRoundTrip(t *testing.T) {
	type tc struct {
		input string
	}

	tests := map[string]tc{
		"simple component": {
			input: `package main

templ Hello() {
	<span>Hello</span>
}
`,
		},
		"complex nested": {
			input: `package main

import (
	"fmt"
)

templ Complex(items []string, selected int) {
	<div border={1}>
		@for i, item := range items {
			@if i == selected {
				<span class="bold">{fmt.Sprintf("> %s", item)}</span>
			} @else {
				<span>{item}</span>
			}
		}
	</div>
}
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fmtr := newTestFormatter()

			// First format
			first, err := fmtr.Format("test.gsx", tt.input)
			if err != nil {
				t.Fatalf("First Format() error = %v", err)
			}

			// Second format (should be identical)
			second, err := fmtr.Format("test.gsx", first)
			if err != nil {
				t.Fatalf("Second Format() error = %v", err)
			}

			if first != second {
				t.Errorf("Round-trip failed:\nfirst:\n%s\nsecond:\n%s", first, second)
			}
		})
	}
}

// TestFormatWithResult tests the FormatWithResult method.
func TestFormatWithResult(t *testing.T) {
	type tc struct {
		input       string
		wantChanged bool
	}

	tests := map[string]tc{
		"already formatted": {
			input: `package main

templ Hello() {
	<span>Hello</span>
}
`,
			wantChanged: false,
		},
		"needs formatting": {
			input: `package main

templ Hello() {
<span>Hello</span>
}
`,
			wantChanged: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fmtr := newTestFormatter()
			res, err := fmtr.FormatWithResult("test.gsx", tt.input)
			if err != nil {
				t.Fatalf("FormatWithResult() error = %v", err)
			}
			if res.Changed != tt.wantChanged {
				t.Errorf("FormatWithResult() Changed = %v, want %v", res.Changed, tt.wantChanged)
			}
		})
	}
}

// TestFormatParseError tests that parse errors are returned.
func TestFormatParseError(t *testing.T) {
	type tc struct {
		input string
	}

	tests := map[string]tc{
		"missing package": {
			input: `templ Hello() {
	<span>Hello</span>
}
`,
		},
		"invalid syntax": {
			input: `package main

@component Hello( {
	<span>Hello</span>
}
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fmtr := newTestFormatter()
			_, err := fmtr.Format("test.gsx", tt.input)
			if err == nil {
				t.Error("Format() expected error, got nil")
			}
		})
	}
}

// TestEscapeString tests the string escaping function.
func TestEscapeString(t *testing.T) {
	type tc struct {
		input string
		want  string
	}

	tests := map[string]tc{
		"no escaping needed": {
			input: "hello world",
			want:  "hello world",
		},
		"newline": {
			input: "hello\nworld",
			want:  `hello\nworld`,
		},
		"tab": {
			input: "hello\tworld",
			want:  `hello\tworld`,
		},
		"quote": {
			input: `hello "world"`,
			want:  `hello \"world\"`,
		},
		"backslash": {
			input: `hello\world`,
			want:  `hello\\world`,
		},
		"multiple escapes": {
			input: "line1\nline2\ttab\"quote\"",
			want:  `line1\nline2\ttab\"quote\"`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := escapeString(tt.input)
			if got != tt.want {
				t.Errorf("escapeString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestFormatPreservesGoExpressions tests that Go expressions are preserved as-is.
func TestFormatPreservesGoExpressions(t *testing.T) {
	input := `package main

import "fmt"

templ Complex() {
	<span>{fmt.Sprintf("%d + %d = %d", 1, 2, 1+2)}</span>
}
`
	fmtr := newTestFormatter()
	got, err := fmtr.Format("test.gsx", input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Check that the Go expression is preserved
	if !strings.Contains(got, `{fmt.Sprintf("%d + %d = %d", 1, 2, 1+2)}`) {
		t.Errorf("Go expression not preserved in output:\n%s", got)
	}
}

// TestFormatAutoImports tests that missing imports are automatically added.
func TestFormatAutoImports(t *testing.T) {
	type tc struct {
		name  string
		input string
		want  []string // import paths that should be present
	}

	tests := map[string]tc{
		"adds tui and element imports": {
			input: `package main

templ Hello() {
	<span>Hello</span>
}
`,
			want: []string{
				"github.com/grindlemire/go-tui/pkg/tui",
				"github.com/grindlemire/go-tui/pkg/tui/element",
			},
		},
		"adds fmt import when used": {
			input: `package main

templ Hello() {
	<span>{fmt.Sprintf("hello")}</span>
}
`,
			want: []string{
				"fmt",
				"github.com/grindlemire/go-tui/pkg/tui",
				"github.com/grindlemire/go-tui/pkg/tui/element",
			},
		},
		"preserves existing imports": {
			input: `package main

import "fmt"

templ Hello() {
	<span>{fmt.Sprintf("hello")}</span>
}
`,
			want: []string{
				"fmt",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fmtr := New() // Use default formatter with FixImports=true
			got, err := fmtr.Format("test.gsx", tt.input)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}

			for _, imp := range tt.want {
				if !strings.Contains(got, `"`+imp+`"`) {
					t.Errorf("missing import %q in output:\n%s", imp, got)
				}
			}
		})
	}
}
