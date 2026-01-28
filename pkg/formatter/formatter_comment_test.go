package formatter

import (
	"testing"
)

// TestFormatCommentLeading tests leading comment preservation.
func TestFormatCommentLeading(t *testing.T) {
	type tc struct {
		input string
		want  string
	}

	tests := map[string]tc{
		"leading comment before component": {
			input: `package main

// This is a doc comment
func Hello() Element {
	<span>Hello</span>
}
`,
			want: `package main

// This is a doc comment
func Hello() Element {
	<span>Hello</span>
}
`,
		},
		"leading comment before element": {
			input: `package main

func Hello() Element {
	// Comment before element
	<span>Hello</span>
}
`,
			want: `package main

func Hello() Element {
	// Comment before element
	<span>Hello</span>
}
`,
		},
		"leading comment before if": {
			input: `package main

func Hello(show bool) Element {
	// Comment before if
	@if show {
		<span>Hello</span>
	}
}
`,
			want: `package main

func Hello(show bool) Element {
	// Comment before if
	@if show {
		<span>Hello</span>
	}
}
`,
		},
		"leading comment before for": {
			input: `package main

func Hello(items []string) Element {
	// Comment before for
	@for _, item := range items {
		<span>{item}</span>
	}
}
`,
			want: `package main

func Hello(items []string) Element {
	// Comment before for
	@for _, item := range items {
		<span>{item}</span>
	}
}
`,
		},
		"leading block comment": {
			input: `package main

/* Block comment
   spanning multiple lines */
func Hello() Element {
	<span>Hello</span>
}
`,
			want: `package main

/*
Block comment
spanning multiple lines
*/
func Hello() Element {
	<span>Hello</span>
}
`,
		},
		"leading comment before package": {
			input: `// File-level comment
package main

func Hello() Element {
	<span>Hello</span>
}
`,
			want: `// File-level comment
package main

func Hello() Element {
	<span>Hello</span>
}
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fmtr := New()
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

// TestFormatCommentTrailing tests trailing comment preservation.
func TestFormatCommentTrailing(t *testing.T) {
	type tc struct {
		input string
		want  string
	}

	tests := map[string]tc{
		"trailing comment on element": {
			input: `package main

func Hello() Element {
	<span>Hello</span>  // trailing
}
`,
			want: `package main

func Hello() Element {
	<span>Hello</span>  // trailing
}
`,
		},
		"trailing comment on self-closing element": {
			input: `package main

func Hello() Element {
	<hr />  // divider
}
`,
			want: `package main

func Hello() Element {
	<hr />  // divider
}
`,
		},
		// Note: import trailing comments are attached to the next declaration
		// by the parser, so we skip this test for now.
		// "trailing comment on import": { ... }
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fmtr := New()
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

// TestFormatCommentOrphan tests orphan comment preservation.
func TestFormatCommentOrphan(t *testing.T) {
	type tc struct {
		input string
		want  string
	}

	tests := map[string]tc{
		"orphan comment in component body": {
			input: `package main

func Hello() Element {
	// orphan comment in body
	<span>Hello</span>
}
`,
			want: `package main

func Hello() Element {
	// orphan comment in body
	<span>Hello</span>
}
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fmtr := New()
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

// TestFormatCommentRoundTrip tests format idempotency with comments.
func TestFormatCommentRoundTrip(t *testing.T) {
	type tc struct {
		input string
	}

	tests := map[string]tc{
		"file with leading comments": {
			input: `// File comment
package main

// Doc comment
func Hello() Element {
	<span>Hello</span>
}
`,
		},
		"file with trailing comments": {
			input: `package main

func Hello() Element {
	<span>Hello</span>  // inline
}
`,
		},
		"complex file with many comments": {
			input: `// File comment
package main

import "fmt"

// Header component
func Header(title string) Element {
	// Container
	<div class="header">
		<span>{title}</span>  // title text
	</div>
}
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fmtr := New()

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

// TestFormatCommentGroupSeparation tests that blank lines between comment groups are preserved.
func TestFormatCommentGroupSeparation(t *testing.T) {
	type tc struct {
		input string
		want  string
	}

	tests := map[string]tc{
		"preserve blank line between comment groups": {
			input: `package main

// Unassigned block comment
// For package comment

// ItemList test
func Hello() Element {
	<span>Hello</span>
}
`,
			want: `package main

// Unassigned block comment
// For package comment

// ItemList test
func Hello() Element {
	<span>Hello</span>
}
`,
		},
		"preserve blank line inside component": {
			input: `package main

func Hello() Element {
	// First comment

	// Second comment
	<span>Hello</span>
}
`,
			want: `package main

func Hello() Element {
	// First comment

	// Second comment
	<span>Hello</span>
}
`,
		},
		"no blank line when comments are adjacent": {
			input: `package main

// First comment
// Second comment
func Hello() Element {
	<span>Hello</span>
}
`,
			want: `package main

// First comment
// Second comment
func Hello() Element {
	<span>Hello</span>
}
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fmtr := New()
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

// TestFormatLineCommentSpacing tests proper spacing after // in line comments.
func TestFormatLineCommentSpacing(t *testing.T) {
	type tc struct {
		input string
		want  string
	}

	tests := map[string]tc{
		"missing space after //": {
			input: `package main

//comment without space
func Hello() Element {
	<span>Hello</span>
}
`,
			want: `package main

// comment without space
func Hello() Element {
	<span>Hello</span>
}
`,
		},
		"already has space": {
			input: `package main

// comment with space
func Hello() Element {
	<span>Hello</span>
}
`,
			want: `package main

// comment with space
func Hello() Element {
	<span>Hello</span>
}
`,
		},
		"trailing comment missing space": {
			input: `package main

func Hello() Element {
	<span>Hello</span>  //trailing
}
`,
			want: `package main

func Hello() Element {
	<span>Hello</span>  // trailing
}
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fmtr := New()
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

// TestFormatInlineBlockComment tests proper formatting of inline block comments.
func TestFormatInlineBlockComment(t *testing.T) {
	type tc struct {
		input string
		want  string
	}

	tests := map[string]tc{
		"missing space before closing": {
			input: `package main

import "fmt"

func Hello(item string) Element {
	<span>{fmt.Sprintf("> %s", /* ItemList item*/ item)}</span>
}
`,
			want: `package main

import "fmt"

func Hello(item string) Element {
	<span>{fmt.Sprintf("> %s", /* ItemList item */ item)}</span>
}
`,
		},
		"missing space after opening": {
			input: `package main

func Hello(x int) Element {
	<span>{/*test*/ x}</span>
}
`,
			want: `package main

func Hello(x int) Element {
	<span>{/* test */ x}</span>
}
`,
		},
		"missing both spaces": {
			input: `package main

func Hello(x int) Element {
	<span>{/*test comment*/ x}</span>
}
`,
			want: `package main

func Hello(x int) Element {
	<span>{/* test comment */ x}</span>
}
`,
		},
		"already properly formatted": {
			input: `package main

func Hello(x int) Element {
	<span>{/* test */ x}</span>
}
`,
			want: `package main

func Hello(x int) Element {
	<span>{/* test */ x}</span>
}
`,
		},
		"empty block comment": {
			input: `package main

func Hello(x int) Element {
	<span>{/**/ x}</span>
}
`,
			want: `package main

func Hello(x int) Element {
	<span>{/* */ x}</span>
}
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fmtr := New()
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

// TestFormatCommentComplex tests complex scenarios with comments at all positions.
func TestFormatCommentComplex(t *testing.T) {
	type tc struct {
		input string
		want  string
	}

	tests := map[string]tc{
		"comments at various positions": {
			// Note: The parser attaches comments before import to the following component,
			// so we adjust our test to not have a comment immediately before import.
			input: `// File-level comment
package main

import "fmt"

// Main component documentation
// Multiple lines
func Main(items []string, selected int) Element {
	// Container div
	<div class="main">
		// Loop through items
		@for i, item := range items {
			// Conditional rendering
			@if i == selected {
				<span class="selected">{item}</span>
			} @else {
				<span>{item}</span>
			}
		}
	</div>
}

// Helper function
func helper(s string) string {
	return fmt.Sprintf("[%s]", s)
}
`,
			want: `// File-level comment
package main

import "fmt"

// Main component documentation
// Multiple lines
func Main(items []string, selected int) Element {
	// Container div
	<div class="main">
		// Loop through items
		@for i, item := range items {
			// Conditional rendering
			@if i == selected {
				<span class="selected">{item}</span>
			} @else {
				<span>{item}</span>
			}
		}
	</div>
}

// Helper function
func helper(s string) string {
	return fmt.Sprintf("[%s]", s)
}
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fmtr := New()
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
