package tuigen

import (
	"regexp"
)

// TailwindMapping represents a parsed Tailwind class and its corresponding Go code
type TailwindMapping struct {
	Option      string // The Go code to generate (e.g., "tui.WithDirection(tui.Column)")
	NeedsImport string // Import path needed, if any (e.g., "layout", "tui")
	IsTextStyle bool   // Whether this is a text style modifier
	TextMethod  string // The method to chain on tui.NewStyle() (e.g., "Bold()", "Foreground(tui.Cyan)")
}

// tailwindClasses maps Tailwind class names to their TUI equivalents
var tailwindClasses = map[string]TailwindMapping{
	// Layout - flex direction
	"flex":     {Option: "tui.WithDirection(tui.Row)", NeedsImport: "tui"},
	"flex-row": {Option: "tui.WithDirection(tui.Row)", NeedsImport: "tui"},
	"flex-col": {Option: "tui.WithDirection(tui.Column)", NeedsImport: "tui"},

	// Flex properties
	"flex-grow":   {Option: "tui.WithFlexGrow(1)", NeedsImport: ""},
	"flex-shrink": {Option: "tui.WithFlexShrink(1)", NeedsImport: ""},

	// Justify content
	"justify-start":   {Option: "tui.WithJustify(tui.JustifyStart)", NeedsImport: "tui"},
	"justify-center":  {Option: "tui.WithJustify(tui.JustifyCenter)", NeedsImport: "tui"},
	"justify-end":     {Option: "tui.WithJustify(tui.JustifyEnd)", NeedsImport: "tui"},
	"justify-between": {Option: "tui.WithJustify(tui.JustifySpaceBetween)", NeedsImport: "tui"},
	"justify-evenly":  {Option: "tui.WithJustify(tui.JustifySpaceEvenly)", NeedsImport: "tui"},
	"justify-around":  {Option: "tui.WithJustify(tui.JustifySpaceAround)", NeedsImport: "tui"},

	// Align items
	"items-start":   {Option: "tui.WithAlign(tui.AlignStart)", NeedsImport: "tui"},
	"items-center":  {Option: "tui.WithAlign(tui.AlignCenter)", NeedsImport: "tui"},
	"items-end":     {Option: "tui.WithAlign(tui.AlignEnd)", NeedsImport: "tui"},
	"items-stretch": {Option: "tui.WithAlign(tui.AlignStretch)", NeedsImport: "tui"},

	// Self-alignment
	"self-start":   {Option: "tui.WithAlignSelf(tui.AlignStart)", NeedsImport: "tui"},
	"self-end":     {Option: "tui.WithAlignSelf(tui.AlignEnd)", NeedsImport: "tui"},
	"self-center":  {Option: "tui.WithAlignSelf(tui.AlignCenter)", NeedsImport: "tui"},
	"self-stretch": {Option: "tui.WithAlignSelf(tui.AlignStretch)", NeedsImport: "tui"},

	// Text alignment
	"text-left":   {Option: "tui.WithTextAlign(tui.TextAlignLeft)", NeedsImport: ""},
	"text-center": {Option: "tui.WithTextAlign(tui.TextAlignCenter)", NeedsImport: ""},
	"text-right":  {Option: "tui.WithTextAlign(tui.TextAlignRight)", NeedsImport: ""},

	// Borders
	"border":         {Option: "tui.WithBorder(tui.BorderSingle)", NeedsImport: "tui"},
	"border-single":  {Option: "tui.WithBorder(tui.BorderSingle)", NeedsImport: "tui"},
	"border-rounded": {Option: "tui.WithBorder(tui.BorderRounded)", NeedsImport: "tui"},
	"border-double":  {Option: "tui.WithBorder(tui.BorderDouble)", NeedsImport: "tui"},
	"border-thick":   {Option: "tui.WithBorder(tui.BorderThick)", NeedsImport: "tui"},

	// Border colors
	"border-red":     {Option: "tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Red))", NeedsImport: "tui"},
	"border-green":   {Option: "tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Green))", NeedsImport: "tui"},
	"border-blue":    {Option: "tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Blue))", NeedsImport: "tui"},
	"border-cyan":    {Option: "tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan))", NeedsImport: "tui"},
	"border-magenta": {Option: "tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Magenta))", NeedsImport: "tui"},
	"border-yellow":  {Option: "tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Yellow))", NeedsImport: "tui"},
	"border-white":   {Option: "tui.WithBorderStyle(tui.NewStyle().Foreground(tui.White))", NeedsImport: "tui"},
	"border-black":   {Option: "tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Black))", NeedsImport: "tui"},

	// Text styles
	"font-bold":  {IsTextStyle: true, TextMethod: "Bold()"},
	"font-dim":   {IsTextStyle: true, TextMethod: "Dim()"},
	"text-dim":   {IsTextStyle: true, TextMethod: "Dim()"},
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
	"bg-red":     {Option: "tui.WithBackground(tui.NewStyle().Background(tui.Red))", NeedsImport: "tui"},
	"bg-green":   {Option: "tui.WithBackground(tui.NewStyle().Background(tui.Green))", NeedsImport: "tui"},
	"bg-blue":    {Option: "tui.WithBackground(tui.NewStyle().Background(tui.Blue))", NeedsImport: "tui"},
	"bg-cyan":    {Option: "tui.WithBackground(tui.NewStyle().Background(tui.Cyan))", NeedsImport: "tui"},
	"bg-magenta": {Option: "tui.WithBackground(tui.NewStyle().Background(tui.Magenta))", NeedsImport: "tui"},
	"bg-yellow":  {Option: "tui.WithBackground(tui.NewStyle().Background(tui.Yellow))", NeedsImport: "tui"},
	"bg-white":   {Option: "tui.WithBackground(tui.NewStyle().Background(tui.White))", NeedsImport: "tui"},
	"bg-black":   {Option: "tui.WithBackground(tui.NewStyle().Background(tui.Black))", NeedsImport: "tui"},

	// Scroll
	"overflow-scroll":   {Option: "tui.WithScrollable(tui.ScrollBoth)", NeedsImport: ""},
	"overflow-y-scroll": {Option: "tui.WithScrollable(tui.ScrollVertical)", NeedsImport: ""},
	"overflow-x-scroll": {Option: "tui.WithScrollable(tui.ScrollHorizontal)", NeedsImport: ""},
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
