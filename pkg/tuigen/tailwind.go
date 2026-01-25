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

	// Align items
	"items-start":  {Option: "element.WithAlign(layout.AlignStart)", NeedsImport: "layout"},
	"items-center": {Option: "element.WithAlign(layout.AlignCenter)", NeedsImport: "layout"},
	"items-end":    {Option: "element.WithAlign(layout.AlignEnd)", NeedsImport: "layout"},

	// Borders
	"border":         {Option: "element.WithBorder(tui.BorderSingle)", NeedsImport: "tui"},
	"border-rounded": {Option: "element.WithBorder(tui.BorderRounded)", NeedsImport: "tui"},
	"border-double":  {Option: "element.WithBorder(tui.BorderDouble)", NeedsImport: "tui"},
	"border-thick":   {Option: "element.WithBorder(tui.BorderThick)", NeedsImport: "tui"},

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
	gapPattern     = regexp.MustCompile(`^gap-(\d+)$`)
	paddingPattern = regexp.MustCompile(`^p-(\d+)$`)
	paddingXPattern = regexp.MustCompile(`^px-(\d+)$`)
	paddingYPattern = regexp.MustCompile(`^py-(\d+)$`)
	marginPattern  = regexp.MustCompile(`^m-(\d+)$`)
	widthPattern   = regexp.MustCompile(`^w-(\d+)$`)
	heightPattern  = regexp.MustCompile(`^h-(\d+)$`)
	minWidthPattern  = regexp.MustCompile(`^min-w-(\d+)$`)
	maxWidthPattern  = regexp.MustCompile(`^max-w-(\d+)$`)
	minHeightPattern = regexp.MustCompile(`^min-h-(\d+)$`)
	maxHeightPattern = regexp.MustCompile(`^max-h-(\d+)$`)
)

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
		return TailwindMapping{Option: "element.WithPaddingX(" + strconv.Itoa(n) + ")"}, true
	}

	if matches := paddingYPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "element.WithPaddingY(" + strconv.Itoa(n) + ")"}, true
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

	for _, class := range strings.Fields(classes) {
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
