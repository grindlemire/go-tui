---
title: "Layout System"
order: 3
---

## Flexbox Model

go-tui uses a CSS flexbox-compatible layout engine. Set direction, justify, align, gap, padding, and margin:

```gsx
// Row layout with centered items
<div class="flex gap-2 items-center justify-center">
    <span>Left</span>
    <span>Center</span>
    <span>Right</span>
</div>

// Column layout with stretch
<div class="flex-col grow">
    <div class="border-single grow">Top section</div>
    <div class="border-single" height={3}>Bottom bar</div>
</div>
```

## Sizing

Elements can use fixed, percentage, or auto sizing:

```gsx
// Fixed size
<div width={40} height={10}>Fixed 40x10</div>

// Percentage of parent
<div widthPercent={50}>Half width</div>

// Flex grow to fill space
<div class="grow">Fills remaining space</div>

// Min/max constraints
<div minWidth={20} maxWidth={80}>Constrained</div>
```
