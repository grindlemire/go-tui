---
sidebar_position: 1
slug: /
---

# go-tui Documentation

This site is a bold placeholder for documenting `go-tui`.

## What this project is

`go-tui` is a terminal UI toolkit in Go focused on composable elements, event handling, rendering, and testability.

## Quick start

```bash
# from repo root
cd docs/site
bun install
bun run start
```

Then open `http://localhost:3000`.

## Why this style exists

- Yellow is the core brand highlight.
- Light mode uses darker amber for better readability.
- Dark mode keeps bright yellow with glow on interactive states.

:::note Try light and dark mode
Use the theme toggle in the navbar and compare link/button readability.
:::

## Navigation guide

- Start here: this page.
- Deep visual test page: `Theme Showcase`.
- Add new docs in `docs/site/docs/`.

## Example snippet

```go
package main

import "fmt"

func main() {
    fmt.Println("go-tui docs preview")
}
```

## Next documentation sections to add

| Section | Goal |
| --- | --- |
| Architecture | Explain render loop, app lifecycle, and event model |
| API guide | Document packages, options, and defaults |
| Recipes | Practical patterns for focus, layout, scrolling, input |
| Testing | Golden tests, integration strategy, and fixtures |

Continue to `Theme Showcase` for a fuller visual test.
