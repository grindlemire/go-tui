package testdata

import (
	tui "github.com/grindlemire/go-tui"
)

type myForm struct {
	app *tui.App
}

func MyForm() *myForm {
	return &myForm{}
}

templ (c *myForm) Render() {
	<div class="flex-col gap-1">
		<textarea placeholder="Enter text..." width={50} border={tui.BorderRounded} onSubmit={c.handleSubmit} />
	</div>
}

func (c *myForm) handleSubmit(text string) {}
