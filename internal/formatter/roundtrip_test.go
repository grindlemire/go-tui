package formatter

import (
	"testing"
)

func TestFormat_Idempotent(t *testing.T) {
	type tc struct {
		input string
	}

	tests := map[string]tc{
		"simple component": {
			input: `package test

templ Hello() {
	<div class="flex-col">
		<span>Hello</span>
	</div>
}
`,
		},
		"component with imports": {
			input: `package test

import "fmt"

templ Hello(name string) {
	<span>{fmt.Sprintf("Hello %s", name)}</span>
}
`,
		},
		"if else": {
			input: `package test

templ Cond(show bool) {
	<div>
		@if show {
			<span>Yes</span>
		} @else {
			<span>No</span>
		}
	</div>
}
`,
		},
		"for loop": {
			input: `package test

templ List(items []string) {
	<div class="flex-col">
		@for _, item := range items {
			<span>{item}</span>
		}
	</div>
}
`,
		},
		"self-closing elements": {
			input: `package test

templ Divider() {
	<div>
		<hr />
		<br />
	</div>
}
`,
		},
		"let binding": {
			input: `package test

import "fmt"

templ Counter(count int) {
	@let countText = <span>{fmt.Sprintf("Count: %d", count)}</span>
	<div>{countText}</div>
}
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			f := New()

			// First format
			result1, err := f.Format("test.gsx", tt.input)
			if err != nil {
				t.Fatalf("first format failed: %v", err)
			}

			// Second format (should be identical)
			result2, err := f.Format("test.gsx", result1)
			if err != nil {
				t.Fatalf("second format failed: %v", err)
			}

			if result1 != result2 {
				t.Errorf("format is not idempotent\nfirst:\n%s\nsecond:\n%s", result1, result2)
			}
		})
	}
}
