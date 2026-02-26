package main

import (
	"fmt"

	tui "github.com/grindlemire/go-tui"
)

type statusApp struct {
	value *tui.State[int]
}

func StatusApp() *statusApp {
	return &statusApp{
		value: tui.NewState(0),
	}
}

func (s *statusApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRune('+', func(ke tui.KeyEvent) {
			s.value.Update(func(v int) int { return v + 1 })
		}),
		tui.OnRune('-', func(ke tui.KeyEvent) {
			s.value.Update(func(v int) int { return v - 1 })
		}),
	}
}

func valueStyle(v int) tui.Style {
	if v > 0 {
		return tui.NewStyle().Bold().Foreground(tui.Green)
	}
	if v < 0 {
		return tui.NewStyle().Bold().Foreground(tui.Red)
	}
	return tui.NewStyle().Dim()
}

templ (s *statusApp) Render() {
	<div class="flex-col items-center justify-center h-full gap-1">
		<span textStyle={valueStyle(s.value.Get())}>{fmt.Sprintf("Value: %d", s.value.Get())}</span>
		<span class="font-dim">Press + / - to change, Esc to quit</span>
	</div>
}
