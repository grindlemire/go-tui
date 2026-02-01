package tuigen

import (
	"strconv"
	"strings"
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
	return "tui.WithPaddingTRBL(" + strconv.Itoa(p.Top) + ", " + strconv.Itoa(p.Right) + ", " + strconv.Itoa(p.Bottom) + ", " + strconv.Itoa(p.Left) + ")"
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
	return "tui.WithMarginTRBL(" + strconv.Itoa(m.Top) + ", " + strconv.Itoa(m.Right) + ", " + strconv.Itoa(m.Bottom) + ", " + strconv.Itoa(m.Left) + ")"
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
		return TailwindMapping{Option: "tui.WithGap(" + strconv.Itoa(n) + ")"}, true
	}

	if matches := paddingPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "tui.WithPadding(" + strconv.Itoa(n) + ")"}, true
	}

	if matches := paddingXPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "tui.WithPaddingTRBL(0, " + strconv.Itoa(n) + ", 0, " + strconv.Itoa(n) + ")"}, true
	}

	if matches := paddingYPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "tui.WithPaddingTRBL(" + strconv.Itoa(n) + ", 0, " + strconv.Itoa(n) + ", 0)"}, true
	}

	if matches := marginPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "tui.WithMargin(" + strconv.Itoa(n) + ")"}, true
	}

	if matches := widthPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "tui.WithWidth(" + strconv.Itoa(n) + ")"}, true
	}

	if matches := heightPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "tui.WithHeight(" + strconv.Itoa(n) + ")"}, true
	}

	if matches := minWidthPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "tui.WithMinWidth(" + strconv.Itoa(n) + ")"}, true
	}

	if matches := maxWidthPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "tui.WithMaxWidth(" + strconv.Itoa(n) + ")"}, true
	}

	if matches := minHeightPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "tui.WithMinHeight(" + strconv.Itoa(n) + ")"}, true
	}

	if matches := maxHeightPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "tui.WithMaxHeight(" + strconv.Itoa(n) + ")"}, true
	}

	// Width fraction patterns (w-1/2, w-2/3, etc.)
	if matches := widthFractionPattern.FindStringSubmatch(class); matches != nil {
		numerator, _ := strconv.Atoi(matches[1])
		denominator, _ := strconv.Atoi(matches[2])
		if denominator != 0 {
			percent := float64(numerator) / float64(denominator) * 100
			return TailwindMapping{Option: "tui.WithWidthPercent(" + strconv.FormatFloat(percent, 'f', 2, 64) + ")"}, true
		}
	}

	// Height fraction patterns (h-1/2, h-2/3, etc.)
	if matches := heightFractionPattern.FindStringSubmatch(class); matches != nil {
		numerator, _ := strconv.Atoi(matches[1])
		denominator, _ := strconv.Atoi(matches[2])
		if denominator != 0 {
			percent := float64(numerator) / float64(denominator) * 100
			return TailwindMapping{Option: "tui.WithHeightPercent(" + strconv.FormatFloat(percent, 'f', 2, 64) + ")"}, true
		}
	}

	// Width keyword patterns (w-full, w-auto)
	if matches := widthKeywordPattern.FindStringSubmatch(class); matches != nil {
		keyword := matches[1]
		switch keyword {
		case "full":
			return TailwindMapping{Option: "tui.WithWidthPercent(100.00)"}, true
		case "auto":
			return TailwindMapping{Option: "tui.WithWidthAuto()"}, true
		}
	}

	// Height keyword patterns (h-full, h-auto)
	if matches := heightKeywordPattern.FindStringSubmatch(class); matches != nil {
		keyword := matches[1]
		switch keyword {
		case "full":
			return TailwindMapping{Option: "tui.WithHeightPercent(100.00)"}, true
		case "auto":
			return TailwindMapping{Option: "tui.WithHeightAuto()"}, true
		}
	}

	// Flex grow pattern (flex-grow-N)
	if matches := flexGrowPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "tui.WithFlexGrow(" + strconv.Itoa(n) + ")"}, true
	}

	// Flex shrink pattern (flex-shrink-N)
	if matches := flexShrinkPattern.FindStringSubmatch(class); matches != nil {
		n, _ := strconv.Atoi(matches[1])
		return TailwindMapping{Option: "tui.WithFlexShrink(" + strconv.Itoa(n) + ")"}, true
	}

	// Gradient patterns
	if matches := textGradientPattern.FindStringSubmatch(class); matches != nil {
		var startColorName, endColorName string
		direction := "tui.GradientHorizontal"
		
		// Check if class ends with a direction suffix
		if strings.HasSuffix(class, "-v") || strings.HasSuffix(class, "-dd") || strings.HasSuffix(class, "-du") || strings.HasSuffix(class, "-h") {
			// Has direction suffix - use first alternative
			startColorName = matches[1]
			endColorName = matches[2]
			if matches[3] != "" {
				switch matches[3] {
				case "v":
					direction = "tui.GradientVertical"
				case "dd":
					direction = "tui.GradientDiagonalDown"
				case "du":
					direction = "tui.GradientDiagonalUp"
				}
			}
		} else {
			// No direction suffix - parse manually by matching known color names from the end
			prefix := "text-gradient-"
			rest := strings.TrimPrefix(class, prefix)
			// Try to match known color names from the end (check longer names first to avoid partial matches)
			colorNames := []string{"bright-red", "bright-green", "bright-blue", "bright-cyan", "bright-magenta", "bright-yellow", "bright-white", "bright-black", "red", "green", "blue", "cyan", "magenta", "yellow", "white", "black"}
			for _, colorName := range colorNames {
				suffix := "-" + colorName
				if strings.HasSuffix(rest, suffix) {
					endColorName = colorName
					startColorName = strings.TrimSuffix(rest, suffix)
					break
				}
			}
			if endColorName == "" {
				// Fallback: use regex matches if available
				if matches[4] != "" && matches[5] != "" {
					startColorName = matches[4]
					endColorName = matches[5]
				} else {
					// Last resort: split on last hyphen
					lastHyphen := strings.LastIndex(rest, "-")
					if lastHyphen > 0 {
						startColorName = rest[:lastHyphen]
						endColorName = rest[lastHyphen+1:]
					}
				}
			}
		}
		
		startColor := colorNameToColor(startColorName)
		endColor := colorNameToColor(endColorName)
		option := "tui.WithTextGradient(tui.NewGradient(" + startColor + ", " + endColor + ").WithDirection(" + direction + "))"
		return TailwindMapping{Option: option, NeedsImport: "tui"}, true
	}

	if matches := bgGradientPattern.FindStringSubmatch(class); matches != nil {
		var startColorName, endColorName string
		direction := "tui.GradientHorizontal"
		
		// Check if class ends with a direction suffix
		if strings.HasSuffix(class, "-v") || strings.HasSuffix(class, "-dd") || strings.HasSuffix(class, "-du") || strings.HasSuffix(class, "-h") {
			// Has direction suffix - use first alternative
			startColorName = matches[1]
			endColorName = matches[2]
			if matches[3] != "" {
				switch matches[3] {
				case "v":
					direction = "tui.GradientVertical"
				case "dd":
					direction = "tui.GradientDiagonalDown"
				case "du":
					direction = "tui.GradientDiagonalUp"
				}
			}
		} else {
			// No direction suffix - parse manually by matching known color names from the end
			prefix := "bg-gradient-"
			rest := strings.TrimPrefix(class, prefix)
			colorNames := []string{"bright-red", "bright-green", "bright-blue", "bright-cyan", "bright-magenta", "bright-yellow", "bright-white", "bright-black", "red", "green", "blue", "cyan", "magenta", "yellow", "white", "black"}
			for _, colorName := range colorNames {
				if strings.HasSuffix(rest, "-"+colorName) {
					endColorName = colorName
					startColorName = strings.TrimSuffix(rest, "-"+colorName)
					break
				}
			}
			if endColorName == "" {
				lastHyphen := strings.LastIndex(rest, "-")
				if lastHyphen > 0 {
					startColorName = rest[:lastHyphen]
					endColorName = rest[lastHyphen+1:]
				}
			}
		}
		
		startColor := colorNameToColor(startColorName)
		endColor := colorNameToColor(endColorName)
		option := "tui.WithBackgroundGradient(tui.NewGradient(" + startColor + ", " + endColor + ").WithDirection(" + direction + "))"
		return TailwindMapping{Option: option, NeedsImport: "tui"}, true
	}

	if matches := borderGradientPattern.FindStringSubmatch(class); matches != nil {
		var startColorName, endColorName string
		direction := "tui.GradientHorizontal"
		
		// Check if class ends with a direction suffix
		if strings.HasSuffix(class, "-v") || strings.HasSuffix(class, "-dd") || strings.HasSuffix(class, "-du") || strings.HasSuffix(class, "-h") {
			// Has direction suffix - use first alternative
			startColorName = matches[1]
			endColorName = matches[2]
			if matches[3] != "" {
				switch matches[3] {
				case "v":
					direction = "tui.GradientVertical"
				case "dd":
					direction = "tui.GradientDiagonalDown"
				case "du":
					direction = "tui.GradientDiagonalUp"
				}
			}
		} else {
			// No direction suffix - parse manually by matching known color names from the end
			prefix := "border-gradient-"
			rest := strings.TrimPrefix(class, prefix)
			colorNames := []string{"bright-red", "bright-green", "bright-blue", "bright-cyan", "bright-magenta", "bright-yellow", "bright-white", "bright-black", "red", "green", "blue", "cyan", "magenta", "yellow", "white", "black"}
			for _, colorName := range colorNames {
				if strings.HasSuffix(rest, "-"+colorName) {
					endColorName = colorName
					startColorName = strings.TrimSuffix(rest, "-"+colorName)
					break
				}
			}
			if endColorName == "" {
				lastHyphen := strings.LastIndex(rest, "-")
				if lastHyphen > 0 {
					startColorName = rest[:lastHyphen]
					endColorName = rest[lastHyphen+1:]
				}
			}
		}
		
		startColor := colorNameToColor(startColorName)
		endColor := colorNameToColor(endColorName)
		option := "tui.WithBorderGradient(tui.NewGradient(" + startColor + ", " + endColor + ").WithDirection(" + direction + "))"
		return TailwindMapping{Option: option, NeedsImport: "tui"}, true
	}

	// Individual padding/margin classes - these are valid but handled separately in ParseTailwindClasses
	if _, ok := parseIndividualSpacing(class); ok {
		// Return a marker mapping - actual handling is done in ParseTailwindClasses
		return TailwindMapping{}, true
	}

	// Unknown class - silently ignore
	return TailwindMapping{}, false
}

// colorNameToColor maps a color name string to the corresponding tui.Color constant.
func colorNameToColor(name string) string {
	switch name {
	case "red":
		return "tui.Red"
	case "green":
		return "tui.Green"
	case "blue":
		return "tui.Blue"
	case "cyan":
		return "tui.Cyan"
	case "magenta":
		return "tui.Magenta"
	case "yellow":
		return "tui.Yellow"
	case "white":
		return "tui.White"
	case "black":
		return "tui.Black"
	case "bright-red":
		return "tui.BrightRed"
	case "bright-green":
		return "tui.BrightGreen"
	case "bright-blue":
		return "tui.BrightBlue"
	case "bright-cyan":
		return "tui.BrightCyan"
	case "bright-magenta":
		return "tui.BrightMagenta"
	case "bright-yellow":
		return "tui.BrightYellow"
	case "bright-white":
		return "tui.BrightWhite"
	case "bright-black":
		return "tui.BrightBlack"
	default:
		// Default to black if unknown
		return "tui.Black"
	}
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
	builder.WriteString("tui.WithTextStyle(tui.NewStyle()")
	for _, method := range methods {
		builder.WriteString(".")
		builder.WriteString(method)
	}
	builder.WriteString(")")
	return builder.String()
}
