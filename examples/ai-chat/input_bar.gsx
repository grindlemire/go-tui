package main

import tui "github.com/grindlemire/go-tui"

type inputBar struct {
	state  *AppState
	events *tui.Events[ChatEvent]
	text   *tui.State[string]
	active *tui.State[bool]
}

func InputBar(state *AppState, events *tui.Events[ChatEvent]) *inputBar {
	return &inputBar{
		state:  state,
		events: events,
		text:   tui.NewState(""),
		active: tui.NewState(true),
	}
}

func (i *inputBar) KeyMap() tui.KeyMap {
	if !i.active.Get() || i.state.IsStreaming.Get() {
		return nil
	}
	return tui.KeyMap{
		tui.OnRunesStop(i.appendChar),
		tui.OnKeyStop(tui.KeyBackspace, i.deleteChar),
		tui.OnKeyStop(tui.KeyEnter, i.submit),
	}
}

func (i *inputBar) appendChar(ke tui.KeyEvent) {
	i.text.Set(i.text.Get() + string(ke.Rune))
}

func (i *inputBar) deleteChar(ke tui.KeyEvent) {
	t := i.text.Get()
	if len(t) > 0 {
		// Handle UTF-8 properly
		runes := []rune(t)
		i.text.Set(string(runes[:len(runes)-1]))
	}
}

func (i *inputBar) submit(ke tui.KeyEvent) {
	t := i.text.Get()
	if t == "" {
		return
	}
	i.events.Emit(ChatEvent{Type: "submit", Payload: t})
	i.text.Set("")
}

func (i *inputBar) border() tui.BorderStyle {
	return tui.BorderRounded
}

func (i *inputBar) borderStyle() tui.Style {
	if i.state.IsStreaming.Get() {
		return tui.NewStyle().Foreground(tui.Yellow)
	}
	return tui.NewStyle().Foreground(tui.Cyan)
}

templ (i *inputBar) Render() {
	<div border={i.border()} borderStyle={i.borderStyle()} height={3} padding={1} direction={tui.Row} align={tui.AlignCenter}>
		@if i.state.IsStreaming.Get() {
			<span class="text-yellow font-dim">{"  Generating..."}</span>
		} @else {
			<span class="text-cyan">{"  "}</span>
			<span class="text-white">{i.text.Get()}</span>
			<span class="text-cyan font-bold">{""}</span>
			<div class="flex-grow-1"></div>
			<span class="font-dim">{"  Enter to send"}</span>
		}
	</div>
}
