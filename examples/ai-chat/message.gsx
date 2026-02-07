package main

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type messageView struct {
	msg      Message
	index    int
	focused  bool
	events   *tui.Events[ChatEvent]
	copyBtn  *tui.Ref
	retryBtn *tui.Ref
}

func MessageView(msg Message, index int, focused bool, events *tui.Events[ChatEvent]) *messageView {
	return &messageView{
		msg:      msg,
		index:    index,
		focused:  focused,
		events:   events,
		copyBtn:  tui.NewRef(),
		retryBtn: tui.NewRef(),
	}
}

func (m *messageView) HandleMouse(me tui.MouseEvent) bool {
	return tui.HandleClicks(me,
		tui.Click(m.copyBtn, func() {
			m.events.Emit(ChatEvent{Type: "copy", Payload: m.msg.Content})
		}),
		tui.Click(m.retryBtn, func() {
			m.events.Emit(ChatEvent{Type: "retry", Payload: fmt.Sprintf("%d", m.index)})
		}),
	)
}

func (m *messageView) borderClass() string {
	if m.msg.Role == "assistant" {
		if m.focused {
			return "border-rounded border-cyan"
		}
		return "border-rounded border-blue"
	}
	if m.focused {
		return "border-rounded border-white"
	}
	return "border-rounded"
}

func (m *messageView) roleIcon() string {
	if m.msg.Role == "assistant" {
		return ""
	}
	return ""
}

func (m *messageView) roleClass() string {
	if m.msg.Role == "assistant" {
		return "text-cyan font-bold"
	}
	return "text-white font-bold"
}

templ (m *messageView) Render() {
	<div class={m.borderClass()} padding={1} margin={1}>
		<div class="flex-col gap-1">
			<div class="flex justify-between">
				<span class={m.roleClass()}>{m.roleIcon() + " " + m.msg.Role}</span>
				<div class="flex gap-1">
					@if m.msg.Duration > 0 {
						<span class="font-dim">{fmt.Sprintf("%.1fs", m.msg.Duration.Seconds())}</span>
					}
					@if m.msg.Role == "assistant" && !m.msg.Streaming {
						<button ref={m.retryBtn} class="font-dim">{"[r]"}</button>
					}
					<button ref={m.copyBtn} class="font-dim">{"[c]"}</button>
				</div>
			</div>
			<div>
				<span class="text-white">{m.msg.Content}</span>
				@if m.msg.Streaming {
					<span class="text-cyan font-bold">{""}</span>
				}
			</div>
		</div>
	</div>
}
