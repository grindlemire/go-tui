# Phase 6: Focus & Rendering Improvements Implementation Plan

Implementation phases for the Focus & Rendering Improvements. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Text on Element

**Reference:** [phase6-focus-rendering-improvements-design.md §3.1](./phase6-focus-rendering-improvements-design.md#31-element-text-fields)

- [x] Update `pkg/tui/element/element.go`
  - Add `text string` field to Element struct
  - Add `textStyle tui.Style` field to Element struct
  - Add `textAlign TextAlign` field to Element struct
  - Move `TextAlign` type and constants from `text.go` to `element.go`
  - Move `stringWidth` function from `render.go` to `element.go` (or keep in render.go)
  - Add `WithText(content string) Option` - sets text and intrinsic size (width=text width, height=1)
  - Add `WithTextStyle(style tui.Style) Option`
  - Add `WithTextAlign(align TextAlign) Option`
  - Add `SetText(content string)` method - updates text and recalculates intrinsic width, marks dirty
  - Add `Text() string` method
  - Add `SetTextStyle(style tui.Style)` method
  - Add `TextStyle() tui.Style` method
  - Add `SetTextAlign(align TextAlign)` method
  - Add `TextAlign() TextAlign` method

- [x] Update `pkg/tui/element/render.go`
  - Update `renderElement` to call `renderTextContent` if `e.text != ""`
  - Update `renderTextContent` to work with `*Element` instead of `*Text`
    - Get text from `e.text`
    - Get style from `e.textStyle`
    - Get align from `e.textAlign`
  - Remove `RenderText(buf *tui.Buffer, t *Text)` function
  - Remove `RenderTextTree` function
  - Remove `renderElementWithText` function

- [x] Remove `pkg/tui/element/text.go`
  - Delete the entire file (Text struct, NewText, TextOption, etc.)

- [x] Update `pkg/tui/element/element_test.go`
  - Add tests for `WithText` setting intrinsic size
  - Add tests for `SetText` updating width and marking dirty
  - Add tests for text style and alignment getters/setters

- [x] Update `pkg/tui/element/render_test.go`
  - Add tests for text rendering in `renderElement`
  - Test text alignment (left, center, right)
  - Test empty text (should not render)

**Tests:** Run `go test ./pkg/tui/element/... -v` once at phase end - ✅ PASSED

---

## Phase 2: Focus on Element

**Reference:** [phase6-focus-rendering-improvements-design.md §3.2](./phase6-focus-rendering-improvements-design.md#32-element-focus-fields)

- [x] Update `pkg/tui/element/element.go`
  - Add `focusable bool` field to Element struct
  - Add `focused bool` field to Element struct
  - Add `onFocus func()` field to Element struct
  - Add `onBlur func()` field to Element struct
  - Add `onEvent func(tui.Event) bool` field to Element struct
  - Add `WithOnFocus(fn func()) Option` - sets onFocus AND sets focusable=true
  - Add `WithOnBlur(fn func()) Option` - sets onBlur AND sets focusable=true
  - Add `WithOnEvent(fn func(tui.Event) bool) Option` - sets onEvent AND sets focusable=true
  - Add `IsFocusable() bool` method - returns e.focusable
  - Add `IsFocused() bool` method - returns e.focused
  - Add `Focus()` method - sets focused=true, calls onFocus if set, cascades to children
  - Add `Blur()` method - sets focused=false, calls onBlur if set, cascades to children
  - Add `HandleEvent(event tui.Event) bool` method - calls onEvent if set, returns result

- [x] Update `pkg/tui/element/element_test.go`
  - Add tests for `WithOnFocus` implying focusable=true
  - Add tests for `WithOnBlur` implying focusable=true
  - Add tests for `WithOnEvent` implying focusable=true
  - Add tests for `Focus()` setting focused and calling callback
  - Add tests for `Blur()` clearing focused and calling callback
  - Add tests for focus cascading to children
  - Add tests for blur cascading to children
  - Add tests for `HandleEvent` delegation

**Tests:** Run `go test ./pkg/tui/element/... -v` once at phase end - ✅ PASSED

---

## Phase 3: Auto-Registration and App Integration

**Reference:** [phase6-focus-rendering-improvements-design.md §3.3, §3.5, §3.6](./phase6-focus-rendering-improvements-design.md#33-child-notification)

- [x] Update `pkg/tui/element/element.go`
  - Add `onChildAdded func(*Element)` field to Element struct
  - Add `onFocusableAdded func(tui.Focusable)` field to Element struct
  - Add `SetOnChildAdded(fn func(*Element))` method
  - Add `SetOnFocusableAdded(fn func(tui.Focusable))` method
  - Add `WalkFocusables(fn func(tui.Focusable))` method
  - Update `AddChild` to call `notifyChildAdded` for each child added
  - Add `notifyChildAdded(child *Element)` method - walks up to root, calls callbacks if set

- [x] Update `pkg/tui/focus.go`
  - Keep `FocusManager.elements` as `[]Focusable` (avoids circular import)
  - Update `NewFocusManager()` to take no arguments and initialize with `current: -1`
  - Element satisfies Focusable interface via compile-time check
  - Existing methods work with Focusable interface

- [x] Update `pkg/tui/app.go`
  - Add `focusableTreeWalker` interface for auto-discovery
  - Update `SetRoot(root Renderable)` to:
    - Store root
    - If root implements focusableTreeWalker, set up callback and discover focusables
  - Add `FocusNext()` method - calls `a.focus.Next()`
  - Add `FocusPrev()` method - calls `a.focus.Prev()`
  - Add `Focused() Focusable` method - returns `a.focus.Focused()`
  - Keep `Focus() *FocusManager` for backward compatibility (deprecated)

- [x] Update `pkg/tui/element/element_test.go`
  - Add tests for `SetOnChildAdded` callback
  - Add tests for `AddChild` triggering callback
  - Add tests for callback walking up to root
  - Add tests for `SetOnFocusableAdded` callback
  - Add tests for `WalkFocusables`

- [x] Update `pkg/tui/focus_test.go`
  - Update tests to use `NewFocusManager()` then `Register()`
  - Keep using mockFocusable for testing

- [x] Update `pkg/tui/app_test.go`
  - Add tests for `SetRoot` discovering focusables via mockFocusableTreeWalker
  - Add tests for auto-registration via callback
  - Add tests for `FocusNext`, `FocusPrev`, `Focused`

- [x] Update `examples/focus/main.go`
  - Remove `FocusableBox` struct
  - Use `element.New` with `WithOnFocus`/`WithOnBlur` directly
  - Remove manual `fm.Register()` calls (auto-registered via SetRoot)
  - Use `app.FocusNext()` instead of `fm.Next()`
  - Remove separate `element.RenderText()` calls (text rendered automatically)
  - Simplify the example significantly

- [x] Update `examples/dashboard/main.go`
  - Update to use new text API (element.New with WithText)
  - Remove `element.RenderText()` calls
  - Use `app.Render()` for rendering

- [x] Update `examples/hello_layout/main.go`
  - Update to use new text API

**Tests:** Run `go test ./pkg/tui/... -v` once at phase end - ✅ PASSED

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Text on Element | ✅ Complete |
| 2 | Focus on Element | ✅ Complete |
| 3 | Auto-Registration and App Integration | ✅ Complete |

## Files to Modify

```
pkg/tui/element/
├── element.go      # Add text and focus fields, options, methods
├── element_test.go # Add tests for new functionality
├── render.go       # Update to render text, remove Text-specific functions
└── render_test.go  # Update tests

pkg/tui/
├── focus.go        # Update to work with *Element
├── focus_test.go   # Update tests
├── app.go          # Add SetRoot discovery, FocusNext/Prev/Focused
└── app_test.go     # Add tests

examples/
├── focus/main.go   # Simplify using new API
└── dashboard/main.go # Update if needed
```

## Files to Remove

```
pkg/tui/element/
└── text.go         # Remove entirely
```

## Breaking Changes

| Change | Migration |
|--------|-----------|
| `Text` struct removed | Use `element.New(WithText(...))` |
| `NewText` removed | Use `element.New(WithText(...))` |
| `RenderText` removed | Text rendered automatically in `RenderTree` |
| `FocusManager` uses `*Element` | Elements now implement focus directly |
| `App.SetRoot` takes `*Element` | Pass Element directly, not Renderable |
