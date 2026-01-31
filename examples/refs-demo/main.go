// Package main demonstrates using element refs (ref={}) in .gsx files.
//
// This example shows how to use the refs feature to access elements
// imperatively from Go code:
//
//   - Simple refs (header, content, statusBar): Direct element access
//   - Loop refs (itemRefs): Slice of elements for items created in @for loops
//   - Keyed loop refs (userRefs): Slice of elements in the keyed demo
//   - Conditional refs (warning): May be nil if the @if condition is false
//
// To build and run:
//
//	cd examples/refs-demo
//	go run ../../cmd/tui generate refs.gsx
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate refs.gsx

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
	refsView := buildRefsDemo(app, items, showWarning, selectedIdx)
	keyedView := buildKeyedDemo(app, users)

	app.SetRoot(refsView.root)

	app.SetGlobalKeyHandler(func(e tui.KeyEvent) bool {
		switch {
		case e.Key == tui.KeyEscape || e.Rune == 'q':
			app.Stop()
			return true

		// Switch between demos
		case e.Rune == 'd':
			activeDemo = (activeDemo + 1) % 2
			if activeDemo == 0 {
				app.SetRoot(refsView.root)
			} else {
				app.SetRoot(keyedView.root)
			}
			return true

		default:
			if activeDemo == 0 {
				return handleRefsDemoKey(app, e, &refsView, &keyedView, items, &showWarning, &selectedIdx, users, &activeDemo)
			}
			return handleKeyedDemoKey(app, e, keyedView, users)
		}
	})

	err = app.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}

type refsDemoState struct {
	root *tui.Element
	view RefsDemoView
}

type keyedDemoState struct {
	root *tui.Element
	view KeyedRefsDemoView
}

func handleRefsDemoKey(
	app *tui.App,
	e tui.KeyEvent,
	refsView *refsDemoState,
	keyedView *keyedDemoState,
	items []string,
	showWarning *bool,
	selectedIdx *int,
	users []User,
	activeDemo *int,
) bool {
	switch {
	// Scroll the Content ref
	case e.Rune == 'j':
		refsView.view.Content.ScrollBy(0, 1)
		return true
	case e.Rune == 'k':
		refsView.view.Content.ScrollBy(0, -1)
		return true
	case e.Rune == 'g':
		refsView.view.Content.ScrollToTop()
		return true
	case e.Rune == 'G':
		refsView.view.Content.ScrollToBottom()
		return true

	// Change selection - demonstrates accessing ItemRefs slice ref
	case e.Rune == '+' || e.Rune == '=':
		if *selectedIdx < len(items)-1 {
			*selectedIdx++
			highlightSelected(refsView.view.ItemRefs, *selectedIdx)
		}
		return true
	case e.Rune == '-' || e.Rune == '_':
		if *selectedIdx > 0 {
			*selectedIdx--
			highlightSelected(refsView.view.ItemRefs, *selectedIdx)
		}
		return true

	// Toggle warning - demonstrates conditional ref
	case e.Key == tui.KeyTab:
		*showWarning = !*showWarning
		*refsView = buildRefsDemo(app, items, *showWarning, *selectedIdx)
		*keyedView = buildKeyedDemo(app, users)
		app.SetRoot(refsView.root)

		// Demonstrate checking if conditional ref is nil
		if refsView.view.Warning != nil {
			refsView.view.Warning.SetBorderStyle(tui.NewStyle().Foreground(tui.Red))
		}
		return true

	// Demonstrate modifying the Header ref
	case e.Rune == 'h':
		refsView.view.Header.SetBorderStyle(tui.NewStyle().Foreground(tui.Green))
		return true

	// Demonstrate modifying the StatusBar ref
	case e.Rune == 's':
		refsView.view.StatusBar.SetBorderStyle(tui.NewStyle().Foreground(tui.Magenta))
		return true
	}

	return false
}

func handleKeyedDemoKey(app *tui.App, e tui.KeyEvent, keyedView keyedDemoState, users []User) bool {
	switch e.Rune {
	case '1':
		highlightUserByIdx(keyedView.view.UserRefs, 0, users)
		return true
	case '2':
		highlightUserByIdx(keyedView.view.UserRefs, 1, users)
		return true
	case '3':
		highlightUserByIdx(keyedView.view.UserRefs, 2, users)
		return true
	}
	return false
}

// buildRefsDemo creates the RefsDemo UI.
func buildRefsDemo(app *tui.App, items []string, showWarning bool, selectedIdx int) refsDemoState {
	width, height := app.Size()

	root := tui.New(
		tui.WithSize(width, height),
		tui.WithDirection(tui.Column),
	)

	view := RefsDemo(items, showWarning, selectedIdx)
	root.AddChild(view.Root)

	return refsDemoState{
		root: root,
		view: view,
	}
}

// buildKeyedDemo creates the KeyedRefsDemo UI.
func buildKeyedDemo(app *tui.App, users []User) keyedDemoState {
	width, height := app.Size()

	root := tui.New(
		tui.WithSize(width, height),
		tui.WithDirection(tui.Column),
		tui.WithJustify(tui.JustifyCenter),
		tui.WithAlign(tui.AlignCenter),
	)

	view := KeyedRefsDemo(users)
	root.AddChild(view.Root)

	return keyedDemoState{
		root: root,
		view: view,
	}
}

// highlightSelected demonstrates using the ItemRefs slice ref to modify
// individual elements created in a @for loop.
func highlightSelected(items []*tui.Element, selectedIdx int) {
	for i, item := range items {
		if i == selectedIdx {
			item.SetTextStyle(tui.NewStyle().Bold().Foreground(tui.Cyan))
		} else {
			item.SetTextStyle(tui.NewStyle().Foreground(tui.White))
		}
	}
}

// highlightUserByIdx demonstrates using loop refs (slice access) to
// highlight a specific user element by index.
func highlightUserByIdx(users []*tui.Element, highlightIdx int, allUsers []User) {
	for i, elem := range users {
		if i == highlightIdx {
			elem.SetTextStyle(tui.NewStyle().Bold().Foreground(tui.Green))
		} else {
			elem.SetTextStyle(tui.NewStyle().Foreground(tui.White))
		}
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
