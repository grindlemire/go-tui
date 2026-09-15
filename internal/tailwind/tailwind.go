package tailwind

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Class is the resolution of a single class name. A class carries either Ops
// or exactly one Text op, never both: the compiler renders a text class as a
// single Style method call.
type Class struct {
	Ops  []Op
	Text []TextOp
}

// Result is the resolution of a whole class attribute. Classes are kept in
// source order; per-side spacing classes are folded into one trailing
// PaddingEdges and one trailing MarginEdges class.
type Result struct {
	Classes []Class
}

// Ops flattens every element option in order.
func (r Result) Ops() []Op {
	var ops []Op
	for _, c := range r.Classes {
		ops = append(ops, c.Ops...)
	}
	return ops
}

// Text flattens every text style modifier in order.
func (r Result) Text() []TextOp {
	var text []TextOp
	for _, c := range r.Classes {
		text = append(text, c.Text...)
	}
	return text
}

// static maps exact class names to their ops.
var static = map[string]Class{
	"block":    {Ops: []Op{Display{}}},
	"flex":     {Ops: []Op{Display{Flex: true}, Direction{}}},
	"flex-row": {Ops: []Op{Display{Flex: true}, Direction{}}},
	"flex-col": {Ops: []Op{Display{Flex: true}, Direction{Column: true}}},

	"flex-wrap":         {Ops: []Op{FlexWrap{Mode: WrapOn}}},
	"flex-wrap-reverse": {Ops: []Op{FlexWrap{Mode: WrapReverse}}},
	"flex-nowrap":       {Ops: []Op{FlexWrap{Mode: WrapNone}}},

	"content-start":   {Ops: []Op{AlignContent{Value: ContentStart}}},
	"content-end":     {Ops: []Op{AlignContent{Value: ContentEnd}}},
	"content-center":  {Ops: []Op{AlignContent{Value: ContentCenter}}},
	"content-stretch": {Ops: []Op{AlignContent{Value: ContentStretch}}},
	"content-between": {Ops: []Op{AlignContent{Value: ContentSpaceBetween}}},
	"content-around":  {Ops: []Op{AlignContent{Value: ContentSpaceAround}}},

	"grow":     {Ops: []Op{FlexGrow{N: 1}}},
	"grow-0":   {Ops: []Op{FlexGrow{N: 0}}},
	"shrink":   {Ops: []Op{FlexShrink{N: 1}}},
	"shrink-0": {Ops: []Op{FlexShrink{N: 0}}},

	"flex-1":       {Ops: []Op{FlexGrow{N: 1}, FlexShrink{N: 1}}},
	"flex-auto":    {Ops: []Op{FlexGrow{N: 1}, FlexShrink{N: 1}}},
	"flex-initial": {Ops: []Op{FlexGrow{N: 0}, FlexShrink{N: 1}}},
	"flex-none":    {Ops: []Op{FlexGrow{N: 0}, FlexShrink{N: 0}}},
	"flex-grow":    {Ops: []Op{FlexGrow{N: 1}}},
	"flex-shrink":  {Ops: []Op{FlexShrink{N: 1}}},

	"justify-start":   {Ops: []Op{Justify{Value: JustifyStart}}},
	"justify-center":  {Ops: []Op{Justify{Value: JustifyCenter}}},
	"justify-end":     {Ops: []Op{Justify{Value: JustifyEnd}}},
	"justify-between": {Ops: []Op{Justify{Value: SpaceBetween}}},
	"justify-evenly":  {Ops: []Op{Justify{Value: SpaceEvenly}}},
	"justify-around":  {Ops: []Op{Justify{Value: SpaceAround}}},

	"items-start":   {Ops: []Op{AlignItems{Value: AlignStart}}},
	"items-center":  {Ops: []Op{AlignItems{Value: AlignCenter}}},
	"items-end":     {Ops: []Op{AlignItems{Value: AlignEnd}}},
	"items-stretch": {Ops: []Op{AlignItems{Value: AlignStretch}}},

	"self-start":   {Ops: []Op{AlignSelf{Value: AlignStart}}},
	"self-end":     {Ops: []Op{AlignSelf{Value: AlignEnd}}},
	"self-center":  {Ops: []Op{AlignSelf{Value: AlignCenter}}},
	"self-stretch": {Ops: []Op{AlignSelf{Value: AlignStretch}}},

	"text-left":   {Ops: []Op{TextAlign{Value: TextLeft}}},
	"text-center": {Ops: []Op{TextAlign{Value: TextCenter}}},
	"text-right":  {Ops: []Op{TextAlign{Value: TextRight}}},

	"border":         {Ops: []Op{Border{Style: BorderSingle}}},
	"border-single":  {Ops: []Op{Border{Style: BorderSingle}}},
	"border-rounded": {Ops: []Op{Border{Style: BorderRounded}}},
	"border-double":  {Ops: []Op{Border{Style: BorderDouble}}},
	"border-thick":   {Ops: []Op{Border{Style: BorderThick}}},

	"font-bold":     {Text: []TextOp{TextAttr{Attr: Bold}}},
	"font-dim":      {Text: []TextOp{TextAttr{Attr: Dim}}},
	"text-dim":      {Text: []TextOp{TextAttr{Attr: Dim}}},
	"italic":        {Text: []TextOp{TextAttr{Attr: Italic}}},
	"underline":     {Text: []TextOp{TextAttr{Attr: Underline}}},
	"blink":         {Text: []TextOp{TextAttr{Attr: Blink}}},
	"reverse":       {Text: []TextOp{TextAttr{Attr: Reverse}}},
	"strikethrough": {Text: []TextOp{TextAttr{Attr: Strikethrough}}},

	"overflow-scroll":   {Ops: []Op{Scroll{Mode: ScrollBoth}}},
	"overflow-y-scroll": {Ops: []Op{Scroll{Mode: ScrollVertical}}},
	"overflow-x-scroll": {Ops: []Op{Scroll{Mode: ScrollHorizontal}}},
	"overflow-hidden":   {Ops: []Op{OverflowHidden{}}},
	"focusable":         {Ops: []Op{Focusable{}}},
	"hidden":            {Ops: []Op{Hidden{}}},
	"truncate":          {Ops: []Op{Truncate{}}},
	"nowrap":            {Ops: []Op{Wrap{Enabled: false}}},
	"wrap":              {Ops: []Op{Wrap{Enabled: true}}},
	"scrollbar-hidden":  {Ops: []Op{ScrollbarHidden{}}},
}

// Named-color families: prefix -> op constructor. Registered into static at init.
var colorFamilies = []struct {
	prefix string
	make   func(Color) Class
}{
	{"border-", func(c Color) Class { return Class{Ops: []Op{BorderColor{Color: c}}} }},
	{"text-", func(c Color) Class { return Class{Text: []TextOp{Foreground{Color: c}}} }},
	{"bg-", func(c Color) Class { return Class{Ops: []Op{Background{Color: c}}} }},
	{"scrollbar-", func(c Color) Class { return Class{Ops: []Op{ScrollbarColor{Color: c}}} }},
	{"scrollbar-thumb-", func(c Color) Class { return Class{Ops: []Op{ScrollbarThumbColor{Color: c}}} }},
}

func init() {
	for _, fam := range colorFamilies {
		for i, name := range ColorNames {
			// Border colors only exist for the 8 base colors in the class set.
			if fam.prefix == "border-" && i >= 8 {
				continue
			}
			static[fam.prefix+name] = fam.make(Color{Index: uint8(i)})
		}
	}
}

// numeric maps a class prefix to an op taking one integer argument.
var numeric = []struct {
	prefix string
	make   func(int) Op
}{
	{"gap-", func(n int) Op { return Gap{N: n} }},
	{"p-", func(n int) Op { return Padding{N: n} }},
	{"m-", func(n int) Op { return Margin{N: n} }},
	{"w-", func(n int) Op { return Width{N: n} }},
	{"h-", func(n int) Op { return Height{N: n} }},
	{"min-w-", func(n int) Op { return MinWidth{N: n} }},
	{"max-w-", func(n int) Op { return MaxWidth{N: n} }},
	{"min-h-", func(n int) Op { return MinHeight{N: n} }},
	{"max-h-", func(n int) Op { return MaxHeight{N: n} }},
	{"flex-grow-", func(n int) Op { return FlexGrow{N: n} }},
	{"flex-shrink-", func(n int) Op { return FlexShrink{N: n} }},
}

// hexFamilies maps a class prefix to an op taking a hex color argument.
var hexFamilies = []struct {
	prefix string
	make   func(Color) Class
}{
	{"text-[#", func(c Color) Class { return Class{Text: []TextOp{Foreground{Color: c}}} }},
	{"bg-[#", func(c Color) Class { return Class{Ops: []Op{Background{Color: c}}} }},
	{"border-[#", func(c Color) Class { return Class{Ops: []Op{BorderColor{Color: c}}} }},
	{"scrollbar-[#", func(c Color) Class { return Class{Ops: []Op{ScrollbarColor{Color: c}}} }},
	{"scrollbar-thumb-[#", func(c Color) Class { return Class{Ops: []Op{ScrollbarThumbColor{Color: c}}} }},
}

var gradientFamilies = []struct {
	prefix string
	make   func(start, end Color, dir GradientDirection) Op
}{
	{"text-gradient-", func(s, e Color, d GradientDirection) Op { return TextGradient{Start: s, End: e, Direction: d} }},
	{"bg-gradient-", func(s, e Color, d GradientDirection) Op { return BackgroundGradient{Start: s, End: e, Direction: d} }},
	{"border-gradient-", func(s, e Color, d GradientDirection) Op { return BorderGradient{Start: s, End: e, Direction: d} }},
}

var (
	fractionPattern = regexp.MustCompile(`^([wh])-(\d+)/(\d+)$`)
	keywordPattern  = regexp.MustCompile(`^([wh])-(full|auto)$`)
	hexPattern      = regexp.MustCompile(`^([0-9a-fA-F]{3}|[0-9a-fA-F]{6})\]$`)
	gradientPattern = regexp.MustCompile(`^[\w-]+-[\w-]+$`)
	// Per-side spacing: p/m + side (t r b l x y) + number.
	sidePattern = regexp.MustCompile(`^([pm])([trblxy])-(\d+)$`)
)

// Known reports whether class is a recognized utility class.
func Known(class string) bool {
	_, ok := Resolve(class)
	return ok
}

// StaticClasses returns every fixed-name class, sorted. Parameterized classes
// (gap-N, w-N/M, hex colors, gradients, per-side spacing) are not listed.
func StaticClasses() []string {
	out := make([]string, 0, len(static))
	for name := range static {
		out = append(out, name)
	}
	slices.Sort(out)
	return out
}

// Resolve resolves one class. Per-side spacing classes (pt-1, mx-2, ...) are
// known but resolve to no ops on their own; Parse folds them into
// PaddingEdges/MarginEdges.
func Resolve(class string) (Class, bool) {
	class = strings.TrimSpace(class)
	if class == "" {
		return Class{}, false
	}
	if c, ok := static[class]; ok {
		return c, true
	}
	for _, n := range numeric {
		if rest, ok := strings.CutPrefix(class, n.prefix); ok {
			if v, ok := digits(rest); ok {
				return Class{Ops: []Op{n.make(v)}}, true
			}
		}
	}
	if m := fractionPattern.FindStringSubmatch(class); m != nil {
		num, _ := strconv.Atoi(m[2])
		den, _ := strconv.Atoi(m[3])
		if den != 0 {
			pct := float64(num) / float64(den) * 100
			if m[1] == "w" {
				return Class{Ops: []Op{WidthPercent{Percent: pct}}}, true
			}
			return Class{Ops: []Op{HeightPercent{Percent: pct}}}, true
		}
	}
	if m := keywordPattern.FindStringSubmatch(class); m != nil {
		switch {
		case m[1] == "w" && m[2] == "full":
			return Class{Ops: []Op{WidthPercent{Percent: 100}}}, true
		case m[1] == "w":
			return Class{Ops: []Op{WidthAuto{}}}, true
		case m[2] == "full":
			return Class{Ops: []Op{HeightPercent{Percent: 100}}}, true
		default:
			return Class{Ops: []Op{HeightAuto{}}}, true
		}
	}
	for _, h := range hexFamilies {
		rest, ok := strings.CutPrefix(class, h.prefix)
		if !ok {
			continue
		}
		m := hexPattern.FindStringSubmatch(rest)
		if m == nil {
			continue
		}
		if c, ok := ParseHex(m[1]); ok {
			return h.make(c), true
		}
	}
	for _, g := range gradientFamilies {
		rest, ok := strings.CutPrefix(class, g.prefix)
		if !ok || !gradientPattern.MatchString(rest) {
			continue
		}
		start, end, dir := parseGradient(rest)
		return Class{Ops: []Op{g.make(start, end, dir)}}, true
	}
	if isPadding, side, n, ok := parseSide(class); ok {
		var e edges
		e.merge(side, n)
		return Class{Ops: []Op{e.op(isPadding)}}, true
	}
	return Class{}, false
}

// Parse resolves a whitespace-separated class list. Unknown classes are skipped.
func Parse(classes string) Result {
	var result Result
	var padding, margin edges
	for class := range strings.FieldsSeq(classes) {
		if isPadding, side, n, ok := parseSide(class); ok {
			if isPadding {
				padding.merge(side, n)
			} else {
				margin.merge(side, n)
			}
			continue
		}
		if c, ok := Resolve(class); ok {
			result.Classes = append(result.Classes, c)
		}
	}
	if padding.set {
		result.Classes = append(result.Classes, Class{Ops: []Op{padding.op(true)}})
	}
	if margin.set {
		result.Classes = append(result.Classes, Class{Ops: []Op{margin.op(false)}})
	}
	return result
}

type edges struct {
	top, right, bottom, left int
	set                      bool
}

// op returns the accumulated edges as a padding or margin op.
func (e edges) op(padding bool) Op {
	if padding {
		return PaddingEdges{Top: e.top, Right: e.right, Bottom: e.bottom, Left: e.left}
	}
	return MarginEdges{Top: e.top, Right: e.right, Bottom: e.bottom, Left: e.left}
}

func (e *edges) merge(side byte, n int) {
	e.set = true
	switch side {
	case 't':
		e.top = n
	case 'r':
		e.right = n
	case 'b':
		e.bottom = n
	case 'l':
		e.left = n
	case 'x':
		e.left, e.right = n, n
	case 'y':
		e.top, e.bottom = n, n
	}
}

// parseSide matches per-side spacing classes like pt-1 or mx-2.
func parseSide(class string) (isPadding bool, side byte, n int, ok bool) {
	m := sidePattern.FindStringSubmatch(class)
	if m == nil {
		return false, 0, 0, false
	}
	n, _ = strconv.Atoi(m[3])
	return m[1] == "p", m[2][0], n, true
}

var directionSuffixes = []struct {
	suffix string
	dir    GradientDirection
}{
	{"-h", Horizontal}, {"-v", Vertical}, {"-dd", DiagonalDown}, {"-du", DiagonalUp},
}

// parseGradient splits "<start>-<end>[-dir]" into colors. The end color is
// matched against known names from the right so hyphenated names like
// bright-red parse; otherwise the last hyphen splits. Unknown names fall
// back to black.
func parseGradient(rest string) (start, end Color, dir GradientDirection) {
	for _, s := range directionSuffixes {
		if before, ok := strings.CutSuffix(rest, s.suffix); ok {
			rest, dir = before, s.dir
			break
		}
	}
	// Longest names first so "bright-red" wins over "red".
	for i := len(ColorNames) - 1; i >= 0; i-- {
		if suffix := "-" + ColorNames[i]; strings.HasSuffix(rest, suffix) {
			return Named(strings.TrimSuffix(rest, suffix)), Named(ColorNames[i]), dir
		}
	}
	var startName, endName string
	if i := strings.LastIndex(rest, "-"); i > 0 {
		startName, endName = rest[:i], rest[i+1:]
	}
	return Named(startName), Named(endName), dir
}

// digits parses a non-empty all-digit string.
func digits(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(s)
	return n, err == nil
}
