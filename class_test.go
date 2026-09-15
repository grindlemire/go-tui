package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/grindlemire/go-tui/internal/tailwind"
)

// classState captures everything a class can touch on an element.
type classState struct {
	Layout              LayoutStyle
	Border              BorderStyle
	BorderStyle         Style
	Background          *Style
	TextStyle           Style
	TextStyleSet        bool
	TextAlign           TextAlign
	Truncate            bool
	Wrap                bool
	Hidden              bool
	Focusable           bool
	Scroll              ScrollMode
	Overflow            OverflowMode
	ScrollbarStyle      Style
	ScrollbarThumbStyle Style
	ScrollbarHidden     bool
	TextGradient        *Gradient
	BgGradient          *Gradient
	BorderGradient      *Gradient
}

func snapshotClassState(e *Element) classState {
	return classState{
		Layout:              e.LayoutStyle(),
		Border:              e.Border(),
		BorderStyle:         e.BorderStyle(),
		Background:          e.Background(),
		TextStyle:           e.TextStyle(),
		TextStyleSet:        e.textStyleSet,
		TextAlign:           e.TextAlign(),
		Truncate:            e.Truncate(),
		Wrap:                e.Wrap(),
		Hidden:              e.Hidden(),
		Focusable:           e.IsFocusable(),
		Scroll:              e.ScrollModeValue(),
		Overflow:            e.Overflow(),
		ScrollbarStyle:      e.scrollbarStyle,
		ScrollbarThumbStyle: e.scrollbarThumbStyle,
		ScrollbarHidden:     e.scrollbarHidden,
		TextGradient:        e.textGradient,
		BgGradient:          e.bgGradient,
		BorderGradient:      e.borderGradient,
	}
}

func TestWithClass_MatchesExplicitOptions(t *testing.T) {
	type tc struct {
		classes string
		opts    []Option
	}

	tests := map[string]tc{
		"block":           {classes: "block", opts: []Option{WithDisplay(DisplayBlock)}},
		"flex row":        {classes: "flex", opts: []Option{WithDisplay(DisplayFlex), WithDirection(Row)}},
		"flex-col":        {classes: "flex-col", opts: []Option{WithDisplay(DisplayFlex), WithDirection(Column)}},
		"flex wrap":       {classes: "flex-wrap-reverse", opts: []Option{WithFlexWrap(WrapReverse)}},
		"align content":   {classes: "content-between", opts: []Option{WithAlignContent(ContentSpaceBetween)}},
		"flex-1":          {classes: "flex-1", opts: []Option{WithFlexGrow(1), WithFlexShrink(1)}},
		"flex-grow-N":     {classes: "flex-grow-3 flex-shrink-0", opts: []Option{WithFlexGrow(3), WithFlexShrink(0)}},
		"justify":         {classes: "justify-evenly", opts: []Option{WithJustify(JustifySpaceEvenly)}},
		"items":           {classes: "items-end", opts: []Option{WithAlign(AlignEnd)}},
		"self":            {classes: "self-center", opts: []Option{WithAlignSelf(AlignCenter)}},
		"text align":      {classes: "text-right", opts: []Option{WithTextAlign(TextAlignRight)}},
		"border":          {classes: "border-thick", opts: []Option{WithBorder(BorderThick)}},
		"border color":    {classes: "border-cyan", opts: []Option{WithBorderStyle(NewStyle().Foreground(Cyan))}},
		"border hex":      {classes: "border-[#0000ff]", opts: []Option{WithBorderStyle(NewStyle().Foreground(RGBColor(0, 0, 255)))}},
		"background":      {classes: "bg-bright-magenta", opts: []Option{WithBackground(NewStyle().Background(BrightMagenta))}},
		"background hex":  {classes: "bg-[#f80]", opts: []Option{WithBackground(NewStyle().Background(RGBColor(255, 136, 0)))}},
		"scroll":          {classes: "overflow-y-scroll", opts: []Option{WithScrollable(ScrollVertical)}},
		"overflow hidden": {classes: "overflow-hidden", opts: []Option{WithOverflow(OverflowHidden)}},
		"focusable":       {classes: "focusable", opts: []Option{WithFocusable(true)}},
		"hidden":          {classes: "hidden", opts: []Option{WithHidden(true)}},
		"truncate":        {classes: "truncate", opts: []Option{WithTruncate(true)}},
		"nowrap":          {classes: "nowrap", opts: []Option{WithWrap(false)}},
		"scrollbar":       {classes: "scrollbar-hidden scrollbar-red scrollbar-thumb-[#abc]", opts: []Option{WithScrollbarHidden(true), WithScrollbarStyle(NewStyle().Foreground(Red)), WithScrollbarThumbStyle(NewStyle().Foreground(RGBColor(0xaa, 0xbb, 0xcc)))}},
		"gap":             {classes: "gap-2", opts: []Option{WithGap(2)}},
		"padding all":     {classes: "p-3", opts: []Option{WithPadding(3)}},
		"margin all":      {classes: "m-2", opts: []Option{WithMargin(2)}},
		"padding sides":   {classes: "pt-1 px-2", opts: []Option{WithPaddingTRBL(1, 2, 0, 2)}},
		"margin sides":    {classes: "mb-3 ml-1", opts: []Option{WithMarginTRBL(0, 0, 3, 1)}},
		"fixed sizes":     {classes: "w-10 h-4", opts: []Option{WithWidth(10), WithHeight(4)}},
		"min max":         {classes: "min-w-1 max-w-20 min-h-2 max-h-8", opts: []Option{WithMinWidth(1), WithMaxWidth(20), WithMinHeight(2), WithMaxHeight(8)}},
		"percent":         {classes: "w-1/2 h-full", opts: []Option{WithWidthPercent(50), WithHeightPercent(100)}},
		"auto":            {classes: "w-auto h-auto", opts: []Option{WithWidthAuto(), WithHeightAuto()}},
		"text gradient":   {classes: "text-gradient-red-blue", opts: []Option{WithTextGradient(NewGradient(Red, Blue).WithDirection(GradientHorizontal))}},
		"bg gradient":     {classes: "bg-gradient-bright-red-bright-blue-v", opts: []Option{WithBackgroundGradient(NewGradient(BrightRed, BrightBlue).WithDirection(GradientVertical))}},
		"border gradient": {classes: "border-gradient-cyan-magenta-dd", opts: []Option{WithBorderGradient(NewGradient(Cyan, Magenta).WithDirection(GradientDiagonalDown))}},
		"text attrs":      {classes: "font-bold font-dim italic underline blink reverse strikethrough", opts: []Option{WithTextStyle(NewStyle().Bold().Dim().Italic().Underline().Blink().Reverse().Strikethrough())}},
		"text color":      {classes: "text-yellow", opts: []Option{WithTextStyle(NewStyle().Foreground(Yellow))}},
		"text hex":        {classes: "text-[#ff8000]", opts: []Option{WithTextStyle(NewStyle().Foreground(RGBColor(255, 128, 0)))}},
		"last color wins": {classes: "text-red text-green", opts: []Option{WithTextStyle(NewStyle().Foreground(Green))}},
		"unknown ignored": {classes: "bogus flex-column", opts: nil},
		"mixed":           {classes: "flex-col gap-1 p-1 border-rounded text-cyan font-bold", opts: []Option{WithDisplay(DisplayFlex), WithDirection(Column), WithGap(1), WithPadding(1), WithBorder(BorderRounded), WithTextStyle(NewStyle().Foreground(Cyan).Bold())}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := snapshotClassState(New(WithClass(tt.classes)))
			want := snapshotClassState(New(tt.opts...))
			if !reflect.DeepEqual(got, want) {
				t.Errorf("WithClass(%q) state =\n  %+v\nwant\n  %+v", tt.classes, got, want)
			}
		})
	}
}

func TestWithClass_RendersLikeExplicitOptions(t *testing.T) {
	classed := New(WithClass("border-rounded p-1 bg-blue text-red font-bold w-12"), WithText("hi"))
	explicit := New(
		WithBorder(BorderRounded), WithPadding(1),
		WithBackground(NewStyle().Background(Blue)),
		WithTextStyle(NewStyle().Foreground(Red).Bold()),
		WithWidth(12), WithText("hi"),
	)

	got := renderElementANSI(classed)
	want := renderElementANSI(explicit)
	if got != want {
		t.Errorf("rendered output differs:\n got: %q\nwant: %q", got, want)
	}
	if !strings.Contains(got, "hi") {
		t.Errorf("rendered output missing text: %q", got)
	}
}

func renderElementANSI(el *Element) string {
	caps := Capabilities{Colors: ColorTrue, TrueColor: true}
	buf, height := renderElementToBuffer(el, 40, caps)
	var rows []string
	esc := newEscBuilder(256)
	for row := range height {
		rows = append(rows, bufferRowToANSI(buf, row, esc, caps))
	}
	return strings.Join(rows, "\n")
}

func TestSetClass(t *testing.T) {
	e := New(WithClass("text-red p-1"))
	e.dirty = false

	e.SetClass("text-green font-bold border")

	if got, want := e.TextStyle(), NewStyle().Foreground(Green).Bold(); got != want {
		t.Errorf("TextStyle after SetClass = %+v, want %+v", got, want)
	}
	if e.Border() != BorderSingle {
		t.Errorf("Border after SetClass = %v, want BorderSingle", e.Border())
	}
	if e.style.Padding != EdgeAll(1) {
		t.Errorf("SetClass should leave properties it does not mention untouched, padding = %+v", e.style.Padding)
	}
	if !e.IsDirty() {
		t.Errorf("SetClass should mark the element dirty")
	}
}

// Every class the shared table knows must apply at runtime without hitting
// the unhandled-op panic. Parameterized forms are sampled. Per-kind behavior
// is pinned by TestWithClass_MatchesExplicitOptions.
func TestWithClass_CoversEveryClass(t *testing.T) {
	classes := append(tailwind.StaticClasses(),
		"gap-1", "p-2", "px-1", "py-1", "pt-1", "pr-1", "pb-1", "pl-1",
		"m-2", "mx-1", "my-1", "mt-1", "mr-1", "mb-1", "ml-1",
		"w-3", "h-3", "min-w-1", "max-w-1", "min-h-1", "max-h-1",
		"w-1/2", "h-1/3", "w-full", "w-auto", "h-full", "h-auto",
		"flex-grow-2", "flex-shrink-2",
		"text-[#abc]", "bg-[#abcdef]", "border-[#123]", "scrollbar-[#123]", "scrollbar-thumb-[#123]",
		"text-gradient-red-blue", "bg-gradient-red-blue-v", "border-gradient-bright-red-bright-blue-dd",
	)
	for _, class := range classes {
		if !tailwind.Known(class) {
			t.Errorf("%q is not a known class", class)
			continue
		}
		New(WithClass(class))
	}
}
