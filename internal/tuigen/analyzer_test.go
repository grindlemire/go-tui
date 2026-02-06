package tuigen

import (
	"strings"
	"testing"
)

func TestAnalyzer_UnknownElementTag(t *testing.T) {
	type tc struct {
		input         string
		wantError     bool
		errorContains string
	}

	tests := map[string]tc{
		"known tag div": {
			input: `package x
templ Test() {
	<div></div>
}`,
			wantError: false,
		},
		"known tag span": {
			input: `package x
templ Test() {
	<span>hello</span>
}`,
			wantError: false,
		},
		"known tag ul": {
			input: `package x
templ Test() {
	<ul><li /></ul>
}`,
			wantError: false,
		},
		"unknown tag": {
			input: `package x
templ Test() {
	<unknownTag></unknownTag>
}`,
			wantError:     true,
			errorContains: "unknown element tag <unknownTag>",
		},
		"unknown tag foobar": {
			input: `package x
templ Test() {
	<foobar />
}`,
			wantError:     true,
			errorContains: "unknown element tag <foobar>",
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

func TestAnalyzer_UnknownAttribute(t *testing.T) {
	type tc struct {
		input         string
		wantError     bool
		errorContains string
	}

	tests := map[string]tc{
		"known attribute width": {
			input: `package x
templ Test() {
	<div width=100></div>
}`,
			wantError: false,
		},
		"known attribute direction": {
			input: `package x
templ Test() {
	<div direction={layout.Column}></div>
}`,
			wantError: false,
		},
		"unknown attribute": {
			input: `package x
templ Test() {
	<div unknownAttr=123></div>
}`,
			wantError:     true,
			errorContains: "unknown attribute unknownAttr",
		},
		"typo colour": {
			input: `package x
templ Test() {
	<div colour="red"></div>
}`,
			wantError:     true,
			errorContains: "unknown attribute colour",
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

func TestAnalyzer_ImportInsertion(t *testing.T) {
	type tc struct {
		input       string
		wantImports []string
	}

	tests := map[string]tc{
		"adds root tui import": {
			input: `package x
templ Test() {
	<div></div>
}`,
			wantImports: []string{
				"github.com/grindlemire/go-tui",
			},
		},
		"adds root import when layout used": {
			input: `package x
templ Test() {
	<div direction={tui.Column}></div>
}`,
			wantImports: []string{
				"github.com/grindlemire/go-tui",
			},
		},
		"adds root import when tui used": {
			input: `package x
templ Test() {
	<div border={tui.BorderSingle}></div>
}`,
			wantImports: []string{
				"github.com/grindlemire/go-tui",
			},
		},
		"preserves existing imports": {
			input: `package x
import "fmt"
templ Test() {
	<span>hello</span>
}`,
			wantImports: []string{
				"fmt",
				"github.com/grindlemire/go-tui",
			},
		},
		"does not duplicate existing root import": {
			input: `package x
import tui "github.com/grindlemire/go-tui"
templ Test() {
	<div></div>
}`,
			wantImports: []string{
				"github.com/grindlemire/go-tui",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			file, err := AnalyzeFile("test.gsx", tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check that all expected imports are present
			for _, wantPath := range tt.wantImports {
				found := false
				for _, imp := range file.Imports {
					if imp.Path == wantPath {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("missing import %q", wantPath)
				}
			}

			// Check we don't have more imports than expected
			if len(file.Imports) != len(tt.wantImports) {
				var paths []string
				for _, imp := range file.Imports {
					paths = append(paths, imp.Path)
				}
				t.Errorf("import count = %d, want %d. Imports: %v", len(file.Imports), len(tt.wantImports), paths)
			}
		})
	}
}

func TestAnalyzer_ValidateElement(t *testing.T) {
	type tc struct {
		tag   string
		valid bool
	}

	tests := map[string]tc{
		"div":      {tag: "div", valid: true},
		"span":     {tag: "span", valid: true},
		"p":        {tag: "p", valid: true},
		"ul":       {tag: "ul", valid: true},
		"li":       {tag: "li", valid: true},
		"button":   {tag: "button", valid: true},
		"input":    {tag: "input", valid: true},
		"table":    {tag: "table", valid: true},
		"progress": {tag: "progress", valid: true},
		"unknown":  {tag: "unknown", valid: false},
		"box":      {tag: "box", valid: false},
		"text":     {tag: "text", valid: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := ValidateElement(tt.tag)
			if result != tt.valid {
				t.Errorf("ValidateElement(%q) = %v, want %v", tt.tag, result, tt.valid)
			}
		})
	}
}

func TestAnalyzer_ValidateAttribute(t *testing.T) {
	type tc struct {
		attr  string
		valid bool
	}

	tests := map[string]tc{
		"width":       {attr: "width", valid: true},
		"height":      {attr: "height", valid: true},
		"direction":   {attr: "direction", valid: true},
		"gap":         {attr: "gap", valid: true},
		"padding":     {attr: "padding", valid: true},
		"margin":      {attr: "margin", valid: true},
		"border":      {attr: "border", valid: true},
		"borderStyle": {attr: "borderStyle", valid: true},
		"text":        {attr: "text", valid: true},
		"textStyle":   {attr: "textStyle", valid: true},
		"onFocus":     {attr: "onFocus", valid: true},
		"flexGrow":    {attr: "flexGrow", valid: true},
		"class":       {attr: "class", valid: true},
		"unknown":     {attr: "unknown", valid: false},
		"style":       {attr: "style", valid: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := ValidateAttribute(tt.attr)
			if result != tt.valid {
				t.Errorf("ValidateAttribute(%q) = %v, want %v", tt.attr, result, tt.valid)
			}
		})
	}
}

func TestAnalyzer_SuggestAttribute(t *testing.T) {
	type tc struct {
		input      string
		suggestion string
	}

	tests := map[string]tc{
		"colour -> color/background": {
			input:      "colour",
			suggestion: "color",
		},
		"onfocus -> onFocus": {
			input:      "onfocus",
			suggestion: "onFocus",
		},
		"flexgrow -> flexGrow": {
			input:      "flexgrow",
			suggestion: "flexGrow",
		},
		"textstyle -> textStyle": {
			input:      "textstyle",
			suggestion: "textStyle",
		},
		"no suggestion for random": {
			input:      "randomattr",
			suggestion: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := SuggestAttribute(tt.input)
			if result != tt.suggestion {
				t.Errorf("SuggestAttribute(%q) = %q, want %q", tt.input, result, tt.suggestion)
			}
		})
	}
}
