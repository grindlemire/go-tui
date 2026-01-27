// Package main demonstrates using named element refs (#Name) in .tui files.
//
// This example shows how to use the named refs feature to access elements
// imperatively from Go code:
//
//   - Simple refs (#Header, #Content, #StatusBar): Direct element access
//   - Loop refs (#Items): Slice of elements for items created in @for loops
//   - Keyed loop refs (#Users key={user.ID}): Map access by key for stable correlation
//   - Conditional refs (#Warning): May be nil if the @if condition is false
//
// To build and run:
//
//	cd examples/refs-demo
//	go run ../../cmd/tui generate refs.tui
//	go run .
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/grindlemire/go-tui/pkg/layout"
	"github.com/grindlemire/go-tui/pkg/tui"
	"github.com/grindlemire/go-tui/pkg/tui/element"
)

//go:generate go run ../../cmd/tui generate refs.tui

// User represents a user for the keyed refs demo.
type User struct {
	ID   string
	Name string
}

func main() {
	app, err := tui.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	// State for RefsDemo
	items := generateItems(50)
	showWarning := false
	selectedIdx := 0

	// State for KeyedRefsDemo
	users := []User{
		{ID: "1", Name: "Alice"},
		{ID: "2", Name: "Bob"},
		{ID: "3", Name: "Charlie"},
	}

	// Track which demo is active (0 = RefsDemo, 1 = KeyedRefsDemo)
	activeDemo := 0

	// Build initial UI
	var refsView RefsDemoView
	var keyedView KeyedRefsDemoView

	refsView = buildRefsDemo(app, items, showWarning, selectedIdx)
	keyedView = buildKeyedDemo(app, users)

	app.SetRoot(refsView.Root)
	app.Focus().Register(refsView.Content)

	// Main event loop
	for {
		event, ok := app.PollEvent(50 * time.Millisecond)
		if ok {
			switch e := event.(type) {
			case tui.KeyEvent:
				switch {
				case e.Key == tui.KeyEscape || e.Rune == 'q':
					return

				// Switch between demos
				case e.Rune == 'd':
					activeDemo = (activeDemo + 1) % 2
					if activeDemo == 0 {
						app.SetRoot(refsView.Root)
						app.Focus().Register(refsView.Content)
					} else {
						app.SetRoot(keyedView.Root)
					}

				default:
					if activeDemo == 0 {
						// RefsDemo controls
						switch {
						// Scroll the Content ref
						case e.Rune == 'j':
							refsView.Content.ScrollBy(0, 1)
						case e.Rune == 'k':
							refsView.Content.ScrollBy(0, -1)
						case e.Rune == 'g':
							refsView.Content.ScrollToTop()
						case e.Rune == 'G':
							refsView.Content.ScrollToBottom()

						// Change selection - demonstrates accessing Items slice ref
						case e.Rune == '+' || e.Rune == '=':
							if selectedIdx < len(items)-1 {
								selectedIdx++
								highlightSelected(refsView.Items, selectedIdx)
							}
						case e.Rune == '-' || e.Rune == '_':
							if selectedIdx > 0 {
								selectedIdx--
								highlightSelected(refsView.Items, selectedIdx)
							}

						// Toggle warning - demonstrates conditional ref
						case e.Key == tui.KeyTab:
							showWarning = !showWarning
							refsView = buildRefsDemo(app, items, showWarning, selectedIdx)
							app.SetRoot(refsView.Root)
							app.Focus().Register(refsView.Content)

							// Demonstrate checking if conditional ref is nil
							if refsView.Warning != nil {
								refsView.Warning.SetBorderStyle(tui.NewStyle().Foreground(tui.Red))
							}

						// Demonstrate modifying the Header ref
						case e.Rune == 'h':
							refsView.Header.SetBorderStyle(tui.NewStyle().Foreground(tui.Green))

						// Demonstrate modifying the StatusBar ref
						case e.Rune == 's':
							refsView.StatusBar.SetBorderStyle(tui.NewStyle().Foreground(tui.Magenta))

						default:
							app.Dispatch(event)
						}
					} else {
						// KeyedRefsDemo controls - access users by key
						switch e.Rune {
						case '1':
							highlightUserByID(keyedView.Users, "1", users)
						case '2':
							highlightUserByID(keyedView.Users, "2", users)
						case '3':
							highlightUserByID(keyedView.Users, "3", users)
						default:
							app.Dispatch(event)
						}
					}
				}

			case tui.ResizeEvent:
				refsView = buildRefsDemo(app, items, showWarning, selectedIdx)
				keyedView = buildKeyedDemo(app, users)
				if activeDemo == 0 {
					app.SetRoot(refsView.Root)
					app.Focus().Register(refsView.Content)
				} else {
					app.SetRoot(keyedView.Root)
				}
			}
		}

		app.Render()
	}
}

// buildRefsDemo creates the RefsDemo UI.
func buildRefsDemo(app *tui.App, items []string, showWarning bool, selectedIdx int) RefsDemoView {
	width, height := app.Size()

	root := element.New(
		element.WithSize(width, height),
		element.WithDirection(layout.Column),
	)

	view := RefsDemo(items, showWarning, selectedIdx)
	root.AddChild(view.Root)

	return RefsDemoView{
		Root:      root,
		Header:    view.Header,
		Content:   view.Content,
		Items:     view.Items,
		Warning:   view.Warning,
		StatusBar: view.StatusBar,
	}
}

// buildKeyedDemo creates the KeyedRefsDemo UI.
func buildKeyedDemo(app *tui.App, users []User) KeyedRefsDemoView {
	width, height := app.Size()

	root := element.New(
		element.WithSize(width, height),
		element.WithDirection(layout.Column),
		element.WithJustify(layout.JustifyCenter),
		element.WithAlign(layout.AlignCenter),
	)

	view := KeyedRefsDemo(users)
	root.AddChild(view.Root)

	return KeyedRefsDemoView{
		Root:  root,
		Users: view.Users,
	}
}

// highlightSelected demonstrates using the Items slice ref to modify
// individual elements created in a @for loop.
func highlightSelected(items []*element.Element, selectedIdx int) {
	for i, item := range items {
		if i == selectedIdx {
			item.SetTextStyle(tui.NewStyle().Bold().Foreground(tui.Cyan))
		} else {
			item.SetTextStyle(tui.NewStyle().Foreground(tui.White))
		}
	}
}

// highlightUserByID demonstrates using keyed refs (map access) to
// highlight a specific user element by their ID.
func highlightUserByID(users map[string]*element.Element, highlightID string, allUsers []User) {
	// Reset all users to normal style
	for _, user := range allUsers {
		if elem, ok := users[user.ID]; ok {
			elem.SetTextStyle(tui.NewStyle().Foreground(tui.White))
		}
	}

	// Highlight the selected user by key
	if elem, ok := users[highlightID]; ok {
		elem.SetTextStyle(tui.NewStyle().Bold().Foreground(tui.Green))
	}
}

// generateItems creates sample items for the list.
func generateItems(count int) []string {
	items := make([]string, count)
	for i := 0; i < count; i++ {
		items[i] = fmt.Sprintf("Item %d - This is a sample item in the scrollable list", i+1)
	}
	return items
}
