// Simple GSX component example
// Tests basic syntax highlighting

package example

import (
	"github.com/grindlemire/go-tui/pkg/tui"
)

func Header(title string) Element {
	<div border={tui.BorderSingle} padding={1}>
		<span>{title}</span>
	</div>
}

func Footer() Element {
	<div padding={1}>
		<span>Footer content</span>
	</div>
}

func SimpleCard(title string, content string) Element {
	<div border={tui.BorderRounded}>
		<span class="font-bold">{title}</span>
		<span>{content}</span>
	</div>
}
