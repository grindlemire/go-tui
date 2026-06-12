package formatter

import (
	"testing"
)

// TestFormatElementAttributesCoverage exercises attribute value printing
// (int, float, bool literals), key={...} extraction in single-line and
// multi-line attribute layouts, and inline-children fallbacks.
func TestFormatElementAttributesCoverage(t *testing.T) {
	type tc struct {
		input string
		want  string
	}

	tests := map[string]tc{
		"literal attribute values normalized to braced form": {
			input: `package main

templ A() {
<div width=10 flexGrow=1.5 disabled=false focusable>hi</div>
}
`,
			want: `package main

templ A() {
	<div width={10} flexGrow={1.5} disabled={false} focusable={true}>hi</div>
}
`,
		},
		"key attribute single line": {
			input: `package main

templ A(items []string) {
for i, it := range items {
<span key={i} class="p-1">{it}</span>
}
}
`,
			want: `package main

templ A(items []string) {
	for i, it := range items {
		<span key={i} class="p-1">{it}</span>
	}
}
`,
		},
		"key attribute multi line": {
			input: `package main

templ A(items []string) {
for i, it := range items {
<div
key={i}
class="p-1">{it}</div>
}
}
`,
			want: `package main

templ A(items []string) {
	for i, it := range items {
		<div
			key={i}
			class="p-1">{it}</div>
	}
}
`,
		},
		"element child forces multi-line": {
			input: `package main

templ A() {
<div><span>a</span></div>
}
`,
			want: `package main

templ A() {
	<div>
		<span>a</span>
	</div>
}
`,
		},
		"go expression with newline forces multi-line": {
			input: "package main\n\ntempl A() {\n<span>{fmt.Sprintf(\n\"x\")}</span>\n}\n",
			want:  "package main\n\ntempl A() {\n\t<span>\n\t\t{fmt.Sprintf(\n\"x\")}\n\t</span>\n}\n",
		},
		"raw string child preserved": {
			input: "package main\n\ntempl A() {\n<span>{fn(`raw /* x */ string`)}</span>\n}\n",
			want:  "package main\n\ntempl A() {\n\t<span>{fn(`raw /* x */ string`)}</span>\n}\n",
		},
		"escaped quote in expression preserved": {
			input: `package main

templ A() {
<span>{fmt.Sprintf("a\"b /* not comment */")}</span>
}
`,
			want: `package main

templ A() {
	<span>{fmt.Sprintf("a\"b /* not comment */")}</span>
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
