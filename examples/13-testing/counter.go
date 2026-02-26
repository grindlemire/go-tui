// Package main demonstrates the Testing guide patterns.
//
// This file defines a simple counter component used by the test file.
// Run tests with: go test -v ./examples/10-testing/
package main

import (
	"fmt"

	tui "github.com/grindlemire/go-tui"
)

type counter struct {
	count *tui.State[int]
}

func NewCounter() *counter {
	return &counter{
		count: tui.NewState(0),
	}
}

func (c *counter) Render(app *tui.App) *tui.Element {
	root := tui.New(
		tui.WithDirection(tui.Column),
		tui.WithFlexGrow(1),
		tui.WithAlign(tui.AlignCenter),
		tui.WithJustify(tui.JustifyCenter),
	)

	label := tui.New(
		tui.WithText(fmt.Sprintf("Count: %d", c.count.Get())),
		tui.WithTextStyle(tui.NewStyle().Bold()),
	)
	root.AddChild(label)

	hint := tui.New(
		tui.WithText("Press + / - to change, Esc to quit"),
		tui.WithTextStyle(tui.NewStyle().Dim()),
	)
	root.AddChild(hint)

	return root
}

func (c *counter) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('+', func(ke tui.KeyEvent) {
			c.count.Update(func(v int) int { return v + 1 })
		}),
		tui.OnRune('-', func(ke tui.KeyEvent) {
			c.count.Update(func(v int) int { return v - 1 })
		}),
	}
}
