package schema

import (
	"strings"
	"testing"
)

func TestGetElement(t *testing.T) {
	type tc struct {
		tag      string
		wantNil  bool
		wantCat  string
		wantSelf bool
	}

	tests := map[string]tc{
		"div is container": {
			tag:     "div",
			wantCat: "container",
		},
		"span is text": {
			tag:     "span",
			wantCat: "text",
		},
		"p is text": {
			tag:     "p",
			wantCat: "text",
		},
		"button is input": {
			tag:     "button",
			wantCat: "input",
		},
		"input is self-closing": {
			tag:      "input",
			wantCat:  "input",
			wantSelf: true,
		},
		"progress is self-closing display": {
			tag:      "progress",
			wantCat:  "display",
			wantSelf: true,
		},
		"hr is self-closing": {
			tag:      "hr",
			wantCat:  "display",
			wantSelf: true,
		},
		"br is self-closing": {
			tag:      "br",
			wantCat:  "display",
			wantSelf: true,
		},
		"unknown tag returns nil": {
			tag:     "foobar",
			wantNil: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			elem := GetElement(tt.tag)
			if tt.wantNil {
				if elem != nil {
					t.Errorf("expected nil for tag %q, got %+v", tt.tag, elem)
				}
				return
			}
			if elem == nil {
				t.Fatalf("expected non-nil for tag %q", tt.tag)
			}
			if elem.Category != tt.wantCat {
				t.Errorf("tag %q: category = %q, want %q", tt.tag, elem.Category, tt.wantCat)
			}
			if elem.SelfClosing != tt.wantSelf {
				t.Errorf("tag %q: selfClosing = %v, want %v", tt.tag, elem.SelfClosing, tt.wantSelf)
			}
			if elem.Description == "" {
				t.Errorf("tag %q: description should not be empty", tt.tag)
			}
		})
	}
}

func TestGetAttribute(t *testing.T) {
	type tc struct {
		tag     string
		attr    string
		wantNil bool
		wantCat string
	}

	tests := map[string]tc{
		"div has id": {
			tag:     "div",
			attr:    "id",
			wantCat: "generic",
		},
		"div has class": {
			tag:     "div",
			attr:    "class",
			wantCat: "generic",
		},
		"div has padding": {
			tag:     "div",
			attr:    "padding",
			wantCat: "spacing",
		},
		"div has direction": {
			tag:     "div",
			attr:    "direction",
			wantCat: "flex",
		},
		"div has border": {
			tag:     "div",
			attr:    "border",
			wantCat: "visual",
		},
		"div has onClick": {
			tag:     "div",
			attr:    "onClick",
			wantCat: "event",
		},
		"span has text": {
			tag:     "span",
			attr:    "text",
			wantCat: "text",
		},
		"input has value": {
			tag:     "input",
			attr:    "value",
			wantCat: "generic",
		},
		"unknown tag returns nil": {
			tag:     "foobar",
			attr:    "id",
			wantNil: true,
		},
		"unknown attr returns nil": {
			tag:     "div",
			attr:    "nonexistent",
			wantNil: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			attr := GetAttribute(tt.tag, tt.attr)
			if tt.wantNil {
				if attr != nil {
					t.Errorf("expected nil for %s.%s, got %+v", tt.tag, tt.attr, attr)
				}
				return
			}
			if attr == nil {
				t.Fatalf("expected non-nil for %s.%s", tt.tag, tt.attr)
			}
			if attr.Category != tt.wantCat {
				t.Errorf("%s.%s: category = %q, want %q", tt.tag, tt.attr, attr.Category, tt.wantCat)
			}
		})
	}
}

func TestGetEventHandler(t *testing.T) {
	type tc struct {
		name    string
		wantNil bool
	}

	tests := map[string]tc{
		"onClick exists": {
			name: "onClick",
		},
		"onFocus exists": {
			name: "onFocus",
		},
		"onBlur exists": {
			name: "onBlur",
		},
		"onKeyPress exists": {
			name: "onKeyPress",
		},
		"onEvent exists": {
			name: "onEvent",
		},
		"unknown handler returns nil": {
			name:    "onSwipe",
			wantNil: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := GetEventHandler(tt.name)
			if tt.wantNil {
				if handler != nil {
					t.Errorf("expected nil for %q, got %+v", tt.name, handler)
				}
				return
			}
			if handler == nil {
				t.Fatalf("expected non-nil for %q", tt.name)
			}
			if handler.Description == "" {
				t.Errorf("%q: description should not be empty", tt.name)
			}
			if handler.Signature == "" {
				t.Errorf("%q: signature should not be empty", tt.name)
			}
		})
	}
}

func TestIsElementTag(t *testing.T) {
	type tc struct {
		tag  string
		want bool
	}

	tests := map[string]tc{
		"div is element":     {tag: "div", want: true},
		"span is element":    {tag: "span", want: true},
		"foobar is not":      {tag: "foobar", want: false},
		"empty is not":       {tag: "", want: false},
		"MyComp is not":      {tag: "MyComponent", want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsElementTag(tt.tag)
			if got != tt.want {
				t.Errorf("IsElementTag(%q) = %v, want %v", tt.tag, got, tt.want)
			}
		})
	}
}

func TestIsEventHandler(t *testing.T) {
	type tc struct {
		name string
		want bool
	}

	tests := map[string]tc{
		"onClick is event":     {name: "onClick", want: true},
		"onFocus is event":     {name: "onFocus", want: true},
		"class is not event":   {name: "class", want: false},
		"padding is not event": {name: "padding", want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsEventHandler(tt.name)
			if got != tt.want {
				t.Errorf("IsEventHandler(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestGetKeyword(t *testing.T) {
	type tc struct {
		name    string
		wantNil bool
	}

	tests := map[string]tc{
		"templ keyword":     {name: "templ"},
		"@for keyword":      {name: "@for"},
		"for keyword":       {name: "for"},
		"@if keyword":       {name: "@if"},
		"if keyword":        {name: "if"},
		"@else keyword":     {name: "@else"},
		"@let keyword":      {name: "@let"},
		"let keyword":       {name: "let"},
		"package keyword":   {name: "package"},
		"import keyword":    {name: "import"},
		"unknown":           {name: "unknown", wantNil: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			kw := GetKeyword(tt.name)
			if tt.wantNil {
				if kw != nil {
					t.Errorf("expected nil for %q, got %+v", tt.name, kw)
				}
				return
			}
			if kw == nil {
				t.Fatalf("expected non-nil for %q", tt.name)
			}
			if kw.Documentation == "" {
				t.Errorf("%q: documentation should not be empty", tt.name)
			}
		})
	}
}

func TestGetClassDoc(t *testing.T) {
	type tc struct {
		class   string
		wantDoc bool
	}

	tests := map[string]tc{
		"flex":           {class: "flex", wantDoc: true},
		"flex-col":       {class: "flex-col", wantDoc: true},
		"items-center":   {class: "items-center", wantDoc: true},
		"border-rounded": {class: "border-rounded", wantDoc: true},
		"font-bold":      {class: "font-bold", wantDoc: true},
		"gap-2":          {class: "gap-2", wantDoc: true},
		"p-3":            {class: "p-3", wantDoc: true},
		"text-red":       {class: "text-red", wantDoc: true},
		"bg-blue":        {class: "bg-blue", wantDoc: true},
		"w-full":         {class: "w-full", wantDoc: true},
		"grow":           {class: "grow", wantDoc: true},
		"unknown-class":  {class: "unknown-class", wantDoc: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			doc := GetClassDoc(tt.class)
			if tt.wantDoc && doc == "" {
				t.Errorf("GetClassDoc(%q) returned empty, expected documentation", tt.class)
			}
			if !tt.wantDoc && doc != "" {
				t.Errorf("GetClassDoc(%q) returned %q, expected empty", tt.class, doc)
			}
		})
	}
}

func TestMatchClasses(t *testing.T) {
	type tc struct {
		prefix   string
		wantMin  int // minimum expected matches
		wantAll  bool // if true, verify all matches have the prefix
	}

	tests := map[string]tc{
		"empty prefix returns all": {
			prefix:  "",
			wantMin: 50, // we have a lot of classes
		},
		"flex prefix": {
			prefix:  "flex",
			wantMin: 2, // at least flex, flex-col, flex-row
			wantAll: true,
		},
		"border prefix": {
			prefix:  "border",
			wantMin: 4, // border, border-none, border-single, etc.
			wantAll: true,
		},
		"no matches": {
			prefix:  "zzz-nonexistent",
			wantMin: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			matches := MatchClasses(tt.prefix)
			if len(matches) < tt.wantMin {
				t.Errorf("MatchClasses(%q) returned %d matches, want at least %d",
					tt.prefix, len(matches), tt.wantMin)
			}
			if tt.wantAll {
				for _, m := range matches {
					if !strings.HasPrefix(m.Name, tt.prefix) {
						t.Errorf("match %q doesn't have prefix %q", m.Name, tt.prefix)
					}
				}
			}
		})
	}
}

func TestAllElementTags(t *testing.T) {
	tags := AllElementTags()
	if len(tags) != len(Elements) {
		t.Errorf("AllElementTags returned %d tags, want %d", len(tags), len(Elements))
	}

	// Ensure all expected tags are present
	expected := []string{"div", "span", "p", "ul", "li", "button", "input", "table", "progress", "hr", "br"}
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}
	for _, tag := range expected {
		if !tagSet[tag] {
			t.Errorf("expected tag %q not found in AllElementTags()", tag)
		}
	}
}

func TestContainerHasEventAttrs(t *testing.T) {
	// Verify that container elements have event handler attributes
	div := GetElement("div")
	if div == nil {
		t.Fatal("div element not found")
	}

	hasOnClick := false
	hasOnFocus := false
	for _, attr := range div.Attributes {
		if attr.Name == "onClick" {
			hasOnClick = true
		}
		if attr.Name == "onFocus" {
			hasOnFocus = true
		}
	}

	if !hasOnClick {
		t.Error("div should have onClick attribute")
	}
	if !hasOnFocus {
		t.Error("div should have onFocus attribute")
	}
}
