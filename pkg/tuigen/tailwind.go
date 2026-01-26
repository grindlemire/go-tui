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

// TailwindValidationResult contains validation results for a class
type TailwindValidationResult struct {
	Valid      bool
	Class      string
	Suggestion string // "did you mean...?" hint
}

// TailwindClassInfo contains metadata about a class for autocomplete
type TailwindClassInfo struct {
	Name        string
	Category    string // "layout", "spacing", "typography", "visual", "flex"
	Description string
	Example     string
}

// TailwindClassWithPosition tracks a class and its position within the attribute value
type TailwindClassWithPosition struct {
	Class      string
	StartCol   int // column offset relative to attribute value start
	EndCol     int // column offset relative to attribute value start
	Valid      bool
	Suggestion string
}

// similarClasses maps common typos/alternatives to correct class names
var similarClasses = map[string]string{
	"flex-column":    "flex-col",
	"flex-columns":   "flex-col",
	"flex-rows":      "flex-row",
	"gap":            "gap-1",
	"padding":        "p-1",
	"margin":         "m-1",
	"bold":           "font-bold",
	"italic":         "italic",
	"dim":            "font-dim",
	"border-single":  "border",
	"width":          "w-1",
	"height":         "h-1",
	"center":         "text-center",
	"left":           "text-left",
	"right":          "text-right",
	"align-center":   "text-center",
	"align-left":     "text-left",
	"align-right":    "text-right",
	"grow":           "flex-grow",
	"shrink":         "flex-shrink",
	"no-grow":        "flex-grow-0",
	"no-shrink":      "flex-shrink-0",
	"padding-top":    "pt-1",
	"padding-bottom": "pb-1",
	"padding-left":   "pl-1",
	"padding-right":  "pr-1",
	"margin-top":     "mt-1",
	"margin-bottom":  "mb-1",
	"margin-left":    "ml-1",
	"margin-right":   "mr-1",
	"col":            "flex-col",
	"row":            "flex-row",
	"column":         "flex-col",
	"columns":        "flex-col",
	"rows":           "flex-row",
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Create a 2D slice for dynamic programming
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
	}

	// Initialize first column
	for i := 0; i <= len(a); i++ {
		matrix[i][0] = i
	}

	// Initialize first row
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	// Fill in the rest of the matrix
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

// getAllKnownClassNames returns all known class names for fuzzy matching
func getAllKnownClassNames() []string {
	classes := make([]string, 0, len(tailwindClasses)+50)

	// Add all static class names
	for name := range tailwindClasses {
		classes = append(classes, name)
	}

	// Add common pattern-based class examples
	patternExamples := []string{
		"gap-1", "gap-2", "gap-3", "gap-4",
		"p-1", "p-2", "p-3", "p-4",
		"px-1", "px-2", "px-3", "px-4",
		"py-1", "py-2", "py-3", "py-4",
		"pt-1", "pt-2", "pt-3", "pt-4",
		"pr-1", "pr-2", "pr-3", "pr-4",
		"pb-1", "pb-2", "pb-3", "pb-4",
		"pl-1", "pl-2", "pl-3", "pl-4",
		"m-1", "m-2", "m-3", "m-4",
		"mt-1", "mt-2", "mt-3", "mt-4",
		"mr-1", "mr-2", "mr-3", "mr-4",
		"mb-1", "mb-2", "mb-3", "mb-4",
		"ml-1", "ml-2", "ml-3", "ml-4",
		"mx-1", "mx-2", "mx-3", "mx-4",
		"my-1", "my-2", "my-3", "my-4",
		"w-1", "w-10", "w-20", "w-50", "w-100",
		"w-full", "w-auto", "w-1/2", "w-1/3", "w-2/3", "w-1/4", "w-3/4",
		"h-1", "h-10", "h-20", "h-50", "h-100",
		"h-full", "h-auto", "h-1/2", "h-1/3", "h-2/3", "h-1/4", "h-3/4",
		"min-w-1", "min-w-10", "max-w-50", "max-w-100",
		"min-h-1", "min-h-10", "max-h-50", "max-h-100",
		"flex-grow-0", "flex-grow-1", "flex-grow-2",
		"flex-shrink-0", "flex-shrink-1", "flex-shrink-2",
	}
	classes = append(classes, patternExamples...)

	return classes
}

// findSimilarClass finds a similar valid class for a given invalid class
func findSimilarClass(class string) string {
	// First check exact match in similarClasses map
	if suggestion, ok := similarClasses[class]; ok {
		return suggestion
	}

	// Use Levenshtein distance for fuzzy matching
	allClasses := getAllKnownClassNames()
	bestMatch := ""
	bestDistance := 4 // Only suggest if distance <= 3

	for _, knownClass := range allClasses {
		dist := levenshteinDistance(class, knownClass)
		if dist < bestDistance {
			bestDistance = dist
			bestMatch = knownClass
		}
	}

	return bestMatch
}

// ValidateTailwindClass validates a single class and returns suggestions
func ValidateTailwindClass(class string) TailwindValidationResult {
	class = strings.TrimSpace(class)
	if class == "" {
		return TailwindValidationResult{Valid: false, Class: class}
	}

	// Check if it's a valid class using ParseTailwindClass
	_, ok := ParseTailwindClass(class)
	if ok {
		return TailwindValidationResult{Valid: true, Class: class}
	}

	// Invalid class - find a suggestion
	suggestion := findSimilarClass(class)
	return TailwindValidationResult{
		Valid:      false,
		Class:      class,
		Suggestion: suggestion,
	}
}

// ParseTailwindClassesWithPositions parses classes and tracks their positions
func ParseTailwindClassesWithPositions(classes string, attrStartCol int) []TailwindClassWithPosition {
	var result []TailwindClassWithPosition

	// Track position as we iterate through the string
	pos := 0
	for pos < len(classes) {
		// Skip leading whitespace
		for pos < len(classes) && (classes[pos] == ' ' || classes[pos] == '\t') {
			pos++
		}
		if pos >= len(classes) {
			break
		}

		// Find the end of this class (next whitespace or end of string)
		startPos := pos
		for pos < len(classes) && classes[pos] != ' ' && classes[pos] != '\t' {
			pos++
		}
		endPos := pos

		class := classes[startPos:endPos]
		if class == "" {
			continue
		}

		validation := ValidateTailwindClass(class)
		result = append(result, TailwindClassWithPosition{
			Class:      class,
			StartCol:   attrStartCol + startPos,
			EndCol:     attrStartCol + endPos,
			Valid:      validation.Valid,
			Suggestion: validation.Suggestion,
		})
	}

	return result
}

// AllTailwindClasses returns all known classes for autocomplete
func AllTailwindClasses() []TailwindClassInfo {
	var classes []TailwindClassInfo

	// Layout classes
	layoutClasses := []TailwindClassInfo{
		{Name: "flex", Category: "layout", Description: "Display flex row", Example: `<div class="flex">`},
		{Name: "flex-row", Category: "layout", Description: "Display flex row", Example: `<div class="flex-row">`},
		{Name: "flex-col", Category: "layout", Description: "Display flex column", Example: `<div class="flex-col">`},
	}
	classes = append(classes, layoutClasses...)

	// Flex utilities
	flexClasses := []TailwindClassInfo{
		{Name: "flex-grow", Category: "flex", Description: "Allow element to grow", Example: `<div class="flex-grow">`},
		{Name: "flex-shrink", Category: "flex", Description: "Allow element to shrink", Example: `<div class="flex-shrink">`},
		{Name: "flex-grow-0", Category: "flex", Description: "Prevent element from growing", Example: `<div class="flex-grow-0">`},
		{Name: "flex-shrink-0", Category: "flex", Description: "Prevent element from shrinking", Example: `<div class="flex-shrink-0">`},
	}
	classes = append(classes, flexClasses...)

	// Justify content
	justifyClasses := []TailwindClassInfo{
		{Name: "justify-start", Category: "flex", Description: "Justify content to start", Example: `<div class="flex justify-start">`},
		{Name: "justify-center", Category: "flex", Description: "Justify content to center", Example: `<div class="flex justify-center">`},
		{Name: "justify-end", Category: "flex", Description: "Justify content to end", Example: `<div class="flex justify-end">`},
		{Name: "justify-between", Category: "flex", Description: "Space between items", Example: `<div class="flex justify-between">`},
		{Name: "justify-around", Category: "flex", Description: "Space around items", Example: `<div class="flex justify-around">`},
		{Name: "justify-evenly", Category: "flex", Description: "Space evenly between items", Example: `<div class="flex justify-evenly">`},
	}
	classes = append(classes, justifyClasses...)

	// Align items
	alignClasses := []TailwindClassInfo{
		{Name: "items-start", Category: "flex", Description: "Align items to start", Example: `<div class="flex items-start">`},
		{Name: "items-center", Category: "flex", Description: "Align items to center", Example: `<div class="flex items-center">`},
		{Name: "items-end", Category: "flex", Description: "Align items to end", Example: `<div class="flex items-end">`},
		{Name: "items-stretch", Category: "flex", Description: "Stretch items to fill", Example: `<div class="flex items-stretch">`},
	}
	classes = append(classes, alignClasses...)

	// Self alignment
	selfClasses := []TailwindClassInfo{
		{Name: "self-start", Category: "flex", Description: "Align self to start", Example: `<div class="self-start">`},
		{Name: "self-center", Category: "flex", Description: "Align self to center", Example: `<div class="self-center">`},
		{Name: "self-end", Category: "flex", Description: "Align self to end", Example: `<div class="self-end">`},
		{Name: "self-stretch", Category: "flex", Description: "Stretch self to fill", Example: `<div class="self-stretch">`},
	}
	classes = append(classes, selfClasses...)

	// Gap classes
	gapClasses := []TailwindClassInfo{
		{Name: "gap-1", Category: "spacing", Description: "Gap of 1 character", Example: `<div class="flex gap-1">`},
		{Name: "gap-2", Category: "spacing", Description: "Gap of 2 characters", Example: `<div class="flex gap-2">`},
		{Name: "gap-3", Category: "spacing", Description: "Gap of 3 characters", Example: `<div class="flex gap-3">`},
		{Name: "gap-4", Category: "spacing", Description: "Gap of 4 characters", Example: `<div class="flex gap-4">`},
	}
	classes = append(classes, gapClasses...)

	// Padding classes
	paddingClasses := []TailwindClassInfo{
		{Name: "p-1", Category: "spacing", Description: "Padding of 1 on all sides", Example: `<div class="p-1">`},
		{Name: "p-2", Category: "spacing", Description: "Padding of 2 on all sides", Example: `<div class="p-2">`},
		{Name: "p-3", Category: "spacing", Description: "Padding of 3 on all sides", Example: `<div class="p-3">`},
		{Name: "p-4", Category: "spacing", Description: "Padding of 4 on all sides", Example: `<div class="p-4">`},
		{Name: "px-1", Category: "spacing", Description: "Horizontal padding of 1", Example: `<div class="px-1">`},
		{Name: "px-2", Category: "spacing", Description: "Horizontal padding of 2", Example: `<div class="px-2">`},
		{Name: "py-1", Category: "spacing", Description: "Vertical padding of 1", Example: `<div class="py-1">`},
		{Name: "py-2", Category: "spacing", Description: "Vertical padding of 2", Example: `<div class="py-2">`},
		{Name: "pt-1", Category: "spacing", Description: "Top padding of 1", Example: `<div class="pt-1">`},
		{Name: "pt-2", Category: "spacing", Description: "Top padding of 2", Example: `<div class="pt-2">`},
		{Name: "pr-1", Category: "spacing", Description: "Right padding of 1", Example: `<div class="pr-1">`},
		{Name: "pr-2", Category: "spacing", Description: "Right padding of 2", Example: `<div class="pr-2">`},
		{Name: "pb-1", Category: "spacing", Description: "Bottom padding of 1", Example: `<div class="pb-1">`},
		{Name: "pb-2", Category: "spacing", Description: "Bottom padding of 2", Example: `<div class="pb-2">`},
		{Name: "pl-1", Category: "spacing", Description: "Left padding of 1", Example: `<div class="pl-1">`},
		{Name: "pl-2", Category: "spacing", Description: "Left padding of 2", Example: `<div class="pl-2">`},
	}
	classes = append(classes, paddingClasses...)

	// Margin classes
	marginClasses := []TailwindClassInfo{
		{Name: "m-1", Category: "spacing", Description: "Margin of 1 on all sides", Example: `<div class="m-1">`},
		{Name: "m-2", Category: "spacing", Description: "Margin of 2 on all sides", Example: `<div class="m-2">`},
		{Name: "m-3", Category: "spacing", Description: "Margin of 3 on all sides", Example: `<div class="m-3">`},
		{Name: "m-4", Category: "spacing", Description: "Margin of 4 on all sides", Example: `<div class="m-4">`},
		{Name: "mx-1", Category: "spacing", Description: "Horizontal margin of 1", Example: `<div class="mx-1">`},
		{Name: "mx-2", Category: "spacing", Description: "Horizontal margin of 2", Example: `<div class="mx-2">`},
		{Name: "my-1", Category: "spacing", Description: "Vertical margin of 1", Example: `<div class="my-1">`},
		{Name: "my-2", Category: "spacing", Description: "Vertical margin of 2", Example: `<div class="my-2">`},
		{Name: "mt-1", Category: "spacing", Description: "Top margin of 1", Example: `<div class="mt-1">`},
		{Name: "mt-2", Category: "spacing", Description: "Top margin of 2", Example: `<div class="mt-2">`},
		{Name: "mr-1", Category: "spacing", Description: "Right margin of 1", Example: `<div class="mr-1">`},
		{Name: "mr-2", Category: "spacing", Description: "Right margin of 2", Example: `<div class="mr-2">`},
		{Name: "mb-1", Category: "spacing", Description: "Bottom margin of 1", Example: `<div class="mb-1">`},
		{Name: "mb-2", Category: "spacing", Description: "Bottom margin of 2", Example: `<div class="mb-2">`},
		{Name: "ml-1", Category: "spacing", Description: "Left margin of 1", Example: `<div class="ml-1">`},
		{Name: "ml-2", Category: "spacing", Description: "Left margin of 2", Example: `<div class="ml-2">`},
	}
	classes = append(classes, marginClasses...)

	// Width classes
	widthClasses := []TailwindClassInfo{
		{Name: "w-full", Category: "layout", Description: "Full width (100%)", Example: `<div class="w-full">`},
		{Name: "w-auto", Category: "layout", Description: "Auto width (size to content)", Example: `<div class="w-auto">`},
		{Name: "w-1/2", Category: "layout", Description: "Half width (50%)", Example: `<div class="w-1/2">`},
		{Name: "w-1/3", Category: "layout", Description: "One-third width (33%)", Example: `<div class="w-1/3">`},
		{Name: "w-2/3", Category: "layout", Description: "Two-thirds width (67%)", Example: `<div class="w-2/3">`},
		{Name: "w-1/4", Category: "layout", Description: "Quarter width (25%)", Example: `<div class="w-1/4">`},
		{Name: "w-3/4", Category: "layout", Description: "Three-quarters width (75%)", Example: `<div class="w-3/4">`},
	}
	classes = append(classes, widthClasses...)

	// Height classes
	heightClasses := []TailwindClassInfo{
		{Name: "h-full", Category: "layout", Description: "Full height (100%)", Example: `<div class="h-full">`},
		{Name: "h-auto", Category: "layout", Description: "Auto height (size to content)", Example: `<div class="h-auto">`},
		{Name: "h-1/2", Category: "layout", Description: "Half height (50%)", Example: `<div class="h-1/2">`},
		{Name: "h-1/3", Category: "layout", Description: "One-third height (33%)", Example: `<div class="h-1/3">`},
		{Name: "h-2/3", Category: "layout", Description: "Two-thirds height (67%)", Example: `<div class="h-2/3">`},
		{Name: "h-1/4", Category: "layout", Description: "Quarter height (25%)", Example: `<div class="h-1/4">`},
		{Name: "h-3/4", Category: "layout", Description: "Three-quarters height (75%)", Example: `<div class="h-3/4">`},
	}
	classes = append(classes, heightClasses...)

	// Border classes
	borderClasses := []TailwindClassInfo{
		{Name: "border", Category: "visual", Description: "Single line border", Example: `<div class="border">`},
		{Name: "border-rounded", Category: "visual", Description: "Rounded border", Example: `<div class="border-rounded">`},
		{Name: "border-double", Category: "visual", Description: "Double line border", Example: `<div class="border-double">`},
		{Name: "border-thick", Category: "visual", Description: "Thick border", Example: `<div class="border-thick">`},
		{Name: "border-red", Category: "visual", Description: "Red border color", Example: `<div class="border border-red">`},
		{Name: "border-green", Category: "visual", Description: "Green border color", Example: `<div class="border border-green">`},
		{Name: "border-blue", Category: "visual", Description: "Blue border color", Example: `<div class="border border-blue">`},
		{Name: "border-cyan", Category: "visual", Description: "Cyan border color", Example: `<div class="border border-cyan">`},
		{Name: "border-magenta", Category: "visual", Description: "Magenta border color", Example: `<div class="border border-magenta">`},
		{Name: "border-yellow", Category: "visual", Description: "Yellow border color", Example: `<div class="border border-yellow">`},
		{Name: "border-white", Category: "visual", Description: "White border color", Example: `<div class="border border-white">`},
		{Name: "border-black", Category: "visual", Description: "Black border color", Example: `<div class="border border-black">`},
	}
	classes = append(classes, borderClasses...)

	// Typography classes
	typographyClasses := []TailwindClassInfo{
		{Name: "font-bold", Category: "typography", Description: "Bold text", Example: `<span class="font-bold">Bold</span>`},
		{Name: "font-dim", Category: "typography", Description: "Dim/faint text", Example: `<span class="font-dim">Dim</span>`},
		{Name: "italic", Category: "typography", Description: "Italic text", Example: `<span class="italic">Italic</span>`},
		{Name: "underline", Category: "typography", Description: "Underlined text", Example: `<span class="underline">Underlined</span>`},
		{Name: "strikethrough", Category: "typography", Description: "Strikethrough text", Example: `<span class="strikethrough">Strikethrough</span>`},
		{Name: "text-left", Category: "typography", Description: "Align text left", Example: `<div class="text-left">`},
		{Name: "text-center", Category: "typography", Description: "Center text", Example: `<div class="text-center">`},
		{Name: "text-right", Category: "typography", Description: "Align text right", Example: `<div class="text-right">`},
	}
	classes = append(classes, typographyClasses...)

	// Text color classes
	textColorClasses := []TailwindClassInfo{
		{Name: "text-red", Category: "visual", Description: "Red text color", Example: `<span class="text-red">Red</span>`},
		{Name: "text-green", Category: "visual", Description: "Green text color", Example: `<span class="text-green">Green</span>`},
		{Name: "text-blue", Category: "visual", Description: "Blue text color", Example: `<span class="text-blue">Blue</span>`},
		{Name: "text-cyan", Category: "visual", Description: "Cyan text color", Example: `<span class="text-cyan">Cyan</span>`},
		{Name: "text-magenta", Category: "visual", Description: "Magenta text color", Example: `<span class="text-magenta">Magenta</span>`},
		{Name: "text-yellow", Category: "visual", Description: "Yellow text color", Example: `<span class="text-yellow">Yellow</span>`},
		{Name: "text-white", Category: "visual", Description: "White text color", Example: `<span class="text-white">White</span>`},
		{Name: "text-black", Category: "visual", Description: "Black text color", Example: `<span class="text-black">Black</span>`},
	}
	classes = append(classes, textColorClasses...)

	// Background color classes
	bgColorClasses := []TailwindClassInfo{
		{Name: "bg-red", Category: "visual", Description: "Red background", Example: `<div class="bg-red">`},
		{Name: "bg-green", Category: "visual", Description: "Green background", Example: `<div class="bg-green">`},
		{Name: "bg-blue", Category: "visual", Description: "Blue background", Example: `<div class="bg-blue">`},
		{Name: "bg-cyan", Category: "visual", Description: "Cyan background", Example: `<div class="bg-cyan">`},
		{Name: "bg-magenta", Category: "visual", Description: "Magenta background", Example: `<div class="bg-magenta">`},
		{Name: "bg-yellow", Category: "visual", Description: "Yellow background", Example: `<div class="bg-yellow">`},
		{Name: "bg-white", Category: "visual", Description: "White background", Example: `<div class="bg-white">`},
		{Name: "bg-black", Category: "visual", Description: "Black background", Example: `<div class="bg-black">`},
	}
	classes = append(classes, bgColorClasses...)

	// Scroll classes
	scrollClasses := []TailwindClassInfo{
		{Name: "overflow-scroll", Category: "layout", Description: "Enable scrolling in both directions", Example: `<div class="overflow-scroll">`},
		{Name: "overflow-y-scroll", Category: "layout", Description: "Enable vertical scrolling", Example: `<div class="overflow-y-scroll">`},
		{Name: "overflow-x-scroll", Category: "layout", Description: "Enable horizontal scrolling", Example: `<div class="overflow-x-scroll">`},
	}
	classes = append(classes, scrollClasses...)

	return classes
}
