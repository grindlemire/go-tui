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
		"known tag div": {
			input: `package x
@component Test() {
	<div></div>
}`,
			wantError: false,
		},
		"known tag span": {
			input: `package x
@component Test() {
	<span>hello</span>
}`,
			wantError: false,
		},
		"known tag ul": {
			input: `package x
@component Test() {
	<ul><li /></ul>
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
	<div width=100></div>
}`,
			wantError: false,
		},
		"known attribute direction": {
			input: `package x
@component Test() {
	<div direction={layout.Column}></div>
}`,
			wantError: false,
		},
		"unknown attribute": {
			input: `package x
@component Test() {
	<div unknownAttr=123></div>
}`,
			wantError:   true,
			errorContains: "unknown attribute unknownAttr",
		},
		"typo colour": {
			input: `package x
@component Test() {
	<div colour="red"></div>
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
	<div></div>
}`,
			wantImports: []string{
				"github.com/grindlemire/go-tui/pkg/tui/element",
			},
		},
		"adds layout import when used": {
			input: `package x
@component Test() {
	<div direction={layout.Column}></div>
}`,
			wantImports: []string{
				"github.com/grindlemire/go-tui/pkg/tui/element",
				"github.com/grindlemire/go-tui/pkg/layout",
			},
		},
		"adds tui import when used": {
			input: `package x
@component Test() {
	<div border={tui.BorderSingle}></div>
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
	<span>hello</span>
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
	<div></div>
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
	<div>
		<div>
			<unknownTag />
		</div>
	</div>
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
	<div>
		@for _, item := range items {
			<span>{item}</span>
		}
	</div>
}`,
			wantError: false,
		},
		"invalid element in for loop": {
			input: `package x
@component Test(items []string) {
	<div>
		@for _, item := range items {
			<badTag />
		}
	</div>
}`,
			wantError:   true,
			errorContains: "unknown element tag <badTag>",
		},
		"valid if statement": {
			input: `package x
@component Test(show bool) {
	<div>
		@if show {
			<span>visible</span>
		}
	</div>
}`,
			wantError: false,
		},
		"invalid element in if then": {
			input: `package x
@component Test(show bool) {
	<div>
		@if show {
			<badTag />
		}
	</div>
}`,
			wantError:   true,
			errorContains: "unknown element tag <badTag>",
		},
		"invalid element in if else": {
			input: `package x
@component Test(show bool) {
	<div>
		@if show {
			<span>yes</span>
		} @else {
			<badTag />
		}
	</div>
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
	@let myText = <span>hello</span>
	<div></div>
}`,
			wantError: false,
		},
		"let binding with invalid element": {
			input: `package x
@component Test() {
	@let myText = <badTag />
	<div></div>
}`,
			wantError:   true,
			errorContains: "unknown element tag <badTag>",
		},
		"let binding with invalid attribute": {
			input: `package x
@component Test() {
	@let myText = <span badAttr="value">hello</span>
	<div></div>
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
	<div ` + attr + `=1></div>
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
		"div", "span", "p", "ul", "li",
		"button", "input", "table", "progress",
		"hr", "br",
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
	<div colour="red"></div>
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

func TestAnalyzer_HRValid(t *testing.T) {
	type tc struct {
		input     string
		wantError bool
	}

	tests := map[string]tc{
		"hr self-closing": {
			input: `package x
@component Test() {
	<div>
		<hr/>
	</div>
}`,
			wantError: false,
		},
		"hr with class": {
			input: `package x
@component Test() {
	<div>
		<hr class="border-double"/>
	</div>
}`,
			wantError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.tui", tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAnalyzer_BRValid(t *testing.T) {
	type tc struct {
		input     string
		wantError bool
	}

	tests := map[string]tc{
		"br self-closing": {
			input: `package x
@component Test() {
	<div>
		<br/>
	</div>
}`,
			wantError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.tui", tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAnalyzer_VoidWithChildren(t *testing.T) {
	type tc struct {
		input         string
		wantError     bool
		errorContains string
	}

	tests := map[string]tc{
		"hr with text child": {
			input: `package x
@component Test() {
	<div>
		<hr>text</hr>
	</div>
}`,
			wantError:     true,
			errorContains: "<hr> is a void element and cannot have children",
		},
		"hr with element child": {
			input: `package x
@component Test() {
	<div>
		<hr><span>nested</span></hr>
	</div>
}`,
			wantError:     true,
			errorContains: "<hr> is a void element and cannot have children",
		},
		"br with text child": {
			input: `package x
@component Test() {
	<div>
		<br>text</br>
	</div>
}`,
			wantError:     true,
			errorContains: "<br> is a void element and cannot have children",
		},
		"input with child": {
			input: `package x
@component Test() {
	<div>
		<input>text</input>
	</div>
}`,
			wantError:     true,
			errorContains: "<input> is a void element and cannot have children",
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
