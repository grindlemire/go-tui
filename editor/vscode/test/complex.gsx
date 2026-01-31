package example

import (
	"fmt"
	"strings"
	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
)

// For loops and conditionals
templ ItemList(items []string, selected int) {
	<div direction={layout.Column} gap={1}>
		@for i, item := range items {
			@if i == selected {
				<div class="border-single">
					<span>{fmt.Sprintf("> %s", item)}</span>
				</div>
			} @else {
				<span>{fmt.Sprintf("  %s", item)}</span>
			}
		}
	</div>
}

// Let bindings
templ Counter(count int, label string) {
	@let countText = <span class="font-bold">{fmt.Sprintf("%d", count)}</span>
	<div class="flex-col gap-1 p-1">
		<span>{label}</span>
		{countText}
	</div>
}

// Conditional rendering
templ ConditionalContent(showHeader bool, showFooter bool) {
	<div class="flex-col">
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

// Component composition
templ Dashboard(user string, items []string, count int) {
	<div class="flex-col gap-2 p-2">
		@Header(fmt.Sprintf("Welcome, %s!", user))
		<div class="flex gap-1">
			@ItemList(items, 0)
			@Counter(count, "Total:")
		</div>
		@Footer()
	</div>
}

// Named refs - simple, loop, keyed, conditional
templ RefsExample(items []string, users map[string]string, showWarning bool) {
	<div #Container class="flex-col gap-1">
		<span #Title class="font-bold">Dashboard</span>

		@for _, item := range items {
			<span #Items>{item}</span>
		}

		@for id, name := range users {
			<span #Users key={id}>{name}</span>
		}

		@if showWarning {
			<div #Warning class="text-red border-single p-1">
				<span>Warning!</span>
			</div>
		}
	</div>
}

// State variables and reactive patterns
templ StatefulCounter() {
	count := tui.NewState(0)
	label := tui.NewState("Counter")
	<div class="flex-col gap-1 p-1 border-rounded">
		<span class="font-bold">{label.Get()}</span>
		<span>{fmt.Sprintf("Count: %d", count.Get())}</span>
		<button onClick={increment(count)}>+</button>
		<button onClick={decrement(count)}>-</button>
	</div>
}

// Event handlers
templ InteractiveElement() {
	<div class="flex-col gap-1">
		<button onClick={handleClick} onFocus={handleFocus}>
			<span>Click me</span>
		</button>
		<div focusable={true} onBlur={handleBlur} onKeyPress={handleKey}>
			<span>Focusable area</span>
		</div>
		<div onEvent={handleGenericEvent}>
			<span>Generic event handler</span>
		</div>
	</div>
}

// Numeric values
templ NumericValues() {
	<div>
		<span padding={10}>Integer: {42}</span>
		<span>{3.14159}</span>
		<span>{0xFF}</span>
		<span>{0b1010}</span>
		<span>{0o755}</span>
	</div>
}

// String values
templ StringValues() {
	<div>
		<span>{"Hello, World!"}</span>
		<span>{`Raw string with newlines`}</span>
		<span>{"Escaped: \n\t\""}</span>
	</div>
}

// Self-closing elements
templ SelfClosing() {
	<div>
		<hr />
		<br />
	</div>
}

// Attributes with various types
templ AttributeTypes(enabled bool, size int) {
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

// Helper function
func helperFunction(s string) string {
	return fmt.Sprintf("[%s]", strings.ToUpper(s))
}
