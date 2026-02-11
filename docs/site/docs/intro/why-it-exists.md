---
title: Why It Exists
description: Understand the problem go-tui solves.
---

# Why It Exists

## What you'll learn
- The main pain points in terminal UI development.
- What `go-tui` simplifies.
- Why the framework uses a compile step for `.gsx`.

## Prerequisites
- Read [What Is the Framework](./what-is-the-framework).

## Steps
1. Terminal UIs are usually hard to structure as they grow.
2. `go-tui` keeps UI declarative and logic in Go.
3. The `tui` tools (`generate`, `check`, `fmt`, `lsp`) give one clear workflow.

## Example
Without structure, many terminal apps become hard to test and hard to change.

With `go-tui`, you split work cleanly:
- `.gsx` for layout and component markup.
- Go for state, handlers, and behavior.
- `tui` CLI for generation, validation, and formatting.

## Recap
The framework exists to make terminal UI code easier to read, test, and maintain.

## Next
[Install and run](../getting-started/install-and-run)
