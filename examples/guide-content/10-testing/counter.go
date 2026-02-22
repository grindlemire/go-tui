// Package main demonstrates the Testing guide patterns.
//
// This file defines a simple counter component used by the test file.
// Run tests with: go test -v ./examples/guide-content/10-testing/
package main

import (
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
