package tui

import (
	"strings"
	"testing"
	"time"
)

// newTestInput creates an Input bound to the shared test app.
func newTestInput(opts ...InputOption) *Input {
	inp := NewInput(opts...)
	inp.BindApp(testApp)
	return inp
}

func TestInput_NewInput_Defaults(t *testing.T) {
	inp := newTestInput()

	if inp.width != 20 {
		t.Errorf("width = %d, want 20", inp.width)
	}
	if inp.border != BorderNone {
		t.Errorf("border = %v, want BorderNone", inp.border)
	}
	if inp.cursorRune != '▌' {
		t.Errorf("cursorRune = %q, want '▌'", inp.cursorRune)
	}
	if inp.placeholderStyle != (Style{}.Dim()) {
		t.Errorf("placeholderStyle = %+v, want dim", inp.placeholderStyle)
	}
	if got := inp.Text(); got != "" {
		t.Errorf("Text() = %q, want empty", got)
	}
	if inp.cursorPos.Get() != 0 || inp.scrollPos.Get() != 0 {
		t.Errorf("cursorPos/scrollPos = %d/%d, want 0/0", inp.cursorPos.Get(), inp.scrollPos.Get())
	}
	if !inp.blink.Get() {
		t.Error("blink should start true")
	}
	if inp.IsFocused() {
		t.Error("input should start unfocused")
	}
}

func TestInput_BindApp_BindsAllStates(t *testing.T) {
	type tc struct {
		stateApp func(*Input) *App
	}

	tests := map[string]tc{
		"text":      {stateApp: func(i *Input) *App { return i.text.app }},
		"cursorPos": {stateApp: func(i *Input) *App { return i.cursorPos.app }},
		"scrollPos": {stateApp: func(i *Input) *App { return i.scrollPos.app }},
		"blink":     {stateApp: func(i *Input) *App { return i.blink.app }},
		"focused":   {stateApp: func(i *Input) *App { return i.focused.app }},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inp := NewInput()
			inp.BindApp(testApp)
			if tt.stateApp(inp) != testApp {
				t.Errorf("state %s not bound to testApp", name)
			}
		})
	}
}

func TestInput_SetTextAndClear(t *testing.T) {
	type tc struct {
		setText       string
		clear         bool
		wantText      string
		wantCursorPos int
	}

	tests := map[string]tc{
		"ascii moves cursor to end": {
			setText: "hello", wantText: "hello", wantCursorPos: 5,
		},
		"multibyte counts runes not bytes": {
			setText: "a界🙂", wantText: "a界🙂", wantCursorPos: 3,
		},
		"clear resets text cursor and scroll": {
			setText: "hello world long", clear: true, wantText: "", wantCursorPos: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inp := newTestInput(WithInputWidth(5))
			inp.SetText(tt.setText)
			if tt.clear {
				inp.scrollPos.Set(3)
				inp.Clear()
				if inp.scrollPos.Get() != 0 {
					t.Errorf("scrollPos after Clear = %d, want 0", inp.scrollPos.Get())
				}
			}
			if got := inp.Text(); got != tt.wantText {
				t.Errorf("Text() = %q, want %q", got, tt.wantText)
			}
			if got := inp.cursorPos.Get(); got != tt.wantCursorPos {
				t.Errorf("cursorPos = %d, want %d", got, tt.wantCursorPos)
			}
		})
	}
}

func TestInput_VisibleWidth(t *testing.T) {
	type tc struct {
		width  int
		border BorderStyle
		want   int
	}

	tests := map[string]tc{
		"no border uses full width":     {width: 20, border: BorderNone, want: 20},
		"border reserves two columns":   {width: 20, border: BorderSingle, want: 18},
		"rounded border also shrinks":   {width: 10, border: BorderRounded, want: 8},
		"zero width without border":     {width: 0, border: BorderNone, want: 0},
		"narrow border can go negative": {width: 1, border: BorderSingle, want: -1},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inp := newTestInput(WithInputWidth(tt.width), WithInputBorder(tt.border))
			if got := inp.visibleWidth(); got != tt.want {
				t.Errorf("visibleWidth() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestInput_ClampCursorPos(t *testing.T) {
	type tc struct {
		text string
		pos  int
		want int
	}

	tests := map[string]tc{
		"negative clamps to zero":    {text: "abc", pos: -1, want: 0},
		"past end clamps to length":  {text: "abc", pos: 99, want: 3},
		"in range passes through":    {text: "abc", pos: 1, want: 1},
		"multibyte clamps to runes":  {text: "a界", pos: 10, want: 2},
		"empty text clamps to zero":  {text: "", pos: 5, want: 0},
		"at exact end stays at end":  {text: "ab", pos: 2, want: 2},
		"at exact start stays there": {text: "ab", pos: 0, want: 0},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inp := newTestInput()
			inp.text.Set(tt.text)
			inp.cursorPos.Set(tt.pos)
			if got := inp.clampCursorPos(); got != tt.want {
				t.Errorf("clampCursorPos() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestInput_EnsureCursorVisible(t *testing.T) {
	type tc struct {
		width      int
		text       string
		cursorPos  int
		scrollPos  int
		wantScroll int
	}

	tests := map[string]tc{
		"cursor left of window scrolls back": {
			width: 5, text: "abcdefgh", cursorPos: 2, scrollPos: 5, wantScroll: 2,
		},
		"cursor right of window scrolls forward": {
			width: 5, text: "abcdefgh", cursorPos: 8, scrollPos: 0, wantScroll: 4,
		},
		"cursor inside window leaves scroll alone": {
			width: 5, text: "abcdefgh", cursorPos: 3, scrollPos: 2, wantScroll: 2,
		},
		"zero visible width returns early": {
			width: 0, text: "abcdefgh", cursorPos: 8, scrollPos: 3, wantScroll: 3,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inp := newTestInput(WithInputWidth(tt.width))
			inp.text.Set(tt.text)
			inp.cursorPos.Set(tt.cursorPos)
			inp.scrollPos.Set(tt.scrollPos)

			inp.ensureCursorVisible()
			if got := inp.scrollPos.Get(); got != tt.wantScroll {
				t.Errorf("scrollPos = %d, want %d", got, tt.wantScroll)
			}
		})
	}
}

func TestInput_HandleEvent_Editing(t *testing.T) {
	type tc struct {
		width      int
		text       string
		cursorPos  int
		events     []KeyEvent
		wantText   string
		wantCursor int
		wantScroll int
	}

	tests := map[string]tc{
		"typing inserts at cursor": {
			width: 20, text: "", cursorPos: 0,
			events: []KeyEvent{
				{Key: KeyRune, Rune: 'h'},
				{Key: KeyRune, Rune: 'i'},
			},
			wantText: "hi", wantCursor: 2, wantScroll: 0,
		},
		"insert in the middle": {
			width: 20, text: "ac", cursorPos: 1,
			events:   []KeyEvent{{Key: KeyRune, Rune: 'b'}},
			wantText: "abc", wantCursor: 2, wantScroll: 0,
		},
		"insert multibyte rune": {
			width: 20, text: "a界", cursorPos: 1,
			events:   []KeyEvent{{Key: KeyRune, Rune: '🙂'}},
			wantText: "a🙂界", wantCursor: 2, wantScroll: 0,
		},
		"typing past width scrolls window": {
			width: 5, text: "", cursorPos: 0,
			events: []KeyEvent{
				{Key: KeyRune, Rune: 'a'},
				{Key: KeyRune, Rune: 'b'},
				{Key: KeyRune, Rune: 'c'},
				{Key: KeyRune, Rune: 'd'},
				{Key: KeyRune, Rune: 'e'},
				{Key: KeyRune, Rune: 'f'},
				{Key: KeyRune, Rune: 'g'},
				{Key: KeyRune, Rune: 'h'},
			},
			wantText: "abcdefgh", wantCursor: 8, wantScroll: 4,
		},
		"backspace deletes before cursor": {
			width: 20, text: "abc", cursorPos: 2,
			events:   []KeyEvent{{Key: KeyBackspace}},
			wantText: "ac", wantCursor: 1, wantScroll: 0,
		},
		"backspace at start is a no-op": {
			width: 20, text: "abc", cursorPos: 0,
			events:   []KeyEvent{{Key: KeyBackspace}},
			wantText: "abc", wantCursor: 0, wantScroll: 0,
		},
		"backspace pins cursor to right edge on long text": {
			// 8 runes, width 5: deleting at the end keeps the window pinned
			// so each delete scrolls one more character into view.
			width: 5, text: "abcdefgh", cursorPos: 8,
			events:   []KeyEvent{{Key: KeyBackspace}},
			wantText: "abcdefg", wantCursor: 7, wantScroll: 3,
		},
		"backspace on short text resets scroll": {
			width: 5, text: "abc", cursorPos: 3,
			events:   []KeyEvent{{Key: KeyBackspace}},
			wantText: "ab", wantCursor: 2, wantScroll: 0,
		},
		"delete removes char at cursor": {
			width: 20, text: "abc", cursorPos: 1,
			events:   []KeyEvent{{Key: KeyDelete}},
			wantText: "ac", wantCursor: 1, wantScroll: 0,
		},
		"delete at end is a no-op": {
			width: 20, text: "abc", cursorPos: 3,
			events:   []KeyEvent{{Key: KeyDelete}},
			wantText: "abc", wantCursor: 3, wantScroll: 0,
		},
		"delete past visible width keeps cursor in view": {
			width: 5, text: "abcdefgh", cursorPos: 5,
			events:   []KeyEvent{{Key: KeyDelete}},
			wantText: "abcdegh", wantCursor: 5, wantScroll: 1,
		},
		"delete on short text resets scroll": {
			width: 5, text: "abc", cursorPos: 0,
			events:   []KeyEvent{{Key: KeyDelete}},
			wantText: "bc", wantCursor: 0, wantScroll: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inp := newTestInput(WithInputWidth(tt.width))
			inp.text.Set(tt.text)
			inp.cursorPos.Set(tt.cursorPos)

			for _, ev := range tt.events {
				if handled := inp.HandleEvent(ev); !handled {
					t.Fatalf("HandleEvent(%+v) = false, want true", ev)
				}
			}
			if got := inp.Text(); got != tt.wantText {
				t.Errorf("Text() = %q, want %q", got, tt.wantText)
			}
			if got := inp.cursorPos.Get(); got != tt.wantCursor {
				t.Errorf("cursorPos = %d, want %d", got, tt.wantCursor)
			}
			if got := inp.scrollPos.Get(); got != tt.wantScroll {
				t.Errorf("scrollPos = %d, want %d", got, tt.wantScroll)
			}
		})
	}
}

func TestInput_HandleEvent_CursorMovement(t *testing.T) {
	type tc struct {
		text       string
		cursorPos  int
		scrollPos  int
		event      KeyEvent
		wantCursor int
		wantScroll int
	}

	tests := map[string]tc{
		"left moves cursor back": {
			text: "abc", cursorPos: 2,
			event:      KeyEvent{Key: KeyLeft},
			wantCursor: 1, wantScroll: 0,
		},
		"left at start is a no-op": {
			text: "abc", cursorPos: 0,
			event:      KeyEvent{Key: KeyLeft},
			wantCursor: 0, wantScroll: 0,
		},
		"left scrolls window when cursor exits": {
			text: "abcdefgh", cursorPos: 4, scrollPos: 4,
			event:      KeyEvent{Key: KeyLeft},
			wantCursor: 3, wantScroll: 3,
		},
		"right moves cursor forward": {
			text: "abc", cursorPos: 1,
			event:      KeyEvent{Key: KeyRight},
			wantCursor: 2, wantScroll: 0,
		},
		"right at end is a no-op": {
			text: "abc", cursorPos: 3,
			event:      KeyEvent{Key: KeyRight},
			wantCursor: 3, wantScroll: 0,
		},
		"home jumps to start and resets scroll": {
			text: "abcdefgh", cursorPos: 8, scrollPos: 4,
			event:      KeyEvent{Key: KeyHome},
			wantCursor: 0, wantScroll: 0,
		},
		"end jumps past last rune and scrolls": {
			text: "abcdefgh", cursorPos: 0, scrollPos: 0,
			event:      KeyEvent{Key: KeyEnd},
			wantCursor: 8, wantScroll: 4,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inp := newTestInput(WithInputWidth(5))
			inp.text.Set(tt.text)
			inp.cursorPos.Set(tt.cursorPos)
			inp.scrollPos.Set(tt.scrollPos)
			inp.blink.Set(false)

			inp.HandleEvent(tt.event)
			if got := inp.cursorPos.Get(); got != tt.wantCursor {
				t.Errorf("cursorPos = %d, want %d", got, tt.wantCursor)
			}
			if got := inp.scrollPos.Get(); got != tt.wantScroll {
				t.Errorf("scrollPos = %d, want %d", got, tt.wantScroll)
			}
			// Successful moves reset the blink so the cursor is visible.
			moved := inp.cursorPos.Get() != tt.cursorPos ||
				tt.event.Key == KeyHome || tt.event.Key == KeyEnd
			if moved && !inp.blink.Get() {
				t.Error("blink should be reset to true after cursor movement")
			}
		})
	}
}

func TestInput_HandleEvent_UnmatchedEvents(t *testing.T) {
	type tc struct {
		event Event
	}

	tests := map[string]tc{
		"mouse events are not handled": {
			event: MouseEvent{Button: MouseLeft, Action: MousePress, X: 1, Y: 1},
		},
		"unbound special key falls through": {
			event: KeyEvent{Key: KeyUp},
		},
		"ctrl-modified key falls through": {
			event: KeyEvent{Key: KeyTab},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inp := newTestInput()
			inp.SetText("abc")

			if handled := inp.HandleEvent(tt.event); handled {
				t.Errorf("HandleEvent(%+v) = true, want false", tt.event)
			}
			if got := inp.Text(); got != "abc" {
				t.Errorf("text changed to %q, want %q", got, "abc")
			}
		})
	}
}

func TestInput_Submit(t *testing.T) {
	type tc struct {
		text       string
		setOnSub   bool
		wantCalled bool
	}

	tests := map[string]tc{
		"enter calls onSubmit with text":  {text: "hello", setOnSub: true, wantCalled: true},
		"enter without onSubmit is no-op": {text: "hello", setOnSub: false, wantCalled: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var got string
			called := false
			opts := []InputOption{}
			if tt.setOnSub {
				opts = append(opts, WithInputOnSubmit(func(s string) {
					called = true
					got = s
				}))
			}
			inp := newTestInput(opts...)
			inp.SetText(tt.text)

			if handled := inp.HandleEvent(KeyEvent{Key: KeyEnter}); !handled {
				t.Fatal("HandleEvent(Enter) = false, want true")
			}
			if called != tt.wantCalled {
				t.Fatalf("onSubmit called = %v, want %v", called, tt.wantCalled)
			}
			if tt.wantCalled && got != tt.text {
				t.Errorf("onSubmit received %q, want %q", got, tt.text)
			}
		})
	}
}

func TestInput_OnChange_FiresForEdits(t *testing.T) {
	type tc struct {
		text      string
		cursorPos int
		event     KeyEvent
		wantValue string
		wantCalls int
	}

	tests := map[string]tc{
		"insert fires onChange":           {text: "ab", cursorPos: 2, event: KeyEvent{Key: KeyRune, Rune: 'c'}, wantValue: "abc", wantCalls: 1},
		"backspace fires onChange":        {text: "abc", cursorPos: 3, event: KeyEvent{Key: KeyBackspace}, wantValue: "ab", wantCalls: 1},
		"delete fires onChange":           {text: "abc", cursorPos: 0, event: KeyEvent{Key: KeyDelete}, wantValue: "bc", wantCalls: 1},
		"movement does not fire onChange": {text: "abc", cursorPos: 1, event: KeyEvent{Key: KeyRight}, wantValue: "", wantCalls: 0},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			calls := 0
			var last string
			inp := newTestInput(WithInputOnChange(func(s string) {
				calls++
				last = s
			}))
			inp.text.Set(tt.text)
			inp.cursorPos.Set(tt.cursorPos)

			inp.HandleEvent(tt.event)
			if calls != tt.wantCalls {
				t.Fatalf("onChange calls = %d, want %d", calls, tt.wantCalls)
			}
			if tt.wantCalls > 0 && last != tt.wantValue {
				t.Errorf("onChange value = %q, want %q", last, tt.wantValue)
			}
		})
	}
}

func TestInput_FocusBlur_Transitions(t *testing.T) {
	inp := newTestInput()

	if !inp.IsFocusable() {
		t.Error("IsFocusable() = false, want true")
	}
	if !inp.IsTabStop() {
		t.Error("IsTabStop() = false, want true")
	}

	// Focus sets focused and resets the blink.
	inp.blink.Set(false)
	inp.Focus()
	if !inp.IsFocused() {
		t.Fatal("IsFocused() = false after Focus()")
	}
	if !inp.blink.Get() {
		t.Error("blink should reset to true on focus")
	}

	// Focus is idempotent: a second call must not reset blink state.
	inp.blink.Set(false)
	inp.Focus()
	if inp.blink.Get() {
		t.Error("second Focus() should be a no-op and leave blink untouched")
	}

	inp.Blur()
	if inp.IsFocused() {
		t.Fatal("IsFocused() = true after Blur()")
	}

	// Blur is idempotent.
	inp.Blur()
	if inp.IsFocused() {
		t.Error("second Blur() should leave input unfocused")
	}
}

func TestInput_EscapeBlursViaApp(t *testing.T) {
	inp := newTestInput()
	root := inp.Render(testApp)

	testApp.focus.Register(root)
	defer testApp.focus.Unregister(root)
	testApp.focus.SetFocus(root)

	if !inp.IsFocused() {
		t.Fatal("input should be focused after SetFocus on its element")
	}

	if handled := inp.HandleEvent(KeyEvent{Key: KeyEscape, app: testApp}); !handled {
		t.Fatal("HandleEvent(Escape) = false, want true")
	}
	if inp.IsFocused() {
		t.Error("input should be blurred after Escape")
	}
	if testApp.Focused() != nil {
		t.Error("app should have no focused element after Escape")
	}
}

func TestInput_EscapeWithoutApp_IsHandled(t *testing.T) {
	inp := newTestInput()
	inp.Focus()

	// No app attached to the event: the handler must not blur the input.
	if handled := inp.HandleEvent(KeyEvent{Key: KeyEscape}); !handled {
		t.Fatal("HandleEvent(Escape) = false, want true")
	}
	if !inp.IsFocused() {
		t.Error("input should remain focused when event has no app")
	}
}

func TestInput_KeyMap_FocusGatedBindings(t *testing.T) {
	inp := newTestInput()
	km := inp.KeyMap()

	if len(km) != 9 {
		t.Fatalf("KeyMap() has %d bindings, want 9", len(km))
	}
	for i, b := range km {
		if !b.Pattern.FocusRequired {
			t.Errorf("binding %d is not focus-gated", i)
		}
		if !b.Stop {
			t.Errorf("binding %d does not stop propagation", i)
		}
	}
}

func TestInput_Watchers_BlinkTimer(t *testing.T) {
	inp := newTestInput()

	ws := inp.Watchers()
	if len(ws) != 1 {
		t.Fatalf("Watchers() returned %d watchers, want 1", len(ws))
	}
	tw, ok := ws[0].(*timerWatcher)
	if !ok {
		t.Fatalf("watcher is %T, want *timerWatcher", ws[0])
	}
	if tw.interval != 500*time.Millisecond {
		t.Errorf("blink interval = %v, want 500ms", tw.interval)
	}

	// Focused: each tick toggles the blink state.
	inp.Focus()
	if !inp.blink.Get() {
		t.Fatal("blink should be true after focus")
	}
	tw.handler()
	if inp.blink.Get() {
		t.Error("blink should toggle to false on tick")
	}
	tw.handler()
	if !inp.blink.Get() {
		t.Error("blink should toggle back to true on tick")
	}

	// Unfocused: ticks leave blink alone.
	inp.Blur()
	inp.blink.Set(true)
	tw.handler()
	if !inp.blink.Get() {
		t.Error("blink should not toggle while unfocused")
	}
}

func TestInput_DisplayText(t *testing.T) {
	type tc struct {
		width     int
		text      string
		cursorPos int
		scrollPos int
		focused   bool
		blink     bool
		want      string
	}

	tests := map[string]tc{
		"unfocused empty shows single space": {
			width: 10, text: "", focused: false, want: " ",
		},
		"unfocused short text shows all": {
			width: 10, text: "abc", cursorPos: 3, focused: false, want: "abc",
		},
		"unfocused long text shows scrolled viewport": {
			// cursor at end forces scroll to 4, viewport shows the tail
			width: 5, text: "abcdefgh", cursorPos: 8, focused: false, want: "efgh",
		},
		"focused shows cursor at position": {
			width: 10, text: "abc", cursorPos: 1, focused: true, blink: true, want: "a▌bc",
		},
		"focused blink off shows space cursor": {
			width: 10, text: "abc", cursorPos: 1, focused: true, blink: false, want: "a bc",
		},
		"focused cursor at end": {
			width: 10, text: "abc", cursorPos: 3, focused: true, blink: true, want: "abc▌",
		},
		"focused empty shows bare cursor": {
			width: 10, text: "", cursorPos: 0, focused: true, blink: true, want: "▌",
		},
		"focused scrolled viewport keeps cursor visible": {
			width: 5, text: "abcdefgh", cursorPos: 8, focused: true, blink: true, want: "efgh▌",
		},
		"scroll beyond cursor shifts view start": {
			// width 0 disables ensureCursorVisible, exercising the
			// scroll > pos adjustment in displayText directly.
			width: 0, text: "abc", cursorPos: 1, scrollPos: 2, focused: true, blink: true, want: "c",
		},
		"negative scroll clamps to zero": {
			width: 0, text: "ab", cursorPos: 0, scrollPos: -3, focused: true, blink: true, want: "▌",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inp := newTestInput(WithInputWidth(tt.width))
			inp.text.Set(tt.text)
			inp.cursorPos.Set(tt.cursorPos)
			inp.scrollPos.Set(tt.scrollPos)
			inp.focused.Set(tt.focused)
			inp.blink.Set(tt.blink)

			if got := inp.displayText(); got != tt.want {
				t.Errorf("displayText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInput_Render_ContentAndPlaceholder(t *testing.T) {
	type tc struct {
		opts      []InputOption
		text      string
		focused   bool
		wantChild string
		wantStyle Style
	}

	tests := map[string]tc{
		"placeholder shown when empty and unfocused": {
			opts:      []InputOption{WithInputPlaceholder("type here")},
			wantChild: "type here",
			wantStyle: Style{}.Dim(),
		},
		"placeholder hidden when focused": {
			opts:      []InputOption{WithInputPlaceholder("type here")},
			focused:   true,
			wantChild: "▌",
		},
		"text shown instead of placeholder": {
			opts:      []InputOption{WithInputPlaceholder("type here")},
			text:      "hi",
			wantChild: "hi",
		},
		"empty without placeholder shows blank": {
			wantChild: " ",
		},
		"custom text style applied to content": {
			opts:      []InputOption{WithInputTextStyle(NewStyle().Bold())},
			text:      "hi",
			wantChild: "hi",
			wantStyle: NewStyle().Bold(),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inp := newTestInput(tt.opts...)
			if tt.text != "" {
				inp.SetText(tt.text)
			}
			inp.focused.Set(tt.focused)

			root := inp.Render(testApp)
			if len(root.children) != 1 {
				t.Fatalf("root has %d children, want 1", len(root.children))
			}
			child := root.children[0]
			if child.text != tt.wantChild {
				t.Errorf("child text = %q, want %q", child.text, tt.wantChild)
			}
			if tt.wantStyle != (Style{}) && child.textStyle != tt.wantStyle {
				t.Errorf("child textStyle = %+v, want %+v", child.textStyle, tt.wantStyle)
			}
		})
	}
}

func TestInput_Render_BorderStates(t *testing.T) {
	focusColor := Cyan
	borderGrad := NewGradient(Red, Blue)
	focusGrad := NewGradient(Green, Yellow)

	type tc struct {
		opts            []InputOption
		focused         bool
		wantBorder      BorderStyle
		wantBorderStyle Style
		wantGradient    *Gradient
		wantWidth       Value
	}

	tests := map[string]tc{
		"no border by default": {
			wantBorder: BorderNone,
			wantWidth:  Fixed(20),
		},
		"border applied when set": {
			opts:       []InputOption{WithInputBorder(BorderSingle)},
			wantBorder: BorderSingle,
			wantWidth:  Fixed(20),
		},
		"focus color applied when focused": {
			opts: []InputOption{
				WithInputBorder(BorderSingle),
				WithInputFocusColor(focusColor),
			},
			focused:         true,
			wantBorder:      BorderSingle,
			wantBorderStyle: NewStyle().Foreground(focusColor),
			wantWidth:       Fixed(20),
		},
		"focus color ignored when unfocused": {
			opts: []InputOption{
				WithInputBorder(BorderSingle),
				WithInputFocusColor(focusColor),
			},
			wantBorder: BorderSingle,
			wantWidth:  Fixed(20),
		},
		"focus gradient wins over focus color": {
			opts: []InputOption{
				WithInputBorder(BorderSingle),
				WithInputFocusColor(focusColor),
				WithInputFocusGradient(focusGrad),
			},
			focused:      true,
			wantBorder:   BorderSingle,
			wantGradient: &focusGrad,
			wantWidth:    Fixed(20),
		},
		"border gradient applied when unfocused": {
			opts: []InputOption{
				WithInputBorder(BorderSingle),
				WithInputBorderGradient(borderGrad),
			},
			wantBorder:   BorderSingle,
			wantGradient: &borderGrad,
			wantWidth:    Fixed(20),
		},
		"zero width skips width option": {
			opts:       []InputOption{WithInputWidth(0)},
			wantBorder: BorderNone,
			wantWidth:  Auto(),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inp := newTestInput(tt.opts...)
			inp.focused.Set(tt.focused)

			root := inp.Render(testApp)
			if root.border != tt.wantBorder {
				t.Errorf("border = %v, want %v", root.border, tt.wantBorder)
			}
			if root.borderStyle != tt.wantBorderStyle {
				t.Errorf("borderStyle = %+v, want %+v", root.borderStyle, tt.wantBorderStyle)
			}
			if tt.wantGradient == nil {
				if root.borderGradient != nil {
					t.Errorf("borderGradient = %+v, want nil", root.borderGradient)
				}
			} else if root.borderGradient == nil || *root.borderGradient != *tt.wantGradient {
				t.Errorf("borderGradient = %+v, want %+v", root.borderGradient, *tt.wantGradient)
			}
			if !root.IsFocusable() {
				t.Error("rendered root should be focusable")
			}
			if root.style.Width != tt.wantWidth {
				t.Errorf("style.Width = %+v, want %+v", root.style.Width, tt.wantWidth)
			}
		})
	}
}

func TestInput_Render_ElementFocusWiring(t *testing.T) {
	inp := newTestInput()
	root := inp.Render(testApp)

	root.Focus()
	if !inp.IsFocused() {
		t.Fatal("element Focus() should focus the input component")
	}
	root.Blur()
	if inp.IsFocused() {
		t.Fatal("element Blur() should blur the input component")
	}
}

func TestInput_Render_Output(t *testing.T) {
	type tc struct {
		opts       []InputOption
		text       string
		focus      bool
		width      int
		wantRows   int
		wantSubstr string
	}

	tests := map[string]tc{
		"borderless input renders text on one row": {
			opts:       []InputOption{WithInputWidth(10)},
			text:       "hello",
			width:      10,
			wantRows:   1,
			wantSubstr: "hello",
		},
		"bordered input renders three rows": {
			opts:       []InputOption{WithInputWidth(10), WithInputBorder(BorderSingle)},
			text:       "hi",
			width:      10,
			wantRows:   3,
			wantSubstr: "hi",
		},
		"focused input renders cursor": {
			opts:       []InputOption{WithInputWidth(10)},
			text:       "ab",
			focus:      true,
			width:      10,
			wantRows:   1,
			wantSubstr: "ab▌",
		},
		"placeholder rendered when empty": {
			opts:       []InputOption{WithInputWidth(12), WithInputPlaceholder("search")},
			width:      12,
			wantRows:   1,
			wantSubstr: "search",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inp := newTestInput(tt.opts...)
			if tt.text != "" {
				inp.SetText(tt.text)
			}
			if tt.focus {
				inp.Focus()
				inp.blink.Set(true)
			}

			rows := renderedRows(t, inp, tt.width)
			if len(rows) != tt.wantRows {
				t.Fatalf("rendered %d rows, want %d:\n%s", len(rows), tt.wantRows, strings.Join(rows, "\n"))
			}
			if !strings.Contains(strings.Join(rows, "\n"), tt.wantSubstr) {
				t.Errorf("rendered output missing %q:\n%s", tt.wantSubstr, strings.Join(rows, "\n"))
			}
		})
	}
}
