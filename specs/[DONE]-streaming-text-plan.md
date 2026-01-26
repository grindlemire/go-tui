# Streaming Text Implementation Plan

Implementation phases for streaming text support. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Add onUpdate Hook to Element

**Reference:** [streaming-text-design.md §3](./streaming-text-design.md#3-core-entities)

**Completed in commit:** (pending review)

- [x] Modify `pkg/tui/element/element.go`
  - Add `onUpdate func()` field to Element struct
  - Add `SetOnUpdate(fn func())` method
  - See [streaming-text-design.md §3 - Element onUpdate Hook](./streaming-text-design.md#element-onupdate-hook)

- [x] Modify `pkg/tui/element/render.go`
  - Call `e.onUpdate()` at the start of `Render()` method (before layout calculation)
  - Only call if `e.onUpdate != nil`

- [x] Modify `pkg/tui/element/options.go`
  - Add `WithOnUpdate(fn func()) Option` function
  - See [streaming-text-design.md §3 - WithOnUpdate Option](./streaming-text-design.md#withonupdate-option)

- [x] Add tests in `pkg/tui/element/element_test.go`
  - Test that onUpdate hook is called during Render()
  - Test that nil hook doesn't cause panic
  - Test WithOnUpdate option sets the hook correctly

**Tests:** PASS - 4 new tests added

---

## Phase 2: Create Streaming Example

**Reference:** [streaming-text-design.md §4](./streaming-text-design.md#4-user-experience)

**Completed in commit:** (pending review)

- [x] Create `examples/streaming/main.go`
  - Define local `StreamBox` struct with:
    - `elem *element.Element` - the scrollable container
    - `textCh <-chan string` - channel for incoming text
    - `textStyle tui.Style` - style for text lines
    - `autoScroll bool` - whether to auto-scroll on new content
  - Implement `NewStreamBox(textCh <-chan string) *StreamBox`
    - Create scrollable Element with vertical scroll and column direction
    - Set up onUpdate hook to call `poll()` method
    - Initialize autoScroll to true
  - Implement `Element() *element.Element` accessor
  - Implement `poll()` method
    - Non-blocking select loop to drain channel
    - Split text on newlines and add child elements
    - Call ScrollToBottom() if autoScroll is true
  - Implement `appendText(text string)` helper
    - Split on `\n`, create Element for each non-empty line
    - Add as children to the scrollable element
  - Implement auto-scroll detection
    - Check if user is at bottom before adding content
    - Set autoScroll = false when user scrolls up
    - Set autoScroll = true when user scrolls to bottom
  - Create main() with:
    - Header explaining controls
    - StreamBox in the main area
    - Footer showing scroll position
    - Simulated process goroutine sending timestamped lines
    - Event loop handling ESC to exit, j/k for scroll, dispatch to focused element
  - See [streaming-text-design.md §3 - Example StreamBox Pattern](./streaming-text-design.md#example-streambox-pattern)

**Tests:** PASS - manual testing via running the example (`go run ./examples/streaming`)

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Add onUpdate hook to Element | Complete |
| 2 | Create streaming example | Complete |

## Files to Modify

| File | Changes |
|------|---------|
| `pkg/tui/element/element.go` | Add onUpdate field and SetOnUpdate method |
| `pkg/tui/element/render.go` | Call onUpdate hook before layout |
| `pkg/tui/element/options.go` | Add WithOnUpdate option |
| `pkg/tui/element/element_test.go` | Add tests for onUpdate hook |

## Files to Create

```
examples/
└── streaming/
    └── main.go        # Streaming demo with local StreamBox
```
