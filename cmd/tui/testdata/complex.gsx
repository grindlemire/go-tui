// Complex.gsx
package testdata

import (
	"fmt"
	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
)

// Unassigned block comment
// For package comment

// ItemList test
func ItemList(items []string, selected int) Element {
	<div direction={layout.Column} gap={1}>
		// ItemList direction
		@for i, item := range items {
			// ItemList for loop
			@if i == selected {
				<div border={tui.BorderSingle}>
					// ItemList border
					<span>{fmt.Sprintf("> %s", /* ItemList item */ item)}</span>
				</div>
			} @else {
				// ItemList else
				<span>{fmt.Sprintf("  %s", item)}</span>
			}
		}
	</div>
}

/*
Counter
tests block comment
*/
func Counter(count int, label string) Element {
	@let countText = <span>{fmt.Sprintf("%d", count)}</span>
	<div direction={layout.Column} gap={1} padding={1}>
		<span class="font-bold">{label}</span>
		{countText}
	</div>
}

func ConditionalContent(showHeader bool, showFooter bool) Element {
	<div direction={layout.Column}>
		@if showHeader {
			<span>Header</span>
		}
		<span>Main Content</span>
		@if showFooter {
			<span>Footer</span>
		} @else {
			<span>No Footer</span>
		}
	</div>
}

func WithHelper(text string) Element {
	shouldShowHeader := true
	otherHelperFunction("test")
	<div>
		<span>{helperFunction(text)}</span>
		@if shouldShowHeader {
			@ConditionalContent(true, false)
		} @else {
			<span>False</span>
		}
	</div>
}

func helperFunction(s string) string {
	return fmt.Sprintf("[%s]", s)
}
