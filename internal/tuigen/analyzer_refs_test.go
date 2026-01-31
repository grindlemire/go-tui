package tuigen

import (
	"strings"
	"testing"
)

func TestAnalyzer_LetBindingValidation(t *testing.T) {
	type tc struct {
		input         string
		wantError     bool
		errorContains string
	}

	tests := map[string]tc{
		"valid let binding": {
			input: `package x
templ Test() {
	@let myText = <span>hello</span>
	<div></div>
}`,
			wantError: false,
		},
		"let binding with invalid element": {
			input: `package x
templ Test() {
	@let myText = <badTag />
	<div></div>
}`,
			wantError:     true,
			errorContains: "unknown element tag <badTag>",
		},
		"let binding with invalid attribute": {
			input: `package x
templ Test() {
	@let myText = <span badAttr="value">hello</span>
	<div></div>
}`,
			wantError:     true,
			errorContains: "unknown attribute badAttr",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.gsx", tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAnalyzer_RefValidation(t *testing.T) {
	type tc struct {
		input         string
		wantError     bool
		errorContains string
	}

	tests := map[string]tc{
		"valid ref name": {
			input: `package x
templ Test() {
	<div ref={content}></div>
}`,
			wantError: false,
		},
		"valid ref name with digits": {
			input: `package x
templ Test() {
	<div ref={content2}></div>
}`,
			wantError: false,
		},
		"valid ref name with underscore": {
			input: `package x
templ Test() {
	<div ref={my_Content}></div>
}`,
			wantError: false,
		},
		"reserved name Root": {
			input: `package x
templ Test() {
	<div ref={root}></div>
}`,
			wantError:     true,
			errorContains: "ref name \"root\" is reserved",
		},
		"duplicate ref name": {
			input: `package x
templ Test() {
	<div ref={content}></div>
	<div ref={content}></div>
}`,
			wantError:     true,
			errorContains: "duplicate ref name",
		},
		"duplicate ref name across branches": {
			input: `package x
templ Test(show bool) {
	@if show {
		<div ref={content}></div>
	} @else {
		<div ref={content}></div>
	}
}`,
			wantError:     true,
			errorContains: "duplicate ref name",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.gsx", tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAnalyzer_RefInLoop(t *testing.T) {
	type tc struct {
		input         string
		wantError     bool
		errorContains string
	}

	tests := map[string]tc{
		"ref in loop is valid": {
			input: `package x
templ Test(items []string) {
	<ul>
		@for _, item := range items {
			<li ref={items}>{item}</li>
		}
	</ul>
}`,
			wantError: false,
		},
		"ref with key in loop is valid": {
			input: `package x
templ Test(items []Item) {
	<ul>
		@for _, item := range items {
			<li ref={items} key={item.ID}>{item.Name}</li>
		}
	</ul>
}`,
			wantError: false,
		},
		"ref with key outside loop is invalid": {
			input: `package x
templ Test() {
	<div ref={content} key={someKey}></div>
}`,
			wantError:     true,
			errorContains: "key attribute on ref",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.gsx", tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAnalyzer_RefInConditional(t *testing.T) {
	input := `package x
templ Test(show bool) {
	<div>
		@if show {
			<span ref={label}>hello</span>
		}
	</div>
}`

	_, err := AnalyzeFile("test.gsx", input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Ref inside conditional is valid, it just may be nil at runtime
}

func TestAnalyzer_CollectRefs(t *testing.T) {
	input := `package x
templ Test(items []Item, show bool) {
	<div>
		<div ref={header}></div>
		@if show {
			<span ref={label}>hello</span>
		}
		@for _, item := range items {
			<li ref={items}>{item.Name}</li>
		}
	</div>
}`

	l := NewLexer("test.gsx", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	refs := analyzer.CollectRefs(file.Components[0])

	if len(refs) != 3 {
		t.Fatalf("expected 3 refs, got %d", len(refs))
	}

	// Check header ref
	if refs[0].Name != "header" {
		t.Errorf("refs[0].Name = %q, want 'header'", refs[0].Name)
	}
	if refs[0].ExportName != "Header" {
		t.Errorf("refs[0].ExportName = %q, want 'Header'", refs[0].ExportName)
	}
	if refs[0].InLoop || refs[0].InConditional {
		t.Error("header should not be in loop or conditional")
	}

	// Check label ref (in conditional)
	if refs[1].Name != "label" {
		t.Errorf("refs[1].Name = %q, want 'label'", refs[1].Name)
	}
	if refs[1].ExportName != "Label" {
		t.Errorf("refs[1].ExportName = %q, want 'Label'", refs[1].ExportName)
	}
	if refs[1].InLoop {
		t.Error("label should not be in loop")
	}
	if !refs[1].InConditional {
		t.Error("label should be in conditional")
	}

	// Check items ref (in loop)
	if refs[2].Name != "items" {
		t.Errorf("refs[2].Name = %q, want 'items'", refs[2].Name)
	}
	if refs[2].ExportName != "Items" {
		t.Errorf("refs[2].ExportName = %q, want 'Items'", refs[2].ExportName)
	}
	if !refs[2].InLoop {
		t.Error("items should be in loop")
	}
}
