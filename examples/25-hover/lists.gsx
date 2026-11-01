package main

import (
	"fmt"
	"math/rand/v2"
	"slices"
	tui "github.com/grindlemire/go-tui"
)

type message struct {
	id      int
	subject string
	unread  bool
}

// listsPage shows hover on a keyed list whose contents change under a still
// pointer: shuffling, live inserts at the top, and rows whose hover color
// comes from a runtime class expression.
type listsPage struct {
	d      *hoverDemo
	msgs   *tui.State[[]message]
	live   *tui.State[bool]
	nextID int
	rows   *tui.RefMap[int]
}

func newListsPage(d *hoverDemo) *listsPage {
	p := &listsPage{d: d, live: tui.NewState(false), rows: tui.NewRefMap[int]()}
	var msgs []message
	for i, subject := range inbox[:8] {
		msgs = append(msgs, message{id: i, subject: subject, unread: i%3 == 0})
	}
	p.nextID = len(msgs)
	p.msgs = tui.NewState(msgs)
	return p
}

// tick runs every two seconds on the root's timer; it inserts only while
// live inserts are on.
func (p *listsPage) tick() {
	if !p.live.Get() {
		return
	}
	msg := message{id: p.nextID, subject: fmt.Sprintf("New message #%d", p.nextID), unread: true}
	p.nextID++
	p.msgs.Set(append([]message{msg}, p.msgs.Get()...)[:min(len(p.msgs.Get())+1, 20)])
}

func (p *listsPage) shuffle() {
	msgs := slices.Clone(p.msgs.Get())
	rand.Shuffle(len(msgs), func(i, j int) { msgs[i], msgs[j] = msgs[j], msgs[i] })
	p.msgs.Set(msgs)
}

func (p *listsPage) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.On(tui.Rune('s'), func(ke tui.KeyEvent) { p.shuffle() }),
		tui.On(tui.Rune('t'), func(ke tui.KeyEvent) { p.live.Set(!p.live.Get()) }),
	}
}

// Clicking a row toggles read/unread. Rows sit inside a scrolled container,
// which ContainsPoint accounts for.
func (p *listsPage) HandleMouse(me tui.MouseEvent) bool {
	if me.Button != tui.MouseLeft || me.Action != tui.MousePress {
		return false
	}
	for id, el := range p.rows.All() {
		if el != nil && el.ContainsPoint(me.X, me.Y) {
			msgs := slices.Clone(p.msgs.Get())
			i := slices.IndexFunc(msgs, func(m message) bool { return m.id == id })
			msgs[i].unread = !msgs[i].unread
			p.msgs.Set(msgs)
			return true
		}
	}
	return false
}

// rowClass is a runtime class expression; its hover: classes resolve at
// runtime through tui.WithClass.
func rowClass(m message) string {
	if m.unread {
		return "px-1 font-bold text-bright-white hover:bg-magenta hover:text-black"
	}
	return "px-1 text-white hover:bg-blue"
}

func rowLabel(m message) string {
	if m.unread {
		return "● " + m.subject
	}
	return "○ " + m.subject
}

func (p *listsPage) liveLabel() string {
	if p.live.Get() {
		return "live inserts: on"
	}
	return "live inserts: off"
}

templ (p *listsPage) Render() {
	<div class="flex gap-2 w-full h-full">
		<div
			class="flex-col border-rounded w-36 shrink-0 overflow-y-scroll scrollbar-bright-black"
			height={12}
			borderTitle="Keyed messages">
			for _, m := range p.msgs.Get() {
				<span ref={p.rows} key={m.id} class={rowClass(m)} onMouseEnter={p.d.logged("enter " + m.subject)} onMouseLeave={p.d.logged("leave " + m.subject)}>{rowLabel(m)}</span>
			}
		</div>
		<div class="flex-col">
			<span class="font-bold">Rows are keyed by message id.</span>
			<span class="font-dim">Hold the pointer still and press s to</span>
			<span class="font-dim">shuffle: the row that leaves gets its</span>
			<span class="font-dim">leave, the one now under the pointer</span>
			<span class="font-dim">gets its enter, right away.</span>
			<span class="font-dim mt-1">t toggles a new message every 2s.</span>
			<span class="font-dim">Click a row to toggle unread. Unread</span>
			<span class="font-dim">rows hover magenta (class expression).</span>
			<span class="text-cyan mt-1">{p.liveLabel()}</span>
		</div>
	</div>
}
