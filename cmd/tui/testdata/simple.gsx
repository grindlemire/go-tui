package testdata

import "github.com/grindlemire/go-tui/pkg/tui"

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
