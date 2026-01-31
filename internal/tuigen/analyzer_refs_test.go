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

func TestAnalyzer_NamedRefValidation(t *testing.T) {
	type tc struct {
		input         string
		wantError     bool
		errorContains string
	}

	tests := map[string]tc{
		"valid ref name": {
			input: `package x
templ Test() {
	<div #Content></div>
}`,
			wantError: false,
		},
		"valid ref name with digits": {
			input: `package x
templ Test() {
	<div #Content2></div>
}`,
			wantError: false,
		},
		"valid ref name with underscore": {
			input: `package x
templ Test() {
	<div #My_Content></div>
}`,
			wantError: false,
		},
		"invalid ref name lowercase": {
			input: `package x
templ Test() {
	<div #content></div>
}`,
			wantError:     true,
			errorContains: "invalid ref name",
		},
		"invalid ref name starts with digit": {
			input: `package x
templ Test() {
	<div #123invalid></div>
}`,
			wantError:     true,
			errorContains: "expected identifier", // Parser rejects this before analyzer
		},
		"reserved name Root": {
			input: `package x
templ Test() {
	<div #Root></div>
}`,
			wantError:     true,
			errorContains: "ref name 'Root' is reserved",
		},
		"duplicate ref name": {
			input: `package x
templ Test() {
	<div #Content></div>
	<div #Content></div>
}`,
			wantError:     true,
			errorContains: "duplicate ref name",
		},
		"duplicate ref name across branches": {
			input: `package x
templ Test(show bool) {
	@if show {
		<div #Content></div>
	} @else {
		<div #Content></div>
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

func TestAnalyzer_NamedRefInLoop(t *testing.T) {
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
			<li #Items>{item}</li>
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
			<li #Items key={item.ID}>{item.Name}</li>
		}
	</ul>
}`,
			wantError: false,
		},
		"ref with key outside loop is invalid": {
			input: `package x
templ Test() {
	<div #Content key={someKey}></div>
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

func TestAnalyzer_NamedRefInConditional(t *testing.T) {
	input := `package x
templ Test(show bool) {
	<div>
		@if show {
			<span #Label>hello</span>
		}
	</div>
}`

	_, err := AnalyzeFile("test.gsx", input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Ref inside conditional is valid, it just may be nil at runtime
}

func TestAnalyzer_CollectNamedRefs(t *testing.T) {
	input := `package x
templ Test(items []Item, show bool) {
	<div>
		<div #Header></div>
		@if show {
			<span #Label>hello</span>
		}
		@for _, item := range items {
			<li #Items>{item.Name}</li>
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
	refs := analyzer.CollectNamedRefs(file.Components[0])

	if len(refs) != 3 {
		t.Fatalf("expected 3 refs, got %d", len(refs))
	}

	// Check Header ref
	if refs[0].Name != "Header" {
		t.Errorf("refs[0].Name = %q, want 'Header'", refs[0].Name)
	}
	if refs[0].InLoop || refs[0].InConditional {
		t.Error("Header should not be in loop or conditional")
	}

	// Check Label ref (in conditional)
	if refs[1].Name != "Label" {
		t.Errorf("refs[1].Name = %q, want 'Label'", refs[1].Name)
	}
	if refs[1].InLoop {
		t.Error("Label should not be in loop")
	}
	if !refs[1].InConditional {
		t.Error("Label should be in conditional")
	}

	// Check Items ref (in loop)
	if refs[2].Name != "Items" {
		t.Errorf("refs[2].Name = %q, want 'Items'", refs[2].Name)
	}
	if !refs[2].InLoop {
		t.Error("Items should be in loop")
	}
}
