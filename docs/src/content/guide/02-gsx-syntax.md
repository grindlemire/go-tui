---
title: "GSX Syntax"
order: 2
---

## Components

Components are defined with the `templ` keyword and return an Element tree:

```gsx
// Simple component with parameters
templ Greeting(name string, age int) {
    <div class="flex-col gap-1 p-1 border-single">
        <span class="font-bold">{name}</span>
        <span class="font-dim">{fmt.Sprintf("Age: %d", age)}</span>
    </div>
}

// Method component on a struct
templ (c *counter) Render() {
    <div class="flex gap-2 items-center">
        <button ref={c.decBtn}>-</button>
        <span>{fmt.Sprintf("%d", c.count.Get())}</span>
        <button ref={c.incBtn}>+</button>
    </div>
}
```

## Control Flow

Use @if, @for, and @let for control flow inside templates:

```gsx
templ TodoList(items []Todo) {
    <div class="flex-col gap-1">
        @for i, item := range items {
            @let label = fmt.Sprintf("%d. %s", i+1, item.Text)
            @if item.Done {
                <span class="font-dim">{label} ✓</span>
            } @else {
                <span>{label}</span>
            }
        }
    </div>
}
```

## Composition

Components can call other components using the @ prefix:

```gsx
templ App() {
    <div class="flex-col h-full">
        @Header("My App")
        <div class="flex grow">
            @Sidebar()
            @Content()
        </div>
        @Footer()
    </div>
}
```
