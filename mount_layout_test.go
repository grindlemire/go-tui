package tui

import (
	"fmt"
	"strings"
	"testing"
)

// dumpTree prints element rects for debugging layout issues.
func dumpTree(el *Element, indent int) string {
	var sb strings.Builder
	prefix := strings.Repeat("  ", indent)
	rect := el.Rect()
	contentRect := el.ContentRect()
	text := el.Text()
	if text == "" {
		text = "(no text)"
	}
	nChildren := len(el.Children())
	sb.WriteString(fmt.Sprintf("%srect=%v content=%v text=%q children=%d dirty=%v\n",
		prefix, rect, contentRect, text, nChildren, el.IsDirty()))
	for _, child := range el.Children() {
		sb.WriteString(dumpTree(child, indent+1))
	}
	return sb.String()
}

// TestMount_HeaderViewLayout tests that a mounted function-templ View
// has correct layout when used inside a method component's Render.
func TestMount_HeaderViewLayout(t *testing.T) {
	cleanup := setupTestMountState()
	defer cleanup()

	// Simulate what the generated code does for Header:
	//   templ Header(title string) {
	//     <div class="border-rounded p-1 flex justify-center">
	//       <span class="font-bold">{title}</span>
	//     </div>
	//   }
	type headerView struct {
		Root *Element
	}
	makeHeader := func(title string) headerView {
		div := New(
			WithBorder(BorderRounded),
			WithPadding(1),
			WithDisplay(DisplayFlex), WithDirection(Row),
			WithJustify(JustifyCenter),
		)
		span := New(WithText(title))
		div.AddChild(span)
		return headerView{Root: div}
	}

	// Make it implement Component (like the generated Render method)
	type headerComponent struct {
		headerView
	}
	_ = headerComponent{} // silence unused

	// Simulate the method component's Render():
	//   outer container (column, padding 2, gap 2)
	//     mounted Header("Component Showcase")
	//     some other content
	parent := &mockParent{}

	outer := New(
		WithDirection(Column),
		WithPadding(2),
		WithGap(2),
	)

	// Mount the header (simulates app.Mount in generated code)
	headerEl := testApp.Mount(parent, 0, func() Component {
		hv := makeHeader("Component Showcase")
		// Return something implementing Component with Render returning Root
		return &simpleViewComponent{root: hv.Root}
	})
	outer.AddChild(headerEl)

	// Add some other content below
	otherContent := New(WithText("other stuff"))
	outer.AddChild(otherContent)

	// Run layout with a reasonable terminal size
	outer.Calculate(80, 24)

	// Dump tree for debugging
	tree := dumpTree(outer, 0)
	t.Logf("Element tree after Calculate:\n%s", tree)

	// Check header element has reasonable dimensions
	headerRect := headerEl.Rect()
	if headerRect.Width == 0 {
		t.Error("header width is 0")
	}
	if headerRect.Height == 0 {
		t.Error("header height is 0")
	}

	// The header should have children (the span)
	if len(headerEl.Children()) == 0 {
		t.Fatal("header has no children - span is missing!")
	}

	// Check the span child has layout
	span := headerEl.Children()[0]
	spanRect := span.Rect()
	t.Logf("Header rect: %v, Span rect: %v, Span text: %q", headerRect, spanRect, span.Text())

	if spanRect.Height == 0 {
		t.Error("span height is 0 - text won't be visible")
	}
	if spanRect.Width == 0 {
		t.Error("span width is 0 - text won't be visible")
	}

	// The span should be within the header's content area
	headerContent := headerEl.ContentRect()
	if spanRect.Y < headerContent.Y || spanRect.Y >= headerContent.Y+headerContent.Height {
		t.Errorf("span Y=%d is outside header content area Y=[%d, %d)",
			spanRect.Y, headerContent.Y, headerContent.Y+headerContent.Height)
	}

	// Now test the SECOND render (what App.Render does - rebuild containers, reuse mounts)
	// This is the critical path: new outer container, cached mount elements
	outer2 := New(
		WithDirection(Column),
		WithPadding(2),
		WithGap(2),
	)

	headerEl2 := testApp.Mount(parent, 0, func() Component {
		hv := makeHeader("Component Showcase")
		return &simpleViewComponent{root: hv.Root}
	})
	outer2.AddChild(headerEl2)

	otherContent2 := New(WithText("other stuff"))
	outer2.AddChild(otherContent2)

	outer2.Calculate(80, 24)

	tree2 := dumpTree(outer2, 0)
	t.Logf("Element tree after SECOND Calculate (cached mount):\n%s", tree2)

	headerRect2 := headerEl2.Rect()
	if headerRect2.Height == 0 {
		t.Error("SECOND RENDER: header height is 0")
	}
	if len(headerEl2.Children()) == 0 {
		t.Fatal("SECOND RENDER: header has no children")
	}

	span2 := headerEl2.Children()[0]
	spanRect2 := span2.Rect()
	t.Logf("SECOND RENDER: Header rect: %v, Span rect: %v", headerRect2, spanRect2)

	if spanRect2.Height == 0 {
		t.Error("SECOND RENDER: span height is 0")
	}
}

// simpleViewComponent wraps a pre-built element tree as a Component.
type simpleViewComponent struct {
	root *Element
}

func (c *simpleViewComponent) Render(app *App) *Element {
	return c.root
}

// TestMount_FullAppRenderPipeline tests the complete App.Render() flow
// to see if mounted function-templ content appears in the buffer.
func TestMount_FullAppRenderPipeline(t *testing.T) {
	mockTerm := NewMockTerminal(80, 24)

	// Build the app manually (avoiding real terminal)
	app := &App{
		terminal:       mockTerm,
		buffer:         NewBuffer(80, 24),
		focus:          newFocusManager(),
		stopCh:         make(chan struct{}),
		stopped:        false,
		eventQueue:     make(chan func(), 256),
		updateQueue:    make(chan func(), 256),
		mounts:         newMountState(),
		batch:          newBatchContext(),
		needsFullRedraw: true,
		frameDuration:  16,
	}
	app.resetRootSession()

	// Create root component that mounts a "Header" function templ
	root := &mountTestRootComponent{app: app}
	app.SetRootComponent(root)

	// Call App.Render() like the main loop does
	app.Render()

	// Check the buffer for "Hello Mount" text
	bufContent := extractBufferText(app.buffer, 80, 24)
	t.Logf("Buffer content:\n%s", bufContent)

	if !strings.Contains(bufContent, "Hello Mount") {
		t.Error("Buffer does not contain 'Hello Mount' - mounted component text is missing!")
	}

	// Also check that the border is there (just the top-left corner)
	found := false
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			c := app.buffer.Cell(x, y)
			if c.Rune == '╭' {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("Buffer does not contain border corner '╭' - border is missing!")
	}

	// Call App.Render() a second time (cached mounts)
	app.Render()
	bufContent2 := extractBufferText(app.buffer, 80, 24)
	t.Logf("Buffer content after second render:\n%s", bufContent2)
	if !strings.Contains(bufContent2, "Hello Mount") {
		t.Error("SECOND RENDER: Buffer does not contain 'Hello Mount'!")
	}
}

// mountTestRootComponent is a struct component that mounts a function templ.
type mountTestRootComponent struct {
	app *App
}

func (c *mountTestRootComponent) Render(app *App) *Element {
	outer := New(
		WithDirection(Column),
		WithPadding(2),
		WithGap(2),
	)

	// Mount a "header" component (simulates @Header("Hello Mount"))
	headerEl := app.Mount(c, 0, func() Component {
		// Simulates what the generated Header() function templ does
		div := New(
			WithBorder(BorderRounded),
			WithPadding(1),
			WithDisplay(DisplayFlex), WithDirection(Row),
			WithJustify(JustifyCenter),
		)
		span := New(WithText("Hello Mount"))
		div.AddChild(span)
		return &simpleViewComponent{root: div}
	})
	outer.AddChild(headerEl)

	return outer
}

// extractBufferText reads all characters from the buffer.
func extractBufferText(buf *Buffer, width, height int) string {
	var sb strings.Builder
	for y := range height {
		for x := range width {
			c := buf.Cell(x, y)
			sb.WriteRune(c.Rune)
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

// TestMount_ExactGeneratedCodePattern tests with the EXACT element options
// from the generated components_gsx.go to match real-world conditions.
func TestMount_ExactGeneratedCodePattern(t *testing.T) {
	mockTerm := NewMockTerminal(120, 50)

	app := &App{
		terminal:        mockTerm,
		buffer:          NewBuffer(120, 50),
		focus:           newFocusManager(),
		stopCh:          make(chan struct{}),
		stopped:         false,
		eventQueue:      make(chan func(), 256),
		updateQueue:     make(chan func(), 256),
		mounts:          newMountState(),
		batch:           newBatchContext(),
		needsFullRedraw: true,
		frameDuration:   16,
	}
	app.resetRootSession()

	root := &exactGenRootComponent{}
	app.SetRootComponent(root)
	app.Render()

	bufContent := extractBufferText(app.buffer, 120, 50)

	// Check Header text
	if !strings.Contains(bufContent, "Component Showcase") {
		t.Error("Buffer missing 'Component Showcase' from Header mount")
	}

	// Check Card titles
	if !strings.Contains(bufContent, "System Info") {
		t.Error("Buffer missing 'System Info' from Card mount")
	}

	// Check StatusLine content
	if !strings.Contains(bufContent, "Version:") {
		t.Error("Buffer missing 'Version:' from StatusLine mount")
	}
	if !strings.Contains(bufContent, "1.2.0") {
		t.Error("Buffer missing '1.2.0' from StatusLine mount")
	}

	// Check StatusBar content
	if !strings.Contains(bufContent, "Build passed") {
		t.Error("Buffer missing 'Build passed' from StatusBar mount")
	}

	// Check help text
	if !strings.Contains(bufContent, "j/k scroll") {
		t.Error("Buffer missing 'j/k scroll' help text")
	}

	t.Logf("Buffer:\n%s", bufContent)

	// Dump the element tree to see layout rects
	if el, ok := app.root.(*Element); ok {
		t.Logf("Tree:\n%s", dumpTree(el, 0))
	}
}

// exactGenRootComponent replicates the EXACT structure from generated components_gsx.go
type exactGenRootComponent struct{}

func (c *exactGenRootComponent) Render(app *App) *Element {
	// Matches: func (a *componentsApp) Render(app *tui.App) *tui.Element
	__tui_0 := New(
		WithDirection(Column),
		WithPadding(2),
		WithGap(2),
	)

	// @Header("Component Showcase")
	__tui_1 := app.Mount(c, 0, func() Component {
		return genHeader("Component Showcase")
	})
	__tui_0.AddChild(__tui_1)

	// User cards row
	__tui_2 := New(WithDisplay(DisplayFlex), WithDirection(Row), WithGap(2))
	__tui_3 := app.Mount(c, 1, func() Component {
		return genUserCard(app, c, "Alice", "Engineer", true)
	})
	__tui_2.AddChild(__tui_3)
	__tui_4 := app.Mount(c, 2, func() Component {
		return genUserCard(app, c, "Bob", "Designer", false)
	})
	__tui_2.AddChild(__tui_4)
	__tui_0.AddChild(__tui_2)

	// Cards with children row
	__tui_6 := New(WithDisplay(DisplayFlex), WithDirection(Row), WithGap(2))

	// Card("System Info") with StatusLine children
	__tui_7_children := []*Element{}
	__tui_8 := app.Mount(c, 5, func() Component {
		return genStatusLine("Version:", "1.2.0")
	})
	__tui_7_children = append(__tui_7_children, __tui_8)
	__tui_9 := app.Mount(c, 6, func() Component {
		return genStatusLine("Uptime:", "3d 14h")
	})
	__tui_7_children = append(__tui_7_children, __tui_9)
	__tui_7 := app.Mount(c, 4, func() Component {
		return genCard("System Info", __tui_7_children)
	})
	__tui_6.AddChild(__tui_7)
	__tui_0.AddChild(__tui_6)

	// @StatusBar()
	__tui_18 := app.Mount(c, 13, func() Component {
		return genStatusBar()
	})
	__tui_0.AddChild(__tui_18)

	// Help text
	__tui_19 := New(WithDisplay(DisplayFlex), WithDirection(Row), WithJustify(JustifyCenter))
	__tui_20 := New(WithText("j/k scroll|q to quit"), WithTextStyle(NewStyle().Dim()))
	__tui_19.AddChild(__tui_20)
	__tui_0.AddChild(__tui_19)

	return __tui_0
}

// --- Replicas of generated function templs ---

type genHeaderView struct{ Root *Element }

func (v genHeaderView) Render(app *App) *Element { return v.Root }

func genHeader(title string) genHeaderView {
	div := New(
		WithBorder(BorderRounded),
		WithBorderGradient(NewGradient(Cyan, Magenta).WithDirection(GradientHorizontal)),
		WithPadding(1),
		WithDisplay(DisplayFlex), WithDirection(Row),
		WithJustify(JustifyCenter),
	)
	span := New(
		WithText(title),
		WithTextGradient(NewGradient(Cyan, Magenta).WithDirection(GradientHorizontal)),
		WithTextStyle(NewStyle().Bold()),
	)
	div.AddChild(span)
	return genHeaderView{Root: div}
}

type genCardView struct{ Root *Element }

func (v genCardView) Render(app *App) *Element { return v.Root }

func genCard(title string, children []*Element) genCardView {
	div := New(
		WithBorder(BorderRounded),
		WithPadding(1),
		WithDirection(Column),
		WithGap(1),
		WithFlexGrow(1.0),
	)
	titleSpan := New(
		WithText(title),
		WithTextGradient(NewGradient(Cyan, Magenta).WithDirection(GradientHorizontal)),
		WithTextStyle(NewStyle().Bold()),
	)
	div.AddChild(titleSpan)
	hr := New(WithHR(), WithBorder(BorderSingle))
	div.AddChild(hr)
	for _, child := range children {
		div.AddChild(child)
	}
	return genCardView{Root: div}
}

type genStatusLineView struct{ Root *Element }

func (v genStatusLineView) Render(app *App) *Element { return v.Root }

func genStatusLine(label, value string) genStatusLineView {
	div := New(WithDisplay(DisplayFlex), WithDirection(Row), WithGap(1))
	l := New(WithText(label), WithTextStyle(NewStyle().Dim()))
	div.AddChild(l)
	v := New(WithText(value), WithTextStyle(NewStyle().Foreground(Cyan).Bold()))
	div.AddChild(v)
	return genStatusLineView{Root: div}
}

type genUserCardView struct{ Root *Element }

func (uc genUserCardView) Render(app *App) *Element { return uc.Root }

func genUserCard(app *App, parent Component, name, role string, online bool) genUserCardView {
	children := []*Element{}
	roleSpan := New(WithText(role), WithTextStyle(NewStyle().Dim()))
	children = append(children, roleSpan)

	statusText := "Offline"
	if online {
		statusText = "Online"
	}
	badge := New(WithText(statusText))
	children = append(children, badge)

	card := genCard(name, children)
	return genUserCardView{Root: card.Root}
}

type genStatusBarView struct{ Root *Element }

func (v genStatusBarView) Render(app *App) *Element { return v.Root }

func genStatusBar() genStatusBarView {
	div := New(
		WithBorder(BorderRounded),
		WithPadding(1),
		WithDisplay(DisplayFlex), WithDirection(Row),
		WithGap(2),
		WithJustify(JustifyCenter),
	)
	b1 := New(WithText("Build passed"))
	div.AddChild(b1)
	sep := New(WithText("|"), WithTextStyle(NewStyle().Dim()))
	div.AddChild(sep)
	b2 := New(WithText("3 warnings"))
	div.AddChild(b2)
	return genStatusBarView{Root: div}
}

