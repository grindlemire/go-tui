package tailwind

import (
	"reflect"
	"testing"
)

func TestSplitHover(t *testing.T) {
	type tc struct {
		input     string
		wantBase  string
		wantHover string
	}

	tests := map[string]tc{
		"no variants":   {input: "flex-col p-1", wantBase: "flex-col p-1"},
		"only variants": {input: "hover:bg-blue hover:font-bold", wantHover: "bg-blue font-bold"},
		"mixed keeps order": {
			input:     "px-1 hover:bg-blue text-white hover:text-black",
			wantBase:  "px-1 text-white",
			wantHover: "bg-blue text-black",
		},
		"extra whitespace": {input: "  p-1\thover:underline  ", wantBase: "p-1", wantHover: "underline"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			base, hover := SplitHover(tt.input)
			if base != tt.wantBase || hover != tt.wantHover {
				t.Errorf("SplitHover(%q) = (%q, %q), want (%q, %q)", tt.input, base, hover, tt.wantBase, tt.wantHover)
			}
		})
	}
}

func TestHoverAllowed(t *testing.T) {
	type tc struct {
		class string
		want  bool
	}

	tests := map[string]tc{
		"background":          {class: "bg-blue", want: true},
		"hex background":      {class: "bg-[#ff0000]", want: true},
		"text color":          {class: "text-red", want: true},
		"text attribute":      {class: "font-bold", want: true},
		"border color":        {class: "border-cyan", want: true},
		"gradient":            {class: "bg-gradient-red-blue", want: true},
		"border gradient":     {class: "border-gradient-red-blue-v", want: true},
		"scrollbar color":     {class: "scrollbar-thumb-red", want: true},
		"padding changes box": {class: "p-2", want: false},
		"border adds a box":   {class: "border-rounded", want: false},
		"hidden":              {class: "hidden", want: false},
		"width":               {class: "w-10", want: false},
		"text alignment":      {class: "text-center", want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c, ok := Resolve(tt.class)
			if !ok {
				t.Fatalf("Resolve(%q) failed", tt.class)
			}
			if got := HoverAllowed(c); got != tt.want {
				t.Errorf("HoverAllowed(%q) = %v, want %v", tt.class, got, tt.want)
			}
		})
	}
}

func TestParseHover_DropsUnknownAndLayoutClasses(t *testing.T) {
	got := ParseHover("bg-blue p-2 bogus font-bold")
	want := Result{Classes: []Class{
		{Ops: []Op{Background{Color: Named("blue")}}},
		{Text: []TextOp{TextAttr{Attr: Bold}}},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseHover = %#v, want %#v", got, want)
	}
}
