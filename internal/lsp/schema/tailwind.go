package schema

import (
	"fmt"
	"strings"
)

// TailwindClassDef describes a Tailwind-style utility class.
type TailwindClassDef struct {
	Name        string
	Description string
	Category    string // "layout", "flex", "spacing", "typography", "visual", "scroll"
}

// GetClassDoc returns hover documentation for a Tailwind-style class.
// Returns empty string if the class is not recognized.
func GetClassDoc(class string) string {
	// Check static classes first
	if doc, ok := staticClassDocs[class]; ok {
		return doc
	}

	// Check parameterized patterns
	for _, p := range paramPatterns {
		if strings.HasPrefix(class, p.prefix) {
			val := strings.TrimPrefix(class, p.prefix)
			if val != "" {
				return fmt.Sprintf(p.docFmt, val)
			}
		}
	}

	return ""
}

// MatchClasses returns all class definitions that match the given prefix.
// Used for completion filtering.
func MatchClasses(prefix string) []TailwindClassDef {
	var matches []TailwindClassDef
	for _, cls := range AllClasses {
		if prefix == "" || strings.HasPrefix(cls.Name, prefix) {
			matches = append(matches, cls)
		}
	}
	return matches
}

// paramPattern describes a parameterized class prefix with its documentation format.
type paramPattern struct {
	prefix string
	docFmt string // fmt string with one %s for the value
}

// staticClassDocs maps exact class names to their documentation strings.
var staticClassDocs = map[string]string{
	// Layout
	"flex":     "Display as flexbox with row direction.",
	"flex-col": "Display as flexbox with column direction.",
	"flex-row": "Display as flexbox with row direction (default).",

	// Alignment
	"items-start":   "Align items to the start of the cross axis.",
	"items-center":  "Align items to the center of the cross axis.",
	"items-end":     "Align items to the end of the cross axis.",
	"items-stretch": "Stretch items to fill the cross axis.",

	"self-start":   "Align self to the start of the cross axis.",
	"self-center":  "Align self to the center of the cross axis.",
	"self-end":     "Align self to the end of the cross axis.",
	"self-stretch": "Stretch self to fill the cross axis.",

	"justify-start":   "Justify content to the start of the main axis.",
	"justify-center":  "Justify content to the center of the main axis.",
	"justify-end":     "Justify content to the end of the main axis.",
	"justify-between": "Distribute items with space between them.",
	"justify-around":  "Distribute items with space around them.",
	"justify-evenly":  "Distribute items with equal space around them.",

	// Borders
	"border":         "Default border style.",
	"border-none":    "No border.",
	"border-single":  "Single line border: `\u250c\u2500\u2510\u2502\u2514\u2500\u2518`",
	"border-double":  "Double line border: `\u2554\u2550\u2557\u2551\u255a\u2550\u255d`",
	"border-rounded": "Rounded border: `\u256d\u2500\u256e\u2502\u2570\u2500\u256f`",
	"border-thick":   "Thick border: `\u250f\u2501\u2513\u2503\u2517\u2501\u251b`",

	// Font styles
	"font-bold":          "Bold text style.",
	"font-dim":           "Dim/faint text style.",
	"font-italic":        "Italic text style.",
	"italic":             "Italic text style.",
	"font-underline":     "Underlined text style.",
	"underline":          "Underlined text style.",
	"font-blink":         "Blinking text style.",
	"blink":              "Blinking text style.",
	"font-reverse":       "Reverse video (swap foreground/background).",
	"reverse":            "Reverse video (swap foreground/background).",
	"font-strikethrough": "Strikethrough text style.",
	"strikethrough":      "Strikethrough text style.",

	// Text alignment
	"text-left":   "Align text to the left.",
	"text-center": "Align text to the center.",
	"text-right":  "Align text to the right.",

	// Size keywords
	"w-full":  "Set width to 100%.",
	"w-auto":  "Set width to auto (size to content).",
	"h-full":  "Set height to 100%.",
	"h-auto":  "Set height to auto (size to content).",
	"w-1/2":   "Set width to 50%.",
	"w-1/3":   "Set width to 33.3%.",
	"w-2/3":   "Set width to 66.7%.",
	"w-1/4":   "Set width to 25%.",
	"w-3/4":   "Set width to 75%.",

	// Flex grow/shrink keywords
	"grow":     "Set flex grow factor to 1.",
	"shrink":   "Set flex shrink factor to 1.",
	"shrink-0": "Prevent element from shrinking.",

	// Scroll
	"overflow-scroll":   "Enable scrolling for overflow content.",
	"overflow-y-scroll": "Enable vertical scrolling for overflow content.",
	"overflow-x-scroll": "Enable horizontal scrolling for overflow content.",

	// Gradients
	"text-gradient-red-blue":     "Apply a horizontal gradient to text from red to blue.",
	"bg-gradient-red-blue":        "Apply a horizontal gradient to background from red to blue.",
	"border-gradient-red-blue":    "Apply a horizontal gradient to border from red to blue.",
}

// paramPatterns defines parameterized class prefixes and their documentation.
var paramPatterns = []paramPattern{
	// Gap
	{prefix: "gap-", docFmt: "Set gap between children to %s characters."},

	// Padding
	{prefix: "p-", docFmt: "Set padding to %s on all sides."},
	{prefix: "px-", docFmt: "Set horizontal padding (left and right) to %s."},
	{prefix: "py-", docFmt: "Set vertical padding (top and bottom) to %s."},
	{prefix: "pt-", docFmt: "Set top padding to %s."},
	{prefix: "pb-", docFmt: "Set bottom padding to %s."},
	{prefix: "pl-", docFmt: "Set left padding to %s."},
	{prefix: "pr-", docFmt: "Set right padding to %s."},

	// Margin
	{prefix: "m-", docFmt: "Set margin to %s on all sides."},
	{prefix: "mx-", docFmt: "Set horizontal margin (left and right) to %s."},
	{prefix: "my-", docFmt: "Set vertical margin (top and bottom) to %s."},
	{prefix: "mt-", docFmt: "Set top margin to %s."},
	{prefix: "mb-", docFmt: "Set bottom margin to %s."},
	{prefix: "ml-", docFmt: "Set left margin to %s."},
	{prefix: "mr-", docFmt: "Set right margin to %s."},

	// Gradients (check before generic text/bg/border patterns)
	{prefix: "text-gradient-", docFmt: "Apply a gradient to text from **%s** to another color. Use format: text-gradient-COLOR1-COLOR2[-DIRECTION] where DIRECTION is h (horizontal, default), v (vertical), dd (diagonal down), or du (diagonal up)."},
	{prefix: "bg-gradient-", docFmt: "Apply a gradient to background from **%s** to another color. Use format: bg-gradient-COLOR1-COLOR2[-DIRECTION] where DIRECTION is h (horizontal, default), v (vertical), dd (diagonal down), or du (diagonal up)."},
	{prefix: "border-gradient-", docFmt: "Apply a gradient to border from **%s** to another color. Use format: border-gradient-COLOR1-COLOR2[-DIRECTION] where DIRECTION is h (horizontal, default), v (vertical), dd (diagonal down), or du (diagonal up)."},

	// Text colors
	{prefix: "text-", docFmt: "Set text color to **%s**."},

	// Background colors
	{prefix: "bg-", docFmt: "Set background color to **%s**."},

	// Border colors
	{prefix: "border-", docFmt: "Set border color to **%s**."},

	// Width/height
	{prefix: "w-", docFmt: "Set width to %s characters."},
	{prefix: "h-", docFmt: "Set height to %s rows."},
	{prefix: "min-w-", docFmt: "Set minimum width to %s."},
	{prefix: "min-h-", docFmt: "Set minimum height to %s."},
	{prefix: "max-w-", docFmt: "Set maximum width to %s."},
	{prefix: "max-h-", docFmt: "Set maximum height to %s."},

	// Flex factors
	{prefix: "grow-", docFmt: "Set flex grow factor to %s."},
	{prefix: "shrink-", docFmt: "Set flex shrink factor to %s."},
	{prefix: "flex-grow-", docFmt: "Set flex grow factor to %s."},
	{prefix: "flex-shrink-", docFmt: "Set flex shrink factor to %s."},
}

// AllClasses is the list of all known Tailwind-style classes for completion.
// This includes static classes and common parameterized examples.
var AllClasses = []TailwindClassDef{
	// Layout
	{Name: "flex", Description: "Display as flexbox with row direction", Category: "layout"},
	{Name: "flex-col", Description: "Display as flexbox with column direction", Category: "layout"},
	{Name: "flex-row", Description: "Display as flexbox with row direction (default)", Category: "layout"},

	// Alignment
	{Name: "items-start", Description: "Align items to the start of the cross axis", Category: "flex"},
	{Name: "items-center", Description: "Align items to the center of the cross axis", Category: "flex"},
	{Name: "items-end", Description: "Align items to the end of the cross axis", Category: "flex"},
	{Name: "items-stretch", Description: "Stretch items to fill the cross axis", Category: "flex"},
	{Name: "self-start", Description: "Align self to the start of the cross axis", Category: "flex"},
	{Name: "self-center", Description: "Align self to the center of the cross axis", Category: "flex"},
	{Name: "self-end", Description: "Align self to the end of the cross axis", Category: "flex"},
	{Name: "self-stretch", Description: "Stretch self to fill the cross axis", Category: "flex"},
	{Name: "justify-start", Description: "Justify content to the start", Category: "flex"},
	{Name: "justify-center", Description: "Justify content to the center", Category: "flex"},
	{Name: "justify-end", Description: "Justify content to the end", Category: "flex"},
	{Name: "justify-between", Description: "Distribute items with space between", Category: "flex"},
	{Name: "justify-around", Description: "Distribute items with space around", Category: "flex"},
	{Name: "justify-evenly", Description: "Distribute items with equal space", Category: "flex"},

	// Spacing
	{Name: "gap-1", Description: "Set gap between children to 1", Category: "spacing"},
	{Name: "gap-2", Description: "Set gap between children to 2", Category: "spacing"},
	{Name: "gap-3", Description: "Set gap between children to 3", Category: "spacing"},
	{Name: "gap-4", Description: "Set gap between children to 4", Category: "spacing"},
	{Name: "p-1", Description: "Set padding to 1 on all sides", Category: "spacing"},
	{Name: "p-2", Description: "Set padding to 2 on all sides", Category: "spacing"},
	{Name: "p-3", Description: "Set padding to 3 on all sides", Category: "spacing"},
	{Name: "p-4", Description: "Set padding to 4 on all sides", Category: "spacing"},
	{Name: "px-1", Description: "Set horizontal padding to 1", Category: "spacing"},
	{Name: "px-2", Description: "Set horizontal padding to 2", Category: "spacing"},
	{Name: "py-1", Description: "Set vertical padding to 1", Category: "spacing"},
	{Name: "py-2", Description: "Set vertical padding to 2", Category: "spacing"},
	{Name: "m-1", Description: "Set margin to 1 on all sides", Category: "spacing"},
	{Name: "m-2", Description: "Set margin to 2 on all sides", Category: "spacing"},

	// Typography
	{Name: "font-bold", Description: "Bold text style", Category: "typography"},
	{Name: "font-dim", Description: "Dim/faint text style", Category: "typography"},
	{Name: "font-italic", Description: "Italic text style", Category: "typography"},
	{Name: "italic", Description: "Italic text style", Category: "typography"},
	{Name: "underline", Description: "Underlined text style", Category: "typography"},
	{Name: "strikethrough", Description: "Strikethrough text style", Category: "typography"},
	{Name: "blink", Description: "Blinking text style", Category: "typography"},
	{Name: "reverse", Description: "Reverse video (swap fg/bg)", Category: "typography"},
	{Name: "text-left", Description: "Align text to the left", Category: "typography"},
	{Name: "text-center", Description: "Align text to the center", Category: "typography"},
	{Name: "text-right", Description: "Align text to the right", Category: "typography"},

	// Text colors
	{Name: "text-red", Description: "Set text color to red", Category: "typography"},
	{Name: "text-green", Description: "Set text color to green", Category: "typography"},
	{Name: "text-blue", Description: "Set text color to blue", Category: "typography"},
	{Name: "text-cyan", Description: "Set text color to cyan", Category: "typography"},
	{Name: "text-magenta", Description: "Set text color to magenta", Category: "typography"},
	{Name: "text-yellow", Description: "Set text color to yellow", Category: "typography"},
	{Name: "text-white", Description: "Set text color to white", Category: "typography"},
	{Name: "text-black", Description: "Set text color to black", Category: "typography"},

	// Bright text colors
	{Name: "text-bright-red", Description: "Set text color to bright red", Category: "typography"},
	{Name: "text-bright-green", Description: "Set text color to bright green", Category: "typography"},
	{Name: "text-bright-blue", Description: "Set text color to bright blue", Category: "typography"},
	{Name: "text-bright-cyan", Description: "Set text color to bright cyan", Category: "typography"},
	{Name: "text-bright-magenta", Description: "Set text color to bright magenta", Category: "typography"},
	{Name: "text-bright-yellow", Description: "Set text color to bright yellow", Category: "typography"},
	{Name: "text-bright-white", Description: "Set text color to bright white", Category: "typography"},
	{Name: "text-bright-black", Description: "Set text color to bright black (gray)", Category: "typography"},

	// Background colors
	{Name: "bg-red", Description: "Set background to red", Category: "visual"},
	{Name: "bg-green", Description: "Set background to green", Category: "visual"},
	{Name: "bg-blue", Description: "Set background to blue", Category: "visual"},
	{Name: "bg-cyan", Description: "Set background to cyan", Category: "visual"},
	{Name: "bg-magenta", Description: "Set background to magenta", Category: "visual"},
	{Name: "bg-yellow", Description: "Set background to yellow", Category: "visual"},
	{Name: "bg-white", Description: "Set background to white", Category: "visual"},
	{Name: "bg-black", Description: "Set background to black", Category: "visual"},

	// Bright background colors
	{Name: "bg-bright-red", Description: "Set background to bright red", Category: "visual"},
	{Name: "bg-bright-green", Description: "Set background to bright green", Category: "visual"},
	{Name: "bg-bright-blue", Description: "Set background to bright blue", Category: "visual"},
	{Name: "bg-bright-cyan", Description: "Set background to bright cyan", Category: "visual"},
	{Name: "bg-bright-magenta", Description: "Set background to bright magenta", Category: "visual"},
	{Name: "bg-bright-yellow", Description: "Set background to bright yellow", Category: "visual"},
	{Name: "bg-bright-white", Description: "Set background to bright white", Category: "visual"},
	{Name: "bg-bright-black", Description: "Set background to bright black (dark gray)", Category: "visual"},

	// Borders
	{Name: "border", Description: "Default border style", Category: "visual"},
	{Name: "border-none", Description: "No border", Category: "visual"},
	{Name: "border-single", Description: "Single line border", Category: "visual"},
	{Name: "border-double", Description: "Double line border", Category: "visual"},
	{Name: "border-rounded", Description: "Rounded border", Category: "visual"},
	{Name: "border-thick", Description: "Thick border", Category: "visual"},
	{Name: "border-red", Description: "Set border color to red", Category: "visual"},
	{Name: "border-green", Description: "Set border color to green", Category: "visual"},
	{Name: "border-blue", Description: "Set border color to blue", Category: "visual"},
	{Name: "border-cyan", Description: "Set border color to cyan", Category: "visual"},

	// Sizing
	{Name: "w-full", Description: "Set width to 100%", Category: "layout"},
	{Name: "w-auto", Description: "Set width to auto", Category: "layout"},
	{Name: "h-full", Description: "Set height to 100%", Category: "layout"},
	{Name: "h-auto", Description: "Set height to auto", Category: "layout"},
	{Name: "w-1/2", Description: "Set width to 50%", Category: "layout"},
	{Name: "w-1/3", Description: "Set width to 33.3%", Category: "layout"},
	{Name: "w-2/3", Description: "Set width to 66.7%", Category: "layout"},

	// Flex grow/shrink
	{Name: "grow", Description: "Set flex grow factor to 1", Category: "flex"},
	{Name: "grow-0", Description: "Set flex grow factor to 0", Category: "flex"},
	{Name: "shrink", Description: "Set flex shrink factor to 1", Category: "flex"},
	{Name: "shrink-0", Description: "Prevent element from shrinking", Category: "flex"},

	// Scroll
	{Name: "overflow-scroll", Description: "Enable scrolling for overflow content", Category: "scroll"},
	{Name: "overflow-y-scroll", Description: "Enable vertical scrolling", Category: "scroll"},
	{Name: "overflow-x-scroll", Description: "Enable horizontal scrolling", Category: "scroll"},

	// Gradients
	{Name: "text-gradient-red-blue", Description: "Apply horizontal gradient to text from red to blue", Category: "visual"},
	{Name: "text-gradient-cyan-magenta", Description: "Apply horizontal gradient to text from cyan to magenta", Category: "visual"},
	{Name: "text-gradient-red-blue-v", Description: "Apply vertical gradient to text from red to blue", Category: "visual"},
	{Name: "bg-gradient-red-blue", Description: "Apply horizontal gradient to background from red to blue", Category: "visual"},
	{Name: "bg-gradient-cyan-magenta", Description: "Apply horizontal gradient to background from cyan to magenta", Category: "visual"},
	{Name: "bg-gradient-red-blue-dd", Description: "Apply diagonal-down gradient to background from red to blue", Category: "visual"},
	{Name: "border-gradient-yellow-red", Description: "Apply horizontal gradient to border from yellow to red", Category: "visual"},
}
