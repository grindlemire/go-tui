# Tailwind Expansion Specification

**Status:** Planned\
**Version:** 1.0\
**Last Updated:** 2025-01-25

---

## 1. Overview

### Purpose

Expand Tailwind-style class support in go-tui to provide comprehensive styling convenience, and add robust editor feedback (error diagnostics with suggestions) and autocomplete for Tailwind classes inside the `class` attribute.

### Goals

- Add width/height percentage classes (w-1/2, w-full, h-1/2, etc.)
- Add individual side padding/margin classes (pt-2, pb-2, pl-2, pr-2, mt-2, mb-2, etc.)
- Add more flex utilities (flex-grow-0, flex-shrink-0, self-start, self-end, justify-evenly, etc.)
- Add border color classes (border-red, border-green, border-cyan, etc.)
- Add text alignment classes (text-left, text-center, text-right)
- Report unknown/invalid Tailwind classes as errors in the LSP
- Suggest similar valid classes when an unknown class is used ("did you mean...?")
- Provide autocomplete with documentation for Tailwind classes inside `class=""` attributes

### Non-Goals

- Full Tailwind CSS parity (terminal limitations)
- Responsive variants (sm:, md:, lg:)
- State variants (hover:, focus:, active:) - terminal doesn't support these
- Arbitrary values (e.g., w-[100px])

---

## 2. Architecture

### Directory Structure

```
pkg/
├── tuigen/
│   ├── tailwind.go          # Extended with new classes + validation
│   └── tailwind_test.go     # Extended with new test cases
└── lsp/
    ├── completion.go        # Extended for class attribute completion
    ├── diagnostics.go       # Extended for class validation diagnostics
    └── hover.go             # Possibly extended for class hover info
```

### Component Overview

| Component | Purpose |
|-----------|---------|
| `pkg/tuigen/tailwind.go` | Class parsing, validation, similarity matching |
| `pkg/lsp/completion.go` | Autocomplete for class attributes |
| `pkg/lsp/diagnostics.go` | Error reporting for unknown classes |

### Flow Diagram

```
.tui file edited
        │
        ▼
┌───────────────────┐
│  LSP Document     │
│  didChange        │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│  Analyzer.Analyze │──► ParseTailwindClasses()
└─────────┬─────────┘            │
          │                      ▼
          │              ┌───────────────────┐
          │              │ ValidateTailwind  │
          │              │ class (new)       │
          │              └─────────┬─────────┘
          │                        │
          │              unknown class?
          │                  │    │
          │                 yes   no
          │                  │    │
          │                  ▼    │
          │         SuggestSimilar│
          │              │        │
          │              ▼        │
          ▼         Add Error     │
    Diagnostics ◄────────────────┘
          │
          ▼
    publishDiagnostics
```

---

## 3. Core Entities

### New Tailwind Classes to Support

#### Width/Height Percentage Classes

| Class | Maps To | Description |
|-------|---------|-------------|
| `w-full` | `WithWidthPercent(100)` | Full width |
| `w-1/2` | `WithWidthPercent(50)` | Half width |
| `w-1/3` | `WithWidthPercent(33.33)` | Third width |
| `w-2/3` | `WithWidthPercent(66.67)` | Two-thirds width |
| `w-1/4` | `WithWidthPercent(25)` | Quarter width |
| `w-3/4` | `WithWidthPercent(75)` | Three-quarters width |
| `w-auto` | `WithWidth(layout.Auto())` | Auto width |
| `h-full` | `WithHeightPercent(100)` | Full height |
| `h-1/2` | `WithHeightPercent(50)` | Half height |
| `h-1/3` | `WithHeightPercent(33.33)` | Third height |
| `h-2/3` | `WithHeightPercent(66.67)` | Two-thirds height |
| `h-1/4` | `WithHeightPercent(25)` | Quarter height |
| `h-3/4` | `WithHeightPercent(75)` | Three-quarters height |
| `h-auto` | (default) | Auto height |

#### Individual Padding Classes

| Class Pattern | Maps To | Description |
|---------------|---------|-------------|
| `pt-N` | `WithPaddingTRBL(N, 0, 0, 0)` | Padding top |
| `pr-N` | `WithPaddingTRBL(0, N, 0, 0)` | Padding right |
| `pb-N` | `WithPaddingTRBL(0, 0, N, 0)` | Padding bottom |
| `pl-N` | `WithPaddingTRBL(0, 0, 0, N)` | Padding left |

#### Individual Margin Classes

| Class Pattern | Maps To | Description |
|---------------|---------|-------------|
| `mt-N` | `WithMarginTRBL(N, 0, 0, 0)` | Margin top |
| `mr-N` | `WithMarginTRBL(0, N, 0, 0)` | Margin right |
| `mb-N` | `WithMarginTRBL(0, 0, N, 0)` | Margin bottom |
| `ml-N` | `WithMarginTRBL(0, 0, 0, N)` | Margin left |
| `mx-N` | `WithMarginTRBL(0, N, 0, N)` | Margin horizontal |
| `my-N` | `WithMarginTRBL(N, 0, N, 0)` | Margin vertical |

#### Flex Utilities

| Class | Maps To | Description |
|-------|---------|-------------|
| `flex-grow-0` | `WithFlexGrow(0)` | Don't grow |
| `flex-grow-N` | `WithFlexGrow(N)` | Grow factor N |
| `flex-shrink-0` | `WithFlexShrink(0)` | Don't shrink |
| `flex-shrink-N` | `WithFlexShrink(N)` | Shrink factor N |
| `self-start` | `WithAlignSelf(layout.AlignStart)` | Align self start |
| `self-end` | `WithAlignSelf(layout.AlignEnd)` | Align self end |
| `self-center` | `WithAlignSelf(layout.AlignCenter)` | Align self center |
| `self-stretch` | `WithAlignSelf(layout.AlignStretch)` | Align self stretch |
| `justify-evenly` | `WithJustify(layout.JustifySpaceEvenly)` | Space evenly |
| `items-stretch` | `WithAlign(layout.AlignStretch)` | Align items stretch |
| `justify-around` | `WithJustify(layout.JustifySpaceAround)` | Space around |

#### Border Colors

| Class | Maps To | Description |
|-------|---------|-------------|
| `border-red` | `WithBorderStyle(tui.NewStyle().Foreground(tui.Red))` | Red border |
| `border-green` | `WithBorderStyle(tui.NewStyle().Foreground(tui.Green))` | Green border |
| `border-blue` | `WithBorderStyle(tui.NewStyle().Foreground(tui.Blue))` | Blue border |
| `border-cyan` | `WithBorderStyle(tui.NewStyle().Foreground(tui.Cyan))` | Cyan border |
| `border-magenta` | `WithBorderStyle(tui.NewStyle().Foreground(tui.Magenta))` | Magenta border |
| `border-yellow` | `WithBorderStyle(tui.NewStyle().Foreground(tui.Yellow))` | Yellow border |
| `border-white` | `WithBorderStyle(tui.NewStyle().Foreground(tui.White))` | White border |
| `border-black` | `WithBorderStyle(tui.NewStyle().Foreground(tui.Black))` | Black border |

#### Text Alignment

| Class | Maps To | Description |
|-------|---------|-------------|
| `text-left` | `WithTextAlign(element.TextAlignLeft)` | Align text left |
| `text-center` | `WithTextAlign(element.TextAlignCenter)` | Center text |
| `text-right` | `WithTextAlign(element.TextAlignRight)` | Align text right |

#### Existing Padding/Margin Classes (Already Supported)

For completeness, these classes already exist in the current implementation:

| Class Pattern | Maps To | Description |
|---------------|---------|-------------|
| `p-N` | `WithPadding(N)` | Padding all sides |
| `px-N` | `WithPaddingX(N)` | Padding horizontal |
| `py-N` | `WithPaddingY(N)` | Padding vertical |
| `m-N` | `WithMargin(N)` | Margin all sides |

---

## 3.1 Technical Design Details

### Padding/Margin Accumulation Strategy

**Problem:** Multiple individual side classes will overwrite each other:

```tui
<div class="pt-2 pb-4">  <!-- Without accumulation, only pb-4 takes effect -->
```

**Solution:** Accumulate padding/margin sides during `ParseTailwindClasses()` and emit a single merged call.

```go
// PaddingAccumulator tracks individual padding values
type PaddingAccumulator struct {
    Top, Right, Bottom, Left int
    HasTop, HasRight, HasBottom, HasLeft bool
}

// Merge combines an individual side class into the accumulator
func (p *PaddingAccumulator) Merge(side string, value int)

// ToOption generates WithPaddingTRBL() if any sides are set
func (p *PaddingAccumulator) ToOption() string
```

The `ParseTailwindClasses()` function will:
1. Detect individual padding classes (pt-N, pr-N, pb-N, pl-N)
2. Accumulate them into a `PaddingAccumulator`
3. At the end, emit a single `WithPaddingTRBL(T, R, B, L)` call with merged values
4. Same approach for margin classes

**Example Result:**
```tui
<div class="pt-2 pb-4 pl-1">
```
Generates:
```go
element.WithPaddingTRBL(2, 0, 4, 1)
```

### Width/Height Pattern Additions

Current patterns only match `w-N` and `h-N` (fixed integers). Add patterns for fractions and keywords:

```go
// Fraction patterns: w-1/2, w-2/3, etc.
widthFractionPattern  = regexp.MustCompile(`^w-(\d+)/(\d+)$`)
heightFractionPattern = regexp.MustCompile(`^h-(\d+)/(\d+)$`)

// Keyword patterns: w-full, w-auto, h-full, h-auto
widthKeywordPattern  = regexp.MustCompile(`^w-(full|auto)$`)
heightKeywordPattern = regexp.MustCompile(`^h-(full|auto)$`)
```

### New Element Options for Auto Values

The current `WithWidth(int)` only accepts fixed values. For `w-auto` and `h-auto`, add:

```go
// WithWidthAuto sets width to auto (size to content)
func WithWidthAuto() Option {
    return func(e *Element) {
        e.style.Width = layout.Auto()
    }
}

// WithHeightAuto sets height to auto (size to content)
func WithHeightAuto() Option {
    return func(e *Element) {
        e.style.Height = layout.Auto()
    }
}
```

Alternatively, generate direct style assignments in the Tailwind mapping:
```go
"w-auto": {Option: "element.WithWidthAuto()", NeedsImport: ""},
"h-auto": {Option: "element.WithHeightAuto()", NeedsImport: ""},
```

### Validation Types

```go
// TailwindValidationResult contains validation results for a class
type TailwindValidationResult struct {
    Valid      bool
    Class      string
    Suggestion string // "did you mean...?" hint
}

// ValidateTailwindClass validates a single class and returns suggestions
func ValidateTailwindClass(class string) TailwindValidationResult

// AllTailwindClasses returns all known classes for autocomplete
func AllTailwindClasses() []TailwindClassInfo

// TailwindClassInfo contains metadata about a class for autocomplete
type TailwindClassInfo struct {
    Name        string
    Category    string // "layout", "spacing", "typography", "visual"
    Description string
    Example     string
}
```

### LSP Position Tracking for Class Validation

To underline only the invalid class (not the entire `class` attribute value), track character positions within the class value string:

```go
// TailwindClassWithPosition tracks a class and its position within the attribute value
type TailwindClassWithPosition struct {
    Class    string
    StartCol int  // column offset relative to attribute value start
    EndCol   int  // column offset relative to attribute value start
    Valid    bool
    Suggestion string
}

// ParseTailwindClassesWithPositions parses classes and tracks their positions
func ParseTailwindClassesWithPositions(classes string, attrStartCol int) []TailwindClassWithPosition
```

### LSP Class Attribute Context Detection

For autocomplete, detect when cursor is inside a `class=""` attribute:

```go
// isInClassAttribute checks if cursor is inside a class attribute value
// Returns (isInClass, partialPrefix) where partialPrefix is what user has typed
func (s *Server) isInClassAttribute(doc *Document, pos Position) (bool, string) {
    // 1. Search backwards from cursor for class="
    // 2. If found and not past closing quote, return true
    // 3. Extract partial class being typed (text after last space before cursor)
    ...
}
```

### Similar Class Matching

```go
// similarClasses maps common typos/alternatives to correct class names
var similarClasses = map[string]string{
    "flex-column":     "flex-col",
    "flex-columns":    "flex-col",
    "flex-rows":       "flex-row",
    "gap":             "gap-1",
    "padding":         "p-1",
    "margin":          "m-1",
    "bold":            "font-bold",
    "italic":          "font-italic",
    "dim":             "font-dim",
    "border-single":   "border",
    "width":           "w-1",
    "height":          "h-1",
    "center":          "text-center",
    "left":            "text-left",
    "right":           "text-right",
    "align-center":    "text-center",
    "align-left":      "text-left",
    "align-right":     "text-right",
    "grow":            "flex-grow",
    "shrink":          "flex-shrink",
    "no-grow":         "flex-grow-0",
    "no-shrink":       "flex-shrink-0",
    "padding-top":     "pt-1",
    "padding-bottom":  "pb-1",
    "padding-left":    "pl-1",
    "padding-right":   "pr-1",
    "margin-top":      "mt-1",
    "margin-bottom":   "mb-1",
    "margin-left":     "ml-1",
    "margin-right":    "mr-1",
    // ... etc
}
```

---

## 4. User Experience

### Error Diagnostics

When an unknown class is used:

```tui
<div class="flex-columns gap-2">  // "flex-columns" underlined red
                                  // Error: unknown Tailwind class "flex-columns" (did you mean "flex-col"?)
```

### Autocomplete

When typing inside `class=""`:

```
class="flex-|"     // Cursor at |
                   // Shows:
                   // - flex (Display flex row)
                   // - flex-col (Display flex column)
                   // - flex-grow (Grow to fill space)
                   // - flex-shrink (Shrink when needed)
                   // - flex-grow-0 (Don't grow)
                   // - flex-shrink-0 (Don't shrink)
```

Each completion shows:
- **Label**: The class name (e.g., `flex-col`)
- **Detail**: Short description (e.g., "Display flex column")
- **Documentation**: Full markdown with example usage

### CLI Error Output

```
$ tui check ./...
component.tui:15:12: unknown Tailwind class "flex-columns" (did you mean "flex-col"?)
```

---

## 5. Complexity Assessment

| Size | Phases | When to Use |
|------|--------|-------------|
| Small | 1-2 | Single component, bug fix, minor enhancement |
| Medium | 3-4 | New feature touching multiple files/components |
| Large | 5-6 | Cross-cutting feature, new subsystem |

**Assessed Size:** Medium

**Recommended Phases:** 4

**Rationale:** This feature touches multiple components (tailwind.go, analyzer.go, LSP completion, LSP diagnostics) but doesn't require new subsystems. The changes are additive extensions to existing patterns. The four phases are:
1. Expand Tailwind class mappings in `tailwind.go`
2. Add validation and similarity matching to `tailwind.go`
3. Integrate validation into analyzer and LSP diagnostics
4. Add class autocomplete to LSP completion handler

> **IMPORTANT:** User must approve the complexity assessment before proceeding to implementation plan. The plan MUST use the approved number of phases.

---

## 6. Success Criteria

1. All new Tailwind classes from Section 3 are supported and generate correct Go code:
   - Width/height percentages (w-full, w-1/2, w-1/3, w-2/3, w-1/4, w-3/4, h-full, etc.)
   - Individual padding (pt-N, pr-N, pb-N, pl-N)
   - Individual margin (mt-N, mr-N, mb-N, ml-N, mx-N, my-N)
   - Flex utilities (flex-grow-N, flex-shrink-N, self-start, self-end, self-center, self-stretch, justify-evenly, justify-around, items-stretch)
   - Border colors (border-red, border-green, border-blue, border-cyan, border-magenta, border-yellow, border-white, border-black)
   - Text alignment (text-left, text-center, text-right)
2. Unknown classes produce error diagnostics in the editor with red underlines
3. Error messages include "did you mean...?" suggestions for similar valid classes
4. Autocomplete shows all available Tailwind classes when typing inside `class=""`
5. Each autocomplete item has a description and category
6. `tui check` reports unknown classes as errors
7. All existing tests continue to pass
8. New classes have corresponding test coverage

---

## 7. Open Questions

1. ~~What new classes to support?~~ → All categories: percentages, individual sides, flex utilities, border colors, text alignment
2. ~~Error vs warning for unknown classes?~~ → Error (red squiggly)
3. ~~Include autocomplete?~~ → Yes, with documentation
4. ~~Suggest similar classes?~~ → Yes, "did you mean...?"

