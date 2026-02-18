# Refs and Click Handling Reference

## Overview

Refs give you direct access to rendered elements from your event handlers. Attach a ref to an element in `.gsx`, and after the render pass completes, the ref holds a pointer to that element. From there you can read layout bounds, control scroll position, or wire up click handlers.

go-tui provides three ref types to cover different use cases:

- **`Ref`** -- a single element reference
- **`RefList`** -- an ordered collection built from `@for` loops
- **`RefMap[K]`** -- a keyed collection built from `@for` loops with a `key` attribute

All three are thread-safe.

## Ref

A reference to one element. Declare it as a struct field, create it in the constructor, attach it in `.gsx` with `ref={...}`, and read it in handlers.

### NewRef

```go
func NewRef() *Ref
```

Creates a new empty ref. The framework populates it during each render pass.

```go
type myApp struct {
    header *tui.Ref
}

func MyApp() *myApp {
    return &myApp{
        header: tui.NewRef(),
    }
}
```

### Set

```go
func (r *Ref) Set(v *Element)
```

Stores an element in the ref. You rarely call this yourself -- the generated code calls it during rendering. Each render pass overwrites the previous value, so the ref always points to the latest element.

### El

```go
func (r *Ref) El() *Element
```

Returns the stored element, or `nil` if the ref hasn't been set yet. Always check for `nil` before using the result, since the ref may be empty during the first render or if the element is conditionally hidden.

```go
el := s.header.El()
if el != nil {
    _, maxY := el.MaxScroll()
    // use maxY
}
```

### IsSet

```go
func (r *Ref) IsSet() bool
```

Returns `true` if the ref holds a non-nil element.

```go
if s.header.IsSet() {
    s.header.El().ScrollToBottom()
}
```

### Usage in .gsx

Attach a ref to any element with the `ref` attribute:

```gsx
templ (s *myApp) Render() {
    <div ref={s.header} class="border-rounded p-1">
        <span>Header content</span>
    </div>
}
```

The generated code calls `s.header.Set(element)` after creating the element, making it available in `HandleMouse`, `KeyMap`, and other handler methods.

## RefList

An ordered collection of element references, populated automatically when you use `ref=` inside a `@for` loop without a `key` attribute.

### NewRefList

```go
func NewRefList() *RefList
```

Creates an empty ref list.

```go
type listApp struct {
    items    *tui.State[[]string]
    itemRefs *tui.RefList
}

func ListApp() *listApp {
    return &listApp{
        items:    tui.NewState([]string{"alpha", "beta", "gamma"}),
        itemRefs: tui.NewRefList(),
    }
}
```

### Append

```go
func (r *RefList) Append(el *Element)
```

Adds an element to the list. Called by the generated code for each iteration of a `@for` loop where the element has a `ref` attribute.

### All

```go
func (r *RefList) All() []*Element
```

Returns a copy of all stored elements. The returned slice is safe to modify without affecting the ref list.

```go
for _, el := range s.itemRefs.All() {
    fmt.Println(el.Text())
}
```

### At

```go
func (r *RefList) At(i int) *Element
```

Returns the element at the given index, or `nil` if the index is out of bounds.

```go
el := s.itemRefs.At(2)
if el != nil {
    fmt.Println(el.Text())
}
```

### Len

```go
func (r *RefList) Len() int
```

Returns the number of elements in the list.

### Usage in .gsx

Use `ref=` inside a `@for` loop. The analyzer detects the loop context and generates `Append` calls instead of `Set`:

```gsx
templ (s *listApp) Render() {
    <div class="flex-col">
        @for _, item := range s.items.Get() {
            <span ref={s.itemRefs} class="p-1">{item}</span>
        }
    </div>
}
```

## RefMap[K]

A keyed collection of element references, populated when you use `ref=` inside a `@for` loop that also has a `key` attribute. Lets you look up a specific element by its key rather than by position.

### NewRefMap

```go
func NewRefMap[K comparable]() *RefMap[K]
```

Creates an empty ref map. The type parameter `K` must be comparable (strings, ints, and most value types work).

```go
type tabApp struct {
    tabs    *tui.State[[]string]
    tabRefs *tui.RefMap[string]
}

func TabApp() *tabApp {
    return &tabApp{
        tabs:    tui.NewState([]string{"home", "settings", "help"}),
        tabRefs: tui.NewRefMap[string](),
    }
}
```

### Put

```go
func (r *RefMap[K]) Put(key K, el *Element)
```

Stores an element under the given key. Called by the generated code for each iteration of a `@for` loop where the element has both `ref` and `key` attributes.

### Get

```go
func (r *RefMap[K]) Get(key K) *Element
```

Returns the element stored under the key, or `nil` if the key doesn't exist.

```go
el := s.tabRefs.Get("settings")
if el != nil {
    // do something with the settings tab element
}
```

### All

```go
func (r *RefMap[K]) All() map[K]*Element
```

Returns a copy of all keyed elements. Safe to iterate over without affecting the ref map.

### Len

```go
func (r *RefMap[K]) Len() int
```

Returns the number of entries in the map.

## Click Handling

go-tui pairs refs with handler functions so you don't have to compare mouse coordinates against layout bounds yourself. You list which ref maps to which action, and the framework does the hit testing.

### ClickBinding

```go
type ClickBinding struct {
    Ref *Ref
    Fn  func()
}
```

Pairs a ref with a handler function. Created by `Click` and consumed by `HandleClicks`.

### Click

```go
func Click(ref *Ref, fn func()) ClickBinding
```

Links a ref to a callback.

```go
tui.Click(s.saveBtn, s.save)
```

### HandleClicks

```go
func HandleClicks(me MouseEvent, bindings ...ClickBinding) bool
```

Tests a mouse event against a list of click bindings. Returns `true` if a binding matched and its handler was called.

How it works:

1. Ignores everything except left-button press events (`MouseLeft` + `MousePress`). Right clicks, releases, drags, and scroll wheel events return `false` immediately.
2. Iterates through bindings in order.
3. For each binding, checks that the ref is set (non-nil element) and that the click coordinates fall within the element's layout bounds via `ContainsPoint`.
4. Calls the handler of the first match and returns `true`.
5. If nothing matches, returns `false`.

Use it in your `HandleMouse` implementation:

```go
func (c *colorMixer) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(c.redUpBtn, func() { c.adjustColor(&c.red, 16) }),
        tui.Click(c.redDnBtn, func() { c.adjustColor(&c.red, -16) }),
        tui.Click(c.greenUpBtn, func() { c.adjustColor(&c.green, 16) }),
        tui.Click(c.greenDnBtn, func() { c.adjustColor(&c.green, -16) }),
    )
}
```

Binding order matters. If two elements overlap, the first matching binding wins. Place more specific bindings before general ones.

### Enabling Mouse Support

Mouse events only fire if you enable mouse support on the app:

```go
app, err := tui.NewApp(
    tui.WithRootComponent(MyApp()),
    tui.WithMouse(),
)
```

Without `WithMouse()`, your `HandleMouse` method never gets called.

## Practical Patterns

### Scroll control with refs

Use refs to read layout information for scroll clamping:

```go
type logViewer struct {
    lines    *tui.State[[]string]
    scrollY  *tui.State[int]
    content  *tui.Ref
}

func LogViewer() *logViewer {
    return &logViewer{
        lines:   tui.NewState([]string{}),
        scrollY: tui.NewState(0),
        content: tui.NewRef(),
    }
}

func (v *logViewer) scrollBy(delta int) {
    el := v.content.El()
    if el == nil {
        return
    }
    _, maxY := el.MaxScroll()
    newY := v.scrollY.Get() + delta
    if newY < 0 {
        newY = 0
    } else if newY > maxY {
        newY = maxY
    }
    v.scrollY.Set(newY)
}
```

```gsx
templ (v *logViewer) Render() {
    <div
        ref={v.content}
        class="flex-col border-single p-1"
        scrollable={tui.ScrollVertical}
        scrollOffset={0, v.scrollY.Get()}
        flexGrow={1.0}
    >
        @for _, line := range v.lines.Get() {
            <span>{line}</span>
        }
    </div>
}
```

### Clickable buttons with visual feedback

Combine refs, state, and click handling for interactive buttons:

```gsx
templ (s *myApp) Render() {
    <div class="flex gap-2">
        <button ref={s.incBtn} class="px-2 border-rounded text-green">{" + "}</button>
        <span class="font-bold">{fmt.Sprintf("%d", s.count.Get())}</span>
        <button ref={s.decBtn} class="px-2 border-rounded text-red">{" - "}</button>
    </div>
}
```

```go
func (s *myApp) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(s.incBtn, func() {
            s.count.Update(func(v int) int { return v + 1 })
        }),
        tui.Click(s.decBtn, func() {
            s.count.Update(func(v int) int { return v - 1 })
        }),
    )
}
```

### Conditional refs

When an element is conditionally rendered, its ref may be nil on some frames. Always guard access:

```go
func (s *myApp) doSomething() {
    if el := s.panel.El(); el != nil {
        el.ScrollToTop()
    }
}
```

## Thread Safety

All ref types use `sync.RWMutex` internally. `El()`, `IsSet()`, `All()`, `At()`, `Get()`, and `Len()` acquire a read lock. `Set()`, `Append()`, and `Put()` acquire a write lock.

In practice, the framework writes refs during the render pass (single-threaded) and you read them in handlers (also single-threaded on the event loop). The mutex is there for safety, not because concurrent access is common.
