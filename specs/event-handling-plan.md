# Event Handling Implementation Plan

Implementation phases for the event handling system with push-based watchers, dirty tracking, and app lifecycle management. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Dirty Tracking & Watcher Types

**Reference:** [event-handling-design.md §3.1-3.2](./event-handling-design.md#31-dirty-tracking)

**Status:** Complete

- [x] Create `pkg/tui/dirty.go`
  - Add `dirty atomic.Bool` package-level variable
  - Implement `MarkDirty()` function that sets the flag
  - Implement `checkAndClearDirty()` internal function that swaps and returns
  - See [design §3.1](./event-handling-design.md#31-dirty-tracking)

- [x] Create `pkg/tui/watcher.go`
  - Define `Watcher` interface with `Start(eventQueue, stopCh)` method
  - See [design §3.2](./event-handling-design.md#32-watcher-types)

- [x] Implement `channelWatcher[T any]` in `pkg/tui/watcher.go`
  - Struct with `ch <-chan T` and `handler func(T)` fields
  - Implement `Watch[T any](ch <-chan T, handler func(T)) Watcher` constructor
  - Implement `Start(eventQueue, stopCh)` that spawns goroutine
  - Goroutine selects on `stopCh` and `w.ch`
  - On channel receive, enqueue `func() { w.handler(val) }` to `eventQueue`
  - Exit on channel close or stopCh signal

- [x] Implement `timerWatcher` in `pkg/tui/watcher.go`
  - Struct with `interval time.Duration` and `handler func()` fields
  - Implement `OnTimer(interval time.Duration, handler func()) Watcher` constructor
  - Implement `Start(eventQueue, stopCh)` that spawns goroutine
  - Use `time.NewTicker` with proper cleanup via `defer ticker.Stop()`
  - Select on `stopCh` and `ticker.C`
  - On tick, enqueue handler to `eventQueue`

- [x] Add tests to `pkg/tui/dirty_test.go`
  - Test `MarkDirty()` sets flag to true
  - Test `checkAndClearDirty()` returns true and clears flag
  - Test `checkAndClearDirty()` returns false when not dirty
  - Test concurrent calls to `MarkDirty()` are safe

- [x] Add tests to `pkg/tui/watcher_test.go`
  - Test `Watch()` creates watcher that receives channel values
  - Test channel watcher exits when channel closes
  - Test channel watcher exits when stopCh closes
  - Test `OnTimer()` creates watcher that fires at interval
  - Test timer watcher exits when stopCh closes
  - Test handler is called on main loop (via eventQueue)

**Acceptance:** `go test ./pkg/tui/... -run "Dirty|Watcher"` passes

---

## Phase 2: App.Run() & SetRoot

**Reference:** [event-handling-design.md §3.3](./event-handling-design.md#33-app-changes)

**Status:** Not Started

- [ ] Modify `pkg/tui/app.go` - App struct
  - Add `eventQueue chan func()` field (buffered, size 256)
  - Add `stopCh chan struct{}` field
  - Add `stopped bool` field
  - Add `globalKeyHandler func(KeyEvent) bool` field
  - See [design §3.3](./event-handling-design.md#33-app-changes)

- [ ] Modify `pkg/tui/app.go` - NewApp()
  - Initialize `eventQueue: make(chan func(), 256)`
  - Initialize `stopCh: make(chan struct{})`
  - Initialize `stopped: false`

- [ ] Create `Viewable` interface in `pkg/tui/app.go`
  - `GetRoot() *element.Element`
  - `GetWatchers() []Watcher`

- [ ] Implement `SetRoot(v any)` in `pkg/tui/app.go`
  - Type switch on `Viewable` and `*element.Element`
  - For Viewable: extract root, start all watchers
  - For Element: set root directly

- [ ] Implement `SetGlobalKeyHandler(fn func(KeyEvent) bool)` in `pkg/tui/app.go`
  - Store handler in `globalKeyHandler` field
  - See [design §3.3](./event-handling-design.md#33-app-changes)

- [ ] Implement `Run()` in `pkg/tui/app.go`
  - Set up SIGINT handler with `signal.Notify`
  - Goroutine catches signal and calls `Stop()`
  - Start `readInputEvents()` goroutine
  - Main loop: block on eventQueue or stopCh
  - Drain additional queued events (batch processing)
  - Check `checkAndClearDirty()` and call `Render()` if true
  - Return nil when stopped

- [ ] Implement `Stop()` in `pkg/tui/app.go`
  - Check idempotency (return if already stopped)
  - Set `stopped = true`
  - Close `stopCh` to signal all goroutines

- [ ] Implement `QueueUpdate(fn func())` in `pkg/tui/app.go`
  - Non-blocking send to eventQueue
  - Safe to call from any goroutine

- [ ] Implement `readInputEvents()` in `pkg/tui/app.go`
  - Loop checking stopCh
  - Poll for events with timeout
  - Enqueue handler that:
    - Calls globalKeyHandler first (if set)
    - If handler returns true, event is consumed
    - Otherwise, dispatch to focused element

- [ ] Add tests to `pkg/tui/app_test.go`
  - Test `SetRoot` with Viewable extracts root and starts watchers
  - Test `SetRoot` with raw Element sets root directly
  - Test `Run()` blocks until `Stop()` is called
  - Test `Stop()` is idempotent (multiple calls safe)
  - Test events are batched (multiple events, single render)
  - Test `QueueUpdate()` enqueues function safely from goroutine
  - Test `SetGlobalKeyHandler` intercepts keys before dispatch
  - Test global handler returning true consumes event
  - Test global handler returning false passes to element
  - Test SIGINT triggers graceful shutdown

**Acceptance:** `go test ./pkg/tui/... -run App` passes

---

## Phase 3: Element Handler Changes

**Reference:** [event-handling-design.md §3.4-3.5](./event-handling-design.md#34-element-handler-changes)

**Status:** Not Started

- [ ] Modify `pkg/tui/element/element.go` - handler fields
  - Add `onKeyPress func(tui.KeyEvent)` field (no bool return)
  - Add `onClick func()` field (no bool return)
  - See [design §3.4](./event-handling-design.md#34-element-handler-changes)

- [ ] Implement `SetOnKeyPress(fn func(tui.KeyEvent))` in `pkg/tui/element/element.go`
  - Set handler field directly

- [ ] Implement `SetOnClick(fn func())` in `pkg/tui/element/element.go`
  - Set handler field directly

- [ ] Update mutating methods to call `MarkDirty()` in `pkg/tui/element/element.go`
  - `ScrollBy(dx, dy int)` - call `tui.MarkDirty()` after mutation
  - `SetText(text string)` - call `tui.MarkDirty()` after mutation
  - `AddChild(children ...*Element)` - call `tui.MarkDirty()` after mutation
  - `RemoveAllChildren()` - call `tui.MarkDirty()` after mutation
  - Any other mutating methods that affect rendering

- [ ] Modify `pkg/tui/element/options.go`
  - Add `WithOnKeyPress(fn func(tui.KeyEvent)) Option`
  - Add `WithOnClick(fn func()) Option`
  - See [design §3.5](./event-handling-design.md#35-element-options)

- [ ] Update event dispatch in element
  - Modify `HandleEvent()` (or equivalent) to call handlers without expecting bool return
  - Remove any existing bool return handling

- [ ] Add tests to `pkg/tui/element/element_test.go`
  - Test `SetOnKeyPress` sets handler
  - Test `SetOnClick` sets handler
  - Test `ScrollBy` marks dirty
  - Test `SetText` marks dirty
  - Test `AddChild` marks dirty
  - Test `RemoveAllChildren` marks dirty
  - Test handler is called on event dispatch

- [ ] Add tests to `pkg/tui/element/options_test.go`
  - Test `WithOnKeyPress` option sets handler
  - Test `WithOnClick` option sets handler

**Acceptance:** `go test ./pkg/tui/element/... -run "Element|Options"` passes

---

## Phase 4: Generator Updates

**Reference:** [event-handling-design.md §5](./event-handling-design.md#5-generated-code)

**Status:** Not Started

- [ ] Modify `pkg/tuigen/generator.go` - Viewable interface
  - View structs already have `GetRoot()` from named refs work
  - Add `GetWatchers() []tui.Watcher` method to generated view structs
  - Add `watchers []tui.Watcher` field to view structs

- [ ] Modify `pkg/tuigen/generator.go` - watcher collection
  - Parse `onChannel={tui.Watch(...)}` attributes
  - Parse `onTimer={tui.OnTimer(...)}` attributes
  - Store watcher expressions for later emission

- [ ] Modify `pkg/tuigen/generator.go` - watcher registration
  - Generate `watchers = append(watchers, ...)` for each watcher attribute
  - See [design §5.1](./event-handling-design.md#51-channel-watcher-registration)

- [ ] Modify `pkg/tuigen/generator.go` - watcher aggregation
  - When component contains child components, aggregate their watchers
  - Generate `watchers = append(watchers, childView.GetWatchers()...)`
  - See [design §5.4](./event-handling-design.md#54-watcher-aggregation-from-nested-components)

- [ ] Modify `pkg/tuigen/generator.go` - handler options
  - Generate `element.WithOnKeyPress(handler)` for `onKeyPress={handler}`
  - Generate `element.WithOnClick(handler)` for `onClick={handler}`
  - No bool return expected from handlers

- [ ] Add tests to `pkg/tuigen/generator_test.go`
  - Test `onChannel` attribute generates watcher registration
  - Test `onTimer` attribute generates watcher registration
  - Test view struct has `GetWatchers()` method
  - Test nested component watchers are aggregated
  - Test `onKeyPress` attribute generates option
  - Test `onClick` attribute generates option
  - Test generated code compiles

- [ ] Update `examples/streaming-dsl/` to use event handling
  - Update `streaming.tui` with `onChannel`, `onTimer`, `onKeyPress`
  - Update `main.go` to use `SetRoot(view)` and `Run()`
  - Add `SetGlobalKeyHandler` for quit key
  - Verify example compiles and runs

**Acceptance:** `go test ./pkg/tuigen/... -run Generator` passes, examples build and run correctly

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Dirty Tracking & Watcher Types | Complete |
| 2 | App.Run() & SetRoot | Not Started |
| 3 | Element Handler Changes | Not Started |
| 4 | Generator Updates | Not Started |

## Files to Create

```
pkg/tui/
├── dirty.go           # NEW: Dirty flag management
├── dirty_test.go      # NEW: Dirty flag tests
├── watcher.go         # NEW: Watcher interface and implementations
└── watcher_test.go    # NEW: Watcher tests
```

## Files to Modify

```
pkg/tui/
├── app.go             # Add Run(), Stop(), SetRoot(), SetGlobalKeyHandler()
├── app_test.go        # Add tests for new app functionality

pkg/tui/element/
├── element.go         # Add onKeyPress, onClick; mutations mark dirty
├── element_test.go    # Add handler and dirty tests
├── options.go         # Add WithOnKeyPress, WithOnClick
└── options_test.go    # Add option tests

pkg/tuigen/
├── generator.go       # Generate Viewable interface, watcher collection
└── generator_test.go  # Add watcher generation tests

examples/streaming-dsl/
├── streaming.tui      # Use new event handling patterns
└── main.go            # Use SetRoot(view), Run(), SetGlobalKeyHandler()
```

## Files Unchanged

| File | Reason |
|------|--------|
| `pkg/tuigen/lexer.go` | No new tokens needed |
| `pkg/tuigen/parser.go` | Attributes already parsed |
| `pkg/tuigen/analyzer.go` | No new validation needed |
| `pkg/tuigen/tailwind.go` | No changes for events |
| `pkg/formatter/` | No event-specific formatting |
| `pkg/lsp/` | LSP support can be added later |

## Dependencies

```
Phase 1 ─────► Phase 2 ─────► Phase 4
                  │
                  ▼
              Phase 3 ─────► Phase 4
```

- Phase 2 depends on Phase 1 (needs dirty tracking and watcher types)
- Phase 3 depends on Phase 2 (needs MarkDirty to be available)
- Phase 4 depends on Phases 2 and 3 (generates code using both)

## Success Criteria

From [design §11](./event-handling-design.md#11-success-criteria):

1. `app.Run()` blocks and processes events until `app.Stop()` or SIGINT
2. Channel data triggers handler immediately (no polling delay)
3. Timer fires at specified interval
4. Mutations automatically mark dirty
5. Multiple events batched before single render
6. `SetGlobalKeyHandler()` intercepts keys before dispatch
7. Ctrl+C triggers graceful shutdown
8. Zero CPU when idle
9. Example streaming app works end-to-end
