package tailwind

import (
	"reflect"
	"slices"
	"testing"
)

func TestParse(t *testing.T) {
	type tc struct {
		input    string
		wantOps  []Op
		wantText []TextOp
	}

	red := Named("red")
	tests := map[string]tc{
		"empty":           {input: ""},
		"whitespace only": {input: "   \t "},
		"unknown ignored": {input: "bogus flex-col", wantOps: []Op{Display{Flex: true}, Direction{Column: true}}},
		"flex emits display and row": {
			input:   "flex",
			wantOps: []Op{Display{Flex: true}, Direction{Column: false}},
		},
		"block":          {input: "block", wantOps: []Op{Display{Flex: false}}},
		"flex shorthand": {input: "flex-1", wantOps: []Op{FlexGrow{N: 1}, FlexShrink{N: 1}}},
		"flex-grow-N":    {input: "flex-grow-3", wantOps: []Op{FlexGrow{N: 3}}},
		"justify and align": {
			input:   "justify-between items-center self-end content-around",
			wantOps: []Op{Justify{Value: SpaceBetween}, AlignItems{Value: AlignCenter}, AlignSelf{Value: AlignEnd}, AlignContent{Value: ContentSpaceAround}},
		},
		"wrap modes":         {input: "flex-wrap flex-nowrap flex-wrap-reverse", wantOps: []Op{FlexWrap{Mode: WrapOn}, FlexWrap{Mode: WrapNone}, FlexWrap{Mode: WrapReverse}}},
		"gap padding margin": {input: "gap-2 p-1 m-3", wantOps: []Op{Gap{N: 2}, Padding{N: 1}, Margin{N: 3}}},
		"padding sides accumulate at end": {
			input:   "pt-1 gap-1 px-2 pb-4",
			wantOps: []Op{Gap{N: 1}, PaddingEdges{Top: 1, Right: 2, Bottom: 4, Left: 2}},
		},
		"margin sides accumulate after padding": {
			input:   "mt-3 pt-1 mx-2",
			wantOps: []Op{PaddingEdges{Top: 1}, MarginEdges{Top: 3, Right: 2, Left: 2}},
		},
		"sizes": {
			input:   "w-10 h-5 min-w-1 max-w-20 min-h-2 max-h-9",
			wantOps: []Op{Width{N: 10}, Height{N: 5}, MinWidth{N: 1}, MaxWidth{N: 20}, MinHeight{N: 2}, MaxHeight{N: 9}},
		},
		"fractions and keywords": {
			input:   "w-1/2 h-2/3 w-full h-auto w-auto h-full",
			wantOps: []Op{WidthPercent{Percent: 50}, HeightPercent{Percent: float64(2) / float64(3) * 100}, WidthPercent{Percent: 100}, HeightAuto{}, WidthAuto{}, HeightPercent{Percent: 100}},
		},
		"zero denominator ignored": {input: "w-1/0"},
		"border styles": {
			input:   "border border-rounded border-double border-thick border-single",
			wantOps: []Op{Border{Style: BorderSingle}, Border{Style: BorderRounded}, Border{Style: BorderDouble}, Border{Style: BorderThick}, Border{Style: BorderSingle}},
		},
		"colors": {
			input:   "border-red bg-bright-blue scrollbar-cyan scrollbar-thumb-white",
			wantOps: []Op{BorderColor{Color: red}, Background{Color: Named("bright-blue")}, ScrollbarColor{Color: Named("cyan")}, ScrollbarThumbColor{Color: Named("white")}},
		},
		"hex colors": {
			input:    "text-[#ff8000] bg-[#f80] border-[#0000ff] scrollbar-[#123456] scrollbar-thumb-[#abc]",
			wantOps:  []Op{Background{Color: RGB(255, 136, 0)}, BorderColor{Color: RGB(0, 0, 255)}, ScrollbarColor{Color: RGB(0x12, 0x34, 0x56)}, ScrollbarThumbColor{Color: RGB(0xaa, 0xbb, 0xcc)}},
			wantText: []TextOp{Foreground{Color: RGB(255, 128, 0)}},
		},
		"text style chain in class order": {
			input:    "font-bold text-red italic text-dim underline blink reverse strikethrough font-dim text-green",
			wantText: []TextOp{TextAttr{Attr: Bold}, Foreground{Color: red}, TextAttr{Attr: Italic}, TextAttr{Attr: Dim}, TextAttr{Attr: Underline}, TextAttr{Attr: Blink}, TextAttr{Attr: Reverse}, TextAttr{Attr: Strikethrough}, TextAttr{Attr: Dim}, Foreground{Color: Named("green")}},
		},
		"text align": {input: "text-left text-center text-right", wantOps: []Op{TextAlign{Value: TextLeft}, TextAlign{Value: TextCenter}, TextAlign{Value: TextRight}}},
		"scroll and overflow": {
			input:   "overflow-scroll overflow-y-scroll overflow-x-scroll overflow-hidden scrollbar-hidden",
			wantOps: []Op{Scroll{Mode: ScrollBoth}, Scroll{Mode: ScrollVertical}, Scroll{Mode: ScrollHorizontal}, OverflowHidden{}, ScrollbarHidden{}},
		},
		"flags": {input: "focusable hidden truncate nowrap wrap", wantOps: []Op{Focusable{}, Hidden{}, Truncate{}, Wrap{Enabled: false}, Wrap{Enabled: true}}},
		"gradient default direction": {
			input:   "text-gradient-red-blue",
			wantOps: []Op{TextGradient{Start: red, End: Named("blue"), Direction: Horizontal}},
		},
		"gradient with direction and bright colors": {
			input:   "bg-gradient-bright-red-bright-blue-v border-gradient-cyan-magenta-dd text-gradient-white-black-du bg-gradient-red-green-h",
			wantOps: []Op{BackgroundGradient{Start: Named("bright-red"), End: Named("bright-blue"), Direction: Vertical}, BorderGradient{Start: Named("cyan"), End: Named("magenta"), Direction: DiagonalDown}, TextGradient{Start: Named("white"), End: Named("black"), Direction: DiagonalUp}, BackgroundGradient{Start: red, End: Named("green"), Direction: Horizontal}},
		},
		"gradient unknown colors fall back to black": {
			input:   "text-gradient-foo-bar",
			wantOps: []Op{TextGradient{Start: Named("black"), End: Named("black"), Direction: Horizontal}},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := Parse(tt.input)
			if ops := got.Ops(); !reflect.DeepEqual(ops, tt.wantOps) {
				t.Errorf("Parse(%q).Ops() =\n  %#v\nwant\n  %#v", tt.input, ops, tt.wantOps)
			}
			if text := got.Text(); !reflect.DeepEqual(text, tt.wantText) {
				t.Errorf("Parse(%q).Text() =\n  %#v\nwant\n  %#v", tt.input, text, tt.wantText)
			}
		})
	}
}

func TestParseKeepsClassGrouping(t *testing.T) {
	got := Parse("flex pt-1 gap-2 mx-3")
	want := []Class{
		{Ops: []Op{Display{Flex: true}, Direction{}}},
		{Ops: []Op{Gap{N: 2}}},
		{Ops: []Op{PaddingEdges{Top: 1}}},
		{Ops: []Op{MarginEdges{Right: 3, Left: 3}}},
	}
	if !reflect.DeepEqual(got.Classes, want) {
		t.Errorf("Parse().Classes =\n  %#v\nwant\n  %#v", got.Classes, want)
	}
}

func TestResolveSideSpacing(t *testing.T) {
	type tc struct {
		class string
		want  Op
	}

	tests := map[string]tc{
		"pt": {class: "pt-2", want: PaddingEdges{Top: 2}},
		"px": {class: "px-3", want: PaddingEdges{Right: 3, Left: 3}},
		"py": {class: "py-5", want: PaddingEdges{Top: 5, Bottom: 5}},
		"mb": {class: "mb-4", want: MarginEdges{Bottom: 4}},
		"ml": {class: "ml-1", want: MarginEdges{Left: 1}},
		"my": {class: "my-2", want: MarginEdges{Top: 2, Bottom: 2}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, ok := Resolve(tt.class)
			if !ok {
				t.Fatalf("Resolve(%q) not ok", tt.class)
			}
			if !reflect.DeepEqual(got.Ops, []Op{tt.want}) {
				t.Errorf("Resolve(%q).Ops = %#v, want %#v", tt.class, got.Ops, []Op{tt.want})
			}
		})
	}
}

func TestKnown(t *testing.T) {
	type tc struct {
		class string
		want  bool
	}

	tests := map[string]tc{
		"static":         {class: "flex-col", want: true},
		"dynamic":        {class: "gap-7", want: true},
		"side spacing":   {class: "pt-2", want: true},
		"hex":            {class: "text-[#abc]", want: true},
		"bad hex":        {class: "text-[#abcd]", want: false},
		"unknown":        {class: "flex-column", want: false},
		"empty":          {class: "", want: false},
		"surrounding ws": {class: "  border  ", want: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := Known(tt.class); got != tt.want {
				t.Errorf("Known(%q) = %v, want %v", tt.class, got, tt.want)
			}
		})
	}
}

func TestStaticClasses(t *testing.T) {
	classes := StaticClasses()
	if !slices.IsSorted(classes) {
		t.Errorf("StaticClasses() is not sorted")
	}
	for _, want := range []string{"flex", "border-rounded", "text-bright-cyan", "scrollbar-thumb-black", "truncate"} {
		if !slices.Contains(classes, want) {
			t.Errorf("StaticClasses() missing %q", want)
		}
	}
	for _, c := range classes {
		if !Known(c) {
			t.Errorf("static class %q not Known", c)
		}
		r, _ := Resolve(c)
		if len(r.Text) > 1 || (len(r.Text) > 0 && len(r.Ops) > 0) {
			t.Errorf("static class %q must carry either ops or one text op, got %d ops and %d text ops", c, len(r.Ops), len(r.Text))
		}
	}
}

func TestColor(t *testing.T) {
	type tc struct {
		hex    string
		want   Color
		wantOK bool
	}

	tests := map[string]tc{
		"six digit":            {hex: "ff8000", want: RGB(255, 128, 0), wantOK: true},
		"three digit expanded": {hex: "f80", want: RGB(255, 136, 0), wantOK: true},
		"wrong length":         {hex: "ffff"},
		"empty":                {hex: ""},
		"bad red component":    {hex: "zzff00"},
		"bad green component":  {hex: "ffzz00"},
		"bad blue component":   {hex: "ff00zz"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, ok := ParseHex(tt.hex)
			if ok != tt.wantOK {
				t.Fatalf("ParseHex(%q) ok = %v, want %v", tt.hex, ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Errorf("ParseHex(%q) = %+v, want %+v", tt.hex, got, tt.want)
			}
		})
	}

	if Named("chartreuse") != Named("black") {
		t.Errorf("unknown color name should fall back to black")
	}
	if Named("bright-white").Index != 15 || Named("red").Index != 1 || Named("black").Index != 0 {
		t.Errorf("named colors should map to ANSI-16 indices")
	}
}
