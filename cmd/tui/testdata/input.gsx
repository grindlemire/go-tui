package testdata

import tui "github.com/grindlemire/go-tui"

type myInput struct {
	app *tui.App
}

func MyInput() *myInput {
	return &myInput{}
}

templ (c *myInput) Render() {
	<div class="flex-col gap-1">
		<input placeholder="Type here..." value="hello" width={30} border={tui.BorderRounded} onSubmit={c.handleSubmit} />
	</div>
}

func (c *myInput) handleSubmit(text string) {}
