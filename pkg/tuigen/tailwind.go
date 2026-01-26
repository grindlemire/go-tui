package tuigen

import (
	"regexp"
	"strconv"
	"strings"
)

// TailwindMapping represents a parsed Tailwind class and its corresponding Go code
type TailwindMapping struct {
	Option      string // The Go code to generate (e.g., "element.WithDirection(layout.Column)")
	NeedsImport string // Import path needed, if any (e.g., "layout", "tui")
	IsTextStyle bool   // Whether this is a text style modifier
	TextMethod  string // The method to chain on tui.NewStyle() (e.g., "Bold()", "Foreground(tui.Cyan)")
}

// tailwindClasses maps Tailwind class names to their TUI equivalents
var tailwindClasses = map[string]TailwindMapping{
	// Layout - flex direction
	"flex":     {Option: "element.WithDirection(layout.Row)", NeedsImport: "layout"},
	"flex-row": {Option: "element.WithDirection(layout.Row)", NeedsImport: "layout"},
	"flex-col": {Option: "element.WithDirection(layout.Column)", NeedsImport: "layout"},

	// Flex properties
	"flex-grow":   {Option: "element.WithFlexGrow(1)", NeedsImport: ""},
	"flex-shrink": {Option: "element.WithFlexShrink(1)", NeedsImport: ""},

	// Justify content
	"justify-start":   {Option: "element.WithJustify(layout.JustifyStart)", NeedsImport: "layout"},
	"justify-center":  {Option: "element.WithJustify(layout.JustifyCenter)", NeedsImport: "layout"},
	"justify-end":     {Option: "element.WithJustify(layout.JustifyEnd)", NeedsImport: "layout"},
	"justify-between": {Option: "element.WithJustify(layout.JustifySpaceBetween)", NeedsImport: "layout"},
	"justify-evenly":  {Option: "element.WithJustify(layout.JustifySpaceEvenly)", NeedsImport: "layout"},
	"justify-around":  {Option: "element.WithJustify(layout.JustifySpaceAround)", NeedsImport: "layout"},

	// Align items
	"items-start":   {Option: "element.WithAlign(layout.AlignStart)", NeedsImport: "layout"},
	"items-center":  {Option: "element.WithAlign(layout.AlignCenter)", NeedsImport: "layout"},
	"items-end":     {Option: "element.WithAlign(layout.AlignEnd)", NeedsImport: "layout"},
	"items-stretch": {Option: "element.WithAlign(layout.AlignStretch)", NeedsImport: "layout"},

	// Self-alignment
	"self-start":   {Option: "element.WithAlignSelf(layout.AlignStart)", NeedsImport: "layout"},
	"self-end":     {Option: "element.WithAlignSelf(layout.AlignEnd)", NeedsImport: "layout"},
	"self-center":  {Option: "element.WithAlignSelf(layout.AlignCenter)", NeedsImport: "layout"},
	"self-stretch": {Option: "element.WithAlignSelf(layout.AlignStretch)", NeedsImport: "layout"},

	// Text alignment
	"text-left":   {Option: "element.WithTextAlign(element.TextAlignLeft)", NeedsImport: ""},
	"text-center": {Option: "element.WithTextAlign(element.TextAlignCenter)", NeedsImport: ""},
	"text-right":  {Option: "element.WithTextAlign(element.TextAlignRight)", NeedsImport: ""},

	// Borders
	"border":         {Option: "element.WithBorder(tui.BorderSingle)", NeedsImport: "tui"},
	"border-rounded": {Option: "element.WithBorder(tui.BorderRounded)", NeedsImport: "tui"},
	"border-double":  {Option: "element.WithBorder(tui.BorderDouble)", NeedsImport: "tui"},
	"border-thick":   {Option: "element.WithBorder(tui.BorderThick)", NeedsImport: "tui"},

	// Border colors
	"border-red":     {Option: "element.WithBorderStyle(tui.NewStyle().Foreground(tui.Red))", NeedsImport: "tui"},
	"border-green":   {Option: "element.WithBorderStyle(tui.NewStyle().Foreground(tui.Green))", NeedsImport: "tui"},
	"border-blue":    {Option: "element.WithBorderStyle(tui.NewStyle().Foreground(tui.Blue))", NeedsImport: "tui"},
	"border-cyan":    {Option: "element.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan))", NeedsImport: "tui"},
	"border-magenta": {Option: "element.WithBorderStyle(tui.NewStyle().Foreground(tui.Magenta))", NeedsImport: "tui"},
	"border-yellow":  {Option: "element.WithBorderStyle(tui.NewStyle().Foreground(tui.Yellow))", NeedsImport: "tui"},
	"border-white":   {Option: "element.WithBorderStyle(tui.NewStyle().Foreground(tui.White))", NeedsImport: "tui"},
	"border-black":   {Option: "element.WithBorderStyle(tui.NewStyle().Foreground(tui.Black))", NeedsImport: "tui"},

	// Text styles
	"font-bold":  {IsTextStyle: true, TextMethod: "Bold()"},
	"font-dim":   {IsTextStyle: true, TextMethod: "Dim()"},
	"italic":     {IsTextStyle: true, TextMethod: "Italic()"},
	"underline":  {IsTextStyle: true, TextMethod: "Underline()"},
	"blink":      {IsTextStyle: true, TextMethod: "Blink()"},
	"reverse":    {IsTextStyle: true, TextMethod: "Reverse()"},
	"strikethrough": {IsTextStyle: true, TextMethod: "Strikethrough()"},

	// Text colors
	"text-red":     {IsTextStyle: true, TextMethod: "Foreground(tui.Red)", NeedsImport: "tui"},
	"text-green":   {IsTextStyle: true, TextMethod: "Foreground(tui.Green)", NeedsImport: "tui"},
	"text-blue":    {IsTextStyle: true, TextMethod: "Foreground(tui.Blue)", NeedsImport: "tui"},
	"text-cyan":    {IsTextStyle: true, TextMethod: "Foreground(tui.Cyan)", NeedsImport: "tui"},
	"text-magenta": {IsTextStyle: true, TextMethod: "Foreground(tui.Magenta)", NeedsImport: "tui"},
	"text-yellow":  {IsTextStyle: true, TextMethod: "Foreground(tui.Yellow)", NeedsImport: "tui"},
	"text-white":   {IsTextStyle: true, TextMethod: "Foreground(tui.White)", NeedsImport: "tui"},
	"text-black":   {IsTextStyle: true, TextMethod: "Foreground(tui.Black)", NeedsImport: "tui"},

	// Background colors
	"bg-red":     {IsTextStyle: true, TextMethod: "Background(tui.Red)", NeedsImport: "tui"},
	"bg-green":   {IsTextStyle: true, TextMethod: "Background(tui.Green)", NeedsImport: "tui"},
	"bg-blue":    {IsTextStyle: true, TextMethod: "Background(tui.Blue)", NeedsImport: "tui"},
	"bg-cyan":    {IsTextStyle: true, TextMethod: "Background(tui.Cyan)", NeedsImport: "tui"},
	"bg-magenta": {IsTextStyle: true, TextMethod: "Background(tui.Magenta)", NeedsImport: "tui"},
	"bg-yellow":  {IsTextStyle: true, TextMethod: "Background(tui.Yellow)", NeedsImport: "tui"},
	"bg-white":   {IsTextStyle: true, TextMethod: "Background(tui.White)", NeedsImport: "tui"},
	"bg-black":   {IsTextStyle: true, TextMethod: "Background(tui.Black)", NeedsImport: "tui"},

	// Scroll
	"overflow-scroll":   {Option: "element.WithScrollable(element.ScrollBoth)", NeedsImport: ""},
	"overflow-y-scroll": {Option: "element.WithScrollable(element.ScrollVertical)", NeedsImport: ""},
	"overflow-x-scroll": {Option: "element.WithScrollable(element.ScrollHorizontal)", NeedsImport: ""},
}

// Regex patterns for dynamic classes
var (
	gapPattern       = regexp.MustCompile(`^gap-(\d+)$`)
	paddingPattern   = regexp.MustCompile(`^p-(\d+)$`)
	paddingXPattern  = regexp.MustCompile(`^px-(\d+)$`)
	paddingYPattern  = regexp.MustCompile(`^py-(\d+)$`)
	marginPattern    = regexp.MustCompile(`^m-(\d+)$`)
	widthPattern     = regexp.MustCompile(`^w-(\d+)$`)
	heightPattern    = regexp.MustCompile(`^h-(\d+)$`)
	minWidthPattern  = regexp.MustCompile(`^min-w-(\d+)$`)
	maxWidthPattern  = regexp.MustCompile(`^max-w-(\d+)$`)
	minHeightPattern = regexp.MustCompile(`^min-h-(\d+)$`)
	maxHeightPattern = regexp.MustCompile(`^max-h-(\d+)$`)

	// Width/height fraction and keyword patterns
	widthFractionPattern  = regexp.MustCompile(`^w-(\d+)/(\d+)$`)
	heightFractionPattern = regexp.MustCompile(`^h-(\d+)/(\d+)$`)
	widthKeywordPattern   = regexp.MustCompile(`^w-(full|auto)$`)
	heightKeywordPattern  = regexp.MustCompile(`^h-(full|auto)$`)

	// Individual padding patterns
	ptPattern = regexp.MustCompile(`^pt-(\d+)$`)
	prPattern = regexp.MustCompile(`^pr-(\d+)$`)
	pbPattern = regexp.MustCompile(`^pb-(\d+)$`)
	plPattern = regexp.MustCompile(`^pl-(\d+)$`)

	// Individual margin patterns
	mtPattern = regexp.MustCompile(`^mt-(\d+)$`)
	mrPattern = regexp.MustCompile(`^mr-(\d+)$`)
	mbPattern = regexp.MustCompile(`^mb-(\d+)$`)
	mlPattern = regexp.MustCompile(`^ml-(\d+)$`)
	mxPattern = regexp.MustCompile(`^mx-(\d+)$`)
	myPattern = regexp.MustCompile(`^my-(\d+)$`)

	// Flex grow/shrink patterns
	flexGrowPattern   = regexp.MustCompile(`^flex-grow-(\d+)$`)
	flexShrinkPattern = regexp.MustCompile(`^flex-shrink-(\d+)$`)
)

// PaddingAccumulator tracks individual padding values for accumulation
type PaddingAccumulator struct {
	Top, Right, Bottom, Left             int
	HasTop, HasRight, HasBottom, HasLeft bool
}

// Merge combines an individual side class into the accumulator
func (p *PaddingAccumulator) Merge(side string, value int) {
	switch side {
	case "top":
		p.Top = value
		p.HasTop = true
	case "right":
		p.Right = value
		p.HasRight = true
	case "bottom":
		p.Bottom = value
		p.HasBottom = true
	case "left":
		p.Left = value
		p.HasLeft = true
	case "x": // horizontal (left and right)
		p.Left = value
		p.Right = value
		p.HasLeft = true
		p.HasRight = true
	case "y": // vertical (top and bottom)
		p.Top = value
		p.Bottom = value
		p.HasTop = true
		p.HasBottom = true
	}
}

// HasAny returns true if any side has been set
func (p *PaddingAccumulator) HasAny() bool {
	return p.HasTop || p.HasRight || p.HasBottom || p.HasLeft
}

// ToOption generates WithPaddingTRBL() if any sides are set
func (p *PaddingAccumulator) ToOption() string {
	if !p.HasAny() {
		return ""
	}
	return "element.WithPaddingTRBL(" + strconv.Itoa(p.Top) + ", " + strconv.Itoa(p.Right) + ", " + strconv.Itoa(p.Bottom) + ", " + strconv.Itoa(p.Left) + ")"
}

// MarginAccumulator tracks individual margin values for accumulation
type MarginAccumulator struct {
	Top, Right, Bottom, Left             int
	HasTop, HasRight, HasBottom, HasLeft bool
}

// Merge combines an individual side class into the accumulator
func (m *MarginAccumulator) Merge(side string, value int) {
	switch side {
	case "top":
		m.Top = value
		m.HasTop = true
	case "right":
		m.Right = value
		m.HasRight = true
	case "bottom":
		m.Bottom = value
		m.HasBottom = true
	case "left":
		m.Left = value
		m.HasLeft = true
	case "x": // horizontal (left and right)
		m.Left = value
		m.Right = value
		m.HasLeft = true
		m.HasRight = true
	case "y": // vertical (top and bottom)
		m.Top = value
		m.Bottom = value
		m.HasTop = true
		m.HasBottom = true
	}
}

// HasAny returns true if any side has been set
func (m *MarginAccumulator) HasAny() bool {
	return m.HasTop || m.HasRight || m.HasBottom || m.HasLeft
}

// ToOption generates WithMarginTRBL() if any sides are set
func (m *MarginAccumulator) ToOption() string {
	if !m.HasAny() {
		return ""
	}
	return "element.WithMarginTRBL(" + strconv.Itoa(m.Top) + ", " + strconv.Itoa(m.Right) + ", " + strconv.Itoa(m.Bottom) + ", " + strconv.Itoa(m.Left) + ")"
}

// IndividualSpacingResult indicates an individual padding/margin class was parsed
type IndividualSpacingResult struct {
	IsPadding bool   // true for padding, false for margin
	Side      string // "top", "right", "bottom", "left", "x", "y"
	Value     int
}

// parseIndividualSpacing checks if a class is an individual padding/margin class
// Returns the result and true if it matched, or zero value and false if not
func parseIndividualSpacing(class string) (IndividualSpacingResult, bool) {
	// Individual padding
	if matches := ptPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return IndividualSpacingResult{IsPadding: true, Side: "top", Value: n}, true
	}
	if matches := prPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return IndividualSpacingResult{IsPadding: true, Side: "right", Value: n}, true
	}
	if matches := pbPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return IndividualSpacingResult{IsPadding: true, Side: "bottom", Value: n}, true
	}
	if matches := plPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return IndividualSpacingResult{IsPadding: true, Side: "left", Value: n}, true
	}
	if matches := paddingXPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return IndividualSpacingResult{IsPadding: true, Side: "x", Value: n}, true
	}
	if matches := paddingYPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return IndividualSpacingResult{IsPadding: true, Side: "y", Value: n}, true
	}

	// Individual margin
	if matches := mtPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return IndividualSpacingResult{IsPadding: false, Side: "top", Value: n}, true
	}
	if matches := mrPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return IndividualSpacingResult{IsPadding: false, Side: "right", Value: n}, true
	}
	if matches := mbPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return IndividualSpacingResult{IsPadding: false, Side: "bottom", Value: n}, true
	}
	if matches := mlPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return IndividualSpacingResult{IsPadding: false, Side: "left", Value: n}, true
	}
	if matches := mxPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return IndividualSpacingResult{IsPadding: false, Side: "x", Value: n}, true
	}
	if matches := myPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return IndividualSpacingResult{IsPadding: false, Side: "y", Value: n}, true
	}

	return IndividualSpacingResult{}, false
}

// ParseTailwindClass parses a single Tailwind class and returns its mapping
func ParseTailwindClass(class string) (TailwindMapping, bool) {
	class = strings.TrimSpace(class)
	if class == "" {
		return TailwindMapping{}, false
	}

	// Check static mappings first
	if mapping, ok := tailwindClasses[class]; ok {
		return mapping, true
	}

	// Check dynamic patterns
	if matches := gapPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "element.WithGap(" + strconv.Itoa(n) + ")"}, true
	}

	if matches := paddingPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "element.WithPadding(" + strconv.Itoa(n) + ")"}, true
	}

	if matches := paddingXPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "element.WithPaddingTRBL(0, " + strconv.Itoa(n) + ", 0, " + strconv.Itoa(n) + ")"}, true
	}

	if matches := paddingYPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "element.WithPaddingTRBL(" + strconv.Itoa(n) + ", 0, " + strconv.Itoa(n) + ", 0)"}, true
	}

	if matches := marginPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "element.WithMargin(" + strconv.Itoa(n) + ")"}, true
	}

	if matches := widthPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "element.WithWidth(" + strconv.Itoa(n) + ")"}, true
	}

	if matches := heightPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "element.WithHeight(" + strconv.Itoa(n) + ")"}, true
	}

	if matches := minWidthPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "element.WithMinWidth(" + strconv.Itoa(n) + ")"}, true
	}

	if matches := maxWidthPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "element.WithMaxWidth(" + strconv.Itoa(n) + ")"}, true
	}

	if matches := minHeightPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "element.WithMinHeight(" + strconv.Itoa(n) + ")"}, true
	}

	if matches := maxHeightPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "element.WithMaxHeight(" + strconv.Itoa(n) + ")"}, true
	}

	// Width fraction patterns (w-1/2, w-2/3, etc.)
	if matches := widthFractionPattern.FindStringSubmatch(class); matches != nil {
		numerator, _ := strconv.Atoi(matches[1])
		denominator, _ := strconv.Atoi(matches[2])
		if denominator != 0 {
			percent := float64(numerator) / float64(denominator) * 100
			return TailwindMapping{Option: "element.WithWidthPercent(" + strconv.FormatFloat(percent, 'f', 2, 64) + ")"}, true
		}
	}

	// Height fraction patterns (h-1/2, h-2/3, etc.)
	if matches := heightFractionPattern.FindStringSubmatch(class); matches != nil {
		numerator, _ := strconv.Atoi(matches[1])
		denominator, _ := strconv.Atoi(matches[2])
		if denominator != 0 {
			percent := float64(numerator) / float64(denominator) * 100
			return TailwindMapping{Option: "element.WithHeightPercent(" + strconv.FormatFloat(percent, 'f', 2, 64) + ")"}, true
		}
	}

	// Width keyword patterns (w-full, w-auto)
	if matches := widthKeywordPattern.FindStringSubmatch(class); matches != nil {
		keyword := matches[1]
		switch keyword {
		case "full":
			return TailwindMapping{Option: "element.WithWidthPercent(100.00)"}, true
		case "auto":
			return TailwindMapping{Option: "element.WithWidthAuto()"}, true
		}
	}

	// Height keyword patterns (h-full, h-auto)
	if matches := heightKeywordPattern.FindStringSubmatch(class); matches != nil {
		keyword := matches[1]
		switch keyword {
		case "full":
			return TailwindMapping{Option: "element.WithHeightPercent(100.00)"}, true
		case "auto":
			return TailwindMapping{Option: "element.WithHeightAuto()"}, true
		}
	}

	// Flex grow pattern (flex-grow-N)
	if matches := flexGrowPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "element.WithFlexGrow(" + strconv.Itoa(n) + ")"}, true
	}

	// Flex shrink pattern (flex-shrink-N)
	if matches := flexShrinkPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "element.WithFlexShrink(" + strconv.Itoa(n) + ")"}, true
	}

	// Individual padding/margin classes - these are valid but handled separately in ParseTailwindClasses
	if _, ok := parseIndividualSpacing(class); ok {
		// Return a marker mapping - actual handling is done in ParseTailwindClasses
		return TailwindMapping{}, true
	}

	// Unknown class - silently ignore
	return TailwindMapping{}, false
}

// TailwindParseResult contains the parsed results from a class string
type TailwindParseResult struct {
	Options      []string          // Direct element options
	TextMethods  []string          // Text style methods to chain
	NeedsImports map[string]bool   // Imports needed
}

// ParseTailwindClasses parses a full class attribute string
func ParseTailwindClasses(classes string) TailwindParseResult {
	result := TailwindParseResult{
		NeedsImports: make(map[string]bool),
	}

	// Accumulators for individual padding/margin classes
	var paddingAcc PaddingAccumulator
	var marginAcc MarginAccumulator

	for _, class := range strings.Fields(classes) {
		// First, check if it's an individual padding/margin class
		if spacing, ok := parseIndividualSpacing(class); ok {
			if spacing.IsPadding {
				paddingAcc.Merge(spacing.Side, spacing.Value)
			} else {
				marginAcc.Merge(spacing.Side, spacing.Value)
			}
			continue
		}

		mapping, ok := ParseTailwindClass(class)
		if !ok {
			continue
		}

		if mapping.IsTextStyle {
			result.TextMethods = append(result.TextMethods, mapping.TextMethod)
		} else if mapping.Option != "" {
			result.Options = append(result.Options, mapping.Option)
		}

		if mapping.NeedsImport != "" {
			result.NeedsImports[mapping.NeedsImport] = true
		}
	}

	// Add accumulated padding if any sides were set
	if paddingOpt := paddingAcc.ToOption(); paddingOpt != "" {
		result.Options = append(result.Options, paddingOpt)
	}

	// Add accumulated margin if any sides were set
	if marginOpt := marginAcc.ToOption(); marginOpt != "" {
		result.Options = append(result.Options, marginOpt)
	}

	return result
}

// BuildTextStyleOption builds the combined text style option from accumulated methods
func BuildTextStyleOption(methods []string) string {
	if len(methods) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("element.WithTextStyle(tui.NewStyle()")
	for _, method := range methods {
		builder.WriteString(".")
		builder.WriteString(method)
	}
	builder.WriteString(")")
	return builder.String()
}
