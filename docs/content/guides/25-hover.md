# Hover

## Overview

Hover lets elements react when the mouse pointer moves over them. go-tui gives you two tools for it. A `hover:` class prefix restyles an element while the pointer is over it, with no state or handler code. The `onMouseEnter` and `onMouseLeave` attributes call your functions when the pointer arrives and departs, so you can update state, log activity, or show a preview.

This guide builds around the hover example in `examples/25-hover`. Its tabs switch between five pages, and a panel at the bottom logs every enter and leave event as it happens:

- **Basics**: a scrollable inbox with `hover:` rows and a card that stays lit around its buttons.
- **Lists**: a keyed list you can shuffle or feed with live inserts while the pointer sits still, with hover colors from a class expression.
- **Styles**: hover text styles cascading to children and layering on inherited styles, plus gradients.
- **Components**: `<input>` and `<textarea>` with hover, and a drag area.
- **Edge cases**: a button whose enter handler moves it out from under the pointer.

## Pointer Motion

Terminals only report pointer movement when an app asks for it, and they report it often. Every cell the pointer crosses produces an event, which adds up over SSH or inside tmux. So go-tui keeps motion reporting off until something needs it.

You don't have to turn it on yourself. After the first render in which any element has `onMouseEnter`, `onMouseLeave`, or a `hover:` class, the app asks the terminal for motion reports, and it keeps them on from then on. Apps that never use hover never pay for them.

Hover needs the mouse enabled. It is on by default in full screen mode. Inline mode leaves it off so the terminal's own text selection and scrollback keep working, so pass `tui.WithMouse()` to use hover there, as the example does with `-inline`:

```go
opts := []tui.AppOption{
    tui.WithRootComponent(HoverDemo(*inline)),
}
if *inline {
    opts = append(opts, tui.WithInlineHeight(24), tui.WithMouse())
}
```

`tui.WithoutMouse()` turns hover off along with the rest of the mouse.

Once motion is on, your `HandleMouse` method also receives the moves themselves as `MouseEvent` values with `Action: tui.MouseMove` and `Button: tui.MouseNone`. The example uses them, and drags, for its status line:

```go
func (d *hoverDemo) HandleMouse(me tui.MouseEvent) bool {
    d.pointer.Set(fmt.Sprintf("%d,%d", me.X, me.Y))
    switch me.Action {
    case tui.MouseMove:
        d.moves.Set(d.moves.Get() + 1)
    case tui.MouseDrag:
        d.drags.Set(d.drags.Get() + 1)
    }
    // ...
}
```

Most apps never look at `MouseMove` directly. The hover attributes and classes cover the common cases. If you do want moves without using hover, pass `tui.WithMouseMotion()` to turn motion on from the start.

## Styling with hover: Classes

Prefix a class with `hover:` and it applies only while the pointer is over the element. The rows in the example's message list highlight this way:

```gsx
for _, row := range inbox {
    <span key={row} class="px-1 hover:bg-blue hover:text-bright-white">{row}</span>
}
```

When the pointer leaves a row, its background and text color go back to what they were. The styling lives entirely in the class attribute.

Hover counts descendants too: an element stays hovered while the pointer is over any of its children. The card in the example lights its border when the pointer is anywhere inside it, including over the buttons:

```gsx
<div class="flex-col px-1 border-rounded border-[#666666] hover:border-magenta">
    <span class="font-bold">Nested card</span>
    ...
</div>
```

### Allowed classes

`hover:` only accepts classes that change color or text style. Anything that changes layout could move the element out from under the pointer, which would end the hover and make it flicker. The allowed classes are:

| Group | Classes |
|-------|---------|
| Text color | `text-COLOR`, `text-bright-COLOR`, `text-[#hex]` |
| Background | `bg-COLOR`, `bg-bright-COLOR`, `bg-[#hex]` |
| Text style | `font-bold`, `font-dim`, `text-dim`, `italic`, `underline`, `strikethrough`, `blink`, `reverse` |
| Border color | `border-COLOR`, `border-[#hex]` |
| Gradients | `text-gradient-*`, `bg-gradient-*`, `border-gradient-*` |
| Scrollbar | `scrollbar-COLOR`, `scrollbar-thumb-COLOR` and their `[#hex]` forms |

Note that `hover:border-cyan` changes the border's color, while the border itself (`border`, `border-rounded`, and the other shape classes) has to be in the element's normal classes. A layout class after `hover:` is a compile error:

```
hover: only supports color and text style classes, not "p-2", which changes layout
```

The same applies to sizes, margins, `hidden`, text alignment, and flex classes.

### How hover styles combine

Hover text styles layer on top of the element's current style, including anything it inherits from its parent, and cascade to children the same way other text styles do. A `hover:font-bold` on a container bolds the text inside it. Hover backgrounds cascade to children as well.

When the pointer leaves, every property the hover classes changed returns to its value from before the hover. If your own code changes one of those properties while the element is hovered (for example by calling `SetBackground`), your value is kept instead.

### Dynamic classes

Expression-valued classes accept `hover:` too, so `class={expr}` can compute hover styles at runtime. The Lists page gives unread messages a different hover color:

```go
func rowClass(m message) string {
    if m.unread {
        return "px-1 font-bold text-bright-white hover:bg-magenta hover:text-black"
    }
    return "px-1 text-white hover:bg-blue"
}
```

```gsx
<span ref={p.rows} key={m.id} class={rowClass(m)} ...>{rowLabel(m)}</span>
```

`tui.WithClass` and `Element.SetClass` handle the prefix the same way. At runtime there is no compiler to report mistakes, so unknown classes and layout classes after `hover:` are ignored.

## Hover Callbacks

When hovering should change state rather than just color, use `onMouseEnter` and `onMouseLeave`. Both take a `func()`.

The page tabs track which one is hovered in a `*State[string]` and derive their classes from it:

```go
func (d *hoverDemo) enterTab(tab string) func() {
    return func() {
        d.record("enter tab " + tab)
        d.hoveredTab.Set(tab)
    }
}

func (d *hoverDemo) leaveTab(tab string) func() {
    return func() {
        d.record("leave tab " + tab)
        d.hoveredTab.Set("")
    }
}

func (d *hoverDemo) tabClass(tab string) string {
    switch {
    case tab == d.page.Get():
        return "px-1 font-bold bg-cyan text-black"
    case tab == d.hoveredTab.Get():
        return "px-1 underline text-bright-cyan"
    }
    return "px-1 text-white"
}
```

```gsx
<div class="flex">
    for _, tab := range tabs {
        <span ref={d.tabRefs} key={tab} class={d.tabClass(tab)} onMouseEnter={d.enterTab(tab)} onMouseLeave={d.leaveTab(tab)}>{tab}</span>
    }
</div>
```

A `hover:underline hover:text-bright-cyan` class would give the same look. The tabs use callbacks to show the state-driven approach, which you need when something other than the hovered element reacts: a status line naming the hovered item, a preview pane, or a tooltip.

Each callback can set state like any other handler. Setting state schedules a render, and the new tree is built with the updated value.

## Nested Elements

Enter and leave follow the same rules as the DOM's `mouseenter` and `mouseleave`:

- Enter fires when the pointer moves onto the element or any of its descendants.
- Leave fires when the pointer is off the element and all of its descendants.
- Moving between children does not leave the parent.

The card in the example has callbacks on the card and on each button inside it. The log indents the button events so the nesting is visible:

```gsx
<div class="flex-col px-1 border-rounded border-[#666666] hover:border-magenta" onMouseEnter={p.d.logged("enter card")} onMouseLeave={p.d.logged("leave card")}>
    ...
    <div class="flex gap-1 mt-1">
        <span class="px-1 border-single hover:border-yellow hover:text-yellow hover:font-bold" onMouseEnter={p.d.logged("  enter button Save")} onMouseLeave={p.d.logged("  leave button Save")}>Save</span>
        <span class="px-1 border-single hover:border-yellow hover:text-yellow hover:font-bold" onMouseEnter={p.d.logged("  enter button Discard")} onMouseLeave={p.d.logged("  leave button Discard")}>Discard</span>
    </div>
</div>
```

Moving the pointer from outside the card straight onto Save logs `enter card` and then `enter button Save`. When one move enters several elements, enters fire outermost first. Moving from Save to Discard logs a leave for Save and an enter for Discard, and the card logs nothing because the pointer never left it. Moving off the card from a button logs the button's leave and then the card's: leaves fire deepest first.

Every enter is followed by exactly one leave. The leave that runs is the `onMouseLeave` the element had when the pointer entered it, even if later renders built the element with a different callback, or removed it from the tree entirely.

## Keys and Lists

go-tui rebuilds elements on every render, so hover cannot remember an element by pointer. It identifies elements by their position in the tree instead. For static layouts that is all you need.

Lists that reorder are different. Without a key, hover sticks to the position: if an item is inserted above the hovered row, the row that slides under the pointer gets no enter, and the row that moved away gets no leave until the pointer leaves that spot. Give list items a `key` and hover tracks the items themselves. The Lists page keys its messages by id:

```gsx
for _, m := range p.msgs.Get() {
    <span ref={p.rows} key={m.id} class={rowClass(m)} onMouseEnter={p.d.logged("enter " + m.subject)} onMouseLeave={p.d.logged("leave " + m.subject)}>{rowLabel(m)}</span>
}
```

Rest the pointer on a row and press `s` to shuffle, or `t` to start inserting a new message at the top every two seconds. The log shows the leave for the message that moved away and the enter for the one that took its place, without the pointer moving.

On a plain element, `key={...}` compiles to `tui.WithKey(key)` in addition to its other uses (the `RefMap` key with `ref`, and component identity). When siblings are inserted, removed, or reordered under a still pointer, the item that moved away gets its leave and the item now under the pointer gets its enter. Keys must be unique among siblings, as with any other `key`.

## Scrolling

Hover works inside scrollable containers. Hit testing accounts for the container's border, its scrollbar, and the current scroll offset, so the row under the pointer is the row that is drawn there. The example's message list is a scrolled container wrapped around the keyed rows:

```gsx
<div
    class="flex-col border-rounded w-32 shrink-0 overflow-y-scroll scrollbar-bright-black"
    height={12}
    borderTitle="Inbox (wheel scrolls)">
    for _, row := range inbox {
        <span key={row} class="px-1 hover:bg-blue hover:text-bright-white" ...>{row}</span>
    }
</div>
```

The list has no `scrollOffset` and no wheel handler: the wheel scrolls it on its own, as described in [Scrolling](scrolling). Scroll the list with the wheel while holding the pointer still, and the highlight moves to whichever row scrolls under it. The leave for the old row and the enter for the new one fire right after the scroll renders, without waiting for the pointer to move. The pointer resting on the border or the scrollbar hovers the container itself and none of its rows.

Clicks work the same way: the Lists page toggles a message's unread marker with `el.ContainsPoint` on a `RefMap` of rows inside its scrolled list.

## Modals

While a modal is open, only the topmost modal's contents receive hover. Content behind the backdrop does not react, and anything that was hovered when the modal opened gets its leave callback.

The example opens a modal with `m`. Its Close button uses a `hover:` class and logs its own events:

```gsx
<modal open={d.showModal} class="justify-center items-center">
    <div class="border-rounded p-2 flex-col gap-1 w-40 items-center">
        <span class="font-bold text-yellow">Modal open</span>
        <span class="font-dim">Content behind the backdrop ignores hover.</span>
        <button class="px-2 border-single hover:border-green hover:text-green hover:font-bold" onMouseEnter={d.logged("enter modal button")} onMouseLeave={d.logged("leave modal button")} onActivate={d.closeModal}>Close</button>
    </div>
</modal>
```

Hover a list row and press `m`. The row's leave fires and its highlight clears as soon as the modal appears, even though the pointer has not moved; the entry is in the log once you close the modal.

## Component Tags

The component tags `<input>`, `<textarea>`, `<markdown>`, and `<modal>` accept `onMouseEnter`, `onMouseLeave`, and `hover:` classes. They apply to the component's root element, the same element a `class` attribute styles:

```gsx
<input
    value={p.name}
    placeholder="hover me"
    width={28}
    border={tui.BorderRounded}
    class="hover:border-cyan"
    onMouseEnter={p.d.logged("enter input")}
    onMouseLeave={p.d.logged("leave input")}
/>
```

The Components page also has a drag area. Drags are reported whenever the mouse is enabled, and hover keeps tracking the pointer while a button is held.

## When Callbacks Fire

Hover is evaluated at two points:

1. **On every mouse event.** Moves, clicks, wheel, and drags all update hover at the pointer's position, before the event reaches any `HandleMouse` method.
2. **After a render that followed input.** When a key press, click, wheel event, or background update causes a render, go-tui checks hover once more at the last known pointer position. This is what makes scrolling, reordering, and opening a modal update hover immediately, while the pointer sits still.

Renders caused only by hover callbacks never trigger that second check. A handler that sets state in `onMouseEnter` therefore cannot start a loop, even if the state change moves things around.

The Edge cases page shows this with a dodge button. Entering it moves the button to the other side, and the re-check after that render gives it its leave, with the pointer still.

Motion arrives in bursts, so `app.Run()` and `app.DispatchEvents()` coalesce it: when several `MouseMove` events are waiting in one drain of the queue, only the latest is dispatched, in its original position relative to other events. Drags are not coalesced. If you call `app.Dispatch` yourself, every event is dispatched as given, and a custom loop that reads `app.Events()` sees every move. The example's move counter shows the coalesced count, which is usually far lower than the number of cells the pointer crossed.

In inline mode, moving the pointer above or below the inline region counts as leaving the app, so everything hovered gets its leave.

## Building Elements in Go

The attributes map to element options, which you can use when building elements in Go:

```go
row := tui.New(
    tui.WithText("Build failed on main"),
    tui.WithKey("build-failed"),
    tui.WithHoverClass("bg-blue text-bright-white"),
    tui.WithOnMouseEnter(func() { status.Set("build-failed") }),
    tui.WithOnMouseLeave(func() { status.Set("") }),
)
```

`WithHoverClass` takes the classes without the `hover:` prefix. `el.SetOnMouseEnter` and `el.SetOnMouseLeave` replace the callbacks on an existing element.

## Limitations

- Terminals do not report the pointer leaving the window. If the pointer exits the terminal while over an element, that element stays hovered until the pointer comes back.
- `hover:` cannot change layout. Use callbacks and state when hovering needs to show or hide content or resize something.
- Once motion is on it stays on, even if you switch to a screen without hover.

## Run

```bash
tui generate .
go run .          # full screen
go run . -inline  # inline mode, 24 rows below the prompt
```

Switch pages with `1` to `5` or by clicking the tabs. Press `m` to open the modal (full screen only), `c` to clear the log, and `q` to quit. In inline mode, move the pointer above the app to see everything get its leave.

## Next Steps

- [Refs and Click Handling](refs-and-clicks) for clicks and the `RefMap` used by the tabs.
- [Scrolling](scrolling) for the scrollable container that holds the message list.
- [Styling](styling) for the color and text style classes that `hover:` accepts.
