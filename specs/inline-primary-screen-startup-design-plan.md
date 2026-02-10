# Inline Primary-Screen Startup Design + Implementation Plan

**Status:** Planned  
**Version:** 1.0  
**Last Updated:** 2026-02-10

---

## 1. Problem

Inline apps (`WithInlineHeight`) can start in a terminal that already has visible content in the history area (including the prior shell prompt and previous inline-session output). The framework currently assumes that history area is empty on startup.

Because of that assumption:
- New inline session content can visually mix with stale lines from previous runs.
- First appended lines are written into what the app believes are blank rows, while stale rows remain above.
- Users see old transcript lines and new transcript lines in the same visible block.

This is visible in `examples/ai-chat` and any inline app that prints session messages with `PrintAboveln`.

---

## 2. Goals

1. Preserve primary-screen behavior (no mandatory alternate-screen mode).
2. Avoid global scrollback wipe (`ESC[3J`) in all inline startup modes.
3. Make inline startup deterministic and artifact-free.
4. Support three startup behaviors: preserve-visible with gradual drain, fresh viewport clear, and soft-reset with scrollback preservation.
5. Keep backward compatibility for existing apps unless they opt in to new behavior.

### 2.1 Mode Semantics

| Mode | Visible on launch | How old rows leave | Scrollback effect |
| --- | --- | --- | --- |
| `InlineStartupPreserveVisible` | Old rows remain visible | Gradual drain during `PrintAbove/PrintAboveln` appends | Preserved |
| `InlineStartupFreshViewport` | Clean viewport immediately | Immediate clear | Old visible rows discarded (not moved to scrollback) |
| `InlineStartupSoftReset` | Clean viewport immediately | Immediate newline-driven push | Preserved (moved to scrollback) |

## 3. Non-Goals

1. Rewriting the inline renderer.
2. Changing `PrintAbove/PrintAboveln` semantics.
3. Forcing all inline apps to render transcript in component state.
4. Requiring terminal CPR support as a hard dependency.

---

## 4. Current Behavior (Root Cause)

Current startup path (`NewApp` and `NewAppWithReader`):
- Inline mode reserves bottom rows by printing newlines, then moves cursor up.
- Inline layout state is initialized as empty.
- Buffer starts as all spaces and first frame is diff-based unless explicitly marked full redraw.

Code touch points:
- `/Users/joelholsteen/go/src/github.com/grindlemire/go-tui/app.go` (inline bootstrap)
- `/Users/joelholsteen/go/src/github.com/grindlemire/go-tui/app_render.go` (inline diff/full redraw path)
- `/Users/joelholsteen/go/src/github.com/grindlemire/go-tui/inline_session.go` (history geometry assumptions)
- `/Users/joelholsteen/go/src/github.com/grindlemire/go-tui/examples/ai-chat/chat_gsx.go` (`PrintAboveln` transcript output)

The important mismatch is: startup assumes history block is empty, but terminal content is often unknown/non-empty.

Secondary issue:
- Preserve-visible startup should model unknown visible history as full/unknown, so runtime append logic scrolls the full history region and naturally drains stale rows.

---

## 5. Proposed Design

### 5.1 Inline Startup Policy (New API)

Add startup policy to inline mode so app authors can choose behavior explicitly.

New type:

```go
type InlineStartupMode int

const (
    // Backward-compatible launch visuals. Keep existing visible lines, then
    // treat the history area conservatively so stale rows drain naturally.
    InlineStartupPreserveVisible InlineStartupMode = iota

    // Clear the visible inline viewport at startup (history area + widget area),
    // but do not run a global scrollback wipe command.
    InlineStartupFreshViewport

    // Soft-reset the viewport by scrolling visible content upward via newline flow.
    // Preserves scrollback by pushing old visible rows into it.
    InlineStartupSoftReset
)
```

New option:

```go
func WithInlineStartupMode(mode InlineStartupMode) AppOption
```

Default value:
- `InlineStartupPreserveVisible` for backward compatibility.

### 5.2 Startup Ownership Handshake

Add a shared internal inline bootstrap helper used by both constructors.

Proposed sequence for inline app startup:
1. Clamp `inlineHeight` to terminal height.
2. Establish safe line boundary before reserving viewport (`\r\n` guard).
3. Reserve widget rows at bottom.
4. Compute `inlineStartRow`, initialize session/buffer.
5. Apply startup policy:
- `PreserveVisible`: keep visible history, initialize layout as **unknown** (`invalid`), and let first append treat the history block as full so scrolling starts from row 0.
- `FreshViewport`: clear visible viewport region once (no scrollback clear), initialize layout as empty.
- `SoftReset`: push visible rows into scrollback by controlled newline flow, initialize layout as empty.
6. Force first inline frame to full redraw (`needsFullRedraw = true`).

### 5.3 Layout Initialization + Runtime Append Rules

Current `newInlineLayoutState(...)` creates a "known-empty" layout. That is wrong for preserve-visible startup.

Startup-mode-aware initialization:
- Preserve mode: `inlineLayout.invalidate(historyCapacity)`
- Fresh/Soft modes: `inlineLayout.resetEmpty(historyCapacity)`

Runtime behavior for preserve mode:
- Existing `appendText` logic already does `resetConservativeFull(...)` when layout is invalid.
- That sets `contentStartRow=0`, `visibleRows=historyCapacity`.
- As a result, existing `appendRow` behavior scrolls from row 0 and naturally drains stale rows into scrollback.

Important constraint:
- Do **not** globally force `appendRow` to scroll from row 0 for all sessions.
- That would introduce blank-line scrollback churn for known-empty startup cases.

This approach gives Claude-like gradual cleanup for unknown startup history without regressing empty-start behavior.

### 5.4 Rendering Rule Change

Guarantee first inline render is full redraw:
- Set `needsFullRedraw = true` at inline startup.
- Keep current behavior for subsequent frames (diff rendering).

This removes startup dependence on unknown prior front-buffer state.

### 5.5 Optional Follow-Up (App-Level Transcript Ownership)

For chat apps that should never mix sessions in visible transcript:
- Render transcript inside component state (scrollbox/list).
- Keep `PrintAboveln` for logs/streaming only.

This is recommended for example apps but not required for framework correctness.

### 5.6 Why Keep Both `FreshViewport` and `SoftReset`

`SoftReset` is not strictly better than `FreshViewport`; they optimize for different tradeoffs.

1. `FreshViewport` provides deterministic clear with minimal scrollback churn, but intentionally discards current visible rows.
2. `SoftReset` preserves retrievability of old visible rows, but can add scrollback noise (including blank rows) and is more terminal-behavior-sensitive than direct clear.

---

## 6. Internal API/Code Changes

### 6.1 App fields

Add to `App`:
- `inlineStartupMode InlineStartupMode`

### 6.2 Options

Update `app_options.go`:
- Add enum definition and docs.
- Add `WithInlineStartupMode(...)` option.
- Document behavior and compatibility notes under `WithInlineHeight` docs.

### 6.3 Constructors

Update both constructors in `app.go`:
- Extract duplicated inline init into one internal helper:
- `setupInlineMode(width, termHeight int)` or similar.
- Include startup policy execution and `needsFullRedraw = true`.

### 6.4 Inline session/layout

Update `inline_session.go`:
- Reuse existing `invalidate`/`resetEmpty` based on startup mode.
- Keep existing append/resize algorithms for known-empty sessions.
- Verify preserve-visible startup triggers conservative-full append behavior (row-0 drain) via invalid layout initialization.

### 6.5 Examples

Update inline examples to demonstrate startup policy:
- `examples/ai-chat/main.go`: set explicit mode (default `InlineStartupPreserveVisible`, optional `InlineStartupSoftReset`).
- Optionally `examples/claude-chat/main.go` and `examples/inline-test/main.go` for consistency.

---

## 7. Implementation Plan

## Phase 1: Core Startup Policy + First Frame Ownership

1. Add `InlineStartupMode` and `WithInlineStartupMode`.
2. Add `inlineStartupMode` field to `App` with preserve-visible default.
3. Refactor inline constructor branch into shared helper.
4. Add startup policy handling and layout initialization rules.
5. Set `needsFullRedraw = true` at inline startup.

Acceptance criteria:
- No behavior change for apps not setting `WithInlineStartupMode`.
- Preserve-visible mode no longer assumes empty history and drains stale rows gradually during appends.
- First inline frame always fully paints widget region.
- Known-empty startup paths do not regress into blank-line scrollback churn.

## Phase 2: Example Policy Adoption

1. Update `ai-chat` to choose explicit startup policy, with `InlineStartupPreserveVisible` as default recommendation and `InlineStartupSoftReset` as immediate-clean alternative.
2. Optionally opt other inline examples into explicit modes.
3. Update docs with startup-mode guidance and tradeoffs.

Acceptance criteria:
- Running `examples/ai-chat` in a dirty terminal either drains stale rows naturally (preserve mode) or starts clean while preserving old rows in scrollback (soft-reset mode).
- Scrollback remains intact.

## Phase 3: Optional UX Enhancements

1. Evaluate transcript-in-state migration for `ai-chat`.
2. Add optional convenience helper (e.g., `WithInlineFreshStart()`) if desired.
3. Add optional cursor-position-aware boundary logic (CPR with timeout) if needed for advanced terminals.

Acceptance criteria:
- Session transcript ownership is explicit at app level.
- Startup behavior remains deterministic across terminal implementations.

---

## 8. Test Plan

### 8.1 Unit tests (framework)

Add to `app_inline_test.go`:
1. `TestInlineStartup_PreserveVisible_UsesConservativeLayout`
2. `TestInlineStartup_PreserveVisible_AppendsDrainFromRowZero`
3. `TestInlineStartup_PreserveVisible_KnownEmptyDoesNotGenerateBlankScrollback`
4. `TestInlineStartup_FreshViewport_ClearsVisibleWithoutScrollbackWipe`
5. `TestInlineStartup_SoftReset_PushesViewportIntoScrollback`
6. `TestInlineStartup_FirstRenderIsFullRedraw`
7. `TestInlineStartup_DefaultModeIsPreserveVisible`

### 8.2 Regression tests (artifact scenarios)

Add emulator-based scenarios:
1. Seed terminal with old prompt + stale transcript lines, then start inline app.
2. Append 1-3 new `PrintAboveln` lines.
3. Verify expected on-screen chronology per startup mode.

### 8.3 Example-level validation

Manual check for `examples/ai-chat`:
1. Start from non-cleared terminal with prior run output visible.
2. Run app, submit several lines.
3. Confirm selected mode behavior: preserve mode drains stale rows naturally as new lines arrive; soft-reset starts with clean viewport and preserved scrollback.

---

## 9. Risks and Mitigations

1. Risk: Startup mode defaults break existing expectations.
- Mitigation: keep default as preserve-visible.

2. Risk: Preserve-visible still shows old rows at launch.
- Mitigation: document that this is intentional Claude-like behavior; apps needing clean launch can select `SoftReset` or `FreshViewport`.

3. Risk: Soft-reset behavior varies across terminals.
- Mitigation: mark soft-reset as advanced/optional; keep fresh viewport simple and deterministic.

4. Risk: Boundary newline adds extra vertical movement in some shells.
- Mitigation: keep boundary logic minimal and document behavior; optionally gate with follow-up CPR enhancement.

5. Risk: Confusion between framework-level viewport ownership and app-level transcript ownership.
- Mitigation: document clearly that `PrintAboveln` writes terminal history, not component-owned transcript state.

---

## 10. Rollout Strategy

1. Merge phase 1 with no default behavior changes.
2. Update `ai-chat` to explicit startup policy in phase 2.
3. Announce recommended startup mode matrix:
- log/monitor apps: preserve-visible
- chat/session apps: preserve-visible (gradual drain) or soft-reset (immediate clean)
- privacy/clean-launch apps: fresh viewport

---

## 11. Success Criteria

1. Preserve-visible mode gives Claude-like gradual cleanup (stale rows drain via normal append flow).
2. Inline apps no longer exhibit stale-line startup artifacts when using soft-reset or fresh-viewport modes.
3. Framework has explicit, documented startup semantics for inline ownership.
4. Existing inline apps remain stable without forced migration.
5. Global scrollback wipe is avoided in all modes.
6. Scrollback-preserving behavior is available for both preserve-visible and soft-reset modes.
