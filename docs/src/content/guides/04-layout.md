# Layout

## Overview

go-tui uses a CSS flexbox-compatible layout engine. Every `<div>` is a flex container, and its children are flex items that you arrange with direction, alignment, spacing, and sizing controls. You can set all of these through Tailwind-style classes or through element attributes directly.

If you've used CSS flexbox, the mental model is the same: a container has a main axis (the direction children flow) and a cross axis (perpendicular to it). Properties like `justify-*` control the main axis, `items-*` control the cross axis, and `gap-*` adds space between children.

## Direction

Every flex container lays out its children along a main axis. The default is `Row` (horizontal, left to right). Use `flex-col` to switch to `Column` (vertical, top to bottom).

```gsx
templ DirectionDemo() {
    <div class="flex-col items-center w-full gap-2 p-1">
        <span class="font-bold">Row (default):</span>
        <div class="flex gap-1">
            <span class="bg-cyan text-black px-1">A</span>
            <span class="bg-cyan text-black px-1">B</span>
            <span class="bg-cyan text-black px-1">C</span>
        </div>

        <span class="font-bold">Column:</span>
        <div class="flex-col">
            <span class="bg-magenta text-black px-1">A</span>
            <span class="bg-magenta text-black px-1">B</span>
            <span class="bg-magenta text-black px-1">C</span>
        </div>
    </div>
}
```

In the row layout, A, B, and C sit side by side. In the column layout, they stack vertically.

| Class | Direction | Description |
|-------|-----------|-------------|
| `flex` or `flex-row` | Row | Children flow left to right (default) |
| `flex-col` | Column | Children flow top to bottom |

You can also set direction via the attribute: `direction={tui.Column}`.

## Justify Content

Justify controls how children are distributed along the **main axis** (horizontal for rows, vertical for columns).

```gsx
templ JustifyDemo() {
    <div class="flex-col items-center w-full gap-1 p-1">
        <span class="font-dim">justify-start (default):</span>
        <div class="flex justify-start border-single w-40">
            <span class="bg-cyan text-black px-1">A</span>
            <span class="bg-cyan text-black px-1">B</span>
            <span class="bg-cyan text-black px-1">C</span>
        </div>

        <span class="font-dim">justify-center:</span>
        <div class="flex justify-center border-single w-40">
            <span class="bg-cyan text-black px-1">A</span>
            <span class="bg-cyan text-black px-1">B</span>
            <span class="bg-cyan text-black px-1">C</span>
        </div>

        <span class="font-dim">justify-end:</span>
        <div class="flex justify-end border-single w-40">
            <span class="bg-cyan text-black px-1">A</span>
            <span class="bg-cyan text-black px-1">B</span>
            <span class="bg-cyan text-black px-1">C</span>
        </div>

        <span class="font-dim">justify-between:</span>
        <div class="flex justify-between border-single w-40">
            <span class="bg-cyan text-black px-1">A</span>
            <span class="bg-cyan text-black px-1">B</span>
            <span class="bg-cyan text-black px-1">C</span>
        </div>

        <span class="font-dim">justify-around:</span>
        <div class="flex justify-around border-single w-40">
            <span class="bg-cyan text-black px-1">A</span>
            <span class="bg-cyan text-black px-1">B</span>
            <span class="bg-cyan text-black px-1">C</span>
        </div>

        <span class="font-dim">justify-evenly:</span>
        <div class="flex justify-evenly border-single w-40">
            <span class="bg-cyan text-black px-1">A</span>
            <span class="bg-cyan text-black px-1">B</span>
            <span class="bg-cyan text-black px-1">C</span>
        </div>
    </div>
}
```

| Class | Behavior |
|-------|----------|
| `justify-start` | Pack children at the start (default) |
| `justify-center` | Center children along the main axis |
| `justify-end` | Pack children at the end |
| `justify-between` | Even spacing between children, no space at edges |
| `justify-around` | Even spacing around each child (half-space at edges) |
| `justify-evenly` | Equal spacing between children and at edges |

You can also use the attribute form: `justify={tui.JustifyCenter}`.

## Align Items

Align controls how children are positioned along the **cross axis** (vertical for rows, horizontal for columns).

```gsx
templ AlignDemo() {
    <div class="flex-col items-center w-full gap-1 p-1">
        <span class="font-dim">items-start:</span>
        <div class="flex items-start gap-1 border-single h-5 w-40">
            <span class="bg-cyan text-black px-1">Short</span>
            <span class="bg-magenta text-black px-1">Taller\ntext</span>
        </div>

        <span class="font-dim">items-center:</span>
        <div class="flex items-center gap-1 border-single h-5 w-40">
            <span class="bg-cyan text-black px-1">Short</span>
            <span class="bg-magenta text-black px-1">Taller\ntext</span>
        </div>

        <span class="font-dim">items-end:</span>
        <div class="flex items-end gap-1 border-single h-5 w-40">
            <span class="bg-cyan text-black px-1">Short</span>
            <span class="bg-magenta text-black px-1">Taller\ntext</span>
        </div>

        <span class="font-dim">items-stretch (default):</span>
        <div class="flex items-stretch gap-1 border-single h-5 w-40">
            <span class="bg-cyan text-black px-1">Short</span>
            <span class="bg-magenta text-black px-1">Taller\ntext</span>
        </div>
    </div>
}
```

| Class | Behavior |
|-------|----------|
| `items-start` | Align children to the start of the cross axis |
| `items-center` | Center children on the cross axis |
| `items-end` | Align children to the end of the cross axis |
| `items-stretch` | Stretch children to fill the cross axis (default) |

Individual children can override the container's alignment with `self-start`, `self-center`, `self-end`, or `self-stretch`.

## Gap

Gap adds uniform spacing between children along the main axis. It does not add space before the first child or after the last.

```gsx
templ GapDemo() {
    <div class="flex-col items-center w-full gap-1 p-1">
        <span class="font-dim">gap-0 (no gap):</span>
        <div class="flex gap-0">
            <span class="bg-cyan text-black px-1">A</span>
            <span class="bg-cyan text-black px-1">B</span>
            <span class="bg-cyan text-black px-1">C</span>
        </div>

        <span class="font-dim">gap-1:</span>
        <div class="flex gap-1">
            <span class="bg-cyan text-black px-1">A</span>
            <span class="bg-cyan text-black px-1">B</span>
            <span class="bg-cyan text-black px-1">C</span>
        </div>

        <span class="font-dim">gap-3:</span>
        <div class="flex gap-3">
            <span class="bg-cyan text-black px-1">A</span>
            <span class="bg-cyan text-black px-1">B</span>
            <span class="bg-cyan text-black px-1">C</span>
        </div>
    </div>
}
```

Use `gap-N` where N is the number of character cells between children. You can also set it via the attribute: `gap={2}`.

## Flex Grow and Shrink

Grow and shrink control how children claim extra space or give it up when the container doesn't match the total size of its children.

### Growing

`grow` (or `flex-grow-1`) tells an element to expand and fill available space. The grow value is relative to siblings: if two siblings both have `grow`, they split the extra space equally. If one has `flex-grow-2` and another has `flex-grow-1`, the first gets twice as much extra space.

```gsx
templ GrowDemo() {
    <div class="flex-col items-center w-full gap-1 p-1">
        <span class="font-dim">Fixed sidebar + growing content:</span>
        <div class="flex gap-1 w-50 border-single">
            <div class="w-12 bg-magenta text-black p-1">
                <span>Sidebar</span>
            </div>
            <div class="grow bg-cyan text-black p-1">
                <span>Content (fills remaining space)</span>
            </div>
        </div>

        <span class="font-dim">Two panels, equal grow:</span>
        <div class="flex gap-1 w-50 border-single">
            <div class="grow bg-cyan text-black p-1">
                <span>Left</span>
            </div>
            <div class="grow bg-magenta text-black p-1">
                <span>Right</span>
            </div>
        </div>
    </div>
}
```

### Shrinking

By default, flex items can shrink below their natural size when the container is too small (`flex-shrink` defaults to 1). Use `shrink-0` to prevent an element from shrinking, or `flex-shrink-N` to set a relative shrink factor.

### Shorthand Classes

| Class | Effect |
|-------|--------|
| `grow` | `flex-grow: 1` — expand to fill space |
| `grow-0` | `flex-grow: 0` — don't grow |
| `shrink` | `flex-shrink: 1` — allow shrinking |
| `shrink-0` | `flex-shrink: 0` — don't shrink |
| `flex-1` | `flex-grow: 1, flex-shrink: 1` — grow and shrink equally |
| `flex-auto` | Same as `flex-1` |
| `flex-initial` | `flex-grow: 0, flex-shrink: 1` — shrink but don't grow |
| `flex-none` | `flex-grow: 0, flex-shrink: 0` — fixed size |
| `flex-grow-N` | Set grow factor to N |
| `flex-shrink-N` | Set shrink factor to N |

You can also use attributes: `flexGrow={1.5}`, `flexShrink={0}`.

## Flex Wrap

By default, children stay on a single line even if they overflow the container. Add `flex-wrap` to let items break onto new lines when they run out of room.

```gsx
templ WrapDemo() {
    <div class="flex flex-wrap gap-1 w-40 border-single p-1">
        <span class="bg-cyan text-black px-2 shrink-0">Alpha</span>
        <span class="bg-cyan text-black px-2 shrink-0">Bravo</span>
        <span class="bg-cyan text-black px-2 shrink-0">Charlie</span>
        <span class="bg-cyan text-black px-2 shrink-0">Delta</span>
        <span class="bg-cyan text-black px-2 shrink-0">Echo</span>
    </div>
}
```

Items that don't fit on the first line wrap to the next. Each line handles grow, shrink, and justify on its own.

| Class | Behavior |
|-------|----------|
| `flex-wrap` | Items wrap to new lines when they overflow |
| `flex-wrap-reverse` | Items wrap in reverse order (last line appears first) |
| `flex-nowrap` | Items stay on one line (default) |

You can also use the attribute form: `flexWrap={tui.Wrap}`.

### Align Content

When wrapping produces multiple lines, `align-content` controls how the lines are spaced along the cross axis. It requires at least two lines and some free cross-axis space to have a visible effect.

```gsx
templ AlignContentDemo() {
    <div class="flex flex-wrap gap-1 h-20 w-40 border-single content-center">
        <span class="bg-cyan text-black px-2 shrink-0">A</span>
        <span class="bg-cyan text-black px-2 shrink-0">B</span>
        <span class="bg-cyan text-black px-2 shrink-0">C</span>
        <span class="bg-cyan text-black px-2 shrink-0">D</span>
    </div>
}
```

| Class | Behavior |
|-------|----------|
| `content-start` | Pack lines at the start of the cross axis (default) |
| `content-end` | Pack lines at the end |
| `content-center` | Center lines in the cross axis |
| `content-stretch` | Stretch lines to fill the cross axis |
| `content-between` | First line at start, last line at end, even spacing between |
| `content-around` | Equal spacing around each line |

You can also use the attribute form: `alignContent={tui.ContentCenter}`.

## Sizing

By default, elements size to their content (`Auto`). You can set explicit sizes in character cells, percentages, or keep the default auto behavior.

### Fixed Sizes

```gsx
<div class="w-30 h-10 border-rounded p-1">
    <span>30 characters wide, 10 rows tall</span>
</div>
```

### Percentage Sizes

```gsx
<div class="flex gap-1 w-full">
    <div class="w-1/3 bg-cyan text-black p-1">
        <span>1/3 width</span>
    </div>
    <div class="w-2/3 bg-magenta text-black p-1">
        <span>2/3 width</span>
    </div>
</div>
```

### Full and Auto

```gsx
<div class="w-full h-full">
    // Takes all available width and height
    <div class="w-auto h-auto">
        // Sizes to its content (the default)
    </div>
</div>
```

### Min and Max Constraints

```gsx
<div class="grow min-w-20 max-w-60 p-1 border-single">
    <span>Grows with available space, but stays between 20 and 60 characters wide</span>
</div>
```

### Size Class Reference

| Class | Effect |
|-------|--------|
| `w-N` | Fixed width of N characters |
| `h-N` | Fixed height of N rows |
| `w-full` | 100% of parent width |
| `h-full` | 100% of parent height |
| `w-auto` | Width sizes to content (default) |
| `h-auto` | Height sizes to content (default) |
| `w-1/2` | 50% of parent width |
| `w-1/3` | 33.3% of parent width |
| `w-2/3` | 66.7% of parent width |
| `h-1/2` | 50% of parent height |
| `h-1/3` | 33.3% of parent height |
| `h-2/3` | 66.7% of parent height |
| `min-w-N` | Minimum width of N characters |
| `max-w-N` | Maximum width of N characters |
| `min-h-N` | Minimum height of N rows |
| `max-h-N` | Maximum height of N rows |

For attributes, use `width={30}`, `widthPercent={50}`, `height={10}`, `heightPercent={100}`, `minWidth={20}`, `maxWidth={60}`, `minHeight={5}`, `maxHeight={20}`.

## Padding and Margin

Padding adds space inside an element's border. Margin adds space outside it. Both are measured in character cells.

```gsx
templ SpacingDemo() {
    <div class="flex justify-center w-full gap-2 p-1">
        <div class="border-single p-2">
            <span>2 cells of padding inside the border</span>
        </div>
        <div class="border-single m-2">
            <span>2 cells of margin outside the border</span>
        </div>
    </div>
}
```

### Per-Side Control

You can set padding and margin on individual sides or axis pairs:

| Class | Sides Affected |
|-------|----------------|
| `p-N` | All four sides |
| `px-N` | Left and right |
| `py-N` | Top and bottom |
| `pt-N` | Top only |
| `pr-N` | Right only |
| `pb-N` | Bottom only |
| `pl-N` | Left only |
| `m-N` | All four sides |
| `mx-N` | Left and right |
| `my-N` | Top and bottom |
| `mt-N` | Top only |
| `mr-N` | Right only |
| `mb-N` | Bottom only |
| `ml-N` | Left only |

For more control via attributes, use `padding={2}` for uniform or set each side explicitly with the `tui.WithPaddingTRBL(top, right, bottom, left)` option in Go.

## Common Layout Patterns

### Sidebar and Main Content

```gsx
templ SidebarLayout() {
    <div class="flex h-full">
        <div class="w-20 border-single flex-col p-1">
            <span class="font-bold">Sidebar</span>
            <span>Navigation</span>
            <span>Settings</span>
        </div>
        <div class="grow flex-col p-1">
            <span class="font-bold">Content</span>
            <span>The main area fills the remaining width.</span>
        </div>
    </div>
}
```

The sidebar has a fixed width of 20 characters. The content area grows to fill the rest of the row.

### Centered Card

```gsx
templ CenteredCard() {
    <div class="flex items-center justify-center h-full">
        <div class="border-rounded p-2 flex-col gap-1 w-40">
            <span class="font-bold text-cyan">Welcome</span>
            <hr />
            <span>This card is centered both horizontally and vertically.</span>
        </div>
    </div>
}
```

The outer container fills the terminal (`h-full`) and centers its child on both axes with `items-center` and `justify-center`. Because the default direction is `Row`, `justify-center` centers horizontally and `items-center` centers vertically.

### Dashboard Grid

```gsx
templ Dashboard() {
    <div class="flex-col h-full gap-1 p-1">
        <div class="flex gap-1 grow">
            <div class="grow border-rounded p-1 flex-col">
                <span class="font-bold text-cyan">CPU</span>
                <span>45%</span>
            </div>
            <div class="grow border-rounded p-1 flex-col">
                <span class="font-bold text-green">Memory</span>
                <span>2.1 GB</span>
            </div>
            <div class="grow border-rounded p-1 flex-col">
                <span class="font-bold text-yellow">Disk</span>
                <span>67%</span>
            </div>
        </div>
        <div class="flex gap-1 grow">
            <div class="w-2/3 border-rounded p-1 flex-col">
                <span class="font-bold">Network Activity</span>
                <span>Sparkline goes here</span>
            </div>
            <div class="grow border-rounded p-1 flex-col">
                <span class="font-bold">Events</span>
                <span>Log feed goes here</span>
            </div>
        </div>
    </div>
}
```

The top row has three equally-sized panels (each with `grow`). The bottom row uses `w-2/3` for a wider panel and `grow` for the remaining one.

### Stacked Form Fields

```gsx
templ FormLayout() {
    <div class="flex-col gap-1 p-2 w-40">
        <div class="flex-col">
            <span class="font-bold">Username</span>
            <input placeholder="Enter username" class="border-single" />
        </div>
        <div class="flex-col">
            <span class="font-bold">Password</span>
            <input placeholder="Enter password" class="border-single" />
        </div>
        <div class="flex justify-end gap-1 pt-1">
            <button class="bg-cyan text-black px-2">Submit</button>
        </div>
    </div>
}
```

Each field is a vertical stack of label and input. The button row uses `justify-end` to push the button to the right.

### Wrapping Tag Grid

```gsx
templ FlexWrapGrid() {
    <div class="flex flex-wrap gap-1 grow content-center">
        for _, label := range labels {
            <div class="border-rounded p-1 w-16 flex-col items-center shrink-0">
                <span>{label}</span>
            </div>
        }
    </div>
}
```

Items have a fixed `w-16` width and `shrink-0` so they won't compress to fit. When the row fills up, `flex-wrap` pushes the rest onto a new line. The `content-center` class centers those lines vertically.

The `examples/04-layout` demo lets you switch between these patterns with Tab. The dashboard view:

![Dashboard layout](/guides/04a.png)

The flex wrap view cycles through align-content modes with arrow keys:

![Flex wrap layout](/guides/04b.png)

## Next Steps

- [State and Reactivity](state) -- Reactive state with `State[T]`
- [Components](components) -- Component patterns, composition, and lifecycle
