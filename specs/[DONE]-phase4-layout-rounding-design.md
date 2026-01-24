# Phase 4: Yoga-Style Layout Rounding

**Status:** Complete
**Version:** 1.0
**Last Updated:** 2025-01-24

---

## 1. Overview

### Purpose

This phase implements jitter-free animation by adopting Yoga's absolute float rounding strategy and consolidating all centering logic into the layout system.

### Problem Statement

When centering happens at multiple levels (e.g., panel centered in root, text centered in panel), independent integer rounding at each level causes visual jitter during animation:

```
Panel width 30→31→32, terminal width 100:
- Panel X: floor(70/2)=35, floor(69/2)=34, floor(68/2)=34 → 35, 34, 34
- Text offset: floor(8/2)=4, floor(9/2)=4, floor(10/2)=5 → 4, 4, 5
- Absolute X: 39, 38, 39 ← JITTER (position oscillates!)
```

The fundamental issue: `round(a) + round(b) ≠ round(a + b)`

### Solution

1. **Absolute Float Rounding**: Track cumulative float positions through the layout tree, round only once when computing the final integer Rect
2. **Centering Consolidation**: Text elements use intrinsic width, parent's `AlignItems` handles centering—eliminating text-level centering

### Goals

- Eliminate animation jitter for centered elements
- Simplify the mental model: layout handles all positioning
- Maintain backward compatibility for explicit-size text elements
- Match industry-standard approach (Yoga/React Native)

### Non-Goals

- Subpixel rendering (terminals are cell-based)
- Font metrics or text wrapping (future work)

---

## 2. Architecture

### Data Flow: Before vs After

**Before (jitter-prone):**
```
Parent int position → add float offset → round to int → pass to child
Child int position → add float offset → round to int → pass to grandchild
```

**After (jitter-free):**
```
Parent float position → add float offset → keep as float → pass to child
Child float position → add float offset → keep as float → pass to grandchild
...
Final: round(absolute float position) → integer Rect for rendering
```

### Modified Types

#### Layout Struct (`pkg/layout/layout.go`)

```go
type Layout struct {
    Rect        Rect    // Integer border box (rounded)
    ContentRect Rect    // Integer content box (rounded)

    // NEW: Absolute float positions for jitter-free child positioning
    AbsoluteX float64   // True X position before rounding
    AbsoluteY float64   // True Y position before rounding
}
```

#### Text Element (`pkg/tui/element/text.go`)

Text elements now have intrinsic sizing:

```go
func NewText(content string, opts ...TextOption) *Text {
    t := &Text{
        Element: New(),
        content: content,
    }
    // Intrinsic size based on content
    t.Element.style.Width = layout.Fixed(stringWidth(content))
    t.Element.style.Height = layout.Fixed(1)
    // Options may override
    for _, opt := range opts {
        opt(t)
    }
    return t
}
```

---

## 3. Implementation Details

### Calculate Entry Point (`pkg/layout/calculate.go`)

```go
func Calculate(root Layoutable, availableWidth, availableHeight int) {
    // ...
    calculateNode(root, available, 0.0, 0.0)  // Start with float (0,0)
}

func calculateNode(node Layoutable, available Rect, absoluteX, absoluteY float64) {
    // 1. Compute dimensions
    borderBox := computeBorderBox(style, available)

    // 2. Set position from ROUNDED absolute float
    borderBox.X = int(math.Round(absoluteX))
    borderBox.Y = int(math.Round(absoluteY))

    // 3. Content rect from float position
    contentAbsX := absoluteX + float64(style.Padding.Left)
    contentAbsY := absoluteY + float64(style.Padding.Top)

    // 4. Pass FLOAT positions to children
    layoutChildren(node, contentRect, contentAbsX, contentAbsY)

    // 5. Store float positions for child calculations
    node.SetLayout(Layout{
        Rect:        borderBox,
        ContentRect: contentRect,
        AbsoluteX:   absoluteX,
        AbsoluteY:   absoluteY,
    })
}
```

### Child Positioning (`pkg/layout/flex.go`)

```go
func layoutChildren(node Layoutable, contentRect Rect, parentAbsX, parentAbsY float64) {
    // ... Phase 1-5 unchanged (compute sizes and relative positions) ...

    // Phase 6: Convert to rects using ABSOLUTE float positions
    for i, child := range children {
        // Compute ABSOLUTE float position
        var childAbsX, childAbsY float64
        if isRow {
            childAbsX = parentAbsX + items[i].mainPos   // mainPos is float64
            childAbsY = parentAbsY + items[i].crossPos
        } else {
            childAbsX = parentAbsX + items[i].crossPos
            childAbsY = parentAbsY + items[i].mainPos
        }

        // Round ONCE for integer Rect
        slot := Rect{
            X:      int(math.Round(childAbsX)),
            Y:      int(math.Round(childAbsY)),
            Width:  items[i].mainSize,
            Height: items[i].crossSize,
        }

        // Recurse with FLOAT position
        calculateNode(child, childBorderBox, childAbsX, childAbsY)
    }
}
```

### Text Rendering (`pkg/tui/element/render.go`)

```go
func renderTextContent(buf *tui.Buffer, t *Text) {
    contentRect := t.ContentRect()
    textWidth := stringWidth(t.content)
    x := contentRect.X

    // Only apply text-level alignment if element is wider than text
    // (user explicitly set larger size)
    if contentRect.Width > textWidth {
        switch t.align {
        case TextAlignCenter:
            x += (contentRect.Width - textWidth) / 2
        case TextAlignRight:
            x += contentRect.Width - textWidth
        }
    }
    // Otherwise: draw at origin, layout already centered the element

    buf.SetString(x, contentRect.Y, t.content, t.contentStyle)
}
```

---

## 4. Why This Works

### Mathematical Proof

With panel width 30→31→32, terminal 100, text 18 chars:

**Old approach (jitter):**
```
Frame 1: round(35.0) + round(4.0) = 35 + 4 = 39
Frame 2: round(34.5) + round(4.5) = 35 + 5 = 40  ← jumped!
Frame 3: round(34.0) + round(5.0) = 34 + 5 = 39  ← jumped back!
```

**New approach (stable):**
```
Frame 1: round(35.0 + 4.0) = round(39.0) = 39
Frame 2: round(34.5 + 4.5) = round(39.0) = 39  ← same!
Frame 3: round(34.0 + 5.0) = round(39.0) = 39  ← same!
```

The fractional parts cancel out when added as floats before rounding.

### Centering Consolidation Benefit

When Text elements have intrinsic width:
- Element width = text width (e.g., 18 for "Layout Engine Demo")
- Parent's `AlignCenter` positions the element
- Only ONE centering calculation exists
- Absolute float rounding handles it perfectly

---

## 5. API Changes

### New Pattern (Recommended)

```go
// Parent handles centering via AlignItems
panel := element.New(
    element.WithDirection(layout.Column),
    element.WithAlign(layout.AlignCenter),  // Centers children
)

// Text element uses intrinsic size - no alignment options needed
title := element.NewText("Hello World",
    element.WithTextStyle(tui.NewStyle().Bold()),
)

panel.AddChild(title.Element)
```

### Legacy Pattern (Still Supported)

```go
// Explicit size with text-level centering (for buttons, etc.)
button := element.NewText("OK",
    element.WithTextAlign(element.TextAlignCenter),
    element.WithElementOption(element.WithSize(20, 3)),  // Fixed width
)
```

When element width > text width, text-level alignment still applies.

---

## 6. Files Modified

| File | Changes |
|------|---------|
| `pkg/layout/layout.go` | Added `AbsoluteX`, `AbsoluteY` float64 fields |
| `pkg/layout/calculate.go` | Pass float positions through `calculateNode` |
| `pkg/layout/flex.go` | Compute absolute float positions, `math.Round()` once |
| `pkg/tui/element/text.go` | Set intrinsic width/height in `NewText()` |
| `pkg/tui/element/render.go` | Conditional text alignment (only if explicit size) |

---

## 7. Testing

All existing tests pass. Key test scenarios:

1. **Backward compatibility**: Explicit-size text with `TextAlignCenter` still centers
2. **Intrinsic sizing**: Text elements sized to content by default
3. **Layout rounding**: Float positions round correctly at tree boundaries
4. **Animation**: Dashboard example animates without jitter at any step size

---

## 8. References

- [Yoga Layout Engine](https://yogalayout.com/) - Facebook's cross-platform layout library
- [YGRoundValueToPixelGrid](https://github.com/facebook/yoga/blob/main/yoga/algorithm/PixelGrid.cpp) - Yoga's rounding implementation
- React Native's approach to preventing layout jitter
