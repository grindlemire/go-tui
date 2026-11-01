# Hover

A five-page tour of hover in go-tui. The tabs switch pages; the panel at the bottom logs every enter and leave event, and the status line counts move and drag events.

- **Basics**: `hover:` classes on a scrollable inbox (the wheel scrolls it), and a card that stays lit while the pointer moves between its buttons.
- **Lists**: a list keyed by message id. With the pointer resting on a row, press `s` to shuffle or `t` to insert a message every two seconds; the row that moved away gets its leave and the new one under the pointer gets its enter right away. Click a row to toggle unread; unread rows hover in a different color from a class expression.
- **Styles**: `hover:text-*` cascading to children, `hover:font-bold` layering onto inherited color, and gradient hover classes.
- **Components**: `<input>` and `<textarea>` with hover classes and callbacks (Tab focuses them so you can type; Escape leaves), and a drag area that reports drags.
- **Edge cases**: a button that dodges the pointer. Its enter handler moves it away, and it gets its leave right after, with the pointer still.

Press `m` for a modal whose backdrop blocks hover and the wheel on everything behind it.

Hover needs no app option: the app turns on pointer motion reporting once an element uses hover. Inline mode leaves the mouse off by default, so the `-inline` run adds `tui.WithMouse()`.

## Run

```bash
go run .          # full screen
go run . -inline  # inline mode: move the pointer above the app to leave everything
```

See the [Hover guide](https://www.go-tui.dev/guide/hover) for a walkthrough.
