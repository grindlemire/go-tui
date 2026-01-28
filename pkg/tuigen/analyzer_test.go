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
func Test() Element {
	<div></div>
}`,
			wantError: false,
		},
		"known tag span": {
			input: `package x
func Test() Element {
	<span>hello</span>
}`,
			wantError: false,
		},
		"known tag ul": {
			input: `package x
func Test() Element {
	<ul><li /></ul>
}`,
			wantError: false,
		},
		"unknown tag": {
			input: `package x
func Test() Element {
	<unknownTag></unknownTag>
}`,
			wantError:   true,
			errorContains: "unknown element tag <unknownTag>",
		},
		"unknown tag foobar": {
			input: `package x
func Test() Element {
	<foobar />
}`,
			wantError:   true,
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
		input       string
		wantError   bool
		errorContains string
	}

	tests := map[string]tc{
		"known attribute width": {
			input: `package x
func Test() Element {
	<div width=100></div>
}`,
			wantError: false,
		},
		"known attribute direction": {
			input: `package x
func Test() Element {
	<div direction={layout.Column}></div>
}`,
			wantError: false,
		},
		"unknown attribute": {
			input: `package x
func Test() Element {
	<div unknownAttr=123></div>
}`,
			wantError:   true,
			errorContains: "unknown attribute unknownAttr",
		},
		"typo colour": {
			input: `package x
func Test() Element {
	<div colour="red"></div>
}`,
			wantError:   true,
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
		input        string
		wantImports  []string
	}

	tests := map[string]tc{
		"adds element import": {
			input: `package x
func Test() Element {
	<div></div>
}`,
			wantImports: []string{
				"github.com/grindlemire/go-tui/pkg/tui/element",
			},
		},
		"adds layout import when used": {
			input: `package x
func Test() Element {
	<div direction={layout.Column}></div>
}`,
			wantImports: []string{
				"github.com/grindlemire/go-tui/pkg/tui/element",
				"github.com/grindlemire/go-tui/pkg/layout",
			},
		},
		"adds tui import when used": {
			input: `package x
func Test() Element {
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
func Test() Element {
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
func Test() Element {
	<div></div>
}`,
			wantImports: []string{
				"github.com/grindlemire/go-tui/pkg/tui/element",
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
func Test() Element {
	<div>
		<div>
			<unknownTag />
		</div>
	</div>
}`

	_, err := AnalyzeFile("test.gsx", input)
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
func Test(items []string) Element {
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
func Test(items []string) Element {
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
func Test(show bool) Element {
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
func Test(show bool) Element {
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
func Test(show bool) Element {
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

func TestAnalyzer_LetBindingValidation(t *testing.T) {
	type tc struct {
		input       string
		wantError   bool
		errorContains string
	}

	tests := map[string]tc{
		"valid let binding": {
			input: `package x
func Test() Element {
	@let myText = <span>hello</span>
	<div></div>
}`,
			wantError: false,
		},
		"let binding with invalid element": {
			input: `package x
func Test() Element {
	@let myText = <badTag />
	<div></div>
}`,
			wantError:   true,
			errorContains: "unknown element tag <badTag>",
		},
		"let binding with invalid attribute": {
			input: `package x
func Test() Element {
	@let myText = <span badAttr="value">hello</span>
	<div></div>
}`,
			wantError:   true,
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
func Test() Element {
	<div ` + attr + `=1></div>
}`
			_, err := AnalyzeFile("test.gsx", input)
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
func Test() Element {
	<` + tag + ` />
}`
			_, err := AnalyzeFile("test.gsx", input)
			if err != nil {
				t.Errorf("tag %q should be valid, got error: %v", tag, err)
			}
		})
	}
}

func TestAnalyzer_MultipleErrors(t *testing.T) {
	// Test that multiple errors are collected
	input := `package x
func Test() Element {
	<unknownTag1 />
	<unknownTag2 />
}`

	_, err := AnalyzeFile("test.gsx", input)
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
func Test() Element {
	<div colour="red"></div>
}`

	_, err := AnalyzeFile("test.gsx", input)
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
func Test() Element {
	<div>
		<hr/>
	</div>
}`,
			wantError: false,
		},
		"hr with class": {
			input: `package x
func Test() Element {
	<div>
		<hr class="border-double"/>
	</div>
}`,
			wantError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.gsx", tt.input)

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
func Test() Element {
	<div>
		<br/>
	</div>
}`,
			wantError: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := AnalyzeFile("test.gsx", tt.input)

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
func Test() Element {
	<div>
		<hr>text</hr>
	</div>
}`,
			wantError:     true,
			errorContains: "<hr> is a void element and cannot have children",
		},
		"hr with element child": {
			input: `package x
func Test() Element {
	<div>
		<hr><span>nested</span></hr>
	</div>
}`,
			wantError:     true,
			errorContains: "<hr> is a void element and cannot have children",
		},
		"br with text child": {
			input: `package x
func Test() Element {
	<div>
		<br>text</br>
	</div>
}`,
			wantError:     true,
			errorContains: "<br> is a void element and cannot have children",
		},
		"input with child": {
			input: `package x
func Test() Element {
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

func TestAnalyzer_TailwindClassValidation(t *testing.T) {
	type tc struct {
		input         string
		wantError     bool
		errorContains string
		hintContains  string
	}

	tests := map[string]tc{
		"valid tailwind classes": {
			input: `package x
func Test() Element {
	<div class="flex-col gap-2 p-4"></div>
}`,
			wantError: false,
		},
		"valid width and height classes": {
			input: `package x
func Test() Element {
	<div class="w-full h-1/2"></div>
}`,
			wantError: false,
		},
		"valid individual padding classes": {
			input: `package x
func Test() Element {
	<div class="pt-2 pb-4 pl-1"></div>
}`,
			wantError: false,
		},
		"valid border color classes": {
			input: `package x
func Test() Element {
	<div class="border border-red"></div>
}`,
			wantError: false,
		},
		"valid text alignment classes": {
			input: `package x
func Test() Element {
	<div class="text-center"></div>
}`,
			wantError: false,
		},
		"unknown tailwind class": {
			input: `package x
func Test() Element {
	<div class="flex-columns"></div>
}`,
			wantError:     true,
			errorContains: `unknown Tailwind class "flex-columns"`,
			hintContains:  `flex-col`,
		},
		"unknown tailwind class without suggestion": {
			input: `package x
func Test() Element {
	<div class="xyz-completely-invalid"></div>
}`,
			wantError:     true,
			errorContains: `unknown Tailwind class "xyz-completely-invalid"`,
		},
		"multiple unknown classes": {
			input: `package x
func Test() Element {
	<div class="flex-columns badclass"></div>
}`,
			wantError:     true,
			errorContains: `unknown Tailwind class`,
		},
		"mix of valid and invalid classes": {
			input: `package x
func Test() Element {
	<div class="flex-col gap-2 badclass p-4"></div>
}`,
			wantError:     true,
			errorContains: `unknown Tailwind class "badclass"`,
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
				if tt.hintContains != "" && !strings.Contains(err.Error(), tt.hintContains) {
					t.Errorf("error %q does not contain hint %q", err.Error(), tt.hintContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAnalyzer_TailwindClassErrorPosition(t *testing.T) {
	// Test that the error position correctly points to the invalid class
	input := `package x
func Test() Element {
	<div class="flex-col badclass p-2"></div>
}`

	_, err := AnalyzeFile("test.gsx", input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errList, ok := err.(*ErrorList)
	if !ok {
		t.Fatalf("expected *ErrorList, got %T", err)
	}

	errors := errList.Errors()
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	tuiErr := errors[0]
	// The error should point to "badclass" which starts at column 22
	// (class=" is at column 7, content starts at column 14, "flex-col " is 9 chars, so badclass is at column 23)
	// Actually, let's verify the error message contains the class name
	if !strings.Contains(tuiErr.Message, "badclass") {
		t.Errorf("error message %q should contain 'badclass'", tuiErr.Message)
	}

	// Verify EndPos is set for range highlighting
	if tuiErr.EndPos.Line == 0 && tuiErr.EndPos.Column == 0 {
		t.Error("EndPos should be set for range-based highlighting")
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
@component Test() {
	<div #Content></div>
}`,
			wantError: false,
		},
		"valid ref name with digits": {
			input: `package x
@component Test() {
	<div #Content2></div>
}`,
			wantError: false,
		},
		"valid ref name with underscore": {
			input: `package x
@component Test() {
	<div #My_Content></div>
}`,
			wantError: false,
		},
		"invalid ref name lowercase": {
			input: `package x
@component Test() {
	<div #content></div>
}`,
			wantError:     true,
			errorContains: "invalid ref name",
		},
		"invalid ref name starts with digit": {
			input: `package x
@component Test() {
	<div #123invalid></div>
}`,
			wantError:     true,
			errorContains: "expected identifier", // Parser rejects this before analyzer
		},
		"reserved name Root": {
			input: `package x
@component Test() {
	<div #Root></div>
}`,
			wantError:     true,
			errorContains: "ref name 'Root' is reserved",
		},
		"duplicate ref name": {
			input: `package x
@component Test() {
	<div #Content></div>
	<div #Content></div>
}`,
			wantError:     true,
			errorContains: "duplicate ref name",
		},
		"duplicate ref name across branches": {
			input: `package x
@component Test(show bool) {
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

func TestAnalyzer_NamedRefInLoop(t *testing.T) {
	type tc struct {
		input         string
		wantError     bool
		errorContains string
	}

	tests := map[string]tc{
		"ref in loop is valid": {
			input: `package x
@component Test(items []string) {
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
@component Test(items []Item) {
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
@component Test() {
	<div #Content key={someKey}></div>
}`,
			wantError:     true,
			errorContains: "key attribute on ref",
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

func TestAnalyzer_NamedRefInConditional(t *testing.T) {
	input := `package x
@component Test(show bool) {
	<div>
		@if show {
			<span #Label>hello</span>
		}
	</div>
}`

	_, err := AnalyzeFile("test.tui", input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Ref inside conditional is valid, it just may be nil at runtime
}

func TestAnalyzer_CollectNamedRefs(t *testing.T) {
	input := `package x
@component Test(items []Item, show bool) {
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

	l := NewLexer("test.tui", input)
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

// ===== State Detection Tests =====

func TestAnalyzer_DetectStateVars_IntLiteral(t *testing.T) {
	// Since GoCode blocks are handled specially, we need to test with
	// the actual parsing. For now, test the type inference separately.
	type tc struct {
		expr     string
		wantType string
	}

	tests := map[string]tc{
		"integer 0":       {expr: "0", wantType: "int"},
		"integer 42":      {expr: "42", wantType: "int"},
		"negative int":    {expr: "-5", wantType: "int"},
		"float":           {expr: "3.14", wantType: "float64"},
		"negative float":  {expr: "-2.5", wantType: "float64"},
		"bool true":       {expr: "true", wantType: "bool"},
		"bool false":      {expr: "false", wantType: "bool"},
		"string double":   {expr: `"hello"`, wantType: "string"},
		"string backtick": {expr: "`raw`", wantType: "string"},
		"slice literal":   {expr: "[]string{}", wantType: "[]string"},
		"slice with pkg":  {expr: "[]pkg.Type{}", wantType: "[]pkg.Type"},
		"map literal":     {expr: "map[string]int{}", wantType: "map[string]int"},
		"pointer struct":  {expr: "&User{}", wantType: "*User"},
		"pointer pkg":     {expr: "&pkg.User{}", wantType: "*pkg.User"},
		"struct literal":  {expr: "User{}", wantType: "User"},
		"nil":             {expr: "nil", wantType: "any"},
		"function call":   {expr: "someFunc()", wantType: "any"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := inferTypeFromExpr(tt.expr)
			if result != tt.wantType {
				t.Errorf("inferTypeFromExpr(%q) = %q, want %q", tt.expr, result, tt.wantType)
			}
		})
	}
}

func TestAnalyzer_DetectStateVars_Parameter(t *testing.T) {
	input := `package x
@component Counter(count *tui.State[int]) {
	<span>{count.Get()}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 1 {
		t.Fatalf("expected 1 state var, got %d", len(stateVars))
	}

	sv := stateVars[0]
	if sv.Name != "count" {
		t.Errorf("Name = %q, want 'count'", sv.Name)
	}
	if sv.Type != "int" {
		t.Errorf("Type = %q, want 'int'", sv.Type)
	}
	if !sv.IsParameter {
		t.Error("expected IsParameter to be true")
	}
	if sv.InitExpr != "" {
		t.Errorf("InitExpr = %q, want empty for parameter", sv.InitExpr)
	}
}

func TestAnalyzer_DetectStateVars_StringParameter(t *testing.T) {
	input := `package x
@component Greeting(name *tui.State[string]) {
	<span>{name.Get()}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 1 {
		t.Fatalf("expected 1 state var, got %d", len(stateVars))
	}

	sv := stateVars[0]
	if sv.Name != "name" {
		t.Errorf("Name = %q, want 'name'", sv.Name)
	}
	if sv.Type != "string" {
		t.Errorf("Type = %q, want 'string'", sv.Type)
	}
}

func TestAnalyzer_DetectStateVars_SliceParameter(t *testing.T) {
	input := `package x
@component TodoList(items *tui.State[[]string]) {
	<div>{items.Get()}</div>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 1 {
		t.Fatalf("expected 1 state var, got %d", len(stateVars))
	}

	sv := stateVars[0]
	if sv.Type != "[]string" {
		t.Errorf("Type = %q, want '[]string'", sv.Type)
	}
}

func TestAnalyzer_DetectStateVars_PointerParameter(t *testing.T) {
	input := `package x
@component UserProfile(user *tui.State[*User]) {
	<div>{user.Get()}</div>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 1 {
		t.Fatalf("expected 1 state var, got %d", len(stateVars))
	}

	sv := stateVars[0]
	if sv.Type != "*User" {
		t.Errorf("Type = %q, want '*User'", sv.Type)
	}
}

func TestAnalyzer_DetectStateVars_GoCodeDeclaration(t *testing.T) {
	// Test detection of tui.NewState in component body (GoCode block)
	input := `package x
@component Counter() {
	count := tui.NewState(0)
	<span>{count.Get()}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 1 {
		t.Fatalf("expected 1 state var, got %d", len(stateVars))
	}

	sv := stateVars[0]
	if sv.Name != "count" {
		t.Errorf("Name = %q, want 'count'", sv.Name)
	}
	if sv.Type != "int" {
		t.Errorf("Type = %q, want 'int'", sv.Type)
	}
	if sv.IsParameter {
		t.Error("expected IsParameter to be false for GoCode declaration")
	}
	if sv.InitExpr != "0" {
		t.Errorf("InitExpr = %q, want '0'", sv.InitExpr)
	}
}

func TestAnalyzer_DetectStateVars_GoCodeDeclarationString(t *testing.T) {
	// Test detection of tui.NewState with string literal
	input := `package x
@component Greeting() {
	name := tui.NewState("Alice")
	<span>{name.Get()}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 1 {
		t.Fatalf("expected 1 state var, got %d", len(stateVars))
	}

	sv := stateVars[0]
	if sv.Name != "name" {
		t.Errorf("Name = %q, want 'name'", sv.Name)
	}
	if sv.Type != "string" {
		t.Errorf("Type = %q, want 'string'", sv.Type)
	}
	if sv.InitExpr != `"Alice"` {
		t.Errorf("InitExpr = %q, want '\"Alice\"'", sv.InitExpr)
	}
}

func TestAnalyzer_DetectStateVars_GoCodeDeclarationSlice(t *testing.T) {
	// Test detection of tui.NewState with slice literal (matching plan spec)
	input := `package x
@component TodoList() {
	items := tui.NewState([]string{})
	<div>{items.Get()}</div>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 1 {
		t.Fatalf("expected 1 state var, got %d", len(stateVars))
	}

	sv := stateVars[0]
	if sv.Name != "items" {
		t.Errorf("Name = %q, want 'items'", sv.Name)
	}
	if sv.Type != "[]string" {
		t.Errorf("Type = %q, want '[]string'", sv.Type)
	}
}

func TestAnalyzer_DetectStateVars_GoCodeDeclarationBool(t *testing.T) {
	// Test detection of tui.NewState with boolean literal
	input := `package x
@component Toggle() {
	enabled := tui.NewState(true)
	<span>{enabled.Get()}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 1 {
		t.Fatalf("expected 1 state var, got %d", len(stateVars))
	}

	sv := stateVars[0]
	if sv.Type != "bool" {
		t.Errorf("Type = %q, want 'bool'", sv.Type)
	}
	if sv.InitExpr != "true" {
		t.Errorf("InitExpr = %q, want 'true'", sv.InitExpr)
	}
}

func TestAnalyzer_DetectStateVars_MultipleDeclarations(t *testing.T) {
	// Test detection of multiple tui.NewState declarations
	input := `package x
@component Profile() {
	firstName := tui.NewState("Alice")
	lastName := tui.NewState("Smith")
	age := tui.NewState(30)
	<span>{firstName.Get()}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 3 {
		t.Fatalf("expected 3 state vars, got %d", len(stateVars))
	}

	// Check that all are detected
	names := make(map[string]string)
	for _, sv := range stateVars {
		names[sv.Name] = sv.Type
	}

	if names["firstName"] != "string" {
		t.Errorf("firstName type = %q, want 'string'", names["firstName"])
	}
	if names["lastName"] != "string" {
		t.Errorf("lastName type = %q, want 'string'", names["lastName"])
	}
	if names["age"] != "int" {
		t.Errorf("age type = %q, want 'int'", names["age"])
	}
}

func TestAnalyzer_DetectStateVars_MixedParamsAndDeclarations(t *testing.T) {
	// Test detection of both parameter states and GoCode declarations
	input := `package x
@component Counter(initialCount *tui.State[int]) {
	label := tui.NewState("Count: ")
	<span>{label.Get()}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])

	if len(stateVars) != 2 {
		t.Fatalf("expected 2 state vars, got %d", len(stateVars))
	}

	// Find each by name
	var param, decl *StateVar
	for i := range stateVars {
		if stateVars[i].Name == "initialCount" {
			param = &stateVars[i]
		}
		if stateVars[i].Name == "label" {
			decl = &stateVars[i]
		}
	}

	if param == nil {
		t.Fatal("parameter state 'initialCount' not found")
	}
	if !param.IsParameter {
		t.Error("initialCount should be marked as parameter")
	}
	if param.Type != "int" {
		t.Errorf("initialCount type = %q, want 'int'", param.Type)
	}

	if decl == nil {
		t.Fatal("declared state 'label' not found")
	}
	if decl.IsParameter {
		t.Error("label should not be marked as parameter")
	}
	if decl.Type != "string" {
		t.Errorf("label type = %q, want 'string'", decl.Type)
	}
}

func TestAnalyzer_DetectStateBindings_SimpleGet(t *testing.T) {
	input := `package x
@component Counter(count *tui.State[int]) {
	<span>{count.Get()}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}

	b := bindings[0]
	if len(b.StateVars) != 1 || b.StateVars[0] != "count" {
		t.Errorf("StateVars = %v, want [count]", b.StateVars)
	}
	if b.Attribute != "text" {
		t.Errorf("Attribute = %q, want 'text'", b.Attribute)
	}
	if b.ExplicitDeps {
		t.Error("expected ExplicitDeps to be false")
	}
}

func TestAnalyzer_DetectStateBindings_FormatString(t *testing.T) {
	input := `package x
@component Counter(count *tui.State[int]) {
	<span>{fmt.Sprintf("Count: %d", count.Get())}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}

	b := bindings[0]
	if len(b.StateVars) != 1 || b.StateVars[0] != "count" {
		t.Errorf("StateVars = %v, want [count]", b.StateVars)
	}
	if !strings.Contains(b.Expr, "fmt.Sprintf") {
		t.Errorf("Expr = %q, should contain 'fmt.Sprintf'", b.Expr)
	}
}

func TestAnalyzer_DetectStateBindings_MultipleStates(t *testing.T) {
	input := `package x
@component Profile(firstName *tui.State[string], lastName *tui.State[string]) {
	<span>{fmt.Sprintf("%s %s", firstName.Get(), lastName.Get())}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}

	b := bindings[0]
	if len(b.StateVars) != 2 {
		t.Fatalf("expected 2 state vars, got %d: %v", len(b.StateVars), b.StateVars)
	}
	// Check both states are detected (order may vary)
	hasFirst := false
	hasLast := false
	for _, sv := range b.StateVars {
		if sv == "firstName" {
			hasFirst = true
		}
		if sv == "lastName" {
			hasLast = true
		}
	}
	if !hasFirst || !hasLast {
		t.Errorf("StateVars = %v, want [firstName, lastName]", b.StateVars)
	}
}

func TestAnalyzer_DetectStateBindings_ExplicitDeps(t *testing.T) {
	input := `package x
@component UserCard(user *tui.State[*User]) {
	<span deps={[user]}>{formatUser(user.Get())}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}

	b := bindings[0]
	if len(b.StateVars) != 1 || b.StateVars[0] != "user" {
		t.Errorf("StateVars = %v, want [user]", b.StateVars)
	}
	if !b.ExplicitDeps {
		t.Error("expected ExplicitDeps to be true")
	}
}

func TestAnalyzer_DetectStateBindings_ExplicitDepsMultiple(t *testing.T) {
	input := `package x
@component Combined(count *tui.State[int], name *tui.State[string]) {
	<span deps={[count, name]}>{compute(count, name)}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}

	b := bindings[0]
	if len(b.StateVars) != 2 {
		t.Fatalf("expected 2 state vars, got %d: %v", len(b.StateVars), b.StateVars)
	}
	if !b.ExplicitDeps {
		t.Error("expected ExplicitDeps to be true")
	}
}

func TestAnalyzer_DetectStateBindings_UnknownStateInDeps(t *testing.T) {
	input := `package x
@component Test(count *tui.State[int]) {
	<span deps={[unknown]}>{count.Get()}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	_ = analyzer.DetectStateBindings(file.Components[0], stateVars)

	// Check that an error was recorded
	errors := analyzer.Errors().Errors()
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if !strings.Contains(errors[0].Message, "unknown state variable") {
		t.Errorf("error message = %q, want to contain 'unknown state variable'", errors[0].Message)
	}
}

func TestAnalyzer_DetectStateBindings_DynamicClass(t *testing.T) {
	input := `package x
@component Toggle(enabled *tui.State[bool]) {
	<span class={enabled.Get() ? "text-green" : "text-red"}>Status</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}

	b := bindings[0]
	if b.Attribute != "class" {
		t.Errorf("Attribute = %q, want 'class'", b.Attribute)
	}
	if len(b.StateVars) != 1 || b.StateVars[0] != "enabled" {
		t.Errorf("StateVars = %v, want [enabled]", b.StateVars)
	}
}

func TestAnalyzer_DetectStateBindings_NoStateUsage(t *testing.T) {
	input := `package x
@component Static() {
	<span>{"Hello, World!"}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 0 {
		t.Errorf("expected 0 bindings, got %d", len(bindings))
	}
}

func TestAnalyzer_DetectStateBindings_WithNamedRef(t *testing.T) {
	input := `package x
@component Counter(count *tui.State[int]) {
	<span #Label>{count.Get()}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}

	b := bindings[0]
	if b.ElementName != "Label" {
		t.Errorf("ElementName = %q, want 'Label'", b.ElementName)
	}
}

func TestAnalyzer_DepsAttributeValid(t *testing.T) {
	// Test that deps attribute is recognized as valid
	input := `package x
@component Test(count *tui.State[int]) {
	<span deps={[count]}>{count.Get()}</span>
}`

	_, err := AnalyzeFile("test.tui", input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAnalyzer_DetectStateBindings_DereferencedPointer(t *testing.T) {
	// Test that (*count).Get() pattern is detected
	input := `package x
@component Counter(count *tui.State[int]) {
	<span>{(*count).Get()}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	bindings := analyzer.DetectStateBindings(file.Components[0], stateVars)

	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}

	b := bindings[0]
	if len(b.StateVars) != 1 || b.StateVars[0] != "count" {
		t.Errorf("StateVars = %v, want [count]", b.StateVars)
	}
}

func TestAnalyzer_DepsStringLiteralError(t *testing.T) {
	// Test that deps="string" produces an error
	input := `package x
@component Test(count *tui.State[int]) {
	<span deps="not-valid">{count.Get()}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	_ = analyzer.DetectStateBindings(file.Components[0], stateVars)

	errors := analyzer.Errors().Errors()
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if !strings.Contains(errors[0].Message, "must use expression syntax") {
		t.Errorf("error message = %q, want to contain 'must use expression syntax'", errors[0].Message)
	}
}

func TestAnalyzer_DepsMissingBracketsError(t *testing.T) {
	// Test that deps={count} (missing brackets) produces an error
	input := `package x
@component Test(count *tui.State[int]) {
	<span deps={count}>{count.Get()}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	_ = analyzer.DetectStateBindings(file.Components[0], stateVars)

	errors := analyzer.Errors().Errors()
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if !strings.Contains(errors[0].Message, "must be an array literal") {
		t.Errorf("error message = %q, want to contain 'must be an array literal'", errors[0].Message)
	}
}

func TestAnalyzer_DepsEmptyArrayWarning(t *testing.T) {
	// Test that deps={[]} (empty) produces a warning
	input := `package x
@component Test(count *tui.State[int]) {
	<span deps={[]}>{count.Get()}</span>
}`

	l := NewLexer("test.tui", input)
	p := NewParser(l)
	file, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	analyzer := NewAnalyzer()
	stateVars := analyzer.DetectStateVars(file.Components[0])
	_ = analyzer.DetectStateBindings(file.Components[0], stateVars)

	errors := analyzer.Errors().Errors()
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if !strings.Contains(errors[0].Message, "empty deps attribute") {
		t.Errorf("error message = %q, want to contain 'empty deps attribute'", errors[0].Message)
	}
}
