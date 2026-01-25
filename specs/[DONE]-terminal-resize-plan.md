# Terminal Resize Resilience Implementation Plan

Implementation phases for terminal resize resilience. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Auto-Upgrade Render After Resize (App)

**Reference:** [terminal-resize-design.md §2](./terminal-resize-design.md#2-architecture)

**Completed in commit:** (pending review)

- [x] Modify `pkg/tui/app.go`
  - Add `needsFullRedraw bool` field to `App` struct
  - Modify `Dispatch()` to set `needsFullRedraw = true` when handling `ResizeEvent`
  - Modify `Render()` to check `needsFullRedraw` flag:
    - If true: call `RenderFull()` logic instead of normal diff-based render, then reset flag
    - If false: use existing diff-based render
  - See [terminal-resize-design.md §3](./terminal-resize-design.md#3-core-entities)

- [x] Create `pkg/tui/app_resize_test.go`
  - Test that `Dispatch(ResizeEvent)` sets `needsFullRedraw` flag
  - Test that `Render()` clears the flag after full redraw
  - Test that multiple Render() calls after single resize only do one full redraw
  - Use mock terminal to verify full vs diff render behavior

**Tests:** PASS

---

## Phase 2: Debounce Resize Events (EventReader)

**Reference:** [terminal-resize-design.md §2](./terminal-resize-design.md#2-architecture)

**Completed in commit:** (pending review)

- [x] Modify `pkg/tui/reader.go`
  - Add `lastResizeTime time.Time` field to `stdinReader` struct
  - Add `pendingResize *ResizeEvent` field to buffer the latest resize
  - Modify `PollEvent()` to debounce SIGWINCH signals:
    - On SIGWINCH: record time and update `pendingResize` with current terminal size
    - Only emit `ResizeEvent` when 16ms has passed since last SIGWINCH
    - If another SIGWINCH arrives within 16ms window, update size but don't emit yet
  - Ensure final resize is always emitted (don't swallow the last one)

- [x] Create `pkg/tui/reader_resize_test.go`
  - Test that single SIGWINCH emits one ResizeEvent after 16ms
  - Test that rapid SIGWINCH signals are coalesced into one event
  - Test that final size is correct (uses last signal's dimensions)
  - Test that normal events (KeyEvent) are not affected by debouncing

**Tests:** PASS

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Auto-upgrade Render() to full redraw after resize | Complete |
| 2 | Debounce rapid resize events in EventReader | Complete |

## Files to Create

```
pkg/tui/
├── app_resize_test.go    (new - resize behavior tests)
└── reader_resize_test.go (new - debounce tests)
```

## Files to Modify

| File | Changes |
|------|---------|
| `pkg/tui/app.go` | Add needsFullRedraw flag, modify Dispatch() and Render() |
| `pkg/tui/reader.go` | Add debouncing fields and logic to PollEvent() |
