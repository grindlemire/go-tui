# Hover

Demonstrates hover in two styles:

- `hover:` classes for pure styling (list rows, the card border, the card and modal buttons), with no state involved.
- `onMouseEnter` / `onMouseLeave` callbacks driving state (the tabs), and logging every enter and leave event to the panel at the bottom.

The list scrolls with the mouse wheel; the highlight follows whatever row ends up under the pointer. The card stays lit while the pointer moves between its buttons, and while the modal is open only the modal reacts to hover.

Hover needs pointer motion reporting, which is opt-in:

```go
app, err := tui.NewApp(
    tui.WithMouseMotion(),
    tui.WithRootComponent(HoverDemo()),
)
```

## Run

```bash
go run .
```
