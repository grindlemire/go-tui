package example

import (
	"fmt"
	"strings"
	"time"
	tui "github.com/grindlemire/go-tui"
)

// =============================================================================
// Struct-based component with KeyMap and Watchers
// =============================================================================
type complexApp struct {
	// State
	count        *tui.State[int]
	selected     *tui.State[int]
	items        *tui.State[[]string]
	showHeader   *tui.State[bool]
	showFooter   *tui.State[bool]
	timerSeconds *tui.State[int]
	messages     *tui.State[[]string]
	msgCh        chan string

	// Refs for click handling
	incrementBtn *tui.Ref
	decrementBtn *tui.Ref
	resetBtn     *tui.Ref
	itemRefs     *tui.RefList
}

func ComplexApp() *complexApp {
	msgCh := make(chan string, 10)
	app := &complexApp{
		count:        tui.NewState(0),
		selected:     tui.NewState(0),
		items:        tui.NewState([]string{"Item 1", "Item 2", "Item 3"}),
		showHeader:   tui.NewState(true),
		showFooter:   tui.NewState(true),
		timerSeconds: tui.NewState(0),
		messages:     tui.NewState([]string{}),
		msgCh:        msgCh,
		incrementBtn: tui.NewRef(),
		decrementBtn: tui.NewRef(),
		resetBtn:     tui.NewRef(),
		itemRefs:     tui.NewRefList(),
	}

	// Simulate background message producer
	go func() {
		for i := 0; ; i++ {
			time.Sleep(5 * time.Second)
			msgCh <- fmt.Sprintf("Message #%d received", i+1)
		}
	}()

	return app
}

// KeyMap demonstrates the new keyboard handling API
func (c *complexApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		// Quit handlers
		tui.OnKey(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnRuneStop('q', func(ke tui.KeyEvent) { ke.App().Stop() }),

		// Counter controls
		tui.OnRune('+', func(ke tui.KeyEvent) { c.count.Set(c.count.Get() + 1) }),
		tui.OnRune('-', func(ke tui.KeyEvent) { c.count.Set(c.count.Get() - 1) }),
		tui.OnRune('r', func(ke tui.KeyEvent) { c.count.Set(0) }),

		// Navigation
		tui.OnKey(tui.KeyUp, func(ke tui.KeyEvent) {
			if c.selected.Get() > 0 {
				c.selected.Set(c.selected.Get() - 1)
			}
		}),
		tui.OnKey(tui.KeyDown, func(ke tui.KeyEvent) {
			if c.selected.Get() < len(c.items.Get())-1 {
				c.selected.Set(c.selected.Get() + 1)
			}
		}),

		// Toggle visibility
		tui.OnRune('h', func(ke tui.KeyEvent) { c.showHeader.Set(!c.showHeader.Get()) }),
		tui.OnRune('f', func(ke tui.KeyEvent) { c.showFooter.Set(!c.showFooter.Get()) }),

		// Catch-all for other runes
		tui.OnRunes(func(ke tui.KeyEvent) {
			// Handle any other character input
		}),
	}
}

// Watchers demonstrates the new watcher API for timers and channels
func (c *complexApp) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.OnTimer(time.Second, c.tick),
		tui.Watch(c.msgCh, c.addMessage),
	}
}

func (c *complexApp) tick() {
	c.timerSeconds.Set(c.timerSeconds.Get() + 1)
}

func (c *complexApp) addMessage(msg string) {
	current := c.messages.Get()
	ts := time.Now().Format("15:04:05")
	entry := fmt.Sprintf("[%s] %s", ts, msg)
	// Keep last 5 messages
	if len(current) >= 5 {
		current = current[1:]
	}
	c.messages.Set(append(current, entry))
}

// HandleMouse demonstrates the new click handling API with refs
func (c *complexApp) HandleMouse(me tui.MouseEvent) bool {
	return tui.HandleClicks(me,
		tui.Click(c.incrementBtn, func() { c.count.Set(c.count.Get() + 1) }),
		tui.Click(c.decrementBtn, func() { c.count.Set(c.count.Get() - 1) }),
		tui.Click(c.resetBtn, func() { c.count.Set(0) }),
	)
}

templ (c *complexApp) Render() {
	<div class="flex-col gap-2 p-2 border-rounded border-cyan">
		<span class="text-gradient-cyan-magenta font-bold">Complex Component Demo</span>

		// Conditional rendering
		@if c.showHeader.Get() {
			<div class="border-single p-1">
				<span class="font-bold">Header Section</span>
			</div>
		}
		// Counter with refs for click handling
		<div class="flex-col gap-1 border-rounded p-1">
			<span class="text-cyan font-bold">Counter:{fmt.Sprintf("%d", c.count.Get())}</span>
			<div class="flex gap-1">
				<button ref={c.incrementBtn} class="px-2">+</button>
				<button ref={c.decrementBtn} class="px-2">-</button>
				<button ref={c.resetBtn} class="px-2">Reset</button>
			</div>
		</div>

		// List with selection
		<div class="flex-col gap-1 border-rounded p-1">
			<span class="font-bold">Items(↑/↓to navigate)</span>
			@for i, item := range c.items.Get() {
				@if i == c.selected.Get() {
					<div ref={c.itemRefs} class="border-single px-1">
						<span class="text-cyan font-bold">{fmt.Sprintf("> %s", item)}</span>
					</div>
				} @else {
					<span ref={c.itemRefs} class="font-dim">{fmt.Sprintf("  %s", item)}</span>
				}
			}
		</div>

		// Timer from watcher
		<div class="border-rounded p-1">
			<span class="font-dim">Uptime:{fmt.Sprintf("%d seconds", c.timerSeconds.Get())}</span>
		</div>

		// Messages from channel watcher
		<div class="flex-col gap-1 border-rounded p-1">
			<span class="font-bold">Live Messages</span>
			@for _, msg := range c.messages.Get() {
				<span class="text-green">{msg}</span>
			}
			@if len(c.messages.Get()) == 0 {
				<span class="font-dim">Waitingmessages...</span>
			}
		</div>

		// Footer with conditional
		@if c.showFooter.Get() {
			<div class="border-single p-1">
				<span class="font-dim">Footer Section</span>
			</div>
		} @else {
			<span class="font-dim">Footer hidden</span>
		}

		// Help text
		<div class="flex-col gap-1">
			<span class="font-dim">+/-counter|↑/↓navigate|h toggle header|f toggle footer|r reset|q quit</span>
		</div>
	</div>
}

// =============================================================================
// Simple component examples (for syntax testing)
// =============================================================================

// Let bindings
templ LetBindingExample(count int, label string) {
	formattedLabel := fmt.Sprintf("%s:", strings.ToUpper(label))
	@let countText = <span class="font-bold">{fmt.Sprintf("%d", count)}</span>
	<div class="flex-col gap-1 p-1">
		<span>{formattedLabel}</span>
		{countText}
	</div>
}

// Refs - simple, loop, keyed, conditional
templ RefsExample(items []string, users map[string]string, showWarning bool) {
	container := tui.NewRef()
	titleRef := tui.NewRef()
	itemRefs := tui.NewRefList()
	userRefs := tui.NewRefMap[string]()
	warning := tui.NewRef()
	<div ref={container} class="flex-col gap-1">
		<span ref={titleRef} class="font-bold">Dashboard</span>

		@for _, item := range items {
			<span ref={itemRefs}>{item}</span>
		}

		@for id, name := range users {
			<span ref={userRefs}>{name}</span>
		}

		@if showWarning {
			<div ref={warning} class="text-red border-single p-1">
				<span>Warning!</span>
			</div>
		}
	</div>
}

// Numeric values
templ NumericValues() {
	<div>
		<span padding={10}>Integer:{42}</span>
		<span>{3.14159}</span>
		<span>{0xFF}</span>
		<span>{0b1010}</span>
		<span>{0o755}</span>
	</div>
}

// String values
templ StringValues() {
	<div>
		<span>{"Hello, World!"}</span>
		<span>{`Raw string with newlines`}</span>
		<span>{"Escaped: \n\t\""}</span>
	</div>
}

// Self-closing elements
templ SelfClosing() {
	<div>
		<hr />
		<br />
	</div>
}

// Attributes with various types
templ AttributeTypes(enabled bool, size int) {
	<div
		border={tui.BorderDouble}
		padding={2}
		margin={1}
		width="100%"
		height={size}
		class="flex-col items-center">
		<span>Content</span>
	</div>
}

// Helper function
func helperFunction(s string) string {
	return fmt.Sprintf("[%s]", strings.ToUpper(s))
}
