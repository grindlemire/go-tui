---
title: What Is the Framework?
description: Learn what go-tui is and how it fits together.
---

# What Is the Framework?

## What you'll learn
- What `go-tui` is for.
- The core pieces you build with.
- How `.gsx` and Go code work together.

## Prerequisites
- Basic Go knowledge.
- A terminal and a text editor.

## Steps
1. Build UI structure in `.gsx` files.
2. Use the `tui` command to generate Go code from `.gsx`.
3. Run your app as a normal Go program.

## Example
```text
app.gsx -> tui generate app.gsx -> app_gsx.go -> go run .
```

```go
app, err := tui.NewApp()
if err != nil {
    panic(err)
}
defer app.Close()

app.SetRootComponent(Home())
_ = app.Run()
```

## Recap
`go-tui` gives you a simple workflow: declare UI, generate code, run app.

## Next
[Why it exists](./why-it-exists)
