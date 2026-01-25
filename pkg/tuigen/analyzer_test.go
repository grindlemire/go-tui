package tuigen

import (
	"strings"
	"testing"
)

func TestAnalyzer_UnknownElementTag(t *testing.T) {
	type tc struct {
		input       string
		wantError   bool
		errorContains string
	}

	tests := map[string]tc{
		"known tag box": {
			input: `package x
@component Test() {
	<box></box>
}`,
			wantError: false,
		},
		"known tag text": {
			input: `package x
@component Test() {
	<text>hello</text>
}`,
			wantError: false,
		},
		"known tag scrollable": {
			input: `package x
@component Test() {
	<scrollable></scrollable>
}`,
			wantError: false,
		},
		"unknown tag": {
			input: `package x
@component Test() {
	<unknownTag></unknownTag>
}`,
			wantError:   true,
			errorContains: "unknown element tag <unknownTag>",
		},
		"unknown tag foobar": {
			input: `package x
@component Test() {
	<foobar />
}`,
			wantError:   true,
			errorContains: "unknown element tag <foobar>",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.tui", tt.input)

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
		input       string
		wantError   bool
		errorContains string
	}

	tests := map[string]tc{
		"known attribute width": {
			input: `package x
@component Test() {
	<box width=100></box>
}`,
			wantError: false,
		},
		"known attribute direction": {
			input: `package x
@component Test() {
	<box direction={layout.Column}></box>
}`,
			wantError: false,
		},
		"unknown attribute": {
			input: `package x
@component Test() {
	<box unknownAttr=123></box>
}`,
			wantError:   true,
			errorContains: "unknown attribute unknownAttr",
		},
		"typo colour": {
			input: `package x
@component Test() {
	<box colour="red"></box>
}`,
			wantError:   true,
			errorContains: "unknown attribute colour",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.tui", tt.input)

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
		input        string
		wantImports  []string
	}

	tests := map[string]tc{
		"adds element import": {
			input: `package x
@component Test() {
	<box></box>
}`,
			wantImports: []string{
				"github.com/grindlemire/go-tui/pkg/tui/element",
			},
		},
		"adds layout import when used": {
			input: `package x
@component Test() {
	<box direction={layout.Column}></box>
}`,
			wantImports: []string{
				"github.com/grindlemire/go-tui/pkg/tui/element",
				"github.com/grindlemire/go-tui/pkg/layout",
			},
		},
		"adds tui import when used": {
			input: `package x
@component Test() {
	<box border={tui.BorderSingle}></box>
}`,
			wantImports: []string{
				"github.com/grindlemire/go-tui/pkg/tui/element",
				"github.com/grindlemire/go-tui/pkg/tui",
			},
		},
		"preserves existing imports": {
			input: `package x
import "fmt"
@component Test() {
	<text>hello</text>
}`,
			wantImports: []string{
				"fmt",
				"github.com/grindlemire/go-tui/pkg/tui/element",
			},
		},
		"does not duplicate existing element import": {
			input: `package x
import "github.com/grindlemire/go-tui/pkg/tui/element"
@component Test() {
	<box></box>
}`,
			wantImports: []string{
				"github.com/grindlemire/go-tui/pkg/tui/element",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			file, err := AnalyzeFile("test.tui", tt.input)
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
		tag    string
		valid  bool
	}

	tests := map[string]tc{
		"box":        {tag: "box", valid: true},
		"text":       {tag: "text", valid: true},
		"scrollable": {tag: "scrollable", valid: true},
		"button":     {tag: "button", valid: true},
		"input":      {tag: "input", valid: true},
		"list":       {tag: "list", valid: true},
		"table":      {tag: "table", valid: true},
		"progress":   {tag: "progress", valid: true},
		"unknown":    {tag: "unknown", valid: false},
		"div":        {tag: "div", valid: false},
		"span":       {tag: "span", valid: false},
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
		attr   string
		valid  bool
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
		"onEvent":     {attr: "onEvent", valid: true},
		"onFocus":     {attr: "onFocus", valid: true},
		"flexGrow":    {attr: "flexGrow", valid: true},
		"unknown":     {attr: "unknown", valid: false},
		"class":       {attr: "class", valid: false},
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
		"onclick -> onEvent": {
			input:      "onclick",
			suggestion: "onEvent",
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

func TestAnalyzer_NestedElements(t *testing.T) {
	// Test that nested elements are all validated
	input := `package x
@component Test() {
	<box>
		<box>
			<unknownTag />
		</box>
	</box>
}`

	_, err := AnalyzeFile("test.tui", input)
	if err == nil {
		t.Error("expected error for nested unknown tag")
		return
	}

	if !strings.Contains(err.Error(), "unknown element tag <unknownTag>") {
		t.Errorf("error %q does not contain expected message", err.Error())
	}
}

func TestAnalyzer_ControlFlowValidation(t *testing.T) {
	type tc struct {
		input       string
		wantError   bool
		errorContains string
	}

	tests := map[string]tc{
		"valid for loop": {
			input: `package x
@component Test(items []string) {
	<box>
		@for _, item := range items {
			<text>{item}</text>
		}
	</box>
}`,
			wantError: false,
		},
		"invalid element in for loop": {
			input: `package x
@component Test(items []string) {
	<box>
		@for _, item := range items {
			<badTag />
		}
	</box>
}`,
			wantError:   true,
			errorContains: "unknown element tag <badTag>",
		},
		"valid if statement": {
			input: `package x
@component Test(show bool) {
	<box>
		@if show {
			<text>visible</text>
		}
	</box>
}`,
			wantError: false,
		},
		"invalid element in if then": {
			input: `package x
@component Test(show bool) {
	<box>
		@if show {
			<badTag />
		}
	</box>
}`,
			wantError:   true,
			errorContains: "unknown element tag <badTag>",
		},
		"invalid element in if else": {
			input: `package x
@component Test(show bool) {
	<box>
		@if show {
			<text>yes</text>
		} @else {
			<badTag />
		}
	</box>
}`,
			wantError:   true,
			errorContains: "unknown element tag <badTag>",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.tui", tt.input)

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

func TestAnalyzer_LetBindingValidation(t *testing.T) {
	type tc struct {
		input       string
		wantError   bool
		errorContains string
	}

	tests := map[string]tc{
		"valid let binding": {
			input: `package x
@component Test() {
	@let myText = <text>hello</text>
	<box></box>
}`,
			wantError: false,
		},
		"let binding with invalid element": {
			input: `package x
@component Test() {
	@let myText = <badTag />
	<box></box>
}`,
			wantError:   true,
			errorContains: "unknown element tag <badTag>",
		},
		"let binding with invalid attribute": {
			input: `package x
@component Test() {
	@let myText = <text badAttr="value">hello</text>
	<box></box>
}`,
			wantError:   true,
			errorContains: "unknown attribute badAttr",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.tui", tt.input)

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

func TestAnalyzer_AllKnownAttributes(t *testing.T) {
	// Test all known attributes are accepted
	attributes := []string{
		"width", "widthPercent", "height", "heightPercent",
		"minWidth", "minHeight", "maxWidth", "maxHeight",
		"direction", "justify", "align", "gap",
		"flexGrow", "flexShrink", "alignSelf",
		"padding", "margin",
		"border", "borderStyle", "background",
		"text", "textStyle", "textAlign",
		"onFocus", "onBlur", "onEvent",
		"scrollable", "scrollbarStyle", "scrollbarThumbStyle",
		"disabled", "id",
	}

	for _, attr := range attributes {
		t.Run(attr, func(t *testing.T) {
			input := `package x
@component Test() {
	<box ` + attr + `=1></box>
}`
			_, err := AnalyzeFile("test.tui", input)
			if err != nil {
				t.Errorf("attribute %q should be valid, got error: %v", attr, err)
			}
		})
	}
}

func TestAnalyzer_AllKnownTags(t *testing.T) {
	// Test all known tags are accepted
	tags := []string{
		"box", "text", "scrollable", "button",
		"input", "list", "table", "progress",
	}

	for _, tag := range tags {
		t.Run(tag, func(t *testing.T) {
			input := `package x
@component Test() {
	<` + tag + ` />
}`
			_, err := AnalyzeFile("test.tui", input)
			if err != nil {
				t.Errorf("tag %q should be valid, got error: %v", tag, err)
			}
		})
	}
}

func TestAnalyzer_MultipleErrors(t *testing.T) {
	// Test that multiple errors are collected
	input := `package x
@component Test() {
	<unknownTag1 />
	<unknownTag2 />
}`

	_, err := AnalyzeFile("test.tui", input)
	if err == nil {
		t.Fatal("expected errors, got nil")
	}

	errStr := err.Error()

	if !strings.Contains(errStr, "unknownTag1") {
		t.Error("missing error for unknownTag1")
	}

	if !strings.Contains(errStr, "unknownTag2") {
		t.Error("missing error for unknownTag2")
	}
}

func TestAnalyzer_ErrorHint(t *testing.T) {
	input := `package x
@component Test() {
	<box colour="red"></box>
}`

	_, err := AnalyzeFile("test.tui", input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errStr := err.Error()

	// Should have hint about similar attribute
	if !strings.Contains(errStr, "did you mean") {
		t.Errorf("error should contain hint, got: %s", errStr)
	}
}
