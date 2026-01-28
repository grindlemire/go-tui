package example

import (
	"fmt"
	"strings"
	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
)

func ItemList(items []string, selected int) Element {
	<div direction={layout.Column} gap={1}>
		@for i, item := range items {
			@if i == selected {
				<div border={tui.BorderSingle}>
					<span>{fmt.Sprintf("> %s", item)}</span>
				</div>
			} @else {
				<span>{fmt.Sprintf("  %s", item)}</span>
			}
		}
	</div>
}

func Counter(count int, label string) Element {
	@let countText = <span class="font-bold">{fmt.Sprintf("%d", count)}</span>
	<div direction={layout.Column} gap={1} padding={1}>
		<span>{label}</span>
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
	<div>
		<span>{helperFunction(text)}</span>
	</div>
}

func Dashboard(user string, items []string, count int) Element {
	<div class="flex-col gap-2 p-2">
		@Header(fmt.Sprintf("Welcome, %s!", user))
		<div class="flex gap-1">
			@ItemList(items, 0)
			@Counter(count, "Total:")
		</div>
		@Footer()
	</div>
}

func NumericValues() Element {
	<div>
		<span padding={10}>
			Integer
			{42}
		</span>
		<span padding={3.14}>
			Float
			{3.14159}
		</span>
		<span>{0xFF}</span>
		<span>{0b1010}</span>
		<span>{0o755}</span>
	</div>
}

func StringValues() Element {
	<div>
		<span>{"Hello, World!"}</span>
		<span>
			{`Raw string
with newlines`}
		</span>
		<span>{"Escaped: \n\t\""}</span>
	</div>
}

func SelfClosing() Element {
	<div>
		<hr />
		<br />
		<spacer height={1} />
	</div>
}

func AttributeTypes(enabled bool, size int) Element {
	<div border={tui.BorderDouble}
	     padding={2}
	     margin={1}
	     width="100%"
	     height={size}
	     visible={enabled && size > 0}
	     class="flex-col items-center">
		<span>Content</span>
	</div>
}

func AttributeTypes2(enabled bool, size int) Element {
	<div width="100%"
	     border={tui.BorderDouble}
	     visible={enabled && size > 0}
	     class="flex-col items-center">
		<span>Content</span>
	</div>
}

func helperFunction(s string) string {
	return fmt.Sprintf("[%s]", strings.ToUpper(s))
}
