---
title: Install and Run
description: Set up go-tui and run your first example.
---

# Install and Run

## What you'll learn
- What you need before starting.
- How to install the `tui` tool.
- How to run a working example.

## Prerequisites
- Go `1.25.1` or newer.
- Git.

## Steps
1. Clone the repository.
2. Install the CLI tool.
3. Generate code from a `.gsx` file.
4. Run the example app.

## Example
```bash
# from repo root
go install ./cmd/tui

# verify install
tui version

# run the hello example
cd examples/00-hello
tui generate hello.gsx
go run .
```

Press `q` or `Esc` to quit.

## Recap
You now have the CLI installed and have run a generated `go-tui` app.

## Next
[Create your first project](./first-project)
