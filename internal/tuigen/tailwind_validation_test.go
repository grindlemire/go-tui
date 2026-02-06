package tuigen

import (
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	type tc struct {
		a        string
		b        string
		expected int
	}

	tests := map[string]tc{
		"identical strings": {
			a:        "flex",
			b:        "flex",
			expected: 0,
		},
		"one character different": {
			a:        "flex",
			b:        "fles",
			expected: 1,
		},
		"one character added": {
			a:        "flex",
			b:        "flexs",
			expected: 1,
		},
		"one character removed": {
			a:        "flex",
			b:        "fle",
			expected: 1,
		},
		"completely different": {
			a:        "flex",
			b:        "border",
			expected: 5, // flex->blex->bolex->borex->borde->border or equivalent 5-edit path
		},
		"empty first string": {
			a:        "",
			b:        "flex",
			expected: 4,
		},
		"empty second string": {
			a:        "flex",
			b:        "",
			expected: 4,
		},
		"both empty": {
			a:        "",
			b:        "",
			expected: 0,
		},
		"flex-col vs flex-column": {
			a:        "flex-col",
			b:        "flex-column",
			expected: 3,
		},
		"flex-columns vs flex-col": {
			a:        "flex-columns",
			b:        "flex-col",
			expected: 4,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := levenshteinDistance(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}

func TestFindSimilarClass(t *testing.T) {
	type tc struct {
		input    string
		expected string
	}

	tests := map[string]tc{
		"exact match in similarClasses map": {
			input:    "flex-columns",
			expected: "flex-col",
		},
		"flex-column typo": {
			input:    "flex-column",
			expected: "flex-col",
		},
		"bold to font-bold": {
			input:    "bold",
			expected: "font-bold",
		},
		"center to text-center": {
			input:    "center",
			expected: "text-center",
		},
		"no-grow to grow-0": {
			input:    "no-grow",
			expected: "grow-0",
		},
		"no-shrink to shrink-0": {
			input:    "no-shrink",
			expected: "shrink-0",
		},
		"padding-top to pt-1": {
			input:    "padding-top",
			expected: "pt-1",
		},
		"fuzzy match - fex to flex": {
			input:    "fex",
			expected: "flex",
		},
		"fuzzy match - border-rounde to border-rounded": {
			input:    "border-rounde",
			expected: "border-rounded",
		},
		"no match for very different string": {
			input:    "xyzabc123",
			expected: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := findSimilarClass(tt.input)
			if got != tt.expected {
				t.Errorf("findSimilarClass(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestValidateTailwindClass_Valid(t *testing.T) {
	type tc struct {
		input string
	}

	tests := map[string]tc{
		"flex":           {input: "flex"},
		"flex-col":       {input: "flex-col"},
		"flex-row":       {input: "flex-row"},
		"gap-2":          {input: "gap-2"},
		"p-4":            {input: "p-4"},
		"pt-2":           {input: "pt-2"},
		"m-1":            {input: "m-1"},
		"mt-3":           {input: "mt-3"},
		"w-full":         {input: "w-full"},
		"w-1/2":          {input: "w-1/2"},
		"h-auto":         {input: "h-auto"},
		"border":         {input: "border"},
		"border-rounded": {input: "border-rounded"},
		"font-bold":      {input: "font-bold"},
		"text-cyan":      {input: "text-cyan"},
		"bg-red":         {input: "bg-red"},
		"justify-center": {input: "justify-center"},
		"items-center":   {input: "items-center"},
		"self-start":     {input: "self-start"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := ValidateTailwindClass(tt.input)
			if !result.Valid {
				t.Errorf("ValidateTailwindClass(%q) Valid = false, want true", tt.input)
			}
			if result.Class != tt.input {
				t.Errorf("ValidateTailwindClass(%q) Class = %q, want %q", tt.input, result.Class, tt.input)
			}
		})
	}
}

func TestValidateTailwindClass_Invalid(t *testing.T) {
	type tc struct {
		input          string
		wantSuggestion string
	}

	tests := map[string]tc{
		"flex-columns typo": {
			input:          "flex-columns",
			wantSuggestion: "flex-col",
		},
		"bold typo": {
			input:          "bold",
			wantSuggestion: "font-bold",
		},
		"center typo": {
			input:          "center",
			wantSuggestion: "text-center",
		},
		"completely unknown": {
			input:          "xyzabc123",
			wantSuggestion: "",
		},
		"empty string": {
			input:          "",
			wantSuggestion: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := ValidateTailwindClass(tt.input)
			if result.Valid {
				t.Errorf("ValidateTailwindClass(%q) Valid = true, want false", tt.input)
			}
			if result.Suggestion != tt.wantSuggestion {
				t.Errorf("ValidateTailwindClass(%q) Suggestion = %q, want %q", tt.input, result.Suggestion, tt.wantSuggestion)
			}
		})
	}
}

func TestParseTailwindClassesWithPositions(t *testing.T) {
	type tc struct {
		input        string
		attrStartCol int
		wantCount    int
		checkClass   int    // index of class to check
		wantClass    string
		wantStartCol int
		wantEndCol   int
		wantValid    bool
	}

	tests := map[string]tc{
		"single valid class": {
			input:        "flex",
			attrStartCol: 10,
			wantCount:    1,
			checkClass:   0,
			wantClass:    "flex",
			wantStartCol: 10,
			wantEndCol:   14,
			wantValid:    true,
		},
		"multiple classes": {
			input:        "flex gap-2",
			attrStartCol: 5,
			wantCount:    2,
			checkClass:   1,
			wantClass:    "gap-2",
			wantStartCol: 10,
			wantEndCol:   15,
			wantValid:    true,
		},
		"invalid class": {
			input:        "flex-columns",
			attrStartCol: 0,
			wantCount:    1,
			checkClass:   0,
			wantClass:    "flex-columns",
			wantStartCol: 0,
			wantEndCol:   12,
			wantValid:    false,
		},
		"mixed valid and invalid": {
			input:        "flex flex-columns gap-2",
			attrStartCol: 0,
			wantCount:    3,
			checkClass:   1,
			wantClass:    "flex-columns",
			wantStartCol: 5,
			wantEndCol:   17,
			wantValid:    false,
		},
		"with extra whitespace": {
			input:        "  flex   gap-2  ",
			attrStartCol: 0,
			wantCount:    2,
			checkClass:   0,
			wantClass:    "flex",
			wantStartCol: 2,
			wantEndCol:   6,
			wantValid:    true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := ParseTailwindClassesWithPositions(tt.input, tt.attrStartCol)

			if len(result) != tt.wantCount {
				t.Errorf("ParseTailwindClassesWithPositions count = %d, want %d", len(result), tt.wantCount)
				return
			}

			if tt.checkClass >= len(result) {
				t.Fatalf("checkClass index %d out of range", tt.checkClass)
			}

			class := result[tt.checkClass]
			if class.Class != tt.wantClass {
				t.Errorf("Class = %q, want %q", class.Class, tt.wantClass)
			}
			if class.StartCol != tt.wantStartCol {
				t.Errorf("StartCol = %d, want %d", class.StartCol, tt.wantStartCol)
			}
			if class.EndCol != tt.wantEndCol {
				t.Errorf("EndCol = %d, want %d", class.EndCol, tt.wantEndCol)
			}
			if class.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", class.Valid, tt.wantValid)
			}
		})
	}
}

func TestParseTailwindClassesWithPositions_Suggestions(t *testing.T) {
	result := ParseTailwindClassesWithPositions("flex-columns", 0)

	if len(result) != 1 {
		t.Fatalf("expected 1 class, got %d", len(result))
	}

	if result[0].Suggestion != "flex-col" {
		t.Errorf("Suggestion = %q, want %q", result[0].Suggestion, "flex-col")
	}
}

func TestAllTailwindClasses(t *testing.T) {
	classes := AllTailwindClasses()

	// Should return a non-empty list
	if len(classes) == 0 {
		t.Error("AllTailwindClasses() returned empty list")
	}

	// Check that we have classes from different categories
	categories := make(map[string]int)
	for _, c := range classes {
		categories[c.Category]++
	}

	expectedCategories := []string{"layout", "flex", "spacing", "typography", "visual"}
	for _, cat := range expectedCategories {
		if categories[cat] == 0 {
			t.Errorf("expected category %q to have classes, got 0", cat)
		}
	}

	// Check that all classes have required fields
	for _, c := range classes {
		if c.Name == "" {
			t.Error("found class with empty Name")
		}
		if c.Category == "" {
			t.Errorf("class %q has empty Category", c.Name)
		}
		if c.Description == "" {
			t.Errorf("class %q has empty Description", c.Name)
		}
		if c.Example == "" {
			t.Errorf("class %q has empty Example", c.Name)
		}
	}
}

func TestAllTailwindClasses_SpecificClasses(t *testing.T) {
	classes := AllTailwindClasses()

	// Build a map for easy lookup
	classMap := make(map[string]TailwindClassInfo)
	for _, c := range classes {
		classMap[c.Name] = c
	}

	// Check for specific classes that should exist
	expectedClasses := []string{
		"flex", "flex-col", "flex-row",
		"flex-grow", "flex-shrink", "flex-grow-0", "flex-shrink-0",
		"justify-start", "justify-center", "justify-end", "justify-between", "justify-around", "justify-evenly",
		"items-start", "items-center", "items-end", "items-stretch",
		"self-start", "self-center", "self-end", "self-stretch",
		"gap-1", "gap-2",
		"p-1", "p-2", "px-1", "py-1", "pt-1", "pr-1", "pb-1", "pl-1",
		"m-1", "m-2", "mx-1", "my-1", "mt-1", "mr-1", "mb-1", "ml-1",
		"w-full", "w-auto", "w-1/2",
		"h-full", "h-auto", "h-1/2",
		"border", "border-rounded", "border-double", "border-thick",
		"border-red", "border-green", "border-blue", "border-cyan",
		"font-bold", "font-dim", "italic", "underline",
		"text-left", "text-center", "text-right",
		"text-red", "text-green", "text-cyan",
		"bg-red", "bg-green", "bg-blue",
		"overflow-scroll", "overflow-y-scroll", "overflow-x-scroll",
	}

	for _, name := range expectedClasses {
		if _, ok := classMap[name]; !ok {
			t.Errorf("expected class %q not found in AllTailwindClasses()", name)
		}
	}
}
