package formatter

import (
	"strings"
	"testing"
)

// TestFormatWithResultCoverage exercises the changed/unchanged/error paths.
func TestFormatWithResultCoverage(t *testing.T) {
	type tc struct {
		input       string
		wantContent string
		wantChanged bool
		wantErr     string
	}

	tests := map[string]tc{
		"already formatted is unchanged": {
			input: `package main

templ A() {
	<span>hi</span>
}
`,
			wantContent: `package main

templ A() {
	<span>hi</span>
}
`,
			wantChanged: false,
		},
		"unformatted input is changed": {
			input: `package main

templ A() {
<span>hi</span>
}
`,
			wantContent: `package main

templ A() {
	<span>hi</span>
}
`,
			wantChanged: true,
		},
		"parse error returns empty result": {
			input:   "package main\n\ntempl A() {\n<span>\n}\n",
			wantErr: "test.gsx:6:1: error: expected closing tag </span>",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fmtr := newTestFormatter()
			got, err := fmtr.FormatWithResult("test.gsx", tt.input)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("FormatWithResult() expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("FormatWithResult() error = %q, want it to contain %q", err.Error(), tt.wantErr)
				}
				if got.Content != "" || got.Changed {
					t.Errorf("FormatWithResult() result = %+v, want zero value on error", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("FormatWithResult() error = %v", err)
			}
			if got.Content != tt.wantContent {
				t.Errorf("Content mismatch:\ngot:\n%s\nwant:\n%s", got.Content, tt.wantContent)
			}
			if got.Changed != tt.wantChanged {
				t.Errorf("Changed = %v, want %v", got.Changed, tt.wantChanged)
			}
		})
	}
}

// TestFixImportsTuiFilename verifies the .tui filename is converted for
// goimports resolution and imports are still fixed.
func TestFixImportsTuiFilename(t *testing.T) {
	input := `package main

templ A() {
<span>{fmt.Sprintf("hi")}</span>
}
`
	want := `package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

templ A() {
	<span>{fmt.Sprintf("hi")}</span>
}
`

	fmtr := New() // FixImports enabled
	got, err := fmtr.Format("test.tui", input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if got != want {
		t.Errorf("Format() mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// TestFixImportsGenerationFailure verifies that when the generator cannot
// produce valid Go (invalid expression code), import fixing is skipped but
// formatting still succeeds with the original imports untouched.
func TestFixImportsGenerationFailure(t *testing.T) {
	input := `package main

templ A() {
<span>{a +}</span>
}
`
	want := `package main

templ A() {
	<span>{a +}</span>
}
`

	fmtr := New() // FixImports enabled
	got, err := fmtr.Format("test.gsx", input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if got != want {
		t.Errorf("Format() mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// TestExtractImports exercises the import extraction helper directly,
// including the parse error path.
func TestExtractImports(t *testing.T) {
	type tc struct {
		src       string
		wantPaths []string
		wantErr   bool
	}

	tests := map[string]tc{
		"valid imports": {
			src:       "package p\n\nimport (\n\t\"fmt\"\n\talias \"strings\"\n)\n",
			wantPaths: []string{"fmt", "strings"},
		},
		"invalid go source": {
			src:     "not go code",
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := extractImports([]byte(tt.src))
			if tt.wantErr {
				if err == nil {
					t.Fatal("extractImports() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("extractImports() error = %v", err)
			}
			if len(got) != len(tt.wantPaths) {
				t.Fatalf("extractImports() returned %d imports, want %d", len(got), len(tt.wantPaths))
			}
			for i, p := range tt.wantPaths {
				if got[i].Path != p {
					t.Errorf("import[%d].Path = %q, want %q", i, got[i].Path, p)
				}
			}
			if got[1].Alias != "alias" {
				t.Errorf("import[1].Alias = %q, want %q", got[1].Alias, "alias")
			}
		})
	}
}
