package formatter

import (
	"strings"
	"testing"
)

func TestCommentPreservationInElements(t *testing.T) {
	type tc struct {
		source   string
		expected string // if empty, expected == source (idempotent)
	}

	tests := map[string]tc{
		"comments before closing tag stay inside element": {
			source: `package main

templ Foo() {
	<div>
		<span>Hello</span>
		// comment before closing
	</div>
}
`,
		},
		"comments as only children stay inside element": {
			source: `package main

templ Foo() {
	<div>
		// only comments as children
		// another comment
	</div>
}
`,
		},
		"comments at end of nested div stay inside": {
			source: `package main

templ Foo() {
	<div>
		<div>
			<span>Text</span>
			// trailing comment in inner div
		</div>
	</div>
}
`,
		},
		"comments between elements become leading comments": {
			source: `package main

templ Foo() {
	<div>
		<span>Before</span>
		// comment 1
		// comment 2
		<span>After</span>
	</div>
}
`,
		},
		"orphan in inner div and outer div stay in place": {
			source: `package main

templ Foo() {
	<div>
		<div>
			<span>Text</span>
			// orphan inside inner div
		</div>
		<span>After inner div</span>
		// orphan inside outer div
	</div>
}
`,
		},
		"comments before closing nested divs stay in place": {
			source: `package main

templ Foo() {
	<div>
		<div>
			// comment at end of inner
		</div>
		// comment at end of outer
	</div>
}
`,
		},
		"commented out closing div preserves all comments": {
			source: `package main

templ Foo() {
	<div class="flex-col">
		<span>Bold text</span>
		<span>Dim text</span>
		// <span>Italic text</span>
		// </div>
		//
		// <br />
		// <div class="flex gap-2">
		// 	<span>Red</span>
		<span>Blue</span>
		<span>White</span>
	</div>
}
`,
		},
		"orphan comments in component body preserved": {
			source: `package main

templ Foo() {
	// orphan at start
	<div>
		<span>Text</span>
	</div>
	// orphan at end
}
`,
		},
	}

	fmtr := &Formatter{
		IndentString: "\t",
		FixImports:   false,
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			formatted, err := fmtr.Format("test.gsx", tt.source)
			if err != nil {
				t.Fatalf("format error: %v", err)
			}

			expected := tt.expected
			if expected == "" {
				expected = tt.source
			}

			// Verify no comments were lost
			sourceComments := countCommentLines(tt.source)
			formattedComments := countCommentLines(formatted)
			if formattedComments < sourceComments {
				t.Errorf("comments lost: source had %d comment lines, formatted has %d\nFormatted:\n%s",
					sourceComments, formattedComments, formatted)
			}

			// Verify comments stay at the same nesting depth
			sourceCommentDepths := commentDepths(tt.source)
			formattedCommentDepths := commentDepths(formatted)
			if sourceCommentDepths != formattedCommentDepths {
				t.Errorf("comment nesting changed:\n  source depths:    %s\n  formatted depths: %s\nFormatted:\n%s",
					sourceCommentDepths, formattedCommentDepths, formatted)
			}

			// Verify idempotency
			formatted2, err := fmtr.Format("test.gsx", formatted)
			if err != nil {
				t.Fatalf("second format error: %v", err)
			}
			if formatted2 != formatted {
				t.Errorf("not idempotent!\nFirst format:\n%s\nSecond format:\n%s", formatted, formatted2)
			}
		})
	}
}

// countCommentLines counts lines starting with // (after trimming whitespace).
func countCommentLines(s string) int {
	count := 0
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			count++
		}
	}
	return count
}

// commentDepths returns a string showing the tab depth of each comment line.
// e.g., "2,2,3,3" means 4 comment lines at depths 2,2,3,3.
func commentDepths(s string) string {
	var depths []string
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			tabs := 0
			for _, ch := range line {
				if ch == '\t' {
					tabs++
				} else {
					break
				}
			}
			depths = append(depths, strings.Repeat(".", tabs))
		}
	}
	return strings.Join(depths, ",")
}
