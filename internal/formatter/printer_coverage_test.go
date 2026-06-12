package formatter

import (
	"testing"
)

// TestFormatBodyStatementsCoverage exercises printNode paths for raw Go
// statements (GoCode) and component expressions (ComponentExpr) in bodies.
func TestFormatBodyStatementsCoverage(t *testing.T) {
	type tc struct {
		input string
		want  string
	}

	tests := map[string]tc{
		"go statements and component expression": {
			input: `package main

templ (c *card) Render() {
count := 1
fmt.Println("hi")
@c.textarea
<span>{count}</span>
}
`,
			want: `package main

templ (c *card) Render() {
	count := 1
	fmt.Println("hi")
	@c.textarea
	<span>{count}</span>
}
`,
		},
		"component expression with leading comment": {
			input: `package main

templ (c *card) Render() {
// the text area
@c.textarea
<span>done</span>
}
`,
			want: `package main

templ (c *card) Render() {
	// the text area
	@c.textarea
	<span>done</span>
}
`,
		},
		"go statement with leading comment": {
			input: `package main

templ A() {
// compute a label
label := "hi"
<span>{label}</span>
}
`,
			want: `package main

templ A() {
	// compute a label
	label := "hi"
	<span>{label}</span>
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

			// Formatting must be idempotent.
			again, err := fmtr.Format("test.gsx", got)
			if err != nil {
				t.Fatalf("second Format() error = %v", err)
			}
			if again != got {
				t.Errorf("Format() not idempotent:\nfirst:\n%s\nsecond:\n%s", got, again)
			}
		})
	}
}
