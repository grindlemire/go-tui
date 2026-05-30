// Package main demonstrates the <markdown> element and the Markdown component.
//
// To build and run:
//
//	go run ../../cmd/tui generate markdown.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate markdown.gsx

// sampleDoc exercises every markdown construct the renderer supports, plus a few
// edge cases. It lives here as plain Go (a double-quoted string can hold the
// backticks that code fences and inline code need) and is referenced by the
// generated component constructor.
const sampleDoc = "# go-tui Markdown (ATX h1)\n" +
	"\n" +
	"## Inline styles (ATX h2)\n" +
	"\n" +
	"This paragraph mixes **bold**, *italic*, ***bold italic***, `inline code`, " +
	"and a [link to go.dev](https://go.dev) on terminals that support OSC 8. " +
	"Underscores work too: __bold__ and _italic_.\n" +
	"\n" +
	"### Edge cases (ATX h3)\n" +
	"\n" +
	"Delimiters without a closer stay literal: \"see **docs\" does not turn bold, " +
	"and math like 3 * 4 keeps its asterisk. This long sentence also shows that " +
	"ordinary paragraph text wraps to the configured width without manual breaks.\n" +
	"\n" +
	"Setext Level 1\n" +
	"==============\n" +
	"\n" +
	"Setext Level 2\n" +
	"--------------\n" +
	"\n" +
	"## Code block\n" +
	"\n" +
	"Inline `fmt.Println` first, then a fenced block (blank lines preserved):\n" +
	"\n" +
	"```go\n" +
	"func main() {\n" +
	"    fmt.Println(\"hello\")\n" +
	"\n" +
	"    fmt.Println(\"the blank line above is kept\")\n" +
	"}\n" +
	"```\n" +
	"\n" +
	"## Table\n" +
	"\n" +
	"| Language | Typed | Notes         |\n" +
	"| -------- | ----- | ------------- |\n" +
	"| Go       | yes   | `gofmt` clean |\n" +
	"| Python   | no    | *dynamic*     |\n" +
	"\n" +
	"## Lists\n" +
	"\n" +
	"- first item\n" +
	"- second item with a deliberately long line so it wraps onto more than one row at the fixed render width\n" +
	"  - nested item one\n" +
	"  - nested item two\n" +
	"- third item\n" +
	"\n" +
	"Ordered:\n" +
	"\n" +
	"1. one\n" +
	"2. two\n" +
	"3. three\n" +
	"\n" +
	"The `*` and `+` markers also start unordered lists:\n" +
	"\n" +
	"* star item\n" +
	"+ plus item\n" +
	"\n" +
	"## Blockquotes\n" +
	"\n" +
	"> A simple quote, long enough to show that quoted text wraps inside the bar column at a fixed width.\n" +
	">\n" +
	"> > A nested quote inside a quote.\n" +
	"\n" +
	"> A quote that contains a list:\n" +
	">\n" +
	"> - quoted item one\n" +
	"> - quoted item two\n"

func main() {
	app, err := tui.NewApp(
		tui.WithRootComponent(Viewer()),
		// Mouse reporting is left off so the terminal keeps native text
		// selection and OSC 8 link clicking. Full-screen mode then enables
		// alternate-scroll, so the mouse wheel still scrolls (the terminal
		// sends arrow keys, which the keymap handles).
		tui.WithoutMouse(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
