package tuigen

import "strings"

// TailwindValidationResult contains validation results for a class
type TailwindValidationResult struct {
	Valid      bool
	Class      string
	Suggestion string // "did you mean...?" hint
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
	"width":          "w-1",
	"height":         "h-1",
	"center":         "text-center",
	"left":           "text-left",
	"right":          "text-right",
	"align-center":   "text-center",
	"align-left":     "text-left",
	"align-right":    "text-right",
	"no-grow":        "grow-0",
	"no-shrink":      "shrink-0",
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
		// Tailwind standard flex utilities
		"grow", "grow-0", "shrink", "shrink-0",
		"flex-1", "flex-auto", "flex-initial", "flex-none",
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
