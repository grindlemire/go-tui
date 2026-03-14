# GSX Language Support for VS Code

Syntax highlighting and language support for `.gsx` files used with the [go-tui](https://github.com/grindlemire/go-tui) framework.

## Features

- **Syntax Highlighting**: Full highlighting support for the GSX DSL
  - Component declarations: `templ Name(params) { ... }` and `templ (c *Type) Render() { ... }`
  - Keywords: `for`, `if`, `else`, `:=`
  - Element tags: `<div>`, `<span>`, `<p>`, `<button>`, `<input>`, `<textarea>`, `<table>`, `<progress>`, etc.
  - Ref bindings: `ref={myRef}` on elements
  - Reactive state: `tui.NewState()`, `.Get()`, `.Set()`, `.Update()`
  - Event attributes: `onFocus`, `onBlur`
  - Attributes with string, number, and expression values
  - Go expressions inside `{}`
  - Comments: `//` and `/* */`

- **Language Configuration**
  - Auto-closing brackets and quotes
  - Comment toggling
  - Code folding
  - Smart indentation

- **LSP Support** (via `tui lsp`)
  - Real-time diagnostics
  - Go-to-definition for components, functions, refs, and state
  - Hover documentation for elements, attributes, keywords, and state
  - Auto-completion for elements, attributes, Tailwind classes, and Go expressions
  - Find references across workspace
  - Document and workspace symbols
  - Semantic token highlighting
  - Code formatting

## Installation

### From Source

1. Clone the go-tui repository:

   ```bash
   git clone https://github.com/grindlemire/go-tui.git
   ```

2. Copy the extension to your VS Code extensions folder:

   ```bash
   cp -r go-tui/editor/vscode ~/.vscode/extensions/gsx-language
   ```

3. Reload VS Code

### From VSIX (Package)

1. Build the extension:

   ```bash
   cd go-tui/editor/vscode
   npm install
   npx vsce package
   ```

2. Install the `.vsix` file:
   - Open VS Code
   - Go to Extensions (Ctrl+Shift+X)
   - Click the `...` menu and select "Install from VSIX..."
   - Select the generated `.vsix` file

## Usage

Simply open any `.gsx` file and the syntax highlighting will be applied automatically.

### Example

```gsx
package main

import (
    "fmt"
    tui "github.com/grindlemire/go-tui"
)

type counter struct {
    count  *tui.State[int]
    incBtn *tui.Ref
    decBtn *tui.Ref
}

func Counter() *counter {
    return &counter{
        count:  tui.NewState(0),
        incBtn: tui.NewRef(),
        decBtn: tui.NewRef(),
    }
}

func (c *counter) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnRune('+', func(ke tui.KeyEvent) { c.count.Update(func(v int) int { return v + 1 }) }),
        tui.OnRune('-', func(ke tui.KeyEvent) { c.count.Update(func(v int) int { return v - 1 }) }),
        tui.OnRune('q', func(ke tui.KeyEvent) { ke.App().Stop() }),
    }
}

func (c *counter) HandleMouse(me tui.MouseEvent) bool {
    return tui.HandleClicks(me,
        tui.Click(c.incBtn, func() { c.count.Update(func(v int) int { return v + 1 }) }),
        tui.Click(c.decBtn, func() { c.count.Update(func(v int) int { return v - 1 }) }),
    )
}

templ (c *counter) Render() {
    <div class="flex-col gap-2 p-2 border-rounded items-center">
        <span class="font-bold">{fmt.Sprintf("Count: %d", c.count.Get())}</span>
        <div class="flex gap-2">
            <button ref={c.decBtn} class="px-1">{"-"}</button>
            <button ref={c.incBtn} class="px-1">{"+"}</button>
        </div>
        <span class="font-dim">+/- or click buttons, q to quit</span>
    </div>
}
```

## Supported Constructs

| Construct | Example |
|-----------|---------|
| Component | `templ Name(params) { ... }` |
| Method component | `templ (c *Type) Render() { ... }` |
| For loop | `for i, v := range items { ... }` |
| If/Else | `if condition { ... } else { ... }` |
| Let binding | `label := <span>text</span>` |
| Component call | `@ComponentName(args)` |
| Element | `<div class="flex-col gap-1">children</div>` |
| Self-closing element | `<hr />`, `<br />`, `<input />`, `<progress />` |
| Ref binding | `<button ref={myRef}>text</button>` |
| State access | `c.count.Get()`, `c.count.Set(v)`, `c.count.Update(fn)` |
| Event attributes | `onFocus={handler}`, `onBlur={handler}` |
| Go expression | `{fmt.Sprintf(...)}` |
| Helper function | `func helper(s string) string { ... }` |

## LSP Configuration

The extension automatically starts the GSX language server when you open a `.gsx` file. Configure via VS Code settings:

| Setting | Default | Description |
|---------|---------|-------------|
| `gsx.lsp.enabled` | `true` | Enable/disable the language server |
| `gsx.lsp.path` | `tui` | Path to the `tui` binary |
| `gsx.lsp.logPath` | `""` | Path for LSP log file (empty = no logging) |

If the `tui` binary is not found, the extension will offer to install it automatically via `go install`.

## Contributing

Contributions are welcome! Please see the [go-tui repository](https://github.com/grindlemire/go-tui) for contribution guidelines.

## License

MIT License. See the [LICENSE](https://github.com/grindlemire/go-tui/blob/main/LICENSE) file for details.
