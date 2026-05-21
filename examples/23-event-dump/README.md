# Event Dump

A diagnostic that logs every key and mouse event the app receives, with a status line showing scroll state. Built as the manual verification step for the Windows reader rewrite, useful as a general input-handling smoke test on any new terminal or platform.

See the full [Event Dump guide](https://www.go-tui.dev/guide/event-dump) for what to watch for.

## Run

```bash
go run .
```

Press any key to log it. Click, drag, or scroll the mouse. Drag the terminal corner to verify resize reflow. Esc or q to quit.
