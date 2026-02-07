package main

import tui "github.com/grindlemire/go-tui"

type messageList struct {
	state      *AppState
	events     *tui.Events[ChatEvent]
	focusedIdx *tui.State[int]
	content    *tui.Ref
}

func MessageList(state *AppState, events *tui.Events[ChatEvent]) *messageList {
	m := &messageList{
		state:      state,
		events:     events,
		focusedIdx: tui.NewState(-1),
		content:    tui.NewRef(),
	}

	// Subscribe to done events to scroll to bottom
	events.Subscribe(func(e ChatEvent) {
		if e.Type == "token" || e.Type == "done" {
			if el := m.content.El(); el != nil {
				el.ScrollToBottom()
			}
		}
	})

	return m
}

func (m *messageList) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnRune('j', func(ke tui.KeyEvent) { m.moveDown() }),
		tui.OnRune('k', func(ke tui.KeyEvent) { m.moveUp() }),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) { m.moveDown() }),
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) { m.moveUp() }),
		tui.OnRune('g', func(ke tui.KeyEvent) { m.focusedIdx.Set(0) }),
		tui.OnRune('G', func(ke tui.KeyEvent) {
			msgs := m.state.Messages.Get()
			if len(msgs) > 0 {
				m.focusedIdx.Set(len(msgs) - 1)
			}
		}),
		tui.OnRune('c', func(ke tui.KeyEvent) { m.copyFocused() }),
		tui.OnRune('r', func(ke tui.KeyEvent) { m.retryFocused() }),
	}
}

func (m *messageList) moveDown() {
	msgs := m.state.Messages.Get()
	idx := m.focusedIdx.Get()
	if idx < len(msgs)-1 {
		m.focusedIdx.Set(idx + 1)
	}
}

func (m *messageList) moveUp() {
	idx := m.focusedIdx.Get()
	if idx > 0 {
		m.focusedIdx.Set(idx - 1)
	}
}

func (m *messageList) copyFocused() {
	msgs := m.state.Messages.Get()
	idx := m.focusedIdx.Get()
	if idx >= 0 && idx < len(msgs) {
		m.events.Emit(ChatEvent{Type: "copy", Payload: msgs[idx].Content})
	}
}

func (m *messageList) retryFocused() {
	msgs := m.state.Messages.Get()
	idx := m.focusedIdx.Get()
	if idx >= 0 && idx < len(msgs) && msgs[idx].Role == "assistant" {
		m.events.Emit(ChatEvent{Type: "retry"})
	}
}

templ (m *messageList) Render() {
	<div
		ref={m.content}
		class="flex-col flex-grow-1"
		scrollable={tui.ScrollVertical}
		focusable={true}>
		@for i, msg := range m.state.Messages.Get() {
			@MessageView(msg, i, i == m.focusedIdx.Get(), m.events)
		}
		@if len(m.state.Messages.Get()) == 0 {
			<div class="flex-col flex-grow-1 justify-center items-center">
				<span class="font-dim">{"No messages yet. Start typing below!"}</span>
			</div>
		}
	</div>
}
