# GSX Language Support for VS Code

Syntax highlighting and language support for `.gsx` files used with the [go-tui](https://github.com/grindlemire/go-tui) framework.

## Features

- **Syntax Highlighting**: Full highlighting support for the GSX DSL
  - Component declarations: `templ Name(params) { ... }`
  - Keywords: `@for`, `@if`, `@else`, `@let`
  - Element tags: `<div>`, `<span>`, `<p>`, `<button>`, `<input>`, etc.
  - Named references: `#RefName` on elements
  - Reactive state: `tui.NewState()`, `.Get()`, `.Set()`
  - Event handlers: `onClick`, `onFocus`, `onBlur`, `onKeyPress`
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
  - Auto-completion for elements, attributes, tailwind classes, and Go expressions
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
    "github.com/grindlemire/go-tui/pkg/tui"
)

templ Counter() {
    count := tui.NewState(0)
    <div class="flex-col gap-2 p-2 border-rounded">
        <span class="font-bold">{fmt.Sprintf("Count: %d", count.Get())}</span>
        <button onClick={increment(count)}>+</button>
    </div>
}

templ Dashboard(items []string) {
    <div #Main class="flex-col gap-1">
        @for _, item := range items {
            <span #Items>{item}</span>
        }
    </div>
}
```

## Supported Constructs

| Construct | Example |
|-----------|---------|
| Component | `templ Name(params) { ... }` |
| For loop | `@for i, v := range items { ... }` |
| If/Else | `@if condition { ... } @else { ... }` |
| Let binding | `@let x = <element>` |
| Component call | `@ComponentName(args)` |
| Element | `<div attr={value}>children</div>` |
| Named ref | `<div #Header>...</div>` |
| State | `count := tui.NewState(0)` |
| State access | `count.Get()`, `count.Set(v)` |
| Event handler | `onClick={handler}`, `onFocus={fn}` |
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

MIT License - see the [LICENSE](https://github.com/grindlemire/go-tui/blob/main/LICENSE) file for details.
